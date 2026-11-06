package fastly

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

var fixturePath = filepath.Join("..", "testdata", "fastly_sample.json")

func loadFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(fixturePath)
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

func TestParseIPRangesFixture(t *testing.T) {
	got, err := parseIPRanges(loadFixture(t))
	require.NoError(t, err)
	require.Len(t, got, 24)
	assert.Equal(t, "23.235.32.0/20", got[0])
	assert.Equal(t, "199.232.0.0/16", got[20])
	assert.Equal(t, []string{"2a04:4e40::/32", "2a04:4e42::/31", "2a04:4e44::/32"}, got[21:])
}

func TestParseIPRangesAcceptsOneFamily(t *testing.T) {
	got, err := parseIPRanges(`{"ipv6_addresses":["2a04:4e40::/32"]}`)
	require.NoError(t, err)
	assert.Equal(t, []string{"2a04:4e40::/32"}, got)
}

func TestParseIPRangesValidation(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "malformed JSON", raw: `{`, want: "parse fastly public-ip-list json"},
		{name: "empty payload", raw: `{}`, want: "no IP ranges found"},
		{name: "empty arrays", raw: `{"addresses":[],"ipv6_addresses":[]}`, want: "no IP ranges found"},
		{name: "string addresses", raw: `{"addresses":"151.101.0.0/16"}`, want: "parse fastly public-ip-list json"},
		{name: "invalid IPv4 entry", raw: `{"addresses":["151.101.0.0/16","not-a-cidr"]}`, want: `addresses entry 2: "not-a-cidr"`},
		{name: "invalid IPv6 entry", raw: `{"addresses":["151.101.0.0/16"],"ipv6_addresses":[""]}`, want: `ipv6_addresses entry 1: "" is not a valid CIDR`},
		{name: "bare address", raw: `{"addresses":["151.101.0.1"]}`, want: `addresses entry 1: "151.101.0.1"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseIPRanges(tt.raw)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.want)
		})
	}
}

func TestFilterIPType(t *testing.T) {
	mixed := []string{"151.101.0.0/16", "2a04:4e40::/32", "199.232.0.0/16"}

	assert.Equal(t, []string{"151.101.0.0/16", "199.232.0.0/16"}, filterIPType(mixed, "ipv4"))
	assert.Equal(t, []string{"2a04:4e40::/32"}, filterIPType(mixed, "ipv6"))
	assert.Equal(t, mixed, filterIPType(mixed, "both"))
	assert.Nil(t, filterIPType([]string{"151.101.0.0/16"}, "ipv6"))
}

func TestPrintIPRanges(t *testing.T) {
	tests := []struct {
		name      string
		ipRanges  []string
		verbosity string
		want      string
	}{
		{name: "empty", verbosity: "none", want: ""},
		{
			name:      "none",
			ipRanges:  []string{"151.101.0.0/16", "2a04:4e40::/32"},
			verbosity: "none",
			want:      "151.101.0.0/16\n2a04:4e40::/32\n",
		},
		{
			name:      "mini behaves like none",
			ipRanges:  []string{"151.101.0.0/16"},
			verbosity: "mini",
			want:      "151.101.0.0/16\n",
		},
		{
			name:      "full",
			ipRanges:  []string{"151.101.0.0/16", "2a04:4e40::/32"},
			verbosity: "full",
			want:      "Fastly IP: 151.101.0.0/16\nFastly IP: 2a04:4e40::/32\n",
		},
		{
			name:      "duplicates print once in source order",
			ipRanges:  []string{"199.232.0.0/16", "151.101.0.0/16", "199.232.0.0/16"},
			verbosity: "none",
			want:      "199.232.0.0/16\n151.101.0.0/16\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() { printIPRanges(tt.ipRanges, tt.verbosity) })
			assert.Equal(t, tt.want, out)
		})
	}
}

func TestGetIPRanges(t *testing.T) {
	out := captureStdout(t, func() {
		require.NoError(t, GetIPRanges(context.Background(), Config{Source: fixturePath, IPType: "ipv6", Verbosity: "full"}))
	})
	assert.Equal(t, "Fastly IP: 2a04:4e40::/32\nFastly IP: 2a04:4e42::/31\nFastly IP: 2a04:4e44::/32\n", out)
}

func TestGetIPRangesInvalidSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"addresses":["nope"]}`), 0o600))

	err := GetIPRanges(context.Background(), Config{Source: path})
	require.Error(t, err)
	assert.ErrorContains(t, err, "validate fastly addresses entry 1")
}

func TestGetIPRangesPropagatesSourceError(t *testing.T) {
	err := GetIPRanges(context.Background(), Config{Source: filepath.Join(t.TempDir(), "missing.json")})
	require.Error(t, err)
	assert.ErrorContains(t, err, "read")
}
