package digitalocean

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiltrateIPRanges(t *testing.T) {
	ipRanges := []IPRange{
		{IPRange: "192.168.1.0/24", Country: "US", Region: "California", City: "San Francisco", Zip: "94107"},
		{IPRange: "2607:f8b0:4005:805::200e", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
		{IPRange: "10.0.0.0/8", Country: "US", Region: "New York", City: "New York", Zip: "10001"},
		{IPRange: "2001:4860:4860::8888", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
	}

	tests := []struct {
		name     string
		config   Config
		expected []IPRange
	}{
		{
			name: "Filter by IPv4",
			config: Config{
				IPType: "ipv4",
			},
			expected: []IPRange{
				{IPRange: "192.168.1.0/24", Country: "US", Region: "California", City: "San Francisco", Zip: "94107"},
				{IPRange: "10.0.0.0/8", Country: "US", Region: "New York", City: "New York", Zip: "10001"},
			},
		},
		{
			name: "Filter by IPv6",
			config: Config{
				IPType: "ipv6",
			},
			expected: []IPRange{
				{IPRange: "2607:f8b0:4005:805::200e", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
				{IPRange: "2001:4860:4860::8888", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
			},
		},
		{
			name: "Filter by both IPv4 and IPv6",
			config: Config{
				IPType: "both",
			},
			expected: ipRanges,
		},
		{
			name: "Filter by country",
			config: Config{
				IPType: "both",
				Filters: Filters{
					Country: []string{"US"},
				},
			},
			expected: ipRanges,
		},
		{
			name: "Filter by region",
			config: Config{
				IPType: "both",
				Filters: Filters{
					Region: []string{"California"},
				},
			},
			expected: []IPRange{
				{IPRange: "192.168.1.0/24", Country: "US", Region: "California", City: "San Francisco", Zip: "94107"},
				{IPRange: "2607:f8b0:4005:805::200e", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
				{IPRange: "2001:4860:4860::8888", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
			},
		},
		{
			name: "Filter by city",
			config: Config{
				IPType: "both",
				Filters: Filters{
					City: []string{"Mountain View"},
				},
			},
			expected: []IPRange{
				{IPRange: "2607:f8b0:4005:805::200e", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
				{IPRange: "2001:4860:4860::8888", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
			},
		},
		{
			name: "Filter by zip",
			config: Config{
				IPType: "both",
				Filters: Filters{
					Zip: []string{"94043"},
				},
			},
			expected: []IPRange{
				{IPRange: "2607:f8b0:4005:805::200e", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
				{IPRange: "2001:4860:4860::8888", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filtrateIPRanges(ipRanges, tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "Item present in slice",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "banana",
			expected: true,
		},
		{
			name:     "Item present in slice with different case",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "Banana",
			expected: true,
		},
		{
			name:     "Item not present in slice",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "grape",
			expected: false,
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			item:     "apple",
			expected: false,
		},
		{
			name:     "Empty item",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "",
			expected: false,
		},
		{
			name:     "Item present in slice with mixed case",
			slice:    []string{"apple", "banana", "Cherry"},
			item:     "cherry",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsIgnoreCase(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintIPRanges(t *testing.T) {
	ipRanges := []IPRange{
		{IPRange: "192.168.1.0/24", Country: "US", Region: "California", City: "San Francisco", Zip: "94107"},
		{IPRange: "2607:f8b0:4005:805::200e", Country: "US", Region: "California", City: "Mountain View", Zip: "94043"},
	}

	tests := []struct {
		name      string
		ipRanges  []IPRange
		verbosity string
		expected  string
	}{
		{
			name:      "No IP ranges",
			ipRanges:  []IPRange{},
			verbosity: "none",
			expected:  "No IP ranges to display.\n",
		},
		{
			name:      "Verbosity none",
			ipRanges:  ipRanges,
			verbosity: "none",
			expected:  "192.168.1.0/24\n2607:f8b0:4005:805::200e\n",
		},
		{
			name:      "Verbosity mini",
			ipRanges:  ipRanges,
			verbosity: "mini",
			expected:  "192.168.1.0/24,US,California,San Francisco\n2607:f8b0:4005:805::200e,US,California,Mountain View\n",
		},
		{
			name:      "Verbosity full",
			ipRanges:  ipRanges,
			verbosity: "full",
			expected:  "IP Range: 192.168.1.0/24, Country: US, Region: California, City: San Francisco, ZIP: 94107\nIP Range: 2607:f8b0:4005:805::200e, Country: US, Region: California, City: Mountain View, ZIP: 94043\n",
		},
		{
			name:      "Default verbosity",
			ipRanges:  ipRanges,
			verbosity: "unknown",
			expected:  "192.168.1.0/24\n2607:f8b0:4005:805::200e\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, _ := os.Pipe()
			oldStdout := os.Stdout
			os.Stdout = w

			printIPRanges(tt.ipRanges, tt.verbosity)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			assert.Equal(t, tt.expected, buf.String())
		})
	}
}
