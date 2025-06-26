package fetch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client represents an HTTP client with configuration
type Client struct {
	httpClient *http.Client
	userAgent  string
	logger     *zap.Logger
}

// ClientConfig holds configuration for the HTTP client
type ClientConfig struct {
	Timeout         time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	UserAgent       string
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

// NewClient creates a new HTTP client with the given configuration
func NewClient(config ClientConfig, logger *zap.Logger) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.UserAgent == "" {
		config.UserAgent = "Godar/1.0"
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 10
	}
	if config.IdleConnTimeout == 0 {
		config.IdleConnTimeout = 90 * time.Second
	}

	transport := &http.Transport{
		MaxIdleConns:       config.MaxIdleConns,
		IdleConnTimeout:    config.IdleConnTimeout,
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	httpClient := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	return &Client{
		httpClient: httpClient,
		userAgent:  config.UserAgent,
		logger:     logger,
	}
}

// Get performs an HTTP GET request with retry logic
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= 3; attempt++ {
		if attempt > 0 {
			c.logger.Debug("Retrying request",
				zap.String("url", url),
				zap.Int("attempt", attempt+1),
				zap.Error(lastErr))
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		resp.Body.Close()

		// Don't retry on client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", 3, lastErr)
}
