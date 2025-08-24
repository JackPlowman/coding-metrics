package main

import (
	"os"

	"github.com/twpayne/go-svg"
	"go.uber.org/zap"
)

// initialize global logger
func init() {
	logger, err := initLogger()
	zap.ReplaceGlobals(zap.Must(logger, err))
}

// initLogger initializes and returns a zap logger according to the
// DEBUG environment variable. If DEBUG=="true" a development logger
// will be returned, otherwise a production logger is used.
func initLogger() (*zap.Logger, error) {
	if os.Getenv("INPUT_DEBUG") == "true" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

// main is the entry point for the application.
func main() {

	svgElements := []svg.Element{}
	svgElements = append(svgElements, generateSVGContent()...)
	svg := createSVG(svgElements)
	file := createLocalFile(svg)
	commitSVGChanges(file)
}

var title = "Jack Plowman - GitHub Stats"
var desc = "GitHub profile statistics visualization"

// Color palette - light theme to match target
var (
	textPrimary   = "#24292f"    // Dark text for light mode
	textSecondary = "#656d76"    // Secondary gray text
	accentBlue    = "#0969da"    // GitHub blue
	greenLight    = "#9be9a8"    // Light green for contribution graph
	greenMedium   = "#40c463"    // Medium green
	greenDark     = "#216e39"    // Dark green
)

func generateSVGContent() []svg.Element {
	elements := []svg.Element{
		svg.Title(svg.CharData(title)),
		svg.Desc(svg.CharData(desc)),

		// Main container background
		svg.Rect().Fill(svg.String("#ffffff")).Width(svg.Px(800)).Height(svg.Px(380)).X(svg.Px(0)).Y(svg.Px(0)),

		// Profile section (top left)
		generateProfileSection(),

		// Contribution graph (top right)
		generateContributionGraph(),

		// Stats sections (middle row)
		generateStatsRow(),

		// Languages section (bottom)
		generateLanguagesSection(),
	}

	return elements
}

func generateProfileSection() svg.Element {
	return svg.G().AppendChildren(
		// Profile avatar (circle) - smaller and positioned like target
		svg.Circle().Fill(svg.String(accentBlue)).CXCYR(30, 40, 12, svg.Px),

		// Name - positioned next to avatar
		svg.Text(svg.CharData("Jack Plowman")).XY(50, 45, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 18px; font-weight: 600;")),

		// Joined info
		svg.Text(svg.CharData("⏰ Joined GitHub 5 years ago")).XY(20, 70, svg.Px).Fill(svg.String(textSecondary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),

		// Followed by
		svg.Text(svg.CharData("👥 Followed by 6 users")).XY(20, 88, svg.Px).Fill(svg.String(textSecondary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),

		// Activity header
		svg.Text(svg.CharData("📈 Activity")).XY(20, 115, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),

		// Activity stats - more compact
		svg.Text(svg.CharData("○ 5560 Commits")).XY(20, 135, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("📋 122 Pull requests reviewed")).XY(20, 150, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("🔀 4892 Pull requests opened")).XY(20, 165, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("⭕ 1420 Issues opened")).XY(20, 180, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("💬 1872 Issue comments")).XY(20, 195, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),

		// Languages
		svg.Text(svg.CharData("🗣️ 21 Languages")).XY(20, 220, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),
	)
}

func generateContributionGraph() svg.Element {
	// Create a simple contribution graph with green squares like in the target
	squares := []svg.Element{}

	// Generate contribution squares pattern
	daysInMonth := 31
	squareSize := 10
	squareGap := 2
	startX := 530
	startY := 50

	for day := 0; day < 31; day++ {
		x := startX + (day%7)*(squareSize+squareGap)
		y := startY + (day/7)*(squareSize+squareGap)

		// Vary the green intensity to simulate real contribution data
		var color string
		intensity := day % 4
		switch intensity {
		case 0:
			color = "#ebedf0" // Light gray (no contributions)
		case 1:
			color = greenLight
		case 2:
			color = greenMedium
		case 3:
			color = greenDark
		}

		squares = append(squares, svg.Rect().
			Fill(svg.String(color)).
			Width(svg.Px(float64(squareSize))).
			Height(svg.Px(float64(squareSize))).
			X(svg.Px(float64(x))).
			Y(svg.Px(float64(y))).
			RX(svg.Px(2)))
	}

	// Draw empty squares for the rest of the month (up to 35 squares for 5 weeks)
	for i := daysInMonth; i < 35; i++ {
		x := startX + (i%7)*(squareSize+squareGap)
		y := startY + (i/7)*(squareSize+squareGap)

		squares = append(squares, svg.Rect().
			Fill(svg.String("#ffffff")).
			Stroke(svg.String("#ebedf0")).
			Width(svg.Px(float64(squareSize))).
			Height(svg.Px(float64(squareSize))).
			X(svg.Px(float64(x))).
			Y(svg.Px(float64(y))).
			RX(svg.Px(2)))
	}

	// Add contribution graph header
	headerElements := []svg.Element{
		svg.Text(svg.CharData("Contributed to 52 repositories")).XY(530, 40, svg.Px).Fill(svg.String(textSecondary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
	}

	return svg.G().AppendChildren(append(headerElements, squares...)...)
}

func generateStatsRow() svg.Element {
	return svg.G().AppendChildren(
		// Community stats section
		svg.Text(svg.CharData("🐙 Community stats")).XY(20, 265, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),

		svg.Text(svg.CharData("📊 Member of 0 organizations")).XY(250, 265, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("👤 Following 13 users")).XY(250, 280, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("💝 Sponsoring 0 repositories")).XY(250, 295, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("⭐ Starred 136 repositories")).XY(450, 265, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("👀 Watching 42 repositories")).XY(450, 280, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),

		// Repository stats
		svg.Text(svg.CharData("📚 56 Repositories")).XY(530, 180, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),

		svg.Text(svg.CharData("⚖️ Prefers MIT license")).XY(530, 200, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("🏷️ 77 Releases")).XY(530, 215, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("📦 3 Packages")).XY(530, 230, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("💾 69.1 MB used")).XY(530, 245, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),

		// Right column stats
		svg.Text(svg.CharData("💖 0 Sponsors")).XY(680, 180, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("⭐ 9 Stargazers")).XY(680, 195, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("🍴 9 Forkers")).XY(680, 210, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("👁️ 38 Watchers")).XY(680, 225, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
	)
}

func generateLanguagesSection() svg.Element {
	languages := []struct {
		name  string
		color string
		width int
	}{
		{"Python", "#3776ab", 120},
		{"TypeScript", "#2b7489", 100},
		{"Just", "#384d54", 80},
		{"HCL", "#844fba", 60},
		{"Shell", "#89e051", 70},
		{"Go", "#00add8", 85},
		{"JavaScript", "#f1e05a", 90},
		{"CSS", "#563d7c", 40},
	}

	elements := []svg.Element{
		// Most used languages header
		svg.Text(svg.CharData("Most used languages")).XY(20, 335, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),
	}

	// Single continuous language bar (like in the target)
	currentX := 20
	for _, lang := range languages {
		// Language bar segment
		elements = append(elements, svg.Rect().
			Fill(svg.String(lang.color)).
			Width(svg.Px(float64(lang.width))).
			Height(svg.Px(8)).
			X(svg.Px(float64(currentX))).
			Y(svg.Px(350)))

		currentX += lang.width
	}

	// Language labels below the bar
	currentX = 20
	for _, lang := range languages {
		// Language color dot
		elements = append(elements, svg.Circle().
			Fill(svg.String(lang.color)).
			CXCYR(float64(currentX+8), 370, 4, svg.Px))

		// Language label
		elements = append(elements, svg.Text(svg.CharData(lang.name)).
			XY(float64(currentX+16), 374, svg.Px).
			Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 12px;")))

		currentX += lang.width + 20
	}

	return svg.G().AppendChildren(elements...)
}
