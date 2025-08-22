package main

import (
	"go.uber.org/zap"
	"os"
)

func init() {
	// initialize global logger
	logger, err := initLogger()
	zap.ReplaceGlobals(zap.Must(logger, err))
}

// initLogger initializes and returns a zap logger according to the
// DEBUG environment variable. If DEBUG=="true" a development logger
// will be returned, otherwise a production logger is used.
func initLogger() (*zap.Logger, error) {
	if os.Getenv("DEBUG") == "true" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

// main is the entry point for the application.
func main() {

	file := createLocalFile()
	svgCanvas := setupSVG(file)
	getPullRequestTotal()

	// Card coordinates used by multiple sections.
	cardX, cardY := 40, 30
	cardW, cardH := 720, 200
	drawCard(svgCanvas, cardX, cardY, cardW, cardH)
	drawHeader(svgCanvas, cardX, cardY)
	drawAvatar(svgCanvas, cardX, cardY, cardW, "JackPlowman") // TODO: get username from GITHUB_TOKEN
	drawMetrics(svgCanvas, cardX, cardY)
	drawLanguageBars(svgCanvas, cardX, cardY)
	drawFooter(svgCanvas, cardX, cardY, cardH)

	completeSVG(svgCanvas)
}
