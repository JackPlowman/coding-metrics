package main

import (
	"os"

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
func createLocalFile() *os.File {
	file, err := os.Create("output.svg")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	return file
}
