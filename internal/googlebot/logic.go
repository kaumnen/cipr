package googlebot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kaumnen/cipr/internal/utils"
)

type sourcePrefix struct {
	IPv4Prefix string `json:"ipv4Prefix"`
	IPv6Prefix string `json:"ipv6Prefix"`
}

type IPsData struct {
	CreationTime string         `json:"creationTime"`
	Prefixes     []sourcePrefix `json:"prefixes"`
}

type Config struct {
	Source    string
	IPType    string
	Verbosity string
}

func GetIPRanges(ctx context.Context, config Config) error {
	rawData, err := utils.GetRawData(ctx, config.Source)
	if err != nil {
		return err
	}
	ipRanges, err := parseIPRanges(rawData)
	if err != nil {
		utils.DropCache(config.Source)
		return err
	}
	printIPRanges(filterIPType(ipRanges, config.IPType), config.Verbosity)
	return nil
}

func parseIPRanges(rawData string) ([]string, error) {
	var data IPsData
	if err := json.Unmarshal([]byte(rawData), &data); err != nil {
		return nil, fmt.Errorf("parse googlebot ip-ranges json: %w", err)
	}
	if len(data.Prefixes) == 0 {
		return nil, fmt.Errorf("validate googlebot ip-ranges json: no IP ranges found")
	}

	ipRanges := make([]string, 0, len(data.Prefixes))
	for i, source := range data.Prefixes {
		if source.IPv4Prefix == "" && source.IPv6Prefix == "" {
			return nil, fmt.Errorf("validate googlebot prefix %d: missing ipv4Prefix or ipv6Prefix", i+1)
		}
		if source.IPv4Prefix != "" && source.IPv6Prefix != "" {
			return nil, fmt.Errorf("validate googlebot prefix %d: both ipv4Prefix and ipv6Prefix are set", i+1)
		}

		address := source.IPv4Prefix
		family := "IPv4"
		if address == "" {
			address = source.IPv6Prefix
			family = "IPv6"
		}
		if !utils.IsCIDR(address) {
			return nil, fmt.Errorf("validate googlebot %s prefix %d: %q is not a valid CIDR", family, i+1, address)
		}
		wrongFamily := (family == "IPv4" && !utils.IsIPv4(address)) ||
			(family == "IPv6" && !utils.IsIPv6(address))
		if wrongFamily {
			return nil, fmt.Errorf("validate googlebot %s prefix %d: %q has the wrong address family", family, i+1, address)
		}
		ipRanges = append(ipRanges, address)
	}
	return ipRanges, nil
}

func filterIPType(ipRanges []string, ipType string) []string {
	var keep func(string) bool
	switch ipType {
	case "ipv4":
		keep = utils.IsIPv4
	case "ipv6":
		keep = utils.IsIPv6
	default:
		return ipRanges
	}
	var result []string
	for _, ip := range ipRanges {
		if keep(ip) {
			result = append(result, ip)
		}
	}
	return result
}

func printIPRanges(ipRanges []string, verbosity string) {
	if len(ipRanges) == 0 {
		fmt.Fprintln(os.Stderr, "No IP ranges to display.")
		return
	}

	seen := make(map[string]struct{}, len(ipRanges))
	for _, ip := range ipRanges {
		if _, dup := seen[ip]; dup {
			continue
		}
		seen[ip] = struct{}{}
		if verbosity == "full" {
			fmt.Printf("Googlebot IP: %s\n", ip)
		} else {
			fmt.Println(ip)
		}
	}
}
