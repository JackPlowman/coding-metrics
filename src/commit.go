package main

import (
	"os"
	"context"
	"fmt"
	"strings"
	"go.uber.org/zap"
	"github.com/go-git/go-git/v5"
)

// commitSVGChanges commits the changes made to the SVG file.
func commitSVGChanges() {
	ownerRepo := os.Getenv("INPUT_REPOSITORY")
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		zap.L().Error("Invalid repository format", zap.String("repository", ownerRepo))
		return
	}
	owner, repo := parts[0], parts[1]

	repoPath, err := cloneRepo(owner, repo)
	if err != nil {
		zap.L().Error("Failed to clone repo", zap.Error(err))
	} else {
		zap.L().Debug("Repo cloned", zap.String("path", repoPath))
	}
}

// cloneRepo clones a GitHub repository given the owner and repo name.
// It clones into a temp directory and returns the path or an error.
func cloneRepo(owner, repo string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	url := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	_, err = git.PlainCloneContext(context.Background(), tmpDir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repo: %w", err)
	}
	return tmpDir, nil
}
