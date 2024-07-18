package cmd

import (
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"

	"github.com/kaumnen/cipr/internal/aws"
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Get AWS IP ranges.",
	Long:  `Get AWS IPv4 and IPv6 ranges.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := utils.GetCiprLogger()

		logger.Println("AWS subcommand called")

		if ipv4Flag || (!ipv4Flag && !ipv6Flag) {
			aws.GetIPRanges("ipv4")
		}
		if ipv6Flag || (!ipv4Flag && !ipv6Flag) {
			aws.GetIPRanges("ipv6")
		}
	},
}

func init() {
	rootCmd.AddCommand(awsCmd)

	awsCmd.Flags().BoolVar(&ipv4Flag, "ipv4", false, "Get only IPv4 ranges")
	awsCmd.Flags().BoolVar(&ipv6Flag, "ipv6", false, "Get only IPv6 ranges")
}
