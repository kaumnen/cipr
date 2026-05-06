package cmd

import (
	"errors"
	"fmt"

	"github.com/kaumnen/cipr/internal/aws"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Get AWS IP ranges.",
	Long:  `Get AWS IPv4 and IPv6 ranges with optional filtering.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity := resolveVerbosity(cmd)

		ipv4 := viper.GetBool("aws_ipv4")
		ipv6 := viper.GetBool("aws_ipv6")
		both := !ipv4 && !ipv6
		var ipVersion []string
		if ipv4 || both {
			ipVersion = append(ipVersion, "ipv4")
		}
		if ipv6 || both {
			ipVersion = append(ipVersion, "ipv6")
		}

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
		for _, version := range ipVersion {
			if err := aws.GetIPRanges(cmd.Context(), aws.Config{
				Source:    source,
				IPType:    version,
				Filter:    awsFilter,
				Verbosity: verbosity,
			}); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	awsCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	awsCmd.Flags().String("filter", "", "Filter results. Syntax: region,service,network-border-group")

	awsCmd.Flags().String("filter-region", "", "Filter results by AWS region")
	awsCmd.Flags().String("filter-service", "", "Filter results by AWS service")
	awsCmd.Flags().String("filter-network-border-group", "", "Filter results by AWS network border group")

	viper.BindPFlag("aws_ipv4", awsCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("aws_ipv6", awsCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("aws-filter", awsCmd.Flags().Lookup("filter"))
	viper.BindPFlag("aws-filter-region", awsCmd.Flags().Lookup("filter-region"))
	viper.BindPFlag("aws-filter-service", awsCmd.Flags().Lookup("filter-service"))
	viper.BindPFlag("aws-filter-network-border-group", awsCmd.Flags().Lookup("filter-network-border-group"))
}
