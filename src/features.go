package main

import (
	svg "github.com/ajstarks/svgo"
	"go.uber.org/zap"
)

// generateSVGContent orchestrates the drawing of the SVG card and its components,
// including the header, metrics, language bars, and footer.
func generateSVGContent(svgCanvas *svg.SVG) {
	// Card coordinates used by multiple sections.
	cardX, cardY := 40, 30
	cardW, cardH := 720, 200
	drawCard(svgCanvas, cardX, cardY, cardW, cardH)

	githubUserInfo, err := getGitHubUserInfo()
	if err != nil {
		zap.L().Fatal("Failed to get GitHub user info", zap.Error(err))
	}

	handle := "@" + githubUserInfo.Login
	createStandardHeader(svgCanvas, cardX, cardY, cardW, githubUserInfo.AvatarURL, handle, githubUserInfo.Name)
	createMetricsBar(svgCanvas, cardX, cardY, githubUserInfo.Login)

	drawLanguageBars(svgCanvas, cardX, cardY)
	drawFooter(svgCanvas, cardX, cardY, cardH)
}

// createStandardHeader draws the standard header section of the SVG card, including the user's name,
// handle, and avatar
func createStandardHeader(svgCanvas *svg.SVG, cardX, cardY, cardW int, avatarURL string, handle string, userName string) {
	drawHeader(svgCanvas, cardX, cardY, userName, handle)
	drawAvatar(svgCanvas, cardX, cardY, cardW, avatarURL)
}

// createMetricsBar fetches user metrics and draws the metrics bar on the SVG card.
func createMetricsBar(svgCanvas *svg.SVG, cardX, cardY int, userName string) {
	prTotal, prErr := getPullRequestTotal(userName)
	if prErr != nil {
		zap.L().Fatal("Failed to get pull request total", zap.Error(prErr))
	}
	issuesTotal, err := getIssuesTotal(userName)
	if err != nil {
		zap.L().Fatal("Failed to get GitHub issues total", zap.Error(err))
	}

	commits, err := getCommitsTotal(userName)
	if err != nil {
		zap.L().Fatal("Failed to get commits total", zap.Error(err))
	}
	drawMetrics(svgCanvas, cardX, cardY, prTotal, issuesTotal, commits)
}
