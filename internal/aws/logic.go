package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
)

type IPv4Prefix struct {
	IPAddress          string `json:"ip_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

type IPv6Prefix struct {
	IPv6Address        string `json:"ipv6_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

type IPsData struct {
	SyncToken    string       `json:"syncToken"`
	CreateDate   string       `json:"createDate"`
	Prefixes     []IPv4Prefix `json:"prefixes"`
	IPv6Prefixes []IPv6Prefix `json:"ipv6_prefixes"`
}

type IPPrefix interface {
	GetIPAddress() string
	GetRegion() string
	GetService() string
	GetNetworkBorderGroup() string
}

func (p IPv4Prefix) GetIPAddress() string          { return p.IPAddress }
func (p IPv4Prefix) GetRegion() string             { return p.Region }
func (p IPv4Prefix) GetService() string            { return p.Service }
func (p IPv4Prefix) GetNetworkBorderGroup() string { return p.NetworkBorderGroup }

func (p IPv6Prefix) GetIPAddress() string          { return p.IPv6Address }
func (p IPv6Prefix) GetRegion() string             { return p.Region }
func (p IPv6Prefix) GetService() string            { return p.Service }
func (p IPv6Prefix) GetNetworkBorderGroup() string { return p.NetworkBorderGroup }

type Config struct {
	Source    string
	IPType    string
	Filter    string
	Filters   Filters
	List      string
	Verbosity string
}

type Filters struct {
	Region             []string
	Service            []string
	NetworkBorderGroup []string
}

func GetIPRanges(ctx context.Context, config Config) error {
	rawData, err := utils.GetRawData(ctx, config.Source)
	if err != nil {
		return err
	}

	filters := config.Filters
	if config.Filter != "" {
		filters = filtersFromComposite(config.Filter)
	}
	readyIPs, err := filtrateIPRanges(rawData, config.IPType, filters)
	if err != nil {
		return err
	}
	if config.List != "" {
		return printListedValues(readyIPs, config.List)
	}
	printIPRanges(readyIPs, config.Verbosity)
	return nil
}

func printListedValues(prefixes []IPPrefix, dim string) error {
	var get func(IPPrefix) string
	switch dim {
	case "regions":
		get = IPPrefix.GetRegion
	case "services":
		get = IPPrefix.GetService
	case "network-border-groups":
		get = IPPrefix.GetNetworkBorderGroup
	default:
		return fmt.Errorf("unknown list dimension %q (valid: regions, services, network-border-groups)", dim)
	}

	values := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		values = append(values, get(p))
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

func separateFilters(filterFlagValues string) []string {
	var filterSlice []string

	removeFilterWhitespace := strings.ReplaceAll(filterFlagValues, " ", "")
	filterContents := strings.Split(removeFilterWhitespace, ",")

	for _, val := range filterContents {
		if val == "" {
			filterSlice = append(filterSlice, "*")
		} else {
			filterSlice = append(filterSlice, val)
		}
	}

	for len(filterSlice) < 3 {
		filterSlice = append(filterSlice, "*")
	}

	if len(filterSlice) > 3 {
		filterSlice = filterSlice[:3]
	}

	return filterSlice
}

func filtersFromComposite(filter string) Filters {
	parts := separateFilters(filter)
	return Filters{
		Region:             wildcardToEmpty(parts[0]),
		Service:            wildcardToEmpty(parts[1]),
		NetworkBorderGroup: wildcardToEmpty(parts[2]),
	}
}

func wildcardToEmpty(value string) []string {
	if value == "*" {
		return nil
	}
	return []string{value}
}

func filtrateIPRanges(rawData, ipType string, filters Filters) ([]IPPrefix, error) {
	var data IPsData
	if err := json.Unmarshal([]byte(rawData), &data); err != nil {
		return nil, fmt.Errorf("parse aws ip-ranges json: %w", err)
	}
	if len(data.Prefixes) == 0 && len(data.IPv6Prefixes) == 0 {
		return nil, fmt.Errorf("validate aws ip-ranges json: no IP ranges found")
	}
	for i, prefix := range data.Prefixes {
		if !utils.IsCIDR(prefix.IPAddress) {
			return nil, fmt.Errorf("validate aws IPv4 prefix %d: %q is not a valid CIDR", i+1, prefix.IPAddress)
		}
	}
	for i, prefix := range data.IPv6Prefixes {
		if !utils.IsCIDR(prefix.IPv6Address) {
			return nil, fmt.Errorf("validate aws IPv6 prefix %d: %q is not a valid CIDR", i+1, prefix.IPv6Address)
		}
	}

	var prefixes []IPPrefix
	if ipType == "ipv4" || ipType == "both" || ipType == "" {
		for _, prefix := range data.Prefixes {
			prefixes = append(prefixes, prefix)
		}
	}
	if ipType == "ipv6" || ipType == "both" || ipType == "" {
		for _, prefix := range data.IPv6Prefixes {
			prefixes = append(prefixes, prefix)
		}
	}

	var result []IPPrefix
	for _, prefix := range prefixes {
		if matchesFilter(prefix, filters) {
			result = append(result, prefix)
		}
	}
	return result, nil
}

func matchesFilter(prefix IPPrefix, filters Filters) bool {
	return (len(filters.Region) == 0 || utils.ContainsIgnoreCase(filters.Region, prefix.GetRegion())) &&
		(len(filters.Service) == 0 || utils.ContainsIgnoreCase(filters.Service, prefix.GetService())) &&
		(len(filters.NetworkBorderGroup) == 0 || utils.ContainsIgnoreCase(filters.NetworkBorderGroup, prefix.GetNetworkBorderGroup()))
}

func printIPRanges(ipRanges []IPPrefix, verbosity string) {
	if len(ipRanges) == 0 {
		fmt.Println("No IP ranges to display.")
		return
	}

	var printFunc func(IPPrefix)

	switch verbosity {
	case "none":
		printFunc = func(ip IPPrefix) {
			fmt.Println(ip.GetIPAddress())
		}
	case "mini":
		printFunc = func(ip IPPrefix) {
			fmt.Printf("%s,%s,%s,%s\n",
				ip.GetIPAddress(), ip.GetRegion(), ip.GetService(), ip.GetNetworkBorderGroup())
		}
	case "full":
		printFunc = func(ip IPPrefix) {
			fmt.Printf("IP Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
				ip.GetIPAddress(), ip.GetRegion(), ip.GetService(), ip.GetNetworkBorderGroup())
		}
	default:
		printFunc = func(ip IPPrefix) {
			fmt.Println(ip.GetIPAddress())
		}
	}

	for _, ip := range ipRanges {
		printFunc(ip)
	}
}
