package main

import (
	"os"

	"github.com/twpayne/go-svg"
	"go.uber.org/zap"
)

// initialize global logger
func init() {
	logger, err := initLogger()
	zap.ReplaceGlobals(zap.Must(logger, err))
}

// initLogger initializes and returns a zap logger according to the
// DEBUG environment variable. If DEBUG=="true" a development logger
// will be returned, otherwise a production logger is used.
func initLogger() (*zap.Logger, error) {
	if os.Getenv("INPUT_DEBUG") == "true" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

// main is the entry point for the application.
func main() {

	svgElements := []svg.Element{}
	svgElements = append(svgElements, generateSVGContent()...)
	svg := createSVG(svgElements)
	file := createLocalFile(svg)
	commitSVGChanges(file)
}
