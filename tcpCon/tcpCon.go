package tcpcon

import (
	"net"
	"sync"
	"time"
)

type Scanner struct {
	HostsWStatus map[string]bool
	timeout      time.Duration
	mu           sync.Mutex // protect hostsWStatus
}

func (s *Scanner) Listen4Port() {
	var wg sync.WaitGroup

	// copy keys first so we don't range the map while goroutines write to it
	addrs := make([]string, 0, len(s.HostsWStatus))
	for addr := range s.HostsWStatus {
		addrs = append(addrs, addr)
	}

	for _, addr := range addrs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			open := s.isPortOpen(addr)
			s.mu.Lock()
			s.HostsWStatus[addr] = open
			s.mu.Unlock()
		}()
	}

	wg.Wait()
}

func (s *Scanner) isPortOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, s.timeout)
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
