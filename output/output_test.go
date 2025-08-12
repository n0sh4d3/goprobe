package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func sampleResults() map[string]bool {
	return map[string]bool{
		"host1:22":  true,
		"host2:80":  false,
		"host3:443": true,
	}
}

func TestWriteCSVReport(t *testing.T) {
	f, err := os.CreateTemp("", "testcsv*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err := WriteCSVReport(f.Name(), sampleResults()); err != nil {
		t.Fatalf("WriteCSVReport failed: %v", err)
	}
	f.Seek(0, io.SeekStart)
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("CSV read failed: %v", err)
	}
	if len(records) != 4 { // header + 3 rows
		t.Errorf("expected 4 rows, got %d", len(records))
	}
	if records[0][0] != "hostname" || records[0][2] != "status" {
		t.Errorf("header mismatch: %v", records[0])
	}
}

func TestWriteJSONReport(t *testing.T) {
	f, err := os.CreateTemp("", "testjson*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err := WriteJSONReport(f.Name(), sampleResults()); err != nil {
		t.Fatalf("WriteJSONReport failed: %v", err)
	}
	f.Seek(0, io.SeekStart)
	var out []HostStatus
	dec := json.NewDecoder(f)
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("JSON decode failed: %v", err)
	}
	if len(out) != 3 {
		t.Errorf("expected 3 objects, got %d", len(out))
	}
	if out[0].Status != "open" && out[0].Status != "closed" {
		t.Errorf("unexpected status: %v", out[0].Status)
	}
}

func TestPrintTable(t *testing.T) {
	buf := new(bytes.Buffer)
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
		w.Close()
	}()
	PrintTable(sampleResults())
	w.Close()
	io.Copy(buf, r)
	out := buf.String()
	if !strings.Contains(out, "hostname") || !strings.Contains(out, "open") {
		t.Errorf("table output missing expected content: %s", out)
	}
}

func FuzzWriteCSVReport(f *testing.F) {
	f.Add("host:22", true)
	f.Fuzz(func(t *testing.T, addr string, open bool) {
		results := map[string]bool{addr: open}
		f, err := os.CreateTemp("", "fuzzcsv*.csv")
		if err != nil {
			t.Skip()
		}
		defer os.Remove(f.Name())
		defer f.Close()
		_ = WriteCSVReport(f.Name(), results)
	})
}

func FuzzWriteJSONReport(f *testing.F) {
	f.Add("host:22", true)
	f.Fuzz(func(t *testing.T, addr string, open bool) {
		results := map[string]bool{addr: open}
		f, err := os.CreateTemp("", "fuzzjson*.json")
		if err != nil {
			t.Skip()
		}
		defer os.Remove(f.Name())
		defer f.Close()
		_ = WriteJSONReport(f.Name(), results)
	})
}

func TestWriteCSVReport_Empty(t *testing.T) {
	f, err := os.CreateTemp("", "emptycsv*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err := WriteCSVReport(f.Name(), map[string]bool{}); err != nil {
		t.Fatalf("WriteCSVReport failed: %v", err)
	}
}

func TestWriteJSONReport_Empty(t *testing.T) {
	f, err := os.CreateTemp("", "emptyjson*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err := WriteJSONReport(f.Name(), map[string]bool{}); err != nil {
		t.Fatalf("WriteJSONReport failed: %v", err)
	}
}

func BenchmarkWriteCSVReport(b *testing.B) {
	results := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		results[fmt.Sprintf("host%d:%d", i, i)] = i%2 == 0
	}
	f, err := os.CreateTemp("", "benchcsv*.csv")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WriteCSVReport(f.Name(), results)
	}
}

func BenchmarkWriteJSONReport(b *testing.B) {
	results := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		results[fmt.Sprintf("host%d:%d", i, i)] = i%2 == 0
	}
	f, err := os.CreateTemp("", "benchjson*.json")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WriteJSONReport(f.Name(), results)
	}
}

func BenchmarkPrintTable(b *testing.B) {
	results := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		results[fmt.Sprintf("host%d:%d", i, i)] = i%2 == 0
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PrintTable(results)
	}
}

func TestWriteCSVReport_InvalidPath(t *testing.T) {
	err := WriteCSVReport("/invalid/path/to/file.csv", sampleResults())
	if err == nil {
		t.Errorf("expected error for invalid path")
	}
}

func TestWriteJSONReport_InvalidPath(t *testing.T) {
	err := WriteJSONReport("/invalid/path/to/file.json", sampleResults())
	if err == nil {
		t.Errorf("expected error for invalid path")
	}
}

func TestWriteCSVReport_Overwrite(t *testing.T) {
	f, err := os.CreateTemp("", "overwritecsv*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	_ = WriteCSVReport(f.Name(), sampleResults())
	_ = WriteCSVReport(f.Name(), sampleResults())
}

func TestWriteJSONReport_Overwrite(t *testing.T) {
	f, err := os.CreateTemp("", "overwritejson*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	_ = WriteJSONReport(f.Name(), sampleResults())
	_ = WriteJSONReport(f.Name(), sampleResults())
}

func TestWriteCSVReport_SpecialChars(t *testing.T) {
	results := map[string]bool{"h@st$:22": true, "host,2:80": false, "host3:443\n": true}
	f, err := os.CreateTemp("", "specialcharscsv*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	_ = WriteCSVReport(f.Name(), results)
}

func TestWriteJSONReport_SpecialChars(t *testing.T) {
	results := map[string]bool{"h@st$:22": true, "host,2:80": false, "host3:443\n": true}
	f, err := os.CreateTemp("", "specialcharsjson*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	_ = WriteJSONReport(f.Name(), results)
}

func TestPrintTable_AllOpenClosed(t *testing.T) {
    results := map[string]bool{"host1:22": true, "host2:80": true}
    buf := new(bytes.Buffer)
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    PrintTable(results)
    w.Close()
    os.Stdout = old
    io.Copy(buf, r)
    out := buf.String()
    if !strings.Contains(out, "open") {
        t.Errorf("expected 'open' in output")
    }

    results = map[string]bool{"host1:22": false, "host2:80": false}
    buf.Reset()
    r, w, _ = os.Pipe()
    os.Stdout = w
    PrintTable(results)
    w.Close()
    os.Stdout = old
    io.Copy(buf, r)
    out = buf.String()
    if !strings.Contains(out, "closed") {
        t.Errorf("expected 'closed' in output")
    }
}

func FuzzPrintTable(f *testing.F) {
	f.Add("host:22", true)
	f.Fuzz(func(t *testing.T, addr string, open bool) {
		results := map[string]bool{addr: open}
		PrintTable(results)
	})
}
