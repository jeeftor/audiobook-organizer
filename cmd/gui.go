package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// guiCmd represents the gui command
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Information about the desktop GUI application",
	Long: `The audiobook-organizer-gui is a separate desktop application with a graphical interface.

It provides the same organization features as the CLI/TUI but with a visual interface
for selecting directories, previewing changes, and configuring field mappings.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🖥️  Audiobook Organizer GUI")
		fmt.Println("==========================")
		fmt.Println()
		fmt.Println("The GUI is a separate desktop application available from GitHub releases.")
		fmt.Println()
		fmt.Println("📥 Download:")
		fmt.Println("   https://github.com/jeeftor/audiobook-organizer/releases")
		fmt.Println()
		fmt.Printf("   Look for: audiobook-organizer-gui_%s_%s\n", osName(), archName())
		fmt.Println()
		fmt.Println("📦 Installation:")
		switch runtime.GOOS {
		case "darwin":
			fmt.Println("   1. Download the .tar.gz for your architecture")
			fmt.Println("   2. Extract: tar -xzf audiobook-organizer-gui_Darwin_*.tar.gz")
			fmt.Println("   3. Move to Applications or run directly")
			fmt.Println()
			fmt.Println("   Note: You may need to allow the app in System Settings > Privacy & Security")
		case "windows":
			fmt.Println("   1. Download the .zip for your architecture")
			fmt.Println("   2. Extract the zip file")
			fmt.Println("   3. Run audiobook-organizer-gui.exe")
		case "linux":
			fmt.Println("   1. Download the .tar.gz for your architecture")
			fmt.Println("   2. Extract: tar -xzf audiobook-organizer-gui_Linux_*.tar.gz")
			fmt.Println("   3. Run: ./audiobook-organizer-gui")
		default:
			fmt.Println("   1. Download the appropriate archive for your platform")
			fmt.Println("   2. Extract and run the binary")
		}
		fmt.Println()
		fmt.Println("💡 Tip: Use 'audiobook-organizer tui' for an interactive terminal interface")
		fmt.Println("        that doesn't require a separate download.")
	},
}

func osName() string {
	switch runtime.GOOS {
	case "darwin":
		return "Darwin"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}

func archName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "arm64"
	case "386":
		return "i386"
	default:
		return runtime.GOARCH
	}
}

func init() {
	rootCmd.AddCommand(guiCmd)
}
