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
	Ability      api.AbilityResponse
	RuleTypeName string
}

type ListItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data Data

	titleOrder          *orvyn.SimpleRenderable
	titleRuleType       *orvyn.SimpleRenderable
	titleAbilityTarget  *orvyn.SimpleRenderable
	titleRuleTypeTarget *orvyn.SimpleRenderable
	titleAbility        *orvyn.SimpleRenderable

	mvsRuleTypeTarget *multivalueselector.Widget[cdata.Target]
	btRuleType        *button.Widget
	mvsAbilityTarget  *multivalueselector.Widget[cdata.Target]
	btAbility         *button.Widget

	btRuleTypePlaceHolder string
	btAbilityPlaceHolder  string

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout
}

func Constructor(data Data) list.ListItem[Data] {
	inputKeymap := bubblehelp.NewKeymap(2)
	inputKeymap.Style = style.MainHelpStyle
	inputKeymap.NewKeyBinding(keybind.Tab, true)
	inputKeymap.NewKeyBinding(keybind.ShiftTab, true)
	inputKeymap.NewKeyBinding(keybind.Space, true)
	inputKeymap.SetHelpDesc(keybind.Space, lokyn.L("open selection"))
	inputKeymap.NewKeyBinding(keybind.Esc, true)
	inputKeymap.SetHelpDesc(keybind.Esc, lokyn.L("stop editing"))
	inputKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextScriptEditorRulesListItem, inputKeymap)

	w := new(ListItem)

	t := orvyn.GetTheme()
	dts := t.Style(theme.DimTextStyleID)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.titleOrder = orvyn.NewSimpleRenderable("")

	w.titleRuleType = orvyn.NewSimpleRenderable(lokyn.L("Rule type"))
	w.titleRuleType.Style = dts
	w.titleRuleType.SizeConstraint = true

	w.titleAbilityTarget = orvyn.NewSimpleRenderable(lokyn.L("Ability target"))
	w.titleAbilityTarget.Style = dts
	w.titleAbilityTarget.SizeConstraint = true

	w.titleRuleTypeTarget = orvyn.NewSimpleRenderable(lokyn.L("Rule type target"))
	w.titleRuleTypeTarget.Style = dts
	w.titleRuleTypeTarget.SizeConstraint = true
	w.titleRuleTypeTarget.SetActive(false)

	w.titleAbility = orvyn.NewSimpleRenderable(lokyn.L("Ability"))
	w.titleAbility.Style = dts
	w.titleAbility.SizeConstraint = true

	w.btRuleTypePlaceHolder = lokyn.L("Select a rule type")
	w.btRuleType = button.New(w.btRuleTypePlaceHolder)
	w.btRuleType.OnFocusCallback = w.btOnFocus
	w.btRuleType.OnBlurCallback = w.btOnBlur
	w.btRuleType.OnClickedCallback = w.btRuleTypeOnClicked

	w.mvsRuleTypeTarget = multivalueselector.New[cdata.Target]()
	w.mvsRuleTypeTarget.SetValues(cdata.TargetKeys, cdata.Targets)
	w.mvsRuleTypeTarget.Looping = true
	w.mvsRuleTypeTarget.OnBlur()
	w.mvsRuleTypeTarget.SetActive(false)

	w.btAbilityPlaceHolder = lokyn.L("Select an ability")
	w.btAbility = button.New(w.btAbilityPlaceHolder)
	w.btAbility.OnFocusCallback = w.btOnFocus
	w.btAbility.OnBlurCallback = w.btOnBlur
	w.btAbility.OnClickedCallback = w.btAbilityOnClicked

	w.mvsAbilityTarget = multivalueselector.New[cdata.Target]()
	w.mvsAbilityTarget.SetValues(cdata.TargetKeys, cdata.Targets)
	w.mvsAbilityTarget.Looping = true
	w.mvsAbilityTarget.OnBlur()

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.btAbility)
	w.focusManager.Add(w.mvsAbilityTarget)
	w.focusManager.Add(w.btRuleType)
	w.focusManager.Add(w.mvsRuleTypeTarget)

	titles1Layout := layout.NewHBoxGrowLayout(1, 1,
		[]orvyn.Renderable{
			w.titleAbility,
			w.titleRuleType,
		})

	controls1Layout := layout.NewHBoxGrowLayout(1, 1,
		[]orvyn.Renderable{
			w.btAbility,
			w.btRuleType,
		})

	titles2Layout := layout.NewHBoxGrowLayout(1, 1,
		[]orvyn.Renderable{
			w.titleAbilityTarget,
			w.titleRuleTypeTarget,
		})

	controls2Layout := layout.NewHBoxGrowLayout(1, 1,
		[]orvyn.Renderable{
			w.mvsAbilityTarget,
			w.mvsRuleTypeTarget,
		})

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		[]orvyn.Renderable{
			w.titleOrder,
			titles1Layout,
			controls1Layout,
			titles2Layout,
			controls2Layout,
		})

	w.OnBlur()

	w.UpdateData(data)

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
				w.data.Ability = val
				w.data.AbilityCode = val.Code
				w.btAbility.SetLabel(val.Name)

				// manage target mvs
				w.updateAbilityTargetValues()
			}
		}

		bubblehelp.SwitchToPreviousContext()
	}

	cmd := w.focusManager.Update(msg)

	cmds = append(cmds, cmd)

	w.updateData()

	return tea.Batch(cmds...)
}

func (w *ListItem) UpdateData(data Data) {
	w.data = data

	w.titleOrder.SetValue(fmt.Sprintf(lokyn.L("Order : %d"), w.data.Order))

	w.updateAbilityTargetValues()

	w.mvsAbilityTarget.SetSelectedValue(cdata.GetTarget(w.data.AbilityTarget))

	if w.data.RuleTypeTarget != nil {
		w.mvsRuleTypeTarget.SetSelectedValue(cdata.GetTarget(*w.data.RuleTypeTarget))
	}

	w.updateData()

	abilityName := w.btAbilityPlaceHolder
	ruleTypeName := w.btRuleTypePlaceHolder

	if w.data.Ability.Name != "" {
		abilityName = w.data.Ability.Name
	}

	if w.data.RuleTypeName != "" {
		ruleTypeName = w.data.RuleTypeName
	}

	w.btAbility.SetLabel(abilityName)
	w.btRuleType.SetLabel(ruleTypeName)
}

func (w *ListItem) GetData() Data {
	return w.data
}

// updateData updates the data based on the widgets values
func (w *ListItem) updateData() {
	scriptTarget := w.mvsAbilityTarget.GetSelectedValue().ScriptTarget

	if w.data.AbilityTarget != scriptTarget {
		w.data.AbilityTarget = scriptTarget

	}

	if scriptTarget == api.TargetSelf || w.data.Ability.TargetGroup {
		w.mvsRuleTypeTarget.SetActive(true)
		w.titleRuleTypeTarget.SetActive(true)

		ruleTarget := w.mvsRuleTypeTarget.GetSelectedValue().ScriptTarget

		if w.data.RuleTypeTarget == nil {
			w.data.RuleTypeTarget = &ruleTarget
		} else if *w.data.RuleTypeTarget != ruleTarget {
			w.data.RuleTypeTarget = &ruleTarget
		}
	} else {
		w.mvsRuleTypeTarget.SetActive(false)
		w.titleRuleTypeTarget.SetActive(false)
	}

}

func (w *ListItem) Resize(size orvyn.Size) {
	size.Height = 11

	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
}

func (w *ListItem) Render() string {
	contentSize := w.GetContentSize()

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(w.layout.Render())
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
	orvyn.OpenDialog("selectAbility", abilityselection.New(), nil)

	return nil
}

func (w *ListItem) updateAbilityTargetValues() {
	keys, data := cdata.GetFilteredTargets(w.data.Ability.CanTargetSelf,
		w.data.Ability.CanTargetAllies, w.data.Ability.CanTargetEnemies)

	w.mvsAbilityTarget.SetValues(keys, data)
	w.mvsAbilityTarget.SetSelected(0)

	if len(keys) == 1 {
		w.mvsAbilityTarget.ShowControls = false
		w.focusManager.RemoveWidget(w.mvsAbilityTarget)
	} else {
		w.mvsAbilityTarget.ShowControls = true
		w.focusManager.Insert(1, w.mvsAbilityTarget)
	}
}
