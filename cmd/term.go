package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/jeeftor/audiobook-organizer/internal/tui/terminalimage"
	"github.com/spf13/cobra"
)

var termCmd = &cobra.Command{
	Use:    "term",
	Short:  "Print terminal image diagnostics",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		startupProtocol := terminalimage.DetectTerminalImageProtocol()

		fmt.Fprintln(out, "Terminal logo diagnostics")
		fmt.Fprintf(out, "  executable: %s\n", os.Args[0])
		fmt.Fprintf(
			out,
			"  override: %s\n",
			valueOrUnset(os.Getenv("AUDIOBOOK_ORGANIZER_TERMINAL_IMAGE_PROTOCOL")),
		)
		fmt.Fprintf(
			out,
			"  no images: %s\n",
			valueOrUnset(os.Getenv("AUDIOBOOK_ORGANIZER_NO_TERMINAL_IMAGES")),
		)
		fmt.Fprintf(out, "  NO_COLOR: %s\n", valueOrUnset(os.Getenv("NO_COLOR")))
		fmt.Fprintf(out, "  CI: %s\n", valueOrUnset(os.Getenv("CI")))
		fmt.Fprintf(out, "  TERM: %s\n", valueOrUnset(os.Getenv("TERM")))
		fmt.Fprintf(out, "  TERM_PROGRAM: %s\n", valueOrUnset(os.Getenv("TERM_PROGRAM")))
		fmt.Fprintf(
			out,
			"  TERM_PROGRAM_VERSION: %s\n",
			valueOrUnset(os.Getenv("TERM_PROGRAM_VERSION")),
		)
		fmt.Fprintf(out, "  TERM_SESSION_ID: %s\n", valueOrUnset(os.Getenv("TERM_SESSION_ID")))
		fmt.Fprintf(out, "  LC_TERMINAL: %s\n", valueOrUnset(os.Getenv("LC_TERMINAL")))
		fmt.Fprintf(out, "  COLORTERM: %s\n", valueOrUnset(os.Getenv("COLORTERM")))
		fmt.Fprintf(out, "  TERM_FEATURES: %s\n", valueOrUnset(os.Getenv("TERM_FEATURES")))
		fmt.Fprintf(out, "  TMUX: %s\n", valueOrUnset(os.Getenv("TMUX")))
		fmt.Fprintf(out, "  stdin TTY: %t\n", fileIsCharDevice(os.Stdin))
		fmt.Fprintf(out, "  stdout TTY: %t\n", fileIsCharDevice(os.Stdout))
		fmt.Fprintf(out, "  startup protocol: %s\n", startupProtocol)
		fmt.Fprintln(out)

		if startupProtocol == terminalimage.ProtocolANSI {
			printANSILogoSamples(out)
			printLogoHelpPreview(out, cmd.Root(), startupProtocol)
			return nil
		}
		if startupProtocol == terminalimage.ProtocolASCII {
			fmt.Fprintln(out, "ASCII fallback selected.")
			printLogoHelpPreview(out, cmd.Root(), startupProtocol)
			return nil
		}
		if startupProtocol == terminalimage.ProtocolOff {
			fmt.Fprintln(out, "No image protocol selected; not rendering.")
			printRootHelpPreview(out, cmd.Root())
			return nil
		}
		printLogoHelpPreview(out, cmd.Root(), startupProtocol)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(termCmd)
}

func printRootHelpPreview(out io.Writer, root *cobra.Command) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Root help preview")
	fmt.Fprintln(out)
	fmt.Fprint(out, root.UsageString())
}

func printLogoHelpPreview(
	out io.Writer,
	root *cobra.Command,
	protocol terminalimage.ImageProtocol,
) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Root help preview with selected logo")
	fmt.Fprintln(out)
	fmt.Fprint(out, terminalimage.NewStartupLogo(protocol).ViewBesideText(root.UsageString()))
}

func printANSILogoSamples(out io.Writer) {
	const samples = 4
	fmt.Fprintln(out, "ANSI rotating logo samples")
	fmt.Fprintln(out)
	for i := range samples {
		fmt.Fprintf(out, "%sSample %02d%s\n", ansiDim(), i+1, ansiReset())
		fmt.Fprintln(
			out,
			terminalimage.NewStartupLogo(
				terminalimage.ProtocolANSI,
			).ViewWithReservedSpace(),
		)
		fmt.Fprintln(out)
	}
}

func valueOrUnset(value string) string {
	if value == "" {
		return "<unset>"
	}
	return value
}

func fileIsCharDevice(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func ansiDim() string {
	return "\x1b[2m"
}

func ansiReset() string {
	return "\x1b[0m"
}
