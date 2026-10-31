package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/twpayne/go-svg"
)

func TestContributionGraphsIgnoreAPIColourPalette(t *testing.T) {
	previousProfile := currentColourProfile
	t.Cleanup(func() { currentColourProfile = previousProfile })
	now := time.Now()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Seasonal and future API colours must not change contribution intensity.
	var days []ContributionDay
	fixture := fmt.Sprintf(`[
		{"date":%q,"contributionCount":0,"contributionLevel":"NONE","color":"#ebedf0"},
		{"date":%q,"contributionCount":1,"contributionLevel":"FIRST_QUARTILE","color":"#ffee4a"},
		{"date":%q,"contributionCount":5,"contributionLevel":"SECOND_QUARTILE","color":"#ffc501"},
		{"date":%q,"contributionCount":10,"contributionLevel":"THIRD_QUARTILE","color":"#fe9600"},
		{"date":%q,"contributionCount":30,"contributionLevel":"FOURTH_QUARTILE","color":"#03001c"}
	]`, month.Format("2006-01-02"), month.AddDate(0, 0, 1).Format("2006-01-02"),
		month.AddDate(0, 0, 2).Format("2006-01-02"), month.AddDate(0, 0, 3).Format("2006-01-02"),
		month.AddDate(0, 0, 4).Format("2006-01-02"))
	if err := json.Unmarshal([]byte(fixture), &days); err != nil {
		t.Fatal(err)
	}
	calendar := mergeContributionCalendars(month, month.AddDate(0, 0, 6),
		&ContributionCalendar{Weeks: []ContributionWeek{{ContributionDays: days}}})
	if calendar.TotalContributions != 46 {
		t.Fatalf(
			"expected contribution counts to survive merging, got %d",
			calendar.TotalContributions,
		)
	}

	profiles := make(map[string]ColourProfile, len(colourProfiles)*2)
	for name, profile := range colourProfiles {
		profiles[name] = profile
		profiles[name+"/halloween"] = getColourProfileAt(name,
			time.Date(2026, time.October, 31, 12, 0, 0, 0, time.UTC))
	}
	for name, profile := range profiles {
		t.Run(name, func(t *testing.T) {
			currentColourProfile = profile
			colours := []string{
				profile.ContributionLevel0, profile.ContributionLevel1,
				profile.ContributionLevel2, profile.ContributionLevel3, profile.ContributionLevel4,
				profile.ContributionLevel0, profile.ContributionLevel0,
			}
			squares := generateMonthContributionSquares(calendar, 700)
			baseTiles := generateIsometricBaseTiles(calendar.Weeks, 1, 7, 0, 0, 12, 6)
			layout := isometricLayout{
				TileW:       12,
				TileH:       6,
				HeightStep:  4,
				MaxHeight:   16,
				StrokeWidth: 0.6,
			}
			for i, want := range colours {
				if got := squares[i].(*svg.RectElement).Attrs["fill"]; got != svg.String(want) {
					t.Errorf("month day %d: expected fill %s, got %v", i+1, want, got)
				}
				if got := baseTiles[i].(*svg.PolygonElement).Attrs["fill"]; got != svg.String(
					want,
				) {
					t.Errorf("year day %d: expected fill %s, got %v", i+1, want, got)
				}
				extrusions := generateIsometricExtrusions(
					[]ContributionWeek{
						{
							ContributionDays: []ContributionDay{
								calendar.Weeks[0].ContributionDays[i],
							},
						},
					},
					1,
					1,
					layout,
				)
				if i >= 1 && i <= 4 {
					if len(extrusions) == 0 {
						t.Errorf("year day %d: positive contributions produced no cube", i+1)
					} else if got := extrusions[0].(*svg.PolygonElement).Attrs["fill"]; got != svg.String(want) {
						t.Errorf("year day %d: expected cube fill %s, got %v", i+1, want, got)
					}
				} else if len(extrusions) != 0 {
					t.Errorf("year day %d: empty day produced a cube", i+1)
				}
			}
		})
	}
}
