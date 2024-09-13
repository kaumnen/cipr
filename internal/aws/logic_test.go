package aws

import (
	"bytes"
	"io"
	"os"
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

func TestPrintIPRanges(t *testing.T) {
	testCases := []struct {
		name           string
		IPranges       []IPPrefix
		verbosity      string
		expectedOutput string
	}{
		{
			name:           "Empty IP range list",
			IPranges:       []IPPrefix{},
			verbosity:      "none",
			expectedOutput: "No IP ranges to display.\n",
		},
		{
			name: "IPv4 - Verbosity None",
			IPranges: []IPPrefix{
				IPv4Prefix{IPAddress: "52.216.160.0/23", Region: "us-east-1", Service: "S3", NetworkBorderGroup: "us-east-1"},
				IPv4Prefix{IPAddress: "54.240.224.0/21", Region: "us-west-1", Service: "dynamodb", NetworkBorderGroup: "us-west-1"},
			},
			verbosity:      "none",
			expectedOutput: "52.216.160.0/23\n54.240.224.0/21\n",
		},
		{
			name: "IPv6 - Verbosity Mini",
			IPranges: []IPPrefix{
				IPv6Prefix{IPv6Address: "2600:1f18:400:d000::/56", Region: "us-east-1", Service: "EC2", NetworkBorderGroup: "us-east-1"},
			},
			verbosity:      "mini",
			expectedOutput: "2600:1f18:400:d000::/56,us-east-1,EC2,us-east-1\n",
		},
		{
			name: "Mixed IPv4/IPv6 - Verbosity Full",
			IPranges: []IPPrefix{
				IPv4Prefix{IPAddress: "3.5.140.0/22", Region: "us-east-1", Service: "AMAZON", NetworkBorderGroup: "us-east-1"},
				IPv6Prefix{IPv6Address: "2600:1f18:480:d000::/56", Region: "us-west-2", Service: "ROUTE53", NetworkBorderGroup: "us-west-2"},
			},
			verbosity:      "full",
			expectedOutput: "IP Prefix: 3.5.140.0/22, Region: us-east-1, Service: AMAZON, Network Border Group: us-east-1\nIP Prefix: 2600:1f18:480:d000::/56, Region: us-west-2, Service: ROUTE53, Network Border Group: us-west-2\n",
		},
		{
			name: "Invalid Verbosity - Defaults to None",
			IPranges: []IPPrefix{
				IPv4Prefix{IPAddress: "15.181.152.0/22", Region: "eu-west-1", Service: "CLOUDFRONT", NetworkBorderGroup: "eu-west-1"},
			},
			verbosity:      "invalid",
			expectedOutput: "15.181.152.0/22\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printIPRanges(tc.IPranges, tc.verbosity)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			assert.Equal(t, tc.expectedOutput, output, "Output mismatch for test case: %s", tc.name)
		})
	}
}
