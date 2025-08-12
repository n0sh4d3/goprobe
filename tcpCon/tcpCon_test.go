package tcpcon

import (
	"net"
	"testing"
	"time"
)

// startTCP starts a listener on 127.0.0.1:0 (OS picks a free port).
// it returns "host:port" and a cleanup func.
func startTCP(t *testing.T) (addr string, cleanup func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return ln.Addr().String(), func() { _ = ln.Close() }
}

func TestScanner_Listen4Port_Mixed(t *testing.T) {
	open1, close1 := startTCP(t)
	defer close1()
	open2, close2 := startTCP(t)
	defer close2()

	// choose a closed port: bind and close to learn an address that will be closed.
	tmp, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen temp: %v", err)
	}
	closed := tmp.Addr().String()
	_ = tmp.Close()

	hosts := []string{open1, open2, closed}

	s := NewScanner(hosts, 300*time.Millisecond)
	// IMPORTANT: your Listen4Port must call wg.Wait() internally
	s.Listen4Port()

	if !s.hostsWStatus[open1] {
		t.Errorf("expected %s open", open1)
	}
	if !s.hostsWStatus[open2] {
		t.Errorf("expected %s open", open2)
	}
	if s.hostsWStatus[closed] {
		t.Errorf("expected %s closed", closed)
	}
}

func TestScanner_Timeout_IsUsed(t *testing.T) {
	// a port that shouldn't accept connections.
	tmp, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := tmp.Addr().String()
	_ = tmp.Close()

	s := NewScanner([]string{closed}, 100*time.Millisecond)

	start := time.Now()
	s.Listen4Port()
	elapsed := time.Since(start)

	if elapsed > 800*time.Millisecond { // generous ceiling on slow CI
		t.Fatalf("scan took too long: %v", elapsed)
	}
}

func TestScanner_EmptyHosts(t *testing.T) {
	s := NewScanner([]string{}, 100*time.Millisecond)
	s.Listen4Port()
	if len(s.hostsWStatus) != 0 {
		t.Errorf("expected no hosts, got %d", len(s.hostsWStatus))
	}
}

func TestScanner_PropertyKeysMatch(t *testing.T) {
	hosts := []string{"127.0.0.1:1", "127.0.0.1:2"}
	s := NewScanner(hosts, 100*time.Millisecond)
	s.Listen4Port()
	for _, h := range hosts {
		if _, ok := s.hostsWStatus[h]; !ok {
			t.Errorf("missing key: %s", h)
		}
	}
}

func ExampleNewScanner() {
	hosts := []string{"127.0.0.1:80"}
	s := NewScanner(hosts, 100*time.Millisecond)
	s.Listen4Port()
	for h, open := range s.hostsWStatus {
		println(h, open)
	}
	// Output:
}

func FuzzNewScanner(f *testing.F) {
	f.Add("127.0.0.1:80", 100)
	f.Fuzz(func(t *testing.T, host string, ms int) {
		timeout := time.Duration(ms%1000) * time.Millisecond
		s := NewScanner([]string{host}, timeout)
		s.Listen4Port()
	})
}

func BenchmarkListen4Port(b *testing.B) {
	hosts := make([]string, 100)
	for i := range hosts {
		hosts[i] = "127.0.0.1:0"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewScanner(hosts, 10*time.Millisecond)
		s.Listen4Port()
	}
}

func TestScanner_InvalidHosts(t *testing.T) {
	hosts := []string{"not_a_host", "256.256.256.256:99999", "localhost:notaport", ""}
	s := NewScanner(hosts, 50*time.Millisecond)
	s.Listen4Port()
	for _, h := range hosts {
		if s.hostsWStatus[h] {
			t.Errorf("expected %s to be closed/invalid", h)
		}
	}
}

func TestScanner_LargeHostList(t *testing.T) {
	hosts := make([]string, 2000)
	for i := range hosts {
		hosts[i] = "127.0.0.1:0"
	}
	s := NewScanner(hosts, 1*time.Millisecond)
	s.Listen4Port()
	if len(s.hostsWStatus) != 2000 {
		t.Errorf("expected 2000 hosts, got %d", len(s.hostsWStatus))
	}
}

func TestScanner_ConcurrentUsage(t *testing.T) {
	hosts := []string{"127.0.0.1:0", "127.0.0.1:0"}
	s1 := NewScanner(hosts, 10*time.Millisecond)
	s2 := NewScanner(hosts, 10*time.Millisecond)
	done := make(chan struct{}, 2)
	go func() { s1.Listen4Port(); done <- struct{}{} }()
	go func() { s2.Listen4Port(); done <- struct{}{} }()
	<-done
	<-done
	// Both scanners should have results
	for _, s := range []*Scanner{s1, s2} {
		for _, h := range hosts {
			_ = s.hostsWStatus[h] // just check no panic
		}
	}
}
