package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "godar",
	Version: "dev", // This will be set during build
	Short:   "Godar is a CLI tool to monitor overflying aircraft",
	Long: `A command-line tool to monitor overflying aircraft from a Virtual Radar Server.

Connect to a Virtual Radar Server and alert the user to any overflying aircraft.
All filtering and options are set via configuration file or environment variables.

Examples:
  godar monitor
  # (set filters in godar.yaml or via GODAR_ environment variables)
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.godar.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %s\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		if runtime.GOOS == "windows" {
			if appData := os.Getenv("APPDATA"); appData != "" {
				viper.AddConfigPath(appData)
			}
		} else {
			viper.AddConfigPath("/etc/godar")
		}
		viper.SetConfigName("godar")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}
