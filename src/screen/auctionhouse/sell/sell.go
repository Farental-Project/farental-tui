package sell

import (
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/auctionstartform"
	"farental/widget/help"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	auctionStartForm *auctionstartform.Widget

	statusMessage *statusmessage.Widget
	help          *help.Widget

	layout *layout.CenterLayout
}

var _ orvyn.Screen = (*Screen)(nil)

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("sell")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.auctionStartForm = auctionstartform.New()

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(10, t.Size(ftheme.LayoutWidthSizeID), orvyn.NewSize(10, 4),
			s.title,
			orvyn.VGap,
			s.auctionStartForm,
			s.statusMessage,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextBasicEditMode)

	s.title.SetValue(lokyn.L("Create an auction"))

	s.auctionStartForm.Init()
	s.auctionStartForm.OnFocus()

	s.statusMessage.Reset()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if k, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()
		}
	}

	cmd := s.auctionStartForm.Update(msg)

	return cmd
}
