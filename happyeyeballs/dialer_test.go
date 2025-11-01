package happyeyeballs

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDialHappyEyeballs_SingleAddress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	ip, port, _ := net.SplitHostPort(addr)

	config := &Config{
		Enabled:         true,
		ResolutionDelay: 50 * time.Millisecond,
		ConnectionDelay: 250 * time.Millisecond,
		MetricsEnabled:  false,
	}

	dialer := NewDialer(net.Dialer{Timeout: 5 * time.Second}, config, nil)

	conn, err := dialer.DialHappyEyeballs(context.Background(), "tcp", "example.com", port, []net.IP{net.ParseIP(ip)})
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
}

func TestDialHappyEyeballs_MultipleAddresses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	ip, port, _ := net.SplitHostPort(addr)

	config := &Config{
		Enabled:         true,
		ResolutionDelay: 10 * time.Millisecond,
		ConnectionDelay: 50 * time.Millisecond,
		MetricsEnabled:  false,
	}

	dialer := NewDialer(net.Dialer{Timeout: 5 * time.Second}, config, nil)

	ips := []net.IP{
		net.ParseIP("255.255.255.255"),
		net.ParseIP(ip),
	}

	conn, err := dialer.DialHappyEyeballs(context.Background(), "tcp", "example.com", port, ips)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
}

func TestDialHappyEyeballs_NoAddresses(t *testing.T) {
	config := &Config{
		Enabled:         true,
		ResolutionDelay: 50 * time.Millisecond,
		ConnectionDelay: 250 * time.Millisecond,
		MetricsEnabled:  false,
	}

	dialer := NewDialer(net.Dialer{Timeout: 5 * time.Second}, config, nil)

	_, err := dialer.DialHappyEyeballs(context.Background(), "tcp", "example.com", "80", []net.IP{})
	if err == nil {
		t.Fatal("expected error for empty address list")
	}
}

func TestDialHappyEyeballs_AllFailed(t *testing.T) {
	config := &Config{
		Enabled:         true,
		ResolutionDelay: 10 * time.Millisecond,
		ConnectionDelay: 50 * time.Millisecond,
		MetricsEnabled:  false,
	}

	dialer := NewDialer(net.Dialer{Timeout: 1 * time.Second}, config, nil)

	ips := []net.IP{
		net.ParseIP("255.255.255.254"),
		net.ParseIP("255.255.255.253"),
	}

	_, err := dialer.DialHappyEyeballs(context.Background(), "tcp", "example.com", "12345", ips)
	if err == nil {
		t.Fatal("expected error when all connections fail")
	}
}

func TestDialHappyEyeballs_ContextCancellation(t *testing.T) {
	config := &Config{
		Enabled:         true,
		ResolutionDelay: 50 * time.Millisecond,
		ConnectionDelay: 250 * time.Millisecond,
		MetricsEnabled:  false,
	}

	dialer := NewDialer(net.Dialer{Timeout: 10 * time.Second}, config, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ips := []net.IP{
		net.ParseIP("192.0.2.1"),
	}

	_, err := dialer.DialHappyEyeballs(ctx, "tcp", "example.com", "80", ips)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestDialHappyEyeballs_WithMetrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	ip, port, _ := net.SplitHostPort(addr)

	config := &Config{
		Enabled:         true,
		ResolutionDelay: 10 * time.Millisecond,
		ConnectionDelay: 50 * time.Millisecond,
		MetricsEnabled:  true,
	}

	metrics := NewMetrics(true, false, nil)
	dialer := NewDialer(net.Dialer{Timeout: 5 * time.Second}, config, metrics)

	ips := []net.IP{net.ParseIP(ip)}

	conn, err := dialer.DialHappyEyeballs(context.Background(), "tcp", "example.com", port, ips)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	summary := metrics.GetSummary()
	if !strings.Contains(summary, "attempts") {
		t.Errorf("expected metrics summary to contain 'attempts', got: %s", summary)
	}
}
