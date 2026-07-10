package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
)

type Config struct {
	Source    string
	Verbosity string
}

func GetIPRanges(ctx context.Context, config Config) error {
	rawData, err := utils.GetRawData(ctx, config.Source)
	if err != nil {
		return err
	}
	ipRanges, err := parseIPRanges(rawData)
	if err != nil {
		return err
	}
	printIPRanges(ipRanges, config.Verbosity)
	return nil
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
		fmt.Println("No IP ranges to display.")
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
