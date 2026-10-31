package main

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCheckGitHubRateLimits(t *testing.T) {
	for _, test := range []struct {
		name      string
		status    int
		body      string
		transport error
		wantError string
	}{
		{
			name: "available quotas", status: http.StatusOK,
			body: `{"resources":{"core":{"remaining":100},"graphql":{"remaining":100}}}`,
		},
		{
			name: "positive quotas with unrelated exhausted resource", status: http.StatusOK,
			body: `{"resources":{"core":{"remaining":1},"graphql":{"remaining":1},"search":{"remaining":0}}}`,
		},
		{
			name: "REST exhausted", status: http.StatusOK,
			body:      `{"resources":{"core":{"remaining":0,"reset":1793448000},"graphql":{"remaining":100}}}`,
			wantError: "GitHub core rate limit exhausted; retry after 2026-10-31T12:00:00Z",
		},
		{
			name: "GraphQL exhausted", status: http.StatusOK,
			body:      `{"resources":{"core":{"remaining":100},"graphql":{"remaining":0,"reset":1793448000}}}`,
			wantError: "GitHub graphql rate limit exhausted; retry after 2026-10-31T12:00:00Z",
		},
		{
			name: "no reset time", status: http.StatusOK,
			body:      `{"resources":{"core":{"remaining":0},"graphql":{"remaining":100}}}`,
			wantError: "GitHub core rate limit exhausted; retry when the quota resets",
		},
		{
			name: "missing core", status: http.StatusOK,
			body:      `{"resources":{"graphql":{"remaining":100}}}`,
			wantError: "missing core remaining quota",
		},
		{
			name: "missing graphql", status: http.StatusOK,
			body:      `{"resources":{"core":{"remaining":100}}}`,
			wantError: "missing graphql remaining quota",
		},
		{
			name: "missing remaining", status: http.StatusOK,
			body:      `{"resources":{"core":{},"graphql":{"remaining":100}}}`,
			wantError: "missing core remaining quota",
		},
		{
			name: "null remaining", status: http.StatusOK,
			body:      `{"resources":{"core":{"remaining":100},"graphql":{"remaining":null}}}`,
			wantError: "missing graphql remaining quota",
		},
		{name: "invalid JSON", status: http.StatusOK, body: `{`, wantError: "failed to decode"},
		{name: "empty response", status: http.StatusOK, wantError: "failed to decode"},
		{name: "unauthorized", status: http.StatusUnauthorized, wantError: "returned status 401"},
		{name: "rate limited", status: http.StatusForbidden, wantError: "returned status 403"},
		{name: "server error", status: http.StatusInternalServerError, wantError: "returned status 500"},
		{name: "network error", transport: errors.New("connection failed"), wantError: "connection failed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("INPUT_GITHUB_TOKEN", "metrics-token")
			t.Setenv("INPUT_WORKFLOW_GITHUB_TOKEN", "different-workflow-token")
			previousTransport := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = previousTransport })
			http.DefaultTransport = contributionRoundTripFunc(
				func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodGet ||
						request.URL.String() != "https://api.github.com/rate_limit" {
						t.Errorf(
							"unexpected rate limit request: %s %s",
							request.Method,
							request.URL,
						)
					}
					if got := request.Header.Get("Authorization"); got != "Bearer metrics-token" {
						t.Errorf("unexpected authorization header: %q", got)
					}
					if got := request.Header.Get("Accept"); got != "application/vnd.github+json" {
						t.Errorf("unexpected Accept header: %q", got)
					}
					if test.transport != nil {
						return nil, test.transport
					}
					return &http.Response{
						StatusCode: test.status,
						Body:       io.NopCloser(strings.NewReader(test.body)),
					}, nil
				},
			)

			err := checkGitHubRateLimits()
			if test.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("expected error containing %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestMainStopsBeforeGenerationWhenRateLimitCheckFails(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	requests := 0
	http.DefaultTransport = contributionRoundTripFunc(
		func(request *http.Request) (*http.Response, error) {
			requests++
			if requests != 1 || request.URL.Path != "/rate_limit" {
				t.Fatalf("metrics or commit request made after failed preflight: %s", request.URL)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(
					`{"resources":{"core":{"remaining":100},"graphql":{"remaining":0}}}`,
				)),
			}, nil
		},
	)
	// Turn fatal logging into a panic so the startup failure is observable in-process.
	t.Cleanup(
		zap.ReplaceGlobals(zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))),
	)
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Error("expected startup to fail before generating an SVG")
		}
		if requests != 1 {
			t.Errorf("expected only the rate limit request, got %d requests", requests)
		}
	}()
	main()
}
