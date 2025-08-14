package tcpcon

import (
	"net"
	"sync"
	"time"
)

type Scanner struct {
	HostsWStatus map[string]bool
	timeout      time.Duration
	mu           sync.Mutex // protect HostsWStatus
}

// listen4Port scans all hosts concurrently and updates HostsWStatus
func (s *Scanner) Listen4Port() {
	var wg sync.WaitGroup

	// copy keys first so we don't range the map while goroutines write to it
	addrs := make([]string, 0, len(s.HostsWStatus))
	for addr := range s.HostsWStatus {
		addrs = append(addrs, addr)
	}

	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			open := isPortOpen(addr, s.timeout)
			s.mu.Lock()
			s.HostsWStatus[addr] = open
			s.mu.Unlock()
		}(addr)
	}

	wg.Wait()
}

func (s *Scanner) IsPortOpenMetrics(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, s.timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func isPortOpen(addr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func NewScanner(hosts []string, timeout time.Duration) *Scanner {
	m := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		m[h] = false
	}
	return &Scanner{
		HostsWStatus: m,
		timeout:      timeout,
	}
}
