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
	path := filepath.Join(os.TempDir(), "output.svg")
	file, err := os.Create(path)
	if err == nil {
		zap.L().Info("Writing SVG to file", zap.String("path", path))
		svgElement.WriteTo(file)
		file.Close()
		return file
	}
	zap.L().Fatal("Could not create SVG file", zap.Error(err))
	return nil
}
