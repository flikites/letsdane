package happyeyeballs

import (
	"net"
	"testing"
)

func TestGetAddressFamily(t *testing.T) {
	tests := []struct {
		name   string
		ip     net.IP
		family int
	}{
		{"ipv4", net.ParseIP("192.0.2.1"), familyIPv4},
		{"ipv6", net.ParseIP("2001:db8::1"), familyIPv6},
		{"ipv4_mapped", net.ParseIP("::ffff:192.0.2.1"), familyIPv4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			family := getAddressFamily(tt.ip)
			if family != tt.family {
				t.Errorf("got %d, want %d", family, tt.family)
			}
		})
	}
}

func TestInterleaveAddresses(t *testing.T) {
	tests := []struct {
		name     string
		input    []net.IP
		expected []net.IP
	}{
		{
			name:     "empty",
			input:    []net.IP{},
			expected: []net.IP{},
		},
		{
			name:     "single_ipv4",
			input:    []net.IP{net.ParseIP("192.0.2.1")},
			expected: []net.IP{net.ParseIP("192.0.2.1")},
		},
		{
			name:     "single_ipv6",
			input:    []net.IP{net.ParseIP("2001:db8::1")},
			expected: []net.IP{net.ParseIP("2001:db8::1")},
		},
		{
			name: "interleaved",
			input: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("2001:db8::2"),
				net.ParseIP("192.0.2.1"),
				net.ParseIP("192.0.2.2"),
			},
			expected: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("192.0.2.1"),
				net.ParseIP("2001:db8::2"),
				net.ParseIP("192.0.2.2"),
			},
		},
		{
			name: "more_ipv6_than_ipv4",
			input: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("2001:db8::2"),
				net.ParseIP("2001:db8::3"),
				net.ParseIP("192.0.2.1"),
			},
			expected: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("192.0.2.1"),
				net.ParseIP("2001:db8::2"),
				net.ParseIP("2001:db8::3"),
			},
		},
		{
			name: "only_ipv6",
			input: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("2001:db8::2"),
			},
			expected: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("2001:db8::2"),
			},
		},
		{
			name: "only_ipv4",
			input: []net.IP{
				net.ParseIP("192.0.2.1"),
				net.ParseIP("192.0.2.2"),
			},
			expected: []net.IP{
				net.ParseIP("192.0.2.1"),
				net.ParseIP("192.0.2.2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InterleaveAddresses(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if !result[i].Equal(tt.expected[i]) {
					t.Errorf("at index %d: got %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSortAndInterleaveAddresses(t *testing.T) {
	tests := []struct {
		name     string
		input    []net.IP
		expected []net.IP
	}{
		{
			name: "mixed_order",
			input: []net.IP{
				net.ParseIP("192.0.2.1"),
				net.ParseIP("2001:db8::1"),
				net.ParseIP("192.0.2.2"),
				net.ParseIP("2001:db8::2"),
			},
			expected: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("192.0.2.1"),
				net.ParseIP("2001:db8::2"),
				net.ParseIP("192.0.2.2"),
			},
		},
		{
			name: "ipv4_first",
			input: []net.IP{
				net.ParseIP("192.0.2.1"),
				net.ParseIP("192.0.2.2"),
				net.ParseIP("2001:db8::1"),
			},
			expected: []net.IP{
				net.ParseIP("2001:db8::1"),
				net.ParseIP("192.0.2.1"),
				net.ParseIP("192.0.2.2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SortAndInterleaveAddresses(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if !result[i].Equal(tt.expected[i]) {
					t.Errorf("at index %d: got %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
