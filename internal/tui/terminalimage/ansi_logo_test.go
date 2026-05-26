package terminalimage

import (
	"regexp"
	"strings"
	"testing"
)

var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestRenderANSILogoUsesColorAndStaysCompact(t *testing.T) {
	logo := renderANSILogo()
	if !strings.Contains(logo, "\x1b[") {
		t.Fatal("renderANSILogo() missing ANSI escapes")
	}

	for _, line := range strings.Split(strings.TrimSuffix(logo, "\n"), "\n") {
		visible := ansiEscapePattern.ReplaceAllString(line, "")
		if width := len([]rune(visible)); width > 80 {
			t.Fatalf("ANSI logo line width = %d, want <= 80: %q", width, visible)
		}
	}
}

func TestRenderANSILogoVariantsUseSoftPalettes(t *testing.T) {
	if len(ansiLogoVariants) < 8 {
		t.Fatalf(
			"ansiLogoVariants has %d variants, want several rotating options",
			len(ansiLogoVariants),
		)
	}

	seen256 := false
	seen16 := false
	for i, variant := range ansiLogoVariants {
		logo := renderANSILogoVariant(i)
		if !strings.Contains(logo, "ORGANIZER") {
			t.Fatalf("variant %q missing ORGANIZER subtitle", variant.name)
		}
		if !strings.Contains(logo, "\x1b[") {
			t.Fatalf("variant %q missing ANSI escapes", variant.name)
		}
		if strings.Contains(logo, "\x1b[48;5;") {
			t.Fatalf("variant %q should not paint a background field", variant.name)
		}
		if strings.Contains(logo, "\x1b[38;5;") {
			seen256 = true
		} else {
			seen16 = true
		}

		for _, line := range strings.Split(strings.TrimSuffix(logo, "\n"), "\n") {
			visible := ansiEscapePattern.ReplaceAllString(line, "")
			if width := len([]rune(visible)); width > 80 {
				t.Fatalf("variant %q line width = %d, want <= 80: %q", variant.name, width, visible)
			}
		}
	}
	if !seen256 {
		t.Fatal("ANSI logo variants did not include a 256-color option")
	}
	if !seen16 {
		t.Fatal("ANSI logo variants did not include a 16-color option")
	}
}

func TestRenderASCIILogoUsesPlainText(t *testing.T) {
	logo := renderASCIILogo()
	if strings.Contains(logo, "\x1b[") {
		t.Fatal("renderASCIILogo() must not include ANSI escapes")
	}

	for _, r := range logo {
		if r == '\n' {
			continue
		}
		if r < 32 || r > 126 {
			t.Fatalf("renderASCIILogo() includes non-ASCII rune %q", r)
		}
	}
}
