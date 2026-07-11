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
		filters := aws.Filters{
			Region:             viper.GetStringSlice("aws-filter-region"),
			Service:            viper.GetStringSlice("aws-filter-service"),
			NetworkBorderGroup: viper.GetStringSlice("aws-filter-network-border-group"),
		}

		if filter != "" && (len(filters.Region) > 0 || len(filters.Service) > 0 || len(filters.NetworkBorderGroup) > 0) {
			return errors.New("--filter cannot be used with individual filter flags")
		}

		source := utils.ResolveSource("aws")

		if list := viper.GetString("aws-list"); list != "" {
			if !slices.Contains(awsListDimensions, list) {
				return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(awsListDimensions, ", "))
			}
			return aws.GetIPRanges(cmd.Context(), aws.Config{
				Source:    source,
				IPType:    "",
				Filter:    filter,
				Filters:   filters,
				List:      list,
				Verbosity: verbosity,
			})
		}

		return aws.GetIPRanges(cmd.Context(), aws.Config{
			Source: source, IPType: ipType, Filter: filter, Filters: filters, Verbosity: verbosity,
		})
	},
}

var awsListDimensions = []string{"regions", "services", "network-border-groups"}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	awsCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	awsCmd.Flags().String("filter", "", "Filter results. Syntax: region,service,network-border-group")

	awsCmd.Flags().StringSlice("filter-region", []string{}, "Filter results by AWS region (comma-separated)")
	awsCmd.Flags().StringSlice("filter-service", []string{}, "Filter results by AWS service (comma-separated)")
	awsCmd.Flags().StringSlice("filter-network-border-group", []string{}, "Filter results by AWS network border group (comma-separated)")
	awsCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: regions, services, network-border-groups. Composes with --filter-* flags; ignores --ipv4/--ipv6.")

	viper.BindPFlag("aws_ipv4", awsCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("aws_ipv6", awsCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("aws-filter", awsCmd.Flags().Lookup("filter"))
	viper.BindPFlag("aws-filter-region", awsCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("aws-filter-service", awsCmd.Flags().Lookup("filter-service"))
	viper.BindPFlag("aws-filter-network-border-group", awsCmd.Flags().Lookup("filter-network-border-group"))
	viper.BindPFlag("aws-list", awsCmd.Flags().Lookup("list"))
}
