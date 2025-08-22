package main

import (
	svg "github.com/ajstarks/svgo"
	"go.uber.org/zap"
)

func generateSVGContent(svgCanvas *svg.SVG) {

	getPullRequestTotal()
	// Card coordinates used by multiple sections.
	cardX, cardY := 40, 30
	cardW, cardH := 720, 200
	drawCard(svgCanvas, cardX, cardY, cardW, cardH)
	drawStandardHeader(svgCanvas, cardX, cardY)

	drawMetrics(svgCanvas, cardX, cardY)
	drawLanguageBars(svgCanvas, cardX, cardY)
	drawFooter(svgCanvas, cardX, cardY, cardH)
}

func drawStandardHeader(svgCanvas *svg.SVG, cardX, cardY, cardW int) {
	drawHeader(svgCanvas, cardX, cardY)
	drawAvatar(svgCanvas, cardX, cardY, cardW, "JackPlowman") // TODO: get username from GITHUB_TOKEN
}

// drawAvatar renders the user's avatar as a rounded square. The
// design intentionally avoids a circle per project requirements.
func drawAvatar(canvas *svg.SVG, cardX, cardY, cardW int, username string) {
	avatarX, avatarY := cardX+cardW-92, cardY+18

	// Get the user's avatar URL from GitHub
	avatarURL, err := getUserAvatarURL(username)
	if err != nil {
		zap.L().Warn("Failed to fetch avatar URL, using placeholder", zap.Error(err))
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
