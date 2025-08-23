package main

import (
	"os"
	"context"
	"fmt"
	"strings"
	"time"
	"go.uber.org/zap"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// commitSVGChanges commits the changes made to the SVG file.
func commitSVGChanges(file *os.File) {
	ownerRepo := os.Getenv("INPUT_REPOSITORY")
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		zap.L().Fatal("Invalid repository format", zap.String("repository", ownerRepo))
		return
	}
	owner, repo := parts[0], parts[1]

	repoPath, err := cloneRepo(owner, repo)
	if err != nil {
		zap.L().Fatal("Failed to clone repo", zap.Error(err))
	} else {
		zap.L().Debug("Repo cloned", zap.String("path", repoPath))
	}

	// Copy the SVG file into the cloned repo at the root.
	svgPath := file.Name()
	outputFileName := os.Getenv("INPUT_OUTPUT_FILE_NAME")
	destPath := fmt.Sprintf("%s/%s", repoPath, outputFileName)

	input, err := os.ReadFile(svgPath)
	if err != nil {
		zap.L().Fatal("Failed to read SVG file", zap.Error(err))
		return
	}
	err = os.WriteFile(destPath, input, 0644)
	if err != nil {
		zap.L().Fatal("Failed to write SVG file to repo", zap.Error(err))
		return
	}
	zap.L().Debug("SVG file copied to repo", zap.String("dest", destPath))

	// Commit the changes to the repo.
	err = commitChanges(repoPath, outputFileName)
	if err != nil {
		zap.L().Fatal("Failed to commit changes", zap.Error(err))
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

// commitChanges commits and pushes changes to the repo using a GitHub Actions bot.
func commitChanges(repoPath string, outputFileName string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all changes
	err = w.AddGlob(".")
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit with bot info
	commitMsg := fmt.Sprintf("Update SVG %s via GitHub Actions:", outputFileName)
	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "github-actions[bot]",
			Email: "github-actions[bot]@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Push using token
	token := os.Getenv("INPUT_GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("missing INPUT_GITHUB_TOKEN")
	}

	testMode := os.Getenv("INPUT_TEST_MODE")
	if testMode == "false" {
		err = repo.Push(&git.PushOptions{
			RemoteName: "origin",
			Auth: &gitHttp.BasicAuth{
				Username: "x-access-token",
				Password: token,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
	}
	return nil
}
