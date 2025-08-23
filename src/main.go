package main

import (
	"os"

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
	file := createLocalFile()
	svgCanvas := setupSVG(file)
	generateSVGContent(svgCanvas)
	completeSVG(svgCanvas)

	commitSVGChanges()
}
