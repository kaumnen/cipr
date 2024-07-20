package cmd

import (
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"

	"github.com/kaumnen/cipr/internal/aws"
)

var (
	awsIPv4Flag     bool
	awsIPv6Flag     bool
	awsIPFilterFlag string
	awsVerbosity    string
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Get AWS IP ranges.",
	Long:  `Get AWS IPv4 and IPv6 ranges.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetCiprLogger()

		logger.Println("AWS subcommand called")

		if awsIPv4Flag || (!awsIPv4Flag && !awsIPv6Flag) {
			aws.GetIPRanges("ipv4", awsIPFilterFlag, awsVerbosity, utils.GetReq)
		}
		if awsIPv6Flag || (!awsIPv4Flag && !awsIPv6Flag) {
			aws.GetIPRanges("ipv6", awsIPFilterFlag, awsVerbosity, utils.GetReq)
		}
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().BoolVar(&awsIPv4Flag, "ipv4", false, "Get only IPv4 ranges")
	awsCmd.Flags().BoolVar(&awsIPv6Flag, "ipv6", false, "Get only IPv6 ranges")
	awsCmd.Flags().StringVar(&awsIPFilterFlag, "filter", "", "Filter results. Syntax: aws-region-az,SERVICE,network-border-group")
	awsCmd.Flags().StringVar(&awsVerbosity, "verbose", "none", "Verbosity. Options: none, mini, full")
}
