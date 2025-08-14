package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/n0sh4d3/goprobe/output"
	tcpcon "github.com/n0sh4d3/goprobe/tcpCon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	hostsFile string
	ports     []string
	timeout   time.Duration

	csvPathOpt  string // holds value if user provided one
	jsonPathOpt string // holds value if user provided one
	writeCSV    bool   // toggled when --csv present without value
	writeJSON   bool   // toggled when --json present without value
	writeStdout bool   // toggled when --stdout present
	metricsAddr string
)

var (
	probeAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goprobe_attempts_total",
			Help: "Total number of probe attempts",
		},
		[]string{"host", "port"},
	)
	probeSuccesses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goprobe_success_total",
			Help: "Total number of successful probes",
		},
		[]string{"host", "port"},
	)
	probeFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goprobe_failure_total",
			Help: "Total number of failed probes",
		},
		[]string{"host", "port"},
	)
	probeLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "goprobe_latency_seconds",
			Help:    "Probe latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"host", "port"},
	)
)

func init() {
	prometheus.MustRegister(probeAttempts, probeSuccesses, probeFailures, probeLatency)
}

// takes hosts as slice (hosts aren't valdiated) and merges em with ports
//
// eg probe([]string{"test.com", "a.b"}, []string{"20","30"})
// will return -> []string{"test.com:20", "test.com:30", "a.b:20", "a.b:30"}
func probe(hostsFileContent []string, ports []string) []string {
	hostWport := []string{}

	for _, host := range hostsFileContent {
		for _, port := range ports {
			if port != "" {
				combo := fmt.Sprintf("%s:%s", host, port)
				hostWport = append(hostWport, combo)
			}
		}
	}

	return hostWport
}

func RunProbe(hostsFile string, ports []string, timeout time.Duration, csvPathOpt, jsonPathOpt string, writeCSV, writeJSON, writeStdout bool) error {
	data, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}
	hosts, err := fileToStrSlice(data)
	if err != nil {
		return err
	}
	addrs := probe(hosts, ports)
	scanner := tcpcon.NewScanner(addrs, timeout)

	// instrumented probe logic
	var wg sync.WaitGroup
	results := make(map[string]bool, len(addrs))
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			parts := strings.Split(addr, ":")
			host, port := parts[0], ""
			if len(parts) > 1 {
				port = parts[1]
			}
			start := time.Now()
			probeAttempts.WithLabelValues(host, port).Inc()
			open := scanner.IsPortOpenMetrics(addr)
			latency := time.Since(start).Seconds()
			probeLatency.WithLabelValues(host, port).Observe(latency)
			if open {
				probeSuccesses.WithLabelValues(host, port).Inc()
			} else {
				probeFailures.WithLabelValues(host, port).Inc()
			}
			results[addr] = open
		}(addr)
	}
	wg.Wait()
	scanner.HostsWStatus = results

	outputSelected := writeCSV || csvPathOpt != "" || writeJSON || jsonPathOpt != "" || writeStdout

	if writeStdout {
		printed := false
		if writeCSV || csvPathOpt != "" {
			output.WriteCSVReport("/dev/stdout", scanner.HostsWStatus)
			printed = true
		}
		if writeJSON || jsonPathOpt != "" {
			output.WriteJSONReport("/dev/stdout", scanner.HostsWStatus)
			printed = true
		}
		if !printed {
			output.PrintTable(scanner.HostsWStatus)
		}
	} else if !outputSelected {
		output.PrintTable(scanner.HostsWStatus)
	}

	if writeCSV || csvPathOpt != "" {
		path := csvPathOpt
		if path == "" {
			path = "goprobe.csv"
		}
		if err := output.WriteCSVReport(path, scanner.HostsWStatus); err != nil {
			return fmt.Errorf("write csv: %w", err)
		}
	}
	if writeJSON || jsonPathOpt != "" {
		path := jsonPathOpt
		if path == "" {
			path = "goprobe.json"
		}
		if err := output.WriteJSONReport(path, scanner.HostsWStatus); err != nil {
			return fmt.Errorf("write json: %w", err)
		}
	}
	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "goprobe [flags]",
		Short: "Check port availability for a list of hosts",
		Long: `goprobe is a simple, user-friendly tool to check if TCP ports are open on a list of hosts.

quick start:
  1. create a text file (e.g. hosts.txt) with one host per line.
  2. run: goprobe --hosts hosts.txt
  3. see results printed in a table.

flags:
  --hosts <file>      (required) path to file with hosts, one per line
  --ports <list>      ports to check, comma-separated (default: 22,80,443)
  --timeout <dur>     timeout for each connection (e.g. 500ms, 2s)
  --csv [file]        write results to CSV (default: goprobe.csv)
  --json [file]       write results to JSON (default: goprobe.json)
  --stdout            print results to terminal (table by default, or CSV/JSON if combined)

examples:
  # basic usage (table output)
  goprobe --hosts hosts.txt

  # custom ports
  goprobe --hosts hosts.txt --ports 8080,8443

  # save results to CSV and JSON
  goprobe --hosts hosts.txt --csv --json

  # custom output filenames
  goprobe --hosts hosts.txt --csv out.csv --json out.json

  # print results as CSV to terminal
  goprobe --hosts hosts.txt --csv --stdout

  # print results as JSON to terminal
  goprobe --hosts hosts.txt --json --stdout

  # print results as table to terminal (explicit)
  goprobe --hosts hosts.txt --stdout

  # error on explicit empty ports list
  goprobe --hosts hosts.txt --ports=

tips:
  - you can use --ports multiple times: --ports 22 --ports 443
  - if you don't specify any output flags, results print as a table by default.
  - use --timeout to avoid waiting too long for slow hosts.
  - all output files are created in the current directory unless you specify a path.
`,
		Example: "see above for examples.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// validate --ports= (explicit empty)
			if f := cmd.Flags().Lookup("ports"); f != nil && f.Changed {
				// if user passed --ports= (empty) we reject with a helpful message
				if len(ports) == 0 || (len(ports) == 1 && strings.TrimSpace(ports[0]) == "") {
					return fmt.Errorf("--ports= provided without any value\nuse --ports for defaults (22,80,443)\nor --ports=<port[,port,...]> for specific ports")
				}
			}
			// start metrics server
			go func() {
				http.Handle("/metrics", promhttp.Handler())
				http.ListenAndServe(metricsAddr, nil)
			}()
			return RunProbe(hostsFile, ports, timeout, csvPathOpt, jsonPathOpt, writeCSV, writeJSON, writeStdout)
		},
	}

	defaultPorts := []string{"22", "80", "443"}

	rootCmd.Flags().StringVar(&hostsFile, "hosts", "", "hosts file to check port availability against")
	rootCmd.Flags().DurationVar(&timeout, "timeout", 5*time.Second,
		"per-connection timeout (e.g., 500ms, 2s, 5s)")
	rootCmd.Flags().StringSliceVar(&ports, "ports", defaultPorts, "ports to check availability")
	_ = rootCmd.MarkFlagRequired("hosts")

	rootCmd.Flags().StringVar(&csvPathOpt, "csv", "", "write results to CSV file (default: goprobe.csv)")
	rootCmd.Flags().StringVar(&jsonPathOpt, "json", "", "write results to JSON file (default: goprobe.json)")
	rootCmd.Flags().BoolVar(&writeStdout, "stdout", false, "print results to stdout as a table")
	rootCmd.Flags().StringVar(&metricsAddr, "metrics-addr", ":9090", "address for Prometheus metrics endpoint (e.g. :9090)")
	if f := rootCmd.Flags().Lookup("metrics-addr"); f != nil {
		f.NoOptDefVal = ":9090"
	}

	if f := rootCmd.Flags().Lookup("csv"); f != nil {
		f.NoOptDefVal = "goprobe.csv"
	}
	if f := rootCmd.Flags().Lookup("json"); f != nil {
		f.NoOptDefVal = "goprobe.json"
	}

	if f := rootCmd.Flags().Lookup("ports"); f != nil {
		f.NoOptDefVal = strings.Join(defaultPorts, ",")
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// takes a file contents as []byte tries to read it and returns each line in []string
func fileToStrSlice(data []byte) ([]string, error) {
	content := strings.TrimRight(string(data), "\r\n")
	if content == "" {
		return []string{}, nil
	}
	return strings.Split(content, "\n"), nil
}
