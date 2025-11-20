package applet

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/lyarwood/godar/pkg/config"
	"github.com/lyarwood/godar/pkg/fetch"
	"github.com/lyarwood/godar/pkg/monitor"
	"go.uber.org/zap"
)

// AircraftDetection represents a detected aircraft for display in the menu
type AircraftDetection struct {
	Callsign string
	Type     string
	Altitude int
	Distance float64
	Time     time.Time
}

// Applet represents the system tray applet
type Applet struct {
	config          *config.Config
	logger          *zap.Logger
	monitor         *monitor.Monitor
	monitorRunning  bool
	mu              sync.Mutex
	recentAircraft  []AircraftDetection
	maxRecent       int
	ctx             context.Context
	cancel          context.CancelFunc
	mToggleMonitor  *systray.MenuItem
	mRecentAircraft *systray.MenuItem
	mQuit           *systray.MenuItem
}

// NewApplet creates a new system tray applet
func NewApplet(cfg *config.Config, logger *zap.Logger) *Applet {
	ctx, cancel := context.WithCancel(context.Background())
	return &Applet{
		config:         cfg,
		logger:         logger,
		monitorRunning: false,
		recentAircraft: make([]AircraftDetection, 0),
		maxRecent:      5,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Run starts the system tray applet
func (a *Applet) Run() {
	systray.Run(a.onReady, a.onExit)
}

// onReady is called when the systray is ready
func (a *Applet) onReady() {
	// Set icon and tooltip
	icon := getIcon()
	if len(icon) > 0 {
		systray.SetIcon(icon)
	}
	systray.SetTitle("Godar")
	systray.SetTooltip("Aircraft Monitor")

	// Create menu items
	a.mToggleMonitor = systray.AddMenuItem("Start Monitoring", "Start aircraft monitoring")
	systray.AddSeparator()

	a.mRecentAircraft = systray.AddMenuItem("Recent Aircraft (0)", "Recently detected aircraft")
	systray.AddSeparator()

	mConfig := systray.AddMenuItem("Configuration", "Show current configuration")
	systray.AddSeparator()

	a.mQuit = systray.AddMenuItem("Quit", "Quit the application")

	// Start event loop
	go a.eventLoop()

	// Show configuration info
	go func() {
		for range mConfig.ClickedCh {
			a.showConfiguration()
		}
	}()

	// Show recent aircraft
	go func() {
		for range a.mRecentAircraft.ClickedCh {
			a.showRecentAircraft()
		}
	}()

	// Auto-start monitoring on startup
	go func() {
		// Give the UI a moment to fully initialize
		time.Sleep(500 * time.Millisecond)
		a.toggleMonitoring()
	}()
}

// onExit is called when the systray is exiting
func (a *Applet) onExit() {
	a.logger.Info("Exiting applet")
	a.cancel()
	if a.monitorRunning && a.monitor != nil {
		a.monitor.Stop()
	}
}

// eventLoop handles menu item clicks
func (a *Applet) eventLoop() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-a.mToggleMonitor.ClickedCh:
			a.toggleMonitoring()
		case <-a.mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

// toggleMonitoring starts or stops the monitoring service
func (a *Applet) toggleMonitoring() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.monitorRunning {
		// Stop monitoring
		if a.monitor != nil {
			a.monitor.Stop()
			a.monitor = nil
		}
		a.monitorRunning = false
		a.mToggleMonitor.SetTitle("Start Monitoring")
		systray.SetTooltip("Aircraft Monitor - Stopped")
		a.logger.Info("Monitoring stopped via applet")
	} else {
		// Start monitoring with custom notifier that updates the applet
		fetcher := fetch.NewFetcher(a.config.Server.URL, a.logger)
		fetcher.SetFilters(
			a.config.Filters.AircraftType,
			a.config.Filters.MinAltitude,
			a.config.Filters.MaxAltitude,
			a.config.Filters.Military,
			a.config.Filters.Operator,
			a.config.Filters.FlightNumber,
		)
		fetcher.SetLocation(
			a.config.Location.Latitude,
			a.config.Location.Longitude,
			a.config.Location.MaxDistance,
		)
		if a.config.Server.Username != "" || a.config.Server.Password != "" {
			fetcher.SetAuth(a.config.Server.Username, a.config.Server.Password)
		}

		notifier := NewAppletNotifier(a, a.config.Notification.Enabled, a.config.Notification.Duration, a.logger)

		mon, err := monitor.NewMonitorWithDeps(a.config, a.logger, fetcher, notifier)
		if err != nil {
			a.logger.Error("Failed to create monitor", zap.Error(err))
			return
		}

		if err := mon.Start(); err != nil {
			a.logger.Error("Failed to start monitor", zap.Error(err))
			return
		}

		a.monitor = mon
		a.monitorRunning = true
		a.mToggleMonitor.SetTitle("Stop Monitoring")
		systray.SetTooltip("Aircraft Monitor - Running")
		a.logger.Info("Monitoring started via applet")
	}
}

// AddAircraftDetection adds a detected aircraft to the recent list
func (a *Applet) AddAircraftDetection(callsign, aircraftType string, altitude int, distance float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	detection := AircraftDetection{
		Callsign: callsign,
		Type:     aircraftType,
		Altitude: altitude,
		Distance: distance,
		Time:     time.Now(),
	}

	// Add to front of list
	a.recentAircraft = append([]AircraftDetection{detection}, a.recentAircraft...)

	// Keep only the most recent items
	if len(a.recentAircraft) > a.maxRecent {
		a.recentAircraft = a.recentAircraft[:a.maxRecent]
	}

	// Update menu
	a.updateRecentAircraftMenu()
}

// updateRecentAircraftMenu updates the recent aircraft submenu
func (a *Applet) updateRecentAircraftMenu() {
	// Update the menu title with count
	title := fmt.Sprintf("Recent Aircraft (%d)", len(a.recentAircraft))
	a.mRecentAircraft.SetTitle(title)
}

// showConfiguration displays the current configuration in a notification
func (a *Applet) showConfiguration() {
	message := fmt.Sprintf("Server: %s\nPoll Interval: %s\nMax Distance: %.0f km\nLatitude: %.2f\nLongitude: %.2f",
		a.config.Server.URL,
		a.config.Monitoring.PollInterval,
		a.config.Location.MaxDistance,
		a.config.Location.Latitude,
		a.config.Location.Longitude)

	err := beeep.Notify("Godar Configuration", message, "")
	if err != nil {
		a.logger.Error("Failed to show configuration notification", zap.Error(err))
	}

	a.logger.Info("Configuration displayed",
		zap.String("server", a.config.Server.URL),
		zap.Duration("poll_interval", a.config.Monitoring.PollInterval),
		zap.Float64("max_distance", a.config.Location.MaxDistance))
}

// showRecentAircraft displays recent aircraft detections in a notification
func (a *Applet) showRecentAircraft() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.recentAircraft) == 0 {
		beeep.Notify("Recent Aircraft", "No recent aircraft detected", "")
		return
	}

	message := ""
	for i, ac := range a.recentAircraft {
		if i > 0 {
			message += "\n"
		}
		message += fmt.Sprintf("%s (%s)\n  Alt: %d ft, Dist: %.1f km\n  %s",
			ac.Callsign, ac.Type, ac.Altitude, ac.Distance,
			ac.Time.Format("15:04:05"))
	}

	err := beeep.Notify("Recent Aircraft", message, "")
	if err != nil {
		a.logger.Error("Failed to show recent aircraft notification", zap.Error(err))
	}

	a.logger.Info("Recent aircraft displayed", zap.Int("count", len(a.recentAircraft)))
}

// getIcon returns the icon data for the system tray
func getIcon() []byte {
	// Try to load icon from file first
	home, err := os.UserHomeDir()
	if err == nil {
		iconPath := filepath.Join(home, ".local", "share", "godar", "icon.png")
		if iconData, err := os.ReadFile(iconPath); err == nil {
			return iconData
		}
	}

	// Fallback to embedded icon if file not found
	iconData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x16,
		0x08, 0x06, 0x00, 0x00, 0x00, 0xC4, 0xB4, 0x6C, 0x3B, 0x00, 0x00, 0x00,
		0xEB, 0x49, 0x44, 0x41, 0x54, 0x48, 0x89, 0xED, 0x95, 0x41, 0x0A, 0x80,
		0x20, 0x10, 0x44, 0x7F, 0x4A, 0x0F, 0xF1, 0x24, 0xDE, 0xA5, 0xE8, 0xD6,
		0x8B, 0xF4, 0xFF, 0x4F, 0x5D, 0x3A, 0x94, 0x1E, 0xA2, 0x07, 0x45, 0x41,
		0x11, 0xBC, 0x09, 0x43, 0x08, 0x83, 0x99, 0x61, 0x42, 0x48, 0x29, 0x6D,
		0x7E, 0x8C, 0x31, 0x46, 0x49, 0x92, 0x24, 0x49, 0x92, 0x64, 0x01, 0x00,
		0xC0, 0x5E, 0xEB, 0xFF, 0xFF, 0xFF, 0x3F, 0x00, 0x00, 0xF0, 0x3C, 0x0F,
		0x00, 0x00, 0x78, 0x9E, 0x07, 0x00, 0x00, 0x3C, 0xCF, 0x03, 0x00, 0x00,
		0x9E, 0xE7, 0x01, 0x00, 0x80, 0xE7, 0x79, 0x00, 0x00, 0xE0, 0x79, 0x1E,
		0x00, 0x00, 0x78, 0x9E, 0x07, 0x00, 0x00, 0x9E, 0xE7, 0x01, 0x00, 0x80,
		0xE7, 0x79, 0x00, 0x00, 0xE0, 0x79, 0x1E, 0x00, 0x00, 0x78, 0x9E, 0x07,
		0x00, 0x00, 0x9E, 0xE7, 0x01, 0x00, 0x80, 0xE7, 0x79, 0x00, 0x00, 0xE0,
		0x79, 0x1E, 0x00, 0x00, 0x78, 0x9E, 0x07, 0x00, 0x00, 0x9E, 0xE7, 0x01,
		0x00, 0x80, 0xE7, 0x79, 0x00, 0x00, 0xE0, 0x79, 0x1E, 0x00, 0x00, 0x78,
		0x9E, 0x07, 0x00, 0x00, 0x9E, 0xE7, 0x01, 0x00, 0x80, 0xE7, 0x79, 0x00,
		0x00, 0xE0, 0x79, 0x1E, 0x00, 0x00, 0x78, 0x9E, 0x07, 0x00, 0x00, 0x9E,
		0xE7, 0x01, 0x00, 0x80, 0xE7, 0x79, 0x00, 0x00, 0xE0, 0x79, 0x1E, 0x00,
		0x00, 0x78, 0x9E, 0x07, 0x00, 0x00, 0x9E, 0xE7, 0x01, 0x00, 0x80, 0xE7,
		0x79, 0x00, 0x00, 0xE0, 0x79, 0x1E, 0x00, 0x00, 0x78, 0x9E, 0x07, 0x00,
		0x00, 0xFC, 0x7F, 0x00, 0x03, 0xD0, 0x5B, 0x78, 0x1E, 0xE9, 0x05, 0xD4,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	return iconData
}
