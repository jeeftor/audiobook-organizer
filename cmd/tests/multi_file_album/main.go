// test_multi_file_album.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

func main() {
	// Create a test directory structure
	testDir := "test_output/multi_file_album_test"
	outputDir := "test_output/multi_file_album_output"

	// Clean up previous test runs
	os.RemoveAll(testDir)
	os.RemoveAll(outputDir)

	// Create test directories
	os.MkdirAll(testDir, 0755)
	os.MkdirAll(outputDir, 0755)

	// Create a test album with multiple MP3 files that have consistent metadata
	albumDir := filepath.Join(testDir, "test_album")
	os.MkdirAll(albumDir, 0755)

	// Copy a controlled set of test MP3 files that we know have consistent metadata
	sourceFiles := []string{
		"strange_audiobook_1_Mystery_Series_Mystery_of_the_Lost_City_Jane_Doe_Tr1.mp3",
		"strange_audiobook_2_Mystery_Series_Mystery_of_the_Lost_City_Jane_Doe_Tr2.mp3",
		"strange_audiobook_3_Mystery_Series_Mystery_of_the_Lost_City_Jane_Doe_Tr3.mp3",
	}
	copySelectedFiles("testdata/mp3", albumDir, sourceFiles)

	// Create a second album with different metadata
	album2Dir := filepath.Join(testDir, "test_album2")
	os.MkdirAll(album2Dir, 0755)

	// Copy a second set of files with consistent metadata
	sourceFiles2 := []string{
		"strange_audiobook_6_Epic_Saga__Adventure__Quest___Glory__John_Smith_Tr1.mp3",
		"strange_audiobook_7_Epic_Saga__Adventure__Quest___Glory__John_Smith_Tr2.mp3",
		"strange_audiobook_8_Epic_Saga__Adventure__Quest___Glory__John_Smith_Tr3.mp3",
	}
	copySelectedFiles("testdata/mp3", album2Dir, sourceFiles2)

	// Initialize the organizer
	config := &organizer.OrganizerConfig{
		BaseDir:            testDir,
		OutputDir:          outputDir,
		Flat:               true,
		Verbose:            true,
		UseEmbeddedMetadata: true,
		Layout:             "author-series-title",
	}

	org := organizer.NewOrganizer(config)

	// Run the organizer
	fmt.Println("Starting multi-file album test...")
	fmt.Println("Processing directory:", testDir)
	fmt.Println("Output directory:", outputDir)

	if err := org.Execute(); err != nil {
		fmt.Printf("Error organizing: %v\n", err)
		os.Exit(1)
	}

	// Print simple summary
	fmt.Println("\nOrganization completed successfully!")

	// List files in output directory
	fmt.Println("\nFiles in output directory:")
	listOutputFiles(outputDir)

	fmt.Println("\nTest completed successfully!")
}

// listOutputFiles recursively lists all files in the output directory
func listOutputFiles(dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			rel = path
		}

		if rel == "." {
			return nil
		}

		if info.IsDir() {
			fmt.Printf("📁 %s/\n", rel)
		} else {
			fmt.Printf("📄 %s (%d bytes)\n", rel, info.Size())
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error listing files: %v\n", err)
	}
}

// copySelectedFiles copies specific files from source directory to destination directory
func copySelectedFiles(sourceDir, destDir string, fileNames []string) {
	for _, fileName := range fileNames {
		sourcePath := filepath.Join(sourceDir, fileName)
		destPath := filepath.Join(destDir, fileName)

		data, err := os.ReadFile(sourcePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", sourcePath, err)
			continue
		}

		err = os.WriteFile(destPath, data, 0644)
		if err != nil {
			fmt.Printf("Error writing file %s: %v\n", destPath, err)
		} else {
			fmt.Printf("Copied %s to %s\n", sourcePath, destPath)
		}
	}
}
