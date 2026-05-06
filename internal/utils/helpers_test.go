package utils

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestResolveSource(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	viper.Set("source", "hosted")
	assert.Equal(t, "aws", ResolveSource("aws"))
	assert.Equal(t, "cloudflare_ipv6", ResolveSource("cloudflare_ipv6"))

	viper.Set("source", "https://example.com/ranges.csv")
	assert.Equal(t, "https://example.com/ranges.csv", ResolveSource("aws"))

	viper.Set("source", "/tmp/local.csv")
	assert.Equal(t, "/tmp/local.csv", ResolveSource("aws"))
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"present same case", []string{"apple", "banana"}, "banana", true},
		{"present different case", []string{"apple", "Banana"}, "banana", true},
		{"absent", []string{"apple", "banana"}, "grape", false},
		{"empty slice", []string{}, "apple", false},
		{"empty item", []string{"apple"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ContainsIgnoreCase(tt.slice, tt.item))
		})
	}
}

func TestIPFamily(t *testing.T) {
	cases := []struct {
		input string
		v4    bool
		v6    bool
	}{
		{"1.2.3.4", true, false},
		{"1.2.3.0/24", true, false},
		{"2606:4700::/32", false, true},
		{"2001:4860:4860::8888", false, true},
		{"::ffff:1.2.3.4", false, true}, // IPv4-mapped IPv6
		{"not-an-ip", false, false},
		{"", false, false},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			assert.Equal(t, c.v4, IsIPv4(c.input), "IsIPv4(%q)", c.input)
			assert.Equal(t, c.v6, IsIPv6(c.input), "IsIPv6(%q)", c.input)
		})
	}
}
