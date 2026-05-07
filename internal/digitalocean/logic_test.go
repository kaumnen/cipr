package digitalocean

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
		input := "1.1.1.0/24, US , CA , San Francisco , 94107\n2.2.2.0/24,US,CA,Oakland\n3.3.3.0/24,US,CA\n4.4.4.0/24,US\n5.5.5.0/24\n"
		got, err := parseRecords(input)
		require.NoError(t, err)
		assert.Equal(t, []IPRange{
			{IPRange: "1.1.1.0/24", Country: "US", Region: "CA", City: "San Francisco", Zip: "94107"},
			{IPRange: "2.2.2.0/24", Country: "US", Region: "CA", City: "Oakland"},
			{IPRange: "3.3.3.0/24", Country: "US", Region: "CA"},
			{IPRange: "4.4.4.0/24", Country: "US"},
			{IPRange: "5.5.5.0/24"},
		}, got)
	})

	t.Run("propagates csv parse errors", func(t *testing.T) {
		_, err := parseRecords("1.1.1.0/24,\"unterminated\n")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse digitalocean csv")
	})

	t.Run("fixture parses to expected row count", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("..", "testdata", "do.csv"))
		require.NoError(t, err)
		got, err := parseRecords(string(data))
		require.NoError(t, err)
		assert.Equal(t, 1144, len(got))
		assert.Equal(t, IPRange{IPRange: "5.101.96.0/21", Country: "NL", Region: "NL-NH", City: "Amsterdam", Zip: "1098 XH"}, got[0])
	})
}

func TestGetIPRanges_FromFixture(t *testing.T) {
	cfg := Config{
		Source:    filepath.Join("..", "testdata", "do.csv"),
		IPType:    "ipv4",
		Filters:   Filters{Country: []string{"NL"}, City: []string{"Amsterdam"}},
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

func TestGetIPRanges_List_FilterComposition(t *testing.T) {
	cfg := Config{
		Source:    filepath.Join("..", "testdata", "do.csv"),
		IPType:    "both",
		Filters:   Filters{Country: []string{"NL"}},
		List:      "cities",
		Verbosity: "none",
	}

	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), cfg))
	})

	lines := bytes.Split(bytes.TrimRight([]byte(out), "\n"), []byte("\n"))
	require.NotEmpty(t, lines)
	assert.Contains(t, lines, []byte("Amsterdam"))

	seen := make(map[string]struct{})
	for _, l := range lines {
		s := string(l)
		assert.NotEmpty(t, s, "empty city surfaced")
		_, dup := seen[s]
		assert.False(t, dup, "duplicate city %q", s)
		seen[s] = struct{}{}
	}
}

func TestPrintListedValues(t *testing.T) {
	ranges := []IPRange{
		{Country: "NL", Region: "NL-NH", City: "Amsterdam", Zip: "1098 XH"},
		{Country: "NL", Region: "NL-NH", City: "Amsterdam", Zip: "1098 XH"},
		{Country: "US", Region: "US-NY", City: "New York", Zip: "10001"},
		{Country: "US", Region: "", City: "", Zip: ""},
	}

	t.Run("zips drops empty and dedupes", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(ranges, "zips"))
		})
		assert.Equal(t, "10001\n1098 XH\n", out)
	})

	t.Run("unknown dim returns error", func(t *testing.T) {
		err := printListedValues(ranges, "bogus")
		require.Error(t, err)
	})
}

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
			expected:  "192.168.1.0/24,US,California,San Francisco,94107\n2607:f8b0:4005:805::200e,US,California,Mountain View,94043\n",
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
