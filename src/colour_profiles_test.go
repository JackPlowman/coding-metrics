package main

import (
	"testing"
	"time"
)

func TestSeasonalContributionColours(t *testing.T) {
	tests := []struct {
		date      string
		halloween bool
	}{
		{"2026-10-30T23:59:59Z", false},
		{"2026-10-31T00:00:00Z", true},
		{"2026-10-31T23:59:59Z", true},
		{"2026-11-01T00:00:00Z", false},
		{"2025-10-31T12:00:00Z", true},
		{"2027-10-31T12:00:00Z", true},
		{"2026-09-30T12:00:00Z", false},
		{"2026-12-31T12:00:00Z", false},
		{"2026-10-31T00:30:00+02:00", false},
		{"2026-10-30T23:30:00-04:00", true},
		{"2026-11-01T00:30:00+02:00", true},
		{"2026-10-31T23:30:00-04:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			date, err := time.Parse(time.RFC3339, tt.date)
			if err != nil {
				t.Fatal(err)
			}
			for name, original := range colourProfiles {
				want := original
				if tt.halloween {
					want.ContributionLevel1 = "#fac68f"
					want.ContributionLevel2 = "#c46212"
					want.ContributionLevel3 = "#984b10"
					want.ContributionLevel4 = "#e3d04f"
				}
				if got := getColourProfileAt(name, date); got != want {
					t.Errorf("%s: expected profile %+v, got %+v", name, want, got)
				}
				if colourProfiles[name] != original {
					t.Errorf("%s: seasonal colours modified the stored profile", name)
				}
			}
		})
	}
}

func TestSeasonalProfileNameResolution(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2026, time.October, 31, 12, 0, 0, 0, time.UTC),
		time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC),
	} {
		for _, tt := range []struct{ name, resolved string }{
			{" DARK ", "dark"},
			{"unknown", "default"},
			{"", "default"},
		} {
			got := getColourProfileAt(tt.name, date)
			want := getColourProfileAt(tt.resolved, date)
			if got != want {
				t.Errorf("%q on %s: expected %+v, got %+v", tt.name, date, want, got)
			}
		}
	}
}
