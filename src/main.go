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

// initColourProfile initializes the global colour profile based on the INPUT_COLOUR_PROFILE environment variable
func initColourProfile() {
	profileName := os.Getenv("INPUT_COLOUR_PROFILE")
	if profileName == "" {
		profileName = "default"
	}
	currentColourProfile = GetColourProfile(profileName)
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
	// Initialize the color profile based on environment variable
	initColourProfile()

	svgElements := []svg.Element{}
	svgElements = append(svgElements, generateSVGContent()...)
	svg := createSVG(svgElements)
	file := createLocalFile(svg)
	commitSVGChanges(file)
}

