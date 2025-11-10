package main

import (
	"fmt"
	"os"
	"path/filepath"

	svg "github.com/twpayne/go-svg"

	"go.uber.org/zap"
)

func createSVG(svgChildren []svg.Element) *svg.SVGElement {
	// Add a background rectangle with the profile's background color as the first element
	bgRect := svg.Rect().
		Fill(svg.String(currentColourProfile.Background)).
		Width(svg.Px(1000)).
		Height(svg.Px(380)).
		X(svg.Px(0)).
		Y(svg.Px(0))

	// Prepend the background to the children
	allChildren := append([]svg.Element{bgRect}, svgChildren...)

	return svg.New().WidthHeight(1000, 380, svg.Px).ViewBox(0, 0, 1000, 380).AppendChildren(
		allChildren...,
	)
}

// createLocalFile creates a new SVG file in the system temp directory.
// If the temp directory is not writable, it falls back to stdout so the
// program doesn't panic in read-only environments (CI containers, Actions).
func createLocalFile(
	svgElement *svg.SVGElement,
) *os.File {
	path := filepath.Join(os.TempDir(), filepath.Clean("output.svg"))
	// #nosec G304 -- The file path is controlled and safe in this context.
	file, err := os.Create(path)
	if err != nil {
		zap.L().Fatal("Could not create SVG file", zap.Error(err))
	}
	zap.L().Info("Writing SVG to file", zap.String("path", path))
	if _, writeErr := svgElement.WriteTo(file); writeErr != nil {
		zap.L().Fatal("Failed to write SVG to file", zap.Error(writeErr))
	}
	if closeErr := file.Close(); closeErr != nil {
		zap.L().Error("Failed to close SVG file", zap.Error(closeErr))
	}

	return file
}

// writeToGitHubStepSummary writes the SVG content to the GitHub Actions step summary.
// It checks if GITHUB_STEP_SUMMARY environment variable is set and writes the SVG
// as an embedded code block. If not running in GitHub Actions, it logs a warning.
func writeToGitHubStepSummary(svgElement *svg.SVGElement) {
	summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryPath == "" {
		zap.L().Warn("GITHUB_STEP_SUMMARY not set, skipping step summary update")
		return
	}

	// #nosec G304 -- Path comes from GitHub Actions environment, controlled by GitHub
	summaryFile, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		zap.L().Error("Failed to open GitHub step summary file", zap.Error(err))
		return
	}
	defer func() {
		if closeErr := summaryFile.Close(); closeErr != nil {
			zap.L().Error("Failed to close GitHub step summary file", zap.Error(closeErr))
		}
	}()

	// Write header
	if _, err := fmt.Fprintln(summaryFile, "## Coding Metrics"); err != nil {
		zap.L().Error("Failed to write header to step summary", zap.Error(err))
		return
	}
	if _, err := fmt.Fprintln(summaryFile); err != nil {
		zap.L().Error("Failed to write newline to step summary", zap.Error(err))
		return
	}
	if _, err := fmt.Fprintln(summaryFile, "Generated coding metrics SVG:"); err != nil {
		zap.L().Error("Failed to write description to step summary", zap.Error(err))
		return
	}
	if _, err := fmt.Fprintln(summaryFile); err != nil {
		zap.L().Error("Failed to write newline to step summary", zap.Error(err))
		return
	}

	// Write SVG in code block
	if _, err := svgElement.WriteTo(summaryFile); err != nil {
		zap.L().Error("Failed to write SVG content to step summary", zap.Error(err))
		return
	}

	zap.L().Info("Successfully added SVG to GitHub Actions step summary")
}
