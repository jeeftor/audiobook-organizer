//go:build abs_e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jeeftor/audiobook-organizer/internal/app"
	"github.com/jeeftor/audiobook-organizer/internal/server"
)

const restHarnessToken = "abs-rest-harness-token"

type restHarness struct {
	server *httptest.Server
}

type restScenarioContext struct {
	harness   *restHarness
	config    map[string]any
	libraryID string
	instance  absInstance
	library   absLibrary
}

type restABSLibrariesResponse struct {
	Libraries []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Folders []struct {
			Path     string `json:"path"`
			FullPath string `json:"fullPath"`
		} `json:"folders"`
	} `json:"libraries"`
}

type restABSLibraryStateResponse struct {
	Items []restABSLibraryItem `json:"items"`
}

type restABSLibraryItem struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	RelPath   string `json:"rel_path"`
	IsMissing bool   `json:"is_missing"`
	IsInvalid bool   `json:"is_invalid"`
	MediaType string `json:"media_type"`
	Title     string `json:"title"`
}

type restABSState struct {
	items        []restABSLibraryItem
	activePaths  []string
	missingPaths []string
	allPaths     []string
}

func TestRESTHarness_MetadataJSONModeLifecycle(t *testing.T) {
	cases := append(metadataJSONAudiobookCases(), metadataJSONBookCases()...)
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runRESTMetadataJSONLifecycle(t, tc)
		})
	}
}

func runRESTMetadataJSONLifecycle(t *testing.T, tc metadataJSONLifecycleCase) {
	harness := newRESTHarness(t)
	defer harness.server.Close()

	var ctx restScenarioContext

	step(t, "01 reset and initial ABS scan", func(t *testing.T) {
		resetAndInitialScan(t)
		ctx = newRESTScenarioContext(t, harness, tc.instance, tc.library)
	})

	step(t, "02 assert REST sees messy active paths", func(t *testing.T) {
		waitForRESTABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.oldAPIPaths,
			absentContains: tc.newAPIPaths,
		})
	})

	step(t, "03 run organizer through REST", func(t *testing.T) {
		var response struct {
			Summary struct {
				Moves []struct {
					From string `json:"from"`
					To   string `json:"to"`
				} `json:"Moves"`
			} `json:"summary"`
		}
		harness.postJSON(t, "/api/organize/run", map[string]any{
			"config": map[string]any{
				"base_dir": pathFromRoot(tc.inputParts...),
				"layout":   "author-title",
			},
		}, &response)
		if tc.expectMove && len(response.Summary.Moves) == 0 {
			t.Fatalf("expected REST organizer run to move files")
		}
		if !tc.expectMove && len(response.Summary.Moves) != 0 {
			t.Fatalf(
				"expected REST organizer run to leave files in place, got %d move(s)",
				len(response.Summary.Moves),
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

	step(t, "05 scan ABS through REST", func(t *testing.T) {
		triggerRESTABSScan(t, ctx)
	})

	if tc.expectMove {
		step(
			t,
			"06 assert REST sees old paths missing and organized paths active",
			func(t *testing.T) {
				waitForRESTABSState(t, ctx, absStateExpectation{
					expectedCount:   tc.expectedCount * 2,
					missingCount:    len(tc.oldAPIPaths),
					activeContains:  tc.newAPIPaths,
					missingContains: tc.oldAPIPaths,
				})
			},
		)

		step(t, "07 clean ABS missing rows through REST", func(t *testing.T) {
			cleanRESTABSMissing(t, ctx, tc.oldAPIPaths)
		})

		step(t, "08 scan ABS through REST after cleanup", func(t *testing.T) {
			triggerRESTABSScan(t, ctx)
		})

		step(t, "09 assert REST sees clean organized final state", func(t *testing.T) {
			waitForRESTABSState(t, ctx, absStateExpectation{
				expectedCount:  tc.expectedCount,
				missingCount:   0,
				activeContains: tc.newAPIPaths,
				absentContains: tc.oldAPIPaths,
			})
		})
		return
	}

	step(t, "06 assert REST sees unchanged clean final state", func(t *testing.T) {
		waitForRESTABSState(t, ctx, absStateExpectation{
			expectedCount:  tc.expectedCount,
			missingCount:   0,
			activeContains: tc.oldAPIPaths,
			absentContains: tc.newAPIPaths,
		})
	})
}

func newRESTHarness(t *testing.T) *restHarness {
	t.Helper()

	service := app.NewService(app.DefaultWebConfig("127.0.0.1", 0, false, "", ""))
	srv, err := server.New(server.Config{Token: restHarnessToken}, service)
	if err != nil {
		t.Fatalf("new REST server: %v", err)
	}
	return &restHarness{server: httptest.NewServer(srv.Handler())}
}

func newRESTScenarioContext(
	t *testing.T,
	harness *restHarness,
	instance absInstance,
	library absLibrary,
) restScenarioContext {
	t.Helper()
	loadABSTestingEnv(t)

	baseURL := os.Getenv(instance.envURL)
	if baseURL == "" {
		t.Fatalf("%s is required", instance.envURL)
	}
	token := os.Getenv("ABS_TOKEN")
	if token == "" {
		t.Fatal("ABS_TOKEN is required")
	}

	cfg := map[string]any{
		"url":   baseURL,
		"token": token,
	}

	var libraries restABSLibrariesResponse
	harness.postJSON(t, "/api/abs/libraries", cfg, &libraries)

	for _, candidate := range libraries.Libraries {
		for _, folder := range candidate.Folders {
			if candidate.Name == library.name &&
				(folder.Path == library.folderPath || folder.FullPath == library.folderPath) {
				cfg["library_id"] = candidate.ID
				cfg["path_mappings"] = []map[string]string{
					{
						"abs_prefix":   library.folderPath,
						"local_prefix": localLibraryPath(instance, library),
					},
				}
				return restScenarioContext{
					harness:   harness,
					config:    cfg,
					libraryID: candidate.ID,
					instance:  instance,
					library:   library,
				}
			}
		}
	}

	t.Fatalf(
		"ABS library %q with folder %q not found on %s",
		library.name,
		library.folderPath,
		instance.name,
	)
	return restScenarioContext{}
}

func localLibraryPath(instance absInstance, library absLibrary) string {
	instanceDir := "plain"
	if instance.name == metadataEnabledInstance.name {
		instanceDir = "metadata"
	}
	return filepath.Join(
		repoRootPath,
		"test",
		"abs",
		"runtime",
		instanceDir,
		strings.TrimPrefix(library.folderPath, "/"),
	)
}

func triggerRESTABSScan(t *testing.T, ctx restScenarioContext) {
	t.Helper()
	var response struct {
		Triggered bool   `json:"triggered"`
		LibraryID string `json:"library_id"`
	}
	ctx.harness.postJSON(t, "/api/abs/scan-trigger", map[string]any{
		"config": ctx.config,
	}, &response)
	if !response.Triggered || response.LibraryID != ctx.libraryID {
		t.Fatalf("unexpected scan response: %+v", response)
	}
}

func cleanRESTABSMissing(t *testing.T, ctx restScenarioContext, oldPathFragments []string) {
	t.Helper()
	state := waitForRESTABSState(t, ctx, absStateExpectation{
		missingCount:    len(oldPathFragments),
		missingContains: oldPathFragments,
	})
	for _, item := range state.items {
		if item.IsMissing && containsAnyFragment(item.Path, oldPathFragments) {
			continue
		}
		if item.IsMissing {
			t.Fatalf(
				"refusing to clean ABS issues because unexpected missing item exists: id=%s path=%s",
				item.ID,
				item.Path,
			)
		}
	}

	var response struct {
		Cleaned   bool   `json:"cleaned"`
		LibraryID string `json:"library_id"`
	}
	ctx.harness.postJSON(t, "/api/abs/clean-missing", map[string]any{
		"config": ctx.config,
	}, &response)
	if !response.Cleaned || response.LibraryID != ctx.libraryID {
		t.Fatalf("unexpected clean response: %+v", response)
	}
}

func waitForRESTABSState(
	t *testing.T,
	ctx restScenarioContext,
	want absStateExpectation,
) restABSState {
	t.Helper()

	deadline := time.Now().Add(2 * time.Minute)
	var last restABSState

	for {
		last = readRESTABSState(t, ctx)
		if matchesRESTABSState(last, want) {
			return last
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf(
		"REST ABS state did not match for %s/%s\nexpected count: %d\nactual count: %d\nexpected missing: %d\nactual missing: %d\nactive contains: %v\nmissing contains: %v\nabsent contains: %v\nitems:\n%s",
		ctx.instance.name,
		ctx.library.name,
		want.expectedCount,
		len(last.items),
		want.missingCount,
		len(last.missingPaths),
		want.activeContains,
		want.missingContains,
		want.absentContains,
		formatRESTABSItems(last.items),
	)
	return restABSState{}
}

func readRESTABSState(t *testing.T, ctx restScenarioContext) restABSState {
	t.Helper()

	var response restABSLibraryStateResponse
	ctx.harness.postJSON(t, "/api/abs/library-state", map[string]any{
		"config": ctx.config,
	}, &response)

	state := restABSState{
		items: response.Items,
	}
	for _, item := range response.Items {
		state.allPaths = append(state.allPaths, item.Path)
		if item.IsMissing {
			state.missingPaths = append(state.missingPaths, item.Path)
		} else {
			state.activePaths = append(state.activePaths, item.Path)
		}
	}
	return state
}

func matchesRESTABSState(state restABSState, want absStateExpectation) bool {
	if want.expectedCount > 0 && len(state.items) != want.expectedCount {
		return false
	}
	if len(state.missingPaths) != want.missingCount {
		return false
	}
	if !containsAll(state.activePaths, want.activeContains) {
		return false
	}
	if !containsAll(state.missingPaths, want.missingContains) {
		return false
	}
	if !containsNone(state.allPaths, want.absentContains) {
		return false
	}
	return true
}

func formatRESTABSItems(items []restABSLibraryItem) string {
	var rows []string
	for _, item := range items {
		rows = append(
			rows,
			item.ID+" missing="+boolString(
				item.IsMissing,
			)+" path="+item.Path+" relPath="+item.RelPath,
		)
	}
	return strings.Join(rows, "\n")
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (h *restHarness) postJSON(t *testing.T, path string, body any, target any) {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal REST request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, h.server.URL+path, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create REST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Audiobook-Organizer-Token", restHarnessToken)

	resp, err := h.server.Client().Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody bytes.Buffer
		_, _ = errorBody.ReadFrom(resp.Body)
		t.Fatalf(
			"POST %s status = %d, want %d\n%s",
			path,
			resp.StatusCode,
			http.StatusOK,
			errorBody.String(),
		)
	}
	if target == nil {
		return
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode REST response from %s: %v", path, err)
	}
}
