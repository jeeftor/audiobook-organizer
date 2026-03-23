//go:build !gui

package guiapp

import "fmt"

// Run is a stub used when the binary is built without the gui build tag.
func Run(inputDir, outputDir string) error {
	return fmt.Errorf(
		"GUI not included in this build.\n\nTo launch the GUI, download the GUI-enabled release from:\nhttps://github.com/jeeftor/audiobook-organizer/releases\n\nOr build locally with: make gui-unified",
	)
}
