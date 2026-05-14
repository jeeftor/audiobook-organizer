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
	absentFiles    [][]string
	activeAPIPaths []string
	absentAPIPaths []string
	expectedCount  int
}

func TestEmbeddedAlreadyIndexed_AudiobooksCurrentBehavior(t *testing.T) {
	runEmbeddedAlreadyIndexedCurrentBehavior(t, embeddedAlreadyIndexedCase{
		name:    "plain_audiobooks_embedded_mode_leaves_already_indexed_nested_library_unchanged",
		library: audiobooksLibrary,
		inputParts: []string{
			"test", "abs", "runtime", "plain", "audiobooks",
		},
		oldFiles: [][]string{
			{"test", "abs", "runtime", "plain", "audiobooks", "unsorted-audio", "drop-001", "not-alice.m4b"},
			{"test", "abs", "runtime", "plain", "audiobooks", "loose", "holiday_story_final.m4b"},
		},
		absentFiles: [][]string{
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
		if !strings.Contains(output, "Metadata files found: 0") {
			t.Fatalf("expected organizer to find no metadata rows in current already-indexed mode\noutput:\n%s", output)
		}
		if !strings.Contains(output, "Moves planned/executed: 0") {
			t.Fatalf("expected organizer to leave already-indexed files in place\noutput:\n%s", output)
		}
	})

	step(t, "04 assert filesystem stays unchanged", func(t *testing.T) {
		assertPathsExist(t, tc.oldFiles)
		assertPathsNotExist(t, tc.absentFiles)
	})

	step(t, "05 scan ABS after organizer run", func(t *testing.T) {
		scanLibraryAndWait(t, ctx)
	})

	step(t, "06 assert ABS final state is unchanged and clean", func(t *testing.T) {
		waitForABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.activeAPIPaths,
			absentContains: tc.absentAPIPaths,
		})
	})
}
