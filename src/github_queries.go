package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
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
	Following    int       `json:"following"`
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

// fetchAvatarAsBase64 fetches an avatar image from a URL and returns it as a base64 data URI
func fetchAvatarAsBase64(avatarURL string) string {
	resp, err := http.Get(avatarURL) // #nosec G107 -- URL is from GitHub API
	if err != nil {
		zap.L().Error("Failed to fetch avatar image", zap.Error(err))
		return ""
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			zap.L().Error("Failed to close avatar response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		zap.L().Error("Failed to fetch avatar image", zap.Int("status", resp.StatusCode))
		return ""
	}

	// Read the image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("Failed to read avatar image data", zap.Error(err))
		return ""
	}

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Determine content type (GitHub avatars are typically PNG)
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}

	// Return as data URI
	return "data:" + contentType + ";base64," + base64Image
}

// GetUserId fetches the user ID for a given username
func getUserId(userName string) string {
	zap.L().Debug("Fetching user ID", zap.String("username", userName))
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
		"login": userName,
	}

	if err := QueryGitHubQLAPI(userQuery, userVariables, &userResult); err != nil {
		zap.L().Fatal("Failed to query user ID", zap.Error(err))
	}

	return userResult.User.ID
}

// getCommitsTotal fetches the total number of commits made by a user to default branches across all repositories
func getCommitsTotal(userName, userId string) int {
	zap.L().
		Debug("Fetching total commits")
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
		Debug("Total commits by user", zap.Int("total_commits", totalCommits))
	return totalCommits
}

type GitHubTotals struct {
	TotalPullRequests          int
	TotalIssues                int
	TotalPullRequestReviews    int
	TotalStarredRepos          int
	TotalSponsors              int
	TotalMemberOfOrganizations int
	TotalWatchers              int
}

func getGitHubTotals(userName, userId string) *GitHubTotals {
	zap.L().
		Debug("Fetching GitHub totals")
	query := `
	query($login: String!) {
		user(login: $login) {
			issues {
				totalCount
			}
			pullRequests {
				totalCount
			}
			contributionsCollection {
				pullRequestReviewContributions {
					totalCount
				}
			}
			starredRepositories {
				totalCount
			}
			sponsorshipsAsMaintainer {
				totalCount
			}
			organizations {
				totalCount
			}
			watching {
				totalCount
			}
		}
	}`

	variables := map[string]interface{}{
		"login": userName,
	}

	var result struct {
		User struct {
			Issues struct {
				TotalCount int `json:"totalCount"`
			} `json:"issues"`
			PullRequests struct {
				TotalCount int `json:"totalCount"`
			} `json:"pullRequests"`
			ContributionsCollection struct {
				PullRequestReviewContributions struct {
					TotalCount int `json:"totalCount"`
				} `json:"pullRequestReviewContributions"`
			} `json:"contributionsCollection"`
			StarredRepositories struct {
				TotalCount int `json:"totalCount"`
			} `json:"starredRepositories"`
			SponsorshipsAsMaintainer struct {
				TotalCount int `json:"totalCount"`
			} `json:"sponsorshipsAsMaintainer"`
			Organizations struct {
				TotalCount int `json:"totalCount"`
			} `json:"organizations"`
			Watching struct {
				TotalCount int `json:"totalCount"`
			} `json:"watching"`
		} `json:"user"`
	}

	if err := QueryGitHubQLAPI(query, variables, &result); err != nil {
		zap.L().Fatal("Failed to get GitHub totals", zap.Error(err))
	}

	response := &GitHubTotals{
		TotalPullRequests:          result.User.PullRequests.TotalCount,
		TotalIssues:                result.User.Issues.TotalCount,
		TotalPullRequestReviews:    result.User.ContributionsCollection.PullRequestReviewContributions.TotalCount,
		TotalStarredRepos:          result.User.StarredRepositories.TotalCount,
		TotalSponsors:              result.User.SponsorshipsAsMaintainer.TotalCount,
		TotalMemberOfOrganizations: result.User.Organizations.TotalCount,
		TotalWatchers:              result.User.Watching.TotalCount,
	}
	zap.L().
		Debug("GitHub totals fetched",
			zap.Int("total_pull_requests", response.TotalPullRequests),
			zap.Int("total_issues", response.TotalIssues),
			zap.Int("total_pr_reviews", response.TotalPullRequestReviews),
			zap.Int("total_starred_repos", response.TotalStarredRepos),
			zap.Int("total_sponsors", response.TotalSponsors),
			zap.Int("total_member_of_organizations", response.TotalMemberOfOrganizations),
			zap.Int("total_watchers", response.TotalWatchers),
		)
	return response
}

type GitHubTotalsStats struct {
	TotalCommits               int
	TotalIssues                int
	TotalPullRequests          int
	TotalPullRequestReviews    int
	TotalStarredRepos          int
	TotalSponsors              int
	TotalMemberOfOrganizations int
	TotalWatchers              int
}

func getGitHubTotalsStats(userName, userId string) *GitHubTotalsStats {
	totals := getGitHubTotals(userName, userId)
	totalCommits := getCommitsTotal(userName, userId)

	return &GitHubTotalsStats{
		TotalCommits:               totalCommits,
		TotalPullRequests:          totals.TotalPullRequests,
		TotalIssues:                totals.TotalIssues,
		TotalPullRequestReviews:    totals.TotalPullRequestReviews,
		TotalStarredRepos:          totals.TotalStarredRepos,
		TotalSponsors:              totals.TotalSponsors,
		TotalMemberOfOrganizations: totals.TotalMemberOfOrganizations,
		TotalWatchers:              totals.TotalWatchers,
	}
}

// LanguageStat represents statistics for a programming language
type LanguageStat struct {
	Name       string
	Color      string
	TotalBytes int64
	Percentage float64
}

// getLanguageStats fetches and aggregates language statistics across all user repositories
func getLanguageStats(userName string) []LanguageStat {
	zap.L().Debug("Fetching language statistics")

	query := `
	query($login: String!, $after: String) {
		user(login: $login) {
			repositories(first: 100, after: $after, ownerAffiliations: [OWNER, ORGANIZATION_MEMBER, COLLABORATOR]) {
				pageInfo {
					hasNextPage
					endCursor
				}
				nodes {
					name
					languages(first: 10, orderBy: {field: SIZE, direction: DESC}) {
						edges {
							size
							node {
								name
								color
							}
						}
					}
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"login": userName,
	}

	// Map to aggregate language bytes across all repositories
	languageMap := make(map[string]*LanguageStat)
	totalBytes := int64(0)

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
						Name      string `json:"name"`
						Languages struct {
							Edges []struct {
								Size int64 `json:"size"`
								Node struct {
									Name  string `json:"name"`
									Color string `json:"color"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"languages"`
					} `json:"nodes"`
				} `json:"repositories"`
			} `json:"user"`
		}

		if err := QueryGitHubQLAPI(query, variables, &result); err != nil {
			zap.L().Fatal("Failed to get language statistics", zap.Error(err))
		}

		// Aggregate language statistics
		for _, repo := range result.User.Repositories.Nodes {
			for _, edge := range repo.Languages.Edges {
				langName := edge.Node.Name
				if stat, exists := languageMap[langName]; exists {
					stat.TotalBytes += edge.Size
				} else {
					languageMap[langName] = &LanguageStat{
						Name:       langName,
						Color:      edge.Node.Color,
						TotalBytes: edge.Size,
					}
				}
				totalBytes += edge.Size
			}
		}

		hasNextPage = result.User.Repositories.PageInfo.HasNextPage
		cursor = result.User.Repositories.PageInfo.EndCursor
	}

	// If no language data found, return empty slice
	if totalBytes == 0 {
		zap.L().Debug("No language data found")
		return []LanguageStat{}
	}

	// Calculate percentages and filter languages with < 1%
	languages := []LanguageStat{}
	for _, stat := range languageMap {
		percentage := float64(stat.TotalBytes) / float64(totalBytes) * 100.0
		if percentage >= 1.0 {
			stat.Percentage = percentage
			languages = append(languages, *stat)
		}
	}

	// Renormalize percentages to sum to 100% after filtering
	if len(languages) > 0 {
		totalPercentage := 0.0
		for _, lang := range languages {
			totalPercentage += lang.Percentage
		}
		for i := range languages {
			languages[i].Percentage = (languages[i].Percentage / totalPercentage) * 100.0
		}
	}

	// Sort languages by percentage in descending order
	// Using a simple bubble sort for clarity
	for i := 0; i < len(languages); i++ {
		for j := i + 1; j < len(languages); j++ {
			if languages[j].Percentage > languages[i].Percentage {
				languages[i], languages[j] = languages[j], languages[i]
			}
		}
	}

	zap.L().Debug("Language statistics fetched",
		zap.Int("total_languages", len(languages)),
		zap.Int64("total_bytes", totalBytes))

	return languages
}

// ContributionDay represents a single day's contribution data
type ContributionDay struct {
	Date              string
	ContributionCount int
	Color             string
}

// ContributionCalendar represents the contribution calendar data
type ContributionCalendar struct {
	TotalContributions int
	Weeks              []ContributionWeek
}

// ContributionWeek represents a week of contribution days
type ContributionWeek struct {
	ContributionDays []ContributionDay
}

// getContributionCalendar fetches the user's contribution calendar from GitHub
func getContributionCalendar(userName string) *ContributionCalendar {
	zap.L().Debug("Fetching contribution calendar")

	query := `
	query($login: String!) {
		user(login: $login) {
			contributionsCollection {
				contributionCalendar {
					totalContributions
					weeks {
						contributionDays {
							date
							contributionCount
							color
						}
					}
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"login": userName,
	}

	var result struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []struct {
							Date              string `json:"date"`
							ContributionCount int    `json:"contributionCount"`
							Color             string `json:"color"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	if err := QueryGitHubQLAPI(query, variables, &result); err != nil {
		zap.L().Fatal("Failed to get contribution calendar", zap.Error(err))
	}

	// Convert the result to our data structure
	calendar := &ContributionCalendar{
		TotalContributions: result.User.ContributionsCollection.ContributionCalendar.TotalContributions,
		Weeks:              make([]ContributionWeek, 0),
	}

	for _, week := range result.User.ContributionsCollection.ContributionCalendar.Weeks {
		contributionWeek := ContributionWeek{
			ContributionDays: make([]ContributionDay, 0),
		}
		for _, day := range week.ContributionDays {
			contributionWeek.ContributionDays = append(
				contributionWeek.ContributionDays,
				ContributionDay{
					Date:              day.Date,
					ContributionCount: day.ContributionCount,
					Color:             day.Color,
				},
			)
		}
		calendar.Weeks = append(calendar.Weeks, contributionWeek)
	}

	zap.L().Debug("Contribution calendar fetched",
		zap.Int("total_contributions", calendar.TotalContributions),
		zap.Int("total_weeks", len(calendar.Weeks)))

	return calendar
}
