package main

import (
	"strings"
	"time"

	"go.uber.org/zap"
)

// GitHub default contribution colours (used as baseline for profiles)
const (
	githubContribNone       = "#ebedf0"
	githubContribLow        = "#9be9a8"
	githubContribMediumLow  = "#40c463"
	githubContribMediumHigh = "#30a14e"
	githubContribHigh       = "#216e39"
)

// ColourProfile defines the colour scheme for the SVG
type ColourProfile struct {
	Name               string
	Background         string
	TextPrimary        string
	TextSecondary      string
	AccentPrimary      string
	AccentSecondary    string
	ContributionLevel0 string // No contributions
	ContributionLevel1 string // Low contributions
	ContributionLevel2 string // Medium-low contributions
	ContributionLevel3 string // Medium-high contributions
	ContributionLevel4 string // High contributions
}

// Available colour profiles
var colourProfiles = map[string]ColourProfile{
	"default": {
		Name:               "Default",
		Background:         "#ffffff",
		TextPrimary:        "#24292f",
		TextSecondary:      "#656d76",
		AccentPrimary:      "#0969da",
		AccentSecondary:    "#1f883d",
		ContributionLevel0: githubContribNone,
		ContributionLevel1: githubContribLow,
		ContributionLevel2: githubContribMediumLow,
		ContributionLevel3: githubContribMediumHigh,
		ContributionLevel4: githubContribHigh,
	},
	"dark": {
		Name:               "Dark",
		Background:         "#0d1117",
		TextPrimary:        "#e6edf3",
		TextSecondary:      "#7d8590",
		AccentPrimary:      "#58a6ff",
		AccentSecondary:    "#3fb950",
		ContributionLevel0: "#161b22",
		ContributionLevel1: "#0e4429",
		ContributionLevel2: "#006d32",
		ContributionLevel3: "#26a641",
		ContributionLevel4: "#39d353",
	},
	"ocean": {
		Name:               "Ocean",
		Background:         "#f0f8ff",
		TextPrimary:        "#0f1419",
		TextSecondary:      "#5c6773",
		AccentPrimary:      "#0077be",
		AccentSecondary:    "#00b4d8",
		ContributionLevel0: "#e6f3ff",
		ContributionLevel1: "#90e0ef",
		ContributionLevel2: "#00b4d8",
		ContributionLevel3: "#0096c7",
		ContributionLevel4: "#023e8a",
	},
	"sunset": {
		Name:               "Sunset",
		Background:         "#fff5f5",
		TextPrimary:        "#2d1b1b",
		TextSecondary:      "#6b5555",
		AccentPrimary:      "#d62828",
		AccentSecondary:    "#f77f00",
		ContributionLevel0: "#ffe6e6",
		ContributionLevel1: "#ffb3ba",
		ContributionLevel2: "#ff8fa3",
		ContributionLevel3: "#ff5c8a",
		ContributionLevel4: "#d62828",
	},
}

// GetColourProfile returns the requested profile with seasonal contribution colours
// for today's UTC date, or the default profile if the name is not found.
func GetColourProfile(name string) ColourProfile {
	return getColourProfileAt(name, time.Now())
}

func getColourProfileAt(name string, now time.Time) ColourProfile {
	// Normalize the name to lowercase for case-insensitive matching
	normalizedName := strings.ToLower(strings.TrimSpace(name))

	profile, exists := colourProfiles[normalizedName]
	if exists {
		zap.L().Info("Using colour profile", zap.String("profile", profile.Name))
	} else {
		zap.L().Warn("Colour profile not found, using default",
			zap.String("requested", name),
			zap.String("using", "default"))
		normalizedName = "default"
		profile = colourProfiles["default"]
	}

	// Match GitHub's Halloween palette for the default (light) and dark profiles
	// across the whole graph. Other profiles keep their contribution colours.
	utc := now.UTC()
	if (normalizedName == "default" || normalizedName == "dark") &&
		utc.Month() == time.October && utc.Day() == 31 {
		profile.ContributionLevel1 = "#fac68f"
		profile.ContributionLevel2 = "#c46212"
		profile.ContributionLevel3 = "#984b10"
		profile.ContributionLevel4 = "#e3d04f"
	}

	return profile
}

// GetAvailableProfiles returns a list of all available colour profile names
func GetAvailableProfiles() []string {
	profiles := make([]string, 0, len(colourProfiles))
	for name := range colourProfiles {
		profiles = append(profiles, name)
	}
	return profiles
}

// GetContributionColour maps contribution intensity to the selected profile.
// API hex colours can change with GitHub's seasonal palettes.
func (cp ColourProfile) GetContributionColour(level int) string {
	switch level {
	case 1:
		return cp.ContributionLevel1
	case 2:
		return cp.ContributionLevel2
	case 3:
		return cp.ContributionLevel3
	case 4:
		return cp.ContributionLevel4
	default:
		return cp.ContributionLevel0
	}
}
