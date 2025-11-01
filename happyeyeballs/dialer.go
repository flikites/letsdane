package happyeyeballs

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"
)

type DialResult struct {
	Conn   net.Conn
	IP     net.IP
	Family int
	Error  error
}

type TLSDialResult struct {
	Conn   *tls.Conn
	IP     net.IP
	Family int
	Error  error
}

type Dialer struct {
	NetDialer net.Dialer
	Config    *Config
	Metrics   *Metrics
}

func NewDialer(netDialer net.Dialer, config *Config, metrics *Metrics) *Dialer {
	return &Dialer{
		NetDialer: netDialer,
		Config:    config,
		Metrics:   metrics,
	}
}

func (d *Dialer) DialHappyEyeballs(ctx context.Context, network, host, port string, ips []net.IP) (net.Conn, error) {
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses to dial")
	}

	if len(ips) == 1 {
		return d.dialSingle(ctx, network, host, port, ips[0])
	}

	sortedIPs := SortAndInterleaveAddresses(ips)

	resultChan := make(chan DialResult, len(sortedIPs))
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	startTime := time.Now()

	for i, ip := range sortedIPs {
		if i > 0 {
			delay := d.Config.ConnectionDelay
			if d.Config.VerboseLogging {
				familyStr := "IPv4"
				if getAddressFamily(ip) == familyIPv6 {
					familyStr = "IPv6"
				}
				fmt.Printf("[DEBUG] Happy Eyeballs: Delaying connection attempt to %s (%s) by %v\n",
					ip, familyStr, delay)
			}

			timer := time.NewTimer(delay)
			select {
			case <-cancelCtx.Done():
				timer.Stop()
				break
			case <-timer.C:
			}
		}

		wg.Add(1)
		go func(ip net.IP, attemptNum int) {
			defer wg.Done()
			d.attemptConnection(cancelCtx, network, host, port, ip, attemptNum, startTime, resultChan)
		}(ip, i)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var firstError error
	for result := range resultChan {
		if result.Error == nil {
			cancel()

			if d.Metrics != nil {
				d.Metrics.RecordConnectionAttempt(ConnectionAttempt{
					Host:      host,
					IP:        result.IP,
					Family:    result.Family,
					StartTime: startTime,
					EndTime:   time.Now(),
					Success:   true,
					Winner:    true,
				})
			}

			return result.Conn, nil
		}

		if firstError == nil {
			firstError = result.Error
		}
	}

	if firstError != nil {
		return nil, firstError
	}

	return nil, fmt.Errorf("all connection attempts failed")
}

func (d *Dialer) attemptConnection(ctx context.Context, network, host, port string, ip net.IP, attemptNum int, startTime time.Time, resultChan chan<- DialResult) {
	attemptStart := time.Now()
	family := getAddressFamily(ip)

	addr := net.JoinHostPort(ip.String(), port)
	conn, err := d.NetDialer.DialContext(ctx, network, addr)

	attemptEnd := time.Now()

	result := DialResult{
		Conn:   conn,
		IP:     ip,
		Family: family,
		Error:  err,
	}

	if d.Metrics != nil {
		d.Metrics.RecordConnectionAttempt(ConnectionAttempt{
			Host:      host,
			IP:        ip,
			Family:    family,
			StartTime: attemptStart,
			EndTime:   attemptEnd,
			Success:   err == nil,
			Error:     err,
			Winner:    false,
		})
	}

	select {
	case resultChan <- result:
	case <-ctx.Done():
		if conn != nil {
			conn.Close()
		}
	}
}

func (d *Dialer) dialSingle(ctx context.Context, network, host, port string, ip net.IP) (net.Conn, error) {
	startTime := time.Now()
	addr := net.JoinHostPort(ip.String(), port)
	conn, err := d.NetDialer.DialContext(ctx, network, addr)
	endTime := time.Now()

	if d.Metrics != nil {
		d.Metrics.RecordConnectionAttempt(ConnectionAttempt{
			Host:      host,
			IP:        ip,
			Family:    getAddressFamily(ip),
			StartTime: startTime,
			EndTime:   endTime,
			Success:   err == nil,
			Error:     err,
			Winner:    err == nil,
		})
	}

	return conn, err
}

func (d *Dialer) DialTLSHappyEyeballs(ctx context.Context, network, host, port string, ips []net.IP, tlsConfig *tls.Config) (*tls.Conn, error) {
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses to dial")
	}

	if len(ips) == 1 {
		return d.dialTLSSingle(ctx, network, host, port, ips[0], tlsConfig)
	}

	sortedIPs := SortAndInterleaveAddresses(ips)

	resultChan := make(chan TLSDialResult, len(sortedIPs))
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	startTime := time.Now()

	tlsDialer := &tls.Dialer{
		NetDialer: &d.NetDialer,
		Config:    tlsConfig,
	}

	for i, ip := range sortedIPs {
		if i > 0 {
			delay := d.Config.ConnectionDelay
			if d.Config.VerboseLogging {
				familyStr := "IPv4"
				if getAddressFamily(ip) == familyIPv6 {
					familyStr = "IPv6"
				}
				fmt.Printf("[DEBUG] Happy Eyeballs: Delaying TLS connection attempt to %s (%s) by %v\n",
					ip, familyStr, delay)
			}

			timer := time.NewTimer(delay)
			select {
			case <-cancelCtx.Done():
				timer.Stop()
				break
			case <-timer.C:
			}
		}

		wg.Add(1)
		go func(ip net.IP, attemptNum int) {
			defer wg.Done()
			d.attemptTLSConnection(cancelCtx, tlsDialer, network, host, port, ip, attemptNum, startTime, resultChan)
		}(ip, i)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var firstError error
	for result := range resultChan {
		if result.Error == nil {
			cancel()

			if d.Metrics != nil {
				d.Metrics.RecordConnectionAttempt(ConnectionAttempt{
					Host:      host,
					IP:        result.IP,
					Family:    result.Family,
					StartTime: startTime,
					EndTime:   time.Now(),
					Success:   true,
					Winner:    true,
				})
			}

			return result.Conn, nil
		}

		if firstError == nil {
			firstError = result.Error
		}
	}

	if firstError != nil {
		return nil, firstError
	}

	return nil, fmt.Errorf("all TLS connection attempts failed")
}

func (d *Dialer) attemptTLSConnection(ctx context.Context, tlsDialer *tls.Dialer, network, host, port string, ip net.IP, attemptNum int, startTime time.Time, resultChan chan<- TLSDialResult) {
	attemptStart := time.Now()
	family := getAddressFamily(ip)

	addr := net.JoinHostPort(ip.String(), port)
	conn, err := tlsDialer.DialContext(ctx, network, addr)

	attemptEnd := time.Now()

	var tlsConn *tls.Conn
	if conn != nil {
		tlsConn = conn.(*tls.Conn)
	}

	result := TLSDialResult{
		Conn:   tlsConn,
		IP:     ip,
		Family: family,
		Error:  err,
	}

	if d.Metrics != nil {
		d.Metrics.RecordConnectionAttempt(ConnectionAttempt{
			Host:      host,
			IP:        ip,
			Family:    family,
			StartTime: attemptStart,
			EndTime:   attemptEnd,
			Success:   err == nil,
			Error:     err,
			Winner:    false,
		})
	}

	select {
	case resultChan <- result:
	case <-ctx.Done():
		if tlsConn != nil {
			tlsConn.Close()
		}
	}
}

func (d *Dialer) dialTLSSingle(ctx context.Context, network, host, port string, ip net.IP, tlsConfig *tls.Config) (*tls.Conn, error) {
	startTime := time.Now()

	tlsDialer := &tls.Dialer{
		NetDialer: &d.NetDialer,
		Config:    tlsConfig,
	}

	addr := net.JoinHostPort(ip.String(), port)
	conn, err := tlsDialer.DialContext(ctx, network, addr)
	endTime := time.Now()

	var tlsConn *tls.Conn
	if conn != nil {
		tlsConn = conn.(*tls.Conn)
	}

	if d.Metrics != nil {
		d.Metrics.RecordConnectionAttempt(ConnectionAttempt{
			Host:      host,
			IP:        ip,
			Family:    getAddressFamily(ip),
			StartTime: startTime,
			EndTime:   endTime,
			Success:   err == nil,
			Error:     err,
			Winner:    err == nil,
		})
	}

	return tlsConn, err
}
