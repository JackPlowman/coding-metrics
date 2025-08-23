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
	pushToBranch := os.Getenv("INPUT_UPLOAD_BRANCH")

	data, err := io.ReadAll(file)
	if err != nil {
		zap.L().Error("failed to read svg file", zap.Error(err))
		return
	}

	repoPath := "output.svg"
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", ownerRepo, repoPath)
	client := &http.Client{Timeout: 15 * time.Second}

	// Committer info
	committerName := "github-actions[bot]"
	committerEmail := "github-actions[bot]@users.noreply.github.com"
	commitMessage := os.Getenv("INPUT_COMMIT_MESSAGE")

	payload := map[string]interface{}{
		"message":   commitMessage,
		"committer": map[string]string{"name": committerName, "email": committerEmail},
		"content":   base64.StdEncoding.EncodeToString(data),
	}

	if pushToBranch != "" {
		payload["branch"] = pushToBranch
	}

	bodyBuf, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("failed to marshal payload", zap.Error(err))
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), "PUT", apiURL, strings.NewReader(string(bodyBuf)))
	if err != nil {
		zap.L().Error("failed to create request", zap.Error(err))
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("failed to call github api", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		zap.L().Error("github api returned non-OK", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		return
	}

	zap.L().Info("uploaded svg to repository", zap.String("repo", ownerRepo), zap.String("path", repoPath), zap.String("branch", pushToBranch))
}
