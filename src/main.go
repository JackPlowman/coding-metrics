package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ajstarks/svgo"
)

// main is the entry point for the application.
func main() {
	width := 500
	height := 500
	file := createLocalFile()

	canvas := svg.New(file)
	canvas.Start(width, height)
	canvas.Circle(width/2, height/2, 100)
	canvas.Text(width/2, height/2, "Hello, SVG", "text-anchor:middle;font-size:30px;fill:white")
	canvas.End()
}

// createLocalFile creates a new SVG file in the local directory.
// createLocalFile creates a new SVG file in the system temp directory.
// If the temp directory is not writable, it falls back to stdout so the
// program doesn't panic in read-only environments (CI containers, Actions).
func createLocalFile() *os.File {
	path := filepath.Join(os.TempDir(), "output.svg")
	file, err := os.Create(path)
	if err == nil {
		fmt.Fprintln(os.Stderr, "writing SVG to", path)
		return file
	}

	// Fallback: write to stdout, but warn to stderr.
	fmt.Fprintln(os.Stderr, "warning: could not create", path, "-> writing to stdout instead:", err)
	return os.Stdout
}
