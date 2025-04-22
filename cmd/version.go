package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// These variables will be set during the build using ldflags
var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildTime    = "unknown"
)

var shortOutput bool

// GetFormattedBuildTime returns the build time in a readable format
func GetFormattedBuildTime() string {
	if buildTime == "unknown" {
		return buildTime
	}

	// Try to parse the timestamp
	t, err := time.Parse(time.RFC3339, buildTime)
	if err != nil {
		return buildTime // Return original if parsing fails
	}

	return t.Format("2006-01-02 15:04:05 MST")
}

// GetDisplayVersion returns a formatted version string
// If we're in dev mode, it shows "dev (last release X.Y.Z)"
func GetDisplayVersion() string {
	// If we're in a release build, just return the build version
	if buildVersion != "dev" {
		return buildVersion
	}

	// We're in dev mode, try to find the last release tag
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	tagBytes, err := cmd.Output()

	if err == nil {
		tag := strings.TrimSpace(string(tagBytes))
		if tag != "" {
			return fmt.Sprintf("dev (last release %s)", tag)
		}
	}

	// Couldn't find a tag, just return dev
	return "dev"
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		displayVersion := GetDisplayVersion()

		if shortOutput {
			// For short output, just show the raw buildVersion for scripts
			fmt.Println(buildVersion)
			return
		}
		versionColor := color.New(color.FgCyan, color.Bold)
		buildColor := color.New(color.FgYellow)
		commitColor := color.New(color.FgGreen)
		osArchColor := color.New(color.FgMagenta)
		goVersionColor := color.New(color.FgRed)
		whiteColor := color.New(color.FgWhite)
		pathColor := color.New(color.FgBlue)

		whiteColor.Printf("Version: ")
		versionColor.Printf("%s\n", displayVersion)

		whiteColor.Printf("Built:   ")
		buildColor.Printf("%s\n", GetFormattedBuildTime())

		whiteColor.Printf("Commit:  ")
		commitColor.Printf("%s\n", buildCommit)

		whiteColor.Printf("OS/Arch: ")
		osArchColor.Printf("%s/%s\n", runtime.GOOS, runtime.GOARCH)

		whiteColor.Printf("Go:      ")
		goVersionColor.Printf("%s\n", runtime.Version())

		exe, err := os.Executable()
		exePath := "Unknown"
		if err == nil {
			exePath, _ = filepath.Abs(exe)
		}

		whiteColor.Printf("Binary:  ")
		pathColor.Printf("%s\n", exePath)
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&shortOutput, "short", "s", false, "Print only version number")
	rootCmd.AddCommand(versionCmd)
}
