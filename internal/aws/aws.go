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

func GetIPRanges(ipType string, filter string) {
	raw_data := utils.GetReq("https://ip-ranges.amazonaws.com/ip-ranges.json")

	filterValues := separateFilters(filter, false)

	printIPRanges(raw_data, ipType, filterValues)
}

func separateFilters(filterFlagValues string, returnSlice bool) []string {
	var filterSlice []string

	removeFilterWhitespace := strings.ReplaceAll(filterFlagValues, " ", "")
	filterContents := strings.Split(removeFilterWhitespace, ",")

	for _, val := range filterContents {
		filterSlice = append(filterSlice, strings.TrimSpace(val))
	}

	return filterSlice
}

func printIPRanges(rawData, ipType string, filterSlice []string) {
	logger := utils.GetCiprLogger()
	var data IPsData

	err := json.Unmarshal([]byte(rawData), &data)
	if err != nil {
		logger.Fatalf("Error unmarshalling JSON: %v", err)
	}

	fmt.Printf("Sync Token: %s\n", data.SyncToken)
	fmt.Printf("Create Date: %s\n", data.CreateDate)

	if ipType == "ipv4" {
		fmt.Println("Prefixes:")
		for _, prefix := range data.Prefixes {
			switch len(filterSlice) {
			case 1:
				if prefix.Region == filterSlice[0] {
					fmt.Printf("  IP Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
						prefix.IPAddress, prefix.Region, prefix.Service, prefix.NetworkBorderGroup)
				}
			case 2:
				if prefix.Region == filterSlice[0] && prefix.Service == filterSlice[1] {
					fmt.Printf("  IP Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
						prefix.IPAddress, prefix.Region, prefix.Service, prefix.NetworkBorderGroup)
				}
			case 3:
				if prefix.Region == filterSlice[0] && prefix.Service == filterSlice[1] && prefix.NetworkBorderGroup == filterSlice[2] {
					fmt.Printf("  IP Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
						prefix.IPAddress, prefix.Region, prefix.Service, prefix.NetworkBorderGroup)
				}
			default:
				fmt.Println("Nothing found!")
				return
			}
		}
	} else if ipType == "ipv6" {
		fmt.Println("IPv6 Prefixes:")
		for _, ipv6prefix := range data.IPv6Prefixes {
			switch len(filterSlice) {
			case 1:
				if ipv6prefix.Region == filterSlice[0] {
					fmt.Printf("  IPv6 Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
						ipv6prefix.IPv6Address, ipv6prefix.Region, ipv6prefix.Service, ipv6prefix.NetworkBorderGroup)
				}
			case 2:
				if ipv6prefix.Region == filterSlice[0] && ipv6prefix.Service == filterSlice[1] {
					fmt.Printf("  IPv6 Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
						ipv6prefix.IPv6Address, ipv6prefix.Region, ipv6prefix.Service, ipv6prefix.NetworkBorderGroup)
				}
			case 3:
				if ipv6prefix.Region == filterSlice[0] && ipv6prefix.Service == filterSlice[1] && ipv6prefix.NetworkBorderGroup == filterSlice[2] {
					fmt.Printf("  IPv6 Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
						ipv6prefix.IPv6Address, ipv6prefix.Region, ipv6prefix.Service, ipv6prefix.NetworkBorderGroup)
				}
			default:
				fmt.Println("Nothing found!")
				return
			}
		}
	}
}
