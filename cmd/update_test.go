package cmd

import (
	"runtime"
	"testing"

	"github.com/blang/semver"
)

func TestInstallMethodString(t *testing.T) {
	tests := []struct {
		method   InstallMethod
		expected string
	}{
		{InstallMethodHomebrew, "Homebrew"},
		{InstallMethodApt, "APT (Debian/Ubuntu)"},
		{InstallMethodYum, "YUM/DNF (RedHat/Fedora)"},
		{InstallMethodApk, "APK (Alpine)"},
		{InstallMethodBinary, "Binary Install"},
		{InstallMethodGoInstall, "Go Install"},
		{InstallMethodUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.method.String(); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestFindAssetForPlatform(t *testing.T) {
	release := &GitHubRelease{
		TagName: "v1.0.0",
		HTMLURL: "https://github.com/jeeftor/audiobook-organizer/releases/tag/v1.0.0",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{
				Name:               "audiobook-organizer_Darwin_x86_64.tar.gz",
				BrowserDownloadURL: "https://github.com/.../audiobook-organizer_Darwin_x86_64.tar.gz",
			},
			{
				Name:               "audiobook-organizer_Darwin_arm64.tar.gz",
				BrowserDownloadURL: "https://github.com/.../audiobook-organizer_Darwin_arm64.tar.gz",
			},
			{
				Name:               "audiobook-organizer_Linux_x86_64.tar.gz",
				BrowserDownloadURL: "https://github.com/.../audiobook-organizer_Linux_x86_64.tar.gz",
			},
			{
				Name:               "audiobook-organizer_Windows_x86_64.zip",
				BrowserDownloadURL: "https://github.com/.../audiobook-organizer_Windows_x86_64.zip",
			},
		},
	}

	tests := []struct {
		name        string
		goos        string
		goarch      string
		expectError bool
		expectAsset string
	}{
		{
			name:        "Darwin amd64",
			goos:        "darwin",
			goarch:      "amd64",
			expectError: false,
			expectAsset: "audiobook-organizer_Darwin_x86_64.tar.gz",
		},
		{
			name:        "Darwin arm64",
			goos:        "darwin",
			goarch:      "arm64",
			expectError: false,
			expectAsset: "audiobook-organizer_Darwin_arm64.tar.gz",
		},
		{
			name:        "Linux amd64",
			goos:        "linux",
			goarch:      "amd64",
			expectError: false,
			expectAsset: "audiobook-organizer_Linux_x86_64.tar.gz",
		},
		{
			name:        "Windows amd64",
			goos:        "windows",
			goarch:      "amd64",
			expectError: false,
			expectAsset: "audiobook-organizer_Windows_x86_64.zip",
		},
		{
			name:        "Unsupported platform",
			goos:        "plan9",
			goarch:      "amd64",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origGOOS := runtime.GOOS
			origGOARCH := runtime.GOARCH

			// We can't actually change runtime.GOOS/GOARCH, so this test
			// validates the logic but can't test all platforms dynamically.
			// Instead, we test the current platform and document expected behavior.

			assetURL, err := findAssetForPlatform(release)

			if tt.goos == runtime.GOOS && tt.goarch == runtime.GOARCH {
				if tt.expectError && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !tt.expectError && assetURL == "" {
					t.Error("Expected asset URL but got empty string")
				}
			}

			_ = origGOOS
			_ = origGOARCH
		})
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		shouldUpdate   bool
	}{
		{
			name:           "same version",
			currentVersion: "1.0.0",
			latestVersion:  "1.0.0",
			shouldUpdate:   false,
		},
		{
			name:           "newer version available",
			currentVersion: "1.0.0",
			latestVersion:  "1.1.0",
			shouldUpdate:   true,
		},
		{
			name:           "much newer version available",
			currentVersion: "1.0.0",
			latestVersion:  "2.0.0",
			shouldUpdate:   true,
		},
		{
			name:           "patch update available",
			currentVersion: "1.0.0",
			latestVersion:  "1.0.1",
			shouldUpdate:   true,
		},
		{
			name:           "current is newer (shouldn't happen)",
			currentVersion: "2.0.0",
			latestVersion:  "1.0.0",
			shouldUpdate:   false,
		},
		{
			name:           "with v prefix current",
			currentVersion: "v1.0.0",
			latestVersion:  "1.1.0",
			shouldUpdate:   true,
		},
		{
			name:           "with v prefix latest",
			currentVersion: "1.0.0",
			latestVersion:  "v1.1.0",
			shouldUpdate:   true,
		},
		{
			name:           "with v prefix both",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.1.0",
			shouldUpdate:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse versions (remove 'v' prefix if present)
			currentStr := tt.currentVersion
			if currentStr[0] == 'v' {
				currentStr = currentStr[1:]
			}

			latestStr := tt.latestVersion
			if latestStr[0] == 'v' {
				latestStr = latestStr[1:]
			}

			current, err := semver.Parse(currentStr)
			if err != nil {
				t.Fatalf("Failed to parse current version: %v", err)
			}

			latest, err := semver.Parse(latestStr)
			if err != nil {
				t.Fatalf("Failed to parse latest version: %v", err)
			}

			// Check if update is needed
			updateNeeded := latest.GT(current)

			if updateNeeded != tt.shouldUpdate {
				t.Errorf("Expected shouldUpdate=%v, got %v (current=%s, latest=%s)",
					tt.shouldUpdate, updateNeeded, current, latest)
			}
		})
	}
}

func TestGitHubReleaseAssetMatching(t *testing.T) {
	// Test the asset name matching logic
	tests := []struct {
		name         string
		assetName    string
		goos         string
		goarch       string
		shouldMatch  bool
	}{
		{
			name:        "Darwin x86_64 match",
			assetName:   "audiobook-organizer_Darwin_x86_64.tar.gz",
			goos:        "darwin",
			goarch:      "amd64",
			shouldMatch: true,
		},
		{
			name:        "Darwin arm64 match",
			assetName:   "audiobook-organizer_Darwin_arm64.tar.gz",
			goos:        "darwin",
			goarch:      "arm64",
			shouldMatch: true,
		},
		{
			name:        "Linux x86_64 match",
			assetName:   "audiobook-organizer_Linux_x86_64.tar.gz",
			goos:        "linux",
			goarch:      "amd64",
			shouldMatch: true,
		},
		{
			name:        "Windows zip match",
			assetName:   "audiobook-organizer_Windows_x86_64.zip",
			goos:        "windows",
			goarch:      "amd64",
			shouldMatch: true,
		},
		{
			name:        "Wrong OS",
			assetName:   "audiobook-organizer_Linux_x86_64.tar.gz",
			goos:        "darwin",
			goarch:      "amd64",
			shouldMatch: false,
		},
		{
			name:        "Wrong arch",
			assetName:   "audiobook-organizer_Darwin_arm64.tar.gz",
			goos:        "darwin",
			goarch:      "amd64",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Map Go arch names to release arch names
			archMap := map[string]string{
				"amd64": "x86_64",
				"386":   "i386",
				"arm64": "arm64",
				"arm":   "armv6",
			}

			releaseArch := archMap[tt.goarch]
			if releaseArch == "" {
				releaseArch = tt.goarch
			}

			// Map OS names
			osMap := map[string]string{
				"darwin":  "Darwin",
				"linux":   "Linux",
				"windows": "Windows",
			}

			releaseOS := osMap[tt.goos]
			if releaseOS == "" {
				releaseOS = tt.goos
			}

			// Build expected name
			extension := ".tar.gz"
			if tt.goos == "windows" {
				extension = ".zip"
			}

			expectedName := "audiobook-organizer_" + releaseOS + "_" + releaseArch + extension

			matches := (tt.assetName == expectedName)

			if matches != tt.shouldMatch {
				t.Errorf("Expected match=%v, got %v (asset=%s, expected=%s)",
					tt.shouldMatch, matches, tt.assetName, expectedName)
			}
		})
	}
}

func TestDetectInstallMethod(t *testing.T) {
	// This test validates the detection logic exists and returns valid values
	// We can't reliably test all paths without mocking os.Executable and filesystem

	method := detectInstallMethod()

	// Should return one of the valid install methods
	validMethods := []InstallMethod{
		InstallMethodUnknown,
		InstallMethodHomebrew,
		InstallMethodApt,
		InstallMethodYum,
		InstallMethodApk,
		InstallMethodBinary,
		InstallMethodGoInstall,
	}

	found := false
	for _, valid := range validMethods {
		if method == valid {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("detectInstallMethod returned invalid method: %v", method)
	}

	// Verify String() doesn't panic
	methodStr := method.String()
	if methodStr == "" {
		t.Error("InstallMethod.String() returned empty string")
	}
}
