package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
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
		{Date: "2026-07-12", ContributionCount: 2, ContributionLevel: "FIRST_QUARTILE"},
		{Date: "2026-07-13", ContributionCount: 3, ContributionLevel: "SECOND_QUARTILE"},
	}}}}
	right := &ContributionCalendar{Weeks: []ContributionWeek{{ContributionDays: []ContributionDay{
		{Date: "2026-07-13", ContributionCount: 4, ContributionLevel: "THIRD_QUARTILE"},
		{Date: "2026-07-19", ContributionCount: 5, ContributionLevel: "FOURTH_QUARTILE"},
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
	if got.Weeks[0].ContributionDays[2].ContributionLevel != "NONE" {
		t.Fatalf(
			"expected missing day to use the no-contribution level, got %#v",
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

type contributionRoundTripFunc func(*http.Request) (*http.Response, error)

func (f contributionRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestQueryContributionCalendarRequestsLevels(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	http.DefaultTransport = contributionRoundTripFunc(
		func(request *http.Request) (*http.Response, error) {
			var body GitHubGraphQLRequest
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(body.Query, "contributionLevel") {
				t.Error(
					"calendar query must request contribution levels independently of API colours",
				)
			}
			if body.Variables["from"] != "2026-10-31T00:00:00Z" ||
				body.Variables["to"] != "2026-10-31T23:59:59Z" {
				t.Errorf("unexpected calendar range: %v", body.Variables)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{"data":{"user":{"contributionsCollection":{
				"contributionCalendar":{"totalContributions":12,"weeks":[{"contributionDays":[
					{"date":"2026-10-31","contributionCount":12,"contributionLevel":"FOURTH_QUARTILE"}
				]}]}
			}}}}`)),
			}, nil
		},
	)

	date := time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC)
	calendar, err := queryContributionCalendar("example", &date, &date)
	if err != nil {
		t.Fatal(err)
	}
	if calendar.TotalContributions != 12 || len(calendar.Weeks) != 1 ||
		len(calendar.Weeks[0].ContributionDays) != 1 {
		t.Fatalf("unexpected contribution calendar: %#v", calendar)
	}
	day := calendar.Weeks[0].ContributionDays[0]
	if day.Date != "2026-10-31" || day.ContributionCount != 12 ||
		day.ContributionLevel != "FOURTH_QUARTILE" {
		t.Fatalf("calendar query did not preserve contribution data: %#v", day)
	}
}
