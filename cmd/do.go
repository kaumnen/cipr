package cmd

import (
	"fmt"
	"os"

	"github.com/kaumnen/cipr/internal/digitalocean"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Get Digital Ocean IP ranges",
	Long:  `Retrieve Digital Ocean IPv4 and IPv6 ranges with optional verbosity levels.`,
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
		if viper.GetBool("do_ipv4") || (!viper.GetBool("do_ipv4") && !viper.GetBool("do_ipv6")) {
			ipVersions = append(ipVersions, "ipv4")
		}
		if viper.GetBool("do_ipv6") || (!viper.GetBool("do_ipv4") && !viper.GetBool("do_ipv6")) {
			ipVersions = append(ipVersions, "ipv6")
		}

		for _, version := range ipVersions {
			config := digitalocean.Config{
				IPType: version,
				Filters: digitalocean.Filters{
					Country: viper.GetStringSlice("do-filter-country"),
					Region:  viper.GetStringSlice("do-filter-region"),
					City:    viper.GetStringSlice("do-filter-city"),
					Zip:     viper.GetStringSlice("do-filter-zip"),
				},
				Verbosity: verbosity,
			}
			digitalocean.GetIPRanges(config)
		}
	},
}

func init() {

	doCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	doCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	doCmd.Flags().StringSlice("filter-country", []string{}, "Filter results by country")
	doCmd.Flags().StringSlice("filter-region", []string{}, "Filter results by region")
	doCmd.Flags().StringSlice("filter-city", []string{}, `Filter results by city (use quotes for names with spaces, e.g. "New York")`)
	doCmd.Flags().StringSlice("filter-zip", []string{}, `Filter results by ZIP code`)

	viper.BindPFlag("do_ipv4", doCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("do_ipv6", doCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("do-filter-country", doCmd.Flags().Lookup("filter-country"))
	viper.BindPFlag("do-filter-region", doCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("do-filter-city", doCmd.Flags().Lookup("filter-city"))
	viper.BindPFlag("do-filter-zip", doCmd.Flags().Lookup("filter-zip"))

	rootCmd.AddCommand(doCmd)
}
