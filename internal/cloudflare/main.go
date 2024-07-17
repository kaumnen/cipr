package cloudflare

import (
	"fmt"

	"github.com/kaumnen/cipr/internal/utils"
)

func GetCloudflareIPv4Ranges() {
	raw_data := utils.GetReq("https://www.cloudflare.com/ips-v4/")

	fmt.Println(raw_data)
}

func GetCloudflareIPv6Ranges() {
	raw_data := utils.GetReq("https://www.cloudflare.com/ips-v6/")

	fmt.Println(raw_data)
}
