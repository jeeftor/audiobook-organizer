package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBetaVersionIgnoresPrereleaseTags(t *testing.T) {
	repo := newGitRepo(t)
	git(t, repo, "tag", "v0.11.0")
	git(t, repo, "tag", "v0.11.0-beta.cd9c813")
	git(t, repo, "tag", "v0.11.0-beta.cd9c813-beta.707d4f7")

	output := runBetaVersion(t, repo, "707d4f7", "develop")

	assertOutputLine(t, output, "version", "v0.11.0-beta.707d4f7")
	assertOutputLine(t, output, "release_name", "v0.11.0-beta (develop)")
	assertOutputLine(t, output, "branch", "develop")
	if strings.Contains(output["version"], "beta.cd9c813-beta") {
		t.Fatalf("version contains nested beta suffix: %q", output["version"])
	}
}

func TestBetaVersionUsesNewestStableSemverTag(t *testing.T) {
	repo := newGitRepo(t)
	git(t, repo, "tag", "v0.9.0")
	git(t, repo, "tag", "v0.10.0")
	git(t, repo, "tag", "v0.10.1-beta.old")
	git(t, repo, "tag", "v0.11.0")

	output := runBetaVersion(t, repo, "abc1234", "release/beta-sync")

	assertOutputLine(t, output, "version", "v0.11.0-beta.abc1234")
	assertOutputLine(t, output, "release_name", "v0.11.0-beta (release-beta-sync)")
	assertOutputLine(t, output, "branch", "release-beta-sync")
}

func TestBetaVersionFallsBackWithoutStableTags(t *testing.T) {
	repo := newGitRepo(t)
	git(t, repo, "tag", "v0.1.0-beta.old")

	output := runBetaVersion(t, repo, "abc1234", "develop")

	assertOutputLine(t, output, "version", "v0.0.0-beta.abc1234")
	assertOutputLine(t, output, "release_name", "v0.0.0-beta (develop)")
}

func newGitRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	git(t, repo, "init")
	git(t, repo, "config", "user.email", "test@example.invalid")
	git(t, repo, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("test\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	git(t, repo, "add", "README.md")
	git(t, repo, "commit", "-m", "initial")
	return repo
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func runBetaVersion(t *testing.T, repo, shortSHA, branchName string) map[string]string {
	t.Helper()

	script, err := filepath.Abs("beta-version.sh")
	if err != nil {
		t.Fatalf("resolve script path: %v", err)
	}
	cmd := exec.Command("sh", script)
	cmd.Dir = repo
	cmd.Env = append(
		os.Environ(),
		"BETA_SHORT_SHA="+shortSHA,
		"BETA_BRANCH_NAME="+branchName,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run beta-version.sh: %v\n%s", err, output)
	}
	return parseOutput(t, string(output))
}

func parseOutput(t *testing.T, output string) map[string]string {
	t.Helper()

	values := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("output line %q does not contain key=value", line)
		}
		values[key] = value
	}
	return values
}

func assertOutputLine(t *testing.T, output map[string]string, key, want string) {
	t.Helper()

	if got := output[key]; got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}
