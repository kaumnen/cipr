package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"

	"github.com/kaumnen/cipr/internal/utils"
)

type IPRange struct {
	CIDR    string
	Service string
}

type Config struct {
	Source         string
	IPType         string
	FilterServices []string
	List           string
	Verbosity      string
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

	filtered := filtrate(ranges, config.IPType, config.FilterServices)
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
		services = append(services, k)
	}
	sort.Strings(services)

	// /meta also holds keys and other metadata. A service is a list
	// with at least one CIDR.
	var ranges []IPRange
	for _, service := range services {
		var cidrs []string
		if err := json.Unmarshal(raw[service], &cidrs); err != nil {
			continue
		}
		if !slices.ContainsFunc(cidrs, utils.IsCIDR) {
			continue
		}
		for i, cidr := range cidrs {
			if !utils.IsCIDR(cidr) {
				return nil, fmt.Errorf("validate github service %q prefix %d: %q is not a valid CIDR", service, i+1, cidr)
			}
			ranges = append(ranges, IPRange{CIDR: cidr, Service: service})
		}
	}
	if len(ranges) == 0 {
		return nil, fmt.Errorf("validate github meta json: no IP ranges found")
	}
	return ranges, nil
}

func filtrate(ranges []IPRange, ipType string, filterServices []string) []IPRange {
	var result []IPRange
	for _, r := range ranges {
		if ipType == "ipv4" && !utils.IsIPv4(r.CIDR) {
			continue
		}
		if ipType == "ipv6" && !utils.IsIPv6(r.CIDR) {
			continue
		}
		if len(filterServices) > 0 && !utils.ContainsIgnoreCase(filterServices, r.Service) {
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
		fmt.Fprintln(os.Stderr, "No values to display.")
		return nil
	}
	for _, v := range values {
		fmt.Println(v)
	}
	return nil
}

func printIPRanges(ranges []IPRange, verbosity string) {
	if len(ranges) == 0 {
		fmt.Fprintln(os.Stderr, "No IP ranges to display.")
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
		// CIDRs repeat across services.
		seen := make(map[string]struct{}, len(ranges))
		printFunc = func(r IPRange) {
			if _, dup := seen[r.CIDR]; dup {
				return
			}
			seen[r.CIDR] = struct{}{}
			fmt.Println(r.CIDR)
		}
	}

	for _, r := range ranges {
		printFunc(r)
	}
}
