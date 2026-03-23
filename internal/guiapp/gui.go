//go:build gui

package guiapp

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

// Run launches the Wails GUI window. It blocks until the window is closed.
func Run(inputDir, outputDir string) error {
	app := NewAppWithDirs(inputDir, outputDir)
	return wails.Run(&options.App{
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
}
