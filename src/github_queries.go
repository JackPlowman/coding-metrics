package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func getPullRequestTotal() {
	username := "JackPlowman" // TODO: get username from GITHUB_TOKEN
	url := fmt.Sprintf("https://api.github.com/search/issues?q=author:%s+type:pr", username)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create request for pull request total: %v\n", err)
		return
	}

	// Add GitHub token if available for higher rate limits
	if token := os.Getenv("INPUT_GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to query GitHub API for pull request total: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GitHub API returned status %d for pull request total\n", resp.StatusCode)
		return
	}

	var result struct {
		TotalCount int `json:"total_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Failed to decode pull request total response: %v\n", err)
		return
	}

	fmt.Printf("Total pull requests by %s: %d\n", username, result.TotalCount)
}

// getUserAvatarURL fetches the user's avatar URL from GitHub REST API
func getUserAvatarURL(username string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add GitHub token if available for higher rate limits
	if token := os.Getenv("INPUT_GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var user struct {
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return user.AvatarURL, nil
}

