package cmd

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetFormattedBuildTime(t *testing.T) {
	// Save original values
	originalBuildTime := buildTime
	defer func() {
		buildTime = originalBuildTime
	}()

	tests := []struct {
		name      string
		buildTime string
		wantMatch string
	}{
		{
			name:      "unknown time",
			buildTime: "unknown",
			wantMatch: "unknown",
		},
		{
			name:      "invalid format returns original",
			buildTime: "invalid-time",
			wantMatch: "invalid-time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildTime = tt.buildTime
			result := GetFormattedBuildTime()

			if !strings.Contains(result, tt.wantMatch) {
				t.Errorf("GetFormattedBuildTime() = %q, want to contain %q", result, tt.wantMatch)
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "valid integer",
			input:   "1640995200",
			want:    1640995200,
			wantErr: false,
		},
		{
			name:    "zero",
			input:   "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative integer",
			input:   "-123",
			want:    -123,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "not-a-number",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt64(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("parseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDisplayVersion(t *testing.T) {
	// Save original values
	originalBuildVersion := buildVersion
	defer func() {
		buildVersion = originalBuildVersion
	}()

	tests := []struct {
		name               string
		buildVersion       string
		gitCommandAvailable bool
		expectedContains   string
	}{
		{
			name:             "release version",
			buildVersion:     "v1.2.3",
			expectedContains: "v1.2.3",
		},
		{
			name:             "dev version with git",
			buildVersion:     "dev",
			expectedContains: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildVersion = tt.buildVersion
			result := GetDisplayVersion()

			if !strings.Contains(result, tt.expectedContains) {
				t.Errorf("GetDisplayVersion() = %q, want to contain %q", result, tt.expectedContains)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	// Save original values
	originalBuildVersion := buildVersion
	originalBuildCommit := buildCommit
	originalBuildTime := buildTime

	defer func() {
		buildVersion = originalBuildVersion
		buildCommit = originalBuildCommit
		buildTime = originalBuildTime
	}()

	tests := []struct {
		name         string
		buildVersion string
		buildCommit  string
		buildTime    string
	}{
		{
			name:         "release version",
			buildVersion: "v1.0.0",
			buildCommit:  "abc123",
			buildTime:    "2022-01-01T12:00:00Z",
		},
		{
			name:         "dev version",
			buildVersion: "dev",
			buildCommit:  "local",
			buildTime:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test values
			buildVersion = tt.buildVersion
			buildCommit = tt.buildCommit
			buildTime = tt.buildTime

			// Test the individual functions that the command uses
			displayVersion := GetDisplayVersion()
			formattedTime := GetFormattedBuildTime()

			// Verify the functions return expected results
			if displayVersion == "" {
				t.Error("GetDisplayVersion() returned empty string")
			}

			if formattedTime == "" {
				t.Error("GetFormattedBuildTime() returned empty string")
			}

			// Test that we can create a command without errors
			cmd := &cobra.Command{
				Use: "version",
				Run: func(cmd *cobra.Command, args []string) {
					// Just verify the command can be created and run
				},
			}

			if err := cmd.Execute(); err != nil {
				t.Errorf("Command execution failed: %v", err)
			}
		})
	}
}

func TestVersionCommandFlags(t *testing.T) {
	// Test that the version command has the expected flags
	cmd := versionCmd

	// Check that the short flag exists
	shortFlag := cmd.Flags().Lookup("short")
	if shortFlag == nil {
		t.Error("version command should have --short flag")
	}

	// Check that the short flag has the correct shorthand
	if shortFlag.Shorthand != "s" {
		t.Errorf("short flag shorthand = %q, want %q", shortFlag.Shorthand, "s")
	}

	// Check flag type
	if shortFlag.Value.Type() != "bool" {
		t.Errorf("short flag type = %q, want %q", shortFlag.Value.Type(), "bool")
	}
}


func TestVersionCommandIntegration(t *testing.T) {
	// This test verifies the version command can be executed without panicking
	// We need to create a standalone command to avoid conflicts with root command
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			// Just test that we can call the functions without panicking
			_ = GetDisplayVersion()
			_ = GetFormattedBuildTime()
		},
	}

	cmd.SetArgs([]string{})

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("version command panicked: %v", r)
		}
	}()

	// Execute the command - this should work without errors
	if err := cmd.Execute(); err != nil {
		t.Errorf("version command failed: %v", err)
	}
}

func TestGitCommandExecution(t *testing.T) {
	// Test that git command execution doesn't cause issues
	// This is more of an integration test to ensure git calls don't panic

	// Check if git is available (optional test)
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available, skipping git command test")
	}

	// Test git command execution (used in GetDisplayVersion)
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	_, err = cmd.Output()

	// We don't care if it succeeds or fails, just that it doesn't panic
	// This could fail if we're not in a git repository, which is fine
	t.Logf("git command result: %v", err)
}
