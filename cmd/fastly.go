package cmd

import (
	"github.com/kaumnen/cipr/internal/fastly"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fastlyCmd = &cobra.Command{
	Use:   "fastly",
	Short: "Get Fastly IP ranges.",
	Long:  `Get Fastly IPv4 and IPv6 ranges.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity, err := resolveVerbosity(cmd)
		if err != nil {
			return err
		}

		return fastly.GetIPRanges(cmd.Context(), fastly.Config{
			Source:    utils.ResolveSource("fastly"),
			IPType:    resolveIPType(viper.GetBool("fastly_ipv4"), viper.GetBool("fastly_ipv6")),
			Verbosity: verbosity,
		})
	},
}

func init() {
	rootCmd.AddCommand(fastlyCmd)

	fastlyCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	fastlyCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")

	viper.BindPFlag("fastly_ipv4", fastlyCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("fastly_ipv6", fastlyCmd.Flags().Lookup("ipv6"))
}
