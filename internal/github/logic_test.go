package github

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func TestParseMeta(t *testing.T) {
	raw := mockGetReq()

	ranges, err := parseMeta(raw)
	require.NoError(t, err)

	seen := make(map[string]int)
	for _, r := range ranges {
		seen[r.Service]++
	}

	for _, k := range []string{"verifiable_password_authentication", "ssh_key_fingerprints", "ssh_keys"} {
		_, present := seen[k]
		assert.False(t, present, "non-CIDR key %q must not leak into parsed ranges", k)
	}

	expectedCounts := map[string]int{
		"hooks":                      3,
		"web":                        4,
		"api":                        4,
		"git":                        4,
		"github_enterprise_importer": 3,
		"packages":                   3,
		"pages":                      4,
		"importer":                   3,
		"actions":                    4,
	}
	for service, want := range expectedCounts {
		assert.Equal(t, want, seen[service], "service %q CIDR count", service)
	}

	assert.Len(t, ranges, 32)
}

func TestFiltrate(t *testing.T) {
	raw := mockGetReq()
	ranges, err := parseMeta(raw)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		ipType        string
		filterService string
		expectedLen   int
		assertService string
	}{
		{
			name:        "ipv4 only, no service filter",
			ipType:      "ipv4",
			expectedLen: 25,
		},
		{
			name:        "ipv6 only, no service filter",
			ipType:      "ipv6",
			expectedLen: 7,
		},
		{
			name:          "both families, filter actions",
			ipType:        "",
			filterService: "actions",
			expectedLen:   4,
			assertService: "actions",
		},
		{
			name:          "ipv4 + service filter case-insensitive",
			ipType:        "ipv4",
			filterService: "ACTIONS",
			expectedLen:   3,
			assertService: "actions",
		},
		{
			name:          "nonexistent service yields empty",
			ipType:        "",
			filterService: "nonexistent",
			expectedLen:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filtrate(ranges, tc.ipType, tc.filterService)
			assert.Equal(t, tc.expectedLen, len(result))
			if tc.assertService != "" {
				for _, r := range result {
					assert.Equal(t, tc.assertService, r.Service)
				}
			}
		})
	}
}

func TestPrintIPRanges(t *testing.T) {
	testCases := []struct {
		name           string
		ranges         []IPRange
		verbosity      string
		expectedOutput string
	}{
		{
			name:           "empty list",
			ranges:         []IPRange{},
			verbosity:      "none",
			expectedOutput: "No IP ranges to display.\n",
		},
		{
			name: "none prints CIDR only",
			ranges: []IPRange{
				{CIDR: "192.30.252.0/22", Service: "hooks"},
				{CIDR: "2606:50c0::/32", Service: "packages"},
			},
			verbosity:      "none",
			expectedOutput: "192.30.252.0/22\n2606:50c0::/32\n",
		},
		{
			name: "mini prints CIDR,service",
			ranges: []IPRange{
				{CIDR: "4.148.0.0/16", Service: "actions"},
			},
			verbosity:      "mini",
			expectedOutput: "4.148.0.0/16,actions\n",
		},
		{
			name: "full prints labeled output",
			ranges: []IPRange{
				{CIDR: "140.82.112.0/20", Service: "api"},
				{CIDR: "2a0a:a440::/29", Service: "web"},
			},
			verbosity:      "full",
			expectedOutput: "IP Prefix: 140.82.112.0/20, Service: api\nIP Prefix: 2a0a:a440::/29, Service: web\n",
		},
		{
			name: "invalid verbosity defaults to none",
			ranges: []IPRange{
				{CIDR: "192.30.252.153/32", Service: "pages"},
			},
			verbosity:      "bogus",
			expectedOutput: "192.30.252.153/32\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				printIPRanges(tc.ranges, tc.verbosity)
			})
			assert.Equal(t, tc.expectedOutput, out)
		})
	}
}

func TestPrintListedValues(t *testing.T) {
	ranges := []IPRange{
		{CIDR: "1.1.1.0/24", Service: "web"},
		{CIDR: "1.1.2.0/24", Service: "actions"},
		{CIDR: "1.1.3.0/24", Service: "web"},
		{CIDR: "1.1.4.0/24", Service: "api"},
	}

	t.Run("services sorted and deduped", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(ranges, "services"))
		})
		assert.Equal(t, "actions\napi\nweb\n", out)
	})

	t.Run("empty input prints placeholder", func(t *testing.T) {
		out := captureStdout(t, func() {
			require.NoError(t, printListedValues(nil, "services"))
		})
		assert.Equal(t, "No values to display.\n", out)
	})

	t.Run("unknown dim returns error", func(t *testing.T) {
		err := printListedValues(ranges, "bogus")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown list dimension")
	})
}

func TestGetIPRanges_List(t *testing.T) {
	cfg := Config{
		Source:    filepath.Join("..", "testdata", "github_meta_sample.json"),
		IPType:    "",
		List:      "services",
		Verbosity: "none",
	}

	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), cfg))
	})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Len(t, lines, 9)
	assert.True(t, sort.StringsAreSorted(lines), "services output not sorted: %v", lines)

	seen := make(map[string]struct{})
	for _, l := range lines {
		assert.NotEmpty(t, l, "empty service surfaced")
		_, dup := seen[l]
		assert.False(t, dup, "duplicate service %q", l)
		seen[l] = struct{}{}
	}
	for _, k := range []string{"ssh_keys", "ssh_key_fingerprints", "verifiable_password_authentication"} {
		_, leaked := seen[k]
		assert.False(t, leaked, "non-CIDR key %q leaked into services list", k)
	}
}
