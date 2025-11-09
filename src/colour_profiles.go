package main

import (
	"strings"

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

// GetColourProfile returns the colour profile for the given name, or the default profile if not found
func GetColourProfile(name string) ColourProfile {
	// Normalize the name to lowercase for case-insensitive matching
	normalizedName := strings.ToLower(strings.TrimSpace(name))

	if profile, exists := colourProfiles[normalizedName]; exists {
		zap.L().Info("Using colour profile", zap.String("profile", profile.Name))
		return profile
	}

	// If profile not found, log warning and return default
	zap.L().Warn("Colour profile not found, using default",
		zap.String("requested", name),
		zap.String("using", "default"))
	return colourProfiles["default"]
}

// GetAvailableProfiles returns a list of all available colour profile names
func GetAvailableProfiles() []string {
	profiles := make([]string, 0, len(colourProfiles))
	for name := range colourProfiles {
		profiles = append(profiles, name)
	}
	return profiles
}

// GetContributionColour returns the appropriate contribution colour based on the contribution level
// from the GitHub API colour string
func (cp ColourProfile) GetContributionColour(apiColour string) string {
	// GitHub API returns colours like:
	// #ebedf0 (no contributions)
	// #9be9a8 (low)
	// #40c463 (medium-low)
	// #30a14e (medium-high)
	// #216e39 (high)

	switch apiColour {
	case githubContribNone:
		return cp.ContributionLevel0
	case githubContribLow:
		return cp.ContributionLevel1
	case githubContribMediumLow:
		return cp.ContributionLevel2
	case githubContribMediumHigh:
		return cp.ContributionLevel3
	case githubContribHigh:
		return cp.ContributionLevel4
	default:
		// If we get an unexpected colour, try to map it intelligently
		// or just return level 0 (no contributions)
		return cp.ContributionLevel0
	}
}
