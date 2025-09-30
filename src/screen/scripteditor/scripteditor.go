package scripteditor

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/widget/help"
	"farental/widget/ruletypeinspector"
	"farental/widget/scriptinfoinput"
	"farental/widget/scriptrulelist"
	"farental/widget/scriptrulelistitem"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type Screen struct {
	title             *orvyn.SimpleRenderable
	scriptInfo        *scriptinfoinput.Widget
	list              *scriptrulelist.Widget
	ruleTypeInspector *ruletypeinspector.Widget
	statusMessage     *statusmessage.Widget
	help              *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	data api.ScriptBody

	new bool

	returnErr error
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Script editor"))
	s.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	s.scriptInfo = scriptinfoinput.New()

	s.list = scriptrulelist.New()
	s.list.SetFilterable(false)
	s.list.CursorMovedCallback = s.ruleListCursorMoved

	s.ruleTypeInspector = ruletypeinspector.New()

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.scriptInfo)
	s.focusManager.Add(s.list)
	s.focusManager.Add(s.ruleTypeInspector)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				layout.NewHBoxFixedRatioLayout(
					0, 1, 1,
					[]layout.FixedRatioRenderable{
						layout.NewFixedRatioRenderable(0.2, s.scriptInfo),
						layout.NewFixedRatioRenderable(0.6, s.list),
						layout.NewFixedRatioRenderable(0.2, s.ruleTypeInspector),
					},
				),
				s.statusMessage,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	script, ok := i.(*api.ScriptBasicResponse)

	if !ok || script == nil {
		s.new = true
	} else {
		s.new = false

		resp, err := helper.SendRequest(request.ScriptGetDetail(script.ID))

		if err != nil {
			s.returnErr = fmt.Errorf(lokyn.L("Cannot open selected script"))
			return orvyn.SwitchToPreviousScreen()
		}

		scriptDetail, ok := resp.Result().(*api.ScriptResponse)

		if !ok {
			s.returnErr = fmt.Errorf(lokyn.L("Cannot open selected script"))
			return orvyn.SwitchToPreviousScreen()
		}

		s.data = api.ScriptBody{
			ID:          scriptDetail.ID,
			Name:        scriptDetail.Name,
			Description: scriptDetail.Description,
			IsPrivate:   scriptDetail.IsPrivate,
			Rules:       make([]api.ScriptRuleBody, 0),
		}

		for _, r := range scriptDetail.Rules {
			s.data.Rules = append(s.data.Rules, r.ScriptRuleBody)
		}

		s.scriptInfo.SetData(&s.data)
	}

	s.focusManager.FocusFirst()

	s.list.Init()
	s.ruleTypeInspector.Init()

	return nil
}

func (s *Screen) OnExit() any {
	return s.returnErr
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		s.statusMessage.Reset()

		switch {
		case key.Matches(m, keybind.Esc):
			if !s.focusManager.IsInputting() && !s.list.IsInputting() {
				return orvyn.SwitchToPreviousScreen()
			}
		}
	}

	switch msg := msg.(type) {
	case scriptrulelistitem.ChangedRuleTypeMsg:
		s.inspectorUpdate(string(msg))
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) inspectorUpdate(code string) {
	err := s.ruleTypeInspector.SetRuleType(code)

	if err != nil {
		s.statusMessage.SetError(err)
	}
}

func (s *Screen) ruleListCursorMoved(index int) {
	// TODO: Finish the reset of the inspector
	items := s.list.GetItems()

	if len(items) > 0 {
		s.inspectorUpdate(items[index].RuleTypeCode)
	}
}
