// Package logo renders the Scoreline wordmark in a stylized way.
package logo

import (
	"strings"

	"github.com/RuchikG/scoreline/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// letterform represents a letterform. It can be stretched horizontally
// via the boolean argument.
type letterform func(bool) string

const diag = `╱`

// Opts are the options for rendering the Scoreline title art.
type Opts struct {
	FieldColorHex    string // diagonal lines color
	GradientStartHex string // left gradient ramp point
	GradientEndHex   string // right gradient ramp point
	Width            int    // width of the rendered logo
}

// DefaultOpts returns default options using the theme colors.
func DefaultOpts() Opts {
	startHex, endHex := design.AdaptiveGradientColors()
	return Opts{
		FieldColorHex:    startHex,
		GradientStartHex: startHex,
		GradientEndHex:   endHex,
		Width:            80,
	}
}

// Render renders the Scoreline logo.
// The compact argument determines whether it renders compact (for sidebar)
// or wider (for main pane).
func Render(version string, compact bool, o Opts) string {
	width := o.Width
	if compact {
		width = min(width, 48)
	}
	title := design.RenderHeader("SCORELINE", width)
	if version == "" {
		return title
	}
	versionStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color(o.GradientEndHex)).
		Render(version)
	return strings.TrimSpace(title + "\n" + lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(versionStyled))
}

// RenderCompact renders a smaller inline version suitable for headers.
func RenderCompact(width int) string {
	return design.RenderHeader("SCORELINE", width)
}

// applyLineGradient applies a gradient to a single line of text.
func applyLineGradient(text string, startHex, endHex string) string {
	startColor, err1 := colorful.Hex(startHex)
	endColor, err2 := colorful.Hex(endHex)
	if err1 != nil || err2 != nil {
		return text
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	var result strings.Builder
	for i, char := range runes {
		if char == ' ' {
			result.WriteRune(' ')
			continue
		}
		ratio := float64(i) / float64(max(len(runes)-1, 1))
		color := startColor.BlendLab(endColor, ratio)
		hexColor := color.Hex()
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor)).Bold(true)
		result.WriteString(charStyle.Render(string(char)))
	}

	return result.String()
}

// renderWord renders letterforms to form a word.
func renderWord(spacing int, stretchIndex int, letterforms ...letterform) string {
	if spacing < 0 {
		spacing = 0
	}

	rendered := make([]string, len(letterforms))
	for i, letter := range letterforms {
		rendered[i] = letter(i == stretchIndex)
	}

	// Add spacing between letters
	if spacing > 0 {
		spaced := make([]string, 0, len(rendered)*2-1)
		for i, r := range rendered {
			spaced = append(spaced, r)
			if i < len(rendered)-1 {
				spaced = append(spaced, strings.Repeat(" ", spacing))
			}
		}
		rendered = spaced
	}

	return strings.TrimSpace(
		lipgloss.JoinHorizontal(lipgloss.Top, rendered...),
	)
}

// truncateAnsi truncates a string with ANSI codes to a given width.
func truncateAnsi(s string, width int) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	// Simple truncation - not perfect but works for most cases
	runes := []rune(s)
	if len(runes) > width {
		return string(runes[:width])
	}
	return s
}

// Letterform definitions using Unicode block characters
// ▄ ▀ █ ▌ ▐

func letterG(stretch bool) string {
	left := "▄\n█\n "
	center := "▀\n \n▀"
	right := "▀\n▀█\n▀"

	return joinLetterform(
		left,
		stretchPart(center, stretch, 3, 5, 8),
		right,
	)
}

func letterO(stretch bool) string {
	left := "▄\n█\n "
	center := "▀\n \n▀"
	right := "▄\n█\n "

	return joinLetterform(
		left,
		stretchPart(center, stretch, 3, 5, 8),
		right,
	)
}

func letterL(stretch bool) string {
	left := "█\n█\n▀"
	bottom := " \n \n▀"

	return joinLetterform(
		left,
		stretchPart(bottom, stretch, 3, 5, 8),
	)
}

func letterA(stretch bool) string {
	left := " ▄\n█▀\n▀ "
	center := "▀\n▀\n "
	right := "▄ \n▀█\n ▀"

	return joinLetterform(
		left,
		stretchPart(center, stretch, 2, 4, 7),
		right,
	)
}

func letterZ(stretch bool) string {
	// Z shape with thick diagonal:
	// ▀▀▀▀█
	//  █▀▀
	// ▀▀▀▀▀
	topWidth := 4
	if stretch {
		topWidth = cachedRandN(4) + 5 // 5-8
	}

	// Build each line with proper alignment
	line1 := strings.Repeat("▀", topWidth) + "█"
	line2 := strings.Repeat(" ", topWidth-3) + "█▀▀"
	line3 := "▀" + strings.Repeat("▀", topWidth-1) + "▀"

	return line1 + "\n" + line2 + "\n" + line3
}

func joinLetterform(parts ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func stretchPart(s string, stretch bool, baseWidth, minStretch, maxStretch int) string {
	n := baseWidth
	if stretch {
		n = cachedRandN(maxStretch-minStretch) + minStretch
	}

	parts := make([]string, n)
	for i := range parts {
		parts[i] = s
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// cachedRandN returns a cached random number for consistent rendering.
// Uses a simple deterministic approach for now.
var randCache = make(map[int]int)
var randSeed = 0

func cachedRandN(n int) int {
	if n <= 0 {
		return 0
	}
	if v, ok := randCache[n]; ok {
		return v
	}
	// Simple deterministic "random" based on seed
	randSeed = (randSeed*1103515245 + 12345) & 0x7fffffff
	v := randSeed % n
	randCache[n] = v
	return v
}
