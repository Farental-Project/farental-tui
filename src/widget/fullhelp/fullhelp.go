package fullhelp

import (
	ftheme "farental/internal/theme"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Style struct {
	Title lipgloss.Style
	Help  lipgloss.Style
}

type Widget struct {
	orvyn.BaseWidget

	Style Style

	title string
}

func New() *Widget {
	w := new(Widget)
	t := orvyn.GetTheme()

	w.title = lokyn.L("Help")

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Style = Style{
		Title: t.Style(ftheme.DimUnderlinedTextStyleID),
		Help: t.Style(theme.NormalTextStyleID).
			Align(lipgloss.Center, lipgloss.Center),
	}

	return w
}

func (w *Widget) Render() string {
	var b strings.Builder

	contentSize := w.GetContentSize()

	w.Style.Title = w.Style.Title.Width(contentSize.Width)

	title := w.Style.Title.Render(w.title)

	b.WriteString(title)
	b.WriteString("\n")

	// The help block only gets what is left under the title. It used to be given
	// the whole content height with the title then stacked on top, so the widget
	// rendered one title taller than the height it had been allocated - and the
	// title is two rows, being underlined.
	//
	// MaxHeight is what actually makes the block shrink: Height is a minimum in
	// lipgloss, so without it the keybindings rendered at their natural height
	// whatever the layout allocated, and the panel overflowed the terminal.
	helpHeight := max(contentSize.Height-lipgloss.Height(title), 0)

	b.WriteString(w.Style.Help.
		Width(contentSize.Width).
		Height(helpHeight).
		MaxHeight(helpHeight).
		Render(w.helpView(w.GetSize().Width)))

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(b.String())
}

// helpView renders the keybinding block for the current context. There is no
// current context until a screen switches to its own, and ViewAll dereferences the
// keymap it is handed, so guard against it being unset.
func (w *Widget) helpView(width int) string {
	keymap := bubblehelp.GetCurrentContextKeymap()

	if keymap == nil {
		return ""
	}

	return bubblehelp.ViewAll(keymap, width)
}

// chromeHeight is everything Render draws around the help block: the title, which
// is two rows because it is underlined, plus the widget frame. The newline written
// after the title terminates its last row rather than adding one.
func (w *Widget) chromeHeight() int {
	return lipgloss.Height(w.Style.Title.Render(w.title)) +
		w.GetStyle().GetVerticalFrameSize()
}

// GetMinSize keeps room for the chrome plus a single help row. It used to claim a
// height of 1, which is less than the title alone.
func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(1, w.chromeHeight()+1)
}

// GetPreferredSize measures the help block at the width the widget was actually
// given, then adds the chrome. It used to measure at a hardcoded width of 50 and
// count neither the title nor the frame, so it under-reported its height by the
// whole chrome and mis-measured how the keybindings wrap.
//
// Width stays at 1 so the help never drives the layout width, for the same reason
// documented on the help widget.
func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(1, w.chromeHeight()+lipgloss.Height(w.helpView(w.GetSize().Width)))
}

func (w *Widget) SetTitle(title string) {
	w.title = title
}
