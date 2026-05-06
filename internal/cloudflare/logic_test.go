package cloudflare

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", name, err)
	}
	return string(data)
}

func TestParseIPRanges(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "Whitespace-only input",
			input:    "   \n\t\n  ",
			expected: nil,
		},
		{
			name:     "Single line, no trailing newline",
			input:    "1.1.1.0/24",
			expected: []string{"1.1.1.0/24"},
		},
		{
			name:     "Multiple lines with blanks and surrounding whitespace",
			input:    "  1.1.1.0/24\n\n  2.2.2.0/24  \n",
			expected: []string{"1.1.1.0/24", "2.2.2.0/24"},
		},
		{
			name:     "IPv6 lines",
			input:    "2400:cb00::/32\n2606:4700::/32",
			expected: []string{"2400:cb00::/32", "2606:4700::/32"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseIPRanges(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseIPRanges_Fixtures(t *testing.T) {
	v4 := parseIPRanges(loadFixture(t, "cloudflare_ipv4.txt"))
	assert.Len(t, v4, 15, "cloudflare_ipv4.txt should yield 15 prefixes")
	assert.Equal(t, "173.245.48.0/20", v4[0])

	v6 := parseIPRanges(loadFixture(t, "cloudflare_ipv6.txt"))
	assert.Len(t, v6, 7, "cloudflare_ipv6.txt should yield 7 prefixes")
	assert.Equal(t, "2400:cb00::/32", v6[0])
}

func TestPrintCloudflareIPRanges(t *testing.T) {
	testCases := []struct {
		name           string
		ipRanges       []string
		verbosity      string
		expectedOutput string
	}{
		{
			name:           "Empty list",
			ipRanges:       []string{},
			verbosity:      "none",
			expectedOutput: "No IP ranges to display.\n",
		},
		{
			name:           "Verbosity none",
			ipRanges:       []string{"1.1.1.0/24", "2.2.2.0/24"},
			verbosity:      "none",
			expectedOutput: "1.1.1.0/24\n2.2.2.0/24\n",
		},
		{
			name:           "Verbosity mini behaves like none",
			ipRanges:       []string{"1.1.1.0/24"},
			verbosity:      "mini",
			expectedOutput: "1.1.1.0/24\n",
		},
		{
			name:           "Verbosity full",
			ipRanges:       []string{"1.1.1.0/24", "2606:4700::/32"},
			verbosity:      "full",
			expectedOutput: "Cloudflare IP: 1.1.1.0/24\nCloudflare IP: 2606:4700::/32\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printCloudflareIPRanges(tc.ipRanges, tc.verbosity)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)

			assert.Equal(t, tc.expectedOutput, buf.String())
		})
	}
}
