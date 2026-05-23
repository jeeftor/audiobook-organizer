package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blacktop/go-termimg"
	"github.com/jeeftor/audiobook-organizer/internal/tui/terminalimage"
	"github.com/spf13/cobra"
)

var termCmd = &cobra.Command{
	Use:    "term",
	Short:  "Print terminal image diagnostics",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		detected := termimg.DetectProtocol()
		protocols := termimg.DetermineProtocols()
		startupProtocol := terminalimage.DetectTerminalImageProtocol()

		fmt.Fprintln(out, "Terminal image diagnostics")
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
		fmt.Fprintf(out, "  TERM: %s\n", valueOrUnset(os.Getenv("TERM")))
		fmt.Fprintf(out, "  TERM_PROGRAM: %s\n", valueOrUnset(os.Getenv("TERM_PROGRAM")))
		fmt.Fprintf(
			out,
			"  TERM_PROGRAM_VERSION: %s\n",
			valueOrUnset(os.Getenv("TERM_PROGRAM_VERSION")),
		)
		fmt.Fprintf(out, "  TERM_SESSION_ID: %s\n", valueOrUnset(os.Getenv("TERM_SESSION_ID")))
		fmt.Fprintf(out, "  LC_TERMINAL: %s\n", valueOrUnset(os.Getenv("LC_TERMINAL")))
		fmt.Fprintf(out, "  TERM_FEATURES: %s\n", valueOrUnset(os.Getenv("TERM_FEATURES")))
		fmt.Fprintf(out, "  KITTY_WINDOW_ID: %s\n", valueOrUnset(os.Getenv("KITTY_WINDOW_ID")))
		fmt.Fprintf(out, "  WEZTERM_PANE: %s\n", valueOrUnset(os.Getenv("WEZTERM_PANE")))
		fmt.Fprintf(out, "  TMUX: %s\n", valueOrUnset(os.Getenv("TMUX")))
		fmt.Fprintf(out, "  stdin TTY: %t\n", fileIsCharDevice(os.Stdin))
		fmt.Fprintf(out, "  stdout TTY: %t\n", fileIsCharDevice(os.Stdout))
		fmt.Fprintf(out, "  termimg DetectProtocol: %s\n", detected)
		fmt.Fprintf(out, "  termimg DetermineProtocols: %s\n", formatTermProtocols(protocols))
		fmt.Fprintf(out, "  startup protocol: %s\n", startupProtocol)
		fmt.Fprintln(out)

		if startupProtocol == terminalimage.ProtocolANSI {
			printANSILogoOptions(out)
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
		if startupProtocol == terminalimage.ProtocolHalfblocks {
			fmt.Fprintln(
				out,
				"Halfblocks detected; image payload skipped while reviewing ANSI text logo prototypes.",
			)
			printRootHelpPreview(out, cmd.Root())
			return nil
		}

		rendered := terminalimage.NewStartupLogo(startupProtocol).ViewWithReservedSpace()
		fmt.Fprintf(out, "Rendered payload bytes: %d\n", len(rendered))
		fmt.Fprintf(out, "Rendered payload prefix: %q\n", firstRunes(rendered, 80))
		fmt.Fprintln(out, "Image output follows:")
		fmt.Fprint(out, rendered)
		fmt.Fprintln(out, "Image output ended.")
		printRootHelpPreview(out, cmd.Root())
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
	fmt.Fprint(out, terminalimage.NewStartupLogo(protocol).ViewWithReservedSpace())
	fmt.Fprint(out, root.UsageString())
}

type ansiLogoOption struct {
	name string
	body string
}

func printANSILogoOptions(out io.Writer) {
	fmt.Fprintln(out, "ANSI text logo prototypes")
	fmt.Fprintln(out)
	for i, option := range ansiLogoOptions() {
		fmt.Fprintf(out, "%sOption %02d: %s%s\n", ansiDim(), i+1, option.name, ansiReset())
		fmt.Fprintln(out, option.body)
		fmt.Fprintln(out)
	}
}

func ansiLogoOptions() []ansiLogoOption {
	return []ansiLogoOption{
		{
			name: "plasma board",
			body: ansiBG(16, strings.Repeat(" ", 68)) + "\n" +
				ansiBG(
					16,
					"  ",
				) + ansiFG(201, "█████  █   █ ████  ███  ████  ████  ████  ████  █  █") + ansiBG(16, "  ") + "\n" +
				ansiBG(
					16,
					"  ",
				) + ansiFG(207, "█   █  █   █ █   █  █  █    █ █   █ █    █ █    █ █ ") + ansiBG(16, "  ") + "\n" +
				ansiBG(
					16,
					"  ",
				) + ansiFG(93, "█████  █   █ █   █  █  █    █ ████  █    █ █    ██  ") + ansiBG(16, "  ") + "\n" +
				ansiBG(
					16,
					"  ",
				) + ansiFG(57, "█   █  █   █ █   █  █  █    █ █   █ █    █ █    █ █ ") + ansiBG(16, "  ") + "\n" +
				ansiBG(
					16,
					"  ",
				) + ansiFG(45, "█   █  █████ ████  ███  ████  ████  ████  ████  █  █") + ansiBG(16, "  ") + "\n" +
				ansiBG(53, "                       -- organizer --                       "),
		},
		{
			name: "ice draw",
			body: ansiFG(
				45,
				"╔══════════════════════════════════════════════════════════════╗",
			) + "\n" +
				ansiFG(
					45,
					"║",
				) + ansiFG(
				201,
				" ▄▄▄· ▄• ▄▌ ·▄▄▄▄  ▪        ▄▄▄▄·             ▄ •▄ ",
			) + ansiFG(
				45,
				" ║",
			) + "\n" +
				ansiFG(
					45,
					"║",
				) + ansiFG(
				207,
				"▐█ ▀█ █▪██▌██▪ ██ ██ ▪     ▐█ ▀█▪▪     ▪     █▌▄▌▪",
			) + ansiFG(
				45,
				" ║",
			) + "\n" +
				ansiFG(
					45,
					"║",
				) + ansiFG(
				93,
				"▄█▀▀█ █▌▐█▌▐█· ▐█▌▐█· ▄█▀▄ ▐█▀▀█▄ ▄█▀▄  ▄█▀▄ ▐▀▀▄·",
			) + ansiFG(
				45,
				" ║",
			) + "\n" +
				ansiFG(
					45,
					"║",
				) + ansiFG(
				57,
				"▐█ ▪▐▌▐█▄█▌██. ██ ▐█▌▐█▌.▐▌██▄▪▐█▐█▌.▐▌▐█▌.▐▌▐█.█▌",
			) + ansiFG(
				45,
				" ║",
			) + "\n" +
				ansiFG(
					45,
					"║",
				) + ansiFG(
				45,
				" ▀  ▀  ▀▀▀ ▀▀▀▀▀• ▀▀▀ ▀█▄▀▪·▀▀▀▀  ▀█▄▀▪ ▀█▄▀▪·▀  ▀",
			) + ansiFG(
				45,
				" ║",
			) + "\n" +
				ansiFG(
					45,
					"╚══════════════════════ -- organizer -- ══════════════════════╝",
				),
		},
		{
			name: "neon wall",
			body: ansiBG(
				17,
				"                                                                      ",
			) + "\n" +
				ansiBG(
					17,
					"  ",
				) + ansiBG(
				53,
				"  ",
			) + ansiFG(
				201,
				"  ██   ██ ██  ██ ████  ██  ████  ████   ████   ████  ██  ██  ",
			) + ansiBG(
				53,
				"  ",
			) + "\n" +
				ansiBG(
					17,
					"  ",
				) + ansiBG(
				53,
				"  ",
			) + ansiFG(
				207,
				" ████  ██  ██ ██  ██  ██  ██  ██ ██  ██ ██  ██ ██  ██ ██ ██   ",
			) + ansiBG(
				53,
				"  ",
			) + "\n" +
				ansiBG(
					17,
					"  ",
				) + ansiBG(
				53,
				"  ",
			) + ansiFG(
				93,
				"██  ██ ██  ██ ████   ██  ██  ██ ████  ██  ██ ██  ██ ███     ",
			) + ansiBG(
				53,
				"  ",
			) + "\n" +
				ansiBG(
					17,
					"  ",
				) + ansiBG(
				53,
				"  ",
			) + ansiFG(
				45,
				"██  ██  ████  ██    ████  ████  ████   ████   ████  ██ ██   ",
			) + ansiBG(
				53,
				"  ",
			) + "\n" +
				ansiBG(
					17,
					"                      ",
				) + ansiFG(
				45,
				"-- organizer --",
			) + ansiBG(
				17,
				"                      ",
			),
		},
		{
			name: "color blocks",
			body: ansiBlockRow(16, 53, 89, 125, 161, 197, 201, 207, 213, 219) + "\n" +
				ansiFG(231, "     AUDIOBOOK   AUDIOBOOK   AUDIOBOOK   AUDIOBOOK") + "\n" +
				ansiBlockRow(17, 18, 19, 20, 21, 27, 33, 39, 45, 51) + "\n" +
				ansiFG(45, "                   -- organizer --"),
		},
		{
			name: "old board ad",
			body: ansiBG(
				17,
				"                                                                  ",
			) + "\n" +
				ansiBG(
					17,
					"  ",
				) + ansiCyan(
				" A U D I O B O O K  //  A U D I O B O O K  //  AUDIOBOOK ",
			) + ansiBG(
				17,
				" ",
			) + "\n" +
				ansiBG(
					17,
					"  ",
				) + ansiPink(
				"   uploads sorted / series stacked / metadata cleaned    ",
			) + ansiBG(
				17,
				"  ",
			) + "\n" +
				ansiBG(
					17,
					"                                                                  ",
				) + "\n" +
				ansiGray(
					"                         -- organizer --",
				),
		},
		{
			name: "shade stack",
			body: ansiFG(
				201,
				"  ░█████╗ ██╗   ██╗██████╗ ██╗ ██████╗ ██████╗  ██████╗  ██████╗ ██╗  ██╗",
			) + "\n" +
				ansiFG(
					207,
					"  ██╔══██╗██║   ██║██╔══██╗██║██╔═══██╗██╔══██╗██╔═══██╗██╔═══██╗██║ ██╔╝",
				) + "\n" +
				ansiFG(
					93,
					"  ███████║██║   ██║██║  ██║██║██║   ██║██████╔╝██║   ██║██║   ██║█████╔╝ ",
				) + "\n" +
				ansiFG(
					57,
					"  ██╔══██║██║   ██║██║  ██║██║██║   ██║██╔══██╗██║   ██║██║   ██║██╔═██╗ ",
				) + "\n" +
				ansiFG(
					45,
					"  ██║  ██║╚██████╔╝██████╔╝██║╚██████╔╝██████╔╝╚██████╔╝╚██████╔╝██║  ██╗",
				) + "\n" +
				ansiFG(
					45,
					"                         -- organizer --",
				),
		},
		{
			name: "checker glass",
			body: ansiBG(
				53,
				"    ",
			) + ansiBG(
				16,
				"    ",
			) + ansiBG(
				53,
				"    ",
			) + ansiBG(
				16,
				"    ",
			) + ansiBG(
				53,
				"    ",
			) + ansiBG(
				16,
				"    ",
			) + ansiBG(
				53,
				"    ",
			) + "\n" +
				ansiFG(
					201,
					"  A U D I O B O O K",
				) + ansiFG(
				45,
				"  A U D I O B O O K",
			) + "\n" +
				ansiFG(
					207,
					"  ▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄▀▄",
				) + "\n" +
				ansiFG(
					45,
					"          -- organizer --",
				) + "\n" +
				ansiBG(
					16,
					"    ",
				) + ansiBG(
				53,
				"    ",
			) + ansiBG(
				16,
				"    ",
			) + ansiBG(
				53,
				"    ",
			) + ansiBG(
				16,
				"    ",
			) + ansiBG(
				53,
				"    ",
			) + ansiBG(
				16,
				"    ",
			),
		},
		{
			name: "cassette wall",
			body: ansiFG(45, "  ╔════════════════════════════════════════════════════╗") + "\n" +
				ansiFG(
					45,
					"  ║",
				) + ansiBG(17, "  ") + ansiFG(201, "AUDIOBOOK") + ansiBG(17, "          ") + ansiFG(207, "◉       ◉") + ansiBG(17, "          ") + ansiFG(201, "AUDIOBOOK") + ansiBG(17, "  ") + ansiFG(45, "║") + "\n" +
				ansiFG(
					45,
					"  ║",
				) + ansiBG(17, "  ") + ansiFG(93, "████████████████████████████████████████") + ansiBG(17, "  ") + ansiFG(45, "║") + "\n" +
				ansiFG(45, "  ╚════════════════════ -- organizer -- ══════════════╝"),
		},
		{
			name: "scanlines",
			body: ansiFG(201, " ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄") + "\n" +
				ansiFG(207, " █ A U D I O B O O K  ░▒▓  A U D I O B O O K  ▓▒░ █") + "\n" +
				ansiFG(93, " ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀") + "\n" +
				ansiFG(45, "                    -- organizer --"),
		},
		{
			name: "ansi logo candidate",
			body: terminalimage.NewStartupLogo(terminalimage.ProtocolANSI).ViewWithReservedSpace(),
		},
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

func formatTermProtocols(protocols []termimg.Protocol) string {
	if len(protocols) == 0 {
		return "<none>"
	}
	values := make([]string, 0, len(protocols))
	for _, protocol := range protocols {
		values = append(values, protocol.String())
	}
	return strings.Join(values, ", ")
}

func firstRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func ansiPink(value string) string {
	return "\x1b[38;5;207m" + value + ansiReset()
}

func ansiBlue(value string) string {
	return "\x1b[38;5;81m" + value + ansiReset()
}

func ansiCyan(value string) string {
	return "\x1b[38;5;45m" + value + ansiReset()
}

func ansiGray(value string) string {
	return "\x1b[38;5;245m" + value + ansiReset()
}

func ansiFG(color int, value string) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s%s", color, value, ansiReset())
}

func ansiBG(color int, value string) string {
	return fmt.Sprintf("\x1b[48;5;%dm%s%s", color, value, ansiReset())
}

func ansiBlockRow(colors ...int) string {
	var b strings.Builder
	for _, color := range colors {
		b.WriteString(ansiBG(color, "      "))
	}
	return b.String()
}

func ansiDim() string {
	return "\x1b[2m"
}

func ansiReset() string {
	return "\x1b[0m"
}
