package aircraft

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Aircraft Types", func() {
	Describe("LastDvValue", func() {
		It("should unmarshal string value correctly", func() {
			data := []byte(`"12345"`)
			var ldv LastDvValue
			err := json.Unmarshal(data, &ldv)

			Expect(err).To(BeNil())
			Expect(ldv.Int64()).To(Equal(int64(12345)))
		})

		It("should unmarshal numeric value correctly", func() {
			data := []byte(`67890`)
			var ldv LastDvValue
			err := json.Unmarshal(data, &ldv)

			Expect(err).To(BeNil())
			Expect(ldv.Int64()).To(Equal(int64(67890)))
		})

		It("should return error for invalid string", func() {
			data := []byte(`"invalid"`)
			var ldv LastDvValue
			err := json.Unmarshal(data, &ldv)

			Expect(err).NotTo(BeNil())
		})

		It("should return error for invalid JSON", func() {
			data := []byte(`{invalid}`)
			var ldv LastDvValue
			err := json.Unmarshal(data, &ldv)

			Expect(err).NotTo(BeNil())
		})
	})

	Describe("SqkValue", func() {
		It("should unmarshal string value correctly", func() {
			data := []byte(`"7500"`)
			var sqk SqkValue
			err := json.Unmarshal(data, &sqk)

			Expect(err).To(BeNil())
			Expect(int(sqk)).To(Equal(7500))
		})

		It("should unmarshal numeric value correctly", func() {
			data := []byte(`7600`)
			var sqk SqkValue
			err := json.Unmarshal(data, &sqk)

			Expect(err).To(BeNil())
			Expect(int(sqk)).To(Equal(7600))
		})

		It("should handle empty string as zero", func() {
			data := []byte(`""`)
			var sqk SqkValue
			err := json.Unmarshal(data, &sqk)

			Expect(err).To(BeNil())
			Expect(int(sqk)).To(Equal(0))
		})

		It("should return error for invalid string", func() {
			data := []byte(`"invalid"`)
			var sqk SqkValue
			err := json.Unmarshal(data, &sqk)

			Expect(err).NotTo(BeNil())
		})

		It("should return error for invalid JSON", func() {
			data := []byte(`{invalid}`)
			var sqk SqkValue
			err := json.Unmarshal(data, &sqk)

			Expect(err).NotTo(BeNil())
		})
	})

	Describe("WTCValue", func() {
		It("should unmarshal string value correctly", func() {
			data := []byte(`"M"`)
			var wtc WTCValue
			err := json.Unmarshal(data, &wtc)

			Expect(err).To(BeNil())
			Expect(string(wtc)).To(Equal("M"))
		})

		It("should unmarshal numeric value as string", func() {
			data := []byte(`1`)
			var wtc WTCValue
			err := json.Unmarshal(data, &wtc)

			Expect(err).To(BeNil())
			Expect(string(wtc)).To(Equal("1"))
		})

		It("should handle empty string", func() {
			data := []byte(`""`)
			var wtc WTCValue
			err := json.Unmarshal(data, &wtc)

			Expect(err).To(BeNil())
			Expect(string(wtc)).To(Equal(""))
		})

		It("should return error for invalid JSON", func() {
			data := []byte(`{invalid}`)
			var wtc WTCValue
			err := json.Unmarshal(data, &wtc)

			Expect(err).NotTo(BeNil())
		})
	})

	Describe("SpeciesValue", func() {
		It("should unmarshal string value correctly", func() {
			data := []byte(`"Landplane"`)
			var species SpeciesValue
			err := json.Unmarshal(data, &species)

			Expect(err).To(BeNil())
			Expect(string(species)).To(Equal("Landplane"))
		})

		It("should unmarshal numeric value as string", func() {
			data := []byte(`2`)
			var species SpeciesValue
			err := json.Unmarshal(data, &species)

			Expect(err).To(BeNil())
			Expect(string(species)).To(Equal("2"))
		})

		It("should handle empty string", func() {
			data := []byte(`""`)
			var species SpeciesValue
			err := json.Unmarshal(data, &species)

			Expect(err).To(BeNil())
			Expect(string(species)).To(Equal(""))
		})

		It("should return error for invalid JSON", func() {
			data := []byte(`{invalid}`)
			var species SpeciesValue
			err := json.Unmarshal(data, &species)

			Expect(err).NotTo(BeNil())
		})
	})

	Describe("EngTypeValue", func() {
		It("should unmarshal string value correctly", func() {
			data := []byte(`"Jet"`)
			var engType EngTypeValue
			err := json.Unmarshal(data, &engType)

			Expect(err).To(BeNil())
			Expect(string(engType)).To(Equal("Jet"))
		})

		It("should unmarshal numeric value as string", func() {
			data := []byte(`1`)
			var engType EngTypeValue
			err := json.Unmarshal(data, &engType)

			Expect(err).To(BeNil())
			Expect(string(engType)).To(Equal("1"))
		})

		It("should handle empty string", func() {
			data := []byte(`""`)
			var engType EngTypeValue
			err := json.Unmarshal(data, &engType)

			Expect(err).To(BeNil())
			Expect(string(engType)).To(Equal(""))
		})

		It("should return error for invalid JSON", func() {
			data := []byte(`{invalid}`)
			var engType EngTypeValue
			err := json.Unmarshal(data, &engType)

			Expect(err).NotTo(BeNil())
		})
	})

	Describe("EngMountValue", func() {
		It("should unmarshal string value correctly", func() {
			data := []byte(`"Wing"`)
			var engMount EngMountValue
			err := json.Unmarshal(data, &engMount)

			Expect(err).To(BeNil())
			Expect(string(engMount)).To(Equal("Wing"))
		})

		It("should unmarshal numeric value as string", func() {
			data := []byte(`1`)
			var engMount EngMountValue
			err := json.Unmarshal(data, &engMount)

			Expect(err).To(BeNil())
			Expect(string(engMount)).To(Equal("1"))
		})

		It("should handle empty string", func() {
			data := []byte(`""`)
			var engMount EngMountValue
			err := json.Unmarshal(data, &engMount)

			Expect(err).To(BeNil())
			Expect(string(engMount)).To(Equal(""))
		})

		It("should return error for invalid JSON", func() {
			data := []byte(`{invalid}`)
			var engMount EngMountValue
			err := json.Unmarshal(data, &engMount)

			Expect(err).NotTo(BeNil())
		})
	})

	Describe("Aircraft struct", func() {
		It("should marshal and unmarshal correctly", func() {
			aircraft := Aircraft{
				ID:     12345,
				TSecs:  1000,
				Rcvr:   1,
				Icao:   "ABCD1234",
				Reg:    "G-ABCD",
				Alt:    30000,
				Call:   "BA123",
				Lat:    51.5074,
				Long:   -0.1278,
				Type:   "A320",
				Op:     "British Airways",
				Mil:    false,
				HasPic: true,
			}

			data, err := json.Marshal(aircraft)
			Expect(err).To(BeNil())

			var unmarshaled Aircraft
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).To(BeNil())

			Expect(unmarshaled.ID).To(Equal(aircraft.ID))
			Expect(unmarshaled.Icao).To(Equal(aircraft.Icao))
			Expect(unmarshaled.Reg).To(Equal(aircraft.Reg))
			Expect(unmarshaled.Alt).To(Equal(aircraft.Alt))
			Expect(unmarshaled.Call).To(Equal(aircraft.Call))
			Expect(unmarshaled.Lat).To(Equal(aircraft.Lat))
			Expect(unmarshaled.Long).To(Equal(aircraft.Long))
			Expect(unmarshaled.Type).To(Equal(aircraft.Type))
			Expect(unmarshaled.Op).To(Equal(aircraft.Op))
			Expect(unmarshaled.Mil).To(Equal(aircraft.Mil))
			Expect(unmarshaled.HasPic).To(Equal(aircraft.HasPic))
		})

		It("should handle optional fields correctly", func() {
			aircraft := Aircraft{
				ID:    12345,
				TSecs: 1000,
				Rcvr:  1,
			}

			data, err := json.Marshal(aircraft)
			Expect(err).To(BeNil())

			var unmarshaled Aircraft
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).To(BeNil())

			Expect(unmarshaled.ID).To(Equal(aircraft.ID))
			Expect(unmarshaled.Icao).To(BeEmpty())
			Expect(unmarshaled.Reg).To(BeEmpty())
			Expect(unmarshaled.Alt).To(Equal(0))
		})
	})

	Describe("AircraftList struct", func() {
		It("should marshal and unmarshal correctly", func() {
			aircraftList := AircraftList{
				LastDv:  LastDvValue(12345),
				TotalAc: 5,
				Src:     1,
				ShowSil: true,
				ShowFlg: false,
				ShowPic: true,
				Stm:     1640995200,
				Aircraft: []Aircraft{
					{ID: 1, Icao: "ABCD1234", Alt: 30000},
					{ID: 2, Icao: "EFGH5678", Alt: 25000},
				},
				Feeds: []Feed{
					{ID: 1, Name: "Feed 1"},
					{ID: 2, Name: "Feed 2"},
				},
			}

			data, err := json.Marshal(aircraftList)
			Expect(err).To(BeNil())

			var unmarshaled AircraftList
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).To(BeNil())

			Expect(unmarshaled.LastDv.Int64()).To(Equal(aircraftList.LastDv.Int64()))
			Expect(unmarshaled.TotalAc).To(Equal(aircraftList.TotalAc))
			Expect(unmarshaled.Src).To(Equal(aircraftList.Src))
			Expect(unmarshaled.ShowSil).To(Equal(aircraftList.ShowSil))
			Expect(unmarshaled.ShowFlg).To(Equal(aircraftList.ShowFlg))
			Expect(unmarshaled.ShowPic).To(Equal(aircraftList.ShowPic))
			Expect(unmarshaled.Stm).To(Equal(aircraftList.Stm))
			Expect(len(unmarshaled.Aircraft)).To(Equal(len(aircraftList.Aircraft)))
			Expect(len(unmarshaled.Feeds)).To(Equal(len(aircraftList.Feeds)))
		})
	})

	Describe("Feed struct", func() {
		It("should marshal and unmarshal correctly", func() {
			feed := Feed{
				ID:   1,
				Name: "Test Feed",
			}

			data, err := json.Marshal(feed)
			Expect(err).To(BeNil())

			var unmarshaled Feed
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).To(BeNil())

			Expect(unmarshaled.ID).To(Equal(feed.ID))
			Expect(unmarshaled.Name).To(Equal(feed.Name))
		})
	})
})
