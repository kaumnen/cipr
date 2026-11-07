package googlebot

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

var fixturePath = filepath.Join("..", "testdata", "googlebot_sample.json")

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
	require.Len(t, got, 15)
	assert.Equal(t, "2001:4860:4801:10::/64", got[0])
	assert.Equal(t, "2001:4860:4801:14::/64", got[4])
	assert.Equal(t, "192.178.4.0/27", got[5])
	assert.Equal(t, "192.178.6.0/27", got[14])
}

func TestParseIPRangesValidation(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "malformed JSON", raw: `{`, want: "parse googlebot ip-ranges json"},
		{name: "empty payload", raw: `{}`, want: "no IP ranges found"},
		{name: "empty prefixes", raw: `{"prefixes":[]}`, want: "no IP ranges found"},
		{name: "string prefixes", raw: `{"prefixes":"192.178.4.0/27"}`, want: "parse googlebot ip-ranges json"},
		{name: "missing prefix", raw: `{"prefixes":[{"ipv4Prefix":"192.178.4.0/27"},{}]}`, want: "prefix 2: missing ipv4Prefix or ipv6Prefix"},
		{name: "both prefixes", raw: `{"prefixes":[{"ipv4Prefix":"192.178.4.0/27","ipv6Prefix":"2001:4860:4801:10::/64"}]}`, want: "prefix 1: both ipv4Prefix and ipv6Prefix are set"},
		{name: "invalid CIDR", raw: `{"prefixes":[{"ipv4Prefix":"not-a-cidr"}]}`, want: `IPv4 prefix 1: "not-a-cidr" is not a valid CIDR`},
		{name: "bare address", raw: `{"prefixes":[{"ipv6Prefix":"2001:4860:4801:10::1"}]}`, want: `IPv6 prefix 1: "2001:4860:4801:10::1" is not a valid CIDR`},
		{name: "IPv6 in ipv4Prefix", raw: `{"prefixes":[{"ipv4Prefix":"2001:4860:4801:10::/64"}]}`, want: "IPv4 prefix 1: \"2001:4860:4801:10::/64\" has the wrong address family"},
		{name: "IPv4 in ipv6Prefix", raw: `{"prefixes":[{"ipv6Prefix":"192.178.4.0/27"}]}`, want: "IPv6 prefix 1: \"192.178.4.0/27\" has the wrong address family"},
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
	mixed := []string{"2001:4860:4801:10::/64", "192.178.4.0/27", "192.178.5.0/27"}

	assert.Equal(t, []string{"192.178.4.0/27", "192.178.5.0/27"}, filterIPType(mixed, "ipv4"))
	assert.Equal(t, []string{"2001:4860:4801:10::/64"}, filterIPType(mixed, "ipv6"))
	assert.Equal(t, mixed, filterIPType(mixed, "both"))
	assert.Nil(t, filterIPType([]string{"192.178.4.0/27"}, "ipv6"))
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
			ipRanges:  []string{"192.178.4.0/27", "2001:4860:4801:10::/64"},
			verbosity: "none",
			want:      "192.178.4.0/27\n2001:4860:4801:10::/64\n",
		},
		{
			name:      "mini behaves like none",
			ipRanges:  []string{"192.178.4.0/27"},
			verbosity: "mini",
			want:      "192.178.4.0/27\n",
		},
		{
			name:      "full",
			ipRanges:  []string{"192.178.4.0/27", "2001:4860:4801:10::/64"},
			verbosity: "full",
			want:      "Googlebot IP: 192.178.4.0/27\nGooglebot IP: 2001:4860:4801:10::/64\n",
		},
		{
			name:      "duplicates print once in source order",
			ipRanges:  []string{"192.178.5.0/27", "192.178.4.0/27", "192.178.5.0/27"},
			verbosity: "none",
			want:      "192.178.5.0/27\n192.178.4.0/27\n",
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
	assert.Equal(t, "Googlebot IP: 2001:4860:4801:10::/64\n"+
		"Googlebot IP: 2001:4860:4801:11::/64\n"+
		"Googlebot IP: 2001:4860:4801:12::/64\n"+
		"Googlebot IP: 2001:4860:4801:13::/64\n"+
		"Googlebot IP: 2001:4860:4801:14::/64\n", out)
}

func TestGetIPRangesInvalidSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"prefixes":[{"ipv4Prefix":"nope"}]}`), 0o600))

	err := GetIPRanges(context.Background(), Config{Source: path})
	require.Error(t, err)
	assert.ErrorContains(t, err, "validate googlebot IPv4 prefix 1")
}

func TestGetIPRangesPropagatesSourceError(t *testing.T) {
	err := GetIPRanges(context.Background(), Config{Source: filepath.Join(t.TempDir(), "missing.json")})
	require.Error(t, err)
	assert.ErrorContains(t, err, "read")
}
