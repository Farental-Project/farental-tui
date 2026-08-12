// Package manual implements the dialog displaying an in-app manual.
package manual

import (
	"farental/internal/keybind"
	manualdoc "farental/internal/manual"
	ftheme "farental/internal/theme"
	"farental/widget/help"
	"farental/widget/simplelogviewer"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

// DialogID is the orvyn dialog ID of the manual, distinct from the IDs the
// screens use for their own dialogs.
const DialogID orvyn.ScreenID = "manual"

// Topics of the available manuals, matching the names of the embedded files.
const (
	TopicScriptEditor = "scripteditor"
)

// pageScrollLines is the number of lines scrolled by the page up and page down
// keys.
const pageScrollLines = 10

// Open returns the command opening the manual dialog on the given topic.
func Open(topic string) tea.Cmd {
	return orvyn.OpenDialog(DialogID, New(), topic)
}

type Screen struct {
	title     *orvyn.SimpleRenderable
	logViewer *simplelogviewer.Widget
	help      *help.Widget

	layout *layout.CenterLayout

	// previousContext holds the bubblehelp context of the screen that opened
	// the dialog, restored on exit. It stays empty when the manual could not
	// be loaded, in which case no context switch happened either.
	previousContext bubblehelp.KeymapContext
}

// New returns the manual dialog. It is meant to be used with
// orvyn.OpenDialog(), with the topic of the manual to display as parameter.
func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.logViewer = simplelogviewer.New("")
	s.logViewer.Style = simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}

	// The manual is read from the beginning, unlike the event log the viewer
	// is usually used for.
	s.logViewer.SetAutoScroll(false)
	s.logViewer.OnFocus()

	s.help = help.New()

	// Grow index 1 is the log viewer: every other element is resized to its
	// min height, so the manual body is what takes the leftover space.
	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 1,
			s.title,
			s.logViewer,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	topic, ok := i.(string)

	if !ok {
		return orvyn.CloseDialog()
	}

	title, content, err := manualdoc.Get(topic)

	if err != nil {
		return orvyn.CloseDialog()
	}

	s.title.SetValue(title)
	s.logViewer.SetContent(content)

	s.previousContext = bubblehelp.CurrentContext
	bubblehelp.SwitchContext(keybind.ContextManual)

	return nil
}

func (s *Screen) OnExit() any {
	if s.previousContext != "" {
		bubblehelp.SwitchContext(s.previousContext)
		s.previousContext = ""
	}

	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Quit):
			return tea.Quit

		case key.Matches(m, keybind.Esc):
			return orvyn.CloseDialog()

		case key.Matches(m, keybind.PrevPage):
			s.logViewer.ScrollUp(pageScrollLines)
			return nil

		case key.Matches(m, keybind.NextPage):
			s.logViewer.ScrollDown(pageScrollLines)
			return nil

		case key.Matches(m, keybind.Home):
			s.logViewer.GotoTop()
			return nil

		case key.Matches(m, keybind.End):
			s.logViewer.GotoBottom()
			return nil
		}
	}

	return s.logViewer.Update(msg)
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
