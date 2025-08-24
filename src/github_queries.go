package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

const bearerPrefix = "Bearer "

// getGitHubUserInfo fetches the user's information from GitHub REST API
type GitHubUserInfo struct {
	AvatarURL    string    `json:"avatar_url"`
	Followers    int       `json:"followers"`
	Id           int       `json:"id"`
	JoinedGitHub time.Time `json:"created_at"`
	Login        string    `json:"login"`
	Name         string    `json:"name"`
	PublicGists  int       `json:"public_gists"`
	PublicRepos  int       `json:"public_repos"`
	Type         string    `json:"type"`
}

func getGitHubUserInfo() *GitHubUserInfo {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		zap.L().Fatal("Failed to create request for GitHub user info", zap.Error(err))
	}
	req.Header.Set("Authorization", bearerPrefix+os.Getenv("INPUT_GITHUB_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Fatal("Failed to make request for GitHub user info", zap.Error(err))
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			zap.L().Fatal("Failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		zap.L().Error("GitHub API returned non-200 status", zap.Int("status", resp.StatusCode))
	}

	var user GitHubUserInfo

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		zap.L().Fatal("Failed to decode GitHub user info", zap.Error(err))
	}
	return &user
}

// getCommitsTotal fetches the total number of commits made by a user to default branches across all repositories
func getCommitsTotal(userName string, userId int) int {
	zap.L().Debug("Fetching total commits", zap.String("username", userName), zap.Int("userId", userId))
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
		"login":  userName,
		"userId": userId,
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
			zap.L().Fatal("Failed to get commits total", zap.Error(err))
		}

		for _, repo := range result.User.Repositories.Nodes {
			if repo.DefaultBranchRef != nil {
				totalCommits += repo.DefaultBranchRef.Target.History.TotalCount
			}
		}

		hasNextPage = result.User.Repositories.PageInfo.HasNextPage
		cursor = result.User.Repositories.PageInfo.EndCursor
	}

	zap.L().
		Debug("Total commits by user", zap.String("username", userName), zap.Int("total_commits", totalCommits))
	return totalCommits
}

type GitHubTotals struct {
	TotalPullRequests  int
	TotalIssues        int
	TotalIssueComments int
}

func getGitHubTotals(userName string, userId int) *GitHubTotals {
	zap.L().Debug("Fetching GitHub totals", zap.String("username", userName), zap.Int("userId", userId))
	query := `
	query($login: String!, $userId: ID!) {
		user(login: $login) {
			issues {
				totalCount
			}
			pullRequests {
				totalCount
			}
			issueComments {
				totalCount
			}
		}
	}`

	variables := map[string]interface{}{
		"login":  userName,
		"userId": userId,
	}

	var result struct {
		User struct {
			Issues struct {
				TotalCount int `json:"totalCount"`
			} `json:"issues"`
			PullRequests struct {
				TotalCount int `json:"totalCount"`
			} `json:"pullRequests"`
			IssueComments struct {
				TotalCount int `json:"totalCount"`
			} `json:"issueComments"`
		} `json:"user"`
	}

	if err := QueryGitHubQLAPI(query, variables, &result); err != nil {
		zap.L().Error("Failed to get GitHub totals", zap.Error(err))
	}

	return &GitHubTotals{
		TotalPullRequests:  result.User.PullRequests.TotalCount,
		TotalIssues:        result.User.Issues.TotalCount,
		TotalIssueComments: result.User.IssueComments.TotalCount,
	}
}

type ActivityStats struct {
	TotalCommits       int
	TotalIssues        int
	TotalPullRequests  int
	TotalIssueComments int
}

func getActivityStats(userName string, userId int) *ActivityStats {
	totals := getGitHubTotals(userName, userId)
	totalCommits := getCommitsTotal(userName, userId)

	return &ActivityStats{
		TotalCommits:       totalCommits,
		TotalPullRequests:  totals.TotalPullRequests,
		TotalIssues:        totals.TotalIssues,
		TotalIssueComments: totals.TotalIssueComments,
	}
}
