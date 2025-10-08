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
	s.list.CursorMovingCallback = s.ruleListCursorMoving
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
	s.list.Init()
	s.ruleTypeInspector.Init()

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
			Rules:       scriptDetail.Rules,
		}

		s.scriptInfo.SetData(&s.data)

		if len(s.data.Rules) > 0 {
			s.list.SetData(&s.data.Rules)
			s.ruleListCursorMoved(0)
		}
	}

	s.focusManager.FocusFirst()

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

		case key.Matches(m, keybind.SKeyCtrl):
			if !s.focusManager.IsInputting() {
				s.save()
			}
		}
	}

	switch msg := msg.(type) {
	case scriptrulelist.ChangedRuleTypeMsg:
		s.inspectorUpdate(string(msg), nil)
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

// ruleDataUpdate updates the rule with the data contained in the rule type inspector.
func (s *Screen) ruleDataUpdate(ruleItemData *scriptrulelist.Data) {
	params := s.ruleTypeInspector.GetItemsData()

	ruleItemData.Parameters = params
}

// inspectorUpdate updates the inspector based on the rule type in the selected rule.
func (s *Screen) inspectorUpdate(code string, data *[]api.ScriptRuleTypeParam) {
	err := s.ruleTypeInspector.SetRuleType(code, data)

	if err != nil {
		s.statusMessage.SetError(err)
	}
}

// ruleListCursorMoving is called before the movement of the cursor in the ruleList.
func (s *Screen) ruleListCursorMoving(index int) {
	items := s.list.GetItems()

	if len(items) > 0 {
		item := &items[index]
		s.ruleDataUpdate(item)
		s.list.SetItem(index, *item)
	}
}

// ruleListCursorMoved is called after the movement of the cursor in the ruleList.
func (s *Screen) ruleListCursorMoved(index int) {
	items := s.list.GetItems()

	if len(items) > 0 {
		s.inspectorUpdate(items[index].RuleTypeCode, &items[index].Parameters)
	}
}

func (s *Screen) save() {
	infoData := s.scriptInfo.GetData()

	s.data.Name = infoData.Name
	s.data.Description = infoData.Description
	s.data.IsPrivate = infoData.Private

	s.ruleListCursorMoving(s.list.GetGlobalIndex())

	rulesData := s.list.GetItems()

	s.data.Rules = make([]api.ScriptRuleBody, 0)

	for _, rd := range rulesData {
		s.data.Rules = append(s.data.Rules, rd.ScriptRuleBody)
	}

	resp, err := helper.SendRequest(request.ScriptSave(&s.data))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == 200 {
		s.statusMessage.SetMessage(lokyn.L("Script save successfully."), statusmessage.SuccessMessage)
	}
}
