package cmd

import (
	"github.com/kaumnen/cipr/internal/cloudflare"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
)

// cloudflareCmd represents the cloudflare command
var cloudflareCmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Get Cloudflare ip ranges",
	Long:  `Get Cloudflare ip ranges`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetCiprLogger()

		logger.Println("cloudflare called")

		cloudflare.GetCloudflareIPv4Ranges()
		cloudflare.GetCloudflareIPv6Ranges()
	},
}

func init() {
	rootCmd.AddCommand(cloudflareCmd)
}
