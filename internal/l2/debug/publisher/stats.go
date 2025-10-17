package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/compose-network/localnet-control-plane/internal/logger"
)

// StatsFetcher fetches stats from the shared publisher
type StatsFetcher struct {
	logger *slog.Logger
}

// NewStatsFetcher creates a new stats fetcher
func NewStatsFetcher() *StatsFetcher {
	return &StatsFetcher{
		logger: logger.Named("publisher_stats"),
	}
}

// Stats represents publisher statistics
type Stats struct {
	Active2PCTransactions any    `json:"active_2pc_transactions"`
	CurrentSlot           any    `json:"current_slot"`
	CurrentState          string `json:"current_state"`
	Raw                   string `json:"-"` // Full JSON response
}

// FetchStats fetches stats from the publisher HTTP endpoint
func (f *StatsFetcher) FetchStats(ctx context.Context, url string) (*Stats, error) {
	f.logger.With("url", url).Debug("fetching publisher stats")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var stats Stats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	stats.Raw = string(body)

	f.logger.Debug("stats fetched successfully")

	return &stats, nil
}

// FormatStats formats stats for display
func FormatStats(url string, stats *Stats, err error) string {
	if err != nil {
		return fmt.Sprintf("Shared publisher stats unavailable: %v", err)
	}

	return fmt.Sprintf("Shared publisher:\n  stats URL: %s\n  current_slot=%v active_2pc=%v state=%s",
		url, stats.CurrentSlot, stats.Active2PCTransactions, stats.CurrentState)
}
