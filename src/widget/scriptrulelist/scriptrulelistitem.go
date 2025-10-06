package scriptrulelist

import (
	cdata "farental/core/data"
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/internal/style"
	"farental/screen/dialog/abilityselection"
	"farental/screen/dialog/ruletypeselection"
	"farental/widget/button"
	"farental/widget/multivalueselector"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type ChangedRuleTypeMsg string

func ChangedRuleTypeCmd(code string) tea.Cmd {
	return func() tea.Msg {
		return ChangedRuleTypeMsg(code)
	}
}

type Data struct {
	api.ScriptRuleBody
	AbilityName  string
	RuleTypeName string
}

type ListItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data *Data

	titleOrder    *orvyn.SimpleRenderable
	titleRuleType *orvyn.SimpleRenderable
	titleTarget   *orvyn.SimpleRenderable
	titleAbility  *orvyn.SimpleRenderable

	btRuleType *button.Widget
	mvsTarget  *multivalueselector.Widget[cdata.Target]
	btAbility  *button.Widget

	btRuleTypePlaceHolder string
	btAbilityPlaceHolder  string

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout

	style lipgloss.Style

	contentSize orvyn.Size
}

func Constructor(data *Data) list.ListItem {
	inputKeymap := bubblehelp.NewKeymap(2)
	inputKeymap.Style = style.MainHelpStyle
	inputKeymap.NewKeyBinding(keybind.Tab, true)
	inputKeymap.NewKeyBinding(keybind.ShiftTab, true)
	inputKeymap.NewKeyBinding(keybind.Space, true)
	inputKeymap.SetHelpDesc(keybind.Space, lokyn.L("Open selection"))
	inputKeymap.NewKeyBinding(keybind.Esc, true)
	inputKeymap.SetHelpDesc(keybind.Esc, lokyn.L("Stop editing"))
	inputKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextScriptEditorRulesListItem, inputKeymap)

	w := new(ListItem)

	w.data = data

	t := orvyn.GetTheme()
	dts := t.Style(theme.DimTextStyleID)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.titleOrder = orvyn.NewSimpleRenderable("")

	w.titleRuleType = orvyn.NewSimpleRenderable(lokyn.L("Rule type"))
	w.titleRuleType.Style = dts
	w.titleRuleType.SizeConstraint = true

	w.titleTarget = orvyn.NewSimpleRenderable(lokyn.L("Target"))
	w.titleTarget.Style = dts
	w.titleTarget.SizeConstraint = true

	w.titleAbility = orvyn.NewSimpleRenderable(lokyn.L("Ability"))
	w.titleAbility.Style = dts
	w.titleAbility.SizeConstraint = true

	w.btRuleTypePlaceHolder = lokyn.L("Select a rule type")
	w.btRuleType = button.New(w.btRuleTypePlaceHolder)
	w.btRuleType.OnFocusCallback = w.btOnFocus
	w.btRuleType.OnBlurCallback = w.btOnBlur
	w.btRuleType.OnClickedCallback = w.btRuleTypeOnClicked

	w.btAbilityPlaceHolder = lokyn.L("Select an ability")
	w.btAbility = button.New(w.btAbilityPlaceHolder)
	w.btAbility.OnFocusCallback = w.btOnFocus
	w.btAbility.OnBlurCallback = w.btOnBlur
	w.btAbility.OnClickedCallback = w.btAbilityOnClicked

	w.mvsTarget = multivalueselector.New[cdata.Target]()
	w.mvsTarget.SetValues(cdata.TargetKeys, cdata.Targets)
	w.mvsTarget.OnBlur()

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.btRuleType)
	w.focusManager.Add(w.mvsTarget)
	w.focusManager.Add(w.btAbility)

	titleLayout := layout.NewHBoxGrowLayout(1, 1,
		[]orvyn.Renderable{
			w.titleRuleType,
			w.titleTarget,
			w.titleAbility,
		})

	controlsLayout := layout.NewHBoxGrowLayout(1, 1,
		[]orvyn.Renderable{
			w.btRuleType,
			w.mvsTarget,
			w.btAbility,
		})

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		[]orvyn.Renderable{
			w.titleOrder,
			titleLayout,
			controlsLayout,
		})

	w.OnBlur()

	w.UpdateData()

	return w
}

func (w *ListItem) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "selectRuleType":
			val, ok := msg.Param.(api.ScriptRuleTypeResponse)

			if ok {
				w.data.RuleTypeName = val.Name
				w.data.RuleTypeCode = val.Code
				w.btRuleType.SetLabel(val.Name)

				cmds = append(cmds, ChangedRuleTypeCmd(val.Code))
			}

		case "selectAbility":
			val, ok := msg.Param.(api.AbilityResponse)

			if ok {
				w.data.AbilityName = val.Name
				w.data.AbilityCode = val.Code
				w.btAbility.SetLabel(val.Name)
			}
		}

		bubblehelp.SwitchToPreviousContext()
	}

	cmd := w.focusManager.Update(msg)

	cmds = append(cmds, cmd)

	w.updateData()

	return tea.Batch(cmds...)
}

func (w *ListItem) UpdateData() {
	w.titleOrder.SetValue(fmt.Sprintf(lokyn.L("Order : %d"), w.data.Order))
	w.mvsTarget.SetSelected(int(w.data.Target))

	abilityName := w.btAbilityPlaceHolder
	ruleTypeName := w.btRuleTypePlaceHolder

	if w.data.AbilityName != "" {
		abilityName = w.data.AbilityName
	}

	if w.data.RuleTypeName != "" {
		ruleTypeName = w.data.RuleTypeName
	}

	w.btAbility.SetLabel(abilityName)
	w.btRuleType.SetLabel(ruleTypeName)
}

// updateData updates the data based on the widgets values
func (w *ListItem) updateData() {
	scriptTarget := w.mvsTarget.GetSelectedValue().ScriptTarget

	if w.data.Target != scriptTarget {
		w.data.Target = scriptTarget
		w.data.AbilityCode = ""
		w.data.AbilityName = ""
	}
}

func (w *ListItem) Resize(size orvyn.Size) {
	size.Height = 7

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
	w.layout.Resize(size)
}

func (w *ListItem) Render() string {
	return w.style.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *ListItem) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *ListItem) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *ListItem) OnEnterInput() {
	bubblehelp.SwitchContext(keybind.ContextScriptEditorRulesListItem)
	w.focusManager.FocusFirst()
}

func (w *ListItem) OnExitInput() {
	w.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(keybind.ContextScriptEditorRulesList)
}

func (w *ListItem) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}

func (w *ListItem) FilterValue() string {
	return ""
}

func (w *ListItem) btOnFocus() {
	// TODO: Need to check current keymap
	bubblehelp.SetKeybindVisible(keybind.Space, true)
}

func (w *ListItem) btOnBlur() {
	// TODO: Need to check current keymap
	bubblehelp.SetKeybindVisible(keybind.Space, false)
}

func (w *ListItem) btRuleTypeOnClicked() tea.Cmd {
	orvyn.OpenDialog("selectRuleType", ruletypeselection.New(), nil)

	return nil
}

func (w *ListItem) btAbilityOnClicked() tea.Cmd {
	orvyn.OpenDialog("selectAbility", abilityselection.New(), w.data.Target)

	return nil
}
