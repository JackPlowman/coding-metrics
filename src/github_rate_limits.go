package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

// checkGitHubRateLimits checks both APIs used to fetch metrics before generation.
// This is a startup snapshot; later requests can still encounter rate limits.
func checkGitHubRateLimits() error {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/rate_limit", nil)
	if err != nil {
		return fmt.Errorf("failed to create GitHub rate limit request: %w", err)
	}
	req.Header.Set("Authorization", bearerPrefix+os.Getenv("INPUT_GITHUB_TOKEN"))
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check GitHub rate limits: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			zap.L().Warn("Failed to close GitHub rate limit response body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub rate limit check returned status %d", resp.StatusCode)
	}

	var limits struct {
		Resources map[string]struct {
			Remaining *int  `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"resources"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&limits); err != nil {
		return fmt.Errorf("failed to decode GitHub rate limits: %w", err)
	}

	for _, resource := range []string{"core", "graphql"} {
		limit, ok := limits.Resources[resource]
		if !ok || limit.Remaining == nil {
			return fmt.Errorf("GitHub rate limit response is missing %s remaining quota", resource)
		}
		if *limit.Remaining <= 0 {
			if limit.Reset > 0 {
				return fmt.Errorf("GitHub %s rate limit exhausted; retry after %s",
					resource, time.Unix(limit.Reset, 0).UTC().Format(time.RFC3339))
			}
			return fmt.Errorf(
				"GitHub %s rate limit exhausted; retry when the quota resets",
				resource,
			)
		}
		zap.L().Debug("GitHub rate limit available",
			zap.String("resource", resource), zap.Int("remaining", *limit.Remaining))
	}
	return nil
}
