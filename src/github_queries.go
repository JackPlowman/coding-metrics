package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

const bearerPrefix = "Bearer "

// getPullRequestTotal fetches the total number of pull requests for a given user
func getPullRequestTotal(username string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/search/issues?q=author:%s+type:pr", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request for pull request total: %w", err)
	}
	req.Header.Set("Authorization", bearerPrefix+os.Getenv("INPUT_GITHUB_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to query GitHub API for pull request total: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			zap.L().Fatal("failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(
			"github API returned status %d for pull request total",
			resp.StatusCode,
		)
	}

	var result struct {
		TotalCount int `json:"total_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode pull request total response: %w", err)
	}

	fmt.Printf("Total pull requests by %s: %d\n", username, result.TotalCount)
	return result.TotalCount, nil
}

// getIssuesTotal fetches the total number of issues for a given user
func getIssuesTotal(username string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/search/issues?q=author:%s+type:issue", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create request for issues total: %v\n", err)
		return 0, err
	}
	req.Header.Set("Authorization", bearerPrefix+os.Getenv("INPUT_GITHUB_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to query GitHub API for issues total: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			zap.L().Fatal("failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github API returned status %d for issues total", resp.StatusCode)
	}

	var result struct {
		TotalCount int `json:"total_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode issues total response: %w", err)
	}

	fmt.Printf("Total issues by %s: %d\n", username, result.TotalCount)
	return result.TotalCount, nil
}

// getGitHubUserInfo fetches the user's information from GitHub REST API
type GitHubUserInfo struct {
	AvatarURL    string    `json:"avatar_url"`
	Followers    int       `json:"followers"`
	JoinedGitHub time.Time `json:"created_at"`
	Login        string    `json:"login"`
	Name         string    `json:"name"`
	PublicGists  int       `json:"public_gists"`
	PublicRepos  int       `json:"public_repos"`
	Type         string    `json:"type"`
}

func getGitHubUserInfo() (*GitHubUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", bearerPrefix+os.Getenv("INPUT_GITHUB_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			zap.L().Fatal("Failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var user GitHubUserInfo

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &user, nil
}

// getCommitsTotal fetches the total number of commits made by a user to default branches across all repositories
func getCommitsTotal(username string) (int, error) {
	// First, get the user's ID
	userQuery := `
	query($login: String!) {
		user(login: $login) {
			id
		}
	}`

	var userResult struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}

	userVariables := map[string]interface{}{
		"login": username,
	}

	if err := QueryGitHubQLAPI(userQuery, userVariables, &userResult); err != nil {
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}

	userID := userResult.User.ID

	// Now query repositories and commits using the user ID
	query := `
	query($login: String!, $userId: ID!, $after: String) {
		user(login: $login) {
			repositories(first: 100, after: $after, ownerAffiliations: [OWNER, ORGANIZATION_MEMBER, COLLABORATOR]) {
				pageInfo {
					hasNextPage
					endCursor
				}
				nodes {
					name
					defaultBranchRef {
						target {
							... on Commit {
								history(author: {id: $userId}) {
									totalCount
								}
							}
						}
					}
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"login":  username,
		"userId": userID,
	}

	totalCommits := 0
	hasNextPage := true
	cursor := ""

	for hasNextPage {
		if cursor != "" {
			variables["after"] = cursor
		}

		var result struct {
			User struct {
				Repositories struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []struct {
						Name             string `json:"name"`
						DefaultBranchRef *struct {
							Target struct {
								History struct {
									TotalCount int `json:"totalCount"`
								} `json:"history"`
							} `json:"target"`
						} `json:"defaultBranchRef"`
					} `json:"nodes"`
				} `json:"repositories"`
			} `json:"user"`
		}

		if err := QueryGitHubQLAPI(query, variables, &result); err != nil {
			return 0, fmt.Errorf("failed to query commits: %w", err)
		}

		for _, repo := range result.User.Repositories.Nodes {
			if repo.DefaultBranchRef != nil {
				totalCommits += repo.DefaultBranchRef.Target.History.TotalCount
			}
		}

		hasNextPage = result.User.Repositories.PageInfo.HasNextPage
		cursor = result.User.Repositories.PageInfo.EndCursor
	}

	fmt.Printf("Total commits by %s: %d\n", username, totalCommits)
	return totalCommits, nil
}

type ActivityStats struct {
	TotalCommits      int
	TotalIssues       int
	TotalPullRequests int
}

func getActivityStats(username string) (*ActivityStats, error) {
	totalCommits, err := getCommitsTotal(username)
	if err != nil {
		zap.L().Fatal("Failed to get total commits", zap.Error(err))
	}

	totalIssues, err := getIssuesTotal(username)
	if err != nil {
		zap.L().Fatal("Failed to get total issues", zap.Error(err))
	}

	totalPullRequests, err := getPullRequestTotal(username)
	if err != nil {
		zap.L().Fatal("Failed to get total PRs", zap.Error(err))
	}

	return &ActivityStats{
		TotalCommits:      totalCommits,
		TotalIssues:       totalIssues,
		TotalPullRequests: totalPullRequests,
	}, nil
}
