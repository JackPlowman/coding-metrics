package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
