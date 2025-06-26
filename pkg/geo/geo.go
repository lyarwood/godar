package geo

import "math"

// CalculateDistance uses the Haversine formula to calculate the distance between two points on Earth.
// lat1, lon1 are the coordinates of the first point (user's location)
// lat2, lon2 are the coordinates of the second point (aircraft's location)
// Returns distance in kilometers
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	la1 := lat1 * math.Pi / 180
	lo1 := lon1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	lo2 := lon2 * math.Pi / 180

	dlat := la2 - la1
	dlon := lo2 - lo1

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(la1)*math.Cos(la2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// CalculateBearing calculates the bearing from point 1 to point 2
// lat1, lon1 are the coordinates of the first point (user's location)
// lat2, lon2 are the coordinates of the second point (aircraft's location)
// Returns bearing in degrees (0-360, where 0 is North, 90 is East, etc.)
func CalculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	la1 := lat1 * math.Pi / 180
	lo1 := lon1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	lo2 := lon2 * math.Pi / 180

	dlon := lo2 - lo1

	y := math.Sin(dlon) * math.Cos(la2)
	x := math.Cos(la1)*math.Sin(la2) - math.Sin(la1)*math.Cos(la2)*math.Cos(dlon)

	bearing := math.Atan2(y, x) * 180 / math.Pi

	// Convert to 0-360 range
	if bearing < 0 {
		bearing += 360
	}

	return bearing
}

// BearingToDirection converts a bearing in degrees to a cardinal direction
// Returns a string like "N", "NE", "E", "SE", "S", "SW", "W", "NW", "NNE", etc.
func BearingToDirection(bearing float64) string {
	// Normalize bearing to 0-360
	bearing = math.Mod(bearing, 360)
	if bearing < 0 {
		bearing += 360
	}

	// Define the 16 cardinal directions
	directions := []string{
		"N", "NNE", "NE", "ENE",
		"E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW",
		"W", "WNW", "NW", "NNW",
	}

	// Each direction covers 22.5 degrees (360/16)
	index := int(math.Round(bearing/22.5)) % 16
	return directions[index]
}

// IsValidLatitude checks if a latitude value is valid
func IsValidLatitude(lat float64) bool {
	return lat >= -90 && lat <= 90
}

// IsValidLongitude checks if a longitude value is valid
func IsValidLongitude(lon float64) bool {
	return lon >= -180 && lon <= 180
}

// IsValidCoordinate checks if both latitude and longitude are valid
func IsValidCoordinate(lat, lon float64) bool {
	return IsValidLatitude(lat) && IsValidLongitude(lon)
}
