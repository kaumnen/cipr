package cmd

import (
	"fmt"

	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"

	"github.com/kaumnen/cipr/internal/aws"
)

var (
	awsIPv4Flag      bool
	awsIPv6Flag      bool
	awsIPFilterFlag  string
	awsVerbosityFlag string

	awsFilterRegionFlag             string
	awsFilterServiceFlag            string
	awsFilterNetworkBorderGroupFlag string
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Get AWS IP ranges.",
	Long:  `Get AWS IPv4 and IPv6 ranges.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetCiprLogger()
		var awsIPFilter string

		if awsIPFilterFlag != "" && (awsFilterRegionFlag != "" || awsFilterServiceFlag != "" || awsFilterNetworkBorderGroupFlag != "") {
			logger.Fatalf("--filter flag cannot be used with individual filter flags")
		}

		if awsIPFilterFlag != "" {
			awsIPFilter = awsIPFilterFlag
		} else {
			awsIPFilter = fmt.Sprintf("%s,%s,%s", awsFilterRegionFlag, awsFilterServiceFlag, awsFilterNetworkBorderGroupFlag)
		}

		if awsIPv4Flag || (!awsIPv4Flag && !awsIPv6Flag) {
			aws.GetIPRanges("ipv4", awsIPFilter, awsVerbosityFlag)
		}
		if awsIPv6Flag || (!awsIPv4Flag && !awsIPv6Flag) {
			aws.GetIPRanges("ipv6", awsIPFilter, awsVerbosityFlag)
		}
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().BoolVar(&awsIPv4Flag, "ipv4", false, "Get only IPv4 ranges")
	awsCmd.Flags().BoolVar(&awsIPv6Flag, "ipv6", false, "Get only IPv6 ranges")
	awsCmd.Flags().StringVar(&awsIPFilterFlag, "filter", "", "Filter results. Syntax: aws-region-az,SERVICE,network-border-group")
	awsCmd.Flags().StringVar(&awsVerbosityFlag, "verbose", "none", "Verbosity. Options: none, mini, full")

	awsCmd.Flags().StringVar(&awsFilterRegionFlag, "filter-region", "", "Filter results by AWS region")
	awsCmd.Flags().StringVar(&awsFilterServiceFlag, "filter-service", "", "Filter results by AWS service")
	awsCmd.Flags().StringVar(&awsFilterNetworkBorderGroupFlag, "filter-network-border-group", "", "Filter results by AWS network border group")
}
