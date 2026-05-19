//go:build abs_e2e

package e2e

import (
	"strings"
	"testing"
)

type embeddedAlreadyIndexedCase struct {
	name           string
	library        absLibrary
	inputParts     []string
	oldFiles       [][]string
	newFiles       [][]string
	logFile        []string
	activeAPIPaths []string
	absentAPIPaths []string
	expectedCount  int
	expectMove     bool
}

func TestEmbeddedAlreadyIndexed_AudiobooksCurrentBehavior(t *testing.T) {
	runEmbeddedAlreadyIndexedCurrentBehavior(t, embeddedAlreadyIndexedCase{
		name:    "plain_audiobooks_embedded_mode_leaves_already_indexed_nested_library_unchanged",
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
			{"test", "abs", "runtime", "plain", "audiobooks", "Lewis Carroll"},
			{"test", "abs", "runtime", "plain", "audiobooks", "Charles Dickens"},
		},
		activeAPIPaths: []string{
			"/audiobooks/loose",
			"/audiobooks/unsorted-audio",
		},
		absentAPIPaths: []string{
			"/audiobooks/Charles Dickens",
			"/audiobooks/Lewis Carroll",
		},
		expectedCount: 2,
	})
}

func TestEmbeddedAlreadyIndexed_BooksCurrentBehavior(t *testing.T) {
	runEmbeddedAlreadyIndexedCurrentBehavior(t, embeddedAlreadyIndexedCase{
		name:    "plain_books_embedded_mode_moves_already_indexed_epub_library",
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
		activeAPIPaths: []string{
			"/books/imported",
			"/books/random",
			"/books/to-sort",
		},
		absentAPIPaths: []string{
			"/books/Jane Austen/Pride and Prejudice",
			"/books/Lewis Carroll/Alice's Adventures in Wonderland",
			"/books/Mary Wollstonecraft Shelley/Frankenstein; or, the modern prometheus",
		},
		expectedCount: 3,
		expectMove:    true,
	})
}

func runEmbeddedAlreadyIndexedCurrentBehavior(t *testing.T, tc embeddedAlreadyIndexedCase) {
	var ctx absScenarioContext
	var output string

	step(t, "01 reset and initial ABS scan", func(t *testing.T) {
		resetAndInitialScan(t)
		ctx = newABSScenarioContext(t, plainInstance, tc.library)
	})

	step(t, "02 assert ABS starts with messy active paths", func(t *testing.T) {
		waitForABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.activeAPIPaths,
			absentContains: tc.absentAPIPaths,
		})
	})

	step(t, "03 run organizer in embedded metadata mode", func(t *testing.T) {
		output = runOrganizer(
			t,
			"--dir", pathFromRoot(tc.inputParts...),
			"--use-embedded-metadata",
			"--layout", "author-title",
		)
		if !tc.expectMove {
			if !strings.Contains(output, "Metadata files found: 0") {
				t.Fatalf(
					"expected organizer to find no metadata rows in current already-indexed mode\noutput:\n%s",
					output,
				)
			}
			if !strings.Contains(output, "Moves planned/executed: 0") {
				t.Fatalf(
					"expected organizer to leave already-indexed files in place\noutput:\n%s",
					output,
				)
			}
			return
		}

		if strings.Count(output, "Found metadata in EPUB file:") != 3 {
			t.Fatalf(
				"expected organizer to read three embedded EPUB metadata files in current already-indexed mode\noutput:\n%s",
				output,
			)
		}
		if !strings.Contains(output, "Moves planned/executed: 3") {
			t.Fatalf(
				"expected organizer to move already-indexed EPUB files in current mode\noutput:\n%s",
				output,
			)
		}
	})

	step(t, "04 assert filesystem result", func(t *testing.T) {
		if tc.expectMove {
			assertPathsNotExist(t, tc.oldFiles)
			assertPathsExist(t, tc.newFiles)
			assertExists(t, pathFromRoot(tc.logFile...))
			return
		}

		assertPathsExist(t, tc.oldFiles)
		assertPathsNotExist(t, tc.newFiles)
	})

	step(t, "05 scan ABS after organizer run", func(t *testing.T) {
		scanLibraryAndWait(t, ctx)
	})

	if tc.expectMove {
		step(
			t,
			"06 assert ABS marks old paths missing and adds organized paths",
			func(t *testing.T) {
				waitForABSState(t, ctx, absStateExpectation{
					expectedCount:   tc.expectedCount * 2,
					missingCount:    len(tc.activeAPIPaths),
					activeContains:  tc.absentAPIPaths,
					missingContains: tc.activeAPIPaths,
				})
			},
		)

		step(t, "07 clean ABS library issues", func(t *testing.T) {
			cleanMissingABSItems(t, ctx, tc.activeAPIPaths)
		})

		step(t, "08 scan ABS after cleanup", func(t *testing.T) {
			scanLibraryAndWait(t, ctx)
		})

		step(t, "09 assert ABS final state is clean and organized", func(t *testing.T) {
			waitForABSState(t, ctx, absStateExpectation{
				expectedCount:  tc.expectedCount,
				missingCount:   0,
				activeContains: tc.absentAPIPaths,
				absentContains: tc.activeAPIPaths,
			})
		})
		return
	}

	step(t, "06 assert ABS final state is unchanged and clean", func(t *testing.T) {
		waitForABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.activeAPIPaths,
			absentContains: tc.absentAPIPaths,
		})
	})
}
