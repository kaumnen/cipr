package cmd

import (
	"fmt"
	"os"

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

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cipr")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
