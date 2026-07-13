package simplelogviewer

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

const (
	arrowUp   = "↑"
	arrowDown = "↓"
)

type Style struct {
	FocusedWidget lipgloss.Style
	BlurredWidget lipgloss.Style
	FocusedTitle  lipgloss.Style
	BlurredTitle  lipgloss.Style
	line          lipgloss.Style
}

type Keybind struct {
	ScrollUp   key.Binding
	ScrollDown key.Binding
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	Style   Style
	Keybind Keybind

	title    string
	content  []string
	viewport viewport.Model

	widgetStyle lipgloss.Style
	titleStyle  lipgloss.Style
	arrowStyle  lipgloss.Style

	titleHeight int

	focusKeybind key.Binding

	autoScroll bool
}

func New(title string) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.title = title
	w.content = make([]string, 0)

	w.viewport = viewport.New(0, 0)

	w.Style = Style{
		FocusedWidget: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()),
		BlurredWidget: lipgloss.NewStyle().
			BorderStyle(lipgloss.HiddenBorder()),
		FocusedTitle: lipgloss.NewStyle().
			Bold(true),
		BlurredTitle: lipgloss.NewStyle().
			Italic(true),
		line: lipgloss.NewStyle(),
	}

	w.SetStyle(lipgloss.NewStyle())

	w.Keybind = Keybind{
		ScrollUp: key.NewBinding(
			key.WithKeys("up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down"),
		),
	}

	w.autoScroll = true

	return w
}

func (w *Widget) OnFocus() {
	w.widgetStyle = w.Style.FocusedWidget
	w.titleStyle = w.Style.FocusedTitle
	w.arrowStyle = orvyn.GetTheme().Style(theme.NormalTextStyleID)
	w.titleHeight = lipgloss.Height(w.titleStyle.Render(w.title))
}

func (w *Widget) OnBlur() {
	w.widgetStyle = w.Style.BlurredWidget
	w.titleStyle = w.Style.BlurredTitle
	w.arrowStyle = orvyn.GetTheme().Style(theme.DimTextStyleID)
	w.titleHeight = lipgloss.Height(w.titleStyle.Render(w.title))
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, w.Keybind.ScrollUp):
			w.viewport.ScrollUp(1)

		case key.Matches(msg, w.Keybind.ScrollDown):
			w.viewport.ScrollDown(1)

		}
	}

	return nil
}

func (w *Widget) Render() string {
	var b strings.Builder

	if w.title != "" {
		b.WriteString(w.titleStyle.Render(w.title))
		b.WriteString("\n")
	}

	b.WriteString(w.renderViewport())

	return w.widgetStyle.Render(b.String())
}

// renderViewport renders the viewport and overlays scroll indicators on the
// last column: an up arrow on the first line when there is content above, and
// a down arrow on the last line when there is content below.
func (w *Widget) renderViewport() string {
	view := w.viewport.View()

	if w.viewport.Width < 1 || w.viewport.Height < 1 {
		return view
	}

	lines := strings.Split(view, "\n")

	if !w.viewport.AtTop() {
		lines[0] = overlayArrow(lines[0], w.viewport.Width, w.arrowStyle.Render(arrowUp))
	}

	if !w.viewport.AtBottom() {
		last := len(lines) - 1
		lines[last] = overlayArrow(lines[last], w.viewport.Width, w.arrowStyle.Render(arrowDown))
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

func (w *Widget) Resize(size orvyn.Size) {
	var marginW, marginH int

	marginW += w.widgetStyle.GetBorderLeftSize()
	marginW += w.widgetStyle.GetBorderRightSize()

	marginH += w.widgetStyle.GetBorderTopSize()
	marginH += w.widgetStyle.GetBorderBottomSize()

	innerW := max(size.Width-marginW, 0)
	innerH := max(size.Height-marginH, 0)

	w.titleStyle = w.titleStyle.Width(innerW)
	w.viewport.Width = innerW
	w.viewport.Height = max(innerH-w.titleHeight, 0)

	if size != w.GetSize() {
		w.refresh()
	}

	w.BaseWidget.Resize(size)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(25, w.titleHeight+15)
}

func (w *Widget) SetContent(content []string) {
	w.content = content
	w.refresh()
}

func (w *Widget) AppendContent(content string) {
	w.content = append(w.content, content)
	w.refresh()
}

func (w *Widget) AppendRune(character rune) {
	var line int

	line = len(w.content) - 1

	if line == -1 {
		w.content = append(w.content, "")
		line++
	}

	if character == '\n' {
		w.content = append(w.content, "")
		return
	}

	w.content[line] += string(character)
	w.refresh()
}

func (w *Widget) GetContent() []string {
	return w.content
}

func (w *Widget) ScrollUp(n int) {
	w.viewport.ScrollUp(n)
}

func (w *Widget) ScrollDown(n int) {
	w.viewport.ScrollDown(n)
}

func (w *Widget) refresh() {
	w.viewport.SetContent(w.Style.line.
		Width(w.viewport.Width).Render(
		strings.Join(w.content, "\n")))

	if w.autoScroll {
		w.viewport.GotoBottom()
	}
}

func (w *Widget) SetAutoScroll(b bool) {
	w.autoScroll = b
}

func (w *Widget) SetTitle(title string) {
	w.title = title
}
