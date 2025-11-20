package cmd

import (
	"fmt"
	"os"

	"github.com/lyarwood/godar/pkg/applet"
	"github.com/lyarwood/godar/pkg/config"
	"github.com/lyarwood/godar/pkg/logger"

	"github.com/spf13/cobra"
)

var appletCmd = &cobra.Command{
	Use:   "applet",
	Short: "Run godar as a system tray applet",
	Long:  `Starts godar as a system tray applet with a graphical interface for controlling the monitoring service`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}

		logger.Init(cfg.Monitoring.Debug)
		log := logger.GetLogger()

		app := applet.NewApplet(cfg, log)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(appletCmd)
}
