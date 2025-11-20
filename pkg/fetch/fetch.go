package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/lyarwood/godar/pkg/aircraft"

	"go.uber.org/zap"
)

// Fetcher handles fetching aircraft data from a Virtual Radar Server
type Fetcher struct {
	BaseURL       string
	AircraftType  string
	MinAltitude   int
	MaxAltitude   int
	Military      bool
	Operator      string
	FlightNumber  string
	UserLat       float64
	UserLong      float64
	MaxDistance   float64
	Username      string // Login username
	Password      string // Login password
	Logger        *zap.Logger
	client        *http.Client
	sessionCookie string
	authMethod    string // Track which auth method worked: "session", "basic", "url", or ""
}

// NewFetcher creates a new Fetcher with the given parameters and logger
func NewFetcher(baseURL string, logger *zap.Logger) *Fetcher {
	return &Fetcher{
		BaseURL: baseURL,
		Logger:  logger,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects - we need to capture cookies from redirect responses
				return http.ErrUseLastResponse
			},
		},
	}
}

// SetFilters sets the filtering parameters for the fetcher
func (f *Fetcher) SetFilters(aircraftType string, minAltitude, maxAltitude int, military bool, operator, flightNumber string) {
	f.AircraftType = aircraftType
	f.MinAltitude = minAltitude
	f.MaxAltitude = maxAltitude
	f.Military = military
	f.Operator = operator
	f.FlightNumber = flightNumber
}

// SetLocation sets the location-based filtering parameters
func (f *Fetcher) SetLocation(lat, lng, maxDistance float64) {
	f.UserLat = lat
	f.UserLong = lng
	f.MaxDistance = maxDistance
}

// SetAuth sets the login credentials
func (f *Fetcher) SetAuth(username, password string) {
	f.Username = username
	f.Password = password
}

// login performs authentication - trying multiple methods
func (f *Fetcher) login() error {
	if f.Username == "" || f.Password == "" {
		return fmt.Errorf("username and password required for login")
	}

	// Try different authentication methods
	authMethods := []struct {
		name string
		fn   func() error
	}{
		{"Form-based login", f.tryFormLogin},
		{"HTTP Basic Auth", f.tryBasicAuth},
		{"URL-based auth", f.tryURLAuth},
		{"No authentication", f.tryNoAuth},
	}

	for _, method := range authMethods {
		f.Logger.Debug("Trying authentication method", zap.String("method", method.name))
		if err := method.fn(); err == nil {
			f.Logger.Debug("Authentication successful", zap.String("method", method.name))
			return nil
		} else {
			f.Logger.Debug("Authentication failed", zap.String("method", method.name), zap.Error(err))
		}
	}

	return fmt.Errorf("all authentication methods failed")
}

// tryFormLogin attempts form-based login
func (f *Fetcher) tryFormLogin() error {
	baseURL, err := url.Parse(f.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	loginURL := fmt.Sprintf("https://%s/login.php", baseURL.Host)

	formData := url.Values{}
	formData.Set("username", f.Username)
	formData.Set("password", f.Password)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", baseURL.Host))

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	f.Logger.Debug("Form login response",
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.Strings("cookies", getCookieNames(resp.Cookies())),
		zap.String("bodySnippet", string(body[:min(1000, len(body))])))

	// Look specifically for the 'rauth' cookie
	for _, cookie := range resp.Cookies() {
		if strings.EqualFold(cookie.Name, "rauth") {
			f.sessionCookie = cookie.Name + "=" + cookie.Value
			f.authMethod = "session"
			f.Logger.Debug("Captured rauth cookie", zap.String("cookie", f.sessionCookie))
			return nil
		}
	}

	return fmt.Errorf("no rauth cookie found in login response")
}

// tryBasicAuth attempts HTTP Basic Auth
func (f *Fetcher) tryBasicAuth() error {
	// Test if the server accepts HTTP Basic Auth by making a test request
	testURL := f.BaseURL + "?test=1"

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.SetBasicAuth(f.Username, f.Password)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", req.URL.Host))

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("test request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and log the response body
	body, _ := io.ReadAll(resp.Body)
	f.Logger.Debug("Basic auth response",
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.String("contentType", resp.Header.Get("Content-Type")),
		zap.String("bodySnippet", string(body[:min(1000, len(body))])))

	// If we get JSON back, Basic Auth worked
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		f.Logger.Debug("HTTP Basic Auth appears to work")
		// Store credentials for future requests
		f.authMethod = "basic"
		return nil
	}

	return fmt.Errorf("HTTP Basic Auth not accepted")
}

// tryURLAuth attempts URL-based authentication
func (f *Fetcher) tryURLAuth() error {
	// Try adding credentials as URL parameters
	u, err := url.Parse(f.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("username", f.Username)
	q.Set("password", f.Password)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create URL auth request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", req.URL.Host))

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("URL auth request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and log the response body
	body, _ := io.ReadAll(resp.Body)
	f.Logger.Debug("URL auth response",
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.String("contentType", resp.Header.Get("Content-Type")),
		zap.String("bodySnippet", string(body[:min(1000, len(body))])))

	// If we get JSON back, URL auth worked
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		f.Logger.Debug("URL-based auth appears to work")
		f.authMethod = "url"
		return nil
	}

	return fmt.Errorf("URL-based auth not accepted")
}

// tryNoAuth attempts to access the data without authentication
func (f *Fetcher) tryNoAuth() error {
	// Try accessing the aircraft data directly without any authentication
	req, err := http.NewRequest("GET", f.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create no-auth request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", req.URL.Host))

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("no-auth request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and log the response body
	body, _ := io.ReadAll(resp.Body)
	f.Logger.Debug("No-auth response",
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.String("contentType", resp.Header.Get("Content-Type")),
		zap.String("bodySnippet", string(body[:min(1000, len(body))])))

	// If we get JSON back, no authentication is needed
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		f.Logger.Debug("No authentication required")
		f.authMethod = "none"
		return nil
	}

	return fmt.Errorf("authentication required")
}

// Fetch fetches aircraft data from the VRS server with applied filters.
func (f *Fetcher) Fetch() (*aircraft.AircraftList, error) {
	if f.Logger == nil {
		f.Logger = zap.NewNop()
	}

	// Login if we have credentials and no session
	if (f.Username != "" || f.Password != "") && f.authMethod == "" {
		if err := f.login(); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	u, err := url.Parse(f.BaseURL)
	if err != nil {
		f.Logger.Error("Failed to parse base URL",
			zap.String("baseURL", f.BaseURL),
			zap.Error(err))
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	q := u.Query()

	// Apply filters
	filtersApplied := 0
	if f.AircraftType != "" {
		q.Set("fTypQ", f.AircraftType)
		filtersApplied++
	}
	if f.MinAltitude > 0 {
		q.Set("fAltL", strconv.Itoa(f.MinAltitude))
		filtersApplied++
	}
	if f.MaxAltitude > 0 {
		q.Set("fAltU", strconv.Itoa(f.MaxAltitude))
		filtersApplied++
	}
	if f.Military {
		q.Set("fMilQ", "1")
		filtersApplied++
	}
	if f.Operator != "" {
		q.Set("fOpQ", f.Operator)
		filtersApplied++
	}
	if f.FlightNumber != "" {
		q.Set("fCallQ", f.FlightNumber)
		filtersApplied++
	}

	if f.UserLat != 0.0 && f.UserLong != 0.0 {
		q.Set("lat", strconv.FormatFloat(f.UserLat, 'f', -1, 64))
		q.Set("lng", strconv.FormatFloat(f.UserLong, 'f', -1, 64))
		if f.MaxDistance > 0 {
			q.Set("fDstU", strconv.FormatFloat(f.MaxDistance, 'f', -1, 64))
		}
		filtersApplied++
	}

	// Apply URL-based auth if that method worked
	if f.authMethod == "url" {
		q.Set("username", f.Username)
		q.Set("password", f.Password)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		f.Logger.Error("Failed to create HTTP request",
			zap.String("url", u.String()),
			zap.Error(err))
		return nil, err
	}

	// Apply authentication based on method
	switch f.authMethod {
	case "session":
		if f.sessionCookie != "" {
			req.Header.Set("Cookie", f.sessionCookie)
		}
	case "basic":
		req.SetBasicAuth(f.Username, f.Password)
	case "url":
		// URL auth is already applied in the query parameters above
	case "none":
		// No authentication needed
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", req.URL.Host))

	f.Logger.Debug("Sending HTTP request",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.Bool("hasSession", f.sessionCookie != ""),
		zap.Strings("headers", getHeaderNames(req.Header)))

	resp, err := f.client.Do(req)
	if err != nil {
		f.Logger.Error("HTTP request failed",
			zap.String("url", u.String()),
			zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	f.Logger.Debug("Received HTTP response",
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.String("contentType", resp.Header.Get("Content-Type")),
		zap.Int64("contentLength", resp.ContentLength),
		zap.Strings("responseHeaders", getHeaderNames(resp.Header)))

	// If we get a redirect to login, try to login again
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		location := resp.Header.Get("Location")
		if strings.Contains(location, "login.php") {
			f.Logger.Info("Session expired, attempting to login again")
			f.sessionCookie = "" // Clear expired session
			f.authMethod = ""    // Clear auth method to force re-authentication
			if err := f.login(); err != nil {
				return nil, fmt.Errorf("re-authentication failed: %w", err)
			}
			// Retry the request with new session
			return f.Fetch()
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		f.Logger.Error("HTTP request returned non-OK status",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.String("url", u.String()),
			zap.String("bodySnippet", string(body[:min(500, len(body))])))
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		f.Logger.Error("Failed to read response body",
			zap.String("url", u.String()),
			zap.Error(err))
		return nil, err
	}

	f.Logger.Debug("Read response body",
		zap.Int("bodySize", len(body)),
		zap.String("url", u.String()))

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || (contentType != "application/json" && !startsWith(contentType, "application/json")) {
		f.Logger.Error("Response is not JSON",
			zap.String("contentType", contentType),
			zap.String("url", u.String()),
			zap.String("bodySnippet", string(body[:min(500, len(body))])))
		return nil, fmt.Errorf("expected JSON response, got content-type: %s", contentType)
	}

	var acList aircraft.AircraftList
	err = json.Unmarshal(body, &acList)
	if err != nil {
		f.Logger.Error("Failed to unmarshal JSON response",
			zap.String("url", u.String()),
			zap.Int("bodySize", len(body)),
			zap.String("bodySnippet", string(body[:min(500, len(body))])),
			zap.Error(err))
		return nil, err
	}

	f.Logger.Info("Successfully fetched aircraft data",
		zap.String("url", u.String()),
		zap.Int("totalAircraft", acList.TotalAc),
		zap.Int("aircraftCount", len(acList.Aircraft)),
		zap.Int64("timestamp", acList.Stm),
		zap.Int("filtersApplied", filtersApplied))

	return &acList, nil
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// startsWith checks if s starts with prefix (case-insensitive)
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// getHeaderNames returns a slice of header names for logging
func getHeaderNames(headers http.Header) []string {
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	return names
}

// getCookieNames returns a slice of cookie names for logging
func getCookieNames(cookies []*http.Cookie) []string {
	names := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		names = append(names, cookie.Name)
	}
	return names
}
