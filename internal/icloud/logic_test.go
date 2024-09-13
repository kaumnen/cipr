package icloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiltrateIPRanges(t *testing.T) {
	ipRanges := []IPRange{
		{"2a02:26f7:f6f9:800::/54", "US", "US-NY", "New York"},
		{"2a02:26f7:f6f9:a06a::/64", "US", "US-NY", "New York"},
		{"2a02:26f7:f6fc:800::/54", "US", "US-NY", "New York"},
		{"2a02:26f7:f6fc:a06a::/64", "US", "US-NY", "New York"},
		{"2606:54c0:a620::/45", "US", "US-NY", "New York"},
		{"2a09:bac2:a620::/45", "US", "US-NY", "New York"},
		{"2a09:bac3:a620::/45", "US", "US-NY", "New York"},
		{"104.28.129.23/32", "DE", "DE-BE", "Berlin"},
		{"104.28.129.24/32", "DE", "DE-BE", "Berlin"},
		{"104.28.129.25/32", "DE", "DE-BE", "Berlin"},
		{"104.28.129.26/32", "DE", "DE-BE", "Berlin"},
		{"140.248.17.50/31", "DE", "DE-BE", "Berlin"},
		{"140.248.34.44/31", "DE", "DE-BE", "Berlin"},
		{"140.248.36.56/31", "DE", "DE-BE", "Berlin"},
		{"146.75.166.14/31", "DE", "DE-BE", "Berlin"},
		{"146.75.169.44/31", "DE", "DE-BE", "Berlin"},
		{"172.224.240.128/27", "JP", "JP-13", "Tokyo"},
		{"172.225.46.64/26", "JP", "JP-13", "Tokyo"},
		{"172.225.46.208/28", "JP", "JP-13", "Tokyo"},
		{"2a04:4e41:0030:0007::/64", "JP", "JP-13", "Tokyo"},
		{"2a04:4e41:0035:0006::/64", "JP", "JP-13", "Tokyo"},
		{"2a04:4e41:0064:000b::/64", "JP", "JP-13", "Tokyo"},
	}

	testCases := []struct {
		name            string
		ipType          string
		filterCountries []string
		filterStates    []string
		filterCities    []string
		expected        []IPRange
	}{
		{
			name:            "IPv6 US-NY-New York",
			ipType:          "ipv6",
			filterCountries: []string{"US"},
			filterStates:    []string{"US-NY"},
			filterCities:    []string{"New York"},
			expected: []IPRange{
				{"2a02:26f7:f6f9:800::/54", "US", "US-NY", "New York"},
				{"2a02:26f7:f6f9:a06a::/64", "US", "US-NY", "New York"},
				{"2a02:26f7:f6fc:800::/54", "US", "US-NY", "New York"},
				{"2a02:26f7:f6fc:a06a::/64", "US", "US-NY", "New York"},
				{"2606:54c0:a620::/45", "US", "US-NY", "New York"},
				{"2a09:bac2:a620::/45", "US", "US-NY", "New York"},
				{"2a09:bac3:a620::/45", "US", "US-NY", "New York"},
			},
		},
		{
			name:            "IPv4 DE-BE-Berlin",
			ipType:          "ipv4",
			filterCountries: []string{"DE"},
			filterStates:    []string{"DE-BE"},
			filterCities:    []string{"Berlin"},
			expected: []IPRange{
				{"104.28.129.23/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.24/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.25/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.26/32", "DE", "DE-BE", "Berlin"},
				{"140.248.17.50/31", "DE", "DE-BE", "Berlin"},
				{"140.248.34.44/31", "DE", "DE-BE", "Berlin"},
				{"140.248.36.56/31", "DE", "DE-BE", "Berlin"},
				{"146.75.166.14/31", "DE", "DE-BE", "Berlin"},
				{"146.75.169.44/31", "DE", "DE-BE", "Berlin"},
			},
		},
		{
			name:            "All IP types in Tokyo",
			ipType:          "ipv4",
			filterCountries: []string{},
			filterStates:    []string{},
			filterCities:    []string{"Tokyo"},
			expected: []IPRange{
				{"172.224.240.128/27", "JP", "JP-13", "Tokyo"},
				{"172.225.46.64/26", "JP", "JP-13", "Tokyo"},
				{"172.225.46.208/28", "JP", "JP-13", "Tokyo"},
			},
		},
		{
			name:            "IPv4 DE-BE without city filter",
			ipType:          "ipv4",
			filterCountries: []string{"DE"},
			filterStates:    []string{"DE-BE"},
			filterCities:    []string{},
			expected: []IPRange{
				{"104.28.129.23/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.24/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.25/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.26/32", "DE", "DE-BE", "Berlin"},
				{"140.248.17.50/31", "DE", "DE-BE", "Berlin"},
				{"140.248.34.44/31", "DE", "DE-BE", "Berlin"},
				{"140.248.36.56/31", "DE", "DE-BE", "Berlin"},
				{"146.75.166.14/31", "DE", "DE-BE", "Berlin"},
				{"146.75.169.44/31", "DE", "DE-BE", "Berlin"},
			},
		},
		{
			name:            "IPv6 US without state and city filter",
			ipType:          "ipv6",
			filterCountries: []string{"US"},
			filterStates:    []string{},
			filterCities:    []string{},
			expected: []IPRange{
				{"2a02:26f7:f6f9:800::/54", "US", "US-NY", "New York"},
				{"2a02:26f7:f6f9:a06a::/64", "US", "US-NY", "New York"},
				{"2a02:26f7:f6fc:800::/54", "US", "US-NY", "New York"},
				{"2a02:26f7:f6fc:a06a::/64", "US", "US-NY", "New York"},
				{"2606:54c0:a620::/45", "US", "US-NY", "New York"},
				{"2a09:bac2:a620::/45", "US", "US-NY", "New York"},
				{"2a09:bac3:a620::/45", "US", "US-NY", "New York"},
			},
		},
		{
			name:            "No filters",
			ipType:          "ipv6",
			filterCountries: []string{},
			filterStates:    []string{},
			filterCities:    []string{},
			expected: []IPRange{
				{"2a02:26f7:f6f9:800::/54", "US", "US-NY", "New York"},
				{"2a02:26f7:f6f9:a06a::/64", "US", "US-NY", "New York"},
				{"2a02:26f7:f6fc:800::/54", "US", "US-NY", "New York"},
				{"2a02:26f7:f6fc:a06a::/64", "US", "US-NY", "New York"},
				{"2606:54c0:a620::/45", "US", "US-NY", "New York"},
				{"2a09:bac2:a620::/45", "US", "US-NY", "New York"},
				{"2a09:bac3:a620::/45", "US", "US-NY", "New York"},
				{"2a04:4e41:0030:0007::/64", "JP", "JP-13", "Tokyo"},
				{"2a04:4e41:0035:0006::/64", "JP", "JP-13", "Tokyo"},
				{"2a04:4e41:0064:000b::/64", "JP", "JP-13", "Tokyo"},
			},
		},
		{
			name:            "City Filter Tokyo",
			ipType:          "ipv4",
			filterCountries: []string{},
			filterStates:    []string{},
			filterCities:    []string{"Tokyo"},
			expected: []IPRange{
				{"172.224.240.128/27", "JP", "JP-13", "Tokyo"},
				{"172.225.46.64/26", "JP", "JP-13", "Tokyo"},
				{"172.225.46.208/28", "JP", "JP-13", "Tokyo"},
			},
		},
		{
			name:            "Multiple Countries and Cities",
			ipType:          "ipv4",
			filterCountries: []string{"US", "DE", "JP"},
			filterStates:    []string{"US-NY", "DE-BE", "JP-13"},
			filterCities:    []string{"New York", "Berlin", "Tokyo"},
			expected: []IPRange{
				{"104.28.129.23/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.24/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.25/32", "DE", "DE-BE", "Berlin"},
				{"104.28.129.26/32", "DE", "DE-BE", "Berlin"},
				{"140.248.17.50/31", "DE", "DE-BE", "Berlin"},
				{"140.248.34.44/31", "DE", "DE-BE", "Berlin"},
				{"140.248.36.56/31", "DE", "DE-BE", "Berlin"},
				{"146.75.166.14/31", "DE", "DE-BE", "Berlin"},
				{"146.75.169.44/31", "DE", "DE-BE", "Berlin"},
				{"172.224.240.128/27", "JP", "JP-13", "Tokyo"},
				{"172.225.46.64/26", "JP", "JP-13", "Tokyo"},
				{"172.225.46.208/28", "JP", "JP-13", "Tokyo"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := filtrateIPRanges(ipRanges, tc.ipType, tc.filterCountries, tc.filterStates, tc.filterCities)
			assert.Equal(t, tc.expected, result, "Test case '%s' failed", tc.name)
		})
	}
}
