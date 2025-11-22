package npc

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/screen"
	"farental/widget/help"
	npclistitem "farental/widget/npclistiem"
	"farental/widget/simplelogviewer"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	list   *list.Widget[api.NpcResponse]
	dialog *simplelogviewer.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("NPCs")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.list = list.New(npclistitem.Constructor)

	logStyle := simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}

	s.dialog = simplelogviewer.New(lokyn.L("Dialog"))
	s.dialog.Style = logStyle
	s.dialog.OnBlur()

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.list)
	s.focusManager.Add(s.dialog)

	mainElements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.3, s.list),
		layout.NewFixedRatioRenderable(0.7, s.dialog),
	}

	mainLayout := layout.NewHBoxFixedRatioLayout(0, 1, 0, mainElements...)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 2,
			s.title,
			orvyn.VGap,
			mainLayout,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextNpc)

	s.focusManager.FocusFirst()

	s.loadNpc()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(msg, keybind.Enter):
			s.speakToNpc()

			return nil

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchScreen(screen.IDDashBoard)

		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadNpc() {
	resp, err := helper.SendRequest(request.NpcGetAvailable())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	npcs := resp.Result().(*[]api.NpcResponse)

	s.list.SetItems(*npcs)
}

func (s *Screen) speakToNpc() {
	npc := s.list.GetSelectedItem()

	resp, err := helper.SendRequest(request.NpcTalkTo(npc.ID))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	dialog := resp.Result().(*api.NpcDialogResponse)

	s.dialog.SetContent(strings.Split(dialog.Dialog, "\n"))
}
