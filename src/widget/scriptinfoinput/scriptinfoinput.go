package scriptinfoinput

import (
	"farental/core/data/api"
	"farental/internal/keybind"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/checkbox"
	"github.com/halsten-dev/orvyn/widget/textarea"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Data struct {
	Name        string
	Description string
	Private     bool
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	title       *orvyn.SimpleRenderable
	name        *textinput.Widget
	description *textarea.Widget
	private     *checkbox.Widget

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout

	data *Data
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.data = &Data{
		Name:        "",
		Description: "",
		Private:     false,
	}

	w.title = orvyn.NewSimpleRenderable(lokyn.L("Script information"))
	w.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	w.name = textinput.New()
	w.name.Placeholder = lokyn.L("Script name")

	w.description = textarea.New()
	w.description.Placeholder = lokyn.L("Script description")
	w.description.ShowLineNumbers = false
	w.description.SetMinSize(orvyn.NewSize(10, 5))

	w.private = checkbox.New(lokyn.L("Private"))

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.name)
	w.focusManager.Add(w.description)
	w.focusManager.Add(w.private)

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		w.title,
		w.name,
		w.description,
		w.private,
	)

	w.OnBlur()

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.data = &Data{}

	w.name.SetValue("")
	w.description.SetValue("")
	w.private.SetChecked(true)

	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	cmd := w.focusManager.Update(msg)

	w.updateData()

	return cmd
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) OnFocus() {
	w.BaseFocusable.OnFocus()
	bubblehelp.SwitchContext(keybind.ContextScriptEditorWidgetNormalMode)
}

func (w *Widget) OnEnterInput() {
	bubblehelp.SwitchContext(keybind.ContextBasicEditMode)

	w.focusManager.FocusFirst()
}

func (w *Widget) OnExitInput() {
	bubblehelp.SwitchContext(keybind.ContextScriptEditorWidgetNormalMode)

	w.focusManager.BlurCurrent()
}

func (w *Widget) GetFocusKeybind() *key.Binding {
	return &keybind.Num1Key
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}

func (w *Widget) GetData() Data {
	return *w.data
}

func (w *Widget) SetData(data *api.ScriptBody) {
	w.data = &Data{
		Name:        data.Name,
		Description: data.Description,
		Private:     data.IsPrivate,
	}

	w.updateWidget()
}

func (w *Widget) updateData() {
	w.data.Name = w.name.Value()
	w.data.Description = w.description.Value()
	w.data.Private = w.private.IsChecked()
}

func (w *Widget) updateWidget() {
	w.name.SetValue(w.data.Name)
	w.description.SetValue(w.data.Description)
	w.private.SetChecked(w.data.Private)
}
