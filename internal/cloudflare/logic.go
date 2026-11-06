package cloudflare

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
)

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

func parseIPRanges(rawData string) ([]string, error) {
	lines := strings.Split(rawData, "\n")
	var ipRanges []string
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			if !utils.IsCIDR(trimmed) {
				return nil, fmt.Errorf("validate cloudflare line %d: %q is not a valid CIDR", i+1, trimmed)
			}
			ipRanges = append(ipRanges, trimmed)
		}
	}
	if len(ipRanges) == 0 {
		return nil, fmt.Errorf("validate cloudflare data: no IP ranges found")
	}
	return ipRanges, nil
}

func printIPRanges(ipRanges []string, verbosity string) {
	if len(ipRanges) == 0 {
		fmt.Fprintln(os.Stderr, "No IP ranges to display.")
		return
	}

	for _, ip := range ipRanges {
		if verbosity == "full" {
			fmt.Printf("Cloudflare IP: %s\n", ip)
		} else {
			fmt.Println(ip)
		}
	}
}
