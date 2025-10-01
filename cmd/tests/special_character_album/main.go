package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// This test script verifies that the album detection and grouping works correctly
// with files that have special characters in their metadata
func main() {
	// Setup test directories
	inputDir := filepath.Join("test_output", "special_chars_test")
	outputDir := filepath.Join("test_output", "special_chars_output")

	// Clean up previous test directories if they exist
	os.RemoveAll(inputDir)
	os.RemoveAll(outputDir)

	// Create test directories
	os.MkdirAll(inputDir, 0755)
	os.MkdirAll(outputDir, 0755)

	// Create test album directories
	fmt.Println("Setting up test albums with special characters...")

	// Create album with accented characters
	accentedDir := filepath.Join(inputDir, "accented_album")
	os.MkdirAll(accentedDir, 0755)

	// Copy test MP3 files from testdata directory - using emoji files
	copyTestFile(
		filepath.Join("testdata", "mp3track", "strange_audiobook_31_Series_With_Emoji____Audiobook_With_Emoji____Author_With_Emoji_____Tr1.mp3"),
		filepath.Join(accentedDir, "01 - Ångström & Café.mp3"),
	)
	copyTestFile(
		filepath.Join("testdata", "mp3track", "strange_audiobook_32_Series_With_Emoji____Audiobook_With_Emoji____Author_With_Emoji_____Tr2.mp3"),
		filepath.Join(accentedDir, "02 - Ångström & Café.mp3"),
	)
	copyTestFile(
		filepath.Join("testdata", "mp3track", "strange_audiobook_33_Series_With_Emoji____Audiobook_With_Emoji____Author_With_Emoji_____Tr3.mp3"),
		filepath.Join(accentedDir, "03 - Ångström & Café.mp3"),
	)

	// Create album with special symbols
	symbolsDir := filepath.Join(inputDir, "symbols_album")
	os.MkdirAll(symbolsDir, 0755)

	copyTestFile(
		filepath.Join("testdata", "mp3flat", "charlesdexterward_01_lovecraft_64kb.mp3"),
		filepath.Join(symbolsDir, "01 - Adventure & Quest+Glory!.mp3"),
	)
	copyTestFile(
		filepath.Join("testdata", "mp3flat", "falstaffswedding1766version_1_kenrick_64kb.mp3"),
		filepath.Join(symbolsDir, "02 - Adventure & Quest+Glory!.mp3"),
	)
	copyTestFile(
		filepath.Join("testdata", "mp3flat", "perouse_01_scott_64kb.mp3"),
		filepath.Join(symbolsDir, "03 - Adventure & Quest+Glory!.mp3"),
	)

	// Run the organizer in flat mode
	fmt.Println("Running organizer in flat mode...")
	org := organizer.NewOrganizer(&organizer.OrganizerConfig{
		BaseDir:   inputDir,
		OutputDir: outputDir,
		Flat:      true,
		Verbose:   true,
		DryRun:    false,
	})

	err := org.Execute()
	if err != nil {
		fmt.Printf("Error running organizer: %v\n", err)
		os.Exit(1)
	}

	// Check the output directory to verify album grouping
	fmt.Println("\nChecking output directory structure...")
	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == outputDir {
			return nil
		}
		rel, _ := filepath.Rel(outputDir, path)
		fmt.Printf("- %s\n", rel)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking output directory: %v\n", err)
	}

	fmt.Println("\nTest completed.")
}

// copyTestFile copies a test MP3 file from source to destination
func copyTestFile(src, dst string) {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Printf("Error opening source file %s: %v\n", src, err)
		os.Exit(1)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		fmt.Printf("Error creating destination file %s: %v\n", dst, err)
		os.Exit(1)
	}
	defer dstFile.Close()

	// Copy the contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Printf("Error copying file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Copied test file to: %s\n", dst)
}
