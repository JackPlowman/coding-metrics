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
		zap.L().Error("Failed to get GitHub user info", zap.Error(err))
		return
	}

	handle := "@" + userName
	drawStandardHeader(svgCanvas, cardX, cardY, cardW, avatarURL, handle, displayName)

	prTotal, prErr := getPullRequestTotal(userName)
	if prErr != nil {
		zap.L().Error("Failed to get pull request total", zap.Error(prErr))
		prTotal = -1 // Sentinel value indicating error
	}
	issuesTotal, err := getIssuesTotal(userName)
	if err != nil {
		zap.L().Error("Failed to get GitHub issues total", zap.Error(err))
		return
	}
	drawMetrics(svgCanvas, cardX, cardY, prTotal, issuesTotal)
	drawLanguageBars(svgCanvas, cardX, cardY)
	drawFooter(svgCanvas, cardX, cardY, cardH)
}

// drawStandardHeader draws the standard header section of the SVG card, including the user's name,
// handle, and avatar
func drawStandardHeader(svgCanvas *svg.SVG, cardX, cardY, cardW int, avatarURL string, handle string, userName string) {
	drawHeader(svgCanvas, cardX, cardY, userName, handle)
	drawAvatar(svgCanvas, cardX, cardY, cardW, avatarURL)
}

// drawAvatar renders the user's avatar as a rounded square. The
// design intentionally avoids a circle per project requirements.
func drawAvatar(canvas *svg.SVG, cardX, cardY, cardW int, avatarURL string) {
	avatarX, avatarY := cardX+cardW-92, cardY+18

	// Create a clipping path for the rounded rectangle
	canvas.ClipPath(`id="avatar-clip"`)
	canvas.Roundrect(avatarX, avatarY, 64, 64, 8, 8, "")
	canvas.ClipEnd()

	// Use the avatar URL directly in an SVG image element
	canvas.Image(avatarX, avatarY, 64, 64, avatarURL, `clip-path="url(#avatar-clip)"`)
}
