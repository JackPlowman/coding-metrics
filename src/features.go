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

	avatarURL, userName, displayName, err := getGitHubUserInfo()
	if err != nil {
		zap.L().Fatal("Failed to get GitHub user info", zap.Error(err))
	}

	handle := "@" + userName
	createStandardHeader(svgCanvas, cardX, cardY, cardW, avatarURL, handle, displayName)
	createMetricsBar(svgCanvas, cardX, cardY, userName)

	drawLanguageBars(svgCanvas, cardX, cardY)
	drawFooter(svgCanvas, cardX, cardY, cardH)
}

// createStandardHeader draws the standard header section of the SVG card, including the user's name,
// handle, and avatar
func createStandardHeader(svgCanvas *svg.SVG, cardX, cardY, cardW int, avatarURL string, handle string, userName string) {
	drawHeader(svgCanvas, cardX, cardY, userName, handle)
	drawAvatar(svgCanvas, cardX, cardY, cardW, avatarURL)
}

func createMetricsBar(svgCanvas *svg.SVG, cardX, cardY int, userName string) {
	prTotal, prErr := getPullRequestTotal(userName)
	if prErr != nil {
		zap.L().Fatal("Failed to get pull request total", zap.Error(prErr))
	}
	issuesTotal, err := getIssuesTotal(userName)
	if err != nil {
		zap.L().Fatal("Failed to get GitHub issues total", zap.Error(err))
	}

	commits := QueryGitHubQLAPI("", nil, nil)
	drawMetrics(svgCanvas, cardX, cardY, prTotal, issuesTotal, commits)
}
