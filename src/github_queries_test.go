package main

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNormalizeAvatarURLAddsSizeParameter(t *testing.T) {
	got := normalizeAvatarURL("https://avatars.githubusercontent.com/u/12345?v=4")

	if !strings.Contains(got, "s=80") {
		t.Fatalf("expected normalized avatar URL to include s=80, got %q", got)
	}
	if !strings.Contains(got, "v=4") {
		t.Fatalf("expected normalized avatar URL to preserve existing query values, got %q", got)
	}
}

func TestNormalizeAvatarURLKeepsExistingSize(t *testing.T) {
	got := normalizeAvatarURL("https://avatars.githubusercontent.com/u/12345?s=40")

	if strings.Contains(got, "s=80") {
		t.Fatalf("expected normalized avatar URL to preserve explicit size, got %q", got)
	}
}

func TestFetchAvatarDataURIEmbedsImage(t *testing.T) {
	avatarBytes := []byte("fake-png-bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "image/*" {
			t.Fatalf("expected image Accept header, got %q", got)
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(avatarBytes)
	}))
	t.Cleanup(server.Close)

	got := fetchAvatarDataURI(server.Client(), server.URL)
	want := "data:image/png;base64," + base64.StdEncoding.EncodeToString(avatarBytes)

	if got != want {
		t.Fatalf("expected embedded avatar data URI %q, got %q", want, got)
	}
}

func TestFetchAvatarDataURIRejectsNonImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"nope"}`))
	}))
	t.Cleanup(server.Close)

	if got := fetchAvatarDataURI(server.Client(), server.URL); got != "" {
		t.Fatalf("expected non-image response to be rejected, got %q", got)
	}
}

func TestMergeContributionCalendarsDeduplicatesAndFillsMissingDays(t *testing.T) {
	from := time.Date(2026, time.July, 12, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 7)
	left := &ContributionCalendar{Weeks: []ContributionWeek{{ContributionDays: []ContributionDay{
		{Date: "2026-07-12", ContributionCount: 2, Color: githubContribLow},
		{Date: "2026-07-13", ContributionCount: 3, Color: githubContribMediumLow},
	}}}}
	right := &ContributionCalendar{Weeks: []ContributionWeek{{ContributionDays: []ContributionDay{
		{Date: "2026-07-13", ContributionCount: 4, Color: githubContribMediumHigh},
		{Date: "2026-07-19", ContributionCount: 5, Color: githubContribHigh},
	}}}}

	got := mergeContributionCalendars(from, to, left, right)

	if got.TotalContributions != 11 {
		t.Fatalf("expected deduplicated total of 11, got %d", got.TotalContributions)
	}
	if len(got.Weeks) != 2 ||
		len(got.Weeks[0].ContributionDays) != 7 ||
		len(got.Weeks[1].ContributionDays) != 1 {
		t.Fatalf("expected one full and one partial week, got %#v", got.Weeks)
	}
	if got.Weeks[0].ContributionDays[1].ContributionCount != 4 {
		t.Fatalf("expected later duplicate to win, got %#v", got.Weeks[0].ContributionDays[1])
	}
	if got.Weeks[0].ContributionDays[2].Color != githubContribNone {
		t.Fatalf(
			"expected missing day to use the no-contribution colour, got %#v",
			got.Weeks[0].ContributionDays[2],
		)
	}
}

func TestIsGraphQLResourceLimitError(t *testing.T) {
	if !isGraphQLResourceLimitError(errors.New("RESOURCE LIMITS for this query exceeded")) {
		t.Fatal("expected resource-limit error to be recognized")
	}
	if isGraphQLResourceLimitError(errors.New("bad credentials")) {
		t.Fatal("did not expect authentication error to be recognized as a resource limit")
	}
}
