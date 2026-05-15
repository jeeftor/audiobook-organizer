//go:build abs_e2e

package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jeeftor/audiobook-organizer/internal/abs"
)

type absInstance struct {
	name   string
	envURL string
}

type absLibrary struct {
	name       string
	folderPath string
}

var (
	plainInstance = absInstance{
		name:   "plain",
		envURL: "ABS_PLAIN_URL",
	}
	metadataEnabledInstance = absInstance{
		name:   "metadata-enabled",
		envURL: "ABS_METADATA_URL",
	}
	audiobooksLibrary = absLibrary{
		name:       "Audiobooks",
		folderPath: "/audiobooks",
	}
	booksLibrary = absLibrary{
		name:       "Ebooks",
		folderPath: "/books",
	}
)

type absScenarioContext struct {
	client    *abs.Client
	libraryID string
	instance  absInstance
	library   absLibrary
}

type absStateExpectation struct {
	expectedCount   int
	missingCount    int
	activeContains  []string
	missingContains []string
	absentContains  []string
}

type absLibraryState struct {
	items        []abs.LibraryItem
	activePaths  []string
	missingPaths []string
	allPaths     []string
}

func newABSScenarioContext(
	t *testing.T,
	instance absInstance,
	library absLibrary,
) absScenarioContext {
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

	client := abs.NewClient(baseURL, token)
	libraries, err := client.GetLibraries()
	if err != nil {
		t.Fatalf("list ABS libraries for %s: %v", instance.name, err)
	}

	for _, candidate := range libraries {
		for _, folder := range candidate.Folders {
			if candidate.Name == library.name &&
				(folder.Path == library.folderPath || folder.FullPath == library.folderPath) {
				return absScenarioContext{
					client:    client,
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
	return absScenarioContext{}
}

func scanLibraryAndWait(t *testing.T, ctx absScenarioContext) {
	t.Helper()
	if err := ctx.client.ScanLibraryForce(ctx.libraryID); err != nil {
		t.Fatalf("scan ABS library %s/%s: %v", ctx.instance.name, ctx.library.name, err)
	}
}

func waitForABSState(
	t *testing.T,
	ctx absScenarioContext,
	want absStateExpectation,
) absLibraryState {
	t.Helper()

	deadline := time.Now().Add(2 * time.Minute)
	var last absLibraryState
	var lastErr error

	for {
		last, lastErr = readABSState(ctx)
		if lastErr == nil && matchesABSState(last, want) {
			return last
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if lastErr != nil {
		t.Fatalf("read ABS state for %s/%s: %v", ctx.instance.name, ctx.library.name, lastErr)
	}
	t.Fatalf(
		"ABS state did not match for %s/%s\nexpected count: %d\nactual count: %d\nexpected missing: %d\nactual missing: %d\nactive contains: %v\nmissing contains: %v\nabsent contains: %v\nitems:\n%s",
		ctx.instance.name,
		ctx.library.name,
		want.expectedCount,
		len(last.items),
		want.missingCount,
		len(last.missingPaths),
		want.activeContains,
		want.missingContains,
		want.absentContains,
		formatABSItems(last.items),
	)
	return absLibraryState{}
}

func cleanMissingABSItems(t *testing.T, ctx absScenarioContext, oldPathFragments []string) {
	t.Helper()

	state := waitForABSState(t, ctx, absStateExpectation{
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
	if err := ctx.client.RemoveLibraryItemsWithIssues(ctx.libraryID); err != nil {
		t.Fatalf("remove ABS library items with issues: %v", err)
	}
}

func readABSState(ctx absScenarioContext) (absLibraryState, error) {
	items, err := ctx.client.GetAllLibraryItems(ctx.libraryID)
	if err != nil {
		return absLibraryState{}, err
	}

	state := absLibraryState{
		items: items,
	}
	for _, item := range items {
		state.allPaths = append(state.allPaths, item.Path)
		if item.IsMissing {
			state.missingPaths = append(state.missingPaths, item.Path)
		} else {
			state.activePaths = append(state.activePaths, item.Path)
		}
	}
	return state, nil
}

func matchesABSState(state absLibraryState, want absStateExpectation) bool {
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

func formatABSItems(items []abs.LibraryItem) string {
	var rows []string
	for _, item := range items {
		rows = append(
			rows,
			fmt.Sprintf(
				"id=%s missing=%t path=%s relPath=%s",
				item.ID,
				item.IsMissing,
				item.Path,
				item.RelPath,
			),
		)
	}
	return strings.Join(rows, "\n")
}

func containsAnyFragment(path string, fragments []string) bool {
	for _, fragment := range fragments {
		if strings.Contains(path, fragment) {
			return true
		}
	}
	return false
}

func loadABSTestingEnv(t *testing.T) {
	t.Helper()
	data, err := os.ReadFile(pathFromRoot("test", "abs", ".env.testing"))
	if err != nil {
		t.Fatalf("read ABS testing env: %v", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("invalid ABS testing env line: %q", line)
		}
		if err := os.Setenv(strings.TrimSpace(key), strings.TrimSpace(value)); err != nil {
			t.Fatalf("set env %s: %v", key, err)
		}
	}
}
