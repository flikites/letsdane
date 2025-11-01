package happyeyeballs

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestConcurrentDNSLookup(t *testing.T) {
	tests := []struct {
		name            string
		ipv6Result      []net.IP
		ipv4Result      []net.IP
		ipv6Err         error
		ipv4Err         error
		resolutionDelay time.Duration
		expectError     bool
		expectedCount   int
	}{
		{
			name:            "both_succeed",
			ipv6Result:      []net.IP{net.ParseIP("2001:db8::1")},
			ipv4Result:      []net.IP{net.ParseIP("192.0.2.1")},
			resolutionDelay: 10 * time.Millisecond,
			expectError:     false,
			expectedCount:   2,
		},
		{
			name:            "ipv6_only",
			ipv6Result:      []net.IP{net.ParseIP("2001:db8::1")},
			ipv4Result:      []net.IP{},
			ipv4Err:         errors.New("no ipv4 addresses"),
			resolutionDelay: 10 * time.Millisecond,
			expectError:     false,
			expectedCount:   1,
		},
		{
			name:            "ipv4_only",
			ipv6Result:      []net.IP{},
			ipv6Err:         errors.New("no ipv6 addresses"),
			ipv4Result:      []net.IP{net.ParseIP("192.0.2.1")},
			resolutionDelay: 10 * time.Millisecond,
			expectError:     false,
			expectedCount:   1,
		},
		{
			name:            "both_fail",
			ipv6Result:      []net.IP{},
			ipv6Err:         errors.New("lookup failed"),
			ipv4Result:      []net.IP{},
			ipv4Err:         errors.New("lookup failed"),
			resolutionDelay: 10 * time.Millisecond,
			expectError:     true,
			expectedCount:   0,
		},
		{
			name: "multiple_addresses",
			ipv6Result: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("2001:db8::2"),
			},
			ipv4Result: []net.IP{
				net.ParseIP("192.0.2.1"),
				net.ParseIP("192.0.2.2"),
			},
			resolutionDelay: 10 * time.Millisecond,
			expectError:     false,
			expectedCount:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookupFunc := func(ctx context.Context, network, host string) ([]net.IP, bool, error) {
				switch network {
				case "ip6":
					return tt.ipv6Result, true, tt.ipv6Err
				case "ip4":
					return tt.ipv4Result, true, tt.ipv4Err
				default:
					return nil, false, errors.New("unsupported network")
				}
			}

			metrics := NewMetrics(true, false, nil)
			ips, _, err := ConcurrentDNSLookup(context.Background(), "example.com", lookupFunc, tt.resolutionDelay, metrics)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(ips) != tt.expectedCount {
				t.Errorf("got %d addresses, want %d", len(ips), tt.expectedCount)
			}
		})
	}
}

func TestConcurrentDNSLookup_ResolutionDelay(t *testing.T) {
	ipv6Called := false
	ipv4Called := false
	var ipv4CallTime time.Time

	lookupFunc := func(ctx context.Context, network, host string) ([]net.IP, bool, error) {
		switch network {
		case "ip6":
			ipv6Called = true
			return []net.IP{net.ParseIP("2001:db8::1")}, true, nil
		case "ip4":
			ipv4Called = true
			ipv4CallTime = time.Now()
			return []net.IP{net.ParseIP("192.0.2.1")}, true, nil
		}
		return nil, false, errors.New("unsupported")
	}

	startTime := time.Now()
	resolutionDelay := 50 * time.Millisecond

	_, _, err := ConcurrentDNSLookup(context.Background(), "example.com", lookupFunc, resolutionDelay, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ipv6Called {
		t.Error("IPv6 lookup was not called")
	}

	if !ipv4Called {
		t.Error("IPv4 lookup was not called")
	}

	ipv4Delay := ipv4CallTime.Sub(startTime)
	if ipv4Delay < resolutionDelay {
		t.Errorf("IPv4 lookup called too early: %v (expected >= %v)", ipv4Delay, resolutionDelay)
	}
}

func TestConcurrentDNSLookup_ContextCancellation(t *testing.T) {
	lookupFunc := func(ctx context.Context, network, host string) ([]net.IP, bool, error) {
		time.Sleep(100 * time.Millisecond)
		return []net.IP{net.ParseIP("2001:db8::1")}, true, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := ConcurrentDNSLookup(ctx, "example.com", lookupFunc, 50*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
