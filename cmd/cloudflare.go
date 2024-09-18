package cmd

import (
	"fmt"
	"os"

	"github.com/kaumnen/cipr/internal/cloudflare"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cloudflareCmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Get Cloudflare IP ranges",
	Long:  `Retrieve Cloudflare IPv4 and IPv6 ranges with optional verbosity levels.`,
	Run: func(cmd *cobra.Command, args []string) {
		var verbosity string
		if cmd.Flags().Changed("verbose-mode") {
			verbosity = viper.GetString("verbose_mode")
		} else if viper.GetBool("verbose") {
			verbosity = "full"
		} else {
			verbosity = "none"
		}

		if !isValidVerbosity(verbosity) {
			fmt.Fprintf(os.Stderr, "Invalid verbosity level: %s. Allowed values are: none, mini, full.\n", verbosity)
			os.Exit(1)
		}

		ipVersions := []string{}
		if viper.GetBool("cloudflare_ipv4") || (!viper.GetBool("cloudflare_ipv4") && !viper.GetBool("cloudflare_ipv6")) {
			ipVersions = append(ipVersions, "ipv4")
		}
		if viper.GetBool("cloudflare_ipv6") || (!viper.GetBool("cloudflare_ipv4") && !viper.GetBool("cloudflare_ipv6")) {
			ipVersions = append(ipVersions, "ipv6")
		}

		for _, version := range ipVersions {
			config := cloudflare.Config{
				IPType:    version,
				Verbosity: verbosity,
			}
			cloudflare.GetCloudflareIPRanges(config)
		}
	},
}

func init() {
	rootCmd.AddCommand(cloudflareCmd)

	cloudflareCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	cloudflareCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")

	viper.BindPFlag("cloudflare_ipv4", cloudflareCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("cloudflare_ipv6", cloudflareCmd.Flags().Lookup("ipv6"))
}
