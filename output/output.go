package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type HostStatus struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Status string `json:"status"`
}

func WriteCSVReport(path string, results map[string]bool) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	w.Write([]string{"hostname", "port", "status"})
	for addr, open := range results {
		parts := strings.Split(addr, ":")
		host, port := parts[0], ""
		if len(parts) > 1 {
			port = parts[1]
		}
		status := "closed"
		if open {
			status = "open"
		}
		w.Write([]string{host, port, status})
	}
	if path != "/dev/stdout" {
		fmt.Printf("\033[35m[INFO]\033[0m CSV file created: %s\n", path)
	}
	return nil
}

func WriteJSONReport(path string, results map[string]bool) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var out []HostStatus
	for addr, open := range results {
		parts := strings.Split(addr, ":")
		host, port := parts[0], ""
		if len(parts) > 1 {
			port = parts[1]
		}
		status := "closed"
		if open {
			status = "open"
		}
		out = append(out, HostStatus{Host: host, Port: port, Status: status})
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return err
	}
	if path != "/dev/stdout" {
		fmt.Printf("\033[35m[INFO]\033[0m JSON file created: %s\n", path)
	}
	return nil
}

func PrintTable(results map[string]bool) {
	const (
		green  = "\033[32m"
		red    = "\033[31m"
		yellow = "\033[33m"
		cyan   = "\033[36m"
		reset  = "\033[0m"
	)
	fmt.Printf(cyan+"%-20s %-8s %-8s\n"+reset, "hostname", "port", "status")
	fmt.Printf(cyan+"%-20s %-8s %-8s\n"+reset, strings.Repeat("-", 20), strings.Repeat("-", 8), strings.Repeat("-", 8))
	for addr, open := range results {
		parts := strings.Split(addr, ":")
		host, port := parts[0], ""
		if len(parts) > 1 {
			port = parts[1]
		}
		status := "closed"
		color := red
		if open {
			status = "open"
			color = green
		}
		fmt.Printf(yellow+"%-20s %-8s "+reset+"%s%-8s%s\n", host, port, color, status, reset)
	}
}