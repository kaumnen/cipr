package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/kaumnen/cipr/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = ""
)

var rootCmd = &cobra.Command{
	Use:     "cipr",
	Version: version,
	Short:   "Retrieve IP ranges from cloud providers and services",
	Long: `cipr is a CLI tool for retrieving IP ranges from various cloud providers
and services (AWS, Cloudflare, DigitalOcean, iCloud Private Relay).

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
	cobra.OnInitialize(initConfig)

	if version != "" {
		utils.UserAgent = "cipr/" + version
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.config/cipr/cipr.toml)")

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output (equivalent to --verbose-mode=full)")
	rootCmd.PersistentFlags().String("verbose-mode", "none", "Verbosity level: none, mini, full. Overrides --verbose")
	rootCmd.PersistentFlags().String("source", "hosted", "Custom data source for ip ranges (url or path). Must have https:// for urls.")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("verbose_mode", rootCmd.PersistentFlags().Lookup("verbose-mode"))
	viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configPath := filepath.Join(home, ".config", "cipr", "cipr.toml")
		viper.SetConfigFile(configPath)
		viper.SetConfigType("toml")

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			createDefaultConfig(configPath)
		}
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading config file:", err)
	}
}

func resolveVerbosity(cmd *cobra.Command) string {
	var verbosity string
	switch {
	case cmd.Flags().Changed("verbose-mode"):
		verbosity = viper.GetString("verbose_mode")
	case viper.GetBool("verbose"):
		verbosity = "full"
	default:
		verbosity = "none"
	}

	if !isValidVerbosity(verbosity) {
		fmt.Fprintf(os.Stderr, "Invalid verbosity level: %s. Allowed values are: none, mini, full.\n", verbosity)
		os.Exit(1)
	}
	return verbosity
}

func isValidVerbosity(v string) bool {
	switch v {
	case "none", "mini", "full":
		return true
	}
	return false
}

func createDefaultConfig(configPath string) {
	config := map[string]interface{}{}
	for key, url := range utils.DefaultEndpoints {
		config[key+"_endpoint"] = url
		config[key+"_local_file"] = ""
	}

	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "Error creating config directory:", err)
			os.Exit(1)
		}
	}

	file, err := os.Create(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating config file:", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing config file:", err)
		os.Exit(1)
	}
}
