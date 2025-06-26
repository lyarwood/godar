package images

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AircraftImageService provides methods to fetch aircraft images
type AircraftImageService struct {
	Client *http.Client
}

// NewAircraftImageService creates a new image service
func NewAircraftImageService() *AircraftImageService {
	return &AircraftImageService{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetAircraftImageURL attempts to find an aircraft image by registration
func (ais *AircraftImageService) GetAircraftImageURL(registration string) string {
	reg := strings.TrimSpace(registration)
	if reg == "" || len(reg) < 3 {
		return ""
	}
	// Try multiple sources in order of preference

	// 1. Try Wikimedia Commons API
	if imageURL := ais.getWikimediaImage(reg); imageURL != "" {
		return imageURL
	}

	// 2. Try OpenSky Network (if you have access)
	// This would require API key and implementation

	// 3. Try generic aircraft type images
	return ""
}

// GetAircraftTypeImageURL attempts to find a generic image for aircraft type
func (ais *AircraftImageService) GetAircraftTypeImageURL(aircraftType string) string {
	typeStr := strings.TrimSpace(aircraftType)
	if typeStr == "" {
		return ""
	}
	// Try Wikimedia Commons for aircraft type images
	return ais.getWikimediaAircraftTypeImage(typeStr)
}

// getWikimediaImage searches Wikimedia Commons for aircraft images
func (ais *AircraftImageService) getWikimediaImage(registration string) string {
	// Wikimedia Commons API endpoint
	baseURL := "https://commons.wikimedia.org/w/api.php"

	// Search for aircraft with registration
	searchQuery := fmt.Sprintf("%s aircraft", registration)

	params := url.Values{}
	params.Set("action", "query")
	params.Set("format", "json")
	params.Set("list", "search")
	params.Set("srsearch", searchQuery)
	params.Set("srnamespace", "6") // File namespace
	params.Set("srlimit", "1")

	resp, err := ais.Client.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Query struct {
			Search []struct {
				Title string `json:"title"`
			} `json:"search"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	if len(result.Query.Search) == 0 {
		return ""
	}

	// Get the actual image URL
	imageTitle := result.Query.Search[0].Title
	return ais.getWikimediaImageURL(imageTitle)
}

// getWikimediaAircraftTypeImage searches for generic aircraft type images
func (ais *AircraftImageService) getWikimediaAircraftTypeImage(aircraftType string) string {
	// Check if this is a military aircraft and use enhanced search
	if ais.IsMilitaryAircraft(aircraftType) {
		if imageURL := ais.getWikimediaMilitaryAircraftImage(aircraftType); imageURL != "" {
			return imageURL
		}
	}

	// Common aircraft type mappings
	typeMappings := map[string]string{
		// Commercial aircraft
		"A320": "Airbus A320",
		"A321": "Airbus A321",
		"A330": "Airbus A330",
		"A350": "Airbus A350",
		"A380": "Airbus A380",
		"B737": "Boeing 737",
		"B747": "Boeing 747",
		"B777": "Boeing 777",
		"B787": "Boeing 787",
		"E190": "Embraer E190",
		"E195": "Embraer E195",

		// Military aircraft
		"F16":   "F-16 Fighting Falcon",
		"F15":   "F-15 Eagle",
		"F18":   "F/A-18 Hornet",
		"F22":   "F-22 Raptor",
		"F35":   "F-35 Lightning II",
		"F14":   "F-14 Tomcat",
		"F4":    "F-4 Phantom II",
		"F5":    "F-5 Tiger II",
		"A10":   "A-10 Thunderbolt II",
		"B1":    "B-1 Lancer",
		"B2":    "B-2 Spirit",
		"B52":   "B-52 Stratofortress",
		"C130":  "C-130 Hercules",
		"C17":   "C-17 Globemaster III",
		"C5":    "C-5 Galaxy",
		"KC135": "KC-135 Stratotanker",
		"KC10":  "KC-10 Extender",
		"E3":    "E-3 Sentry",
		"E4":    "E-4 Nightwatch",
		"P3":    "P-3 Orion",
		"P8":    "P-8 Poseidon",
		"U2":    "U-2 Dragon Lady",
		"SR71":  "SR-71 Blackbird",
		"F117":  "F-117 Nighthawk",
		"B21":   "B-21 Raider",

		// European military aircraft
		"EF2000":  "Eurofighter Typhoon",
		"Rafale":  "Dassault Rafale",
		"Gripen":  "Saab JAS 39 Gripen",
		"Tornado": "Panavia Tornado",
		"Harrier": "Harrier Jump Jet",
		"Hawk":    "BAE Hawk",
		"Typhoon": "Eurofighter Typhoon",

		// Russian military aircraft
		"Su27":  "Sukhoi Su-27",
		"Su30":  "Sukhoi Su-30",
		"Su35":  "Sukhoi Su-35",
		"Su57":  "Sukhoi Su-57",
		"MiG29": "Mikoyan MiG-29",
		"MiG31": "Mikoyan MiG-31",
		"Tu95":  "Tupolev Tu-95",
		"Tu160": "Tupolev Tu-160",
		"Il76":  "Ilyushin Il-76",

		// Helicopters
		"AH64": "AH-64 Apache",
		"UH60": "UH-60 Black Hawk",
		"CH47": "CH-47 Chinook",
		"AH1":  "AH-1 Cobra",
		"V22":  "V-22 Osprey",
		"Ka52": "Kamov Ka-52",
		"Mi24": "Mil Mi-24",
		"Mi28": "Mil Mi-28",
	}

	searchTerm := aircraftType
	if mapped, exists := typeMappings[strings.ToUpper(aircraftType)]; exists {
		searchTerm = mapped
	}

	// Search Wikimedia Commons
	baseURL := "https://commons.wikimedia.org/w/api.php"

	params := url.Values{}
	params.Set("action", "query")
	params.Set("format", "json")
	params.Set("list", "search")
	params.Set("srsearch", fmt.Sprintf("%s aircraft", searchTerm))
	params.Set("srnamespace", "6") // File namespace
	params.Set("srlimit", "1")

	resp, err := ais.Client.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Query struct {
			Search []struct {
				Title string `json:"title"`
			} `json:"search"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	if len(result.Query.Search) == 0 {
		return ""
	}

	// Get the actual image URL
	imageTitle := result.Query.Search[0].Title
	return ais.getWikimediaImageURL(imageTitle)
}

// IsMilitaryAircraft checks if the aircraft type is likely military
func (ais *AircraftImageService) IsMilitaryAircraft(aircraftType string) bool {
	if aircraftType == "" {
		return false
	}
	upperType := strings.ToUpper(strings.TrimSpace(aircraftType))

	// Known military codes (exact matches)
	militaryCodes := map[string]struct{}{
		// US Fighters
		"F16": {}, "F15": {}, "F18": {}, "F22": {}, "F35": {}, "F14": {}, "F4": {}, "F5": {},
		// US Attack
		"A10": {},
		// US Bombers
		"B1": {}, "B2": {}, "B52": {},
		// US Transport
		"C130": {}, "C17": {}, "C5": {},
		// US Tankers
		"KC135": {}, "KC10": {},
		// US Special Mission
		"E3": {}, "E4": {}, "P3": {}, "P8": {}, "U2": {}, "SR71": {}, "F117": {}, "B21": {},
		// European
		"EF2000": {}, "RAFALE": {}, "GRIPEN": {}, "TORNADO": {}, "HARRIER": {}, "HAWK": {}, "TYPHOON": {},
		// Russian
		"SU27": {}, "SU30": {}, "SU35": {}, "SU57": {}, "MIG29": {}, "MIG31": {}, "TU95": {}, "TU160": {}, "IL76": {},
		// Helicopters
		"AH64": {}, "UH60": {}, "CH47": {}, "AH1": {}, "V22": {}, "KA52": {}, "MI24": {}, "MI28": {},
	}
	if _, ok := militaryCodes[upperType]; ok {
		return true
	}

	// Prefix logic for broader detection
	militaryPrefixes := []string{
		"F-", "A-", "B-", "C-", "E-", "P-", "KC-", "U-", "SR-", "F/A-",
		"SU", "MIG", "TU", "IL", "KA", "MI",
		"EF", "RAFALE", "GRIPEN", "TORNADO", "HARRIER", "HAWK", "TYPHOON",
		"AH-", "UH-", "CH-", "V-", "V22",
	}
	if upperType == "V22" || upperType == "V-22" {
		return true
	}
	for _, prefix := range militaryPrefixes {
		if strings.HasPrefix(upperType, prefix) {
			return true
		}
	}
	return false
}

// getWikimediaMilitaryAircraftImage searches for military aircraft images with specific strategies
func (ais *AircraftImageService) getWikimediaMilitaryAircraftImage(aircraftType string) string {
	// Military aircraft type mappings with common variations
	militaryMappings := map[string][]string{
		"F16":   {"F-16 Fighting Falcon", "F-16", "General Dynamics F-16"},
		"F15":   {"F-15 Eagle", "F-15", "McDonnell Douglas F-15"},
		"F18":   {"F/A-18 Hornet", "F-18", "McDonnell Douglas F/A-18"},
		"F22":   {"F-22 Raptor", "F-22", "Lockheed Martin F-22"},
		"F35":   {"F-35 Lightning II", "F-35", "Lockheed Martin F-35"},
		"F14":   {"F-14 Tomcat", "Grumman F-14"},
		"F4":    {"F-4 Phantom II", "McDonnell Douglas F-4"},
		"F5":    {"F-5 Tiger II", "Northrop F-5"},
		"A10":   {"A-10 Thunderbolt II", "A-10", "Fairchild Republic A-10"},
		"B1":    {"B-1 Lancer", "B-1", "Rockwell B-1"},
		"B2":    {"B-2 Spirit", "B-2", "Northrop Grumman B-2"},
		"B52":   {"B-52 Stratofortress", "B-52", "Boeing B-52"},
		"C130":  {"C-130 Hercules", "C-130", "Lockheed C-130"},
		"C17":   {"C-17 Globemaster III", "C-17", "Boeing C-17"},
		"C5":    {"C-5 Galaxy", "C-5", "Lockheed C-5"},
		"KC135": {"KC-135 Stratotanker", "KC-135", "Boeing KC-135"},
		"KC10":  {"KC-10 Extender", "KC-10", "McDonnell Douglas KC-10"},
		"E3":    {"E-3 Sentry", "E-3", "Boeing E-3"},
		"E4":    {"E-4 Nightwatch", "E-4", "Boeing E-4"},
		"P3":    {"P-3 Orion", "P-3", "Lockheed P-3"},
		"P8":    {"P-8 Poseidon", "P-8", "Boeing P-8"},
		"U2":    {"U-2 Dragon Lady", "U-2", "Lockheed U-2"},
		"SR71":  {"SR-71 Blackbird", "SR-71", "Lockheed SR-71"},
		"F117":  {"F-117 Nighthawk", "F-117", "Lockheed F-117"},
		"B21":   {"B-21 Raider", "B-21", "Northrop Grumman B-21"},

		// European military aircraft
		"EF2000":  {"Eurofighter Typhoon", "EF-2000", "Eurofighter EF-2000"},
		"Rafale":  {"Dassault Rafale", "Rafale"},
		"Gripen":  {"Saab JAS 39 Gripen", "JAS 39", "Saab Gripen"},
		"Tornado": {"Panavia Tornado", "Tornado IDS", "Tornado ECR"},
		"Harrier": {"Harrier Jump Jet", "AV-8B Harrier", "Sea Harrier"},
		"Hawk":    {"BAE Hawk", "Hawk T1", "Hawk T2"},
		"Typhoon": {"Eurofighter Typhoon", "EF-2000 Typhoon"},

		// Russian military aircraft
		"Su27":  {"Sukhoi Su-27", "Su-27", "Sukhoi Su-27 Flanker"},
		"Su30":  {"Sukhoi Su-30", "Su-30", "Sukhoi Su-30 Flanker-C"},
		"Su35":  {"Sukhoi Su-35", "Su-35", "Sukhoi Su-35 Flanker-E"},
		"Su57":  {"Sukhoi Su-57", "Su-57", "Sukhoi Su-57 Felon"},
		"MiG29": {"Mikoyan MiG-29", "MiG-29", "Mikoyan MiG-29 Fulcrum"},
		"MiG31": {"Mikoyan MiG-31", "MiG-31", "Mikoyan MiG-31 Foxhound"},
		"Tu95":  {"Tupolev Tu-95", "Tu-95", "Tupolev Tu-95 Bear"},
		"Tu160": {"Tupolev Tu-160", "Tu-160", "Tupolev Tu-160 Blackjack"},
		"Il76":  {"Ilyushin Il-76", "Il-76", "Ilyushin Il-76 Candid"},

		// Helicopters
		"AH64": {"AH-64 Apache", "AH-64", "Boeing AH-64 Apache"},
		"UH60": {"UH-60 Black Hawk", "UH-60", "Sikorsky UH-60"},
		"CH47": {"CH-47 Chinook", "CH-47", "Boeing CH-47 Chinook"},
		"AH1":  {"AH-1 Cobra", "AH-1", "Bell AH-1 Cobra"},
		"V22":  {"V-22 Osprey", "V-22", "Bell Boeing V-22 Osprey"},
		"Ka52": {"Kamov Ka-52", "Ka-52", "Kamov Ka-52 Alligator"},
		"Mi24": {"Mil Mi-24", "Mi-24", "Mil Mi-24 Hind"},
		"Mi28": {"Mil Mi-28", "Mi-28", "Mil Mi-28 Havoc"},
	}

	upperType := strings.ToUpper(aircraftType)

	// Try exact match first
	if searchTerms, exists := militaryMappings[upperType]; exists {
		for _, searchTerm := range searchTerms {
			if imageURL := ais.searchWikimediaCommons(searchTerm); imageURL != "" {
				return imageURL
			}
		}
	}

	// Try partial matches for military aircraft
	for key, searchTerms := range militaryMappings {
		if strings.Contains(upperType, key) || strings.Contains(key, upperType) {
			for _, searchTerm := range searchTerms {
				if imageURL := ais.searchWikimediaCommons(searchTerm); imageURL != "" {
					return imageURL
				}
			}
		}
	}

	return ""
}

// searchWikimediaCommons is a helper function to search Wikimedia Commons
func (ais *AircraftImageService) searchWikimediaCommons(searchTerm string) string {
	baseURL := "https://commons.wikimedia.org/w/api.php"

	params := url.Values{}
	params.Set("action", "query")
	params.Set("format", "json")
	params.Set("list", "search")
	params.Set("srsearch", fmt.Sprintf("%s aircraft", searchTerm))
	params.Set("srnamespace", "6") // File namespace
	params.Set("srlimit", "1")

	resp, err := ais.Client.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Query struct {
			Search []struct {
				Title string `json:"title"`
			} `json:"search"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	if len(result.Query.Search) == 0 {
		return ""
	}

	// Get the actual image URL
	imageTitle := result.Query.Search[0].Title
	return ais.getWikimediaImageURL(imageTitle)
}

// getWikimediaImageURL gets the actual image URL from Wikimedia Commons
func (ais *AircraftImageService) getWikimediaImageURL(imageTitle string) string {
	baseURL := "https://commons.wikimedia.org/w/api.php"

	params := url.Values{}
	params.Set("action", "query")
	params.Set("format", "json")
	params.Set("titles", imageTitle)
	params.Set("prop", "imageinfo")
	params.Set("iiprop", "url")

	resp, err := ais.Client.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Query struct {
			Pages map[string]struct {
				ImageInfo []struct {
					URL string `json:"url"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	// Get the first page (there should only be one)
	for _, page := range result.Query.Pages {
		if len(page.ImageInfo) > 0 {
			return page.ImageInfo[0].URL
		}
	}

	return ""
}
