package utils

import (
	"net/netip"
	"strings"

	"github.com/spf13/viper"
)

// DefaultEndpoints maps a provider config-key prefix to its default URL.
// createDefaultConfig writes <key>_endpoint and <key>_local_file for each entry.
var DefaultEndpoints = map[string]string{
	"aws":             "https://ip-ranges.amazonaws.com/ip-ranges.json",
	"cloudflare_ipv4": "https://www.cloudflare.com/ips-v4/",
	"cloudflare_ipv6": "https://www.cloudflare.com/ips-v6/",
	"icloud":          "https://mask-api.icloud.com/egress-ip-ranges.csv",
	"digitalocean":    "https://digitalocean.com/geo/google.csv",
}

// ResolveSource returns the source token to pass to GetRawData. When the
// global --source is "hosted", returns hostedKey (a config-key prefix);
// otherwise returns the user-provided URL or path verbatim.
func ResolveSource(hostedKey string) string {
	src := viper.GetString("source")
	if src == "hosted" {
		return hostedKey
	}
	return src
}

func ContainsIgnoreCase(slice []string, item string) bool {
	if item == "" {
		return false
	}
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// parseIP accepts either a CIDR ("1.2.3.0/24") or a bare address ("1.2.3.4")
// and returns the address. Returns the zero Addr for unparseable input.
func parseIP(s string) netip.Addr {
	if prefix, err := netip.ParsePrefix(s); err == nil {
		return prefix.Addr()
	}
	if addr, err := netip.ParseAddr(s); err == nil {
		return addr
	}
	return netip.Addr{}
}

func IsIPv4(s string) bool {
	addr := parseIP(s)
	return addr.IsValid() && addr.Is4()
}

func IsIPv6(s string) bool {
	addr := parseIP(s)
	return addr.IsValid() && addr.Is6()
}
