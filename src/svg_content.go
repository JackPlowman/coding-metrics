package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/twpayne/go-svg"
)

var (
	title = "Jack Plowman - GitHub Stats"
	desc  = "GitHub profile statistics visualization"
)

// Common font styles
const (
	fontStyle13px       = "font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 13px;"
	fontStyleHeader15px = "font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 15px; font-weight: 600;"
)

type contributionCalendarStats struct {
	CurrentStreakDays int
	BestStreakDays    int
	HighestInDay      int
	AveragePerDay     float64
}

type isometricLayout struct {
	OriginX     float64
	OriginY     float64
	TileW       float64
	TileH       float64
	HeightStep  float64
	MaxHeight   float64
	StrokeWidth float64
}

// Global colour profile - will be set in main based on user selection
var currentColourProfile ColourProfile

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

		// Year contribution calendar (bottom)
		generateYearContributionCalendarSection(contributionCalendar),
	}

	return elements
}

func generateYearContributionCalendarSection(
	contributionCalendar *ContributionCalendar,
) svg.Element {
	if contributionCalendar == nil || len(contributionCalendar.Weeks) == 0 {
		return svg.G()
	}
	const marginLeft = 20.0

	stats := calculateContributionCalendarStats(contributionCalendar)

	// Reserve space at the right for the notes panel
	const notesX = 700.0
	const graphRightPadding = 15.0

	weeks := contributionCalendar.Weeks
	rows := 7
	cols := len(weeks)

	tileW, tileH, heightStep, maxHeight, _ := calculateIsometricSizing(cols, rows, notesX)
	originY := 360.0

	// Align the isometric grid toward the notes panel to reduce empty right-side space.
	graphRight := (notesX - 40.0) - graphRightPadding
	originX := graphRight - (float64(cols+rows-1) * tileW / 2.0)
	if originX < marginLeft+tileW/2.0 {
		originX = marginLeft + tileW/2.0
	}
	layout := isometricLayout{
		OriginX:     originX,
		OriginY:     originY,
		TileW:       tileW,
		TileH:       tileH,
		HeightStep:  heightStep,
		MaxHeight:   maxHeight,
		StrokeWidth: 0.6,
	}

	elements := []svg.Element{
		svg.Text(svg.CharData("🗓️ Contributions calendar")).
			XY(marginLeft, 320, svg.Px).
			Fill(svg.String(currentColourProfile.AccentPrimary)).
			Style(svg.String(fontStyleHeader15px)),
	}

	elements = append(
		elements,
		generateIsometricBaseTiles(
			weeks,
			cols,
			rows,
			layout.OriginX,
			layout.OriginY,
			layout.TileW,
			layout.TileH,
		)...)
	elements = append(elements, generateIsometricExtrusions(weeks, cols, rows, layout)...)
	elements = append(elements, generateContributionCalendarNotes(stats, notesX, 330, 18)...)

	return svg.G().AppendChildren(elements...)
}

type point struct {
	X float64
	Y float64
}

func diamondPoints(x, y, tileW, tileH float64) []point {
	hw := tileW / 2.0
	hh := tileH / 2.0
	return []point{
		{X: x, Y: y},
		{X: x + hw, Y: y + hh},
		{X: x, Y: y + tileH},
		{X: x - hw, Y: y + hh},
	}
}

func toSVGPoints(points []point) svg.Points {
	out := make(svg.Points, 0, len(points))
	for _, p := range points {
		out = append(out, []float64{p.X, p.Y})
	}
	return out
}

func contributionLevelFromAPIColour(apiColour string) int {
	switch apiColour {
	case githubContribNone:
		return 0
	case githubContribLow:
		return 1
	case githubContribMediumLow:
		return 2
	case githubContribMediumHigh:
		return 3
	case githubContribHigh:
		return 4
	default:
		return 0
	}
}

func calculateContributionCalendarStats(
	contributionCalendar *ContributionCalendar,
) contributionCalendarStats {
	if contributionCalendar == nil {
		return contributionCalendarStats{}
	}

	dates, counts, maxInDay := collectContributionCounts(contributionCalendar)
	if len(dates) == 0 {
		return contributionCalendarStats{}
	}

	currentStreak := calculateCurrentStreak(dates, counts)
	bestStreak := calculateBestStreak(dates, counts)
	avg := calculateAveragePerDay(contributionCalendar.TotalContributions, len(dates))

	return contributionCalendarStats{
		CurrentStreakDays: currentStreak,
		BestStreakDays:    bestStreak,
		HighestInDay:      maxInDay,
		AveragePerDay:     avg,
	}
}

func calculateIsometricSizing(
	cols, rows int,
	notesX float64,
) (tileW, tileH, heightStep, maxHeight, gridWidth float64) {
	// A slightly shallower angle than the classic 2:1 isometric projection.
	// (Lower tileH relative to tileW => shallower projection.)
	tileW = 12.0
	tileH = 4.8
	heightStep = 4.0

	maxHeight = 4.0 * heightStep
	gridWidth = ((float64(cols+rows-2) * tileW) / 2.0) + tileW
	graphMaxWidth := notesX - 40.0

	// If we have extra horizontal room, scale up a bit (capped) so the calendar
	// fills more of the empty space before the notes panel.
	if gridWidth < graphMaxWidth {
		const maxScaleUp = 1.22
		scaleUp := graphMaxWidth / gridWidth
		if scaleUp > maxScaleUp {
			scaleUp = maxScaleUp
		}
		if scaleUp > 1.0 {
			tileW *= scaleUp
			tileH *= scaleUp
			heightStep *= scaleUp
			maxHeight = 4.0 * heightStep
			gridWidth = ((float64(cols+rows-2) * tileW) / 2.0) + tileW
		}
		return tileW, tileH, heightStep, maxHeight, gridWidth
	}

	// Otherwise scale down to fit.
	scaleDown := graphMaxWidth / gridWidth
	tileW *= scaleDown
	tileH *= scaleDown
	heightStep *= scaleDown
	maxHeight = 4.0 * heightStep
	gridWidth = ((float64(cols+rows-2) * tileW) / 2.0) + tileW
	return tileW, tileH, heightStep, maxHeight, gridWidth
}

func generateIsometricBaseTiles(
	weeks []ContributionWeek,
	cols, rows int,
	originX, originY, tileW, tileH float64,
) []svg.Element {
	elements := make([]svg.Element, 0)
	for s := 0; s <= (cols-1)+(rows-1); s++ {
		for row := rows - 1; row >= 0; row-- {
			col := s - row
			if col < 0 || col >= cols {
				continue
			}
			if row >= len(weeks[col].ContributionDays) {
				continue
			}
			day := weeks[col].ContributionDays[row]
			baseColour := currentColourProfile.ContributionLevel0
			if day.Color != "" {
				baseColour = currentColourProfile.GetContributionColour(day.Color)
			}
			x := originX + (float64(col+row) * tileW / 2.0)
			y := originY + (float64(col) * tileH / 2.0) - (float64(row) * tileH / 2.0)
			baseDiamond := diamondPoints(x, y, tileW, tileH)
			elements = append(elements,
				svg.Polygon().
					Points(toSVGPoints(baseDiamond)).
					Fill(svg.String(baseColour)).
					Stroke(svg.String(currentColourProfile.Background)).
					StrokeWidth(svg.Px(0.6)),
			)
		}
	}
	return elements
}

func generateIsometricExtrusions(
	weeks []ContributionWeek,
	cols, rows int,
	layout isometricLayout,
) []svg.Element {
	type cubePos struct {
		Col   int
		Row   int
		BaseX float64
		BaseY float64
		Day   ContributionDay
	}

	// Collect cubes and sort back-to-front by their screen-space base position.
	// This keeps roofs visible and avoids later tiles overpainting earlier tops.
	cubes := make([]cubePos, 0)
	for col := 0; col < cols; col++ {
		if col >= len(weeks) {
			break
		}
		for row := 0; row < rows; row++ {
			if row >= len(weeks[col].ContributionDays) {
				continue
			}
			day := weeks[col].ContributionDays[row]
			if contributionLevelFromAPIColour(day.Color) <= 0 {
				continue
			}
			x := layout.OriginX + (float64(col+row) * layout.TileW / 2.0)
			y := layout.OriginY + (float64(col) * layout.TileH / 2.0) - (float64(row) * layout.TileH / 2.0)
			cubes = append(
				cubes,
				cubePos{Col: col, Row: row, BaseX: x, BaseY: y + layout.TileH, Day: day},
			)
		}
	}
	sort.Slice(cubes, func(i, j int) bool {
		if cubes[i].BaseY == cubes[j].BaseY {
			return cubes[i].BaseX < cubes[j].BaseX
		}
		return cubes[i].BaseY < cubes[j].BaseY
	})

	elements := make([]svg.Element, 0, len(cubes)*6)
	for _, c := range cubes {
		maybeAppendExtrusion(&elements, weeks, cols, rows, c.Day, c.Col, c.Row, layout)
	}
	return elements
}

func contributionLevelAt(weeks []ContributionWeek, col, row int) int {
	if col < 0 || col >= len(weeks) {
		return 0
	}
	if row < 0 || row >= len(weeks[col].ContributionDays) {
		return 0
	}
	return contributionLevelFromAPIColour(weeks[col].ContributionDays[row].Color)
}

func maybeAppendExtrusion(
	elements *[]svg.Element,
	weeks []ContributionWeek,
	cols, rows int,
	day ContributionDay,
	col, row int,
	layout isometricLayout,
) {
	level := contributionLevelFromAPIColour(day.Color)
	if level <= 0 {
		return
	}

	// Occlusion (requested): blocks to the right.
	// If the neighbour "to the right" is equal-or-higher, hide the right face so
	// equal-level runs read as flat.
	drawRightFace := true
	if col+1 < cols {
		if contributionLevelAt(weeks, col+1, row) >= level {
			drawRightFace = false
		}
	}

	height := float64(level) * layout.HeightStep
	if height > layout.MaxHeight {
		height = layout.MaxHeight
	}

	baseColour := currentColourProfile.GetContributionColour(day.Color)
	// Single-colour cubes (no per-face shading)
	leftColour := baseColour
	rightColour := baseColour
	topColour := baseColour

	x := layout.OriginX + (float64(col+row) * layout.TileW / 2.0)
	y := layout.OriginY + (float64(col) * layout.TileH / 2.0) - (float64(row) * layout.TileH / 2.0)

	base := diamondPoints(x, y, layout.TileW, layout.TileH)
	top := diamondPoints(x, y-height, layout.TileW, layout.TileH)

	leftFace := []point{top[3], top[2], base[2], base[3]}
	rightFace := []point{top[1], top[2], base[2], base[1]}

	*elements = append(*elements,
		svg.Polygon().
			Points(toSVGPoints(leftFace)).
			Fill(svg.String(leftColour)),
	)
	if drawRightFace {
		*elements = append(*elements,
			svg.Polygon().
				Points(toSVGPoints(rightFace)).
				Fill(svg.String(rightColour)),
		)
	}
	*elements = append(*elements,
		svg.Polygon().
			Points(toSVGPoints(top)).
			Fill(svg.String(topColour)),
	)

	// Explicit outline (avoids tops reading like triangles)
	stroke := svg.String(currentColourProfile.Background)
	sw := svg.Px(layout.StrokeWidth)

	// Top diamond outline (always)
	*elements = append(*elements,
		svg.Line().X1Y1X2Y2(top[0].X, top[0].Y, top[1].X, top[1].Y).Stroke(stroke).StrokeWidth(sw),
		svg.Line().X1Y1X2Y2(top[1].X, top[1].Y, top[2].X, top[2].Y).Stroke(stroke).StrokeWidth(sw),
		svg.Line().X1Y1X2Y2(top[2].X, top[2].Y, top[3].X, top[3].Y).Stroke(stroke).StrokeWidth(sw),
		svg.Line().X1Y1X2Y2(top[3].X, top[3].Y, top[0].X, top[0].Y).Stroke(stroke).StrokeWidth(sw),
	)

	// Left/front outline edges (always visible in this projection)
	*elements = append(
		*elements,
		svg.Line().
			X1Y1X2Y2(top[3].X, top[3].Y, base[3].X, base[3].Y).
			Stroke(stroke).
			StrokeWidth(sw),
		svg.Line().
			X1Y1X2Y2(top[2].X, top[2].Y, base[2].X, base[2].Y).
			Stroke(stroke).
			StrokeWidth(sw),
		svg.Line().
			X1Y1X2Y2(base[3].X, base[3].Y, base[2].X, base[2].Y).
			Stroke(stroke).
			StrokeWidth(sw),
	)

	// Right outline edges only if the right face is exposed (blocked-to-the-right rule)
	if drawRightFace {
		*elements = append(
			*elements,
			svg.Line().
				X1Y1X2Y2(top[1].X, top[1].Y, base[1].X, base[1].Y).
				Stroke(stroke).
				StrokeWidth(sw),
			svg.Line().
				X1Y1X2Y2(base[1].X, base[1].Y, base[2].X, base[2].Y).
				Stroke(stroke).
				StrokeWidth(sw),
		)
	}
}

func generateContributionCalendarNotes(
	stats contributionCalendarStats,
	notesX, startY, lineGap float64,
) []svg.Element {
	y := startY
	return []svg.Element{
		svg.Text(svg.CharData("📌 Commits streaks")).
			XY(notesX, y, svg.Px).
			Fill(svg.String(currentColourProfile.AccentPrimary)).
			Style(svg.String(fontStyleHeader15px)),
		svg.Text(svg.CharData(fmt.Sprintf("🔥 Current streak %d days", stats.CurrentStreakDays))).
			XY(notesX, y+lineGap*1, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(svg.String(fontStyle13px)),
		svg.Text(svg.CharData(fmt.Sprintf("✨ Best streak %d days", stats.BestStreakDays))).
			XY(notesX, y+lineGap*2, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(svg.String(fontStyle13px)),
		svg.Text(svg.CharData("📈 Commits per day")).
			XY(notesX, y+lineGap*4, svg.Px).
			Fill(svg.String(currentColourProfile.AccentPrimary)).
			Style(svg.String(fontStyleHeader15px)),
		svg.Text(svg.CharData(fmt.Sprintf("🏆 Highest in a day %d", stats.HighestInDay))).
			XY(notesX, y+lineGap*5, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(svg.String(fontStyle13px)),
		svg.Text(svg.CharData(fmt.Sprintf("📊 Average per day ~%.2f", stats.AveragePerDay))).
			XY(notesX, y+lineGap*6, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(svg.String(fontStyle13px)),
	}
}

func collectContributionCounts(
	contributionCalendar *ContributionCalendar,
) ([]time.Time, map[time.Time]int, int) {
	counts := map[time.Time]int{}
	dates := make([]time.Time, 0)
	maxInDay := 0
	for _, week := range contributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			dayDate, err := time.Parse("2006-01-02", day.Date)
			if err != nil {
				continue
			}
			dayDate = time.Date(
				dayDate.Year(),
				dayDate.Month(),
				dayDate.Day(),
				0,
				0,
				0,
				0,
				time.UTC,
			)
			if _, exists := counts[dayDate]; !exists {
				dates = append(dates, dayDate)
			}
			counts[dayDate] = day.ContributionCount
			if day.ContributionCount > maxInDay {
				maxInDay = day.ContributionCount
			}
		}
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates, counts, maxInDay
}

func calculateCurrentStreak(dates []time.Time, counts map[time.Time]int) int {
	if len(dates) == 0 {
		return 0
	}
	streak := 0
	for d := dates[len(dates)-1]; ; d = d.AddDate(0, 0, -1) {
		count, ok := counts[d]
		if !ok || count <= 0 {
			break
		}
		streak++
	}
	return streak
}

func calculateBestStreak(dates []time.Time, counts map[time.Time]int) int {
	if len(dates) == 0 {
		return 0
	}
	best := 0
	running := 0
	prev := dates[0]
	for i, d := range dates {
		if i > 0 {
			expected := prev.AddDate(0, 0, 1)
			if !d.Equal(expected) {
				running = 0
			}
		}
		if counts[d] > 0 {
			running++
			if running > best {
				best = running
			}
		} else {
			running = 0
		}
		prev = d
	}
	return best
}

func calculateAveragePerDay(total, days int) float64 {
	if days <= 0 {
		return 0
	}
	return float64(total) / float64(days)
}

// Generate profile section of svg
func generateProfileSection(userInfo *GitHubUserInfo) svg.Element {
	yearsAgo := time.Since(userInfo.JoinedGitHub).Hours() / 24 / 365

	// Point to hosted avatar so GitHub README sanitization keeps the image
	avatarURL := normalizeAvatarURL(userInfo.AvatarURL)
	avatarImage := svg.Image().
		Href(svg.String(avatarURL)).
		Width(svg.Px(24)).Height(svg.Px(24)).
		X(svg.Px(18)).Y(svg.Px(28)).
		Class(svg.String("avatar"))
	if avatarImage.Attrs == nil {
		avatarImage.Attrs = map[string]svg.AttrValue{}
	}
	avatarImage.Attrs["xlink:href"] = svg.String(avatarURL)

	return svg.G().AppendChildren(
		// Use <image> element with both href variants so sanitizers retain the avatar
		avatarImage,

		// Name - positioned next to avatar
		svg.Text(svg.CharData(userInfo.Name)).
			XY(50, 45, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 18px; font-weight: 600;")),

		// Joined info
		svg.Text(svg.CharData(fmt.Sprintf("⏰ Joined GitHub %.0f years ago", yearsAgo))).
			XY(20, 70, svg.Px).
			Fill(svg.String(currentColourProfile.TextSecondary)).
			Style(svg.String(fontStyle13px)),

		// Followed by
		svg.Text(svg.CharData(fmt.Sprintf("👥 Followed by %d users", userInfo.Followers))).
			XY(20, 88, svg.Px).
			Fill(svg.String(currentColourProfile.TextSecondary)).
			Style(svg.String(fontStyle13px)),
	)
}

// Generate stats row of svg
func generateStatsRow(
	userInfo *GitHubUserInfo,
	githubTotalsStats *GitHubTotalsStats,
	contributionCalendar *ContributionCalendar,
) svg.Element {
	activityStatsX := 20.0
	communityStatsX := 250.0
	repositoriesStatsX := 480.0
	headersRowY := 115.0
	row1Y := 133.0
	row2Y := row1Y + 16.0
	row3Y := row2Y + 16.0
	row4Y := row3Y + 16.0
	headerStyle := svg.String(fontStyleHeader15px)
	textStyle := svg.String(fontStyle13px)
	return svg.G().AppendChildren(
		// Activity stats section
		svg.Text(svg.CharData("📈 Activity")).
			XY(activityStatsX, headersRowY, svg.Px).
			Fill(svg.String(currentColourProfile.AccentPrimary)).
			Style(headerStyle),

		svg.Text(svg.CharData(fmt.Sprintf("💻 %d Commits", githubTotalsStats.TotalCommits))).
			XY(activityStatsX, row1Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("📋 %d Pull requests reviewed", githubTotalsStats.TotalPullRequestReviews))).
			XY(activityStatsX, row2Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("🔀 %d Pull requests opened", githubTotalsStats.TotalPullRequests))).
			XY(activityStatsX, row3Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("❗ %d Issues opened", githubTotalsStats.TotalIssues))).
			XY(activityStatsX, row4Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),

		// Community stats section
		svg.Text(svg.CharData("👥 Community stats")).
			XY(communityStatsX, headersRowY, svg.Px).
			Fill(svg.String(currentColourProfile.AccentPrimary)).
			Style(headerStyle),

		svg.Text(svg.CharData(fmt.Sprintf("🏢 Member of %d organizations", githubTotalsStats.TotalMemberOfOrganizations))).
			XY(communityStatsX, row1Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("👤 Following %d users", userInfo.Following))).
			XY(communityStatsX, row2Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("⭐ Starred %d repositories", githubTotalsStats.TotalStarredRepos))).
			XY(communityStatsX, row3Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData("👀 Watching 42 repositories")).
			XY(communityStatsX, row4Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),

		// Repository stats
		svg.Text(svg.CharData("📚 56 Repositories")).
			XY(repositoriesStatsX, headersRowY, svg.Px).
			Fill(svg.String(currentColourProfile.AccentPrimary)).
			Style(headerStyle),

		svg.Text(svg.CharData(fmt.Sprintf("💖 %d Sponsors", githubTotalsStats.TotalSponsors))).
			XY(repositoriesStatsX, row1Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData("⭐ 9 Stargazers")).
			XY(repositoriesStatsX, row2Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData("🍴 9 Forkers")).
			XY(repositoriesStatsX, row3Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),
		svg.Text(svg.CharData(fmt.Sprintf("👁️ %d Watchers", githubTotalsStats.TotalWatchers))).
			XY(repositoriesStatsX, row4Y, svg.Px).
			Fill(svg.String(currentColourProfile.TextPrimary)).
			Style(textStyle),

		// Contribution graph
		generateContributionGraph(headerStyle, textStyle, contributionCalendar),
	)
}

func generateContributionGraph(
	headerStyle, textStyle svg.String,
	contributionCalendar *ContributionCalendar,
) svg.Element {
	// Create a contribution graph showing the current month in rows of 7 days
	squares := generateMonthContributionSquares(contributionCalendar)

	// Add contribution graph header
	headerElements := []svg.Element{
		svg.Text(svg.CharData("📚 Contributions")).
			XY(630, 115, svg.Px).
			Fill(svg.String(currentColourProfile.AccentPrimary)).
			Style(headerStyle),
		svg.Text(svg.CharData(fmt.Sprintf("%d contributions in the last year", contributionCalendar.TotalContributions))).
			XY(630, 210, svg.Px).
			Fill(svg.String(currentColourProfile.TextSecondary)).
			Style(svg.String(fontStyle13px)),
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

		// Get colour for this day if we have data
		colour := currentColourProfile.ContributionLevel0 // Default: no contributions
		if dayIndex < len(monthContributions) && monthContributions[dayIndex].Color != "" {
			// Map the GitHub API colour to the current colour profile
			colour = currentColourProfile.GetContributionColour(monthContributions[dayIndex].Color)
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

func getMonthContributions(
	contributionCalendar *ContributionCalendar,
	year int,
	month time.Month,
) []ContributionDay {
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
			Fill(svg.String(currentColourProfile.AccentPrimary)).
			Style(svg.String(fontStyleHeader15px)),
		svg.Text(svg.CharData("Most used languages")).
			XY(400, 240, svg.Px).
			Fill(svg.String(currentColourProfile.AccentPrimary)).
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
			Fill(svg.String(currentColourProfile.TextPrimary)).
			TextAnchor(svg.String("middle")).
			Style(svg.String("font-family: -apple-system, BlinkMacSystemFont, Segoe UI; font-size: 12px;")),
		)
	}

	return svg.G().AppendChildren(elements...)
}
