package digitalocean

import (
	"context"
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
	Zip     string
}

type Filters struct {
	Country []string
	Region  []string
	City    []string
	Zip     []string
}

type Config struct {
	Source    string
	IPType    string
	Filters   Filters
	List      string
	Verbosity string
}

func GetIPRanges(ctx context.Context, config Config) error {
	rawData, err := utils.GetRawData(ctx, config.Source)
	if err != nil {
		return err
	}
	ipRanges, err := parseRecords(rawData)
	if err != nil {
		return err
	}
	filtered := filtrateIPRanges(ipRanges, config)
	if config.List != "" {
		return printListedValues(filtered, config.List)
	}
	printIPRanges(filtered, config.Verbosity)
	return nil
}

func printListedValues(ranges []IPRange, dim string) error {
	var get func(IPRange) string
	switch dim {
	case "countries":
		get = func(r IPRange) string { return r.Country }
	case "regions":
		get = func(r IPRange) string { return r.Region }
	case "cities":
		get = func(r IPRange) string { return r.City }
	case "zips":
		get = func(r IPRange) string { return r.Zip }
	default:
		return fmt.Errorf("unknown list dimension %q (valid: countries, regions, cities, zips)", dim)
	}

	values := make([]string, 0, len(ranges))
	for _, r := range ranges {
		values = append(values, get(r))
	}
	values = utils.DedupeSorted(values)
	if len(values) == 0 {
		fmt.Println("No values to display.")
		return nil
	}
	for _, v := range values {
		fmt.Println(v)
	}
	return nil
}

func parseRecords(rawData string) ([]IPRange, error) {
	r := csv.NewReader(strings.NewReader(rawData))
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse digitalocean csv: %w", err)
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
	return ipRanges, nil
}

func filtrateIPRanges(ipRanges []IPRange, config Config) []IPRange {
	var readyIPs []IPRange

	for _, ipRange := range ipRanges {
		isV4 := utils.IsIPv4(ipRange.IPRange)
		isV6 := utils.IsIPv6(ipRange.IPRange)
		var familyMatch bool
		switch config.IPType {
		case "ipv4":
			familyMatch = isV4
		case "ipv6":
			familyMatch = isV6
		case "both":
			familyMatch = isV4 || isV6
		}
		if !familyMatch {
			continue
		}

		if (len(config.Filters.Country) == 0 || utils.ContainsIgnoreCase(config.Filters.Country, ipRange.Country)) &&
			(len(config.Filters.Region) == 0 || utils.ContainsIgnoreCase(config.Filters.Region, ipRange.Region)) &&
			(len(config.Filters.City) == 0 || utils.ContainsIgnoreCase(config.Filters.City, ipRange.City)) &&
			(len(config.Filters.Zip) == 0 || utils.ContainsIgnoreCase(config.Filters.Zip, ipRange.Zip)) {
			readyIPs = append(readyIPs, ipRange)
		}
	}
	return readyIPs
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
