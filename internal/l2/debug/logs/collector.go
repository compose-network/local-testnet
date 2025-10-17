package logs

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/compose-network/localnet-control-plane/internal/logger"
)

// Collector collects Docker logs
type Collector struct {
	logger *slog.Logger
}

// NewCollector creates a new logs collector
func NewCollector() *Collector {
	return &Collector{
		logger: logger.Named("logs_collector"),
	}
}

// LogKeywords are keywords to filter for in logs
var LogKeywords = []string{
	"mailbox",
	"SSV",
	"Send CIRC",
	"putInbox",
	"Tracer captured",
}

// CollectLogs collects filtered Docker logs from a container
func (c *Collector) CollectLogs(ctx context.Context, containerName string, since string) ([]string, error) {
	c.logger.With("container", containerName).With("since", since).Debug("collecting logs")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "logs", containerName, fmt.Sprintf("--since=%s", since))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get docker logs: %w (output: %s)", err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	filtered := c.filterLines(lines)

	if len(filtered) > 20 {
		filtered = filtered[len(filtered)-20:]
	}

	c.logger.With("total_lines", len(lines)).With("filtered_lines", len(filtered)).Debug("logs collected")

	return filtered, nil
}

// filterLines filters log lines for keywords
func (c *Collector) filterLines(lines []string) []string {
	var filtered []string

	// Build regex pattern from keywords
	pattern := strings.Join(LogKeywords, "|")
	re, err := regexp.Compile("(?i)" + pattern) // Case-insensitive
	if err != nil {
		c.logger.With("err", err).Warn("failed to compile regex, returning unfiltered")
		return lines
	}

	for _, line := range lines {
		if re.MatchString(line) {
			filtered = append(filtered, line)
		}
	}

	return filtered
}

// FormatLogs formats log lines for display
func FormatLogs(containerName string, lines []string) string {
	if len(lines) == 0 {
		return fmt.Sprintf("[%s]\n  (no matching log lines in window)", containerName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s]\n", containerName))
	for _, line := range lines {
		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}
