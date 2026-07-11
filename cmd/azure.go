package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/kaumnen/cipr/internal/azure"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Get Azure IP ranges.",
	Long:  `Get Azure IPv4 and IPv6 ranges from the Public cloud service tags, with optional filtering.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity, err := resolveVerbosity(cmd)
		if err != nil {
			return err
		}

		ipv4 := viper.GetBool("azure_ipv4")
		ipv6 := viper.GetBool("azure_ipv6")
		ipType := resolveIPType(ipv4, ipv6)

		filter := viper.GetString("azure-filter")
		filters := azure.Filters{
			Region:  viper.GetStringSlice("azure-filter-region"),
			Service: viper.GetStringSlice("azure-filter-service"),
		}

		if filter != "" && (len(filters.Region) > 0 || len(filters.Service) > 0) {
			return errors.New("--filter cannot be used with individual filter flags")
		}

		source := utils.ResolveSource("azure")

		if list := viper.GetString("azure-list"); list != "" {
			if !slices.Contains(azureListDimensions, list) {
				return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(azureListDimensions, ", "))
			}
			return azure.GetIPRanges(cmd.Context(), azure.Config{
				Source:    source,
				IPType:    "",
				Filter:    filter,
				Filters:   filters,
				List:      list,
				Verbosity: verbosity,
			})
		}

		return azure.GetIPRanges(cmd.Context(), azure.Config{
			Source: source, IPType: ipType, Filter: filter, Filters: filters, Verbosity: verbosity,
		})
	},
}

var azureListDimensions = []string{"regions", "services"}

func init() {
	rootCmd.AddCommand(azureCmd)

	azureCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	azureCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	azureCmd.Flags().String("filter", "", "Filter results. Syntax: region,service")

	azureCmd.Flags().StringSlice("filter-region", []string{}, "Filter results by Azure region (comma-separated, e.g. westeurope,eastus)")
	azureCmd.Flags().StringSlice("filter-service", []string{}, "Filter results by Azure system service (comma-separated, e.g. AzureStorage,AzureKeyVault)")
	azureCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: regions, services. Composes with --filter-* flags; ignores --ipv4/--ipv6.")

	viper.BindPFlag("azure_ipv4", azureCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("azure_ipv6", azureCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("azure-filter", azureCmd.Flags().Lookup("filter"))
	viper.BindPFlag("azure-filter-region", azureCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("azure-filter-service", azureCmd.Flags().Lookup("filter-service"))
	viper.BindPFlag("azure-list", azureCmd.Flags().Lookup("list"))
}
