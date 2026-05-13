package simplelogviewer

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
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
	w.titleHeight = lipgloss.Height(w.titleStyle.Render(w.title))
}

func (w *Widget) OnBlur() {
	w.widgetStyle = w.Style.BlurredWidget
	w.titleStyle = w.Style.BlurredTitle
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

	b.WriteString(w.viewport.View())

	return w.widgetStyle.Render(b.String())
}

func (w *Widget) Resize(size orvyn.Size) {
	var marginW, marginH int

	marginW += w.widgetStyle.GetBorderLeftSize()
	marginW += w.widgetStyle.GetBorderRightSize()

	marginH += w.widgetStyle.GetBorderTopSize()

	size.Width -= marginW
	size.Height -= w.titleHeight + marginH

	size.Width = max(size.Width, 0)
	size.Height = max(size.Height, 0)

	w.titleStyle = w.titleStyle.Width(size.Width)
	w.viewport.Width = size.Width
	w.viewport.Height = size.Height

	if size != w.GetSize() {
		w.refresh()
	}

	w.BaseWidget.Resize(size)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(10, w.titleHeight+1)
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
