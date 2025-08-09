package maildetaileditor

import (
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/style"
	"farental/widget/mailattachmentlist"
	"farental/widget/statusmessage"
	"farental/widget/textinput"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	tiMoney         *textinput.Widget
	tiMoneyStatus   *statusmessage.Widget
	attachmentsList *mailattachmentlist.Widget

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout

	contentSize orvyn.Size
}

func New() *Widget {
	editModeKeymap := bubblehelp.NewKeymap(2)
	editModeKeymap.Style = style.MainHelpStyle
	editModeKeymap.NewKeyBinding(keybind.Tab, true)
	editModeKeymap.NewKeyBinding(keybind.ShiftTab, true)
	editModeKeymap.NewKeyBinding(keybind.Esc, true)
	editModeKeymap.SetHelpDesc(keybind.Esc, lang.L("stop editing"))
	editModeKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextMailDetailEditorEditMode, editModeKeymap)

	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.tiMoney = textinput.New()
	w.tiMoney.Placeholder = lang.L("Money amount to send")
	w.tiMoney.Validate = helper.NumericalValidate

	w.tiMoneyStatus = statusmessage.New()

	w.attachmentsList = mailattachmentlist.New(ListItemDelegate{})

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiMoney)
	w.focusManager.Add(w.attachmentsList)

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		[]orvyn.Renderable{
			w.tiMoney,
			w.tiMoneyStatus,
			w.attachmentsList,
		},
	)

	return w
}

func (w *Widget) Init() tea.Cmd {
	cmd := w.tiMoney.Init()
	w.attachmentsList.Init()

	w.focusManager.BlurCurrent()

	return cmd
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

	if w.tiMoney.Err != nil {
		w.tiMoneyStatus.SetError(w.tiMoney.Err)
	} else {
		w.tiMoneyStatus.Reset()
	}

	return cmd
}

func (w *Widget) OnFocus() {
	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)
}

func (w *Widget) OnBlur() {}

func (w *Widget) OnEnterInput() {
	w.focusManager.Focus(0)
	bubblehelp.SwitchContext(keybind.ContextMailDetailEditorEditMode)
}

func (w *Widget) OnExitInput() {
	w.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}
