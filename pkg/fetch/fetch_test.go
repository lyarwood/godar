package fetch_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/lyarwood/godar/pkg/aircraft"
	"github.com/lyarwood/godar/pkg/fetch"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFetch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fetch Package")
}

var _ = Describe("Fetcher", func() {
	var (
		server               *httptest.Server
		expectedAircraftList *aircraft.AircraftList
	)

	BeforeEach(func() {
		expectedAircraftList = &aircraft.AircraftList{
			LastDv:  aircraft.LastDvValue(12345),
			TotalAc: 1,
			Src:     1,
			Stm:     time.Now().UnixNano() / int64(time.Millisecond),
			Aircraft: []aircraft.Aircraft{
				{
					ID:   1,
					Icao: "ABCDEF",
					Call: "TEST123",
					Type: "A320",
					Alt:  38000,
					Lat:  51.5,
					Long: -0.1,
					Mil:  false,
					Op:   "Test Airlines",
				},
			},
		}
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("Fetch with filters", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check query parameters
				query := r.URL.Query()
				Expect(query.Get("fTypQ")).To(Equal("A320"))
				Expect(query.Get("fAltL")).To(Equal("10000"))
				Expect(query.Get("fAltU")).To(Equal("40000"))
				Expect(query.Get("fMilQ")).To(Equal("1"))
				Expect(query.Get("fOpQ")).To(Equal("Test Airlines"))
				Expect(query.Get("fCallQ")).To(Equal("TEST123"))
				Expect(query.Get("lat")).To(Equal("51.5"))
				Expect(query.Get("lng")).To(Equal("-0.1"))
				Expect(query.Get("fDstU")).To(Equal("100"))

				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(expectedAircraftList); err != nil {
					panic(err)
				}
			}))
		})

		It("should fetch aircraft data with correct query parameters", func() {
			fetcher := fetch.NewFetcher(server.URL, nil)
			fetcher.SetFilters("A320", 10000, 40000, true, "Test Airlines", "TEST123")
			fetcher.SetLocation(51.5, -0.1, 100)
			acList, err := fetcher.Fetch()
			Expect(err).NotTo(HaveOccurred())
			Expect(acList).To(Equal(expectedAircraftList))
		})
	})

	Describe("Fetch with HTTP Basic Auth", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check Basic Auth
				username, password, ok := r.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("testuser"))
				Expect(password).To(Equal("testpass"))

				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(expectedAircraftList); err != nil {
					panic(err)
				}
			}))
		})

		It("should send HTTP Basic Auth headers", func() {
			fetcher := fetch.NewFetcher(server.URL, nil)
			fetcher.SetAuth("testuser", "testpass")
			acList, err := fetcher.Fetch()
			Expect(err).NotTo(HaveOccurred())
			Expect(acList).To(Equal(expectedAircraftList))
		})

		It("should work without Basic Auth when credentials are empty", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Should not have Basic Auth
				_, _, ok := r.BasicAuth()
				Expect(ok).To(BeFalse())

				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(expectedAircraftList); err != nil {
					panic(err)
				}
			}))

			fetcher := fetch.NewFetcher(server.URL, nil)
			acList, err := fetcher.Fetch()
			Expect(err).NotTo(HaveOccurred())
			Expect(acList).To(Equal(expectedAircraftList))
		})
	})

	Describe("Fetch with partial filters", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()
				// Only some filters should be present
				Expect(query.Get("fTypQ")).To(Equal("A320"))
				Expect(query.Get("fAltL")).To(Equal(""))
				Expect(query.Get("fAltU")).To(Equal(""))
				Expect(query.Get("fMilQ")).To(Equal(""))

				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(expectedAircraftList); err != nil {
					panic(err)
				}
			}))
		})

		It("should only send non-empty filter parameters", func() {
			fetcher := fetch.NewFetcher(server.URL, nil)
			fetcher.SetFilters("A320", 0, 0, false, "", "")
			acList, err := fetcher.Fetch()
			Expect(err).NotTo(HaveOccurred())
			Expect(acList).To(Equal(expectedAircraftList))
		})
	})

	Describe("Error handling", func() {
		It("should handle invalid URL", func() {
			fetcher := fetch.NewFetcher("invalid://url", nil)
			_, err := fetcher.Fetch()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported protocol scheme"))
		})

		It("should handle server errors", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal Server Error"))
			}))

			fetcher := fetch.NewFetcher(server.URL, nil)
			_, err := fetcher.Fetch()
			Expect(err).To(HaveOccurred())
		})

		It("should handle invalid JSON response", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("invalid json"))
			}))

			fetcher := fetch.NewFetcher(server.URL, nil)
			_, err := fetcher.Fetch()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Distance filtering", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()
				Expect(query.Get("lat")).To(Equal("51.5"))
				Expect(query.Get("lng")).To(Equal("-0.1"))
				Expect(query.Get("fDstU")).To(Equal("100"))

				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(expectedAircraftList); err != nil {
					panic(err)
				}
			}))
		})

		It("should include distance parameters when location is provided", func() {
			fetcher := fetch.NewFetcher(server.URL, nil)
			fetcher.SetLocation(51.5, -0.1, 100)
			acList, err := fetcher.Fetch()
			Expect(err).NotTo(HaveOccurred())
			Expect(acList).To(Equal(expectedAircraftList))
		})
	})

	Describe("VRS JSON compatibility", func() {
		It("should decode real VRS JSON data without error", func() {
			data, err := os.ReadFile("vrs_testdata.json")
			Expect(err).NotTo(HaveOccurred())
			var acList aircraft.AircraftList
			err = json.Unmarshal(data, &acList)
			Expect(err).NotTo(HaveOccurred())
			Expect(acList.Aircraft).To(HaveLen(2))
			Expect(acList.Aircraft[0].Call).To(Equal("TOM88K"))
			Expect(acList.Aircraft[1].Call).To(Equal("RYR39ZW"))
		})
	})
})
