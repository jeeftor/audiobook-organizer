//go:build abs_e2e

package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

type absMetadataOrganizeCase struct {
	name          string
	library       absLibrary
	inputParts    []string
	oldFiles      [][]string
	newFiles      [][]string
	logFile       []string
	oldAPIPaths   []string
	newAPIPaths   []string
	expectedCount int
	seedAuthors   map[string]string
}

func TestABSMetadataMode_PreviewAndScanTrigger(t *testing.T) {
	var audioCtx absScenarioContext
	var booksCtx absScenarioContext

	step(t, "01 reset and initial ABS scan", func(t *testing.T) {
		resetAndInitialScan(t)
		audioCtx = newABSScenarioContext(t, plainInstance, audiobooksLibrary)
		booksCtx = newABSScenarioContext(t, plainInstance, booksLibrary)
	})

	step(t, "02 discovery lists both libraries", func(t *testing.T) {
		output := runOrganizer(
			t,
			"abs", "scan",
			"--abs-url", os.Getenv(plainInstance.envURL),
			"--abs-token", os.Getenv("ABS_TOKEN"),
		)
		assertOutputContains(
			t,
			output,
			"ABS Discovery Mode",
			"Found 2 libraries",
			"Audiobooks",
			"Ebooks",
		)
	})

	step(t, "03 manual mapping preview loads audiobook metadata", func(t *testing.T) {
		output := runOrganizer(
			t,
			"abs", "scan",
			"--abs-url", os.Getenv(plainInstance.envURL),
			"--abs-token", os.Getenv("ABS_TOKEN"),
			"--abs-library", audiobooksLibrary.name,
			"--abs-path-map", pathMap(audiobooksLibrary, plainInstance),
			"--dir", localLibraryPath(plainInstance, audiobooksLibrary),
			"--check-files",
		)
		assertOutputContains(
			t,
			output,
			"Using API-only mode",
			"Found 2 items",
			"Charles Dickens",
			"Lewis Carroll",
		)
	})

	step(t, "04 all libraries preview loads both mappings", func(t *testing.T) {
		output := runOrganizer(
			t,
			"abs", "scan",
			"--abs-url", os.Getenv(plainInstance.envURL),
			"--abs-token", os.Getenv("ABS_TOKEN"),
			"--abs-all-libraries",
			"--abs-path-map", pathMap(audiobooksLibrary, plainInstance),
			"--abs-path-map", pathMap(booksLibrary, plainInstance),
			"--dir", pathFromRoot("test", "abs", "runtime", "plain"),
			"--check-files",
		)
		assertOutputContains(
			t,
			output,
			"Using ALL LIBRARIES mode",
			"Found 5 items",
			"Items by library",
		)
	})

	step(t, "05 scan trigger accepts resolved library id", func(t *testing.T) {
		output := runOrganizer(
			t,
			"abs", "scan-trigger",
			"--abs-url", os.Getenv(plainInstance.envURL),
			"--abs-token", os.Getenv("ABS_TOKEN"),
			"--abs-library", audioCtx.libraryID,
		)
		assertOutputContains(t, output, "Library scan triggered successfully")
		waitForABSState(t, audioCtx, absStateExpectation{
			expectedCount:  2,
			missingCount:   0,
			activeContains: []string{"/audiobooks/loose", "/audiobooks/unsorted-audio"},
		})
		waitForABSState(t, booksCtx, absStateExpectation{
			expectedCount:  3,
			missingCount:   0,
			activeContains: []string{"/books/imported", "/books/random", "/books/to-sort"},
		})
	})
}

func TestABSMetadataMode_OrganizeAudiobooksLifecycle(t *testing.T) {
	runABSMetadataOrganizeLifecycle(t, absMetadataAudiobookCase())
}

func TestABSMetadataMode_OrganizeBooksLifecycle(t *testing.T) {
	runABSMetadataOrganizeLifecycle(t, absMetadataBookCase())
}

func absMetadataAudiobookCase() absMetadataOrganizeCase {
	return absMetadataOrganizeCase{
		name:    "plain_audiobooks_move_using_abs_metadata",
		library: audiobooksLibrary,
		inputParts: []string{
			"test", "abs", "runtime", "plain", "audiobooks",
		},
		oldFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"audiobooks",
				"unsorted-audio",
				"drop-001",
				"not-alice.m4b",
			},
			{"test", "abs", "runtime", "plain", "audiobooks", "loose", "holiday_story_final.m4b"},
		},
		newFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"audiobooks",
				"Charles Dickens",
				"A Christmas Carol_ A Ghost Story of Christmas",
				"holiday_story_final.m4b",
			},
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"audiobooks",
				"Lewis Carroll",
				"Alice's Adventures in Wonderland Version 2",
				"not-alice.m4b",
			},
		},
		logFile: []string{
			"test", "abs", "runtime", "plain", "audiobooks", ".abook-org.log",
		},
		oldAPIPaths: []string{
			"/audiobooks/loose",
			"/audiobooks/unsorted-audio",
		},
		newAPIPaths: []string{
			"/audiobooks/Charles Dickens/A Christmas Carol_ A Ghost Story of Christmas",
			"/audiobooks/Lewis Carroll/Alice's Adventures in Wonderland Version 2",
		},
		expectedCount: 2,
	}
}

func absMetadataBookCase() absMetadataOrganizeCase {
	return absMetadataOrganizeCase{
		name:    "plain_books_move_using_abs_metadata",
		library: booksLibrary,
		inputParts: []string{
			"test", "abs", "runtime", "plain", "books",
		},
		oldFiles: [][]string{
			{"test", "abs", "runtime", "plain", "books", "imported", "ebook-001.epub"},
			{"test", "abs", "runtime", "plain", "books", "random", "shelley-book.epub"},
			{"test", "abs", "runtime", "plain", "books", "to-sort", "austen.epub"},
		},
		newFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"books",
				"Lewis Carroll",
				"Alice's Adventures in Wonderland",
				"ebook-001.epub",
			},
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"books",
				"Mary Wollstonecraft Shelley",
				"Frankenstein; or, the modern prometheus",
				"shelley-book.epub",
			},
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"books",
				"Jane Austen",
				"Pride and Prejudice",
				"austen.epub",
			},
		},
		logFile: []string{
			"test", "abs", "runtime", "plain", "books", ".abook-org.log",
		},
		oldAPIPaths: []string{
			"/books/imported",
			"/books/random",
			"/books/to-sort",
		},
		newAPIPaths: []string{
			"/books/Jane Austen/Pride and Prejudice",
			"/books/Lewis Carroll/Alice's Adventures in Wonderland",
			"/books/Mary Wollstonecraft Shelley/Frankenstein; or, the modern prometheus",
		},
		expectedCount: 3,
		seedAuthors: map[string]string{
			"/books/imported": "Lewis Carroll",
			"/books/random":   "Mary Wollstonecraft Shelley",
			"/books/to-sort":  "Jane Austen",
		},
	}
}

func runABSMetadataOrganizeLifecycle(t *testing.T, tc absMetadataOrganizeCase) {
	var ctx absScenarioContext

	step(t, "01 reset and initial ABS scan", func(t *testing.T) {
		resetAndInitialScan(t)
		ctx = newABSScenarioContext(t, plainInstance, tc.library)
	})

	step(t, "02 seed ABS metadata required by fixture", func(t *testing.T) {
		seedABSAuthors(t, ctx, tc.seedAuthors)
	})

	step(t, "03 assert ABS starts with messy active paths", func(t *testing.T) {
		waitForABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.oldAPIPaths,
			absentContains: tc.newAPIPaths,
		})
	})

	step(t, "04 run abs organize", func(t *testing.T) {
		output := runOrganizer(
			t,
			"abs", "organize",
			"--abs-url", os.Getenv(plainInstance.envURL),
			"--abs-token", os.Getenv("ABS_TOKEN"),
			"--abs-library", tc.library.name,
			"--abs-path-map", pathMap(tc.library, plainInstance),
			"--dir", pathFromRoot(tc.inputParts...),
			"--layout", "author-title",
		)
		assertOutputContains(t, output, "Organizing with ABS metadata", "Moves planned/executed")
	})

	step(t, "05 assert filesystem result", func(t *testing.T) {
		assertPathsNotExist(t, tc.oldFiles)
		assertPathsExist(t, tc.newFiles)
		assertExists(t, pathFromRoot(tc.logFile...))
	})

	step(t, "06 scan ABS after organizer run", func(t *testing.T) {
		scanLibraryAndWait(t, ctx)
	})

	step(t, "07 assert ABS marks old paths missing and adds organized paths", func(t *testing.T) {
		waitForABSState(t, ctx, absStateExpectation{
			expectedCount:   tc.expectedCount * 2,
			missingCount:    len(tc.oldAPIPaths),
			activeContains:  tc.newAPIPaths,
			missingContains: tc.oldAPIPaths,
		})
	})

	step(t, "08 clean ABS library issues", func(t *testing.T) {
		cleanMissingABSItems(t, ctx, tc.oldAPIPaths)
	})

	step(t, "09 scan ABS after cleanup", func(t *testing.T) {
		scanLibraryAndWait(t, ctx)
	})

	step(t, "10 assert ABS final state is clean and organized", func(t *testing.T) {
		waitForABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.newAPIPaths,
			absentContains: tc.oldAPIPaths,
		})
	})
}

func seedABSAuthors(t *testing.T, ctx absScenarioContext, authorsByPath map[string]string) {
	t.Helper()
	if len(authorsByPath) == 0 {
		return
	}

	state := waitForABSState(t, ctx, absStateExpectation{
		expectedCount:  len(authorsByPath),
		activeContains: keys(authorsByPath),
	})
	for path, author := range authorsByPath {
		itemID := ""
		for _, item := range state.items {
			if item.Path == path {
				itemID = item.ID
				break
			}
		}
		if itemID == "" {
			t.Fatalf("could not find ABS item for metadata seed path %s", path)
		}

		err := ctx.client.UpdateLibraryItemMedia(itemID, map[string]interface{}{
			"metadata": map[string]interface{}{
				"authors": []map[string]string{{"name": author}},
			},
		})
		if err != nil {
			t.Fatalf("seed ABS author metadata for %s: %v", path, err)
		}
	}
}

func pathMap(library absLibrary, instance absInstance) string {
	return fmt.Sprintf("%s:%s", library.folderPath, localLibraryPath(instance, library))
}

func assertOutputContains(t *testing.T, output string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(output, needle) {
			t.Fatalf("expected output to contain %q\noutput:\n%s", needle, output)
		}
	}
}

func keys(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}
