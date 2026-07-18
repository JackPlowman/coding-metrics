package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

const (
	// DefaultGitHubGraphQLEndpoint is the default GitHub GraphQL API endpoint
	DefaultGitHubGraphQLEndpoint = "https://api.github.com/graphql"
)

// GitHubGraphQLRequest represents a GitHub GraphQL API request
type GitHubGraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GitHubGraphQLResponse represents a GitHub GraphQL API response
type GitHubGraphQLResponse struct {
	Data   interface{}          `json:"data"`
	Errors []GitHubGraphQLError `json:"errors,omitempty"`
}

// GitHubGraphQLError represents an error in a GraphQL response
type GitHubGraphQLError struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
	Path []interface{} `json:"path,omitempty"`
}

// GitHubGraphQLClient provides a client for making GraphQL requests to GitHub
type GitHubGraphQLClient struct {
	Token  string
	Client *http.Client
}

// NewGitHubGraphQLClient creates a new GitHub GraphQL client
func NewGitHubGraphQLClient(token string) *GitHubGraphQLClient {
	return &GitHubGraphQLClient{
		Token: token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Query executes a GraphQL query against the GitHub API
func (c *GitHubGraphQLClient) Query(
	query string,
	variables map[string]interface{},
	result interface{},
) error {
	requestBody := GitHubGraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	reqBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", DefaultGitHubGraphQLEndpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	startedAt := time.Now()
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf(
			"failed to query GitHub GraphQL API after %s: %w",
			time.Since(startedAt),
			err,
		)
	}
	zap.L().Debug(
		"GitHub GraphQL response",
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", time.Since(startedAt)),
		zap.String("rate_limit_limit", resp.Header.Get("X-RateLimit-Limit")),
		zap.String("rate_limit_remaining", resp.Header.Get("X-RateLimit-Remaining")),
		zap.String("rate_limit_used", resp.Header.Get("X-RateLimit-Used")),
		zap.String("rate_limit_reset", resp.Header.Get("X-RateLimit-Reset")),
		zap.String("rate_limit_resource", resp.Header.Get("X-RateLimit-Resource")),
		zap.String("retry_after", resp.Header.Get("Retry-After")),
		zap.String("request_id", resp.Header.Get("X-GitHub-Request-Id")),
	)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			zap.L().Fatal("Failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"GitHub API returned non-200 status code %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var graphqlResp GitHubGraphQLResponse
	graphqlResp.Data = result

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&graphqlResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		errorMessages := ""
		for _, graphqlErr := range graphqlResp.Errors {
			zap.L().Warn(
				"GitHub GraphQL error",
				zap.String("type", graphqlErr.Type),
				zap.Any("path", graphqlErr.Path),
				zap.String("message", graphqlErr.Message),
			)
			errorMessages += fmt.Sprintf(
				"%s at %v: %s; ",
				graphqlErr.Type,
				graphqlErr.Path,
				graphqlErr.Message,
			)
		}
		return fmt.Errorf("GraphQL errors: %s", errorMessages)
	}

	return nil
}

// QueryGitHubQLAPI is a convenience function for making GitHub GraphQL queries
func QueryGitHubQLAPI(query string, variables map[string]interface{}, result interface{}) error {
	token := os.Getenv("INPUT_GITHUB_TOKEN")
	client := NewGitHubGraphQLClient(token)
	return client.Query(query, variables, result)
}
