//go:build abs_e2e

package e2e

import (
	"strings"
	"testing"
)

type metadataJSONLifecycleCase struct {
	name          string
	instance      absInstance
	library       absLibrary
	inputParts    []string
	oldFiles      [][]string
	newFiles      [][]string
	logFile       []string
	oldAPIPaths   []string
	newAPIPaths   []string
	expectedCount int
	expectMove    bool
}

func TestMetadataJSONMode_AudiobooksLifecycle(t *testing.T) {
	for _, tc := range metadataJSONAudiobookCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMetadataJSONLifecycle(t, tc)
		})
	}
}

func metadataJSONAudiobookCases() []metadataJSONLifecycleCase {
	return []metadataJSONLifecycleCase{
		{
			name:     "metadata_enabled_audiobooks_moves_and_cleans_missing_abs_items",
			instance: metadataEnabledInstance,
			library:  audiobooksLibrary,
			inputParts: []string{
				"test", "abs", "runtime", "metadata", "audiobooks",
			},
			oldFiles: [][]string{
				{"test", "abs", "runtime", "metadata", "audiobooks", "unsorted-audio", "drop-001", "not-alice.m4b"},
				{"test", "abs", "runtime", "metadata", "audiobooks", "loose", "holiday_story_final.m4b"},
			},
			newFiles: [][]string{
				{"test", "abs", "runtime", "metadata", "audiobooks", "Lewis Carroll", "Alice's Adventures in Wonderland (Abridged)", "not-alice.m4b"},
				{"test", "abs", "runtime", "metadata", "audiobooks", "Charles Dickens", "A Christmas Carol", "holiday_story_final.m4b"},
			},
			logFile: []string{
				"test", "abs", "runtime", "metadata", "audiobooks", ".abook-org.log",
			},
			oldAPIPaths: []string{
				"/audiobooks/loose",
				"/audiobooks/unsorted-audio",
			},
			newAPIPaths: []string{
				"/audiobooks/Charles Dickens/A Christmas Carol",
				"/audiobooks/Lewis Carroll/Alice's Adventures in Wonderland (Abridged)",
			},
			expectedCount: 2,
			expectMove:    true,
		},
		{
			name:     "plain_audiobooks_no_sidecars_stays_clean",
			instance: plainInstance,
			library:  audiobooksLibrary,
			inputParts: []string{
				"test", "abs", "runtime", "plain", "audiobooks",
			},
			oldFiles: [][]string{
				{"test", "abs", "runtime", "plain", "audiobooks", "unsorted-audio", "drop-001", "not-alice.m4b"},
				{"test", "abs", "runtime", "plain", "audiobooks", "loose", "holiday_story_final.m4b"},
			},
			newFiles: [][]string{
				{"test", "abs", "runtime", "plain", "audiobooks", "Lewis Carroll"},
				{"test", "abs", "runtime", "plain", "audiobooks", "Charles Dickens"},
			},
			oldAPIPaths: []string{
				"/audiobooks/loose",
				"/audiobooks/unsorted-audio",
			},
			newAPIPaths: []string{
				"/audiobooks/Charles Dickens",
				"/audiobooks/Lewis Carroll",
			},
			expectedCount: 2,
			expectMove:    false,
		},
	}
}

func runMetadataJSONLifecycle(t *testing.T, tc metadataJSONLifecycleCase) {
	var ctx absScenarioContext

	step(t, "01 reset and initial ABS scan", func(t *testing.T) {
		resetAndInitialScan(t)
		ctx = newABSScenarioContext(t, tc.instance, tc.library)
	})

	step(t, "02 assert ABS starts with messy active paths", func(t *testing.T) {
		waitForABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.oldAPIPaths,
			absentContains: tc.newAPIPaths,
		})
	})

	step(t, "03 run organizer in metadata json mode", func(t *testing.T) {
		runOrganizer(t, "--dir", pathFromRoot(tc.inputParts...), "--layout", "author-title")
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
		step(t, "06 assert ABS marks old paths missing and adds organized paths", func(t *testing.T) {
			waitForABSState(t, ctx, absStateExpectation{
				expectedCount:   tc.expectedCount * 2,
				missingCount:    len(tc.oldAPIPaths),
				activeContains:  tc.newAPIPaths,
				missingContains: tc.oldAPIPaths,
			})
		})

		step(t, "07 clean ABS library issues", func(t *testing.T) {
			cleanMissingABSItems(t, ctx, tc.oldAPIPaths)
		})

		step(t, "08 scan ABS after cleanup", func(t *testing.T) {
			scanLibraryAndWait(t, ctx)
		})

		step(t, "09 assert ABS final state is clean and organized", func(t *testing.T) {
			waitForABSState(t, ctx, absStateExpectation{
				expectedCount:  tc.expectedCount,
				missingCount:   0,
				activeContains: tc.newAPIPaths,
				absentContains: tc.oldAPIPaths,
			})
		})
		return
	}

	step(t, "06 assert ABS final state is unchanged and clean", func(t *testing.T) {
		waitForABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.oldAPIPaths,
			absentContains: tc.newAPIPaths,
		})
	})
}

func step(t *testing.T, name string, fn func(*testing.T)) {
	t.Helper()
	if !t.Run(strings.ReplaceAll(name, " ", "_"), fn) {
		t.FailNow()
	}
}

func assertPathsExist(t *testing.T, paths [][]string) {
	t.Helper()
	for _, path := range paths {
		assertExists(t, pathFromRoot(path...))
	}
}

func assertPathsNotExist(t *testing.T, paths [][]string) {
	t.Helper()
	for _, path := range paths {
		assertNotExists(t, pathFromRoot(path...))
	}
}
