package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const bearerPrefix = "Bearer "

// getPullRequestTotal fetches the total number of pull requests for a given user
func getPullRequestTotal(username string) {
	url := fmt.Sprintf("https://api.github.com/search/issues?q=author:%s+type:pr", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create request for pull request total: %v\n", err)
		return
	}
	req.Header.Set("Authorization", bearerPrefix+os.Getenv("INPUT_GITHUB_TOKEN"))

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

// getIssuesTotal fetches the total number of issues for a given user
func getIssuesTotal(username string) {
	url := fmt.Sprintf("https://api.github.com/search/issues?q=author:%s+type:issue", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create request for issues total: %v\n", err)
		return
	}
	req.Header.Set("Authorization", bearerPrefix+os.Getenv("INPUT_GITHUB_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to query GitHub API for issues total: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GitHub API returned status %d for issues total\n", resp.StatusCode)
		return
	}

	var result struct {
		TotalCount int `json:"total_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Failed to decode issues total response: %v\n", err)
		return
	}

	fmt.Printf("Total issues by %s: %d\n", username, result.TotalCount)
}

// getGitHubUserInfo fetches the user's avatar URL, login (tag), and display name from GitHub REST API
func getGitHubUserInfo() (avatarURL, login, name string, err error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", bearerPrefix+os.Getenv("INPUT_GITHUB_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var user struct {
		AvatarURL string `json:"avatar_url"`
		Login     string `json:"login"`
		Name      string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	return user.AvatarURL, user.Login, user.Name, nil
}
