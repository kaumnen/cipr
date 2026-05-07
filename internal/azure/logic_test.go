package azure

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", "azure_servicetags_sample.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(data)
}

func TestSeparateFilters(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"both set", "westeurope,AzureStorage", []string{"westeurope", "AzureStorage"}},
		{"region only", "westeurope,", []string{"westeurope", "*"}},
		{"service only", ",AzureStorage", []string{"*", "AzureStorage"}},
		{"empty", "", []string{"*", "*"}},
		{"trailing tokens dropped", "westeurope,AzureStorage,extra", []string{"westeurope", "AzureStorage"}},
		{"whitespace stripped", " westeurope , AzureStorage ", []string{"westeurope", "AzureStorage"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, separateFilters(tc.input))
		})
	}
}

func TestFiltrateIPRanges(t *testing.T) {
	raw := loadFixture(t)

	cases := []struct {
		name     string
		ipType   string
		filter   []string
		expected []Prefix
	}{
		{
			name:   "ipv4 no filter",
			ipType: "ipv4",
			filter: []string{"*", "*"},
			expected: []Prefix{
				{Address: "13.66.143.220/30", Region: "", Service: "ActionGroup"},
				{Address: "20.50.32.0/19", Region: "westeurope", Service: "AzureStorage"},
				{Address: "20.42.0.0/15", Region: "eastus", Service: ""},
				{Address: "20.55.0.0/16", Region: "eastus", Service: ""},
			},
		},
		{
			name:   "ipv6 no filter",
			ipType: "ipv6",
			filter: []string{"*", "*"},
			expected: []Prefix{
				{Address: "2603:1000:4::10c/126", Region: "", Service: "ActionGroup"},
				{Address: "2603:1020:206::/48", Region: "westeurope", Service: "AzureStorage"},
			},
		},
		{
			name:     "filter by region",
			ipType:   "ipv4",
			filter:   []string{"westeurope", "*"},
			expected: []Prefix{{Address: "20.50.32.0/19", Region: "westeurope", Service: "AzureStorage"}},
		},
		{
			name:     "filter by service",
			ipType:   "ipv4",
			filter:   []string{"*", "ActionGroup"},
			expected: []Prefix{{Address: "13.66.143.220/30", Region: "", Service: "ActionGroup"}},
		},
		{
			name:     "filter by region and service combined",
			ipType:   "ipv6",
			filter:   []string{"westeurope", "AzureStorage"},
			expected: []Prefix{{Address: "2603:1020:206::/48", Region: "westeurope", Service: "AzureStorage"}},
		},
		{
			name:     "filter case-insensitive",
			ipType:   "ipv4",
			filter:   []string{"WESTEUROPE", "azurestorage"},
			expected: []Prefix{{Address: "20.50.32.0/19", Region: "westeurope", Service: "AzureStorage"}},
		},
		{
			name:     "no matches",
			ipType:   "ipv4",
			filter:   []string{"southpole", "*"},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := filtrateIPRanges(raw, tc.ipType, tc.filter)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestFiltrateIPRangesInvalidJSON(t *testing.T) {
	_, err := filtrateIPRanges("not json", "ipv4", []string{"*", "*"})
	assert.Error(t, err)
}

func TestJSONURLRegex(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "matches typical href",
			input: `<a href="https://download.microsoft.com/download/7/1/d/71d86715-5596-4529-9b13-da13a5de5b63/ServiceTags_Public_20260504.json">link</a>`,
			want:  "https://download.microsoft.com/download/7/1/d/71d86715-5596-4529-9b13-da13a5de5b63/ServiceTags_Public_20260504.json",
		},
		{
			name:  "matches with single quotes",
			input: `data='https://download.microsoft.com/download/a/b/c/abc/ServiceTags_Public_20260101.json'`,
			want:  "https://download.microsoft.com/download/a/b/c/abc/ServiceTags_Public_20260101.json",
		},
		{
			name:  "no match returns empty",
			input: `nothing here at all`,
			want:  "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, jsonURLRegex.FindString(tc.input))
		})
	}
}

func TestScrapeMissError(t *testing.T) {
	cases := []struct {
		name     string
		body     []byte
		wantSub  string
		wantHint bool
	}{
		{
			name:     "small body looks like anti-bot stub",
			body:     bytes.Repeat([]byte("x"), 4*1024),
			wantSub:  "anti-bot stub",
			wantHint: true,
		},
		{
			name:     "large body without match is treated as layout change",
			body:     bytes.Repeat([]byte("x"), 60*1024),
			wantSub:  "page layout may have changed",
			wantHint: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := scrapeMissError("https://example.test/page", tc.body)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantSub)
			if tc.wantHint {
				assert.Contains(t, err.Error(), "Workarounds:")
			}
		})
	}
}

func TestPrintIPRanges(t *testing.T) {
	cases := []struct {
		name      string
		prefixes  []Prefix
		verbosity string
		want      string
	}{
		{
			name:      "empty",
			prefixes:  []Prefix{},
			verbosity: "none",
			want:      "No IP ranges to display.\n",
		},
		{
			name: "none verbosity",
			prefixes: []Prefix{
				{Address: "20.50.32.0/19", Region: "westeurope", Service: "AzureStorage"},
			},
			verbosity: "none",
			want:      "20.50.32.0/19\n",
		},
		{
			name: "mini verbosity",
			prefixes: []Prefix{
				{Address: "20.50.32.0/19", Region: "westeurope", Service: "AzureStorage"},
			},
			verbosity: "mini",
			want:      "20.50.32.0/19,westeurope,AzureStorage\n",
		},
		{
			name: "full verbosity",
			prefixes: []Prefix{
				{Address: "2603:1020:206::/48", Region: "westeurope", Service: "AzureStorage"},
			},
			verbosity: "full",
			want:      "IP Prefix: 2603:1020:206::/48, Region: westeurope, Service: AzureStorage\n",
		},
		{
			name: "unknown verbosity falls back to addresses only",
			prefixes: []Prefix{
				{Address: "13.66.143.220/30", Region: "", Service: "ActionGroup"},
			},
			verbosity: "bogus",
			want:      "13.66.143.220/30\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printIPRanges(tc.prefixes, tc.verbosity)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			assert.Equal(t, tc.want, buf.String())
		})
	}
}

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

func TestPrintListedValues(t *testing.T) {
	prefixes := []Prefix{
		{Address: "13.66.143.220/30", Region: "", Service: "ActionGroup"},
		{Address: "20.50.32.0/19", Region: "westeurope", Service: "AzureStorage"},
		{Address: "20.42.0.0/15", Region: "eastus", Service: ""},
		{Address: "2603:1020:206::/48", Region: "westeurope", Service: "AzureStorage"},
	}

	t.Run("regions drops empty and dedupes", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(prefixes, "regions"))
		})
		assert.Equal(t, "eastus\nwesteurope\n", out)
	})

	t.Run("services drops empty and dedupes", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(prefixes, "services"))
		})
		assert.Equal(t, "ActionGroup\nAzureStorage\n", out)
	})

	t.Run("unknown dim returns error", func(t *testing.T) {
		err := printListedValues(prefixes, "bogus")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown list dimension")
	})
}

func TestGetIPRanges_List(t *testing.T) {
	cfg := Config{
		Source:    filepath.Join("..", "testdata", "azure_servicetags_sample.json"),
		IPType:    "",
		List:      "services",
		Verbosity: "none",
	}

	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), cfg))
	})
	assert.Equal(t, "ActionGroup\nAzureStorage\n", out)
}

// Regression: azure's two-stage hosted fetch (HTML scrape -> JSON download)
// previously bypassed the cache because both steps reduced to raw URL calls
// that GetRawData doesn't cache. Verify the cache wrap now keys the resolved
// JSON under "azure" so a second call within TTL returns from disk.
func TestFetchRawData_HostedCachesUnderAzureKey(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	jsonBody := loadFixture(t)

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(jsonBody))
	}))
	defer srv.Close()

	// Point at a .json endpoint to skip the page-scrape branch (HTTP-redirecting
	// download.microsoft.com URLs into a test server is more trouble than it's
	// worth here). What we're proving is that fetchRawData wraps its inner
	// work in GetCached at all — the scrape branch shares the same wrapper.
	viper.Set("azure_endpoint", srv.URL+"/ServiceTags_Public_20260101.json")

	for i := 0; i < 2; i++ {
		got, err := fetchRawData(context.Background(), "azure")
		require.NoError(t, err)
		assert.Equal(t, jsonBody, got)
	}
	assert.Equal(t, 1, hits, "second call should hit cache, not refetch")
}

func TestFetchRawData_NoCacheBypass(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	jsonBody := loadFixture(t)

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(jsonBody))
	}))
	defer srv.Close()

	viper.Set("azure_endpoint", srv.URL+"/x.json")
	viper.Set("no_cache", true)

	for i := 0; i < 2; i++ {
		_, err := fetchRawData(context.Background(), "azure")
		require.NoError(t, err)
	}
	assert.Equal(t, 2, hits, "--no-cache should refetch every call")
}
