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
	return enc.Encode(out)
}

func PrintTable(results map[string]bool) {
	fmt.Printf("%-20s %-8s %-8s\n", "hostname", "port", "status")
	fmt.Printf("%-20s %-8s %-8s\n", strings.Repeat("-", 20), strings.Repeat("-", 8), strings.Repeat("-", 8))
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
		fmt.Printf("%-20s %-8s %-8s\n", host, port, status)
	}
}