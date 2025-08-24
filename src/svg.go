package main

import (
	"os"
	"path/filepath"

	svg "github.com/twpayne/go-svg"

	"go.uber.org/zap"
)

func createSVG(svgChildren []svg.Element) *svg.SVGElement {
	return svg.New().WidthHeight(1000, 380, svg.Px).ViewBox(0, 0, 1000, 380).AppendChildren(
		svgChildren...,
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
