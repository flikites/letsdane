package happyeyeballs

import (
	"net"
	"sort"
)

type addressWithFamily struct {
	IP     net.IP
	Family int
}

const (
	familyIPv6 = 6
	familyIPv4 = 4
)

func getAddressFamily(ip net.IP) int {
	if ip.To4() != nil {
		return familyIPv4
	}
	return familyIPv6
}

func InterleaveAddresses(ips []net.IP) []net.IP {
	if len(ips) == 0 {
		return ips
	}

	var ipv6Addrs []net.IP
	var ipv4Addrs []net.IP

	for _, ip := range ips {
		if getAddressFamily(ip) == familyIPv6 {
			ipv6Addrs = append(ipv6Addrs, ip)
		} else {
			ipv4Addrs = append(ipv4Addrs, ip)
		}
	}

	result := make([]net.IP, 0, len(ips))
	i6, i4 := 0, 0

	for i6 < len(ipv6Addrs) || i4 < len(ipv4Addrs) {
		if i6 < len(ipv6Addrs) {
			result = append(result, ipv6Addrs[i6])
			i6++
		}
		if i4 < len(ipv4Addrs) {
			result = append(result, ipv4Addrs[i4])
			i4++
		}
	}

	return result
}

func SortAndInterleaveAddresses(ips []net.IP) []net.IP {
	if len(ips) == 0 {
		return ips
	}

	withFamily := make([]addressWithFamily, len(ips))
	for i, ip := range ips {
		withFamily[i] = addressWithFamily{
			IP:     ip,
			Family: getAddressFamily(ip),
		}
	}

	sort.SliceStable(withFamily, func(i, j int) bool {
		return withFamily[i].Family > withFamily[j].Family
	})

	sorted := make([]net.IP, len(ips))
	for i, addr := range withFamily {
		sorted[i] = addr.IP
	}

	return InterleaveAddresses(sorted)
}
