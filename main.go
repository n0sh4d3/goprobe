package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/n0sh4d3/goprobe/output"
	tcpcon "github.com/n0sh4d3/goprobe/tcpCon"
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
)

/*
	 logic here

	 first open hosts file into slice
		(in goroutine?) create new slice with using hostsFile + all given ports
	 (go routine) start async fetching
	 now "wrtier function" that will write or to std out or to .csv of to .json all of the results
*/

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

// Refactored main logic for testability
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
	scanner.Listen4Port()

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

QUICK START:
  1. Create a text file (e.g. hosts.txt) with one host per line.
  2. Run: goprobe --hosts hosts.txt
  3. See results printed in a table.

FLAGS:
  --hosts <file>      (required) Path to file with hosts, one per line
  --ports <list>      Ports to check, comma-separated (default: 22,80,443)
  --timeout <dur>     Timeout for each connection (e.g. 500ms, 2s)
  --csv [file]        Write results to CSV (default: goprobe.csv)
  --json [file]       Write results to JSON (default: goprobe.json)
  --stdout            Print results to terminal (table by default, or CSV/JSON if combined)

EXAMPLES:
  # Basic usage (table output)
  goprobe --hosts hosts.txt

  # Custom ports
  goprobe --hosts hosts.txt --ports 8080,8443

  # Save results to CSV and JSON
  goprobe --hosts hosts.txt --csv --json

  # Custom output filenames
  goprobe --hosts hosts.txt --csv out.csv --json out.json

  # Print results as CSV to terminal
  goprobe --hosts hosts.txt --csv --stdout

  # Print results as JSON to terminal
  goprobe --hosts hosts.txt --json --stdout

  # Print results as table to terminal (explicit)
  goprobe --hosts hosts.txt --stdout

  # Error on explicit empty ports list
  goprobe --hosts hosts.txt --ports=

TIPS:
  - You can use --ports multiple times: --ports 22 --ports 443
  - If you don't specify any output flags, results print as a table by default.
  - Use --timeout to avoid waiting too long for slow hosts.
  - All output files are created in the current directory unless you specify a path.
`,
		Example: "See above for examples.",
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
