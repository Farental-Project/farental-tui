package scriptrulelistitem

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

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data *api.ScriptRuleBody

	titleOrder    *orvyn.SimpleRenderable
	titleRuleType *orvyn.SimpleRenderable
	titleTarget   *orvyn.SimpleRenderable
	titleAbility  *orvyn.SimpleRenderable

	btRuleType *button.Widget
	mvsTarget  *multivalueselector.Widget[cdata.Target]
	btAbility  *button.Widget

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout

	style lipgloss.Style

	contentSize orvyn.Size
}

func Constructor(data *api.ScriptRuleBody) list.IListItem {
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

	w := new(Widget)

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

	w.btRuleType = button.New(lokyn.L("Select a rule type"))
	w.btRuleType.OnFocusCallback = w.btOnFocus
	w.btRuleType.OnBlurCallback = w.btOnBlur
	w.btRuleType.OnClickedCallback = w.btRuleTypeOnClicked

	w.btAbility = button.New(lokyn.L("Select an ability"))
	w.btAbility.OnFocusCallback = w.btOnFocus
	w.btAbility.OnBlurCallback = w.btOnBlur
	w.btAbility.OnClickedCallback = w.btAbilityOnClicked

	w.mvsTarget = multivalueselector.New[cdata.Target]()
	w.mvsTarget.SetValues(cdata.TargetKeys, cdata.Targets)
	w.mvsTarget.Style = multivalueselector.Style{
		FocusedWidget:  t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget:  t.Style(theme.BlurredWidgetStyleID),
		BlurredControl: dts,
		FocusedControl: t.Style(theme.HighlightTextStyleID),
		BlurredValue:   dts,
		FocusedValue:   t.Style(theme.NormalTextStyleID),
	}
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

	w.updateWidgets()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "selectRuleType":
			val, ok := msg.Param.(api.ScriptRuleTypeResponse)

			if ok {
				w.btRuleType.SetLabel(val.Name)
			}

		case "selectAbility":
			val, ok := msg.Param.(api.AbilityResponse)

			if ok {
				w.btAbility.SetLabel(val.Name)
			}
		}

		bubblehelp.SwitchToPreviousContext()
	}

	cmd := w.focusManager.Update(msg)

	w.updateData()

	return cmd
}

func (w *Widget) updateWidgets() {
	// TODO: Init all widgets with data
	w.titleOrder.SetValue(fmt.Sprintf(lokyn.L("Order : %d"), w.data.Order))
	w.mvsTarget.SetSelected(int(w.data.Target))
}

func (w *Widget) updateData() {
	// TODO: Update data
	w.data.Target = w.mvsTarget.GetSelectedValue().ScriptTarget
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 7

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
	w.layout.Resize(size)
}

func (w *Widget) Render() string {
	w.updateWidgets()

	return w.style.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {
	bubblehelp.SwitchContext(keybind.ContextScriptEditorRulesListItem)
	w.focusManager.FocusFirst()
}

func (w *Widget) OnExitInput() {
	w.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(keybind.ContextScriptEditorRulesList)
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}

func (w *Widget) FilterValue() string {
	return ""
}

func (w *Widget) btOnFocus() {
	// TODO: Need to check current keymap
	bubblehelp.SetKeybindVisible(keybind.Space, true)
}

func (w *Widget) btOnBlur() {
	// TODO: Need to check current keymap
	bubblehelp.SetKeybindVisible(keybind.Space, false)
}

func (w *Widget) btRuleTypeOnClicked() tea.Cmd {
	orvyn.OpenDialog("selectRuleType", ruletypeselection.New(), nil)

	return nil
}

func (w *Widget) btAbilityOnClicked() tea.Cmd {
	orvyn.OpenDialog("selectAbility", abilityselection.New(), nil)

	return nil
}
