package terminalimage

import "strings"

func renderANSILogo() string {
	return strings.Join(ansiLogoLines(true), "\n")
}

func renderASCIILogo() string {
	return strings.Join(ansiLogoLines(false), "\n")
}

func ansiLogoLines(color bool) []string {
	lines := []string{
		"  .----------------------------------------------------------------.",
		"  |     _   _   _   ____  ___  ___  ____   ___   ___  _  __    |",
		"  |    / \\ | | | | |  _ \\|_ _|/ _ \\| __ ) / _ \\ / _ \\| |/ /    |",
		"  |   / _ \\| | | | | | | || || | | |  _ \\| | | | | | | ' /     |",
		"  |  / ___ \\ |_| | | |_| || || |_| | |_) | |_| | |_| | . \\     |",
		"  | /_/   \\_\\___/  |____/___|\\___/|____/ \\___/ \\___/|_|\\_\\    |",
		"  '----------------------------------------------------------------'",
		"                         -- organizer --",
		"",
	}
	if !color {
		return lines
	}
	return []string{
		ansiMagenta("  .----------------------------------------------------------------."),
		ansiMagenta(
			"  |",
		) + ansiPink(
			"     _   _   _   ____  ___  ___  ____   ___   ___  _  __    ",
		) + ansiMagenta(
			"|",
		),
		ansiMagenta(
			"  |",
		) + ansiPink(
			"    / \\ | | | | |  _ \\|_ _|/ _ \\| __ ) / _ \\ / _ \\| |/ /    ",
		) + ansiMagenta(
			"|",
		),
		ansiMagenta(
			"  |",
		) + ansiPink(
			"   / _ \\| | | | | | | || || | | |  _ \\| | | | | | | ' /     ",
		) + ansiMagenta(
			"|",
		),
		ansiMagenta(
			"  |",
		) + ansiPink(
			"  / ___ \\ |_| | | |_| || || |_| | |_) | |_| | |_| | . \\     ",
		) + ansiMagenta(
			"|",
		),
		ansiMagenta(
			"  |",
		) + ansiPink(
			" /_/   \\_\\___/  |____/___|\\___/|____/ \\___/ \\___/|_|\\_\\    ",
		) + ansiMagenta(
			"|",
		),
		ansiMagenta("  '----------------------------------------------------------------'"),
		ansiCyan("                         -- organizer --"),
		"",
	}
}

func ansiPink(value string) string {
	return "\x1b[38;5;207m" + value + ansiReset()
}

func ansiMagenta(value string) string {
	return "\x1b[38;5;93m" + value + ansiReset()
}

func ansiCyan(value string) string {
	return "\x1b[38;5;45m" + value + ansiReset()
}

func ansiReset() string {
	return "\x1b[0m"
}
