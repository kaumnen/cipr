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

func GetIPRanges(ipType string, filterCountry, filterRegion, filterCity []string, verbosity string) {
	ip_ranges_data := loadData()
	readyIPs := filtrateIPRanges(ip_ranges_data, ipType, filterCountry, filterRegion, filterCity)
	printIPRanges(readyIPs, verbosity)
}

func loadData() []IPRange {
	raw_data := utils.GetRawData("https://mask-api.icloud.com/egress-ip-ranges.csv")
	r := csv.NewReader(strings.NewReader(raw_data))
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

func filtrateIPRanges(ipRanges []IPRange, ipType string, filterCountries, filterRegions, filterCities []string) []IPRange {
	var readyIPs []IPRange

	for _, ipRange := range ipRanges {
		if (ipType == "ipv4" && strings.Contains(ipRange.IPRange, ".")) || (ipType == "ipv6" && strings.Contains(ipRange.IPRange, ":")) {
			if (len(filterCountries) == 0 || containsIgnoreCase(filterCountries, ipRange.Country)) &&
				(len(filterRegions) == 0 || containsIgnoreCase(filterRegions, ipRange.Region)) &&
				(len(filterCities) == 0 || containsIgnoreCase(filterCities, ipRange.City)) {
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

func printIPRanges(IPranges []IPRange, verbosity string) {
	if len(IPranges) == 0 {
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

	for _, ip := range IPranges {
		printFunc(ip)
	}
}
