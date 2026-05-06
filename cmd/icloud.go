package cmd

import (
	"github.com/kaumnen/cipr/internal/icloud"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var icloudCmd = &cobra.Command{
	Use:   "icloud",
	Short: "Get iCloud private relay IP ranges.",
	Long:  `Get iCloud private relay IPv4 and IPv6 ranges.`,
	Run: func(cmd *cobra.Command, args []string) {
		verbosity := resolveVerbosity(cmd)

		ipType := "ipv4"
		if viper.GetBool("icloud_ipv6") {
			ipType = "ipv6"
		}

		if !viper.GetBool("icloud_ipv4") && !viper.GetBool("icloud_ipv6") {
			ipType = "both"
		}

		config := icloud.Config{
			IPType: ipType,
			Filters: icloud.Filters{
				Country: viper.GetStringSlice("icloud-filter-country"),
				Region:  viper.GetStringSlice("icloud-filter-region"),
				City:    viper.GetStringSlice("icloud-filter-city"),
			},
			Verbosity: verbosity,
		}

		icloud.GetIPRanges(config)
	},
}

func init() {
	icloudCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	icloudCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	icloudCmd.Flags().StringSlice("filter-country", []string{}, "Filter results by country")
	icloudCmd.Flags().StringSlice("filter-region", []string{}, "Filter results by region")
	icloudCmd.Flags().StringSlice("filter-city", []string{}, `Filter results by city (use quotes for names with spaces, e.g. "New York")`)

	viper.BindPFlag("icloud_ipv4", icloudCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("icloud_ipv6", icloudCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("icloud-filter-country", icloudCmd.Flags().Lookup("filter-country"))
	viper.BindPFlag("icloud-filter-region", icloudCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("icloud-filter-city", icloudCmd.Flags().Lookup("filter-city"))

	rootCmd.AddCommand(icloudCmd)
}
