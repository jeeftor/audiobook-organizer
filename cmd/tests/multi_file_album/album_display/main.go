// cmd/tests/multi_file_album/album_display/main.go
package main

import (
	"fmt"
	"path/filepath"
	"os"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
	"github.com/jeeftor/audiobook-organizer/internal/tui/models"
)

func main() {
	fmt.Println("Testing Multi-File Album Display")
	fmt.Println("===============================")

	// Create a test directory with album files
	testDir := "./test_album_dir"
	albumDir := filepath.Join(testDir, "album")
	singleDir := filepath.Join(testDir, "single")

	// Create test directories
	os.MkdirAll(albumDir, 0755)
	os.MkdirAll(singleDir, 0755)

	// Create test files (just creating empty files for demonstration)
	track1 := filepath.Join(albumDir, "track01.mp3")
	track2 := filepath.Join(albumDir, "track02.mp3")
	track3 := filepath.Join(albumDir, "track03.mp3")
	singleBook := filepath.Join(singleDir, "book.m4b")

	// Create empty test files
	for _, file := range []string{track1, track2, track3, singleBook} {
		f, _ := os.Create(file)
		f.Close()
	}

	// Create test AudioBook objects
	testBooks := []models.AudioBook{
		{
			Path: track1,
			Metadata: organizer.Metadata{
				Title:       "Test Album",
				Authors:     []string{"Test Author"},
				TrackNumber: 1,
			},
			IsPartOfAlbum: true,
			AlbumName:     "Test Album Collection",
			TrackNumber:   1,
			TotalTracks:   3,
		},
		{
			Path: track2,
			Metadata: organizer.Metadata{
				Title:       "Test Album",
				Authors:     []string{"Test Author"},
				TrackNumber: 2,
			},
			IsPartOfAlbum: true,
			AlbumName:     "Test Album Collection",
			TrackNumber:   2,
			TotalTracks:   3,
		},
		{
			Path: track3,
			Metadata: organizer.Metadata{
				Title:       "Test Album",
				Authors:     []string{"Test Author"},
				TrackNumber: 3,
			},
			IsPartOfAlbum: true,
			AlbumName:     "Test Album Collection",
			TrackNumber:   3,
			TotalTracks:   3,
		},
		{
			Path: singleBook,
			Metadata: organizer.Metadata{
				Title:   "Regular Book",
				Authors: []string{"Another Author"},
			},
			IsPartOfAlbum: false,
		},
	}

	// Display book information
	fmt.Println("\nBook Display Information:")
	fmt.Println("------------------------")

	for i, book := range testBooks {
		fmt.Printf("\nBook %d:\n", i+1)
		fmt.Printf("  Path: %s\n", filepath.Base(book.Path))
		fmt.Printf("  Title: %s\n", book.Metadata.Title)
		fmt.Printf("  Author: %s\n", book.Metadata.GetFirstAuthor("Unknown"))
		fmt.Printf("  Is Album: %t\n", book.IsPartOfAlbum)
		if book.IsPartOfAlbum {
			fmt.Printf("  Album Name: %s\n", book.AlbumName)
			fmt.Printf("  Track: %d of %d\n", book.TrackNumber, book.TotalTracks)
		}
	}

	// Simulate how the UI would display these books
	fmt.Println("\nUI Display Simulation:")
	fmt.Println("--------------------")

	for i, book := range testBooks {
		fmt.Printf("\nBook %d UI Representation:\n", i+1)

		// Simulate the Title method
		title := book.Metadata.Title
		if book.IsPartOfAlbum {
			title = fmt.Sprintf("📀 %s (Track %d/%d)", title, book.TrackNumber, book.TotalTracks)
		}
		fmt.Printf("  Title: %s\n", title)

		// Simulate the Description method
		description := fmt.Sprintf("%s by %s", book.Metadata.Title, book.Metadata.GetFirstAuthor("Unknown"))
		if book.IsPartOfAlbum {
			description = fmt.Sprintf("%s | Album: %s | Track %d of %d",
				description, book.AlbumName, book.TrackNumber, book.TotalTracks)
		}
		fmt.Printf("  Description: %s\n", description)
	}

	// Clean up test files
	os.RemoveAll(testDir)

	fmt.Println("\nTest completed successfully!")
}
