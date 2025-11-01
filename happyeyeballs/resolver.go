package happyeyeballs

import (
	"context"
	"net"
	"time"
)

type DNSLookupFunc func(ctx context.Context, network, host string) ([]net.IP, bool, error)

func ConcurrentDNSLookup(ctx context.Context, host string, lookupFunc DNSLookupFunc, resolutionDelay time.Duration, metrics *Metrics) ([]net.IP, bool, error) {
	type result struct {
		ips    []net.IP
		secure bool
		err    error
		family int
	}

	ipv6Chan := make(chan result, 1)
	ipv4Chan := make(chan result, 1)

	go func() {
		startTime := time.Now()
		ips, secure, err := lookupFunc(ctx, "ip6", host)
		endTime := time.Now()

		if metrics != nil {
			metrics.RecordDNSResolution(DNSResolution{
				Host:         host,
				Family:       familyIPv6,
				StartTime:    startTime,
				EndTime:      endTime,
				AddressCount: len(ips),
				Success:      err == nil,
				Error:        err,
			})
		}

		ipv6Chan <- result{ips: ips, secure: secure, err: err, family: familyIPv6}
	}()

	timer := time.NewTimer(resolutionDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case <-timer.C:
	}

	go func() {
		startTime := time.Now()
		ips, secure, err := lookupFunc(ctx, "ip4", host)
		endTime := time.Now()

		if metrics != nil {
			metrics.RecordDNSResolution(DNSResolution{
				Host:         host,
				Family:       familyIPv4,
				StartTime:    startTime,
				EndTime:      endTime,
				AddressCount: len(ips),
				Success:      err == nil,
				Error:        err,
			})
		}

		ipv4Chan <- result{ips: ips, secure: secure, err: err, family: familyIPv4}
	}()

	var allIPs []net.IP
	var secure bool = true
	var lastErr error
	resultsReceived := 0

	for resultsReceived < 2 {
		select {
		case res := <-ipv6Chan:
			resultsReceived++
			if res.err == nil {
				allIPs = append(allIPs, res.ips...)
			} else {
				lastErr = res.err
			}
			secure = secure && res.secure

		case res := <-ipv4Chan:
			resultsReceived++
			if res.err == nil {
				allIPs = append(allIPs, res.ips...)
			} else {
				lastErr = res.err
			}
			secure = secure && res.secure

		case <-ctx.Done():
			return nil, false, ctx.Err()
		}
	}

	if len(allIPs) == 0 {
		if lastErr != nil {
			return nil, false, lastErr
		}
		return nil, false, nil
	}

	return allIPs, secure, nil
}
