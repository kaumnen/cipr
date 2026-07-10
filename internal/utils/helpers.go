package utils

import (
	"net/netip"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

// DefaultEndpoints maps a provider config-key prefix to its default URL.
// createDefaultConfig writes <key>_endpoint and <key>_local_file for each entry.
var DefaultEndpoints = map[string]string{
	"aws":             "https://ip-ranges.amazonaws.com/ip-ranges.json",
	"azure":           "https://www.microsoft.com/en-us/download/details.aspx?id=56519",
	"cloudflare_ipv4": "https://www.cloudflare.com/ips-v4/",
	"cloudflare_ipv6": "https://www.cloudflare.com/ips-v6/",
	"icloud":          "https://mask-api.icloud.com/egress-ip-ranges.csv",
	"digitalocean":    "https://digitalocean.com/geo/google.csv",
	"gcp":             "https://www.gstatic.com/ipranges/cloud.json",
	"github":          "https://api.github.com/meta",
}

// ResolveSource returns the source token to pass to GetRawData. The "config"
// default and the legacy "hosted" alias both resolve to the provider's config
// key; URLs and paths are returned verbatim.
func ResolveSource(hostedKey string) string {
	src := viper.GetString("source")
	if src == "config" || src == "hosted" {
		return hostedKey
	}
	return src
}

// IsConfiguredSource reports whether source names a provider whose endpoint or
// local-file settings should be read from viper. Everything else that is not an
// HTTP(S) URL is treated as a local path, including a bare filename.
func IsConfiguredSource(source string) bool {
	if _, ok := DefaultEndpoints[source]; ok {
		return true
	}
	return viper.IsSet(source+"_endpoint") || viper.IsSet(source+"_local_file")
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

func IsCIDR(s string) bool {
	_, err := netip.ParsePrefix(s)
	return err == nil
}

// DedupeSorted returns the input with empty entries dropped and the
// remaining values deduplicated and sorted ascending.
func DedupeSorted(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
