package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = ""
)

var rootCmd = &cobra.Command{
	Use:          "cipr",
	Version:      version,
	Short:        "Retrieve IP ranges from cloud providers and services",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
	Long: `cipr is a CLI tool for retrieving IP ranges from various cloud providers
and services (AWS, Azure, Cloudflare, DigitalOcean, GitHub, iCloud Private Relay).

It provides a quick and efficient way to access up-to-date IP ranges, useful
for network administrators, security professionals, and developers working
with cloud infrastructure.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	if version != "" {
		utils.UserAgent = "cipr/" + version
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.config/cipr/cipr.toml)")

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output (equivalent to --verbose-mode=full)")
	rootCmd.PersistentFlags().String("verbose-mode", "none", "Verbosity level: none, mini, full. Overrides --verbose")
	rootCmd.PersistentFlags().String("source", "config", "Data source: config, an HTTP(S) URL, or a local file path")
	rootCmd.PersistentFlags().Bool("no-cache", false, "Bypass cache (skip read and write)")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("verbose_mode", rootCmd.PersistentFlags().Lookup("verbose-mode"))
	viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source"))
	viper.BindPFlag("no_cache", rootCmd.PersistentFlags().Lookup("no-cache"))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

func initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configPath := filepath.Join(home, ".config", "cipr", "cipr.toml")
		viper.SetConfigFile(configPath)
		viper.SetConfigType("toml")

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := createDefaultConfig(configPath); err != nil {
				return err
			}
		} else if err != nil {
			return fmt.Errorf("inspect config file: %w", err)
		}
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	return nil
}

func resolveVerbosity(cmd *cobra.Command) (string, error) {
	var verbosity string
	switch {
	case cmd.Flags().Changed("verbose-mode"):
		verbosity, _ = cmd.Flags().GetString("verbose-mode")
	case cmd.Flags().Changed("verbose"):
		enabled, _ := cmd.Flags().GetBool("verbose")
		if enabled {
			verbosity = "full"
		} else {
			verbosity = "none"
		}
	default:
		verbosity = viper.GetString("verbose_mode")
		if verbosity == "" || verbosity == "none" && viper.GetBool("verbose") {
			if viper.GetBool("verbose") {
				verbosity = "full"
			} else {
				verbosity = "none"
			}
		}
	}

	if !isValidVerbosity(verbosity) {
		return "", fmt.Errorf("invalid verbosity level %q (allowed: none, mini, full)", verbosity)
	}
	return verbosity, nil
}

func isValidVerbosity(v string) bool {
	switch v {
	case "none", "mini", "full":
		return true
	}
	return false
}

func createDefaultConfig(configPath string) error {
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
	}

	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	const header = `# cipr config. Override per-provider data sources here, or pass --source on
# the command line. For each provider:
#   <provider>_endpoint   = URL fetched when --source=config (the default)
#   <provider>_local_file = if set, read ranges from this path instead of the network
#   <provider>_cache_ttl  = how long to reuse a cached hosted response. Go duration
#                           string ("24h", "30m"). "0s" disables caching for this
#                           provider; defaults to 24h if unset or unparseable.

`
	if _, err := fmt.Fprint(file, header); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	keys := make([]string, 0, len(utils.DefaultEndpoints))
	for k := range utils.DefaultEndpoints {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		_, err := fmt.Fprintf(file, "%s_endpoint = %q\n%s_local_file = \"\"\n%s_cache_ttl = \"24h\"\n\n", k, utils.DefaultEndpoints[k], k, k)
		if err != nil {
			return fmt.Errorf("write config file: %w", err)
		}
	}
	return file.Close()
}

func resolveIPType(ipv4, ipv6 bool) string {
	switch {
	case ipv4 && !ipv6:
		return "ipv4"
	case ipv6 && !ipv4:
		return "ipv6"
	default:
		return "both"
	}
}
