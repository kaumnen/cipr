package icloud

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestParseRecords(t *testing.T) {
	t.Run("empty input yields no records", func(t *testing.T) {
		got, err := parseRecords("")
		assert.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("trims whitespace and tolerates partial columns", func(t *testing.T) {
		input := "172.224.224.0/27, GB , GB-EN , London\n2a02:26f7::/64,US,US-NY\n3.3.3.0/24,US\n4.4.4.0/24\n"
		got, err := parseRecords(input)
		require.NoError(t, err)
		assert.Equal(t, []IPRange{
			{IPRange: "172.224.224.0/27", Country: "GB", Region: "GB-EN", City: "London"},
			{IPRange: "2a02:26f7::/64", Country: "US", Region: "US-NY"},
			{IPRange: "3.3.3.0/24", Country: "US"},
			{IPRange: "4.4.4.0/24"},
		}, got)
	})

	t.Run("propagates csv parse errors", func(t *testing.T) {
		_, err := parseRecords("172.224.224.0/27,\"unterminated\n")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse icloud csv")
	})

	t.Run("fixture first row matches", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("..", "testdata", "icloud.csv"))
		require.NoError(t, err)
		got, err := parseRecords(string(data))
		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Equal(t, IPRange{IPRange: "172.224.224.0/27", Country: "GB", Region: "GB-EN", City: "London"}, got[0])
	})
}

func TestGetIPRanges_FromFixture(t *testing.T) {
	cfg := Config{
		Source:    filepath.Join("..", "testdata", "icloud.csv"),
		IPType:    "ipv4",
		Filters:   Filters{Country: []string{"GB"}, City: []string{"London"}},
		Verbosity: "none",
	}

	out := captureStdout(t, func() {
		err := GetIPRanges(context.Background(), cfg)
		require.NoError(t, err)
	})

	lines := bytes.Split(bytes.TrimRight([]byte(out), "\n"), []byte("\n"))
	assert.NotEmpty(t, lines)
	for _, line := range lines {
		assert.Contains(t, string(line), "/")
	}
}

func TestGetIPRanges_SourceError(t *testing.T) {
	cfg := Config{
		Source:    "./does-not-exist/missing.csv",
		IPType:    "ipv4",
		Verbosity: "none",
	}
	err := GetIPRanges(context.Background(), cfg)
	require.Error(t, err)
}

func TestGetIPRanges_List_Countries(t *testing.T) {
	cfg := Config{
		Source:    filepath.Join("..", "testdata", "icloud.csv"),
		IPType:    "both",
		List:      "countries",
		Verbosity: "none",
	}

	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), cfg))
	})

	lines := bytes.Split(bytes.TrimRight([]byte(out), "\n"), []byte("\n"))
	require.NotEmpty(t, lines)
	seen := make(map[string]struct{})
	for _, l := range lines {
		s := string(l)
		assert.NotEmpty(t, s, "empty country surfaced")
		_, dup := seen[s]
		assert.False(t, dup, "duplicate country %q", s)
		seen[s] = struct{}{}
	}
	assert.Contains(t, seen, "GB")
}

func TestPrintListedValues(t *testing.T) {
	ranges := []IPRange{
		{Country: "GB", Region: "GB-EN", City: "London"},
		{Country: "GB", Region: "GB-EN", City: "London"},
		{Country: "JP", Region: "JP-13", City: "Tokyo"},
		{Country: "DE", Region: "", City: ""},
	}

	t.Run("countries dedupes", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(ranges, "countries"))
		})
		assert.Equal(t, "DE\nGB\nJP\n", out)
	})

	t.Run("cities drops empty", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(ranges, "cities"))
		})
		assert.Equal(t, "London\nTokyo\n", out)
	})

	t.Run("unknown dim returns error", func(t *testing.T) {
		err := printListedValues(ranges, "bogus")
		require.Error(t, err)
	})
}

func TestPrintIPRanges(t *testing.T) {
	ipRanges := []IPRange{
		{IPRange: "172.224.224.0/27", Country: "GB", Region: "GB-EN", City: "London"},
		{IPRange: "2a04:4e41:0030:0007::/64", Country: "JP", Region: "JP-13", City: "Tokyo"},
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
			expected:  "172.224.224.0/27\n2a04:4e41:0030:0007::/64\n",
		},
		{
			name:      "Verbosity mini",
			ipRanges:  ipRanges,
			verbosity: "mini",
			expected:  "172.224.224.0/27,GB,GB-EN,London\n2a04:4e41:0030:0007::/64,JP,JP-13,Tokyo\n",
		},
		{
			name:      "Verbosity full",
			ipRanges:  ipRanges,
			verbosity: "full",
			expected:  "IP Range: 172.224.224.0/27, Country: GB, Region: GB-EN, City: London\nIP Range: 2a04:4e41:0030:0007::/64, Country: JP, Region: JP-13, City: Tokyo\n",
		},
		{
			name:      "Default verbosity",
			ipRanges:  ipRanges,
			verbosity: "unknown",
			expected:  "172.224.224.0/27\n2a04:4e41:0030:0007::/64\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				printIPRanges(tt.ipRanges, tt.verbosity)
			})
			assert.Equal(t, tt.expected, out)
		})
	}
}

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
		filterRegions   []string
		filterCities    []string
		expected        []IPRange
	}{
		{
			name:            "IPv6 US-NY-New York",
			ipType:          "ipv6",
			filterCountries: []string{"US"},
			filterRegions:   []string{"US-NY"},
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
			filterRegions:   []string{"DE-BE"},
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
			filterRegions:   []string{},
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
			filterRegions:   []string{"DE-BE"},
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
			name:            "IPv6 US without Region and city filter",
			ipType:          "ipv6",
			filterCountries: []string{"US"},
			filterRegions:   []string{},
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
			filterRegions:   []string{},
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
			filterRegions:   []string{},
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
			filterRegions:   []string{"US-NY", "DE-BE", "JP-13"},
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
			config := Config{
				IPType: tc.ipType,
				Filters: Filters{
					Country: tc.filterCountries,
					Region:  tc.filterRegions,
					City:    tc.filterCities,
				},
				Verbosity: "none",
			}

			result := filtrateIPRanges(ipRanges, config)

			assert.Equal(t, tc.expected, result, "Test case '%s' failed", tc.name)
		})
	}
}
