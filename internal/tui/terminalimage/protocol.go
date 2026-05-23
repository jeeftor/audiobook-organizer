package terminalimage

import (
	"os"
	"strings"

	"github.com/blacktop/go-termimg"
)

const (
	noImagesEnv      = "AUDIOBOOK_ORGANIZER_NO_TERMINAL_IMAGES"
	imageProtocolEnv = "AUDIOBOOK_ORGANIZER_TERMINAL_IMAGE_PROTOCOL"
)

// ImageProtocol names the terminal image protocol selected for startup logos.
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
	detect      func() termimg.Protocol
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
	detect := cfg.detect
	if detect == nil {
		detect = termimg.DetectProtocol
	}

	if getenv(noImagesEnv) != "" {
		return ProtocolOff
	}
	if !interactive() {
		return ProtocolOff
	}
	if getenv("CI") != "" {
		return ProtocolOff
	}

	rawOverride := strings.TrimSpace(getenv(imageProtocolEnv))
	if rawOverride != "" {
		override := normalizeProtocol(rawOverride)
		if override == ProtocolAuto {
			return detectAutoProtocol(getenv, detect)
		}
		return override
	}

	// tmux image passthrough varies by terminal and configuration. Fall back to
	// text art unless the user explicitly opts into a native protocol above.
	if getenv("TMUX") != "" {
		return detectTextProtocol(getenv)
	}

	return detectAutoProtocol(getenv, detect)
}

func detectAutoProtocol(getenv func(string) string, detect func() termimg.Protocol) ImageProtocol {
	if isITerm2Environment(getenv) {
		return ProtocolITerm2
	}
	if strings.EqualFold(getenv("TERM"), "dumb") {
		return ProtocolASCII
	}

	detected := detect()
	if detected == termimg.Halfblocks {
		return detectTextProtocol(getenv)
	}
	if detected := detectProtocol(detected); detected != ProtocolOff {
		return detected
	}
	return detectTextProtocol(getenv)
}

func detectTextProtocol(getenv func(string) string) ImageProtocol {
	if supportsANSI(getenv) {
		return ProtocolANSI
	}
	return ProtocolASCII
}

func isITerm2Environment(getenv func(string) string) bool {
	if strings.EqualFold(getenv("TERM_PROGRAM"), "iTerm.app") {
		return true
	}
	if strings.EqualFold(getenv("LC_TERMINAL"), "iTerm2") {
		return true
	}

	termSessionID := getenv("TERM_SESSION_ID")
	return strings.HasPrefix(termSessionID, "w") && strings.Contains(termSessionID, ":")
}

func supportsANSI(getenv func(string) string) bool {
	if getenv("NO_COLOR") != "" {
		return false
	}
	if getenv("TERM_PROGRAM") != "" || getenv("COLORTERM") != "" || getenv("TMUX") != "" {
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

func detectProtocol(protocol termimg.Protocol) ImageProtocol {
	switch protocol {
	case termimg.Kitty:
		return ProtocolKitty
	case termimg.ITerm2:
		return ProtocolITerm2
	case termimg.Sixel:
		return ProtocolSixel
	default:
		return ProtocolOff
	}
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
