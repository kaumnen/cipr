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
	State   string
	City    string
}

func GetIPRanges(ipType string, filterCountry, filterState, filterCity []string) {
	ip_ranges_data := loadData()
	readyIPs := filtrateIPRanges(ip_ranges_data, ipType, filterCountry, filterState, filterCity)
	fmt.Println(readyIPs)
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
			State:   record[2],
			City:    record[3],
		}
		ipRanges = append(ipRanges, ipRange)
	}
	return ipRanges
}

func filtrateIPRanges(ipRanges []IPRange, ipType string, filterCountries, filterStates, filterCities []string) []IPRange {
	var readyIPs []IPRange

	for _, ipRange := range ipRanges {
		if (ipType == "ipv4" && strings.Contains(ipRange.IPRange, ".")) || (ipType == "ipv6" && strings.Contains(ipRange.IPRange, ":")) {
			if (len(filterCountries) == 0 || containsIgnoreCase(filterCountries, ipRange.Country)) &&
				(len(filterStates) == 0 || containsIgnoreCase(filterStates, ipRange.State)) &&
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
