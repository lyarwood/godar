package geo

import (
	"fmt"
	"math"
	"time"
)

// ClosestApproach represents the predicted closest approach of an aircraft
type ClosestApproach struct {
	Distance      float64       // Distance at closest approach in km
	TimeToClosest time.Duration // Time until closest approach
	WillApproach  bool          // True if aircraft is getting closer
}

// CalculateClosestApproach predicts when an aircraft will be closest to a location
// based on current position, heading, and speed.
func CalculateClosestApproach(observerLat, observerLon, aircraftLat, aircraftLon, heading, speedKnots float64) *ClosestApproach {
	// Convert speed from knots to km/h
	speedKmh := speedKnots * 1.852

	// If speed is negligible, aircraft is not moving
	if speedKmh < 1.0 {
		currentDist := CalculateDistance(observerLat, observerLon, aircraftLat, aircraftLon)
		return &ClosestApproach{
			Distance:      currentDist,
			TimeToClosest: 0,
			WillApproach:  false,
		}
	}

	// Calculate current distance and bearing from aircraft to observer
	currentDistance := CalculateDistance(observerLat, observerLon, aircraftLat, aircraftLon)
	bearingToObserver := CalculateBearing(aircraftLat, aircraftLon, observerLat, observerLon)

	// Calculate the angle between aircraft heading and bearing to observer
	// This tells us if the aircraft is flying towards or away from the observer
	angleDiff := math.Abs(normalizeAngle(bearingToObserver - heading))

	// If angle is > 90 degrees, aircraft is flying away
	if angleDiff > 90 {
		return &ClosestApproach{
			Distance:      currentDistance,
			TimeToClosest: 0,
			WillApproach:  false,
		}
	}

	// Calculate closest approach using vector math
	// Convert heading to radians
	headingRad := heading * math.Pi / 180.0

	// Project the observer's position onto the aircraft's flight path
	// Using the cosine of the angle between the current position vector and heading
	angleToObserver := bearingToObserver * math.Pi / 180.0

	// Distance along flight path to closest point
	distanceAlongPath := currentDistance * math.Cos((angleToObserver-headingRad))

	// If negative, we've already passed the closest point
	if distanceAlongPath < 0 {
		return &ClosestApproach{
			Distance:      currentDistance,
			TimeToClosest: 0,
			WillApproach:  false,
		}
	}

	// Perpendicular distance from flight path to observer (closest approach distance)
	closestDistance := currentDistance * math.Abs(math.Sin(angleToObserver-headingRad))

	// Time to closest approach
	if speedKmh > 0 {
		timeHours := distanceAlongPath / speedKmh
		timeToClosest := time.Duration(timeHours * float64(time.Hour))

		return &ClosestApproach{
			Distance:      closestDistance,
			TimeToClosest: timeToClosest,
			WillApproach:  true,
		}
	}

	return &ClosestApproach{
		Distance:      currentDistance,
		TimeToClosest: 0,
		WillApproach:  false,
	}
}

// normalizeAngle normalizes an angle to be between -180 and 180 degrees
func normalizeAngle(angle float64) float64 {
	for angle > 180 {
		angle -= 360
	}
	for angle < -180 {
		angle += 360
	}
	return angle
}

// FormatTimeToClosest formats a duration into a human-readable string
func FormatTimeToClosest(d time.Duration) string {
	if d < time.Minute {
		return "< 1 min"
	} else if d < time.Hour {
		mins := int(d.Minutes())
		return fmt.Sprintf("%d min", mins)
	} else {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
}
