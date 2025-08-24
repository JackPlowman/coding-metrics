package main

import (
	"os"
	"time"
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
	)
}

func generateContributionGraph() svg.Element {
	// Create a simple contribution graph with green squares like in the target
	squares := []svg.Element{}

	// Generate contribution squares pattern
	squareSize := 10
	squareGap := 2
	startX := 530
	startY := 50

	// Determine days in the current month
	now := time.Now()
	year, month := now.Year(), now.Month()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	nextMonth := firstOfMonth.AddDate(0, 1, 0)
	daysInMonth := int(nextMonth.Sub(firstOfMonth).Hours() / 24)

	daysPerRow := 12

	for day := 0; day < daysInMonth; day++ {
		x := startX + (day%daysPerRow)*(squareSize+squareGap)
		y := startY + (day/daysPerRow)*(squareSize+squareGap)

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

	// Draw empty squares for the rest of the grid (up to 36 squares for 3 rows of 12)
	for i := daysInMonth; i < 36; i++ {
		x := startX + (i%daysPerRow)*(squareSize+squareGap)
		y := startY + (i/daysPerRow)*(squareSize+squareGap)

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
	activityStatsX := 20.0
	communityStatsX := 250.0
	repositoriesStatsX := 480.0
	headersRowY := 115.0
	row1Y := 135.0
	row2Y := 150.0
	row3Y := 165.0
	row4Y := 180.0
	row5Y := 195.0
	return svg.G().AppendChildren(
		// Activity header
		svg.Text(svg.CharData("📈 Activity")).XY(activityStatsX, headersRowY, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),

		// Activity stats - more compact
		svg.Text(svg.CharData("○ 5560 Commits")).XY(activityStatsX, row1Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("📋 122 Pull requests reviewed")).XY(activityStatsX, row2Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("🔀 4892 Pull requests opened")).XY(activityStatsX, row3Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("⭕ 1420 Issues opened")).XY(activityStatsX, row4Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("💬 1872 Issue comments")).XY(activityStatsX, row5Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),

		// Community stats section
		svg.Text(svg.CharData("🐙 Community stats")).XY(communityStatsX, headersRowY, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),

		svg.Text(svg.CharData("📊 Member of 0 organizations")).XY(communityStatsX, row1Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("👤 Following 13 users")).XY(communityStatsX, row2Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("💝 Sponsoring 0 repositories")).XY(communityStatsX, row3Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("⭐ Starred 136 repositories")).XY(communityStatsX, row4Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("👀 Watching 42 repositories")).XY(communityStatsX, row5Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),

		// Repository stats
		svg.Text(svg.CharData("📚 56 Repositories")).XY(repositoriesStatsX, headersRowY, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),

		svg.Text(svg.CharData("💖 0 Sponsors")).XY(repositoriesStatsX, row1Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("⭐ 9 Stargazers")).XY(repositoriesStatsX, row2Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("🍴 9 Forkers")).XY(repositoriesStatsX, row3Y, svg.Px).Fill(svg.String(textPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;")),
		svg.Text(svg.CharData("👁️ 38 Watchers")).XY(repositoriesStatsX, row4Y, svg.Px).Fill(svg.String(textPrimary)).
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
		// Languages
		svg.Text(svg.CharData("🗣️ 21 Languages")).XY(20, 220, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;")),
		svg.Text(svg.CharData("Most used languages")).XY(100, 335, svg.Px).Fill(svg.String(accentBlue)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 12px; font-weight: 600;")),
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
