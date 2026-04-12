package gtfs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BARTTripUpdatesURL = "https://api.bart.gov/gtfsrt/tripupdate.aspx"

func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func FetchTripUpdates(ctx context.Context, client *http.Client, feedURL string) ([]byte, error) {
	if client == nil {
		client = NewHTTPClient(10 * time.Second)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch trip updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("empty response body")
	}

	return body, nil
}
