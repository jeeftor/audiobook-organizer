//go:build abs_e2e

package e2e

import "testing"

type organizerImportCase struct {
	name          string
	sourceParts   []string
	outputParts   []string
	args          []string
	oldFiles      [][]string
	newFiles      [][]string
	logFile       []string
	newAPIPaths   []string
	expectedCount int
}

func TestEmbeddedMetadataImport_AudiobooksLifecycle(t *testing.T) {
	runOrganizerImportLifecycle(t, organizerImportCase{
		name: "plain_audiobooks_imports_hierarchical_embedded_files",
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
			{"test", "abs", "runtime", "import-input", "audiobooks", "dropbox", "jane-doe-mess", "source.m4b"},
			{"test", "abs", "runtime", "import-input", "audiobooks", "dropbox", "longname-mess", "source.m4b"},
		},
		newFiles: [][]string{
			{"test", "abs", "runtime", "plain", "audiobooks", "Jane Doe", "Mystery of the Lost City", "source.m4b"},
			{"test", "abs", "runtime", "plain", "audiobooks", "Alexander von Longname", "The Epic Tale That Spans Generations", "source.m4b"},
		},
		logFile: []string{
			"test", "abs", "runtime", "plain", "audiobooks", ".abook-org.log",
		},
		newAPIPaths: []string{
			"/audiobooks/Alexander von Longname/The Epic Tale That Spans Generations",
			"/audiobooks/Jane Doe/Mystery of the Lost City",
		},
		expectedCount: 4,
	})
}

func runOrganizerImportLifecycle(t *testing.T, tc organizerImportCase) {
	t.Helper()

	var ctx absScenarioContext

	t.Run(tc.name, func(t *testing.T) {
		step(t, "01 reset and initial ABS scan", func(t *testing.T) {
			resetAndInitialScan(t)
			ctx = newABSScenarioContext(t, plainInstance, audiobooksLibrary)
		})

		step(t, "02 assert ABS starts without imported paths", func(t *testing.T) {
			waitForABSState(t, ctx, absStateExpectation{
				expectedCount:  2,
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
