package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kaumnen/cipr/internal/atlassian"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var atlassianCmd = &cobra.Command{
	Use:   "atlassian",
	Short: "Get Atlassian IP ranges.",
	Long:  `Get Atlassian IPv4 and IPv6 ranges with optional product, direction, region and perimeter filtering.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity, err := resolveVerbosity(cmd)
		if err != nil {
			return err
		}

		ipType := resolveIPType(viper.GetBool("atlassian_ipv4"), viper.GetBool("atlassian_ipv6"))
		filters := atlassian.Filters{
			Product:   viper.GetStringSlice("atlassian-filter-product"),
			Direction: viper.GetStringSlice("atlassian-filter-direction"),
			Region:    viper.GetStringSlice("atlassian-filter-region"),
			Perimeter: viper.GetStringSlice("atlassian-filter-perimeter"),
		}

		list := viper.GetString("atlassian-list")
		if list != "" && !slices.Contains(atlassianListDimensions, list) {
			return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(atlassianListDimensions, ", "))
		}

		return atlassian.GetIPRanges(cmd.Context(), atlassian.Config{
			Source:    utils.ResolveSource("atlassian"),
			IPType:    ipType,
			Filters:   filters,
			List:      list,
			Verbosity: verbosity,
		})
	},
}

var atlassianListDimensions = []string{"products", "directions", "regions", "perimeters"}

func init() {
	rootCmd.AddCommand(atlassianCmd)

	atlassianCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	atlassianCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	atlassianCmd.Flags().StringSlice("filter-product", []string{}, "Filter results by Atlassian product (comma-separated; e.g. jira, confluence, bitbucket, trello, loom)")
	atlassianCmd.Flags().StringSlice("filter-direction", []string{}, "Filter results by traffic direction (comma-separated; egress, ingress)")
	atlassianCmd.Flags().StringSlice("filter-region", []string{}, "Filter results by region (comma-separated; e.g. global, us-east-1, eu-central-1)")
	atlassianCmd.Flags().StringSlice("filter-perimeter", []string{}, "Filter results by perimeter (comma-separated; e.g. commercial, fedramp-moderate)")
	atlassianCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: products, directions, regions, perimeters. Composes with --filter-* flags; ignores --ipv4/--ipv6.")

	viper.BindPFlag("atlassian_ipv4", atlassianCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("atlassian_ipv6", atlassianCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("atlassian-filter-product", atlassianCmd.Flags().Lookup("filter-product"))
	viper.BindPFlag("atlassian-filter-direction", atlassianCmd.Flags().Lookup("filter-direction"))
	viper.BindPFlag("atlassian-filter-region", atlassianCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("atlassian-filter-perimeter", atlassianCmd.Flags().Lookup("filter-perimeter"))
	viper.BindPFlag("atlassian-list", atlassianCmd.Flags().Lookup("list"))
}
