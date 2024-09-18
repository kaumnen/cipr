package icloud

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
)

type IPRange struct {
	IPRange string
	Country string
	Region  string
	City    string
}

type Filters struct {
	Country []string
	Region  []string
	City    []string
}

type Config struct {
	IPType    string
	Filters   Filters
	Verbosity string
}

func GetIPRanges(config Config) {
	ipRangesData := loadData()
	readyIPs := filtrateIPRanges(ipRangesData, config)
	printIPRanges(readyIPs, config.Verbosity)
}

func loadData() []IPRange {
	rawData := utils.GetRawData("https://mask-api.icloud.com/egress-ip-ranges.csv")
	r := csv.NewReader(strings.NewReader(rawData))
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	var ipRanges []IPRange
	for _, record := range records {
		if len(record) < 4 {
			continue
		}
		ipRange := IPRange{
			IPRange: record[0],
			Country: record[1],
			Region:  record[2],
			City:    record[3],
		}
		ipRanges = append(ipRanges, ipRange)
	}
	return ipRanges
}

func filtrateIPRanges(ipRanges []IPRange, config Config) []IPRange {
	var readyIPs []IPRange

	for _, ipRange := range ipRanges {
		if (config.IPType == "ipv4" && strings.Contains(ipRange.IPRange, ".")) ||
			(config.IPType == "ipv6" && strings.Contains(ipRange.IPRange, ":")) ||
			(config.IPType == "both" && (strings.Contains(ipRange.IPRange, ".") || strings.Contains(ipRange.IPRange, ":"))) {

			if (len(config.Filters.Country) == 0 || containsIgnoreCase(config.Filters.Country, ipRange.Country)) &&
				(len(config.Filters.Region) == 0 || containsIgnoreCase(config.Filters.Region, ipRange.Region)) &&
				(len(config.Filters.City) == 0 || containsIgnoreCase(config.Filters.City, ipRange.City)) {
				readyIPs = append(readyIPs, ipRange)
			}
		}
	}
	return readyIPs
}

func containsIgnoreCase(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func printIPRanges(ipRanges []IPRange, verbosity string) {
	if len(ipRanges) == 0 {
		fmt.Println("No IP ranges to display.")
		return
	}

	var printFunc func(IPRange)

	switch verbosity {
	case "none":
		printFunc = func(ip IPRange) {
			fmt.Println(ip.IPRange)
		}
	case "mini":
		printFunc = func(ip IPRange) {
			fmt.Printf("%s,%s,%s,%s\n",
				ip.IPRange, ip.Country, ip.Region, ip.City)
		}
	case "full":
		printFunc = func(ip IPRange) {
			fmt.Printf("IP Range: %s, Country: %s, Region: %s, City: %s\n",
				ip.IPRange, ip.Country, ip.Region, ip.City)
		}
	default:
		printFunc = func(ip IPRange) {
			fmt.Println(ip.IPRange)
		}
	}

	for _, ip := range ipRanges {
		printFunc(ip)
	}
}
