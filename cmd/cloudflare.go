package cmd

import (
	"github.com/kaumnen/cipr/internal/cloudflare"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cloudflareIPv4Flag bool
	cloudflareIPv6Flag bool
)

var cloudflareCmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Get Cloudflare IP ranges",
	Long:  `Get Cloudflare IPv4 and IPv6 ranges.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetCiprLogger()

		logger.Println("cloudflare called")

		if cloudflareIPv4Flag || (!cloudflareIPv4Flag && !cloudflareIPv6Flag) {
			cloudflare.GetCloudflareIPv4Ranges()
		}
		if cloudflareIPv6Flag || (!cloudflareIPv4Flag && !cloudflareIPv6Flag) {
			cloudflare.GetCloudflareIPv6Ranges()
		}
	},
}

func init() {
	rootCmd.AddCommand(cloudflareCmd)

	cloudflareCmd.Flags().BoolVar(&cloudflareIPv4Flag, "ipv4", false, "Get only IPv4 ranges")
	cloudflareCmd.Flags().BoolVar(&cloudflareIPv6Flag, "ipv6", false, "Get only IPv6 ranges")
}
