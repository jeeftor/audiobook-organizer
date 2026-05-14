//go:build abs_e2e

package e2e

import "testing"

type organizerImportCase struct {
	name          string
	library       absLibrary
	sourceParts   []string
	outputParts   []string
	args          []string
	oldFiles      [][]string
	newFiles      [][]string
	logFile       []string
	newAPIPaths   []string
	initialCount  int
	expectedCount int
}

func TestEmbeddedMetadataImport_AudiobooksLifecycle(t *testing.T) {
	runOrganizerImportLifecycle(t, organizerImportCase{
		name:    "plain_audiobooks_imports_hierarchical_embedded_files",
		library: audiobooksLibrary,
		sourceParts: []string{
			"test", "abs", "runtime", "import-input", "audiobooks",
		},
		outputParts: []string{
			"test", "abs", "runtime", "plain", "audiobooks",
		},
		args: []string{
			"--use-embedded-metadata",
			"--layout", "author-title",
		},
		oldFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"import-input",
				"audiobooks",
				"dropbox",
				"jane-doe-mess",
				"source.m4b",
			},
			{
				"test",
				"abs",
				"runtime",
				"import-input",
				"audiobooks",
				"dropbox",
				"longname-mess",
				"source.m4b",
			},
		},
		newFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"audiobooks",
				"Jane Doe",
				"Mystery of the Lost City",
				"source.m4b",
			},
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"audiobooks",
				"Alexander von Longname",
				"The Epic Tale That Spans Generations",
				"source.m4b",
			},
		},
		logFile: []string{
			"test", "abs", "runtime", "plain", "audiobooks", ".abook-org.log",
		},
		newAPIPaths: []string{
			"/audiobooks/Alexander von Longname/The Epic Tale That Spans Generations",
			"/audiobooks/Jane Doe/Mystery of the Lost City",
		},
		initialCount:  2,
		expectedCount: 4,
	})
}

func TestEmbeddedMetadataImport_BooksLifecycle(t *testing.T) {
	runOrganizerImportLifecycle(t, organizerImportCase{
		name:    "plain_books_imports_hierarchical_embedded_epubs",
		library: booksLibrary,
		sourceParts: []string{
			"test", "abs", "runtime", "import-input", "books",
		},
		outputParts: []string{
			"test", "abs", "runtime", "plain", "books",
		},
		args: []string{
			"--use-embedded-metadata",
			"--layout", "author-title",
		},
		oldFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"import-input",
				"books",
				"dropbox",
				"cool-stuff",
				"source.epub",
			},
			{
				"test",
				"abs",
				"runtime",
				"import-input",
				"books",
				"dropbox",
				"testing-knowledge",
				"source.epub",
			},
		},
		newFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"books",
				"Jeef of Github,Some random guy",
				"The book of cool stuff",
				"source.epub",
			},
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"books",
				"Jeef of Github,Some random guy",
				"First book of testing knowledge",
				"source.epub",
			},
		},
		logFile: []string{
			"test", "abs", "runtime", "plain", "books", ".abook-org.log",
		},
		newAPIPaths: []string{
			"/books/Jeef of Github,Some random guy/First book of testing knowledge",
			"/books/Jeef of Github,Some random guy/The book of cool stuff",
		},
		initialCount:  3,
		expectedCount: 5,
	})
}

func TestFlatModeImport_AudiobooksLifecycle(t *testing.T) {
	runOrganizerImportLifecycle(t, organizerImportCase{
		name:    "plain_audiobooks_imports_loose_flat_files",
		library: audiobooksLibrary,
		sourceParts: []string{
			"test", "abs", "runtime", "flat-input", "audiobooks",
		},
		outputParts: []string{
			"test", "abs", "runtime", "plain", "audiobooks",
		},
		args: []string{
			"--flat",
			"--layout", "author-title",
		},
		oldFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"flat-input",
				"audiobooks",
				"inbox",
				"charlesdexterward_01_lovecraft_64kb.mp3",
			},
			{
				"test",
				"abs",
				"runtime",
				"flat-input",
				"audiobooks",
				"inbox",
				"falstaffswedding1766version_1_kenrick_64kb.mp3",
			},
			{
				"test",
				"abs",
				"runtime",
				"flat-input",
				"audiobooks",
				"inbox",
				"perouse_01_scott_64kb.mp3",
			},
		},
		newFiles: [][]string{
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"audiobooks",
				"H. P. Lovecraft",
				"01 - Chapter 1_ A Result and a Prologue",
				"01 - charlesdexterward_01_lovecraft_64kb.mp3",
			},
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"audiobooks",
				"William Kenrick",
				"01 - Act 1",
				"01 - falstaffswedding1766version_1_kenrick_64kb.mp3",
			},
			{
				"test",
				"abs",
				"runtime",
				"plain",
				"audiobooks",
				"Ernest Scott",
				"01 - Family, youth and influences",
				"01 - perouse_01_scott_64kb.mp3",
			},
		},
		logFile: []string{
			"test", "abs", "runtime", "plain", "audiobooks", ".abook-org.log",
		},
		newAPIPaths: []string{
			"/audiobooks/Ernest Scott/01 - Family, youth and influences",
			"/audiobooks/H. P. Lovecraft/01 - Chapter 1_ A Result and a Prologue",
			"/audiobooks/William Kenrick/01 - Act 1",
		},
		initialCount:  2,
		expectedCount: 5,
	})
}

func runOrganizerImportLifecycle(t *testing.T, tc organizerImportCase) {
	t.Helper()

	var ctx absScenarioContext

	t.Run(tc.name, func(t *testing.T) {
		step(t, "01 reset and initial ABS scan", func(t *testing.T) {
			resetAndInitialScan(t)
			ctx = newABSScenarioContext(t, plainInstance, tc.library)
		})

		step(t, "02 assert ABS starts without imported paths", func(t *testing.T) {
			waitForABSState(t, ctx, absStateExpectation{
				expectedCount:  tc.initialCount,
				missingCount:   0,
				absentContains: tc.newAPIPaths,
			})
		})

		step(t, "03 run organizer import", func(t *testing.T) {
			args := []string{
				"--dir", pathFromRoot(tc.sourceParts...),
				"--out", pathFromRoot(tc.outputParts...),
			}
			args = append(args, tc.args...)
			runOrganizer(t, args...)
		})

		step(t, "04 assert filesystem import result", func(t *testing.T) {
			assertPathsNotExist(t, tc.oldFiles)
			assertPathsExist(t, tc.newFiles)
			assertExists(t, pathFromRoot(tc.logFile...))
		})

		step(t, "05 scan ABS after import", func(t *testing.T) {
			scanLibraryAndWait(t, ctx)
		})

		step(t, "06 assert ABS imported paths are active and clean", func(t *testing.T) {
			waitForABSState(t, ctx, absStateExpectation{
				expectedCount:  tc.expectedCount,
				missingCount:   0,
				activeContains: tc.newAPIPaths,
			})
		})
	})
}
