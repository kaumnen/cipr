package aws

import (
	"encoding/json"
	"log"

	"github.com/kaumnen/cipr/internal/utils"
)

type IPv4Prefix struct {
	IPAddress          string `json:"ip_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

// IPv6Prefix represents the structure of each IPv6 prefix.
type IPv6Prefix struct {
	IPv6Address        string `json:"ipv6_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

// Data represents the entire JSON structure.
type IPsData struct {
	SyncToken    string       `json:"syncToken"`
	CreateDate   string       `json:"createDate"`
	Prefixes     []IPv4Prefix `json:"prefixes"`
	IPv6Prefixes []IPv6Prefix `json:"ipv6_prefixes"`
}

func GetIPRanges(ipType string) {
	logger := utils.GetCiprLogger()
	raw_data := utils.GetReq("https://ip-ranges.amazonaws.com/ip-ranges.json")

	var data IPsData

	err := json.Unmarshal([]byte(raw_data), &data)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	logger.Printf("Sync Token: %s\n", data.SyncToken)
	logger.Printf("Create Date: %s\n", data.CreateDate)

	if ipType == "ipv4" {
		logger.Println("Prefixes:")
		for _, prefix := range data.Prefixes {
			logger.Printf("  IP Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
				prefix.IPAddress, prefix.Region, prefix.Service, prefix.NetworkBorderGroup)
		}
	} else if ipType == "ipv6" {
		logger.Println("IPv6 Prefixes:")
		for _, ipv6prefix := range data.IPv6Prefixes {
			logger.Printf("  IPv6 Prefix: %s, Region: %s, Service: %s, Network Border Group: %s\n",
				ipv6prefix.IPv6Address, ipv6prefix.Region, ipv6prefix.Service, ipv6prefix.NetworkBorderGroup)
		}
	}
}
