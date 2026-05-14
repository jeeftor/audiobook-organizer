//go:build abs_e2e

package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

var (
	repoRootPath string
	binaryPath   string
	binaryTmpDir string
)

func TestMain(m *testing.M) {
	releaseLock, err := acquireABSRunLock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "acquire ABS E2E run lock: %v\n", err)
		os.Exit(1)
	}

	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "find repo root: %v\n", err)
		releaseLock()
		os.Exit(1)
	}
	repoRootPath = root

	binaryTmpDir, err = os.MkdirTemp("", "aobook-org-abs-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		releaseLock()
		os.Exit(1)
	}
	binaryPath = filepath.Join(binaryTmpDir, "audiobook-organizer")

	if output, err := runCommand(root, 2*time.Minute, "go", "build", "-o", binaryPath, "."); err != nil {
		fmt.Fprintf(os.Stderr, "build test binary: %v\n%s\n", err, output)
		os.RemoveAll(binaryTmpDir)
		releaseLock()
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(binaryTmpDir)
	releaseLock()
	os.Exit(code)
}

func acquireABSRunLock() (func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:23378")
	if err != nil {
		return nil, fmt.Errorf(
			"another ABS E2E run appears to be active; stop it before running this matrix again: %w",
			err,
		)
	}
	return func() {
		_ = listener.Close()
	}, nil
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}

func resetAndInitialScan(t *testing.T) {
	t.Helper()
	output, err := runCommand(repoRootPath, 8*time.Minute, "make", "abs-dev-reset-scan")
	if err != nil {
		t.Fatalf("reset and initial ABS scan failed: %v\n%s", err, output)
	}
}

func runOrganizer(t *testing.T, args ...string) string {
	t.Helper()
	output, err := runCommand(repoRootPath, 2*time.Minute, binaryPath, args...)
	if err != nil {
		t.Fatalf(
			"organizer command failed: %v\ncommand: %s %s\n%s",
			err,
			binaryPath,
			strings.Join(args, " "),
			output,
		)
	}
	return output
}

func rescanABS(t *testing.T) {
	t.Helper()
	output, err := runCommand(repoRootPath, 4*time.Minute, "test/abs/scripts/scan-libraries.sh")
	if err != nil {
		t.Fatalf("ABS rescan failed: %v\n%s", err, output)
	}
}

func runCommand(dir string, timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(
		os.Environ(),
		"ABS_ENV_FILE=test/abs/.env.testing",
		"NO_COLOR=1",
		"TERM=dumb",
	)

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("%s timed out after %s", name, timeout)
	}
	return string(output), err
}

func pathFromRoot(parts ...string) string {
	all := append([]string{repoRootPath}, parts...)
	return filepath.Join(all...)
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path to exist: %s\nstat error: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected path to be absent: %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("expected path to be absent: %s\nstat error: %v", path, err)
	}
}

func assertLibraryStable(
	t *testing.T,
	dbPath string,
	folderPath string,
	expectedCount int,
	wantContains []string,
	wantAbsent []string,
) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Minute)
	var last libraryState
	var err error

	for {
		last, err = readLibraryState(dbPath, folderPath)
		if err != nil {
			t.Fatalf("read ABS library state: %v", err)
		}

		if last.Count == expectedCount && last.Missing == 0 &&
			containsAll(last.Paths, wantContains) && containsNone(last.Paths, wantAbsent) {
			return
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf(
		"ABS library %s did not stabilize\nexpected count: %d\nactual count: %d\nmissing: %d\nwanted path fragments: %v\nabsent path fragments: %v\npaths:\n%s",
		folderPath,
		expectedCount,
		last.Count,
		last.Missing,
		wantContains,
		wantAbsent,
		strings.Join(last.Paths, "\n"),
	)
}

func assertLibraryContains(
	t *testing.T,
	dbPath string,
	folderPath string,
	wantContains []string,
) libraryState {
	t.Helper()

	deadline := time.Now().Add(2 * time.Minute)
	var state libraryState
	var err error

	for {
		state, err = readLibraryState(dbPath, folderPath)
		if err != nil {
			t.Fatalf("read ABS library state: %v", err)
		}

		if containsAll(state.Paths, wantContains) {
			if state.Missing > 0 {
				t.Logf(
					"ABS library %s has %d missing item(s) after scan; post-move reconciliation is still pending",
					folderPath,
					state.Missing,
				)
			}
			return state
		}

		if time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf(
		"ABS library %s does not contain expected paths\nwanted path fragments: %v\nmissing count: %d\npaths:\n%s",
		folderPath,
		wantContains,
		state.Missing,
		strings.Join(state.Paths, "\n"),
	)
	return libraryState{}
}

type libraryState struct {
	Count   int
	Missing int
	Paths   []string
}

func readLibraryState(dbPath string, folderPath string) (libraryState, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return libraryState{}, err
	}
	defer db.Close()

	var libraryID string
	if err := db.QueryRow(`
		select l.id
		from libraries l
		join libraryFolders f on f.libraryId = l.id
		where f.path = ?
	`, folderPath).Scan(&libraryID); err != nil {
		return libraryState{}, fmt.Errorf("lookup library %s: %w", folderPath, err)
	}

	var state libraryState
	if err := db.QueryRow(`select count(*), coalesce(sum(isMissing), 0) from libraryItems where libraryId = ?`, libraryID).
		Scan(&state.Count, &state.Missing); err != nil {
		return libraryState{}, fmt.Errorf("count items for %s: %w", folderPath, err)
	}

	rows, err := db.Query(
		`select path from libraryItems where libraryId = ? order by path`,
		libraryID,
	)
	if err != nil {
		return libraryState{}, fmt.Errorf("list paths for %s: %w", folderPath, err)
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return libraryState{}, err
		}
		state.Paths = append(state.Paths, path)
	}
	if err := rows.Err(); err != nil {
		return libraryState{}, err
	}

	return state, nil
}

func containsAll(paths []string, fragments []string) bool {
	for _, fragment := range fragments {
		found := false
		for _, path := range paths {
			if strings.Contains(path, fragment) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func containsNone(paths []string, fragments []string) bool {
	for _, fragment := range fragments {
		for _, path := range paths {
			if strings.Contains(path, fragment) {
				return false
			}
		}
	}
	return true
}
