package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// InstallMethod represents how the binary was installed
type InstallMethod int

const (
	InstallMethodUnknown InstallMethod = iota
	InstallMethodHomebrew
	InstallMethodApt
	InstallMethodYum
	InstallMethodApk
	InstallMethodBinary
	InstallMethodGoInstall
)

var (
	checkOnly bool
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update audiobook-organizer to the latest version",
	Long: `Check for and install the latest version of audiobook-organizer from GitHub releases.

This command will:
  - Check for the latest release on GitHub
  - Compare with your current version
  - Download and install the update (if not using --check)

Examples:
  # Check for updates without installing
  audiobook-organizer update --check

  # Update to the latest version
  audiobook-organizer update
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
}

// runUpdate checks for and optionally installs updates
func runUpdate() error {
	fmt.Println("🔍 Checking for updates...")

	// Detect installation method
	installMethod := detectInstallMethod()
	fmt.Printf("📍 Installation method detected: %s\n\n", installMethod)

	// Fetch latest release from GitHub API
	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Get current version
	currentVersionStr := buildVersion
	if currentVersionStr == "" || currentVersionStr == "dev" {
		fmt.Println("⚠️  Development build detected - cannot determine current version")
		fmt.Printf("Latest available version: %s\n", release.TagName)
		fmt.Printf("Visit: %s\n", release.HTMLURL)
		return nil
	}

	// Parse versions (remove 'v' prefix if present)
	currentVersionStr = strings.TrimPrefix(currentVersionStr, "v")
	latestVersionStr := strings.TrimPrefix(release.TagName, "v")

	currentVersion, err := semver.Parse(currentVersionStr)
	if err != nil {
		return fmt.Errorf("failed to parse current version %q: %w", currentVersionStr, err)
	}

	latestVersion, err := semver.Parse(latestVersionStr)
	if err != nil {
		return fmt.Errorf("failed to parse latest version %q: %w", latestVersionStr, err)
	}

	// Compare versions
	if latestVersion.LTE(currentVersion) {
		fmt.Printf("✅ Already up to date (v%s)\n", currentVersion)
		return nil
	}

	// Show update available
	fmt.Printf("\n📦 Update available!\n")
	fmt.Printf("   Current version: v%s\n", currentVersion)
	fmt.Printf("   Latest version:  v%s\n", latestVersion)
	fmt.Printf("   Release URL:     %s\n", release.HTMLURL)

	// If check-only mode, stop here
	if checkOnly {
		fmt.Println("\n💡 Run 'audiobook-organizer update' to install the update")
		return nil
	}

	// Perform update based on installation method
	switch installMethod {
	case InstallMethodHomebrew:
		return updateViaHomebrew()

	case InstallMethodApt:
		return updateViaApt(latestVersion.String())

	case InstallMethodYum:
		return updateViaYum()

	case InstallMethodApk:
		return updateViaApk()

	case InstallMethodBinary, InstallMethodGoInstall, InstallMethodUnknown:
		// Use self-update for binary installs
		return updateViaSelfUpdate(release, latestVersion)

	default:
		return fmt.Errorf("unsupported installation method: %s", installMethod)
	}
}

// fetchLatestRelease fetches the latest release from GitHub
func fetchLatestRelease() (*GitHubRelease, error) {
	url := "https://api.github.com/repos/jeeftor/audiobook-organizer/releases/latest"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

// findAssetForPlatform finds the correct download asset for the current platform
func findAssetForPlatform(release *GitHubRelease) (string, error) {
	// Determine the expected asset name based on OS and architecture
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch names to release arch names
	archMap := map[string]string{
		"amd64": "x86_64",
		"386":   "i386",
		"arm64": "arm64",
		"arm":   "armv6", // or armv7, we'll try both
	}

	releaseArch, ok := archMap[arch]
	if !ok {
		releaseArch = arch
	}

	// Capitalize OS name for Darwin/Linux/Windows
	osMap := map[string]string{
		"darwin":  "Darwin",
		"linux":   "Linux",
		"windows": "Windows",
	}

	releaseOS, ok := osMap[osName]
	if !ok {
		releaseOS = osName
	}

	// Expected filename patterns:
	// audiobook-organizer_Darwin_x86_64.tar.gz
	// audiobook-organizer_Linux_x86_64.tar.gz
	// audiobook-organizer_Windows_x86_64.zip

	extension := ".tar.gz"
	if osName == "windows" {
		extension = ".zip"
	}

	expectedName := fmt.Sprintf("audiobook-organizer_%s_%s%s", releaseOS, releaseArch, extension)

	// Search for the asset
	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return asset.BrowserDownloadURL, nil
		}
	}

	// If exact match not found, list available assets
	var assetNames []string
	for _, asset := range release.Assets {
		assetNames = append(assetNames, asset.Name)
	}

	return "", fmt.Errorf("no compatible asset found for %s/%s (expected: %s)\nAvailable assets: %v",
		osName, arch, expectedName, assetNames)
}

// detectInstallMethod determines how the binary was installed
func detectInstallMethod() InstallMethod {
	exe, err := os.Executable()
	if err != nil {
		return InstallMethodUnknown
	}

	// Check for Homebrew (macOS/Linux)
	if strings.Contains(exe, "/Cellar/") || strings.Contains(exe, "/homebrew/") {
		return InstallMethodHomebrew
	}

	// Check if managed by package manager
	if runtime.GOOS == "linux" {
		// Check dpkg (Debian/Ubuntu)
		if _, err := os.Stat("/var/lib/dpkg/info/audiobook-organizer.list"); err == nil {
			return InstallMethodApt
		}

		// Check rpm (RedHat/CentOS/Fedora)
		if _, err := os.Stat("/var/lib/rpm"); err == nil {
			// Simple check - if rpm database exists, assume yum/dnf system
			return InstallMethodYum
		}

		// Check apk (Alpine)
		if _, err := os.Stat("/etc/apk"); err == nil {
			if _, err := os.Stat("/lib/apk/db/installed"); err == nil {
				return InstallMethodApk
			}
		}
	}

	// Check if installed via go install (typically in GOPATH/bin or GOBIN)
	if strings.Contains(exe, "/go/bin/") || strings.Contains(exe, "go/bin/") {
		return InstallMethodGoInstall
	}

	// Default to binary install
	return InstallMethodBinary
}

// String returns the name of the install method
func (m InstallMethod) String() string {
	switch m {
	case InstallMethodHomebrew:
		return "Homebrew"
	case InstallMethodApt:
		return "APT (Debian/Ubuntu)"
	case InstallMethodYum:
		return "YUM/DNF (RedHat/Fedora)"
	case InstallMethodApk:
		return "APK (Alpine)"
	case InstallMethodBinary:
		return "Binary Install"
	case InstallMethodGoInstall:
		return "Go Install"
	default:
		return "Unknown"
	}
}

// updateViaHomebrew updates via Homebrew
func updateViaHomebrew() error {
	fmt.Println("\n🍺 Updating via Homebrew...")

	cmd := exec.Command("brew", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run 'brew update': %w", err)
	}

	cmd = exec.Command("brew", "upgrade", "audiobook-organizer")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run 'brew upgrade': %w", err)
	}

	fmt.Println("\n✅ Successfully updated via Homebrew!")
	return nil
}

// updateViaApt updates via APT (Debian/Ubuntu)
func updateViaApt(version string) error {
	fmt.Println("\n📦 APT package manager detected")
	fmt.Println("⚠️  Automated APT updates require sudo privileges")
	fmt.Println("\nTo update, please run:")
	fmt.Println("  sudo apt-get update")
	fmt.Println("  sudo apt-get install --only-upgrade audiobook-organizer")
	fmt.Println("\nOr download the .deb package manually from:")
	fmt.Printf("  https://github.com/jeeftor/audiobook-organizer/releases/latest\n")

	return nil
}

// updateViaYum updates via YUM/DNF (RedHat/Fedora)
func updateViaYum() error {
	fmt.Println("\n📦 YUM/DNF package manager detected")
	fmt.Println("⚠️  Automated YUM updates require sudo privileges")
	fmt.Println("\nTo update, please run:")
	fmt.Println("  sudo yum update audiobook-organizer")
	fmt.Println("  # or on newer systems:")
	fmt.Println("  sudo dnf update audiobook-organizer")
	fmt.Println("\nOr download the .rpm package manually from:")
	fmt.Printf("  https://github.com/jeeftor/audiobook-organizer/releases/latest\n")

	return nil
}

// updateViaApk updates via APK (Alpine)
func updateViaApk() error {
	fmt.Println("\n📦 APK package manager detected")
	fmt.Println("⚠️  Automated APK updates require root privileges")
	fmt.Println("\nTo update, please run:")
	fmt.Println("  sudo apk update")
	fmt.Println("  sudo apk upgrade audiobook-organizer")
	fmt.Println("\nOr download the .apk package manually from:")
	fmt.Printf("  https://github.com/jeeftor/audiobook-organizer/releases/latest\n")

	return nil
}

// updateViaSelfUpdate performs a self-update for binary installations
func updateViaSelfUpdate(release *GitHubRelease, latestVersion semver.Version) error {
	// Find the correct asset for this platform
	assetURL, err := findAssetForPlatform(release)
	if err != nil {
		return err
	}

	// Perform the update
	fmt.Printf("\n⬇️  Downloading update from: %s\n", assetURL)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(assetURL, exe); err != nil {
		return fmt.Errorf("failed to update binary: %w", err)
	}

	fmt.Printf("\n✅ Successfully updated to v%s!\n", latestVersion)
	fmt.Println("🔄 Please restart the application to use the new version")

	return nil
}
