package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/kaumnen/cipr/internal/aws"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Get AWS IP ranges.",
	Long:  `Get AWS IPv4 and IPv6 ranges with optional filtering.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity, err := resolveVerbosity(cmd)
		if err != nil {
			return err
		}

		ipv4 := viper.GetBool("aws_ipv4")
		ipv6 := viper.GetBool("aws_ipv6")
		ipType := resolveIPType(ipv4, ipv6)

		filter := viper.GetString("aws-filter")
		filterRegion := viper.GetString("aws-filter-region")
		filterService := viper.GetString("aws-filter-service")
		filterNetworkBorderGroup := viper.GetString("aws-filter-network-border-group")

		if filter != "" && (filterRegion != "" || filterService != "" || filterNetworkBorderGroup != "") {
			return errors.New("--filter cannot be used with individual filter flags")
		}

		var awsFilter string
		if filter != "" {
			awsFilter = filter
		} else {
			awsFilter = fmt.Sprintf("%s,%s,%s", filterRegion, filterService, filterNetworkBorderGroup)
		}

		source := utils.ResolveSource("aws")

		if list := viper.GetString("aws-list"); list != "" {
			if !slices.Contains(awsListDimensions, list) {
				return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(awsListDimensions, ", "))
			}
			return aws.GetIPRanges(cmd.Context(), aws.Config{
				Source:    source,
				IPType:    "",
				Filter:    awsFilter,
				List:      list,
				Verbosity: verbosity,
			})
		}

		return aws.GetIPRanges(cmd.Context(), aws.Config{
			Source: source, IPType: ipType, Filter: awsFilter, Verbosity: verbosity,
		})
	},
}

var awsListDimensions = []string{"regions", "services", "network-border-groups"}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	awsCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	awsCmd.Flags().String("filter", "", "Filter results. Syntax: region,service,network-border-group")

	awsCmd.Flags().String("filter-region", "", "Filter results by AWS region")
	awsCmd.Flags().String("filter-service", "", "Filter results by AWS service")
	awsCmd.Flags().String("filter-network-border-group", "", "Filter results by AWS network border group")
	awsCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: regions, services, network-border-groups. Composes with --filter-* flags; ignores --ipv4/--ipv6.")

	viper.BindPFlag("aws_ipv4", awsCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("aws_ipv6", awsCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("aws-filter", awsCmd.Flags().Lookup("filter"))
	viper.BindPFlag("aws-filter-region", awsCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("aws-filter-service", awsCmd.Flags().Lookup("filter-service"))
	viper.BindPFlag("aws-filter-network-border-group", awsCmd.Flags().Lookup("filter-network-border-group"))
	viper.BindPFlag("aws-list", awsCmd.Flags().Lookup("list"))
}
