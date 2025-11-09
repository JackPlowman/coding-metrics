package main

import (
	"fmt"
	"time"

	"github.com/twpayne/go-svg"
)

var (
	title = "Jack Plowman - GitHub Stats"
	desc  = "GitHub profile statistics visualization"
)

var (
	textPrimary   = "#24292f" // Dark text for light mode
	textSecondary = "#656d76" // Secondary gray text
	accentBlue    = "#0969da" // GitHub blue
)

// Generate the main SVG content
func generateSVGContent() []svg.Element {
	userInfo := getGitHubUserInfo()
	userId := getUserId(userInfo.Login)
	githubTotalsStats := getGitHubTotalsStats(userInfo.Login, userId)
	languageStats := getLanguageStats(userInfo.Login)
	contributionCalendar := getContributionCalendar(userInfo.Login)
	elements := []svg.Element{
		svg.Title(svg.CharData(title)),
		svg.Desc(svg.CharData(desc)),

		// Profile section (top left)
		generateProfileSection(userInfo),

		// Stats sections (middle row)
		generateStatsRow(userInfo, githubTotalsStats, contributionCalendar),

		// Languages section (bottom)
		generateLanguagesSection(languageStats),
	}

	return elements
}

// Generate profile section of svg
func generateProfileSection(userInfo *GitHubUserInfo) svg.Element {
	yearsAgo := time.Since(userInfo.JoinedGitHub).Hours() / 24 / 365
	return svg.G().AppendChildren(
		// Use <foreignObject> for rounded avatar if <image> can't have rounded corners
		svg.Image().
			Href(svg.String(userInfo.AvatarURL)).
			Width(svg.Px(24)).Height(svg.Px(24)).
			X(svg.Px(18)).Y(svg.Px(28)).
			Class(svg.String("avatar")),

		// Name - positioned next to avatar
		svg.Text(svg.CharData(userInfo.Name)).XY(50, 45, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 18px; font-weight: 600;")),

		// Joined info
		svg.Text(svg.CharData(fmt.Sprintf("⏰ Joined GitHub %.0f years ago", yearsAgo))).
			XY(20, 70, svg.Px).
			Fill(svg.String(textSecondary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),

		// Followed by
		svg.Text(svg.CharData(fmt.Sprintf("👥 Followed by %d users", userInfo.Followers))).
			XY(20, 88, svg.Px).
			Fill(svg.String(textSecondary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
	)
}

// Generate stats row of svg
func generateStatsRow(userInfo *GitHubUserInfo, githubTotalsStats *GitHubTotalsStats, contributionCalendar *ContributionCalendar) svg.Element {
	activityStatsX := 20.0
	communityStatsX := 250.0
	repositoriesStatsX := 480.0
	headersRowY := 115.0
	row1Y := 133.0
	row2Y := row1Y + 16.0
	row3Y := row2Y + 16.0
	row4Y := row3Y + 16.0
	headerStyle := svg.String(
		"font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;",
	)
	textStyle := svg.String(
		"font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;",
	)
	return svg.G().AppendChildren(
		// Activity stats section
		svg.Text(svg.CharData("📈 Activity")).
			XY(activityStatsX, headersRowY, svg.Px).
			Fill(svg.String(accentBlue)).
			Style(headerStyle),

		svg.Text(svg.CharData(fmt.Sprintf("○ %d Commits", githubTotalsStats.TotalCommits))).
			XY(activityStatsX, row1Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("📋 %d Pull requests reviewed", githubTotalsStats.TotalPullRequestReviews))).
			XY(activityStatsX, row2Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("🔀 %d Pull requests opened", githubTotalsStats.TotalPullRequests))).
			XY(activityStatsX, row3Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("⭕ %d Issues opened", githubTotalsStats.TotalIssues))).
			XY(activityStatsX, row4Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),

		// Community stats section
		svg.Text(svg.CharData("🐙 Community stats")).
			XY(communityStatsX, headersRowY, svg.Px).
			Fill(svg.String(accentBlue)).
			Style(headerStyle),

		svg.Text(svg.CharData(fmt.Sprintf("📊 Member of %d organizations", githubTotalsStats.TotalMemberOfOrganizations))).
			XY(communityStatsX, row1Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("👤 Following %d users", userInfo.Following))).
			XY(communityStatsX, row2Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("⭐ Starred %d repositories", githubTotalsStats.TotalStarredRepos))).
			XY(communityStatsX, row3Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData("👀 Watching 42 repositories")).
			XY(communityStatsX, row4Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),

		// Repository stats
		svg.Text(svg.CharData("📚 56 Repositories")).
			XY(repositoriesStatsX, headersRowY, svg.Px).
			Fill(svg.String(accentBlue)).
			Style(headerStyle),

		svg.Text(svg.CharData(fmt.Sprintf("💖 %d Sponsors", githubTotalsStats.TotalSponsors))).
			XY(repositoriesStatsX, row1Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData("⭐ 9 Stargazers")).
			XY(repositoriesStatsX, row2Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData("🍴 9 Forkers")).
			XY(repositoriesStatsX, row3Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("👁️ %d Watchers", githubTotalsStats.TotalWatchers))).
			XY(repositoriesStatsX, row4Y, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(textStyle),

		// Contribution graph
		generateContributionGraph(headerStyle, textStyle, contributionCalendar),
	)
}

func generateContributionGraph(headerStyle, textStyle svg.String, contributionCalendar *ContributionCalendar) svg.Element {
	// Create a contribution graph showing the current month in rows of 7 days
	squares := generateMonthContributionSquares(contributionCalendar)

	// Add contribution graph header
	headerElements := []svg.Element{
		svg.Text(svg.CharData("📚 Contributions")).XY(630, 115, svg.Px).Fill(svg.String(accentBlue)).
			Style(headerStyle),
		svg.Text(svg.CharData(fmt.Sprintf("%d contributions in the last year", contributionCalendar.TotalContributions))).
			XY(630, 210, svg.Px).
			Fill(svg.String(textSecondary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
	}

	return svg.G().AppendChildren(append(headerElements, squares...)...)
}

func generateMonthContributionSquares(contributionCalendar *ContributionCalendar) []svg.Element {
	squares := []svg.Element{}

	// Generate contribution squares pattern
	squareSize := 11
	squareGap := 2
	startX := 650
	startY := 125

	// Get current month data
	now := time.Now()
	currentYear, currentMonth := now.Year(), now.Month()

	// Determine days in the current month
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC)
	firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)
	daysInMonth := int(firstOfNextMonth.Sub(firstOfMonth).Hours() / 24)

	// Collect contribution data for the current month
	monthContributions := getMonthContributions(contributionCalendar, currentYear, currentMonth)

	// Draw contribution squares in rows of 7 days
	daysPerRow := 7
	for dayIndex := 0; dayIndex < daysInMonth; dayIndex++ {
		row := dayIndex / daysPerRow
		col := dayIndex % daysPerRow

		x := startX + col*(squareSize+squareGap)
		y := startY + row*(squareSize+squareGap)

		// Get color for this day if we have data
		colour := "#ebedf0" // Default: no contributions
		if dayIndex < len(monthContributions) && monthContributions[dayIndex].Color != "" {
			colour = monthContributions[dayIndex].Color
		}

		squares = append(squares, svg.Rect().
			Fill(svg.String(colour)).
			Width(svg.Px(float64(squareSize))).
			Height(svg.Px(float64(squareSize))).
			X(svg.Px(float64(x))).
			Y(svg.Px(float64(y))).
			RX(svg.Px(2)))
	}

	return squares
}

func getMonthContributions(contributionCalendar *ContributionCalendar, year int, month time.Month) []ContributionDay {
	monthContributions := []ContributionDay{}
	for _, week := range contributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			dayDate, err := time.Parse("2006-01-02", day.Date)
			if err != nil {
				continue
			}
			// Check if day is in the specified month
			if dayDate.Year() == year && dayDate.Month() == month {
				monthContributions = append(monthContributions, day)
			}
		}
	}
	return monthContributions
}

func generateLanguagesSection(languages []LanguageStat) svg.Element {
	// If no languages data, return empty group
	if len(languages) == 0 {
		return svg.G()
	}

	// Calculate total width available for the bar (SVG width minus margins)
	const svgWidth = 1000.0
	const marginLeft = 20.0
	const marginRight = 20.0
	const barWidth = svgWidth - marginLeft - marginRight

	elements := []svg.Element{
		// Most used languages header
		// Languages
		svg.Text(svg.CharData(fmt.Sprintf("🗣️ %d Languages", len(languages)))).
			XY(20, 220, svg.Px).
			Fill(svg.String(accentBlue)).
			Style(svg.String(
				"font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;",
			)),
		svg.Text(svg.CharData("Most used languages")).
			XY(400, 240, svg.Px).
			Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 12px; font-weight: 600;")),
	}

	// Single continuous language bar spanning full width
	currentX := marginLeft
	segmentWidths := make([]float64, len(languages))

	for i, lang := range languages {
		// Calculate proportional width based on percentage (percentages sum to 100)
		segmentWidth := (lang.Percentage / 100.0) * barWidth
		segmentWidths[i] = segmentWidth

		// Language bar segment
		elements = append(elements, svg.Rect().
			Fill(svg.String(lang.Color)).
			Width(svg.Px(segmentWidth)).
			Height(svg.Px(8)).
			X(svg.Px(currentX)).
			Y(svg.Px(260)))

		currentX += segmentWidth
	}

	// Language labels below the bar - all centered with good spacing
	const labelSpacing = 80.0 // Horizontal spacing between labels
	totalLabelsWidth := float64(len(languages)-1) * labelSpacing
	startLabelX := (svgWidth - totalLabelsWidth) / 2 // Center the entire label group
	labelY := 290.0

	for i, lang := range languages {
		labelX := startLabelX + (float64(i) * labelSpacing)

		// Position dot to the left of the label with consistent gap
		// Dot has 4px radius, so we need dot center at (text start - 4px radius - gap)
		const dotRadius = 4.0
		const gapBetweenDotAndText = 6.0 // Visual gap between dot edge and text

		textWidthEstimate := float64(len(lang.Name)) * 6.0
		textStartX := labelX - (textWidthEstimate / 2) // Since text is centered
		dotX := textStartX - dotRadius - gapBetweenDotAndText

		// Language colour dot - to the left of the label text
		elements = append(elements, svg.Circle().
			Fill(svg.String(lang.Color)).
			CXCYR(dotX, labelY-4, dotRadius, svg.Px))

		// Language label - centered
		elements = append(elements, svg.Text(svg.CharData(lang.Name)).
			XY(labelX, labelY, svg.Px).
			Fill(svg.String(textPrimary)).
			TextAnchor(svg.String("middle")).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 12px;")),
		)
	}

	return svg.G().AppendChildren(elements...)
}
