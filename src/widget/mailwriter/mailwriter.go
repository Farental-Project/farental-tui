package mailwriter

import (
	"farental/core/data/api"
	"farental/internal/keybind"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/textarea"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	tiReceiver *textinput.Widget
	tiSubject  *textinput.Widget
	taContent  *textarea.Widget

	focusManager *orvyn.FocusManager

	layout *layout.VBoxFullLayout
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.tiReceiver = textinput.New()
	w.tiReceiver.Placeholder = lokyn.L("Receiver name")

	w.tiSubject = textinput.New()
	w.tiSubject.Placeholder = lokyn.L("Subject")

	w.taContent = textarea.New()
	w.taContent.Placeholder = lokyn.L("Mail content")
	w.taContent.ShowLineNumbers = false
	w.taContent.SetMinSize(orvyn.NewSize(10, 3))

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiReceiver)
	w.focusManager.Add(w.tiSubject)
	w.focusManager.Add(w.taContent)

	w.layout = layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0),
		2,
		w.tiReceiver,
		w.tiSubject,
		w.taContent,
	)

	w.OnBlur()

	return w
}

func (w *Widget) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, w.tiReceiver.Init())
	cmds = append(cmds, w.tiSubject.Init())
	cmds = append(cmds, w.taContent.Init())

	w.focusManager.BlurCurrent()

	return tea.Batch(cmds...)
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	return w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	if w.IsInputting() {
		return w.inputUpdate(msg)
	}

	return nil
}

func (w *Widget) inputUpdate(msg tea.Msg) tea.Cmd {
	cmd := w.focusManager.Update(msg)

	return cmd
}

func (w *Widget) OnFocus() {
	w.BaseFocusable.OnFocus()

	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)
}

func (w *Widget) OnEnterInput() tea.Cmd {
	w.focusManager.Focus(0)
	bubblehelp.SwitchContext(keybind.ContextBasicEditMode)

	return nil
}

func (w *Widget) OnExitInput() tea.Cmd {
	w.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)

	return nil
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}

func (w *Widget) GetMailBody() api.MailSendBody {
	return api.MailSendBody{
		Receiver: w.tiReceiver.Value(),
		Subject:  w.tiSubject.Value(),
		Content:  w.taContent.Value(),
	}
}
