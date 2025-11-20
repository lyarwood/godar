package applet

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/lyarwood/godar/pkg/config"
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

	a.mRecentAircraft = systray.AddMenuItem("Recent Aircraft", "Recently detected aircraft")
	a.mRecentAircraft.Disable()
	systray.AddSeparator()

	mConfig := systray.AddMenuItem("Configuration", "Show current configuration")
	systray.AddSeparator()

	a.mQuit = systray.AddMenuItem("Quit", "Quit the application")

	// Start event loop
	go a.eventLoop()

	// Show configuration info
	go func() {
		for range mConfig.ClickedCh {
			a.logger.Info("Configuration",
				zap.String("server", a.config.Server.URL),
				zap.Duration("poll_interval", a.config.Monitoring.PollInterval),
				zap.Float64("max_distance", a.config.Location.MaxDistance))
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
		// Start monitoring
		mon, err := monitor.NewMonitor(a.config, a.logger)
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
	// This is a simplified version - in a real implementation,
	// you would need to manage submenu items dynamically
	if len(a.recentAircraft) > 0 {
		title := fmt.Sprintf("Recent Aircraft (%d)", len(a.recentAircraft))
		a.mRecentAircraft.SetTitle(title)
		a.mRecentAircraft.Enable()
	} else {
		a.mRecentAircraft.SetTitle("Recent Aircraft")
		a.mRecentAircraft.Disable()
	}
}

// getIcon returns the icon data for the system tray
func getIcon() []byte {
	// Simple 16x16 airplane icon in PNG format
	// This is a basic airplane silhouette facing right
	iconData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff, 0x61, 0x00, 0x00, 0x00,
		0x9a, 0x49, 0x44, 0x41, 0x54, 0x38, 0x8d, 0x63, 0x60, 0x18, 0x05, 0xa4,
		0x80, 0x81, 0x81, 0x01, 0x06, 0x06, 0x06, 0x86, 0x61, 0x18, 0x18, 0x18,
		0x60, 0x60, 0x60, 0xf8, 0xff, 0xff, 0xff, 0x3f, 0x03, 0x03, 0x03, 0x43,
		0x30, 0x30, 0x30, 0xfc, 0xff, 0xff, 0xff, 0x5f, 0x80, 0x81, 0x81, 0xe1,
		0x1f, 0xc4, 0xc0, 0xc0, 0xf0, 0x0f, 0x62, 0x60, 0x60, 0xf8, 0x87, 0x30,
		0x30, 0x30, 0xfc, 0xc3, 0x18, 0x18, 0x18, 0xfe, 0x61, 0x0c, 0x0c, 0x0c,
		0xff, 0x30, 0x06, 0x06, 0x86, 0x7f, 0x18, 0x83, 0x80, 0x80, 0xe0, 0x1f,
		0xc6, 0x40, 0x20, 0x20, 0xf8, 0x87, 0x31, 0x28, 0x08, 0x08, 0xfe, 0x61,
		0x0c, 0x0a, 0x02, 0x82, 0x7f, 0x18, 0x83, 0x82, 0x80, 0xe0, 0x1f, 0xc6,
		0xa0, 0x20, 0x20, 0xf8, 0x87, 0x31, 0x28, 0x08, 0x08, 0xfe, 0x61, 0x60,
		0x60, 0x60, 0x00, 0x00, 0x33, 0x00, 0x19, 0x00, 0xf7, 0x5f, 0xbd, 0x42,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	return iconData
}
