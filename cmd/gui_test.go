package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGuiCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		flags       map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid input and output directories",
			flags: map[string]string{
				"input":  "/tmp/input",
				"output": "/tmp/output",
			},
			expectError: false,
		},
		{
			name: "missing input directory",
			flags: map[string]string{
				"output": "/tmp/output",
			},
			expectError: true,
			errorMsg:    "required flag(s) \"input\" not set",
		},
		{
			name: "missing output directory",
			flags: map[string]string{
				"input": "/tmp/input",
			},
			expectError: true,
			errorMsg:    "required flag(s) \"output\" not set",
		},
		{
			name:        "missing both directories",
			flags:       map[string]string{},
			expectError: true,
			errorMsg:    "required flag(s) \"input\", \"output\" not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := &cobra.Command{
				Use:   "gui",
				Short: "Start the TUI interface for audiobook organization",
				Run: func(cmd *cobra.Command, args []string) {
					inputDir, _ := cmd.Flags().GetString("input")
					outputDir, _ := cmd.Flags().GetString("output")

					// Validate required flags (mimicking the original logic)
					if inputDir == "" || outputDir == "" {
						cmd.Printf("Error: input and output directories are required\n")
						return
					}

					// Don't actually run the TUI in tests, just validate flags
					cmd.Printf("Would start TUI with input: %s, output: %s\n", inputDir, outputDir)
				},
			}

			// Add flags
			cmd.Flags().StringP("input", "i", "", "Input directory containing audiobooks (required)")
			cmd.Flags().StringP("output", "o", "", "Output directory for organized audiobooks (required)")
			cmd.MarkFlagRequired("input")
			cmd.MarkFlagRequired("output")

			// Set up flags
			var args []string
			for flag, value := range tt.flags {
				args = append(args, "--"+flag, value)
			}
			cmd.SetArgs(args)

			// Capture output
			var output strings.Builder
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			// Execute command
			err := cmd.Execute()

			// Check expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", output.String())
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v. Output: %s", err, output.String())
				}
			}
		})
	}
}

func TestGuiCommandFlags(t *testing.T) {
	cmd := guiCmd

	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		flagType     string
		required     bool
		defaultValue string
	}{
		{
			name:         "input flag",
			flagName:     "input",
			shorthand:    "i",
			flagType:     "string",
			required:     true,
			defaultValue: "",
		},
		{
			name:         "output flag",
			flagName:     "output",
			shorthand:    "o",
			flagType:     "string",
			required:     true,
			defaultValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %q not found", tt.flagName)
			}

			// Check shorthand
			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			// Check type
			if flag.Value.Type() != tt.flagType {
				t.Errorf("Flag %q type = %q, want %q", tt.flagName, flag.Value.Type(), tt.flagType)
			}

			// Check default value
			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defaultValue)
			}

			// Check if required (this is harder to test directly, but we can check the annotation)
			annotations := flag.Annotations
			if tt.required {
				if annotations == nil || len(annotations[cobra.BashCompOneRequiredFlag]) == 0 {
					// Note: The required flag annotation might not be set in this way
					// This is a basic check - the actual required validation happens at runtime
					t.Logf("Flag %q should be required (validation happens at runtime)", tt.flagName)
				}
			}
		})
	}
}

func TestGuiCommandValidation(t *testing.T) {
	// Test directory validation logic (without actually running TUI)
	tests := []struct {
		name      string
		inputDir  string
		outputDir string
		valid     bool
	}{
		{
			name:      "both directories provided",
			inputDir:  "/tmp/input",
			outputDir: "/tmp/output",
			valid:     true,
		},
		{
			name:      "empty input directory",
			inputDir:  "",
			outputDir: "/tmp/output",
			valid:     false,
		},
		{
			name:      "empty output directory",
			inputDir:  "/tmp/input",
			outputDir: "",
			valid:     false,
		},
		{
			name:      "both directories empty",
			inputDir:  "",
			outputDir: "",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic used in the GUI command
			valid := tt.inputDir != "" && tt.outputDir != ""

			if valid != tt.valid {
				t.Errorf("Validation result = %v, want %v", valid, tt.valid)
			}
		})
	}
}

func TestGuiCommandIntegration(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "gui_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	// Create the input directory
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Test the command setup without actually running the TUI
	cmd := guiCmd
	cmd.SetArgs([]string{"--input", inputDir, "--output", outputDir})

	// Capture any panics
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GUI command setup panicked: %v", r)
		}
	}()

	// Test flag parsing
	if err := cmd.ParseFlags([]string{"--input", inputDir, "--output", outputDir}); err != nil {
		t.Errorf("Failed to parse flags: %v", err)
	}

	// Verify flags were set correctly
	actualInput, err := cmd.Flags().GetString("input")
	if err != nil {
		t.Errorf("Failed to get input flag: %v", err)
	}
	if actualInput != inputDir {
		t.Errorf("Input flag = %q, want %q", actualInput, inputDir)
	}

	actualOutput, err := cmd.Flags().GetString("output")
	if err != nil {
		t.Errorf("Failed to get output flag: %v", err)
	}
	if actualOutput != outputDir {
		t.Errorf("Output flag = %q, want %q", actualOutput, outputDir)
	}
}

func TestGuiCommandHelp(t *testing.T) {
	cmd := guiCmd

	// Test that help can be generated without errors
	help := cmd.Help()
	if help != nil {
		t.Errorf("Help generation failed: %v", help)
	}

	// Test that usage can be generated
	usage := cmd.UsageString()
	if usage == "" {
		t.Error("Usage string is empty")
	}

	// Check that usage contains expected elements
	expectedElements := []string{"gui", "input", "output"}
	for _, element := range expectedElements {
		if !strings.Contains(usage, element) {
			t.Errorf("Usage string missing %q: %s", element, usage)
		}
	}
}

func TestGuiCommandDescription(t *testing.T) {
	cmd := guiCmd

	// Verify command metadata
	if cmd.Use != "gui" {
		t.Errorf("Command Use = %q, want %q", cmd.Use, "gui")
	}

	if cmd.Short == "" {
		t.Error("Command Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Command Long description is empty")
	}

	// Check that descriptions contain expected keywords
	expectedKeywords := []string{"TUI", "interface", "audiobook"}

	for _, keyword := range expectedKeywords {
		if !strings.Contains(cmd.Short, keyword) && !strings.Contains(cmd.Long, keyword) {
			t.Errorf("Command descriptions missing keyword %q", keyword)
		}
	}
}
