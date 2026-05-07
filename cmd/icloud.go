package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kaumnen/cipr/internal/icloud"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var icloudCmd = &cobra.Command{
	Use:   "icloud",
	Short: "Get iCloud private relay IP ranges.",
	Long:  `Get iCloud private relay IPv4 and IPv6 ranges.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity := resolveVerbosity(cmd)

		ipType := "ipv4"
		if viper.GetBool("icloud_ipv6") {
			ipType = "ipv6"
		}

		if !viper.GetBool("icloud_ipv4") && !viper.GetBool("icloud_ipv6") {
			ipType = "both"
		}

		filters := icloud.Filters{
			Country: viper.GetStringSlice("icloud-filter-country"),
			Region:  viper.GetStringSlice("icloud-filter-region"),
			City:    viper.GetStringSlice("icloud-filter-city"),
		}

		if list := viper.GetString("icloud-list"); list != "" {
			if !slices.Contains(icloudListDimensions, list) {
				return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(icloudListDimensions, ", "))
			}
			return icloud.GetIPRanges(cmd.Context(), icloud.Config{
				Source:    utils.ResolveSource("icloud"),
				IPType:    "both",
				Filters:   filters,
				List:      list,
				Verbosity: verbosity,
			})
		}

		return icloud.GetIPRanges(cmd.Context(), icloud.Config{
			Source:    utils.ResolveSource("icloud"),
			IPType:    ipType,
			Filters:   filters,
			Verbosity: verbosity,
		})
	},
}

var icloudListDimensions = []string{"countries", "regions", "cities"}

func init() {
	icloudCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	icloudCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	icloudCmd.Flags().StringSlice("filter-country", []string{}, "Filter results by country")
	icloudCmd.Flags().StringSlice("filter-region", []string{}, "Filter results by region")
	icloudCmd.Flags().StringSlice("filter-city", []string{}, `Filter results by city (use quotes for names with spaces, e.g. "New York")`)
	icloudCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: countries, regions, cities. Composes with --filter-* flags; ignores --ipv4/--ipv6.")

	viper.BindPFlag("icloud_ipv4", icloudCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("icloud_ipv6", icloudCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("icloud-filter-country", icloudCmd.Flags().Lookup("filter-country"))
	viper.BindPFlag("icloud-filter-region", icloudCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("icloud-filter-city", icloudCmd.Flags().Lookup("filter-city"))
	viper.BindPFlag("icloud-list", icloudCmd.Flags().Lookup("list"))

	rootCmd.AddCommand(icloudCmd)
}
