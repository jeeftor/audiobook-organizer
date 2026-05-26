package terminalimage

import (
	"strings"
	"testing"
)

func TestDetectTerminalImageProtocolGuards(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		interactive bool
		want        ImageProtocol
	}{
		{
			name:        "non interactive",
			interactive: false,
			want:        ProtocolOff,
		},
		{
			name:        "ci",
			env:         map[string]string{"CI": "true"},
			interactive: true,
			want:        ProtocolOff,
		},
		{
			name:        "dumb terminal gets ascii fallback",
			env:         map[string]string{"TERM": "dumb"},
			interactive: true,
			want:        ProtocolASCII,
		},
		{
			name:        "opt out",
			env:         map[string]string{noImagesEnv: "1"},
			interactive: true,
			want:        ProtocolOff,
		},
		{
			name:        "tmux defaults to ansi fallback",
			env:         map[string]string{"TMUX": "/tmp/tmux-501/default,123,0"},
			interactive: true,
			want:        ProtocolANSI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTerminalImageProtocol(protocolDetectionConfig{
				getenv:      envGetter(tt.env),
				interactive: func() bool { return tt.interactive },
			})

			if got != tt.want {
				t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectTerminalImageProtocolAutoUsesTextFallbacks(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want ImageProtocol
	}{
		{name: "xterm color", env: map[string]string{"TERM": "xterm-256color"}, want: ProtocolANSI},
		{name: "kitty terminal", env: map[string]string{"TERM": "xterm-kitty"}, want: ProtocolANSI},
		{
			name: "iterm terminal",
			env:  map[string]string{"TERM_PROGRAM": "iTerm.app"},
			want: ProtocolANSI,
		},
		{
			name: "truecolor terminal",
			env:  map[string]string{"COLORTERM": "truecolor"},
			want: ProtocolANSI,
		},
		{name: "plain vt terminal", env: map[string]string{"TERM": "vt52"}, want: ProtocolASCII},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTerminalImageProtocol(protocolDetectionConfig{
				getenv:      envGetter(tt.env),
				interactive: func() bool { return true },
			})

			if got != tt.want {
				t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectTerminalImageProtocolUnsupportedWithoutANSIGetsASCII(t *testing.T) {
	got := detectTerminalImageProtocol(protocolDetectionConfig{
		getenv:      envGetter(map[string]string{"TERM": "vt52"}),
		interactive: func() bool { return true },
	})

	if got != ProtocolASCII {
		t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, ProtocolASCII)
	}
}

func TestDetectTerminalImageProtocolNoColorUsesASCIIFallback(t *testing.T) {
	got := detectTerminalImageProtocol(protocolDetectionConfig{
		getenv: envGetter(map[string]string{
			"NO_COLOR": "1",
			"TERM":     "xterm-256color",
		}),
		interactive: func() bool { return true },
	})

	if got != ProtocolASCII {
		t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, ProtocolASCII)
	}
}

func TestDetectTerminalImageProtocolIgnoresITerm2GraphicsInAutoMode(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "term program",
			env:  map[string]string{"TERM_PROGRAM": "iTerm.app"},
		},
		{
			name: "lc terminal",
			env:  map[string]string{"LC_TERMINAL": "iTerm2"},
		},
		{
			name: "session id",
			env: map[string]string{
				"TERM_SESSION_ID": "w0t0p2:FA6177F0-D137-4EFF-811C-C92B9A1EB526",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTerminalImageProtocol(protocolDetectionConfig{
				getenv:      envGetter(tt.env),
				interactive: func() bool { return true },
			})

			if got != ProtocolANSI {
				t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, ProtocolANSI)
			}
		})
	}
}

func TestDetectTerminalImageProtocolOverride(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  ImageProtocol
	}{
		{name: "off", value: "off", want: ProtocolOff},
		{name: "kitty", value: "kitty", want: ProtocolKitty},
		{name: "iterm2 alias", value: "iterm", want: ProtocolITerm2},
		{name: "sixel", value: "sixel", want: ProtocolSixel},
		{name: "ansi", value: "ansi", want: ProtocolANSI},
		{name: "ascii", value: "ascii", want: ProtocolASCII},
		{name: "halfblocks", value: "halfblocks", want: ProtocolHalfblocks},
		{name: "invalid", value: "unknown", want: ProtocolOff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTerminalImageProtocol(protocolDetectionConfig{
				getenv:      envGetter(map[string]string{imageProtocolEnv: tt.value}),
				interactive: func() bool { return true },
			})

			if got != tt.want {
				t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectTerminalImageProtocolForcedOverrideBypassesTTYGuard(t *testing.T) {
	got := detectTerminalImageProtocol(protocolDetectionConfig{
		getenv:      envGetter(map[string]string{imageProtocolEnv: "ansi"}),
		interactive: func() bool { return false },
	})

	if got != ProtocolANSI {
		t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, ProtocolANSI)
	}
}

func TestDetectTerminalImageProtocolAutoOverrideUsesTextFallback(t *testing.T) {
	got := detectTerminalImageProtocol(protocolDetectionConfig{
		getenv: envGetter(map[string]string{
			imageProtocolEnv: "auto",
			"TERM_PROGRAM":   "iTerm.app",
		}),
		interactive: func() bool { return true },
	})

	if got != ProtocolANSI {
		t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, ProtocolANSI)
	}
}

func TestDetectTerminalImageProtocolForcedOverrideWinsOverITerm2Environment(t *testing.T) {
	got := detectTerminalImageProtocol(protocolDetectionConfig{
		getenv: envGetter(map[string]string{
			imageProtocolEnv: "kitty",
			"TERM_PROGRAM":   "iTerm.app",
		}),
		interactive: func() bool { return true },
	})

	if got != ProtocolKitty {
		t.Fatalf("detectTerminalImageProtocol() = %q, want %q", got, ProtocolKitty)
	}
}

func TestStartupLogoViewDoesNotRenderWhenOff(t *testing.T) {
	logo := NewStartupLogo(ProtocolOff)
	logo.render = func(ImageProtocol, int, int) (string, error) {
		t.Fatal("renderer should not be called when protocol is off")
		return "", nil
	}

	if got := logo.View(); got != "" {
		t.Fatalf("View() = %q, want empty fallback", got)
	}
}

func TestStartupLogoViewCachesRenderedLogo(t *testing.T) {
	logo := NewStartupLogo(ProtocolKitty)
	calls := 0
	logo.render = func(protocol ImageProtocol, width, height int) (string, error) {
		calls++
		if protocol != ProtocolKitty {
			t.Fatalf("protocol = %q, want %q", protocol, ProtocolKitty)
		}
		if height != defaultLogoHeightCells {
			t.Fatalf("height = %d, want %d", height, defaultLogoHeightCells)
		}
		return "rendered-logo", nil
	}

	if got := logo.View(); got != "rendered-logo" {
		t.Fatalf("View() = %q, want rendered logo", got)
	}
	if got := logo.View(); got != "rendered-logo" {
		t.Fatalf("second View() = %q, want cached rendered logo", got)
	}
	if calls != 1 {
		t.Fatalf("renderer called %d times, want 1", calls)
	}
}

func TestStartupLogoViewWithReservedSpace(t *testing.T) {
	logo := NewStartupLogo(ProtocolITerm2)
	logo.height = 3
	logo.render = func(ImageProtocol, int, int) (string, error) {
		return "rendered-logo", nil
	}

	want := "rendered-logo" + strings.Repeat("\n", nativeLogoReservedRows)
	if got := logo.ViewWithReservedSpace(); got != want {
		t.Fatalf("ViewWithReservedSpace() = %q, want rendered logo plus reserved lines", got)
	}
}

func TestStartupLogoViewWithReservedSpaceANSI(t *testing.T) {
	logo := NewStartupLogo(ProtocolANSI)
	logo.height = 3
	logo.render = func(ImageProtocol, int, int) (string, error) {
		return "rendered-ansi", nil
	}

	if got := logo.ViewWithReservedSpace(); got != "rendered-ansi" {
		t.Fatalf("ViewWithReservedSpace() = %q, want no extra reserved lines", got)
	}
}

func TestStartupLogoViewWithReservedSpaceASCII(t *testing.T) {
	logo := NewStartupLogo(ProtocolASCII)
	logo.height = 3
	logo.render = func(ImageProtocol, int, int) (string, error) {
		return "rendered-ascii", nil
	}

	if got := logo.ViewWithReservedSpace(); got != "rendered-ascii" {
		t.Fatalf("ViewWithReservedSpace() = %q, want no extra reserved lines", got)
	}
}

func TestStartupLogoViewWithReservedSpaceHalfblocks(t *testing.T) {
	logo := NewStartupLogo(ProtocolHalfblocks)
	logo.height = 3
	logo.render = func(ImageProtocol, int, int) (string, error) {
		return "rendered-halfblocks", nil
	}

	if got := logo.ViewWithReservedSpace(); got != "rendered-halfblocks" {
		t.Fatalf("ViewWithReservedSpace() = %q, want no extra reserved lines", got)
	}
}

func TestStartupLogoViewBesideTextNativeProtocol(t *testing.T) {
	logo := NewStartupLogo(ProtocolKitty)
	logo.render = func(ImageProtocol, int, int) (string, error) {
		return "rendered-logo", nil
	}

	got := logo.ViewBesideText("Usage:\n  audiobook-organizer")
	want := "rendered-logo" + strings.Repeat("\n", nativeLogoReservedRows) +
		"Usage:\n  audiobook-organizer"
	if got != want {
		t.Fatalf("ViewBesideText() = %q, want %q", got, want)
	}
}

func TestStartupLogoViewBesideTextTextProtocol(t *testing.T) {
	logo := NewStartupLogo(ProtocolANSI)
	logo.render = func(ImageProtocol, int, int) (string, error) {
		return "rendered-ansi", nil
	}

	got := logo.ViewBesideText("Usage:")
	want := "rendered-ansi\nUsage:"
	if got != want {
		t.Fatalf("ViewBesideText() = %q, want %q", got, want)
	}
}

func envGetter(env map[string]string) func(string) string {
	return func(key string) string {
		return env[key]
	}
}
