package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
)

type Item struct {
	CIDR      string   `json:"cidr"`
	Region    []string `json:"region"`
	Product   []string `json:"product"`
	Direction []string `json:"direction"`
	Perimeter string   `json:"perimeter"`
}

type IPsData struct {
	CreationDate string `json:"creationDate"`
	Items        []Item `json:"items"`
}

type Config struct {
	Source    string
	IPType    string
	Filters   Filters
	List      string
	Verbosity string
}

type Filters struct {
	Product   []string
	Direction []string
	Region    []string
	Perimeter []string
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
	items, err := filtrateIPRanges(rawData, ipType, config.Filters)
	if err != nil {
		utils.DropCache(config.Source)
		return err
	}
	if config.List != "" {
		return printListedValues(items, config.List)
	}
	printIPRanges(items, config.Verbosity)
	return nil
}

func filtrateIPRanges(rawData, ipType string, filters Filters) ([]Item, error) {
	var data IPsData
	if err := json.Unmarshal([]byte(rawData), &data); err != nil {
		return nil, fmt.Errorf("parse atlassian ip-ranges json: %w", err)
	}
	if len(data.Items) == 0 {
		return nil, fmt.Errorf("validate atlassian ip-ranges json: no IP ranges found")
	}

	var result []Item
	for i, item := range data.Items {
		if !utils.IsCIDR(item.CIDR) {
			return nil, fmt.Errorf("validate atlassian item %d: %q is not a valid CIDR", i+1, item.CIDR)
		}
		if !matchesFilter(item, filters) || !ipVersionMatches(item.CIDR, ipType) {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func matchesFilter(item Item, filters Filters) bool {
	return anyMatches(filters.Product, item.Product) &&
		anyMatches(filters.Direction, item.Direction) &&
		anyMatches(filters.Region, item.Region) &&
		anyMatches(filters.Perimeter, []string{item.Perimeter})
}

// anyMatches reports whether any item value is in the filter. An empty filter matches all.
func anyMatches(filter, values []string) bool {
	if len(filter) == 0 {
		return true
	}
	for _, value := range values {
		if utils.ContainsIgnoreCase(filter, value) {
			return true
		}
	}
	return false
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

func printListedValues(items []Item, dimension string) error {
	var get func(Item) []string
	switch dimension {
	case "products":
		get = func(item Item) []string { return item.Product }
	case "directions":
		get = func(item Item) []string { return item.Direction }
	case "regions":
		get = func(item Item) []string { return item.Region }
	case "perimeters":
		get = func(item Item) []string { return []string{item.Perimeter} }
	default:
		return fmt.Errorf("unknown list dimension %q (valid: products, directions, regions, perimeters)", dimension)
	}

	var values []string
	for _, item := range items {
		values = append(values, get(item)...)
	}
	values = utils.DedupeSorted(values)
	if len(values) == 0 {
		fmt.Fprintln(os.Stderr, "No values to display.")
		return nil
	}
	for _, value := range values {
		fmt.Println(value)
	}
	return nil
}

func printIPRanges(items []Item, verbosity string) {
	if len(items) == 0 {
		fmt.Fprintln(os.Stderr, "No IP ranges to display.")
		return
	}

	var printFunc func(Item)
	switch verbosity {
	case "mini":
		printFunc = func(item Item) {
			fmt.Printf("%s,%s,%s,%s,%s\n", item.CIDR, join(item.Region), join(item.Product), join(item.Direction), item.Perimeter)
		}
	case "full":
		printFunc = func(item Item) {
			fmt.Printf("IP Prefix: %s, Region: %s, Product: %s, Direction: %s, Perimeter: %s\n",
				item.CIDR, join(item.Region), join(item.Product), join(item.Direction), item.Perimeter)
		}
	default:
		seen := make(map[string]struct{}, len(items))
		printFunc = func(item Item) {
			if _, dup := seen[item.CIDR]; dup {
				return
			}
			seen[item.CIDR] = struct{}{}
			fmt.Println(item.CIDR)
		}
	}

	for _, item := range items {
		printFunc(item)
	}
}

func join(values []string) string {
	return strings.Join(values, ";")
}
