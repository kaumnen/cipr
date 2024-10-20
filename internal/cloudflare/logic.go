package cloudflare

import (
	"fmt"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/viper"
)

type Config struct {
	IPType    string
	Verbosity string
}

func GetCloudflareIPRanges(config Config) {
	var rawData string
	ipRangesSource := viper.GetString("source")

	if ipRangesSource == "hosted" {
		if config.IPType == "ipv4" {
			rawData = utils.GetRawData("cloudflare_ipv4")
		} else if config.IPType == "ipv6" {
			rawData = utils.GetRawData("cloudflare_ipv6")
		}
	} else {
		rawData = utils.GetRawData(ipRangesSource)
	}

	ipRanges := parseIPRanges(rawData)

	printCloudflareIPRanges(ipRanges, config.Verbosity)
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

func printCloudflareIPRanges(ipRanges []string, verbosity string) {
	if len(ipRanges) == 0 {
		fmt.Println("No IP ranges to display.")
		return
	}

	switch verbosity {
	case "none":
		for _, ip := range ipRanges {
			fmt.Println(ip)
		}
	case "mini":
		for _, ip := range ipRanges {
			fmt.Println(ip)
		}
	case "full":
		for _, ip := range ipRanges {
			fmt.Printf("Cloudflare IP: %s\n", ip)
		}
	default:
		for _, ip := range ipRanges {
			fmt.Println(ip)
		}
	}
}
