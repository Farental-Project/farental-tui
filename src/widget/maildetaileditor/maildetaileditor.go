package maildetaileditor

import (
	"errors"
	"farental/art"
	"farental/core/data"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/style"
	"farental/widget/mailattachmentlist"
	"farental/widget/statusmessage"
	"farental/widget/textinput"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"strconv"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	tiMoney         *textinput.Widget
	tiMoneyStatus   *statusmessage.Widget
	attachmentsList *mailattachmentlist.Widget

	attachments []ListItem

	focusManager *orvyn.FocusManager

	layout *layout.VBoxFullLayout

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
	w.tiMoney.Prompt = string(art.CharGrynars)
	w.tiMoney.Placeholder = lang.L("Money amount to send")
	w.tiMoney.Validate = helper.NumericalValidate

	w.tiMoneyStatus = statusmessage.New()

	w.attachmentsList = mailattachmentlist.New(ListItemDelegate{})

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiMoney)
	w.focusManager.Add(w.attachmentsList)

	w.layout = layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0),
		2,
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

	w.attachments = make([]ListItem, 0)

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

func (w *Widget) GetAttachedMoneyAmount() int {
	amount, err := strconv.Atoi(w.tiMoney.Value())

	if err != nil {
		return 0
	}

	return amount
}

func (w *Widget) HasAttachments() bool {
	if len(w.attachmentsList.Items()) > 0 ||
		w.GetAttachedMoneyAmount() > 0 {
		return true
	}

	return false
}

func (w *Widget) AddAttachment(item ListItem) (tea.Cmd, error) {
	var existingIndex int

	existingIndex = -1

	for i, a := range w.attachments {
		if a.StackID == item.StackID {
			existingIndex = i
			break
		}
	}

	if existingIndex != -1 {
		w.attachments[existingIndex].Amount += item.Amount
	} else {
		if len(w.attachments) == data.ConstMailAttachmentStackCount {
			return nil, errors.New(fmt.Sprintf(
				lang.L("Mails cannot have more than %d attachments"),
				data.ConstMailAttachmentStackCount))
		}

		w.attachments = append(w.attachments, item)
	}

	cmd := w.setListItems(w.attachments)

	return cmd, nil
}

func (w *Widget) RemoveAttachment(index int) tea.Cmd {
	if index < 0 || index >= len(w.attachments) {
		return nil
	}

	w.attachments = append(w.attachments[:index], w.attachments[index+1:]...)

	cmd := w.setListItems(w.attachments)

	return cmd
}

func (w *Widget) GetAttachments() []ListItem {
	return w.attachments
}

func (w *Widget) SetFocusOnAttachmentList() {
	w.focusManager.Focus(1)
}

func (w *Widget) setListItems(items []ListItem) tea.Cmd {
	var listItems []tealist.Item

	listItems = make([]tealist.Item, 0)

	for _, i := range items {
		listItems = append(listItems, i)
	}

	return w.attachmentsList.SetItems(listItems)
}
