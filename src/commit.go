package main

import (
	"context"
	"os"
	"strings"

	"github.com/google/go-github/v61/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// commitSVGChanges commits the changes made to the SVG file.
func commitSVGChanges(file *os.File) {
	testMode := os.Getenv("INPUT_TEST_MODE") == "true"
	if testMode {
		zap.L().Warn("Running in test mode")
		return
	}
	ownerRepo := os.Getenv("INPUT_REPOSITORY")
	parts := strings.Split(ownerRepo, "/")
	branch := "main"
	token := os.Getenv("INPUT_WORKFLOW_GITHUB_TOKEN")
	path := "output.svg"
	if len(parts) != 2 {
		zap.L().Fatal("Invalid repository format", zap.String("repository", ownerRepo))
		return
	}
	owner, repo := parts[0], parts[1]

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	gh := github.NewClient(tc)

	// Get current file SHA (omit if creating a new file)
	fileContent, _, _, _ := gh.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	var sha *string
	if fileContent != nil {
		sha = fileContent.SHA
	}

	contentBytes, err := os.ReadFile(file.Name())
	if err != nil {
		zap.L().Fatal("Failed to read SVG file", zap.Error(err))
		return
	}
	opts := &github.RepositoryContentFileOptions{
		Message: github.String("Update SVG file"),
		Content: contentBytes,
		SHA:     sha, // nil if new file
		Branch:  github.String(branch),
		// Leave Author/Committer nil to get a bot-verified signature
	}
	_, _, err = gh.Repositories.CreateFile(ctx, owner, repo, path, opts) // or UpdateFile if you set SHA
	if err != nil {
		zap.L().Fatal("Failed to upload SVG file", zap.Error(err))
	}
}
