package cmd

import (
	"github.com/kaumnen/cipr/internal/googlebot"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var googlebotCmd = &cobra.Command{
	Use:   "googlebot",
	Short: "Get Googlebot IP ranges.",
	Long: `Get IPv4 and IPv6 ranges for Google's common crawlers, including Googlebot.

Google publishes other crawler lists in the same format. Pass one with --source:
  https://developers.google.com/static/crawling/ipranges/special-crawlers.json
  https://developers.google.com/static/crawling/ipranges/user-triggered-fetchers.json
  https://developers.google.com/static/crawling/ipranges/user-triggered-fetchers-google.json
  https://developers.google.com/static/crawling/ipranges/user-triggered-agents.json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity, err := resolveVerbosity(cmd)
		if err != nil {
			return err
		}

		return googlebot.GetIPRanges(cmd.Context(), googlebot.Config{
			Source:    utils.ResolveSource("googlebot"),
			IPType:    resolveIPType(viper.GetBool("googlebot_ipv4"), viper.GetBool("googlebot_ipv6")),
			Verbosity: verbosity,
		})
	},
}

func init() {
	rootCmd.AddCommand(googlebotCmd)

	googlebotCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	googlebotCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")

	viper.BindPFlag("googlebot_ipv4", googlebotCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("googlebot_ipv6", googlebotCmd.Flags().Lookup("ipv6"))
}
