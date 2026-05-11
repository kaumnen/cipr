package github

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
)

type IPRange struct {
	CIDR    string
	Service string
}

type Config struct {
	Source        string
	IPType        string
	FilterService string
	List          string
	Verbosity     string
}

// nonCIDRKeys are top-level keys in the /meta payload whose values are not
// CIDR arrays. They are skipped during parsing.
var nonCIDRKeys = map[string]struct{}{
	"verifiable_password_authentication": {},
	"ssh_key_fingerprints":               {},
	"ssh_keys":                           {},
	"domains":                            {},
}

func GetIPRanges(ctx context.Context, config Config) error {
	rawData, err := utils.GetRawData(ctx, config.Source)
	if err != nil {
		return err
	}

	ranges, err := parseMeta(rawData)
	if err != nil {
		return err
	}

	filtered := filtrate(ranges, config.IPType, config.FilterService)
	if config.List != "" {
		return printListedValues(filtered, config.List)
	}
	printIPRanges(filtered, config.Verbosity)
	return nil
}

func parseMeta(rawData string) ([]IPRange, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(rawData), &raw); err != nil {
		return nil, fmt.Errorf("parse github meta json: %w", err)
	}

	services := make([]string, 0, len(raw))
	for k := range raw {
		if _, skip := nonCIDRKeys[k]; skip {
			continue
		}
		services = append(services, k)
	}
	sort.Strings(services)

	var ranges []IPRange
	for _, service := range services {
		var cidrs []string
		if err := json.Unmarshal(raw[service], &cidrs); err != nil {
			continue
		}
		for _, cidr := range cidrs {
			ranges = append(ranges, IPRange{CIDR: cidr, Service: service})
		}
	}
	return ranges, nil
}

func filtrate(ranges []IPRange, ipType, filterService string) []IPRange {
	var result []IPRange
	for _, r := range ranges {
		if ipType == "ipv4" && !utils.IsIPv4(r.CIDR) {
			continue
		}
		if ipType == "ipv6" && !utils.IsIPv6(r.CIDR) {
			continue
		}
		if filterService != "" && !strings.EqualFold(r.Service, filterService) {
			continue
		}
		result = append(result, r)
	}
	return result
}

func printListedValues(ranges []IPRange, dim string) error {
	if dim != "services" {
		return fmt.Errorf("unknown list dimension %q (valid: services)", dim)
	}

	values := make([]string, 0, len(ranges))
	for _, r := range ranges {
		values = append(values, r.Service)
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

func printIPRanges(ranges []IPRange, verbosity string) {
	if len(ranges) == 0 {
		fmt.Println("No IP ranges to display.")
		return
	}

	var printFunc func(IPRange)
	switch verbosity {
	case "mini":
		printFunc = func(r IPRange) {
			fmt.Printf("%s,%s\n", r.CIDR, r.Service)
		}
	case "full":
		printFunc = func(r IPRange) {
			fmt.Printf("IP Prefix: %s, Service: %s\n", r.CIDR, r.Service)
		}
	default:
		printFunc = func(r IPRange) {
			fmt.Println(r.CIDR)
		}
	}

	for _, r := range ranges {
		printFunc(r)
	}
}
