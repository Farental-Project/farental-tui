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
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

const dialogSpeed time.Duration = 50

func TickCmd(milliseconds time.Duration, tag uint) tea.Cmd {
	return tea.Tick(milliseconds*time.Millisecond, func(t time.Time) tea.Msg {
		return orvyn.TickMsg{
			Time: t,
			Tag:  tag,
		}
	})
}

type Screen struct {
	title *orvyn.SimpleRenderable

	list        *widgetlist.Widget[api.NpcResponse]
	description *simplelogviewer.Widget
	dialog      *simplelogviewer.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	tickTag uint

	currentNPCID    uint
	dialogAnimating bool
}

func New() *Screen {
	s := new(Screen)

	s.currentNPCID = 0
	s.dialogAnimating = false

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("NPCs")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.list = widgetlist.New(npclistitem.Constructor)

	logStyle := simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}

	s.description = simplelogviewer.New("Description")
	s.description.Style = logStyle
	s.description.OnBlur()
	s.description.SetMinSize(orvyn.NewSize(3, 7))

	s.dialog = simplelogviewer.New("Dialog")
	s.dialog.Style = logStyle
	s.dialog.OnBlur()

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.list)
	s.focusManager.Add(s.description)
	s.focusManager.Add(s.dialog)

	dialogLayout := layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
		s.description,
		s.dialog,
	)

	mainElements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.3, s.list),
		layout.NewFixedRatioRenderable(0.7, dialogLayout),
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

	s.title.SetValue(lokyn.L("NPCs"))
	s.description.SetTitle(lokyn.L("Description"))
	s.dialog.SetTitle(lokyn.L("Dialog"))

	s.dialog.SetContent([]string{})
	s.description.SetActive(false)

	s.focusManager.FocusFirst()

	s.loadNpc()

	s.currentNPCID = 0
	s.dialogAnimating = false

	return TickCmd(0, s.tickTag)
}

func (s *Screen) OnExit() any {
	s.dialogAnimating = false
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case orvyn.TickMsg:
		if msg.Tag != s.tickTag {
			return nil
		}

		s.tickTag++
		return TickCmd(dialogSpeed, s.tickTag)
	}

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
	npcs, err := helper.Fetch[[]api.NpcResponse](request.NpcGetAvailable())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.list.SetItems(*npcs)
	s.list.FocusFirst()
}

func (s *Screen) speakToNpc() {
	npc := s.list.GetSelectedItem()

	if s.currentNPCID == npc.ID && !s.dialogAnimating {
		return
	}

	dialog, err := helper.Fetch[api.NpcDialogResponse](request.NpcTalkTo(npc.ID))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.description.SetContent(strings.Split(npc.Description, "\n"))
	s.description.SetActive(true)

	switch {
	case s.currentNPCID == npc.ID && s.dialogAnimating:
		s.dialog.SetContent(strings.Split(dialog.Dialog, "\n"))
		s.dialogAnimating = false
	default:
		s.dialog.SetContent([]string{})
		s.launchAnimation(dialog.Dialog, npc.ID)
	}

	s.currentNPCID = npc.ID
}

func (s *Screen) launchAnimation(dialog string, npcID uint) {
	var runes []rune

	runes = []rune(dialog)

	s.dialogAnimating = true

	go func(screen *Screen, dialog []rune, npcID uint) {
		for _, r := range dialog {
			if !s.dialogAnimating {
				return
			}

			if s.currentNPCID != npcID {
				return
			}

			screen.dialog.AppendRune(r)

			time.Sleep(dialogSpeed * time.Millisecond)
		}
	}(s, runes, npcID)
}
