package geo_test

import (
	"testing"

	"github.com/lyarwood/godar/pkg/geo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGeo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Geo Package")
}

var _ = Describe("Geo Utils", func() {
	Describe("CalculateDistance", func() {
		It("should calculate distance between two points", func() {
			// London (51.5074, 0.1278) to Paris (48.8566, 2.3522)
			distance := geo.CalculateDistance(51.5074, 0.1278, 48.8566, 2.3522)
			Expect(distance).To(BeNumerically("~", 334.6, 0.5)) // Approximately 334.6 km
		})

		It("should return 0 for the same point", func() {
			distance := geo.CalculateDistance(51.5, 0.0, 51.5, 0.0)
			Expect(distance).To(BeNumerically("~", 0.0, 0.001))
		})

		It("should calculate distance across the equator", func() {
			// Point near equator to another point near equator
			distance := geo.CalculateDistance(0.0, 0.0, 0.0, 1.0)
			Expect(distance).To(BeNumerically("~", 111.0, 1.0)) // Approximately 111 km per degree at equator
		})

		It("should calculate distance across the international date line", func() {
			// Point near 180° to another point near -180°
			distance := geo.CalculateDistance(0.0, 179.0, 0.0, -179.0)
			Expect(distance).To(BeNumerically("~", 222.0, 2.0)) // Approximately 222 km
		})

		It("should calculate distance between poles", func() {
			// North pole to south pole
			distance := geo.CalculateDistance(90.0, 0.0, -90.0, 0.0)
			Expect(distance).To(BeNumerically("~", 20015.0, 10.0)) // Approximately 20,015 km (half Earth's circumference)
		})

		It("should handle negative coordinates", func() {
			// New York (40.7128, -74.0060) to London (51.5074, -0.1278)
			distance := geo.CalculateDistance(40.7128, -74.0060, 51.5074, -0.1278)
			Expect(distance).To(BeNumerically("~", 5570.0, 10.0)) // Approximately 5,570 km
		})
	})

	Describe("IsValidLatitude", func() {
		It("should return true for valid latitudes", func() {
			Expect(geo.IsValidLatitude(0.0)).To(BeTrue())
			Expect(geo.IsValidLatitude(45.0)).To(BeTrue())
			Expect(geo.IsValidLatitude(90.0)).To(BeTrue())
			Expect(geo.IsValidLatitude(-45.0)).To(BeTrue())
			Expect(geo.IsValidLatitude(-90.0)).To(BeTrue())
		})

		It("should return false for invalid latitudes", func() {
			Expect(geo.IsValidLatitude(90.1)).To(BeFalse())
			Expect(geo.IsValidLatitude(-90.1)).To(BeFalse())
			Expect(geo.IsValidLatitude(100.0)).To(BeFalse())
			Expect(geo.IsValidLatitude(-100.0)).To(BeFalse())
		})
	})

	Describe("IsValidLongitude", func() {
		It("should return true for valid longitudes", func() {
			Expect(geo.IsValidLongitude(0.0)).To(BeTrue())
			Expect(geo.IsValidLongitude(90.0)).To(BeTrue())
			Expect(geo.IsValidLongitude(180.0)).To(BeTrue())
			Expect(geo.IsValidLongitude(-90.0)).To(BeTrue())
			Expect(geo.IsValidLongitude(-180.0)).To(BeTrue())
		})

		It("should return false for invalid longitudes", func() {
			Expect(geo.IsValidLongitude(180.1)).To(BeFalse())
			Expect(geo.IsValidLongitude(-180.1)).To(BeFalse())
			Expect(geo.IsValidLongitude(200.0)).To(BeFalse())
			Expect(geo.IsValidLongitude(-200.0)).To(BeFalse())
		})
	})

	Describe("IsValidCoordinate", func() {
		It("should return true for valid coordinates", func() {
			Expect(geo.IsValidCoordinate(0.0, 0.0)).To(BeTrue())
			Expect(geo.IsValidCoordinate(45.0, 90.0)).To(BeTrue())
			Expect(geo.IsValidCoordinate(-45.0, -90.0)).To(BeTrue())
			Expect(geo.IsValidCoordinate(90.0, 180.0)).To(BeTrue())
			Expect(geo.IsValidCoordinate(-90.0, -180.0)).To(BeTrue())
		})

		It("should return false for invalid coordinates", func() {
			// Invalid latitude
			Expect(geo.IsValidCoordinate(90.1, 0.0)).To(BeFalse())
			Expect(geo.IsValidCoordinate(-90.1, 0.0)).To(BeFalse())

			// Invalid longitude
			Expect(geo.IsValidCoordinate(0.0, 180.1)).To(BeFalse())
			Expect(geo.IsValidCoordinate(0.0, -180.1)).To(BeFalse())

			// Both invalid
			Expect(geo.IsValidCoordinate(100.0, 200.0)).To(BeFalse())
			Expect(geo.IsValidCoordinate(-100.0, -200.0)).To(BeFalse())
		})
	})

	Describe("CalculateBearing", func() {
		It("should calculate bearing from north", func() {
			// From (0,0) to (1,0) should be north (0 degrees)
			bearing := geo.CalculateBearing(0.0, 0.0, 1.0, 0.0)
			Expect(bearing).To(BeNumerically("~", 0.0, 1.0))
		})

		It("should calculate bearing from east", func() {
			// From (0,0) to (0,1) should be east (90 degrees)
			bearing := geo.CalculateBearing(0.0, 0.0, 0.0, 1.0)
			Expect(bearing).To(BeNumerically("~", 90.0, 1.0))
		})

		It("should calculate bearing from south", func() {
			// From (0,0) to (-1,0) should be south (180 degrees)
			bearing := geo.CalculateBearing(0.0, 0.0, -1.0, 0.0)
			Expect(bearing).To(BeNumerically("~", 180.0, 1.0))
		})

		It("should calculate bearing from west", func() {
			// From (0,0) to (0,-1) should be west (270 degrees)
			bearing := geo.CalculateBearing(0.0, 0.0, 0.0, -1.0)
			Expect(bearing).To(BeNumerically("~", 270.0, 1.0))
		})

		It("should calculate bearing from northeast", func() {
			// From (0,0) to (1,1) should be northeast (45 degrees)
			bearing := geo.CalculateBearing(0.0, 0.0, 1.0, 1.0)
			Expect(bearing).To(BeNumerically("~", 45.0, 1.0))
		})

		It("should handle same point", func() {
			bearing := geo.CalculateBearing(51.5, 0.0, 51.5, 0.0)
			Expect(bearing).To(BeNumerically("~", 0.0, 1.0))
		})
	})

	Describe("BearingToDirection", func() {
		It("should convert north bearing to N", func() {
			Expect(geo.BearingToDirection(0.0)).To(Equal("N"))
			Expect(geo.BearingToDirection(359.9)).To(Equal("N"))
		})

		It("should convert northeast bearing to NE", func() {
			Expect(geo.BearingToDirection(45.0)).To(Equal("NE"))
		})

		It("should convert east bearing to E", func() {
			Expect(geo.BearingToDirection(90.0)).To(Equal("E"))
		})

		It("should convert southeast bearing to SE", func() {
			Expect(geo.BearingToDirection(135.0)).To(Equal("SE"))
		})

		It("should convert south bearing to S", func() {
			Expect(geo.BearingToDirection(180.0)).To(Equal("S"))
		})

		It("should convert southwest bearing to SW", func() {
			Expect(geo.BearingToDirection(225.0)).To(Equal("SW"))
		})

		It("should convert west bearing to W", func() {
			Expect(geo.BearingToDirection(270.0)).To(Equal("W"))
		})

		It("should convert northwest bearing to NW", func() {
			Expect(geo.BearingToDirection(315.0)).To(Equal("NW"))
		})

		It("should convert intermediate bearings correctly", func() {
			Expect(geo.BearingToDirection(22.5)).To(Equal("NNE"))
			Expect(geo.BearingToDirection(67.5)).To(Equal("ENE"))
			Expect(geo.BearingToDirection(112.5)).To(Equal("ESE"))
			Expect(geo.BearingToDirection(157.5)).To(Equal("SSE"))
			Expect(geo.BearingToDirection(202.5)).To(Equal("SSW"))
			Expect(geo.BearingToDirection(247.5)).To(Equal("WSW"))
			Expect(geo.BearingToDirection(292.5)).To(Equal("WNW"))
			Expect(geo.BearingToDirection(337.5)).To(Equal("NNW"))
		})

		It("should handle negative bearings", func() {
			Expect(geo.BearingToDirection(-45.0)).To(Equal("NW"))
			Expect(geo.BearingToDirection(-90.0)).To(Equal("W"))
		})

		It("should handle bearings over 360 degrees", func() {
			Expect(geo.BearingToDirection(405.0)).To(Equal("NE"))
			Expect(geo.BearingToDirection(720.0)).To(Equal("N"))
		})
	})
})
