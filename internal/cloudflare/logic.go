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
	printIPRanges(parseIPRanges(rawData), config.Verbosity)
	return nil
}

func parseIPRanges(rawData string) []string {
	lines := strings.Split(rawData, "\n")
	var ipRanges []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			ipRanges = append(ipRanges, trimmed)
		}
	}
	return ipRanges
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
