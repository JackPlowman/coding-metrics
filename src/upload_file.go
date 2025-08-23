package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// uploadSVGChanges copies the generated SVG into the repository and commits it
// via the GitHub REST API to the path `output.svg`. It is defensive and will
// log-and-return if required environment variables or a proper file are not
// present.
func uploadSVGChanges(file *os.File) {
	ownerRepo := os.Getenv("INPUT_REPOSITORY") // "owner/repo"
	token := os.Getenv("INPUT_GITHUB_TOKEN")
	pushToBranch := os.Getenv("INPUT_BRANCH")

	logger := zap.L()

	if ownerRepo == "" || token == "" {
		logger.Info("INPUT_REPOSITORY or INPUT_GITHUB_TOKEN not set: skipping upload")
		return
	}

	srcPath := ""
	if file != nil {
		srcPath = file.Name()
	}
	if srcPath == "" || srcPath == "-" {
		logger.Info("SVG written to stdout or unnamed file; skipping upload")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logger.Error("failed to seek svg file", zap.Error(err))
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		logger.Error("failed to read svg file", zap.Error(err))
		return
	}

	repoPath := "output.svg"
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", ownerRepo, repoPath)
	client := &http.Client{Timeout: 15 * time.Second}

	// Try to fetch existing file metadata to include sha if updating
	existingSHA := ""
	reqGet, _ := http.NewRequestWithContext(context.Background(), "GET", apiURL, nil)
	if pushToBranch != "" {
		q := reqGet.URL.Query()
		q.Set("ref", pushToBranch)
		reqGet.URL.RawQuery = q.Encode()
	}
	reqGet.Header.Set("Authorization", "Bearer "+token)
	reqGet.Header.Set("Accept", "application/vnd.github+json")
	reqGet.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if respGet, err := client.Do(reqGet); err == nil {
		defer respGet.Body.Close()
		if respGet.StatusCode == http.StatusOK {
			var meta struct {
				SHA string `json:"sha"`
			}
			if err := json.NewDecoder(respGet.Body).Decode(&meta); err == nil {
				existingSHA = meta.SHA
			}
		}
	}

	// Committer info
	committerName := os.Getenv("INPUT_COMMITTER_NAME")
	committerEmail := os.Getenv("INPUT_COMMITTER_EMAIL")
	actor := os.Getenv("GITHUB_ACTOR")
	if committerName == "" {
		if actor != "" {
			committerName = actor
		} else {
			committerName = "coding-metrics"
		}
	}
	if committerEmail == "" {
		if actor != "" {
			committerEmail = fmt.Sprintf("%s@users.noreply.github.com", actor)
		} else {
			committerEmail = "noreply@github.com"
		}
	}

	payload := map[string]interface{}{
		"message":   "Upload output.svg",
		"committer": map[string]string{"name": committerName, "email": committerEmail},
		"content":   base64.StdEncoding.EncodeToString(data),
	}
	if pushToBranch != "" {
		payload["branch"] = pushToBranch
	}
	if existingSHA != "" {
		payload["sha"] = existingSHA
	}

	bodyBuf, err := json.Marshal(payload)
	if err != nil {
		logger.Error("failed to marshal payload", zap.Error(err))
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), "PUT", apiURL, strings.NewReader(string(bodyBuf)))
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to call github api", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("github api returned non-OK", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		return
	}

	logger.Info("uploaded svg to repository", zap.String("repo", ownerRepo), zap.String("path", repoPath), zap.String("branch", pushToBranch))
}
