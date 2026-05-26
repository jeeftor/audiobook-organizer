package terminalimage

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type ansiLogoColorMode int

const (
	ansiLogoColor16 ansiLogoColorMode = iota
	ansiLogoColor256
)

type ansiLogoVariant struct {
	name       string
	colorMode  ansiLogoColorMode
	top        int
	body       int
	shadow     int
	highlight  int
	subtitle   int
	ramp       int
	accentCols map[int]struct{}
}

var ansiLogoLines = []string{
	"  ▄█▄  █   █ ████▄ █████ ▄██▄  ████▄  ▄██▄   ▄██▄  █  █▀",
	" █▀ ▀█ █   █ █   █   █  █▀  ▀█ █   █ █▀  ▀█ █▀  ▀█ █▄█▀ ",
	" █████ █   █ █   █   █  █    █ ████▄ █    █ █    █ ██▄  ",
	" █   █ █   █ █   █   █  █▄  ▄█ █   █ █▄  ▄█ █▄  ▄█ █ ▀▄ ",
	" ▀   ▀  ▀██▀ ████▀ █████ ▀██▀  ████▀  ▀██▀   ▀██▀  █  ▀▄",
	"             ▁▂▃▄ ORGANIZER ▄▃▂▁",
}

var ansiLogoVariants = []ansiLogoVariant{
	{
		name:      "dusty rose classic",
		colorMode: ansiLogoColor256,
		top:       218, body: 175, shadow: 132,
		highlight: 225, subtitle: 175, ramp: 132,
	},
	{
		name:      "dusty rose deeper",
		colorMode: ansiLogoColor256,
		top:       181, body: 139, shadow: 96,
		highlight: 218, subtitle: 145, ramp: 96,
	},
	{
		name:      "rose plum",
		colorMode: ansiLogoColor256,
		top:       182, body: 140, shadow: 96,
		highlight: 224, subtitle: 176, ramp: 132,
	},
	{
		name:      "plum glow classic",
		colorMode: ansiLogoColor256,
		top:       183, body: 141, shadow: 60,
		highlight: 225, subtitle: 141, ramp: 99,
	},
	{
		name:      "plum night",
		colorMode: ansiLogoColor256,
		top:       147, body: 103, shadow: 60,
		highlight: 189, subtitle: 147, ramp: 96,
	},
	{
		name:      "blueberry classic",
		colorMode: ansiLogoColor256,
		top:       117, body: 75, shadow: 25,
		highlight: 159, subtitle: 117, ramp: 99,
		accentCols: ansiLogoAccentCols(
			45, 46, 47, 48, 49, 50, 51, 52, 53, 54,
		),
	},
	{
		name:      "blueberry muted",
		colorMode: ansiLogoColor256,
		top:       110, body: 68, shadow: 24,
		highlight: 153, subtitle: 110, ramp: 97,
		accentCols: ansiLogoAccentCols(46, 47, 48, 49, 50, 51, 52),
	},
	{
		name:      "blueberry rose subtle",
		colorMode: ansiLogoColor256,
		top:       111, body: 74, shadow: 31,
		highlight: 153, subtitle: 110, ramp: 132,
		accentCols: ansiLogoAccentCols(
			42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
		),
	},
	{
		name:      "low contrast pink",
		colorMode: ansiLogoColor256,
		top:       181, body: 139, shadow: 96,
		highlight: 218, subtitle: 145, ramp: 96,
	},
	{
		name:      "low contrast mauve",
		colorMode: ansiLogoColor256,
		top:       181, body: 138, shadow: 95,
		highlight: 217, subtitle: 144, ramp: 95,
	},
	{
		name:      "smoky rose",
		colorMode: ansiLogoColor256,
		top:       174, body: 132, shadow: 89,
		highlight: 217, subtitle: 174, ramp: 95,
	},
	{
		name:      "soft 16 magenta",
		colorMode: ansiLogoColor16,
		top:       95, body: 35, shadow: 35,
		highlight: 97, subtitle: 95, ramp: 35,
	},
}

func renderANSILogo() string {
	return renderANSILogoVariant(randomANSILogoVariantIndex())
}

func renderANSILogoVariant(index int) string {
	if len(ansiLogoVariants) == 0 {
		return ""
	}
	if index < 0 || index >= len(ansiLogoVariants) {
		index = 0
	}
	variant := ansiLogoVariants[index]

	var rendered []string
	for row, line := range ansiLogoLines {
		var b strings.Builder
		var active ansiLogoSpan
		for col, char := range []rune(line) {
			if char == ' ' {
				if active.set {
					b.WriteString(ansiReset())
					active = ansiLogoSpan{}
				}
				b.WriteRune(char)
				continue
			}
			next := ansiLogoColorFor(variant, line, row, col, char)
			if active != next {
				b.WriteString(ansiLogoSGR(variant.colorMode, next.color, next.bold))
				active = next
			}
			b.WriteRune(char)
		}
		if active.set {
			b.WriteString(ansiReset())
		}
		rendered = append(rendered, b.String())
	}
	return strings.Join(rendered, "\n") + "\n"
}

func renderASCIILogo() string {
	return strings.Join([]string{
		"  .--------------------------------------------------------.",
		"  |      _    _   _ ____ ___ ___  ____   ___   ___  _  __ |",
		"  |     / \\  | | | |  _ \\_ _/ _ \\| __ ) / _ \\ / _ \\| |/ / |",
		"  |    / _ \\ | | | | | | | | | | |  _ \\| | | | | | | ' /  |",
		"  |   / ___ \\| |_| | |_| | | |_| | |_) | |_| | |_| | . \\  |",
		"  |  /_/   \\_\\\\___/|____/___\\___/|____/ \\___/ \\___/|_|\\_\\ |",
		"  '--------------------------------------------------------'",
		"                       -- organizer --",
		"",
	}, "\n")
}

type ansiLogoSpan struct {
	color int
	bold  bool
	set   bool
}

func ansiLogoColorFor(
	variant ansiLogoVariant,
	line string,
	row int,
	col int,
	char rune,
) ansiLogoSpan {
	if strings.Contains(line, "ORGANIZER") {
		if strings.ContainsRune("▁▂▃▄", char) {
			return ansiLogoSpan{color: variant.ramp, set: true}
		}
		return ansiLogoSpan{color: variant.subtitle, bold: true, set: true}
	}

	if _, ok := variant.accentCols[col]; ok && row >= 2 && row <= 4 {
		return ansiLogoSpan{color: variant.ramp, set: true}
	}
	switch row {
	case 0:
		if strings.ContainsRune("▄█", char) && col%9 <= 1 {
			return ansiLogoSpan{color: variant.highlight, bold: true, set: true}
		}
		return ansiLogoSpan{color: variant.top, bold: true, set: true}
	case 1, 2:
		if char == '▀' && col%11 == 0 {
			return ansiLogoSpan{color: variant.highlight, bold: true, set: true}
		}
		return ansiLogoSpan{color: variant.body, bold: row == 2, set: true}
	case 3:
		return ansiLogoSpan{color: variant.body, set: true}
	default:
		return ansiLogoSpan{color: variant.shadow, set: true}
	}
}

func ansiLogoSGR(mode ansiLogoColorMode, color int, bold bool) string {
	var codes []string
	if bold {
		codes = append(codes, "1")
	}
	if mode == ansiLogoColor16 {
		codes = append(codes, strconv.Itoa(color))
	} else {
		codes = append(codes, "38;5;"+strconv.Itoa(color))
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

func ansiLogoAccentCols(cols ...int) map[int]struct{} {
	lookup := make(map[int]struct{}, len(cols))
	for _, col := range cols {
		lookup[col] = struct{}{}
	}
	return lookup
}

func randomANSILogoVariantIndex() int {
	count := len(ansiLogoVariants)
	if count <= 1 {
		return 0
	}
	index, err := rand.Int(rand.Reader, big.NewInt(int64(count)))
	if err == nil {
		return int(index.Int64())
	}
	return int(time.Now().UnixNano() % int64(count))
}

func ansiReset() string {
	return "\x1b[0m"
}
