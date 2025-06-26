package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lyarwood/godar/pkg/config"
	"github.com/lyarwood/godar/pkg/logger"
	"github.com/lyarwood/godar/pkg/monitor"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor aircraft using the godar monitoring service",
	Long:  `Starts the godar monitoring service using the configuration file and environment variables`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}

		logger.Init(cfg.Monitoring.Debug)
		log := logger.GetLogger()

		mon, err := monitor.NewMonitor(cfg, log)
		if err != nil {
			log.Fatal("Failed to create monitor", zap.Error(err))
			os.Exit(1)
		}

		if err := mon.Start(); err != nil {
			log.Fatal("Failed to start monitor", zap.Error(err))
			os.Exit(1)
		}

		// Handle graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Info("Received shutdown signal, stopping monitor...")
		mon.Stop()
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)
}
