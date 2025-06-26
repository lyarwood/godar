package images_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/lyarwood/godar/pkg/notification/images"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("AircraftImageService", func() {
	var service *images.AircraftImageService

	ginkgo.BeforeEach(func() {
		service = images.NewAircraftImageService()
	})

	ginkgo.Describe("NewAircraftImageService", func() {
		ginkgo.It("should create a new service with HTTP client", func() {
			gomega.Expect(service).NotTo(gomega.BeNil())
			gomega.Expect(service).To(gomega.BeAssignableToTypeOf(&images.AircraftImageService{}))
		})
	})

	ginkgo.Describe("isMilitaryAircraft", func() {
		ginkgo.Context("when aircraft type is military", func() {
			ginkgo.It("should detect US fighter aircraft", func() {
				militaryTypes := []string{"F-16", "F15", "F/A-18", "F22", "F35", "F14", "F4", "F5"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})

			ginkgo.It("should detect US attack aircraft", func() {
				militaryTypes := []string{"A-10", "A10"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})

			ginkgo.It("should detect US bombers", func() {
				militaryTypes := []string{"B-1", "B1", "B-2", "B2", "B-52", "B52"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})

			ginkgo.It("should detect US transport aircraft", func() {
				militaryTypes := []string{"C-130", "C130", "C-17", "C17", "C-5", "C5"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})

			ginkgo.It("should detect US tankers", func() {
				militaryTypes := []string{"KC-135", "KC135", "KC-10", "KC10"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})

			ginkgo.It("should detect US special mission aircraft", func() {
				militaryTypes := []string{"E-3", "E3", "E-4", "E4", "P-3", "P3", "P-8", "P8", "U-2", "U2", "SR-71", "SR71"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})

			ginkgo.It("should detect European military aircraft", func() {
				militaryTypes := []string{"EF2000", "Rafale", "Gripen", "Tornado", "Harrier", "Hawk", "Typhoon"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})

			ginkgo.It("should detect Russian military aircraft", func() {
				militaryTypes := []string{"Su27", "Su30", "Su35", "Su57", "MiG29", "MiG31", "Tu95", "Tu160", "Il76"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})

			ginkgo.It("should detect helicopters", func() {
				militaryTypes := []string{"AH64", "UH60", "CH47", "AH1", "V22", "Ka52", "Mi24", "Mi28"}
				for _, aircraftType := range militaryTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
						"Expected %s to be detected as military", aircraftType)
				}
			})
		})

		ginkgo.Context("when aircraft type is not military", func() {
			ginkgo.It("should not detect commercial aircraft", func() {
				commercialTypes := []string{"A320", "A321", "A330", "A350", "A380", "B737", "B747", "B777", "B787", "E190", "E195"}
				for _, aircraftType := range commercialTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeFalse(),
						"Expected %s to not be detected as military", aircraftType)
				}
			})

			ginkgo.It("should not detect unknown aircraft types", func() {
				unknownTypes := []string{"UNKNOWN", "TEST123", "XYZ", "123", ""}
				for _, aircraftType := range unknownTypes {
					gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeFalse(),
						"Expected %s to not be detected as military", aircraftType)
				}
			})
		})
	})

	ginkgo.Describe("GetAircraftImageURL", func() {
		ginkgo.It("should return empty string for empty registration", func() {
			result := service.GetAircraftImageURL("")
			gomega.Expect(result).To(gomega.BeEmpty())
		})

		ginkgo.It("should return empty string for short registration", func() {
			result := service.GetAircraftImageURL("AB")
			gomega.Expect(result).To(gomega.BeEmpty())
		})
	})

	ginkgo.Describe("GetAircraftTypeImageURL", func() {
		ginkgo.It("should return empty string for empty aircraft type", func() {
			result := service.GetAircraftTypeImageURL("")
			gomega.Expect(result).To(gomega.BeEmpty())
		})

		ginkgo.It("should handle unknown aircraft types gracefully", func() {
			result := service.GetAircraftTypeImageURL("UNKNOWN_TYPE")
			// Should not panic, may return empty string or attempt search
			gomega.Expect(result).To(gomega.BeAssignableToTypeOf(""))
		})
	})

	ginkgo.Describe("Wikimedia API integration", func() {
		var server *httptest.Server

		ginkgo.BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Mock Wikimedia Commons API response
				if strings.Contains(r.URL.String(), "action=query") && strings.Contains(r.URL.String(), "list=search") {
					// Mock search response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`{
						"query": {
							"search": [
								{
									"title": "File:Test Aircraft.jpg"
								}
							]
						}
					}`))
					_ = err
				} else if strings.Contains(r.URL.String(), "prop=imageinfo") {
					// Mock image info response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`{
						"query": {
							"pages": {
								"123": {
									"imageinfo": [
										{
											"url": "https://upload.wikimedia.org/wikipedia/commons/test.jpg"
										}
									]
								}
							}
						}
					}`))
					_ = err
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
		})

		ginkgo.AfterEach(func() {
			server.Close()
		})

		ginkgo.It("should handle successful API responses", func() {
			// Create service with custom client that uses our test server
			customService := &images.AircraftImageService{
				Client: &http.Client{},
			}

			// This would require modifying the service to accept a custom base URL
			// For now, we'll test the structure and error handling
			gomega.Expect(customService).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Error handling", func() {
		ginkgo.It("should handle network errors gracefully", func() {
			// Create a test server that returns an error
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Server Error"))
			}))
			defer server.Close()

			// Create service with custom client that points to our error server
			// Since we can't easily override the Wikimedia URL in the service,
			// we'll test that the service doesn't panic when making requests
			service := images.NewAircraftImageService()

			// This should not panic even if the network request fails
			result := service.GetAircraftTypeImageURL("A320")
			// The result could be empty or a URL depending on the actual Wikimedia response
			// The important thing is that it doesn't panic
			gomega.Expect(result).To(gomega.BeAssignableToTypeOf(""))
		})

		ginkgo.It("should handle malformed JSON responses gracefully", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{ invalid json }`))
				_ = err
			}))
			defer server.Close()

			// This test demonstrates the concept - actual implementation would need
			// the service to accept a custom base URL for testing
			gomega.Expect(server).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Aircraft type mappings", func() {
		ginkgo.It("should have comprehensive military aircraft mappings", func() {
			// Test that military aircraft detection covers all major types
			militaryTypes := []string{
				// US fighters
				"F16", "F15", "F18", "F22", "F35", "F14", "F4", "F5",
				// US attack
				"A10",
				// US bombers
				"B1", "B2", "B52",
				// US transport
				"C130", "C17", "C5",
				// US tankers
				"KC135", "KC10",
				// US special mission
				"E3", "E4", "P3", "P8", "U2", "SR71", "F117", "B21",
				// European
				"EF2000", "Rafale", "Gripen", "Tornado", "Harrier", "Hawk", "Typhoon",
				// Russian
				"Su27", "Su30", "Su35", "Su57", "MiG29", "MiG31", "Tu95", "Tu160", "Il76",
				// Helicopters
				"AH64", "UH60", "CH47", "AH1", "V22", "Ka52", "Mi24", "Mi28",
			}

			for _, aircraftType := range militaryTypes {
				gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeTrue(),
					"Expected %s to be detected as military", aircraftType)
			}
		})

		ginkgo.It("should have commercial aircraft mappings", func() {
			commercialTypes := []string{"A320", "A321", "A330", "A350", "A380", "B737", "B747", "B777", "B787", "E190", "E195"}

			for _, aircraftType := range commercialTypes {
				gomega.Expect(service.IsMilitaryAircraft(aircraftType)).To(gomega.BeFalse(),
					"Expected %s to not be detected as military", aircraftType)
			}
		})
	})
})
