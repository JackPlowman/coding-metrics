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
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
	Path []interface{} `json:"path,omitempty"`
}

// GitHubGraphQLClient provides a client for making GraphQL requests to GitHub
type GitHubGraphQLClient struct {
	Token    string
	Endpoint string
	Client   *http.Client
}

// NewGitHubGraphQLClient creates a new GitHub GraphQL client
func NewGitHubGraphQLClient(token string) *GitHubGraphQLClient {
	return &GitHubGraphQLClient{
		Token:    token,
		Endpoint: DefaultGitHubGraphQLEndpoint,
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

	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		zap.L().Fatal("Failed to query GitHub GraphQL API", zap.Error(err))
	}
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
		for _, err := range graphqlResp.Errors {
			errorMessages += err.Message + "; "
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
