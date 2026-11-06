package fastly

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/kaumnen/cipr/internal/utils"
)

type IPsData struct {
	Addresses     []string `json:"addresses"`
	IPv6Addresses []string `json:"ipv6_addresses"`
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
		return nil, fmt.Errorf("parse fastly public-ip-list json: %w", err)
	}
	if err := validateCIDRs("addresses", data.Addresses); err != nil {
		return nil, err
	}
	if err := validateCIDRs("ipv6_addresses", data.IPv6Addresses); err != nil {
		return nil, err
	}

	ipRanges := slices.Concat(data.Addresses, data.IPv6Addresses)
	if len(ipRanges) == 0 {
		return nil, fmt.Errorf("validate fastly public-ip-list json: no IP ranges found")
	}
	return ipRanges, nil
}

func validateCIDRs(field string, values []string) error {
	for i, value := range values {
		if !utils.IsCIDR(value) {
			return fmt.Errorf("validate fastly %s entry %d: %q is not a valid CIDR", field, i+1, value)
		}
	}
	return nil
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
			fmt.Printf("Fastly IP: %s\n", ip)
		} else {
			fmt.Println(ip)
		}
	}
}
