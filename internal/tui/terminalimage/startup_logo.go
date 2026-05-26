package terminalimage

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"strings"
	"sync"

	"github.com/blacktop/go-termimg"
)

const (
	defaultLogoWidthCells  = 20
	defaultLogoHeightCells = 5
	defaultLogoPixels      = 80
	nativeLogoReservedRows = defaultLogoHeightCells
)

// StartupLogo renders the embedded Audiobook Organizer logo for supported terminals.
type StartupLogo struct {
	protocol ImageProtocol
	width    int
	height   int
	render   func(ImageProtocol, int, int) (string, error)

	once sync.Once
	view string
}

// NewStartupLogo creates a startup logo renderer for the selected protocol.
func NewStartupLogo(protocol ImageProtocol) *StartupLogo {
	return &StartupLogo{
		protocol: protocol,
		width:    defaultLogoWidthCells,
		height:   defaultLogoHeightCells,
		render:   renderLogo,
	}
}

// NewAutoStartupLogo creates a startup logo renderer after detecting terminal support.
func NewAutoStartupLogo() *StartupLogo {
	return NewStartupLogo(DetectTerminalImageProtocol())
}

// View returns the rendered logo, or an empty string when images are disabled or fail.
func (l *StartupLogo) View() string {
	if l == nil || l.protocol == ProtocolOff {
		return ""
	}

	l.once.Do(func() {
		rendered, err := l.render(l.protocol, l.width, l.height)
		if err != nil {
			return
		}
		l.view = rendered
	})

	return l.view
}

// ViewWithReservedSpace returns the rendered logo followed by enough newlines
// for native image protocols that do not move the terminal cursor.
func (l *StartupLogo) ViewWithReservedSpace() string {
	view := l.View()
	if view == "" {
		return ""
	}
	if l.protocol == ProtocolANSI || l.protocol == ProtocolASCII ||
		l.protocol == ProtocolHalfblocks {
		return view
	}
	return view + strings.Repeat("\n", nativeLogoReservedRows)
}

// ViewBesideText returns the logo followed by text.
func (l *StartupLogo) ViewBesideText(text string) string {
	view := l.View()
	if view == "" {
		return text
	}
	if l.protocol == ProtocolANSI || l.protocol == ProtocolASCII ||
		l.protocol == ProtocolHalfblocks {
		return view + "\n" + text
	}
	return view + strings.Repeat("\n", nativeLogoReservedRows) + text
}

func renderLogo(protocol ImageProtocol, width, height int) (string, error) {
	if protocol == ProtocolANSI {
		return renderANSILogo(), nil
	}
	if protocol == ProtocolASCII {
		return renderASCIILogo(), nil
	}

	img, _, err := image.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		return "", fmt.Errorf("decode terminal logo: %w", err)
	}

	termProtocol, ok := termimgProtocol(protocol)
	if !ok {
		return "", nil
	}

	termImage := termimg.New(img).Protocol(termProtocol)
	if protocol == ProtocolHalfblocks {
		termImage = termImage.Width(width).Height(height)
	} else {
		termImage = termImage.WidthPixels(defaultLogoPixels).HeightPixels(defaultLogoPixels)
	}

	rendered, err := termImage.Render()
	if err != nil {
		return "", fmt.Errorf("render terminal logo: %w", err)
	}
	return rendered, nil
}

func termimgProtocol(protocol ImageProtocol) (termimg.Protocol, bool) {
	switch protocol {
	case ProtocolKitty:
		return termimg.Kitty, true
	case ProtocolITerm2:
		return termimg.ITerm2, true
	case ProtocolSixel:
		return termimg.Sixel, true
	case ProtocolHalfblocks:
		return termimg.Halfblocks, true
	default:
		return termimg.Unsupported, false
	}
}
