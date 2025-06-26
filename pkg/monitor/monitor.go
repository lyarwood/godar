package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lyarwood/godar/pkg/aircraft"
	"github.com/lyarwood/godar/pkg/config"
	"github.com/lyarwood/godar/pkg/fetch"
	"github.com/lyarwood/godar/pkg/geo"
	"github.com/lyarwood/godar/pkg/notification"

	"go.uber.org/zap"
)

// Fetcher defines the interface for fetching aircraft data
// (You can use mockgen or hand-written mocks)
//
//go:generate mockgen -destination=fetcher_mock.go -package=monitor godar/pkg/monitor Fetcher
type Fetcher interface {
	Fetch() (*aircraft.AircraftList, error)
	SetFilters(aircraftType string, minAlt, maxAlt int, military bool, operator, flightNumber string)
	SetLocation(lat, lon, maxDistance float64)
	SetAuth(username, password string)
}

// Notifier defines the interface for sending notifications
type Notifier interface {
	Send(callsign, aircraftType string, altitude int, speed float64, distance float64, direction string, previousDistance ...float64) error
}

// AircraftTracker tracks aircraft distance history
type AircraftTracker struct {
	LastDistance float64
	LastSeen     time.Time
	Notified     bool
}

// Monitor represents the aircraft monitoring service
type Monitor struct {
	config          *config.Config
	fetcher         Fetcher
	notifier        Notifier
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	stopChan        chan struct{}
	aircraftHistory map[string]*AircraftTracker // Key: ICAO or callsign
	historyMutex    sync.RWMutex
}

// NewMonitorWithDeps creates a new monitoring service with injected dependencies
func NewMonitorWithDeps(cfg *config.Config, logger *zap.Logger, fetcher Fetcher, notifier Notifier) (*Monitor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		config:          cfg,
		fetcher:         fetcher,
		notifier:        notifier,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		stopChan:        make(chan struct{}),
		aircraftHistory: make(map[string]*AircraftTracker),
		historyMutex:    sync.RWMutex{},
	}, nil
}

// NewMonitor creates a new monitoring service
func NewMonitor(cfg *config.Config, logger *zap.Logger) (*Monitor, error) {
	// Create fetcher with logger
	fetcher := fetch.NewFetcher(cfg.Server.URL, logger)

	// Set filters
	fetcher.SetFilters(
		cfg.Filters.AircraftType,
		cfg.Filters.MinAltitude,
		cfg.Filters.MaxAltitude,
		cfg.Filters.Military,
		cfg.Filters.Operator,
		cfg.Filters.FlightNumber,
	)

	// Set location
	fetcher.SetLocation(
		cfg.Location.Latitude,
		cfg.Location.Longitude,
		cfg.Location.MaxDistance,
	)

	// Set auth credentials if provided
	if cfg.Server.Username != "" || cfg.Server.Password != "" {
		fetcher.SetAuth(cfg.Server.Username, cfg.Server.Password)
	}

	notifier := notification.NewNotifier(cfg.Notification.Enabled, cfg.Notification.Duration, logger)

	return NewMonitorWithDeps(cfg, logger, fetcher, notifier)
}

// Start begins the monitoring process
func (m *Monitor) Start() error {
	m.logger.Info("Starting aircraft monitoring",
		zap.String("server", m.config.Server.URL),
		zap.Duration("poll_interval", m.config.Monitoring.PollInterval))

	m.wg.Add(1)
	go m.monitorLoop()

	return nil
}

// Stop gracefully stops the monitoring service
func (m *Monitor) Stop() {
	m.logger.Info("Stopping aircraft monitoring")

	// Signal stop
	close(m.stopChan)

	// Cancel context
	m.cancel()

	// Wait for goroutines to finish
	m.wg.Wait()

	m.logger.Info("Aircraft monitoring stopped")
}

// monitorLoop is the main monitoring loop
func (m *Monitor) monitorLoop() {
	defer m.wg.Done()

	// Perform initial fetch immediately
	if err := m.fetchAndProcess(); err != nil {
		m.logger.Error("Failed to fetch and process aircraft data on startup", zap.Error(err))
	}

	// Set up ticker for subsequent polls
	ticker := time.NewTicker(m.config.Monitoring.PollInterval)
	defer ticker.Stop()

	// Set up cleanup ticker (clean up every 10 minutes)
	cleanupInterval := m.config.Notification.CleanupInterval
	if cleanupInterval == 0 {
		cleanupInterval = 10 * time.Minute // Default cleanup interval
	}
	cleanupTicker := time.NewTicker(cleanupInterval)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			if err := m.fetchAndProcess(); err != nil {
				m.logger.Error("Failed to fetch and process aircraft data", zap.Error(err))
			}
		case <-cleanupTicker.C:
			m.cleanupAircraftHistory()
		}
	}
}

// fetchAndProcess fetches aircraft data and processes it
func (m *Monitor) fetchAndProcess() error {
	acList, err := m.fetcher.Fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch aircraft data: %w", err)
	}

	m.logger.Debug("Fetched aircraft data",
		zap.Int("total_aircraft", acList.TotalAc),
		zap.Int("filtered_aircraft", len(acList.Aircraft)))

	// Process each aircraft
	for _, ac := range acList.Aircraft {
		if err := m.processAircraft(ac); err != nil {
			m.logger.Error("Failed to process aircraft",
				zap.String("callsign", ac.Call),
				zap.Error(err))
		}
	}

	return nil
}

// processAircraft processes a single aircraft
func (m *Monitor) processAircraft(ac aircraft.Aircraft) error {
	distance := m.calculateDistance(ac.Lat, ac.Long)
	bearing := m.calculateBearing(ac.Lat, ac.Long)
	direction := geo.BearingToDirection(bearing)

	// Get aircraft identifier (prefer ICAO, fallback to callsign)
	aircraftID := m.getAircraftIdentifier(ac)

	// Get previous distance before updating tracker
	var previousDistance float64
	m.historyMutex.RLock()
	if tracker, exists := m.aircraftHistory[aircraftID]; exists {
		previousDistance = tracker.LastDistance
	}
	m.historyMutex.RUnlock()

	// Check if we should notify based on distance tracking
	shouldNotify := m.shouldNotifyAircraft(aircraftID, distance)

	m.logger.Info("Aircraft detected",
		zap.String("callsign", ac.Call),
		zap.String("icao", ac.Icao),
		zap.String("type", ac.Type),
		zap.Int("altitude", ac.Alt),
		zap.Float64("distance_km", distance),
		zap.Float64("bearing_degrees", bearing),
		zap.String("direction", direction),
		zap.Float64("previous_distance_km", previousDistance),
		zap.Bool("military", ac.Mil),
		zap.Bool("notifying", shouldNotify))

	// Send notification if enabled and aircraft is getting closer
	if m.config.Notification.Enabled && shouldNotify {
		if err := m.notifier.Send(ac.Call, ac.Type, ac.Alt, ac.Spd, distance, direction, previousDistance); err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}
	}

	return nil
}

// getAircraftIdentifier returns a unique identifier for the aircraft
func (m *Monitor) getAircraftIdentifier(ac aircraft.Aircraft) string {
	if ac.Icao != "" {
		return ac.Icao
	}
	if ac.Call != "" {
		return ac.Call
	}
	// Fallback to a combination of type and altitude if no other identifier
	return fmt.Sprintf("%s_%d", ac.Type, ac.Alt)
}

// shouldNotifyAircraft determines if we should notify about this aircraft
func (m *Monitor) shouldNotifyAircraft(aircraftID string, currentDistance float64) bool {
	m.historyMutex.Lock()
	defer m.historyMutex.Unlock()

	tracker, exists := m.aircraftHistory[aircraftID]
	now := time.Now()

	if !exists {
		// First time seeing this aircraft - always notify
		m.aircraftHistory[aircraftID] = &AircraftTracker{
			LastDistance: currentDistance,
			LastSeen:     now,
			Notified:     true,
		}
		return true
	}

	// Check if aircraft is getting closer (distance is decreasing)
	isGettingCloser := currentDistance < tracker.LastDistance

	// Check if we should re-notify based on time interval
	shouldNotifyByTime := false
	if m.config.Notification.ReNotifyAfter > 0 {
		timeSinceLastSeen := now.Sub(tracker.LastSeen)
		shouldNotifyByTime = timeSinceLastSeen > m.config.Notification.ReNotifyAfter
	}

	// Update tracker
	tracker.LastDistance = currentDistance
	tracker.LastSeen = now

	// Determine if we should notify based on configuration
	if m.config.Notification.NotifyOnCloserOnly {
		// Only notify if getting closer or if re-notify time has passed
		if isGettingCloser || shouldNotifyByTime {
			tracker.Notified = true
			return true
		}
	} else {
		// Original behavior: always notify
		tracker.Notified = true
		return true
	}

	// Aircraft is moving away or staying at same distance
	return false
}

// calculateDistance calculates the distance between user location and aircraft
func (m *Monitor) calculateDistance(acLat, acLong float64) float64 {
	if m.config.Location.Latitude == 0.0 && m.config.Location.Longitude == 0.0 {
		return 0.0
	}
	return geo.CalculateDistance(m.config.Location.Latitude, m.config.Location.Longitude, acLat, acLong)
}

// calculateBearing calculates the bearing between user location and aircraft
func (m *Monitor) calculateBearing(acLat, acLong float64) float64 {
	if m.config.Location.Latitude == 0.0 && m.config.Location.Longitude == 0.0 {
		return 0.0
	}
	return geo.CalculateBearing(m.config.Location.Latitude, m.config.Location.Longitude, acLat, acLong)
}

// cleanupAircraftHistory removes aircraft that haven't been seen recently
func (m *Monitor) cleanupAircraftHistory() {
	m.historyMutex.Lock()
	defer m.historyMutex.Unlock()

	now := time.Now()
	cutoffTime := now.Add(-15 * time.Minute) // Remove aircraft not seen in 15 minutes

	initialCount := len(m.aircraftHistory)

	for aircraftID, tracker := range m.aircraftHistory {
		if tracker.LastSeen.Before(cutoffTime) {
			delete(m.aircraftHistory, aircraftID)
		}
	}

	finalCount := len(m.aircraftHistory)
	if initialCount != finalCount {
		m.logger.Debug("Cleaned up aircraft history",
			zap.Int("removed", initialCount-finalCount),
			zap.Int("remaining", finalCount))
	}
}
