package mailwriter

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/style"
	"farental/widget/textarea"
	"farental/widget/textinput"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	contentSize orvyn.Size

	tiReceiver *textinput.Widget
	tiSubject  *textinput.Widget
	taContent  *textarea.Widget

	focusManager *orvyn.FocusManager

	layout *layout.VBoxFullLayout
}

func New() *Widget {
	editModeKeymap := bubblehelp.NewKeymap(2)
	editModeKeymap.Style = style.MainHelpStyle
	editModeKeymap.NewKeyBinding(keybind.Tab, true)
	editModeKeymap.NewKeyBinding(keybind.ShiftTab, true)
	editModeKeymap.NewKeyBinding(keybind.Esc, true)
	editModeKeymap.SetHelpDesc(keybind.Esc, lang.L("stop editing"))
	editModeKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextMailWriterEditMode, editModeKeymap)

	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.tiReceiver = textinput.New()
	w.tiReceiver.Placeholder = lang.L("Receiver name")

	w.tiSubject = textinput.New()
	w.tiSubject.Placeholder = lang.L("Subject")

	w.taContent = textarea.New()
	w.taContent.Placeholder = lang.L("Mail content")
	w.taContent.ShowLineNumbers = false
	w.taContent.MinHeight = 3

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiReceiver)
	w.focusManager.Add(w.tiSubject)
	w.focusManager.Add(w.taContent)

	w.layout = layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0),
		2,
		[]orvyn.Renderable{
			w.tiReceiver,
			w.tiSubject,
			w.taContent,
		},
	)

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
	var s lipgloss.Style

	if w.IsFocused() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	return s.Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
	w.layout.Resize(size)
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
	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)
}

func (w *Widget) OnBlur() {}

func (w *Widget) OnEnterInput() {
	w.focusManager.Focus(0)
	bubblehelp.SwitchContext(keybind.ContextMailWriterEditMode)
}

func (w *Widget) OnExitInput() {
	w.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)
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
