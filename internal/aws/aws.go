package aws

import (
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

func GetIPRanges(ipType, filter, verbosity string, getReqFunc func(string) string) {
	raw_data := getReqFunc("https://ip-ranges.amazonaws.com/ip-ranges.json")

	filterValues := separateFilters(filter)

	readyIPs := filtrateIPRanges(raw_data, ipType, filterValues)

	printIPRanges(readyIPs, verbosity)
}

func separateFilters(filterFlagValues string) []string {
	logger := utils.GetCiprLogger()
	var filterSlice []string

	removeFilterWhitespace := strings.ReplaceAll(filterFlagValues, " ", "")
	filterContents := strings.Split(removeFilterWhitespace, ",")

	for _, val := range filterContents {
		if len(val) > 0 {
			filterSlice = append(filterSlice, strings.TrimSpace(val))
		}
	}

	if len(filterSlice) == 0 && strings.Contains(filterFlagValues, ",") {
		logger.Fatalf("Filter flag needs actual values!")
	}

	return filterSlice
}

func filtrateIPRanges(rawData, ipType string, filterSlice []string) []IPPrefix {
	logger := utils.GetCiprLogger()
	var data IPsData

	err := json.Unmarshal([]byte(rawData), &data)
	if err != nil {
		logger.Fatalf("Error unmarshalling JSON: %v", err)
	}

	var prefixes []IPPrefix
	result := []IPPrefix{}

	if ipType == "ipv4" {
		for _, prefix := range data.Prefixes {
			prefixes = append(prefixes, prefix)
		}
	} else if ipType == "ipv6" {
		for _, prefix := range data.IPv6Prefixes {
			prefixes = append(prefixes, prefix)
		}
	}

	for _, prefix := range prefixes {
		if matchesFilter(prefix, filterSlice) {
			result = append(result, prefix)
		}
	}

	if len(result) == 0 {
		fmt.Println("Nothing found!")
		return nil
	}

	return result
}

func matchesFilter(prefix IPPrefix, filterSlice []string) bool {
	switch len(filterSlice) {
	case 0:
		return true
	case 1:
		return prefix.GetRegion() == filterSlice[0]
	case 2:
		return prefix.GetRegion() == filterSlice[0] && prefix.GetService() == filterSlice[1]
	case 3:
		return prefix.GetRegion() == filterSlice[0] && prefix.GetService() == filterSlice[1] && prefix.GetNetworkBorderGroup() == filterSlice[2]
	default:
		return false
	}
}

func printIPRanges(IPranges []IPPrefix, verbosity string) {
	if len(IPranges) == 0 {
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

	for _, ip := range IPranges {
		printFunc(ip)
	}
}
