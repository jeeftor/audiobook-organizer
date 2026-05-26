package terminalimage

import (
	"os"
	"strings"
)

const (
	noImagesEnv      = "AUDIOBOOK_ORGANIZER_NO_TERMINAL_IMAGES"
	imageProtocolEnv = "AUDIOBOOK_ORGANIZER_TERMINAL_IMAGE_PROTOCOL"
)

// ImageProtocol names the terminal logo renderer selected for startup logos.
type ImageProtocol string

const (
	ProtocolOff        ImageProtocol = "off"
	ProtocolAuto       ImageProtocol = "auto"
	ProtocolKitty      ImageProtocol = "kitty"
	ProtocolITerm2     ImageProtocol = "iterm2"
	ProtocolSixel      ImageProtocol = "sixel"
	ProtocolANSI       ImageProtocol = "ansi"
	ProtocolASCII      ImageProtocol = "ascii"
	ProtocolHalfblocks ImageProtocol = "halfblocks"
)

type protocolDetectionConfig struct {
	getenv      func(string) string
	interactive func() bool
}

// DetectTerminalImageProtocol detects whether the current terminal should render startup logos.
func DetectTerminalImageProtocol() ImageProtocol {
	return detectTerminalImageProtocol(protocolDetectionConfig{})
}

func detectTerminalImageProtocol(cfg protocolDetectionConfig) ImageProtocol {
	getenv := cfg.getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	interactive := cfg.interactive
	if interactive == nil {
		interactive = isInteractiveTTY
	}
	if getenv(noImagesEnv) != "" {
		return ProtocolOff
	}
	rawOverride := strings.TrimSpace(getenv(imageProtocolEnv))
	if rawOverride != "" {
		override := normalizeProtocol(rawOverride)
		if override != ProtocolAuto {
			return override
		}
	}
	if !interactive() {
		return ProtocolOff
	}
	if getenv("CI") != "" {
		return ProtocolOff
	}

	// tmux image passthrough varies by terminal and configuration. Fall back to
	// text art unless the user explicitly opts into a native protocol above.
	if getenv("TMUX") != "" {
		return detectTextProtocol(getenv)
	}

	return detectAutoProtocol(getenv)
}

func detectAutoProtocol(getenv func(string) string) ImageProtocol {
	if strings.EqualFold(getenv("TERM"), "dumb") {
		return ProtocolASCII
	}
	return detectTextProtocol(getenv)
}

func detectTextProtocol(getenv func(string) string) ImageProtocol {
	if supportsANSI(getenv) {
		return ProtocolANSI
	}
	return ProtocolASCII
}

func supportsANSI(getenv func(string) string) bool {
	if getenv("NO_COLOR") != "" {
		return false
	}
	if getenv("TERM_PROGRAM") != "" ||
		getenv("LC_TERMINAL") != "" ||
		getenv("TERM_SESSION_ID") != "" ||
		getenv("COLORTERM") != "" ||
		getenv("TMUX") != "" {
		return true
	}

	term := strings.ToLower(getenv("TERM"))
	return strings.Contains(term, "xterm") ||
		strings.Contains(term, "color") ||
		strings.Contains(term, "ansi") ||
		strings.Contains(term, "screen") ||
		strings.Contains(term, "tmux") ||
		strings.Contains(term, "rxvt")
}

func normalizeProtocol(value string) ImageProtocol {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ProtocolAuto):
		return ProtocolAuto
	case string(ProtocolOff), "none", "false", "0":
		return ProtocolOff
	case string(ProtocolKitty):
		return ProtocolKitty
	case string(ProtocolITerm2), "iterm":
		return ProtocolITerm2
	case string(ProtocolSixel):
		return ProtocolSixel
	case string(ProtocolANSI):
		return ProtocolANSI
	case string(ProtocolASCII):
		return ProtocolASCII
	case string(ProtocolHalfblocks), "halfblock":
		return ProtocolHalfblocks
	default:
		return ProtocolOff
	}
}
