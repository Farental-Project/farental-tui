package updater

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

// minNotesWidth keeps rendering sane on a very narrow terminal.
const minNotesWidth = 20

// listIndentWidth is the number of spaces one nesting level adds.
const listIndentWidth = 2

// RenderNotes turns release note blocks into styled terminal lines wrapped to
// width. The server decides what the notes say; this decides how they look,
// because only the client knows the width and the active theme.
func RenderNotes(blocks []Block, width int) []string {
	if len(blocks) == 0 {
		return nil
	}

	if width < minNotesWidth {
		width = minNotesWidth
	}

	t := orvyn.GetTheme()
	titleStyle := t.Style(theme.TitleStyleID)
	highlightStyle := t.Style(theme.HighlightTextStyleID)
	dimStyle := t.Style(theme.DimTextStyleID)

	var lines []string

	for i, b := range blocks {
		if i > 0 {
			lines = append(lines, "")
		}

		switch b.Type {
		case "h2":
			text := spansText(b.Spans)
			lines = append(lines, titleStyle.Render(text))
			lines = append(lines, titleStyle.Render(strings.Repeat("─", lipgloss.Width(text))))

		case "h3":
			lines = append(lines, highlightStyle.Render(spansText(b.Spans)))

		case "quote":
			for _, l := range wrapStyled(renderSpans(b.Spans), width-2) {
				lines = append(lines, dimStyle.Render("│ ")+l)
			}

		case "code":
			for _, l := range b.Lines {
				lines = append(lines, dimStyle.Render("  "+l))
			}

		case "list":
			lines = append(lines, renderList(b, width)...)

		default:
			lines = append(lines, wrapStyled(renderSpans(b.Spans), width)...)
		}
	}

	return lines
}

func renderList(b Block, width int) []string {
	var lines []string

	// counters holds one running count per indent level, so ordered
	// numbering resets when nesting resumes after a shallower item.
	var counters []int

	for _, item := range b.Items {
		indent := strings.Repeat(" ", item.Indent*listIndentWidth)

		marker := "• "

		if b.Ordered {
			for len(counters) <= item.Indent {
				counters = append(counters, 0)
			}

			counters[item.Indent]++
			counters = counters[:item.Indent+1]

			marker = fmt.Sprintf("%d. ", counters[item.Indent])
		}

		first := indent + marker
		rest := indent + strings.Repeat(" ", lipgloss.Width(marker))

		wrapped := wrapStyled(renderSpans(item.Spans), width-lipgloss.Width(first))

		for i, l := range wrapped {
			if i == 0 {
				lines = append(lines, first+l)
				continue
			}

			lines = append(lines, rest+l)
		}
	}

	return lines
}

// renderSpans applies inline marks. A link becomes "text (url)" because
// terminal hyperlink escapes are not portable enough to rely on here.
func renderSpans(spans []Span) string {
	var b strings.Builder

	for _, s := range spans {
		text := s.Text

		if s.Href != "" {
			text = fmt.Sprintf("%s (%s)", text, s.Href)
		}

		style := lipgloss.NewStyle().
			Bold(s.Bold).
			Italic(s.Italic).
			Underline(s.Underline).
			Strikethrough(s.Strike)

		b.WriteString(style.Render(text))
	}

	return b.String()
}

func spansText(spans []Span) string {
	var b strings.Builder

	for _, s := range spans {
		b.WriteString(s.Text)
	}

	return b.String()
}

// wrapStyled wraps text that already contains style escapes. ansi.Wordwrap
// measures visible width; lipgloss width math would count the escapes.
func wrapStyled(s string, width int) []string {
	if width < 1 {
		width = 1
	}

	return strings.Split(ansi.Wordwrap(s, width, ""), "\n")
}
