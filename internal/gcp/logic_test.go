package gcp

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

func loadFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", "gcp_cloud_sample.json"))
	require.NoError(t, err)
	return string(data)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = oldStdout })

	fn()

	require.NoError(t, w.Close())
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return buf.String()
}

func TestSeparateFilters(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty", want: []string{"*", "*"}},
		{name: "scope only", input: "global", want: []string{"global", "*"}},
		{name: "service only", input: ",Google Cloud", want: []string{"*", "Google Cloud"}},
		{name: "both", input: "asia-east1,Google Cloud", want: []string{"asia-east1", "Google Cloud"}},
		{name: "trims surrounding whitespace", input: " asia-east1 , Google Cloud ", want: []string{"asia-east1", "Google Cloud"}},
		{name: "truncates extra values", input: "global,Google Cloud,ignored", want: []string{"global", "Google Cloud"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, separateFilters(tt.input))
		})
	}
}

func TestFiltrateIPRanges(t *testing.T) {
	rawData := loadFixture(t)
	tests := []struct {
		name   string
		ipType string
		filter string
		want   []Prefix
	}{
		{
			name:   "IPv4",
			ipType: "ipv4",
			want: []Prefix{
				{Address: "34.1.208.0/20", Scope: "africa-south1", Service: "Google Cloud"},
				{Address: "34.80.0.0/15", Scope: "asia-east1", Service: "Google Cloud"},
				{Address: "8.228.224.0/20", Scope: "global", Service: "Google Cloud"},
			},
		},
		{
			name:   "IPv6",
			ipType: "ipv6",
			want: []Prefix{
				{Address: "2600:1900:8000::/44", Scope: "africa-south1", Service: "Google Cloud"},
				{Address: "2600:1900:4030::/44", Scope: "asia-east1", Service: "Google Cloud"},
				{Address: "2600:1901::/48", Scope: "global", Service: "Google Cloud"},
			},
		},
		{
			name:   "both families preserve source order",
			ipType: "both",
			want: []Prefix{
				{Address: "34.1.208.0/20", Scope: "africa-south1", Service: "Google Cloud"},
				{Address: "2600:1900:8000::/44", Scope: "africa-south1", Service: "Google Cloud"},
				{Address: "34.80.0.0/15", Scope: "asia-east1", Service: "Google Cloud"},
				{Address: "2600:1900:4030::/44", Scope: "asia-east1", Service: "Google Cloud"},
				{Address: "8.228.224.0/20", Scope: "global", Service: "Google Cloud"},
				{Address: "2600:1901::/48", Scope: "global", Service: "Google Cloud"},
			},
		},
		{
			name:   "scope filter is case-insensitive",
			ipType: "ipv4",
			filter: "ASIA-EAST1,",
			want:   []Prefix{{Address: "34.80.0.0/15", Scope: "asia-east1", Service: "Google Cloud"}},
		},
		{
			name:   "service filter preserves embedded space",
			ipType: "ipv6",
			filter: ",google cloud",
			want: []Prefix{
				{Address: "2600:1900:8000::/44", Scope: "africa-south1", Service: "Google Cloud"},
				{Address: "2600:1900:4030::/44", Scope: "asia-east1", Service: "Google Cloud"},
				{Address: "2600:1901::/48", Scope: "global", Service: "Google Cloud"},
			},
		},
		{
			name:   "combined filter",
			ipType: "ipv6",
			filter: "GLOBAL,GOOGLE CLOUD",
			want:   []Prefix{{Address: "2600:1901::/48", Scope: "global", Service: "Google Cloud"}},
		},
		{name: "no matches", ipType: "both", filter: "moon,*", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filtrateIPRanges(rawData, tt.ipType, filtersFromComposite(tt.filter))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFiltrateIPRangesValidation(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "malformed JSON", raw: `{`, want: "parse gcp cloud ip-ranges json"},
		{name: "empty payload", raw: `{}`, want: "no IP ranges found"},
		{name: "empty prefixes", raw: `{"prefixes":[]}`, want: "no IP ranges found"},
		{name: "missing address", raw: `{"prefixes":[{"service":"Google Cloud","scope":"global"}]}`, want: "missing ipv4Prefix or ipv6Prefix"},
		{name: "both addresses", raw: `{"prefixes":[{"ipv4Prefix":"1.1.1.0/24","ipv6Prefix":"2001:db8::/32"}]}`, want: "both ipv4Prefix and ipv6Prefix"},
		{name: "invalid CIDR", raw: `{"prefixes":[{"ipv4Prefix":"not-a-cidr"}]}`, want: `IPv4 prefix 1: "not-a-cidr"`},
		{name: "IPv4 field has IPv6", raw: `{"prefixes":[{"ipv4Prefix":"2001:db8::/32"}]}`, want: "wrong address family"},
		{name: "IPv6 field has IPv4", raw: `{"prefixes":[{"ipv6Prefix":"192.0.2.0/24"}]}`, want: "wrong address family"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := filtrateIPRanges(tt.raw, "both", Filters{})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestFiltrateIPRangesValidatesExcludedEntries(t *testing.T) {
	rawData := `{"prefixes":[
		{"ipv4Prefix":"192.0.2.0/24","service":"Google Cloud","scope":"global"},
		{"ipv6Prefix":"invalid","service":"Google Cloud","scope":"elsewhere"}
	]}`
	_, err := filtrateIPRanges(rawData, "ipv4", filtersFromComposite("global,"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "validate gcp IPv6 prefix 2")
}

func TestFiltrateIPRangesFiltersService(t *testing.T) {
	rawData := `{"prefixes":[
		{"ipv4Prefix":"192.0.2.0/24","service":"Google Cloud","scope":"global"},
		{"ipv4Prefix":"198.51.100.0/24","service":"Future Service","scope":"global"}
	]}`
	got, err := filtrateIPRanges(rawData, "ipv4", filtersFromComposite(",google cloud"))
	require.NoError(t, err)
	assert.Equal(t, []Prefix{
		{Address: "192.0.2.0/24", Scope: "global", Service: "Google Cloud"},
	}, got)
}

func TestFiltrateIPRangesMultipleValues(t *testing.T) {
	rawData := `{"prefixes":[
		{"ipv4Prefix":"192.0.2.0/24","service":"Google Cloud","scope":"global"},
		{"ipv4Prefix":"198.51.100.0/24","service":"Future Service","scope":"asia-east1"},
		{"ipv4Prefix":"203.0.113.0/24","service":"Other Service","scope":"africa-south1"}
	]}`

	got, err := filtrateIPRanges(rawData, "ipv4", Filters{
		Scope:   []string{"GLOBAL", "asia-east1"},
		Service: []string{"google cloud", "FUTURE SERVICE"},
	})
	require.NoError(t, err)
	assert.Equal(t, []Prefix{
		{Address: "192.0.2.0/24", Scope: "global", Service: "Google Cloud"},
		{Address: "198.51.100.0/24", Scope: "asia-east1", Service: "Future Service"},
	}, got)
}

func TestPrintIPRanges(t *testing.T) {
	tests := []struct {
		name      string
		prefixes  []Prefix
		verbosity string
		want      string
	}{
		{name: "empty", verbosity: "none", want: "No IP ranges to display.\n"},
		{
			name: "none",
			prefixes: []Prefix{
				{Address: "34.80.0.0/15", Scope: "asia-east1", Service: "Google Cloud"},
				{Address: "2600:1900:4030::/44", Scope: "asia-east1", Service: "Google Cloud"},
			},
			verbosity: "none",
			want:      "34.80.0.0/15\n2600:1900:4030::/44\n",
		},
		{
			name:      "mini",
			prefixes:  []Prefix{{Address: "8.228.224.0/20", Scope: "global", Service: "Google Cloud"}},
			verbosity: "mini",
			want:      "8.228.224.0/20,global,Google Cloud\n",
		},
		{
			name:      "full",
			prefixes:  []Prefix{{Address: "2600:1901::/48", Scope: "global", Service: "Google Cloud"}},
			verbosity: "full",
			want:      "IP Prefix: 2600:1901::/48, Scope: global, Service: Google Cloud\n",
		},
		{
			name:      "unknown verbosity falls back to CIDR only",
			prefixes:  []Prefix{{Address: "34.1.208.0/20", Scope: "africa-south1", Service: "Google Cloud"}},
			verbosity: "unknown",
			want:      "34.1.208.0/20\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() { printIPRanges(tt.prefixes, tt.verbosity) })
			assert.Equal(t, tt.want, out)
		})
	}
}

func TestPrintListedValues(t *testing.T) {
	prefixes := []Prefix{
		{Address: "34.80.0.0/15", Scope: "asia-east1", Service: "Google Cloud"},
		{Address: "2600:1900:4030::/44", Scope: "asia-east1", Service: "Google Cloud"},
		{Address: "8.228.224.0/20", Scope: "global", Service: "Google Cloud"},
		{Address: "192.0.2.0/24", Scope: "", Service: ""},
	}

	t.Run("scopes are sorted deduplicated and non-empty", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(prefixes, "scopes"))
		})
		assert.Equal(t, "asia-east1\nglobal\n", out)
	})
	t.Run("services are sorted deduplicated and non-empty", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(prefixes, "services"))
		})
		assert.Equal(t, "Google Cloud\n", out)
	})
	t.Run("empty input", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(nil, "scopes"))
		})
		assert.Equal(t, "No values to display.\n", out)
	})
	t.Run("unknown dimension", func(t *testing.T) {
		err := printListedValues(prefixes, "regions")
		require.Error(t, err)
		assert.ErrorContains(t, err, "valid: scopes, services")
	})
}

func TestGetIPRanges(t *testing.T) {
	config := Config{
		Source:    filepath.Join("..", "testdata", "gcp_cloud_sample.json"),
		IPType:    "ipv4",
		Filter:    "global,Google Cloud",
		Verbosity: "mini",
	}
	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), config))
	})
	assert.Equal(t, "8.228.224.0/20,global,Google Cloud\n", out)
}

func TestGetIPRangesListComposesWithFilterAndIgnoresIPType(t *testing.T) {
	config := Config{
		Source: filepath.Join("..", "testdata", "gcp_cloud_sample.json"),
		IPType: "ipv4",
		Filter: ",Google Cloud",
		List:   "scopes",
	}
	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), config))
	})
	assert.Equal(t, "africa-south1\nasia-east1\nglobal\n", out)
}

func TestGetIPRangesPropagatesSourceError(t *testing.T) {
	err := GetIPRanges(context.Background(), Config{Source: filepath.Join(t.TempDir(), "missing.json")})
	require.Error(t, err)
	assert.ErrorContains(t, err, "read")
}
