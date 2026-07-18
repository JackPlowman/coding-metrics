package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	bearerPrefix   = "Bearer "
	maxAvatarBytes = 2 * 1024 * 1024
)

// getGitHubUserInfo fetches the user's information from GitHub REST API
type GitHubUserInfo struct {
	AvatarURL    string    `json:"avatar_url"`
	Followers    int       `json:"followers"`
	Following    int       `json:"following"`
	JoinedGitHub time.Time `json:"created_at"`
	Login        string    `json:"login"`
	Name         string    `json:"name"`
	NodeID       string    `json:"node_id"`
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

// normalizeAvatarURL ensures the avatar URL renders when embedded in sanitized SVGs
func normalizeAvatarURL(avatarURL string) string {
	if avatarURL == "" {
		return avatarURL
	}

	parsed, err := url.Parse(avatarURL)
	if err != nil {
		zap.L().
			Warn("Failed to parse avatar URL", zap.String("avatar_url", avatarURL), zap.Error(err))
		return avatarURL
	}

	query := parsed.Query()
	if query.Get("s") == "" && query.Get("size") == "" {
		query.Set("s", "80")
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

// getAvatarHref returns a self-contained avatar data URI where possible.
// SVGs embedded as images often cannot load external image subresources, so
// using a data URI keeps the GitHub avatar visible in README/profile renders.
func getAvatarHref(avatarURL string) string {
	normalizedURL := normalizeAvatarURL(avatarURL)
	if normalizedURL == "" {
		return ""
	}

	dataURI := fetchAvatarDataURI(
		&http.Client{Timeout: 15 * time.Second},
		normalizedURL,
	)
	if dataURI != "" {
		return dataURI
	}

	return normalizedURL
}

func fetchAvatarDataURI(client *http.Client, avatarURL string) string {
	req, err := http.NewRequest("GET", avatarURL, nil)
	if err != nil {
		zap.L().Warn(
			"Failed to create avatar request",
			zap.String("avatar_url", avatarURL),
			zap.Error(err),
		)
		return ""
	}
	req.Header.Set("Accept", "image/*")
	req.Header.Set("User-Agent", "coding-metrics")

	resp, err := client.Do(req)
	if err != nil {
		zap.L().Warn("Failed to fetch avatar", zap.String("avatar_url", avatarURL), zap.Error(err))
		return ""
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			zap.L().Warn("Failed to close avatar response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		zap.L().Warn("Avatar request returned non-200 status",
			zap.String("avatar_url", avatarURL),
			zap.Int("status", resp.StatusCode))
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAvatarBytes+1))
	if err != nil {
		zap.L().Warn(
			"Failed to read avatar response",
			zap.String("avatar_url", avatarURL),
			zap.Error(err),
		)
		return ""
	}
	if len(body) == 0 || len(body) > maxAvatarBytes {
		zap.L().Warn("Avatar response was empty or too large",
			zap.String("avatar_url", avatarURL),
			zap.Int("bytes", len(body)))
		return ""
	}

	contentType := cleanImageContentType(resp.Header.Get("Content-Type"), body)
	if contentType == "" {
		zap.L().Warn("Avatar response was not an image",
			zap.String("avatar_url", avatarURL),
			zap.String("content_type", resp.Header.Get("Content-Type")))
		return ""
	}

	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(body)
}

func cleanImageContentType(contentType string, body []byte) string {
	contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return ""
	}
	return contentType
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

func getGitHubTotals(userName string) *GitHubTotals {
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
		TotalPullRequestReviews:    getPullRequestReviewTotal(userName),
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

func getPullRequestReviewTotal(userName string) int {
	query := `
	query($login: String!) {
		user(login: $login) {
			contributionsCollection {
				totalPullRequestReviewContributions
			}
		}
	}`

	var result struct {
		User struct {
			ContributionsCollection struct {
				TotalPullRequestReviewContributions int `json:"totalPullRequestReviewContributions"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	variables := map[string]interface{}{"login": userName}
	if err := QueryGitHubQLAPI(query, variables, &result); err != nil {
		zap.L().Warn("Failed to get pull request review total", zap.Error(err))
		return 0
	}

	return result.User.ContributionsCollection.TotalPullRequestReviewContributions
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
	totals := getGitHubTotals(userName)
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

// getContributionCalendar fetches the user's contribution calendar from GitHub.
// Use bounded ranges from the outset: a full-year calendar can exceed GitHub's
// per-query compute limit and failed parent queries can quickly trigger the
// secondary rate limit before an adaptive split reaches a safe range.
func getContributionCalendar(userName string) *ContributionCalendar {
	const (
		calendarChunkDays  = 28
		calendarChunkPause = 200 * time.Millisecond
	)

	today := startOfUTCDay(time.Now())
	start := today.AddDate(-1, 0, 0)
	start = start.AddDate(0, 0, -int(start.Weekday()))
	zap.L().Debug(
		"Fetching contribution calendar in bounded ranges",
		zap.String("from", start.Format("2006-01-02")),
		zap.String("to", today.Format("2006-01-02")),
		zap.Int("chunk_days", calendarChunkDays),
	)

	calendars := make([]*ContributionCalendar, 0, 14)
	chunkStart := start
	chunkIndex := 0
	for !chunkStart.After(today) {
		chunkEnd := chunkStart.AddDate(0, 0, calendarChunkDays-1)
		if chunkEnd.After(today) {
			chunkEnd = today
		}
		chunkIndex++
		zap.L().Debug(
			"Fetching contribution calendar chunk",
			zap.Int("chunk", chunkIndex),
			zap.String("from", chunkStart.Format("2006-01-02")),
			zap.String("to", chunkEnd.Format("2006-01-02")),
		)

		calendar, err := getContributionCalendarRange(userName, chunkStart, chunkEnd)
		if err != nil {
			zap.L().Fatal("Failed to get contribution calendar chunk", zap.Error(err))
		}
		calendars = append(calendars, calendar)
		chunkStart = chunkEnd.AddDate(0, 0, 1)
		if !chunkStart.After(today) {
			time.Sleep(calendarChunkPause)
		}
	}

	calendar := mergeContributionCalendars(start, today, calendars...)
	logContributionCalendarFetched(calendar, true)
	return calendar
}

func queryContributionCalendar(
	userName string,
	from, to *time.Time,
) (*ContributionCalendar, error) {
	query := `
	query($login: String!, $from: DateTime, $to: DateTime) {
		user(login: $login) {
			contributionsCollection(from: $from, to: $to) {
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

	variables := map[string]interface{}{"login": userName}
	if from != nil && to != nil {
		variables["from"] = from.Format(time.RFC3339)
		endOfDay := to.Add(24*time.Hour - time.Second)
		variables["to"] = endOfDay.Format(time.RFC3339)
	}

	var result struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []ContributionDay `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	if err := QueryGitHubQLAPI(query, variables, &result); err != nil {
		return nil, err
	}

	queryCalendar := result.User.ContributionsCollection.ContributionCalendar
	calendar := &ContributionCalendar{
		TotalContributions: queryCalendar.TotalContributions,
		Weeks:              make([]ContributionWeek, 0, len(queryCalendar.Weeks)),
	}
	for _, week := range queryCalendar.Weeks {
		calendar.Weeks = append(
			calendar.Weeks,
			ContributionWeek{ContributionDays: week.ContributionDays},
		)
	}
	return calendar, nil
}

func getContributionCalendarRange(
	userName string,
	from, to time.Time,
) (*ContributionCalendar, error) {
	zap.L().Debug(
		"Fetching contribution calendar range",
		zap.String("from", from.Format("2006-01-02")),
		zap.String("to", to.Format("2006-01-02")),
	)
	calendar, err := queryContributionCalendar(userName, &from, &to)
	if err == nil {
		return calendar, nil
	}

	days := int(to.Sub(from).Hours()/24) + 1
	if !isGraphQLResourceLimitError(err) || days <= 7 {
		return nil, fmt.Errorf(
			"contribution calendar range %s to %s: %w",
			from.Format("2006-01-02"),
			to.Format("2006-01-02"),
			err,
		)
	}

	midpoint := from.AddDate(0, 0, (days/2)-1)
	rightStart := midpoint.AddDate(0, 0, 1)
	zap.L().Warn(
		"Contribution calendar range exceeded GitHub resources; splitting range",
		zap.String("from", from.Format("2006-01-02")),
		zap.String("to", to.Format("2006-01-02")),
		zap.Int("days", days),
		zap.String("left_to", midpoint.Format("2006-01-02")),
		zap.String("right_from", rightStart.Format("2006-01-02")),
	)

	left, err := getContributionCalendarRange(userName, from, midpoint)
	if err != nil {
		return nil, err
	}
	right, err := getContributionCalendarRange(userName, rightStart, to)
	if err != nil {
		return nil, err
	}
	return mergeContributionCalendars(from, to, left, right), nil
}

func mergeContributionCalendars(
	from, to time.Time,
	calendars ...*ContributionCalendar,
) *ContributionCalendar {
	byDate := make(map[string]ContributionDay)
	for _, calendar := range calendars {
		for _, week := range calendar.Weeks {
			for _, day := range week.ContributionDays {
				byDate[day.Date] = day
			}
		}
	}

	merged := &ContributionCalendar{Weeks: make([]ContributionWeek, 0)}
	week := ContributionWeek{ContributionDays: make([]ContributionDay, 0, 7)}
	for date := from; !date.After(to); date = date.AddDate(0, 0, 1) {
		dateString := date.Format("2006-01-02")
		day, ok := byDate[dateString]
		if !ok {
			day = ContributionDay{Date: dateString, Color: githubContribNone}
		}
		merged.TotalContributions += day.ContributionCount
		week.ContributionDays = append(week.ContributionDays, day)
		if len(week.ContributionDays) == 7 {
			merged.Weeks = append(merged.Weeks, week)
			week = ContributionWeek{ContributionDays: make([]ContributionDay, 0, 7)}
		}
	}
	if len(week.ContributionDays) > 0 {
		merged.Weeks = append(merged.Weeks, week)
	}
	return merged
}

func isGraphQLResourceLimitError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "resource limit")
}

func startOfUTCDay(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func logContributionCalendarFetched(calendar *ContributionCalendar, usedBoundedRanges bool) {
	zap.L().Debug(
		"Contribution calendar fetched",
		zap.Int("total_contributions", calendar.TotalContributions),
		zap.Int("total_weeks", len(calendar.Weeks)),
		zap.Bool("used_bounded_ranges", usedBoundedRanges),
	)
}
