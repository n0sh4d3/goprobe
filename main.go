package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	hostsFile string
	ports     []string
	timeout   int

	csvPathOpt  string // holds value if user provided one
	jsonPathOpt string // holds value if user provided one
	writeCSV    bool   // toggled when --csv present without value
	writeJSON   bool   // toggled when --json present without value
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
			combo := fmt.Sprintf("%s:%s", host, port)
			hostWport = append(hostWport, combo)
		}
	}

	return hostWport
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "goprobe [flags]",
		Short: "Check port availability for a list of hosts",
		Long: `goprobe reads a list of hosts from a file (one per line) and probes TCP ports.

Ports can be provided as a comma-separated list or via multiple --ports flags.
If --ports is omitted or given without a value, defaults (22,80,443) are used.

Optionally, results can be written to CSV and/or JSON. Supplying --csv or --json
without a value writes to default filenames (goprobe.csv / goprobe.json)
in the current directory.`,
		Example: `
  # Use default ports (22,80,443)
  goprobe --hosts hosts.txt

  # Custom ports (comma-separated)
  goprobe --hosts hosts.txt -p 8080,8443

  # Multiple flags accumulate
  goprobe --hosts hosts.txt -p 22 -p 443

  # Save CSV and JSON with default filenames
  goprobe --hosts hosts.txt --csv --json

  # Custom output filenames
  goprobe --hosts hosts.txt --csv out/report.csv --json out/report.json

  # Error on explicit empty ports list
  goprobe --hosts hosts.txt --ports=`,
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

			data, err := os.ReadFile(hostsFile)
			if err != nil {
				return err
			}
			hosts, err := fileToStrSlice(data)
			if err != nil {
				return err
			}

			results := probe(hosts, ports)

			for _, r := range results {
				fmt.Fprintln(cmd.OutOrStdout(), r)
			}

			if writeCSV || csvPathOpt != "" {
				path := csvPathOpt
				if path == "" {
					path = "goprobe.csv"
				}
				if err := writeCSVReport(path, results); err != nil {
					return fmt.Errorf("write csv: %w", err)
				}
			}
			if writeJSON || jsonPathOpt != "" {
				path := jsonPathOpt
				if path == "" {
					path = "goprobe.json"
				}
				if err := writeJSONReport(path, results); err != nil {
					return fmt.Errorf("write json: %w", err)
				}
			}

			return nil
		},
	}

	defaultPorts := []string{"22", "80", "443"}

	rootCmd.Flags().StringVar(&hostsFile, "hosts", "", "hosts file to check port availability against")
	rootCmd.Flags().StringSliceVar(&ports, "ports", defaultPorts, "ports to check availability")
	_ = rootCmd.MarkFlagRequired("hosts")

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

func writeCSVReport(path string, contents []string) error {

	return nil
}

func writeJSONReport(path string, contents []string) error {

	return nil
}
