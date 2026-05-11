package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kaumnen/cipr/internal/github"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Get GitHub IP ranges.",
	Long:  `Get GitHub IPv4 and IPv6 ranges with optional service filtering.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbosity := resolveVerbosity(cmd)

		ipv4 := viper.GetBool("github_ipv4")
		ipv6 := viper.GetBool("github_ipv6")
		var ipType string
		switch {
		case ipv4 && !ipv6:
			ipType = "ipv4"
		case ipv6 && !ipv4:
			ipType = "ipv6"
		default:
			ipType = ""
		}

		filterService := viper.GetString("github-filter-service")
		source := utils.ResolveSource("github")

		if list := viper.GetString("github-list"); list != "" {
			if !slices.Contains(githubListDimensions, list) {
				return fmt.Errorf("invalid --list value %q (valid: %s)", list, strings.Join(githubListDimensions, ", "))
			}
			return github.GetIPRanges(cmd.Context(), github.Config{
				Source:        source,
				IPType:        "",
				FilterService: filterService,
				List:          list,
				Verbosity:     verbosity,
			})
		}

		return github.GetIPRanges(cmd.Context(), github.Config{
			Source:        source,
			IPType:        ipType,
			FilterService: filterService,
			Verbosity:     verbosity,
		})
	},
}

var githubListDimensions = []string{"services"}

func init() {
	rootCmd.AddCommand(githubCmd)

	githubCmd.Flags().Bool("ipv4", false, "Get only IPv4 ranges")
	githubCmd.Flags().Bool("ipv6", false, "Get only IPv6 ranges")
	githubCmd.Flags().String("filter-service", "", "Filter results by GitHub service (e.g. actions, web, api, git, hooks, pages, packages, importer, github_enterprise_importer)")
	githubCmd.Flags().String("list", "", "List unique values for a dimension instead of IP ranges. Valid: services. Composes with --filter-service; ignores --ipv4/--ipv6.")

	viper.BindPFlag("github_ipv4", githubCmd.Flags().Lookup("ipv4"))
	viper.BindPFlag("github_ipv6", githubCmd.Flags().Lookup("ipv6"))
	viper.BindPFlag("github-filter-service", githubCmd.Flags().Lookup("filter-service"))
	viper.BindPFlag("github-list", githubCmd.Flags().Lookup("list"))
}
