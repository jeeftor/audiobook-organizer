//go:build abs_e2e

package e2e

import "testing"

func TestRESTHarness_EmbeddedMetadataImportLifecycle(t *testing.T) {
	for _, tc := range restEmbeddedMetadataImportCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runRESTOrganizerImportLifecycle(t, tc)
		})
	}
}

func TestRESTHarness_FlatModeImportLifecycle(t *testing.T) {
	for _, tc := range restFlatModeImportCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runRESTOrganizerImportLifecycle(t, tc)
		})
	}
}

func restEmbeddedMetadataImportCases() []organizerImportCase {
	return []organizerImportCase{
		{
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
		},
		{
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
		},
	}
}

func restFlatModeImportCases() []organizerImportCase {
	return flatModeImportCases()
}

func runRESTOrganizerImportLifecycle(t *testing.T, tc organizerImportCase) {
	t.Helper()

	harness := newRESTHarness(t)
	defer harness.server.Close()

	var ctx restScenarioContext

	step(t, "01 reset and initial ABS scan", func(t *testing.T) {
		resetAndInitialScan(t)
		ctx = newRESTScenarioContext(t, harness, plainInstance, tc.library)
	})

	step(t, "02 assert REST sees no imported paths", func(t *testing.T) {
		waitForRESTABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.initialCount,
			missingCount:   0,
			absentContains: tc.newAPIPaths,
		})
	})

	step(t, "03 run organizer import through REST", func(t *testing.T) {
		var response struct {
			Summary struct {
				Moves []struct {
					From string `json:"from"`
					To   string `json:"to"`
				} `json:"Moves"`
			} `json:"summary"`
		}
		ctx.harness.postJSON(t, "/api/organize/run", map[string]any{
			"config": restOrganizerImportConfig(tc),
		}, &response)
		if len(response.Summary.Moves) == 0 {
			t.Fatalf("expected REST organizer import to move files")
		}
	})

	step(t, "04 assert filesystem import result", func(t *testing.T) {
		assertPathsNotExist(t, tc.oldFiles)
		assertPathsExist(t, tc.newFiles)
		assertExists(t, pathFromRoot(tc.logFile...))
	})

	step(t, "05 scan ABS through REST after import", func(t *testing.T) {
		triggerRESTABSScan(t, ctx)
	})

	step(t, "06 assert REST sees imported paths active and clean", func(t *testing.T) {
		waitForRESTABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.newAPIPaths,
		})
	})
}

func restOrganizerImportConfig(tc organizerImportCase) map[string]any {
	config := map[string]any{
		"base_dir":   pathFromRoot(tc.sourceParts...),
		"output_dir": pathFromRoot(tc.outputParts...),
		"layout":     "author-title",
	}
	for _, arg := range tc.args {
		switch arg {
		case "--use-embedded-metadata":
			config["use_embedded_metadata"] = true
		case "--flat":
			config["flat"] = true
			config["use_embedded_metadata"] = true
		}
	}
	return config
}
