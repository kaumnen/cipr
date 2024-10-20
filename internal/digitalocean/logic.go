package digitalocean

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/viper"
)

type IPRange struct {
	IPRange string
	Country string
	Region  string
	City    string
	Zip     string
}

type Filters struct {
	Country []string
	Region  []string
	City    []string
	Zip     []string
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
	var rawData string
	ipRangesSource := viper.GetString("source")

	if ipRangesSource == "hosted" {
		rawData = utils.GetRawData("digitalocean")
	} else {
		rawData = utils.GetRawData(ipRangesSource)
	}

	r := csv.NewReader(strings.NewReader(rawData))
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()

	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	var ipRanges []IPRange
	for _, record := range records {

		ipRange := IPRange{}

		if len(record) > 0 {
			ipRange.IPRange = strings.TrimSpace(record[0])
		}
		if len(record) > 1 {
			ipRange.Country = strings.TrimSpace(record[1])
		}
		if len(record) > 2 {
			ipRange.Region = strings.TrimSpace(record[2])
		}
		if len(record) > 3 {
			ipRange.City = strings.TrimSpace(record[3])
		}
		if len(record) > 4 {
			ipRange.Zip = strings.TrimSpace(record[4])
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
				(len(config.Filters.City) == 0 || containsIgnoreCase(config.Filters.City, ipRange.City)) &&
				(len(config.Filters.Zip) == 0 || containsIgnoreCase(config.Filters.Zip, ipRange.Zip)) {
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
			fmt.Printf("%s,%s,%s,%s,%s\n",
				ip.IPRange, ip.Country, ip.Region, ip.City, ip.Zip)
		}
	case "full":
		printFunc = func(ip IPRange) {
			fmt.Printf("IP Range: %s, Country: %s, Region: %s, City: %s, ZIP: %s\n",
				ip.IPRange, ip.Country, ip.Region, ip.City, ip.Zip)
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
