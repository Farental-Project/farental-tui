package scripteditor

import (
	"bytes"
	"encoding/json"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/ticker"
	"farental/screen"
	"farental/screen/dialog/popup"
	"farental/widget/help"
	"farental/widget/ruletypeinspector"
	"farental/widget/runningtask"
	"farental/widget/scriptinfoinput"
	"farental/widget/scriptrulelist"
	"fmt"
	"log"

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
	title             *orvyn.SimpleRenderable
	readOnlyTitle     *orvyn.SimpleRenderable
	scriptInfo        *scriptinfoinput.Widget
	list              *scriptrulelist.Widget
	ruleTypeInspector *ruletypeinspector.Widget
	statusMessage     *statusmessage.Widget
	help              *help.Widget

	runningTask *runningtask.Widget

	ticker *ticker.Ticker

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	data       api.ScriptBody
	originData api.ScriptBody

	new bool

	returnErr error
}

func New() *Screen {
	s := new(Screen)

	ts := orvyn.GetTheme().Style(theme.TitleStyleID)

	s.title = orvyn.NewSimpleRenderable("Script editor")
	s.title.Style = ts

	s.readOnlyTitle = orvyn.NewSimpleRenderable("Read only")
	s.readOnlyTitle.Style = ts
	s.readOnlyTitle.SetActive(false)

	s.scriptInfo = scriptinfoinput.New()

	s.list = scriptrulelist.New()
	s.list.SetFilterable(false)
	s.list.CursorMovingCallback = s.ruleListCursorMoving
	s.list.CursorMovedCallback = s.ruleListCursorMoved

	s.ruleTypeInspector = ruletypeinspector.New()

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.runningTask = runningtask.New()
	s.runningTask.SetActive(false)

	s.focusManager = orvyn.NewFocusManager()

	inspectorElements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.2, s.scriptInfo),
		layout.NewFixedRatioRenderable(0.6, s.list),
		layout.NewFixedRatioRenderable(0.2, s.ruleTypeInspector),
	}

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			4,
			s.title,
			s.readOnlyTitle,
			orvyn.VGap,
			s.runningTask,
			layout.NewHBoxFixedRatioLayout(0, 1, 1, inspectorElements...),
			s.statusMessage,
			s.help,
		),
	)

	s.ticker = ticker.New(60, func() {
		if err := context.RefreshRunningTask(); err != nil {
			log.Println(err)
		}

		s.runningTask.SetActive(context.RunningTask != nil)
	})

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	widgets := []orvyn.Focusable{
		s.scriptInfo,
		s.list,
		s.ruleTypeInspector,
	}
	s.focusManager.SetWidgets(widgets)

	s.title.SetValue(lokyn.L("Script editor"))
	s.readOnlyTitle.SetValue(lokyn.L("Read only"))

	s.readOnlyTitle.SetActive(false)

	s.returnErr = nil

	s.scriptInfo.Init()
	s.list.Init()
	s.ruleTypeInspector.Init()

	script, ok := i.(*api.ScriptBasicResponse)

	if !ok || script == nil {
		s.new = true

		s.data = api.ScriptBody{}
		s.originData = api.ScriptBody{}
	} else {
		s.new = false

		if !script.IsEditable && !script.IsDuplicated {
			s.setReadOnly()
		}

		scriptDetail, err := helper.Fetch[api.ScriptResponse](request.ScriptGetDetail(script.ID))

		if err != nil {
			s.returnErr = fmt.Errorf("%s", lokyn.L("Cannot open selected script"))
			return orvyn.SwitchToPreviousScreen()
		}

		s.data = api.ScriptBody{
			ID:          scriptDetail.ID,
			Name:        scriptDetail.Name,
			Description: scriptDetail.Description,
			IsPrivate:   scriptDetail.IsPrivate,
			Rules:       scriptDetail.Rules,
		}

		if script.IsDuplicated {
			s.data.Name += fmt.Sprintf(" %s", lokyn.L("(Duplicated)"))
		}

		s.originData = s.data

		s.scriptInfo.SetData(&s.data)

		if len(s.data.Rules) > 0 {
			s.list.SetData(&s.data.Rules)
			s.ruleListCursorMoved(0)
		}
	}

	s.focusManager.FocusFirst()

	if err := context.RefreshRunningTask(); err != nil {
		log.Println(err)
	}

	s.runningTask.SetActive(context.RunningTask != nil)

	return tea.Batch(s.runningTask.Init(), s.ticker.Start())
}

func (s *Screen) OnExit() any {
	s.focusManager.BlurCurrent()
	return s.returnErr
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		s.statusMessage.Reset()

		switch {
		case key.Matches(m, keybind.Help):
			if bubblehelp.IsKeybindVisible(keybind.Help) {
				bubblehelp.ShowAll = !bubblehelp.ShowAll
			}

		case key.Matches(m, keybind.Esc):
			if !s.focusManager.IsInputting() && !s.list.IsInputting() {

				if s.dataModified() {
					orvyn.OpenDialog("quitConfirm", popup.NewYesNo(
						lokyn.L("Are you sure you want to quit the editor and lose your current progress ?"),
					), nil)
				} else {
					return orvyn.SwitchScreen(screen.IDScriptExplorer)
				}

				return nil
			}

		case key.Matches(m, keybind.SKeyCtrl):
			if bubblehelp.IsKeybindVisible(keybind.SKeyCtrl) {
				if !s.focusManager.IsInputting() {
					s.save()
				}
			}
		}
	}

	var dialogCmd tea.Cmd

	switch msg := msg.(type) {
	case scriptrulelist.ChangedRuleTypeMsg:
		s.inspectorUpdate(string(msg), nil)

	case scriptrulelist.FocusInspectorMsg:
		if !s.ruleTypeInspector.IsEmpty() {
			s.focusManager.Focus(2)
			s.focusManager.ForceInput(2)
			return nil
		}

	case ruletypeinspector.FocusRuleListMsg:
		s.focusManager.Focus(1)

	case orvyn.DialogExitMsg:
		if msg.DialogID == "quitConfirm" && msg.Param.(uint) == 1 {
			return orvyn.SwitchScreen(screen.IDScriptExplorer)
		}

		// Re-arm unconditionally: orvyn routes every message to whichever dialog is
		// open, so the ticker and the running-task spinner die while ANY dialog is
		// open here - not just "quitConfirm" (the rule list also opens "selectRuleType"
		// and "selectAbility" via widget/scriptrulelist/scriptrulelistitem.go).
		// Must NOT return here: this message still needs to reach focusManager.Update
		// below so the rule list item that owns the dialog can consume its result.
		dialogCmd = tea.Batch(s.runningTask.Init(), s.ticker.Restart())

	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return cmd

	}

	cmd := s.focusManager.Update(msg)

	return tea.Batch(cmd, dialogCmd, s.runningTask.Update(msg))
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

func (s *Screen) updateData() {
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
}

func (s *Screen) save() {
	s.updateData()

	resp, err := helper.SendRequest(request.ScriptSave(&s.data))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == 200 {
		s.statusMessage.SetMessage(lokyn.L("Script saved successfully."), statusmessage.SuccessMessage)
		uuidResp, ok := resp.Result().(*api.UUIDResponse)

		if ok {
			s.data.ID = uuidResp.ID
		}

		s.originData = s.data
	}
}

func (s *Screen) dataModified() bool {
	// Compare originData and data to know if something changed.
	s.updateData()

	originJson, err := json.Marshal(s.originData)

	if err != nil {
		log.Println(err)
		return false
	}

	dataJson, err := json.Marshal(s.data)

	if err != nil {
		log.Println(err)
		return false
	}

	if !bytes.Equal(dataJson, originJson) {
		return true
	}

	return false
}

func (s *Screen) setReadOnly() {
	s.focusManager.SetWidgets([]orvyn.Focusable{s.list})
	s.readOnlyTitle.SetActive(true)
	s.list.SetReadOnly()
}
