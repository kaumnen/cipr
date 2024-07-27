package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeparateFilters(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Three items",
			input:    "one,two,three",
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "Two items",
			input:    "one,two",
			expected: []string{"one", "two", "*"},
		},
		{
			name:     "Single item",
			input:    "single,",
			expected: []string{"single", "*", "*"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := separateFilters(tc.input)
			assert.Equal(t, tc.expected, result, "Test case '%s' failed: input=%s, expected=%v, got=%v",
				tc.name, tc.input, tc.expected, result)
		})
	}
}

func TestFiltrateIPRanges(t *testing.T) {
	testCases := []struct {
		name        string
		ipType      string
		filterSlice []string
		expected    []IPPrefix
	}{
		{
			name:        "ipv4 - filters: us-east-1 region, EBS service",
			ipType:      "ipv4",
			filterSlice: []string{"us-east-1", "EBS", "*"},
			expected: []IPPrefix{
				IPv4Prefix{IPAddress: "44.192.140.112/28", Region: "us-east-1", Service: "EBS", NetworkBorderGroup: "us-east-1"},
				IPv4Prefix{IPAddress: "44.192.140.128/29", Region: "us-east-1", Service: "EBS", NetworkBorderGroup: "us-east-1"},
				IPv4Prefix{IPAddress: "44.222.159.166/31", Region: "us-east-1", Service: "EBS", NetworkBorderGroup: "us-east-1"},
				IPv4Prefix{IPAddress: "44.222.159.176/28", Region: "us-east-1", Service: "EBS", NetworkBorderGroup: "us-east-1"},
			},
		},
		{
			name:        "ipv6 - filters: us-east-1 region, EBS service, us-east-1 network border group",
			ipType:      "ipv6",
			filterSlice: []string{"eu-central-1", "S3", "eu-central-1"},
			expected: []IPPrefix{
				IPv6Prefix{IPv6Address: "2a05:d070:4000::/40", Region: "eu-central-1", Service: "S3", NetworkBorderGroup: "eu-central-1"},
				IPv6Prefix{IPv6Address: "2a05:d079:4000::/40", Region: "eu-central-1", Service: "S3", NetworkBorderGroup: "eu-central-1"},
				IPv6Prefix{IPv6Address: "2a05:d034:4000::/40", Region: "eu-central-1", Service: "S3", NetworkBorderGroup: "eu-central-1"},
				IPv6Prefix{IPv6Address: "2a05:d07a:4000::/40", Region: "eu-central-1", Service: "S3", NetworkBorderGroup: "eu-central-1"},
				IPv6Prefix{IPv6Address: "2a05:d078:4000::/40", Region: "eu-central-1", Service: "S3", NetworkBorderGroup: "eu-central-1"},
				IPv6Prefix{IPv6Address: "2a05:d050:4000::/40", Region: "eu-central-1", Service: "S3", NetworkBorderGroup: "eu-central-1"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawData := mockGetReq()

			result := filtrateIPRanges(rawData, tc.ipType, tc.filterSlice)

			assert.Equal(t, len(tc.expected), len(result), "Test case '%s' failed: expected %d items, got %d",
				tc.name, len(tc.expected), len(result))

			for i, expectedIP := range tc.expected {
				assert.Equal(t, expectedIP.GetIPAddress(), result[i].GetIPAddress(), "IP address mismatch at index %d", i)
				assert.Equal(t, expectedIP.GetRegion(), result[i].GetRegion(), "Region mismatch at index %d", i)
				assert.Equal(t, expectedIP.GetService(), result[i].GetService(), "Service mismatch at index %d", i)
				assert.Equal(t, expectedIP.GetNetworkBorderGroup(), result[i].GetNetworkBorderGroup(), "Network border group mismatch at index %d", i)
			}
		})
	}
}
