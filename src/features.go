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
	drawStandardHeader(svgCanvas, cardX, cardY, cardW)

	getPullRequestTotal()
	drawMetrics(svgCanvas, cardX, cardY)
	drawLanguageBars(svgCanvas, cardX, cardY)
	drawFooter(svgCanvas, cardX, cardY, cardH)
}

// drawStandardHeader draws the standard header section of the SVG card, including the user's name,
// handle, and avatar
func drawStandardHeader(svgCanvas *svg.SVG, cardX, cardY, cardW int) {
	// Try to fetch user info (avatar, login, name) from GitHub once.
	avatarURL, userTag, userName, err := getUserInfo()

	// Fallbacks if API didn't return name/login
	displayName := userName
	if displayName == "" {
		displayName = "Name Not Found"
	}
	handle := "@" + userTag
	if userTag == "" {
		handle = "Unknown"
	}

	drawHeader(svgCanvas, cardX, cardY, displayName, handle)
	drawAvatar(svgCanvas, cardX, cardY, cardW, avatarURL, err)
}

// drawAvatar renders the user's avatar as a rounded square. The
// design intentionally avoids a circle per project requirements.
func drawAvatar(canvas *svg.SVG, cardX, cardY, cardW int, avatarURL string, avatarURLErr error) {
	avatarX, avatarY := cardX+cardW-92, cardY+18

	if avatarURLErr != nil {
		zap.L().Warn("Failed to fetch avatar URL, using placeholder", zap.Error(avatarURLErr))
		// Fallback to a coloured rectangle
		canvas.Roundrect(avatarX, avatarY, 64, 64, 8, 8, "fill:#1f6feb")
		return
	}

	// Create a clipping path for the rounded rectangle
	canvas.ClipPath(`id="avatar-clip"`)
	canvas.Roundrect(avatarX, avatarY, 64, 64, 8, 8, "")
	canvas.ClipEnd()

	// Use the avatar URL directly in an SVG image element
	canvas.Image(avatarX, avatarY, 64, 64, avatarURL, `clip-path="url(#avatar-clip)"`)
}
