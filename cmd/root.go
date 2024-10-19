package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "cipr",
	Version: version,
	Short:   "",
	Long:    ``,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.cipr.yaml)")

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output (equivalent to --verbose-mode=full)")
	rootCmd.PersistentFlags().String("verbose-mode", "none", "Verbosity level: none, mini, full. Overrides --verbose")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("verbose_mode", rootCmd.PersistentFlags().Lookup("verbose-mode"))
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

func createDefaultConfig(configPath string) {
	viper.SetDefault("aws_endpoint", "https://ip-ranges.amazonaws.com/ip-ranges.json")
	viper.SetDefault("aws_local_file", "")

	viper.SetDefault("cloudflare_endpoint", "https://www.cloudflare.com/")
	viper.SetDefault("cloudflare_local_file", "")

	viper.SetDefault("icloud_endpoint", "https://mask-api.icloud.com/egress-ip-ranges.csv")
	viper.SetDefault("icloud_local_file", "")

	viper.SetDefault("digitalocean_endpoint", "https://digitalocean.com/geo/google.csv")
	viper.SetDefault("digitalocean_local_file", "")

	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Fprintln(os.Stderr, "Error creating config directory:", err)
			os.Exit(1)
		}
	}

	if err := viper.WriteConfigAs(configPath); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing config file:", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Created default config file at", configPath)
}
