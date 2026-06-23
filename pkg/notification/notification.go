package notification

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/lyarwood/godar/pkg/geo"
	"github.com/lyarwood/godar/pkg/notification/images"
	"go.uber.org/zap"
)

// NotificationSender defines the interface for sending notifications
type NotificationSender interface {
	Notify(title, message, iconPath string) error
}

// DefaultNotificationSender implements NotificationSender using beeep
type DefaultNotificationSender struct{}

// Notify sends a notification using beeep
func (d *DefaultNotificationSender) Notify(title, message, iconPath string) error {
	return beeep.Notify(title, message, iconPath)
}

// Notifier handles sending notifications
type Notifier struct {
	enabled          bool
	duration         time.Duration
	logger           *zap.Logger
	cacheDir         string
	imageService     *images.AircraftImageService
	sender           NotificationSender
	viewableDistance float64
	predictionWindow time.Duration
}

// NewNotifier creates a new notification handler
func NewNotifier(enabled bool, duration time.Duration, logger *zap.Logger, viewableDistance float64, predictionWindow time.Duration) *Notifier {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "godar", "images")
	return &Notifier{
		enabled:          enabled,
		duration:         duration,
		logger:           logger,
		cacheDir:         cacheDir,
		imageService:     images.NewAircraftImageService(),
		sender:           &DefaultNotificationSender{},
		viewableDistance: viewableDistance,
		predictionWindow: predictionWindow,
	}
}

// NewNotifierWithSender creates a new notification handler with a custom sender (for testing)
func NewNotifierWithSender(enabled bool, duration time.Duration, logger *zap.Logger, sender NotificationSender, viewableDistance float64, predictionWindow time.Duration) *Notifier {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "godar", "images")
	return &Notifier{
		enabled:          enabled,
		duration:         duration,
		logger:           logger,
		cacheDir:         cacheDir,
		imageService:     images.NewAircraftImageService(),
		sender:           sender,
		viewableDistance: viewableDistance,
		predictionWindow: predictionWindow,
	}
}

// Send sends a desktop notification for a detected aircraft.
func (n *Notifier) Send(callsign, aircraftType string, altitude int, speed float64, distance float64, direction string, heading float64, aircraftLat, aircraftLon, observerLat, observerLon, userHeading float64, previousDistance ...float64) error {
	if !n.enabled {
		return nil
	}

	notificationTitle := fmt.Sprintf("Aircraft Detected: %s", callsign)

	// Calculate bearing to aircraft
	bearing := geo.CalculateBearing(observerLat, observerLon, aircraftLat, aircraftLon)

	// Calculate clock position based on user's heading
	clockPosition := geo.BearingToClockPosition(userHeading, bearing)

	// Build notification message - always include clock position
	directionInfo := fmt.Sprintf("%s (%d o'clock)", direction, clockPosition)

	// Calculate BRAA (Bearing, Range, Altitude, Aspect)
	rangeNm := geo.KmToNauticalMiles(distance)
	aspect := geo.CalculateAspect(heading, bearing)
	braa := fmt.Sprintf("%03.0f/%.0f/%d/%s", bearing, rangeNm, altitude, aspect)

	notificationMessage := fmt.Sprintf("Type: %s\nAltitude: %d ft\nSpeed: %.1f knots\nDistance: %.2f km\nDirection: %s\nBRAA: %s",
		aircraftType, altitude, speed, distance, directionInfo, braa)

	// Add previous distance information if available
	if len(previousDistance) > 0 && previousDistance[0] > 0 {
		distanceChange := previousDistance[0] - distance
		changeDirection := "closer"
		if distanceChange < 0 {
			changeDirection = "farther"
			distanceChange = -distanceChange
		}
		notificationMessage += fmt.Sprintf("\nPrevious: %.2f km (%s by %.2f km)",
			previousDistance[0], changeDirection, distanceChange)
	}

	// Calculate and add closest approach information if aircraft has valid heading and speed
	if heading > 0 && speed > 1.0 && aircraftLat != 0 && aircraftLon != 0 {
		approach := geo.CalculateClosestApproach(observerLat, observerLon, aircraftLat, aircraftLon, heading, speed)
		// Only show prediction if aircraft will approach within viewable distance and prediction window
		if approach.WillApproach && approach.TimeToClosest > 0 && approach.TimeToClosest < n.predictionWindow && approach.Distance <= n.viewableDistance {
			timeStr := geo.FormatTimeToClosest(approach.TimeToClosest)
			notificationMessage += fmt.Sprintf("\nClosest: %.1f km in %s", approach.Distance, timeStr)
		}
	}

	// Try to get aircraft image
	imagePath := n.getAircraftImage(callsign, aircraftType)

	err := n.sender.Notify(notificationTitle, notificationMessage, imagePath)
	if err != nil {
		n.logger.Error("Failed to send notification",
			zap.String("callsign", callsign),
			zap.Error(err))
		return fmt.Errorf("failed to send notification: %w", err)
	}

	n.logger.Debug("Notification sent",
		zap.String("callsign", callsign),
		zap.String("type", aircraftType),
		zap.String("image", imagePath))

	return nil
}

// getAircraftImage attempts to fetch and cache an aircraft image
func (n *Notifier) getAircraftImage(callsign, aircraftType string) string {
	if n.logger == nil {
		return ""
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(n.cacheDir, 0755); err != nil {
		n.logger.Debug("Failed to create cache directory", zap.Error(err))
		return ""
	}

	// Try to get image by registration first, then by aircraft type
	imagePath := n.tryGetImageByRegistration(callsign)
	if imagePath == "" {
		imagePath = n.tryGetImageByType(aircraftType)
	}

	return imagePath
}

// tryGetImageByRegistration attempts to fetch an image using the aircraft registration
func (n *Notifier) tryGetImageByRegistration(registration string) string {
	if registration == "" {
		return ""
	}

	// Clean registration (remove common prefixes/suffixes)
	cleanReg := strings.TrimSpace(strings.ToUpper(registration))
	if len(cleanReg) < 3 {
		return ""
	}

	cacheFile := filepath.Join(n.cacheDir, fmt.Sprintf("%s.jpg", cleanReg))

	// Check if we already have this image cached
	if _, err := os.Stat(cacheFile); err == nil {
		return cacheFile
	}

	// Try to fetch from JetPhotos API (you'll need an API key)
	// For now, we'll use a placeholder approach
	imageURL := n.getAircraftImageURL(cleanReg)
	if imageURL == "" {
		return ""
	}

	if err := n.downloadImage(imageURL, cacheFile); err != nil {
		n.logger.Debug("Failed to download aircraft image",
			zap.String("registration", cleanReg),
			zap.Error(err))
		return ""
	}

	return cacheFile
}

// tryGetImageByType attempts to fetch a generic image for the aircraft type
func (n *Notifier) tryGetImageByType(aircraftType string) string {
	if aircraftType == "" {
		return ""
	}

	cleanType := strings.TrimSpace(strings.ToUpper(aircraftType))
	cacheFile := filepath.Join(n.cacheDir, fmt.Sprintf("type_%s.jpg", cleanType))

	// Check if we already have this type cached
	if _, err := os.Stat(cacheFile); err == nil {
		return cacheFile
	}

	// Try to fetch generic aircraft type image
	imageURL := n.getAircraftTypeImageURL(cleanType)
	if imageURL == "" {
		return ""
	}

	if err := n.downloadImage(imageURL, cacheFile); err != nil {
		n.logger.Debug("Failed to download aircraft type image",
			zap.String("type", cleanType),
			zap.Error(err))
		return ""
	}

	return cacheFile
}

// getAircraftImageURL returns the URL for an aircraft image by registration
func (n *Notifier) getAircraftImageURL(registration string) string {
	return n.imageService.GetAircraftImageURL(registration)
}

// getAircraftTypeImageURL returns the URL for a generic aircraft type image
func (n *Notifier) getAircraftTypeImageURL(aircraftType string) string {
	return n.imageService.GetAircraftTypeImageURL(aircraftType)
}

// downloadImage downloads an image from URL and saves it to the specified path
func (n *Notifier) downloadImage(imageURL, filePath string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}
