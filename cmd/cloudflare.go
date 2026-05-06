package cmd

import (
	"github.com/kaumnen/cipr/internal/cloudflare"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cloudflareCmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Get Cloudflare IP ranges",
	Long:  `Retrieve Cloudflare IPv4 and IPv6 ranges with optional verbosity levels.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity := resolveVerbosity(cmd)

		ipv4 := viper.GetBool("cloudflare_ipv4")
		ipv6 := viper.GetBool("cloudflare_ipv6")
		both := !ipv4 && !ipv6
		var ipVersions []string
		if ipv4 || both {
			ipVersions = append(ipVersions, "ipv4")
		}
		if ipv6 || both {
			ipVersions = append(ipVersions, "ipv6")
		}

		for _, version := range ipVersions {
			source := utils.ResolveSource("cloudflare_" + version)
			if err := cloudflare.GetIPRanges(cmd.Context(), cloudflare.Config{
				Source:    source,
				Verbosity: verbosity,
			}); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloudflareCmd)

	cloudflareCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	cloudflareCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")

	viper.BindPFlag("cloudflare_ipv4", cloudflareCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("cloudflare_ipv6", cloudflareCmd.Flags().Lookup("ipv6"))
}
