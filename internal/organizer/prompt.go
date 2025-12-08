package organizer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptYesNo asks the user a yes/no question and returns their response
func (o *Organizer) PromptYesNo(question string) bool {
	fmt.Print(RenderPromptIcon("\n❓ " + question + " [y/N] "))

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf(RenderError("Error reading response: %v\n"), err)
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// PromptForDirectoryRemoval asks the user for confirmation before removing an empty directory
func (o *Organizer) PromptForDirectoryRemoval(dir string, isParent bool) bool {
	if isParent {
		fmt.Println(RenderWarning("\n📁 Parent directory is now empty:"))
	} else {
		fmt.Println(RenderWarning("\n📁 Empty directory found:"))
	}

	fmt.Print(RenderPrompt("  Path: "))
	fmt.Println(RenderPath(dir))

	fmt.Print(RenderPromptIcon("\n❓ Remove empty directory? [y/N] "))

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf(RenderError("Error reading response: %v\n"), err)
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
	fmt.Println(RenderWarning("\n📖 Book found:"))

	// Title
	fmt.Print("  ")
	fmt.Print(RenderPrompt("Title: "))
	fmt.Println(RenderHighlight(metadata.Title))

	// Authors
	fmt.Print("  ")
	fmt.Print(RenderPrompt("Authors: "))
	fmt.Println(RenderHighlight(strings.Join(metadata.Authors, ", ")))

	// Series (if present)
	if len(metadata.Series) > 0 {
		cleanedSeries := CleanSeriesName(metadata.Series[0])
		fmt.Print("  ")
		fmt.Print(RenderPrompt("Series: "))
		fmt.Println(RenderHighlight(cleanedSeries))
	}

	fmt.Println(RenderHighlight("\n📝 Proposed move:"))
	fmt.Print("  ")
	fmt.Print(RenderPrompt("From: "))
	fmt.Println(RenderPath(sourcePath))
	fmt.Print("  ")
	fmt.Print(RenderPrompt("To: "))
	fmt.Println(RenderPath(targetPath))

	fmt.Print(RenderPromptIcon("\n❓ Proceed with move? [y/N] "))

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf(RenderError("Error reading response: %v\n"), err)
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
