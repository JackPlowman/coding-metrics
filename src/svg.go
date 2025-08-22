package main

import (
	"os"
	"path/filepath"

	svg "github.com/ajstarks/svgo"
	"go.uber.org/zap"
)

// setupSVG initializes an SVG canvas and draws the background.
func setupSVG(file *os.File) *svg.SVG {
	width := 800
	height := 260
	canvas := svg.New(file)
	canvas.Start(width, height)

	// Draw the page using modular functions. These are kept small so
	// the program can grow and each piece can be tested/maintained
	// independently.
	drawBackground(canvas, width, height)

	return canvas
}

// completeSVG finalizes the SVG canvas by writing the closing tags.
func completeSVG(canvas *svg.SVG) {
	canvas.End()
}

// drawBackground paints the page background.
func drawBackground(canvas *svg.SVG, width, height int) {
	canvas.Rect(0, 0, width, height, "fill:#0d1117")
}

// drawCard draws the rounded card that contains all content.
func drawCard(canvas *svg.SVG, x, y, w, h int) {
	canvas.Roundrect(x, y, w, h, 12, 12, "fill:#0b1220;stroke:#121524;stroke-width:1")
}

// drawHeader renders the name and handle in the card header.
func drawHeader(canvas *svg.SVG, cardX, cardY int) {
	canvas.Text(cardX+22, cardY+36, "Jack Plowman", "fill:#cdd6f4;font-size:18px;font-weight:700")
	canvas.Text(cardX+22, cardY+58, "@jackplowman", "fill:#94a3b8;font-size:12px")
}

// drawMetrics renders the three metric boxes (commits, PRs, issues).
func drawMetrics(canvas *svg.SVG, cardX, cardY int) {
	boxW, boxH := 180, 64
	gap := 16
	startX := cardX + 22
	startY := cardY + 80

	// Commits
	canvas.Roundrect(startX, startY, boxW, boxH, 8, 8, "fill:#071029")
	canvas.Text(startX+14, startY+26, "1,234", "fill:#9be7ff;font-size:20px;font-weight:700")
	canvas.Text(startX+14, startY+46, "commits", "fill:#9aa6b8;font-size:12px")

	// PRs
	prX := startX + boxW + gap
	canvas.Roundrect(prX, startY, boxW, boxH, 8, 8, "fill:#071029")
	canvas.Text(prX+14, startY+26, "84", "fill:#ffd580;font-size:20px;font-weight:700")
	canvas.Text(prX+14, startY+46, "pull requests", "fill:#9aa6b8;font-size:12px")

	// Issues
	issuesX := prX + boxW + gap
	canvas.Roundrect(issuesX, startY, boxW, boxH, 8, 8, "fill:#071029")
	canvas.Text(issuesX+14, startY+26, "12", "fill:#ff9b9b;font-size:20px;font-weight:700")
	canvas.Text(issuesX+14, startY+46, "open issues", "fill:#9aa6b8;font-size:12px")
}

// drawLanguageBars renders the top language bars as simple rectangles.
func drawLanguageBars(canvas *svg.SVG, cardX, cardY int) {
	startX := cardX + 22
	startY := cardY + 80
	langY := startY + 64 + 20
	langX := startX
	canvas.Text(langX, langY, "Top languages", "fill:#94a3b8;font-size:12px")
	barX := langX
	barY := langY + 12
	canvas.Rect(barX, barY, 220, 12, "fill:#2b8bd6;rx:4;ry:4")
	canvas.Rect(barX+230, barY, 140, 12, "fill:#6bcB8b;rx:4;ry:4")
	canvas.Rect(barX+380, barY, 80, 12, "fill:#f6c85f;rx:4;ry:4")
}

// drawFooter renders a small note at the bottom of the card.
func drawFooter(canvas *svg.SVG, cardX, cardY, cardH int) {
	canvas.Text(cardX+22, cardY+cardH-18, "Generated (static) — data will be dynamic later", "fill:#6b7280;font-size:11px")
}

// createLocalFile creates a new SVG file in the system temp directory.
// If the temp directory is not writable, it falls back to stdout so the
// program doesn't panic in read-only environments (CI containers, Actions).
func createLocalFile() *os.File {
	path := filepath.Join(os.TempDir(), "output.svg")
	file, err := os.Create(path)
	if err == nil {
		zap.L().Info("Writing SVG to file", zap.String("path", path))
		return file
	}

	// Fallback: write to stdout, but warn to stderr.
	zap.L().Warn("Could not create SVG file, falling back to stdout", zap.String("path", path), zap.Error(err))
	return os.Stdout
}
