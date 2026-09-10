//go:build abs_e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/abs"
)

func TestABSReviewPaginationSQLiteAndMappings(t *testing.T) {
	resetAndInitialScan(t)
	ctx := newABSScenarioContext(t, plainInstance, booksLibrary)
	fixture, err := os.ReadFile(
		filepath.Join(repoRootPath, "testdata", "epub", "title-author.epub"),
	)
	if err != nil {
		t.Fatal(err)
	}
	libraryRoot := localLibraryPath(plainInstance, booksLibrary)
	for i := 0; i < 101; i++ {
		folder := filepath.Join(libraryRoot, "pagination", fmt.Sprintf("book-%03d", i))
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "book.epub"), fixture, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	scanLibraryAndWait(t, ctx)
	state := waitForABSState(t, ctx, absStateExpectation{expectedCount: 104, missingCount: 0})
	ids := map[string]bool{}
	for _, item := range state.items {
		if ids[item.ID] {
			t.Fatalf("duplicate item %s", item.ID)
		}
		ids[item.ID] = true
	}
	output := runOrganizer(
		t,
		"abs",
		"scan",
		"--abs-url",
		os.Getenv(plainInstance.envURL),
		"--abs-token",
		os.Getenv("ABS_TOKEN"),
		"--abs-library",
		ctx.libraryID,
		"--abs-path-map",
		pathMap(booksLibrary, plainInstance),
		"--dir",
		libraryRoot,
		"--check-files",
	)
	assertOutputContains(t, output, "Found 104 items")

	// Read the live database without stopping ABS or replacing its state.
	dbPath := filepath.Join(
		repoRootPath,
		"test",
		"abs",
		"state",
		"plain",
		"config",
		"absdatabase.sqlite",
	)
	folders, err := abs.ListLibraries(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(folders) != 2 {
		t.Fatalf("library folders = %+v", folders)
	}
	output = runOrganizer(
		t,
		"abs",
		"test-paths",
		"--abs-sqlite",
		dbPath,
		"--dir",
		"/books/pagination",
	)
	assertOutputContains(t, output, "Path discovery SUCCESS!", "ABS: /books -> Local: /books")

	mapper := abs.NewPathMapper([]abs.PathMapping{
		{ABSPrefix: "/books", LocalPrefix: libraryRoot},
		{ABSPrefix: "/books/pagination", LocalPrefix: filepath.Join(libraryRoot, "pagination")},
	})
	for _, item := range state.items {
		local := mapper.ToLocal(item.Path)
		if _, err := os.Stat(local); err != nil {
			t.Fatalf("mapped item is not on disk: %s: %v", local, err)
		}
		if got := mapper.ToABS(local); got != item.Path {
			t.Fatalf("round trip: %s vs %s", got, item.Path)
		}
	}
	if got := mapper.ToLocal("/books-other/book.epub"); got != "/books-other/book.epub" {
		t.Fatalf("unrelated library remapped: %s", got)
	}
}
