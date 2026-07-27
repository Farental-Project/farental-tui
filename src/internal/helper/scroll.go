package helper

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	ArrowUp   = "↑"
	ArrowDown = "↓"
)

// OverlayScrollArrows overlays scroll indicators on the last column of view:
// an up arrow on the first line when hasAbove is true, and a down arrow on the
// last line when hasBelow is true. Arrows are rendered with arrowStyle.
func OverlayScrollArrows(view string, width int, arrowStyle lipgloss.Style,
	hasAbove, hasBelow bool) string {
	if width < 1 || (!hasAbove && !hasBelow) {
		return view
	}

	lines := strings.Split(view, "\n")

	if hasAbove {
		lines[0] = overlayArrow(lines[0], width, arrowStyle.Render(ArrowUp))
	}

	if hasBelow {
		last := len(lines) - 1
		lines[last] = overlayArrow(lines[last], width, arrowStyle.Render(ArrowDown))
	}

	return strings.Join(lines, "\n")
}

// overlayArrow places arrow (already styled) on the last column (width-1) of
// line, preserving the preceding content and padding with spaces when the line
// is too short. ANSI styling on arrow does not count toward its display width.
func overlayArrow(line string, width int, arrow string) string {
	left := ansi.Truncate(line, width-1, "")

	pad := (width - 1) - ansi.StringWidth(left)
	if pad > 0 {
		left += strings.Repeat(" ", pad)
	}

	return left + arrow
}
