package main

import (
	"github.com/ajstarks/svgo"
	"os"
)

func main() {
	width := 500
	height := 500
	file, err := os.Create("output.svg")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	canvas := svg.New(file)
	canvas.Start(width, height)
	canvas.Circle(width/2, height/2, 100)
	canvas.Text(width/2, height/2, "Hello, SVG", "text-anchor:middle;font-size:30px;fill:white")
	canvas.End()
}
