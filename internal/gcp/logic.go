package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
)

type sourcePrefix struct {
	IPv4Prefix string `json:"ipv4Prefix"`
	IPv6Prefix string `json:"ipv6Prefix"`
	Service    string `json:"service"`
	Scope      string `json:"scope"`
}

type IPsData struct {
	SyncToken    string         `json:"syncToken"`
	CreationTime string         `json:"creationTime"`
	Prefixes     []sourcePrefix `json:"prefixes"`
}

type Prefix struct {
	Address string
	Scope   string
	Service string
}

type Config struct {
	Source    string
	IPType    string
	Filter    string
	List      string
	Verbosity string
}

func GetIPRanges(ctx context.Context, config Config) error {
	rawData, err := utils.GetRawData(ctx, config.Source)
	if err != nil {
		return err
	}

	ipType := config.IPType
	if config.List != "" {
		ipType = "both"
	}
	prefixes, err := filtrateIPRanges(rawData, ipType, separateFilters(config.Filter))
	if err != nil {
		return err
	}
	if config.List != "" {
		return printListedValues(prefixes, config.List)
	}
	printIPRanges(prefixes, config.Verbosity)
	return nil
}

func separateFilters(filterFlagValues string) []string {
	values := strings.Split(filterFlagValues, ",")
	filters := make([]string, 0, 2)
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			value = "*"
		}
		filters = append(filters, value)
	}
	for len(filters) < 2 {
		filters = append(filters, "*")
	}
	return filters[:2]
}

func filtrateIPRanges(rawData, ipType string, filterSlice []string) ([]Prefix, error) {
	var data IPsData
	if err := json.Unmarshal([]byte(rawData), &data); err != nil {
		return nil, fmt.Errorf("parse gcp cloud ip-ranges json: %w", err)
	}
	if len(data.Prefixes) == 0 {
		return nil, fmt.Errorf("validate gcp cloud ip-ranges json: no IP ranges found")
	}

	var result []Prefix
	for i, source := range data.Prefixes {
		if source.IPv4Prefix == "" && source.IPv6Prefix == "" {
			return nil, fmt.Errorf("validate gcp prefix %d: missing ipv4Prefix or ipv6Prefix", i+1)
		}
		if source.IPv4Prefix != "" && source.IPv6Prefix != "" {
			return nil, fmt.Errorf("validate gcp prefix %d: both ipv4Prefix and ipv6Prefix are set", i+1)
		}

		address := source.IPv4Prefix
		family := "IPv4"
		if address == "" {
			address = source.IPv6Prefix
			family = "IPv6"
		}
		if !utils.IsCIDR(address) {
			return nil, fmt.Errorf("validate gcp %s prefix %d: %q is not a valid CIDR", family, i+1, address)
		}
		wrongFamily := (family == "IPv4" && !utils.IsIPv4(address)) ||
			(family == "IPv6" && !utils.IsIPv6(address))
		if wrongFamily {
			return nil, fmt.Errorf("validate gcp %s prefix %d: %q has the wrong address family", family, i+1, address)
		}
		if !matchesFilter(source, filterSlice) || !ipVersionMatches(address, ipType) {
			continue
		}
		result = append(result, Prefix{Address: address, Scope: source.Scope, Service: source.Service})
	}
	return result, nil
}

func matchesFilter(prefix sourcePrefix, filterSlice []string) bool {
	return (filterSlice[0] == "*" || strings.EqualFold(prefix.Scope, filterSlice[0])) &&
		(filterSlice[1] == "*" || strings.EqualFold(prefix.Service, filterSlice[1]))
}

func ipVersionMatches(address, ipType string) bool {
	switch ipType {
	case "ipv4":
		return utils.IsIPv4(address)
	case "ipv6":
		return utils.IsIPv6(address)
	default:
		return true
	}
}

func printListedValues(prefixes []Prefix, dimension string) error {
	values := make([]string, 0, len(prefixes))
	switch dimension {
	case "scopes":
		for _, prefix := range prefixes {
			values = append(values, prefix.Scope)
		}
	case "services":
		for _, prefix := range prefixes {
			values = append(values, prefix.Service)
		}
	default:
		return fmt.Errorf("unknown list dimension %q (valid: scopes, services)", dimension)
	}

	values = utils.DedupeSorted(values)
	if len(values) == 0 {
		fmt.Println("No values to display.")
		return nil
	}
	for _, value := range values {
		fmt.Println(value)
	}
	return nil
}

func printIPRanges(prefixes []Prefix, verbosity string) {
	if len(prefixes) == 0 {
		fmt.Println("No IP ranges to display.")
		return
	}

	for _, prefix := range prefixes {
		switch verbosity {
		case "mini":
			fmt.Printf("%s,%s,%s\n", prefix.Address, prefix.Scope, prefix.Service)
		case "full":
			fmt.Printf("IP Prefix: %s, Scope: %s, Service: %s\n", prefix.Address, prefix.Scope, prefix.Service)
		default:
			fmt.Println(prefix.Address)
		}
	}
}
