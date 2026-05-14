package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"

	"github.com/jeeftor/audiobook-organizer/internal/app"
	"github.com/jeeftor/audiobook-organizer/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the local browser-based web UI",
	Long: `Start the Audiobook Organizer local web UI.

The web UI serves from the same binary and binds to localhost by default.
It exposes local API endpoints for scan, preview, rename, and Audiobookshelf
configuration while reusing the existing organizer and ABS packages.`,
	RunE: runWeb,
}

func init() {
	addWebFlags(webCmd)
	rootCmd.AddCommand(webCmd)
}

func addWebFlags(cmd *cobra.Command) {
	cmd.Flags().String("host", "127.0.0.1", "Host interface for the local web UI")
	cmd.Flags().Int("port", 0, "Port for the local web UI (0 chooses an available port)")
	cmd.Flags().Bool("open", true, "Open the web UI in the default browser")
	cmd.Flags().Bool("no-open", false, "Do not open the web UI in the default browser")
}

func runWeb(cmd *cobra.Command, args []string) error {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	openBrowser, _ := cmd.Flags().GetBool("open")
	noOpen, _ := cmd.Flags().GetBool("no-open")
	if noOpen {
		openBrowser = false
	}

	inputDir := firstNonEmpty(viper.GetString("input"), viper.GetString("dir"))
	outputDir := firstNonEmpty(viper.GetString("output"), viper.GetString("out"))

	token, err := newSessionToken()
	if err != nil {
		return err
	}

	webConfig := app.DefaultWebConfig(host, port, openBrowser, inputDir, outputDir)
	service := app.NewService(webConfig)
	webServer, err := server.New(server.Config{Token: token}, service)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	url := server.URL(host, listener, token)

	fmt.Fprintf(cmd.OutOrStdout(), "Audiobook Organizer web UI running at:\n%s\n", url)
	if openBrowser {
		if err := openURL(url); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not open browser automatically: %v\n", err)
		}
	}

	return webServer.Serve(context.Background(), listener)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func newSessionToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("creating web session token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func openURL(url string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		command = "open"
		args = []string{url}
	case "windows":
		command = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		command = "xdg-open"
		args = []string{url}
	}
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
