package main

import (
	"embed"
	"flag"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	// Parse CLI arguments
	var inputDir, outputDir, logLevel string
	flag.StringVar(&inputDir, "dir", "", "Input directory to scan for audiobooks")
	flag.StringVar(&inputDir, "in", "", "Input directory to scan for audiobooks (alias for --dir)")
	flag.StringVar(&outputDir, "out", "", "Output directory for organized audiobooks")
	flag.StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	// Create an instance of the app structure with CLI args
	app := NewAppWithDirs(inputDir, outputDir)
	app.SetLogLevel(logLevel)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Audiobook Organizer",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "Audiobook Organizer",
				Message: "Organize your audiobook library with ease.\n\nhttps://github.com/jeeftor/audiobook-organizer",
				Icon:    icon,
			},
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
