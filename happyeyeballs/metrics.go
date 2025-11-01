package happyeyeballs

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type ConnectionAttempt struct {
	Host            string
	IP              net.IP
	Family          int
	StartTime       time.Time
	EndTime         time.Time
	Success         bool
	Error           error
	Winner          bool
}

type DNSResolution struct {
	Host          string
	Family        int
	StartTime     time.Time
	EndTime       time.Time
	AddressCount  int
	Success       bool
	Error         error
}

type Metrics struct {
	mu                 sync.RWMutex
	enabled            bool
	dnsResolutions     []DNSResolution
	connectionAttempts []ConnectionAttempt
}

func NewMetrics(enabled bool) *Metrics {
	return &Metrics{
		enabled:            enabled,
		dnsResolutions:     make([]DNSResolution, 0),
		connectionAttempts: make([]ConnectionAttempt, 0),
	}
}

func (m *Metrics) RecordDNSResolution(resolution DNSResolution) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	m.dnsResolutions = append(m.dnsResolutions, resolution)
	m.mu.Unlock()

	duration := resolution.EndTime.Sub(resolution.StartTime)
	familyStr := "IPv4"
	if resolution.Family == familyIPv6 {
		familyStr = "IPv6"
	}

	if resolution.Success {
		log.Printf("[INFO] Happy Eyeballs: DNS %s lookup for %s completed in %v (%d addresses)",
			familyStr, resolution.Host, duration, resolution.AddressCount)
	} else {
		log.Printf("[WARN] Happy Eyeballs: DNS %s lookup for %s failed after %v: %v",
			familyStr, resolution.Host, duration, resolution.Error)
	}
}

func (m *Metrics) RecordConnectionAttempt(attempt ConnectionAttempt) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	m.connectionAttempts = append(m.connectionAttempts, attempt)
	m.mu.Unlock()

	duration := attempt.EndTime.Sub(attempt.StartTime)
	familyStr := "IPv4"
	if attempt.Family == familyIPv6 {
		familyStr = "IPv6"
	}

	winnerStr := ""
	if attempt.Winner {
		winnerStr = " [WINNER]"
	}

	if attempt.Success {
		log.Printf("[INFO] Happy Eyeballs: Connection to %s (%s %s) succeeded in %v%s",
			attempt.Host, familyStr, attempt.IP, duration, winnerStr)
	} else {
		log.Printf("[WARN] Happy Eyeballs: Connection to %s (%s %s) failed after %v: %v",
			attempt.Host, familyStr, attempt.IP, duration, attempt.Error)
	}
}

func (m *Metrics) GetSummary() string {
	if !m.enabled {
		return "Metrics disabled"
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var ipv6Wins, ipv4Wins int
	var ipv6Attempts, ipv4Attempts int
	var ipv6Success, ipv4Success int

	for _, attempt := range m.connectionAttempts {
		if attempt.Family == familyIPv6 {
			ipv6Attempts++
			if attempt.Success {
				ipv6Success++
				if attempt.Winner {
					ipv6Wins++
				}
			}
		} else {
			ipv4Attempts++
			if attempt.Success {
				ipv4Success++
				if attempt.Winner {
					ipv4Wins++
				}
			}
		}
	}

	return fmt.Sprintf(`Happy Eyeballs Metrics Summary:
  IPv6: %d attempts, %d successful, %d winning connections
  IPv4: %d attempts, %d successful, %d winning connections
  DNS Resolutions: %d total`,
		ipv6Attempts, ipv6Success, ipv6Wins,
		ipv4Attempts, ipv4Success, ipv4Wins,
		len(m.dnsResolutions))
}
