package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kaumnen/cipr/internal/digitalocean"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Get Digital Ocean IP ranges",
	Long:  `Retrieve Digital Ocean IPv4 and IPv6 ranges with optional verbosity levels.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity := resolveVerbosity(cmd)

		ipv4 := viper.GetBool("do_ipv4")
		ipv6 := viper.GetBool("do_ipv6")
		both := !ipv4 && !ipv6
		var ipVersions []string
		if ipv4 || both {
			ipVersions = append(ipVersions, "ipv4")
		}
		if ipv6 || both {
			ipVersions = append(ipVersions, "ipv6")
		}

		source := utils.ResolveSource("digitalocean")
		filters := digitalocean.Filters{
			Country: viper.GetStringSlice("do-filter-country"),
			Region:  viper.GetStringSlice("do-filter-region"),
			City:    viper.GetStringSlice("do-filter-city"),
			Zip:     viper.GetStringSlice("do-filter-zip"),
		}

		if list := viper.GetString("do-list"); list != "" {
			if !slices.Contains(doListDimensions, list) {
				return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(doListDimensions, ", "))
			}
			return digitalocean.GetIPRanges(cmd.Context(), digitalocean.Config{
				Source:    source,
				IPType:    "both",
				Filters:   filters,
				List:      list,
				Verbosity: verbosity,
			})
		}

		for _, version := range ipVersions {
			if err := digitalocean.GetIPRanges(cmd.Context(), digitalocean.Config{
				Source:    source,
				IPType:    version,
				Filters:   filters,
				Verbosity: verbosity,
			}); err != nil {
				return err
			}
		}
		return nil
	},
}

var doListDimensions = []string{"countries", "regions", "cities", "zips"}

func init() {

	doCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	doCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	doCmd.Flags().StringSlice("filter-country", []string{}, "Filter results by country")
	doCmd.Flags().StringSlice("filter-region", []string{}, "Filter results by region")
	doCmd.Flags().StringSlice("filter-city", []string{}, `Filter results by city (use quotes for names with spaces, e.g. "New York")`)
	doCmd.Flags().StringSlice("filter-zip", []string{}, `Filter results by ZIP code`)
	doCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: countries, regions, cities, zips. Composes with --filter-* flags; ignores --ipv4/--ipv6.")

	viper.BindPFlag("do_ipv4", doCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("do_ipv6", doCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("do-filter-country", doCmd.Flags().Lookup("filter-country"))
	viper.BindPFlag("do-filter-region", doCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("do-filter-city", doCmd.Flags().Lookup("filter-city"))
	viper.BindPFlag("do-filter-zip", doCmd.Flags().Lookup("filter-zip"))
	viper.BindPFlag("do-list", doCmd.Flags().Lookup("list"))

	rootCmd.AddCommand(doCmd)
}
