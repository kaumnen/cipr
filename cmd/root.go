package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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

func createDefaultConfig(configPath string) {
	config := map[string]interface{}{
		"aws_endpoint":               "https://ip-ranges.amazonaws.com/ip-ranges.json",
		"aws_local_file":             "",
		"cloudflare_ipv4_endpoint":   "https://www.cloudflare.com/ips-v4/",
		"cloudflare_ipv4_local_file": "",
		"cloudflare_ipv6_endpoint":   "https://www.cloudflare.com/ips-v6/",
		"cloudflare_ipv6_local_file": "",
		"icloud_endpoint":            "https://mask-api.icloud.com/egress-ip-ranges.csv",
		"icloud_local_file":          "",
		"digitalocean_endpoint":      "https://digitalocean.com/geo/google.csv",
		"digitalocean_local_file":    "",
	}

	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
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
