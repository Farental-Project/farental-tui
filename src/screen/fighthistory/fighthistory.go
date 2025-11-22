package fighthistory

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/screen"
	"farental/widget/fighthistorylistitem"
	"farental/widget/help"
	"farental/widget/simplelogviewer"
	"fmt"

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

	list *list.Widget[fighthistorylistitem.Data]
	log  *simplelogviewer.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	loadedLogsCache map[uint]api.EventLogResponse
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("Fight history")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.list = list.New(fighthistorylistitem.Constructor)

	logStyle := simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}

	s.log = simplelogviewer.New(lokyn.L("Fight log"))
	s.log.Style = logStyle
	s.log.SetAutoScroll(false)
	s.log.OnBlur()

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.list)
	s.focusManager.Add(s.log)

	mainElements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.3, s.list),
		layout.NewFixedRatioRenderable(0.7, s.log),
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
	bubblehelp.SwitchContext(keybind.ContextFightHistory)

	s.loadedLogsCache = make(map[uint]api.EventLogResponse)

	s.focusManager.FocusFirst()

	s.loadFightHistory()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(msg, keybind.Enter):
			s.loadLog()

			return nil

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchScreen(screen.IDFight)

		case key.Matches(msg, keybind.Help):
			bubblehelp.ShowAll = !bubblehelp.ShowAll

			return nil

		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadFightHistory() {
	resp, err := helper.SendRequest(request.FightGetFinished())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	fights := resp.Result().(*[]api.FightResponse)

	data := make([]fighthistorylistitem.Data, 0)

	for _, f := range *fights {
		data = append(data, fighthistorylistitem.Data{
			FightResponse: f,
		})
	}

	s.list.SetItems(data)
}

func (s *Screen) loadLog() {
	fight := s.list.GetSelectedItem()

	log, ok := s.loadedLogsCache[fight.ID]

	if !ok {

		resp, err := helper.SendRequest(request.FightGetLog(fight.ID))

		if err != nil {
			s.statusMessage.SetError(err)
			return
		}

		log = *resp.Result().(*api.EventLogResponse)
	}

	logData := make([]string, 0)

	for _, e := range log.Entries {
		logData = append(logData, fmt.Sprintf("%d : %s",
			e.Order, e.Value))
	}

	s.log.SetContent(logData)
}
