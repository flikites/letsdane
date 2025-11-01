package happyeyeballs

import (
	"context"
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
	mu                sync.RWMutex
	enabled           bool
	dbEnabled         bool
	dnsResolutions    []DNSResolution
	connectionAttempts []ConnectionAttempt
	dbClient          MetricsStore
}

type MetricsStore interface {
	SaveConnectionAttempt(ctx context.Context, attempt *ConnectionAttempt) error
	SaveDNSResolution(ctx context.Context, resolution *DNSResolution) error
}

func NewMetrics(enabled, dbEnabled bool, store MetricsStore) *Metrics {
	return &Metrics{
		enabled:           enabled,
		dbEnabled:         dbEnabled,
		dbClient:          store,
		dnsResolutions:    make([]DNSResolution, 0),
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

	if m.dbEnabled && m.dbClient != nil {
		go func(res DNSResolution) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := m.dbClient.SaveDNSResolution(ctx, &res); err != nil {
				log.Printf("[WARN] Failed to save DNS resolution metrics: %v", err)
			}
		}(resolution)
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

	if m.dbEnabled && m.dbClient != nil {
		go func(att ConnectionAttempt) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := m.dbClient.SaveConnectionAttempt(ctx, &att); err != nil {
				log.Printf("[WARN] Failed to save connection attempt metrics: %v", err)
			}
		}(attempt)
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
