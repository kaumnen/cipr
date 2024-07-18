package cmd

import (
	"github.com/kaumnen/cipr/internal/cloudflare"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
)

var (
	ipv4Flag bool
	ipv6Flag bool
)

var cloudflareCmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Get Cloudflare IP ranges",
	Long:  `Get Cloudflare IPv4 and IPv6 ranges.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetCiprLogger()

		logger.Println("cloudflare called")

		if ipv4Flag || (!ipv4Flag && !ipv6Flag) {
			cloudflare.GetCloudflareIPv4Ranges()
		}
		if ipv6Flag || (!ipv4Flag && !ipv6Flag) {
			cloudflare.GetCloudflareIPv6Ranges()
		}
	},
}

func init() {
	rootCmd.AddCommand(cloudflareCmd)

	cloudflareCmd.Flags().BoolVar(&ipv4Flag, "ipv4", false, "Get only IPv4 ranges")
	cloudflareCmd.Flags().BoolVar(&ipv6Flag, "ipv6", false, "Get only IPv6 ranges")
}
