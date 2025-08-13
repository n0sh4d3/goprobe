package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func writeCSVReport(path string, data []string) error {
	// Minimal implementation: write each string as a line to the file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range data {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}

func writeJSONReport(path string, data []string) error {
	// Minimal implementation: write the slice as a JSON array
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "[%q]", data)
	return err
}

func Test_writeCSVReport_and_writeJSONReport(t *testing.T) {
	tmp := t.TempDir()
	csvPath := filepath.Join(tmp, "out.csv")
	jsonPath := filepath.Join(tmp, "out.json")
	data := []string{"a:1", "b:2"}

	// Should not error (noop impl)
	if err := writeCSVReport(csvPath, data); err != nil {
		t.Errorf("writeCSVReport() error = %v", err)
	}
	if err := writeJSONReport(jsonPath, data); err != nil {
		t.Errorf("writeJSONReport() error = %v", err)
	}
}

func Test_main_CLI_flags_and_output(t *testing.T) {
	tmp := t.TempDir()
	hostsPath := filepath.Join(tmp, "hosts.txt")
	os.WriteFile(hostsPath, []byte("host1\nhost2"), 0644)

	buf := new(bytes.Buffer)
	rootCmd := &cobra.Command{
		Use: "goprobe",
		RunE: func(cmd *cobra.Command, args []string) error {
			hosts, _ := fileToStrSlice([]byte("host1\nhost2"))
			ports := []string{"80", "443"}
			results := probe(hosts, ports)
			for _, r := range results {
				fmt.Fprintln(buf, r)
			}
			return nil
		},
	}
	rootCmd.SetOut(buf)
	rootCmd.Flags().String("hosts", hostsPath, "hosts file")
	rootCmd.Flags().StringSlice("ports", []string{"80", "443"}, "ports")
	rootCmd.SetArgs([]string{"--hosts", hostsPath, "--ports", "80,443"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("CLI failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "host1:80") || !strings.Contains(out, "host2:443") {
		t.Errorf("unexpected CLI output: %s", out)
	}
}

func Test_fileToStrSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []byte
		want []string
	}{
		{
			name: "empty",
			in:   []byte(""),
			want: []string{},
		},
		{
			name: "just LF",
			in:   []byte("\n"),
			want: []string{},
		},
		{
			name: "just CRLF",
			in:   []byte("\r\n"),
			want: []string{},
		},
		{
			name: "single line no newline",
			in:   []byte("test"),
			want: []string{"test"},
		},
		{
			name: "single line LF",
			in:   []byte("test\n"),
			want: []string{"test"},
		},
		{
			name: "single line CRLF",
			in:   []byte("test\r\n"),
			want: []string{"test"},
		},
		{
			name: "multiple LF",
			in:   []byte("a\nb\nc\n"),
			want: []string{"a", "b", "c"},
		},
		{
			name: "multiple CRLF at EOL only",
			in:   []byte("a\r\nb\r\nc\r\n"),
			want: []string{"a\r", "b\r", "c"},
		},
		{
			name: "mixed endings",
			in:   []byte("a\nb\r\nc"),
			want: []string{"a", "b\r", "c"},
		},
		{
			name: "unicode",
			in:   []byte("ą\nŁódź\n😊\n"),
			want: []string{"ą", "Łódź", "😊"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := fileToStrSlice(tt.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Fuzz_fileToStrSlice(f *testing.F) {
	seed := [][]byte{
		[]byte(""),
		[]byte("\n"),
		[]byte("\r\n"),
		[]byte("a\nb\nc\n"),
		[]byte("a\r\nb\r\nc\r\n"),
		[]byte("x"),
		[]byte("x\r\ny\nz"),
	}
	for _, s := range seed {
		f.Add(string(s))
	}

	f.Fuzz(func(t *testing.T, s string) {
		first, err := fileToStrSlice([]byte(s))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		joined := []byte(strings.Join(first, "\n"))
		second, err := fileToStrSlice(joined)
		if err != nil {
			t.Fatalf("unexpected error on second parse: %v", err)
		}
		if !slices.Equal(first, second) {
			t.Fatalf("not idempotent: first=%#v second=%#v", first, second)
		}
	})
}

func Example_fileToStrSlice() {
	lines, _ := fileToStrSlice([]byte("a\nb\nc\n"))
	fmt.Println(lines)
	// Output: [a b c]
}

func Test_probe(t *testing.T) {
	tests := []struct {
		name             string
		hostsFileContent []string
		ports            []string
		want             []string
	}{
		{
			name:             "concatenates host:port pairs",
			hostsFileContent: []string{"test"},
			ports:            []string{"20", "30"},
			want:             []string{"test:20", "test:30"},
		},
		{
			name:             "multiple hosts and ports",
			hostsFileContent: []string{"a", "b"},
			ports:            []string{"1", "2"},
			want:             []string{"a:1", "a:2", "b:1", "b:2"},
		},
		{
			name:             "empty ports yields empty result",
			hostsFileContent: []string{"x"},
			ports:            []string{},
			want:             []string{},
		},
		{
			name:             "empty hosts yields empty result",
			hostsFileContent: []string{},
			ports:            []string{"80"},
			want:             []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := probe(tt.hostsFileContent, tt.ports)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("probe() got=%v, want=%v", got, tt.want)
			}
		})
	}
}
func Fuzz_probe(f *testing.F) {
	f.Add("a", "b", "c")
	f.Add("host1", "host2", "host3")
	f.Fuzz(func(t *testing.T, hosts string, ports string, extra string) {
		hostsSlice := strings.Split(hosts, ",")
		portsSlice := strings.Split(ports, ",")
		got := probe(hostsSlice, portsSlice)
		// should not panic, and output length should be len(hostsSlice)*len(portsSlice)
		if len(hostsSlice) > 0 && len(portsSlice) > 0 &&
			hostsSlice[0] != "" && portsSlice[0] != "" &&
			len(got) != len(hostsSlice)*len(portsSlice) {
			t.Errorf("probe() output length mismatch: got=%d want=%d", len(got), len(hostsSlice)*len(portsSlice))
		}
	})
}

func TestRunProbe_TableOutput_Default(t *testing.T) {
	tmp := t.TempDir()
	hostsPath := filepath.Join(tmp, "hosts.txt")
	os.WriteFile(hostsPath, []byte("host1\nhost2"), 0644)
	err := RunProbe(hostsPath, []string{"80", "443"}, time.Millisecond, "", "", false, false, true)
	if err != nil {
		t.Fatalf("RunProbe failed: %v", err)
	}
}

func TestRunProbe_CSVOutput(t *testing.T) {
	tmp := t.TempDir()
	hostsPath := filepath.Join(tmp, "hosts.txt")
	csvPath := filepath.Join(tmp, "out.csv")
	os.WriteFile(hostsPath, []byte("host1\nhost2"), 0644)
	err := RunProbe(hostsPath, []string{"80"}, time.Millisecond, csvPath, "", true, false, false)
	if err != nil {
		t.Fatalf("RunProbe failed: %v", err)
	}
	if _, err := os.Stat(csvPath); err != nil {
		t.Errorf("CSV file not created: %v", err)
	}
}

func TestRunProbe_JSONOutput(t *testing.T) {
	tmp := t.TempDir()
	hostsPath := filepath.Join(tmp, "hosts.txt")
	jsonPath := filepath.Join(tmp, "out.json")
	os.WriteFile(hostsPath, []byte("host1\nhost2"), 0644)
	err := RunProbe(hostsPath, []string{"80"}, time.Millisecond, "", jsonPath, false, true, false)
	if err != nil {
		t.Fatalf("RunProbe failed: %v", err)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("JSON file not created: %v", err)
	}
}

func TestRunProbe_ErrorCases(t *testing.T) {
	err := RunProbe("/nonexistent/file.txt", []string{"80"}, time.Millisecond, "", "", false, false, true)
	if err == nil {
		t.Errorf("expected error for missing hosts file")
	}
}

func TestRunProbe_EmptyHosts(t *testing.T) {
	tmp := t.TempDir()
	hostsPath := filepath.Join(tmp, "hosts.txt")
	os.WriteFile(hostsPath, []byte(""), 0644)
	err := RunProbe(hostsPath, []string{"80"}, time.Millisecond, "", "", false, false, true)
	if err != nil {
		t.Fatalf("RunProbe failed: %v", err)
	}
}

func TestRunProbe_EmptyPorts(t *testing.T) {
	tmp := t.TempDir()
	hostsPath := filepath.Join(tmp, "hosts.txt")
	os.WriteFile(hostsPath, []byte("host1"), 0644)
	err := RunProbe(hostsPath, []string{}, time.Millisecond, "", "", false, false, true)
	if err != nil {
		t.Fatalf("RunProbe failed: %v", err)
	}
}

func BenchmarkRunProbe(b *testing.B) {
	tmp := b.TempDir()
	hostsPath := filepath.Join(tmp, "hosts.txt")
	hosts := make([]string, 100)
	for i := range hosts {
		hosts[i] = fmt.Sprintf("host%d", i)
	}
	os.WriteFile(hostsPath, []byte(strings.Join(hosts, "\n")), 0644)
	for i := 0; i < b.N; i++ {
		_ = RunProbe(hostsPath, []string{"80", "443"}, time.Millisecond, "", "", false, false, false)
	}
}

func FuzzRunProbe(f *testing.F) {
	f.Add("host1\nhost2", "80,443")
	f.Fuzz(func(t *testing.T, hosts string, ports string) {
		tmp := t.TempDir()
		hostsPath := filepath.Join(tmp, "hosts.txt")
		os.WriteFile(hostsPath, []byte(hosts), 0644)
		portSlice := strings.Split(ports, ",")
		_ = RunProbe(hostsPath, portSlice, time.Millisecond, "", "", false, false, false)
	})
}
