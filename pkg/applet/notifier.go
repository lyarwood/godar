package applet

import (
	"time"

	"github.com/lyarwood/godar/pkg/notification"
	"go.uber.org/zap"
)

// AppletNotifier wraps the standard notifier and also updates the applet
type AppletNotifier struct {
	applet   *Applet
	notifier *notification.Notifier
}

// NewAppletNotifier creates a notifier that updates the applet
func NewAppletNotifier(applet *Applet, enabled bool, duration time.Duration, logger *zap.Logger, viewableDistance float64, predictionWindow time.Duration) *AppletNotifier {
	return &AppletNotifier{
		applet:   applet,
		notifier: notification.NewNotifier(enabled, duration, logger, viewableDistance, predictionWindow),
	}
}

// Send sends a notification and updates the applet's recent aircraft list
func (an *AppletNotifier) Send(callsign, aircraftType string, altitude int, speed float64, distance float64, direction string, heading float64, aircraftLat, aircraftLon, observerLat, observerLon, userHeading float64, previousDistance ...float64) error {
	// Add to applet's recent aircraft list
	an.applet.AddAircraftDetection(callsign, aircraftType, altitude, distance)

	// Send the actual notification
	return an.notifier.Send(callsign, aircraftType, altitude, speed, distance, direction, heading, aircraftLat, aircraftLon, observerLat, observerLon, userHeading, previousDistance...)
}
