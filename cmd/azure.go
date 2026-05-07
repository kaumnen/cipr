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
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity := resolveVerbosity(cmd)

		ipv4 := viper.GetBool("azure_ipv4")
		ipv6 := viper.GetBool("azure_ipv6")
		both := !ipv4 && !ipv6
		var ipVersions []string
		if ipv4 || both {
			ipVersions = append(ipVersions, "ipv4")
		}
		if ipv6 || both {
			ipVersions = append(ipVersions, "ipv6")
		}

		filter := viper.GetString("azure-filter")
		filterRegion := viper.GetString("azure-filter-region")
		filterService := viper.GetString("azure-filter-service")

		if filter != "" && (filterRegion != "" || filterService != "") {
			return errors.New("--filter cannot be used with individual filter flags")
		}

		var azureFilter string
		if filter != "" {
			azureFilter = filter
		} else {
			azureFilter = fmt.Sprintf("%s,%s", filterRegion, filterService)
		}

		source := utils.ResolveSource("azure")

		if list := viper.GetString("azure-list"); list != "" {
			if !slices.Contains(azureListDimensions, list) {
				return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(azureListDimensions, ", "))
			}
			return azure.GetIPRanges(cmd.Context(), azure.Config{
				Source:    source,
				IPType:    "",
				Filter:    azureFilter,
				List:      list,
				Verbosity: verbosity,
			})
		}

		for _, version := range ipVersions {
			if err := azure.GetIPRanges(cmd.Context(), azure.Config{
				Source:    source,
				IPType:    version,
				Filter:    azureFilter,
				Verbosity: verbosity,
			}); err != nil {
				return err
			}
		}
		return nil
	},
}

var azureListDimensions = []string{"regions", "services"}

func init() {
	rootCmd.AddCommand(azureCmd)

	azureCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	azureCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	azureCmd.Flags().String("filter", "", "Filter results. Syntax: region,service")

	azureCmd.Flags().String("filter-region", "", "Filter results by Azure region (e.g. westeurope)")
	azureCmd.Flags().String("filter-service", "", "Filter results by Azure system service (e.g. AzureStorage)")
	azureCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: regions, services. Composes with --filter-* flags; ignores --ipv4/--ipv6.")

	viper.BindPFlag("azure_ipv4", azureCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("azure_ipv6", azureCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("azure-filter", azureCmd.Flags().Lookup("filter"))
	viper.BindPFlag("azure-filter-region", azureCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("azure-filter-service", azureCmd.Flags().Lookup("filter-service"))
	viper.BindPFlag("azure-list", azureCmd.Flags().Lookup("list"))
}
