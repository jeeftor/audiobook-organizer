package organizer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// PromptForDirectoryRemoval asks the user for confirmation before removing an empty directory
func (o *Organizer) PromptForDirectoryRemoval(dir string, isParent bool) bool {
	if isParent {
		color.Yellow("\n📁 Parent directory is now empty:")
	} else {
		color.Yellow("\n📁 Empty directory found:")
	}

	color.White("  Path: ")
	color.Yellow(dir)

	fmt.Print("\n❓ Remove empty directory? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Error reading response: %v", err)
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// PromptForConfirmation asks the user for confirmation before moving files.
// It displays the book metadata and the proposed move operation.
// Returns true if the user confirms with 'y' or 'yes' (case insensitive),
// returns false for any other input including empty input or errors.
func (o *Organizer) PromptForConfirmation(metadata Metadata, sourcePath, targetPath string) bool {
	color.Yellow("\n📖 Book found:")

	// Title
	fmt.Printf("  ")
	color.White("Title: ")
	color.Cyan(metadata.Title)

	// Authors
	fmt.Printf("  ")
	color.White("Authors: ")
	color.Cyan(strings.Join(metadata.Authors, ", "))

	// Series (if present)
	if len(metadata.Series) > 0 {
		cleanedSeries := CleanSeriesName(metadata.Series[0])
		fmt.Printf("  ")
		color.White("Series: ")
		color.Cyan(cleanedSeries)
	}

	color.Cyan("\n📝 Proposed move:")
	fmt.Printf("  ")
	color.White("From: ")
	color.Yellow(sourcePath)
	fmt.Printf("  ")
	color.White("To: ")
	color.Yellow(targetPath)

	fmt.Print("\n❓ Proceed with move? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Error reading response: %v", err)
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
