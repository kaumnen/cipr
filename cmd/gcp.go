package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/kaumnen/cipr/internal/gcp"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Get Google Cloud IP ranges.",
	Long:  `Get Google Cloud IPv4 and IPv6 ranges with optional scope and service filtering.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity, err := resolveVerbosity(cmd)
		if err != nil {
			return err
		}

		ipType := resolveIPType(viper.GetBool("gcp_ipv4"), viper.GetBool("gcp_ipv6"))
		filter := viper.GetString("gcp-filter")
		filters := gcp.Filters{
			Scope:   viper.GetStringSlice("gcp-filter-scope"),
			Service: viper.GetStringSlice("gcp-filter-service"),
		}

		if filter != "" && (len(filters.Scope) > 0 || len(filters.Service) > 0) {
			return errors.New("--filter cannot be used with individual filter flags")
		}

		source := utils.ResolveSource("gcp")
		if list := viper.GetString("gcp-list"); list != "" {
			if !slices.Contains(gcpListDimensions, list) {
				return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(gcpListDimensions, ", "))
			}
			return gcp.GetIPRanges(cmd.Context(), gcp.Config{
				Source:    source,
				IPType:    "both",
				Filter:    filter,
				Filters:   filters,
				List:      list,
				Verbosity: verbosity,
			})
		}

		return gcp.GetIPRanges(cmd.Context(), gcp.Config{
			Source: source, IPType: ipType, Filter: filter, Filters: filters, Verbosity: verbosity,
		})
	},
}

var gcpListDimensions = []string{"scopes", "services"}

func init() {
	rootCmd.AddCommand(gcpCmd)

	gcpCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	gcpCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	gcpCmd.Flags().String("filter", "", "Filter results. Syntax: scope,service")
	gcpCmd.Flags().StringSlice("filter-scope", []string{}, "Filter results by Google Cloud scope (comma-separated, for example, us-central1,global)")
	gcpCmd.Flags().StringSlice("filter-service", []string{}, "Filter results by Google Cloud service (comma-separated)")
	gcpCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: scopes, services. Composes with --filter-* flags; ignores --ipv4/--ipv6.")

	viper.BindPFlag("gcp_ipv4", gcpCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("gcp_ipv6", gcpCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("gcp-filter", gcpCmd.Flags().Lookup("filter"))
	viper.BindPFlag("gcp-filter-scope", gcpCmd.Flags().Lookup("filter-scope"))
	viper.BindPFlag("gcp-filter-service", gcpCmd.Flags().Lookup("filter-service"))
	viper.BindPFlag("gcp-list", gcpCmd.Flags().Lookup("list"))
}
