package maildetaileditor

import (
	"farental/art"
	"farental/core/data"
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/style"
	ftheme "farental/internal/theme"
	"farental/widget/mailattachmentlist"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	moneyTitle    *orvyn.SimpleRenderable
	tiMoney       *textinput.Widget
	tiMoneyStatus *statusmessage.Widget

	paymentTitle    *orvyn.SimpleRenderable
	tiPayment       *textinput.Widget
	tiPaymentStatus *statusmessage.Widget

	attachmentListTitle *orvyn.SimpleRenderable

	attachmentsList *mailattachmentlist.Widget

	inventory *data.Inventory

	focusManager *orvyn.FocusManager

	layout *layout.VBoxFullLayout
}

func New() *Widget {
	editModeKeymap := bubblehelp.NewKeymap(2)
	editModeKeymap.Style = style.MainHelpStyle
	editModeKeymap.NewKeyBinding(keybind.Tab, true)
	editModeKeymap.NewKeyBinding(keybind.ShiftTab, true)
	editModeKeymap.NewKeyBinding(keybind.Esc, true)
	editModeKeymap.SetHelpDesc(keybind.Esc, lokyn.L("stop editing"))
	editModeKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextMailDetailEditorEditMode, editModeKeymap)

	w := new(Widget)

	t := orvyn.GetTheme()
	ds := t.Style(theme.DimTextStyleID)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.moneyTitle = orvyn.NewSimpleRenderable(lokyn.L("Money"))
	w.moneyTitle.Style = ds

	w.tiMoney = textinput.New()
	w.tiMoney.Prompt = string(art.CharGrynars)
	w.tiMoney.Placeholder = lokyn.L("Money amount to send")
	w.tiMoney.Validate = helper.NumericalValidate

	w.tiMoneyStatus = statusmessage.New()

	w.paymentTitle = orvyn.NewSimpleRenderable(lokyn.L("Payment request"))
	w.paymentTitle.Style = ds

	w.tiPayment = textinput.New()
	w.tiPayment.Prompt = string(art.CharGrynars)
	w.tiPayment.Placeholder = lokyn.L("Money amount to request as payment")
	w.tiPayment.Validate = helper.NumericalValidate

	w.tiPaymentStatus = statusmessage.New()

	w.attachmentListTitle = orvyn.NewSimpleRenderable(lokyn.L("Attachments"))
	w.attachmentListTitle.Style = t.Style(ftheme.DimUnderlinedTextStyleID)
	w.attachmentListTitle.SizeConstraint = true

	w.attachmentsList = mailattachmentlist.New()

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiMoney)
	w.focusManager.Add(w.tiPayment)
	w.focusManager.Add(w.attachmentsList)

	w.layout = layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0),
		7,
		[]orvyn.Renderable{
			w.moneyTitle,
			w.tiMoney,
			w.tiMoneyStatus,
			w.paymentTitle,
			w.tiPayment,
			w.tiPaymentStatus,
			w.attachmentListTitle,
			w.attachmentsList,
		},
	)

	w.OnBlur()

	return w
}

func (w *Widget) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, w.tiMoney.Init())
	cmds = append(cmds, w.tiPayment.Init())

	w.attachmentsList.Init()

	w.inventory = data.NewInventory(data.ConstMailAttachmentStackCount)

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

	if w.tiMoney.Err != nil {
		w.tiMoneyStatus.SetError(w.tiMoney.Err)
	} else {
		w.tiMoneyStatus.Reset()
	}

	if w.tiPayment.Err != nil {
		w.tiPaymentStatus.SetError(w.tiPayment.Err)
	} else {
		w.tiPaymentStatus.Reset()
	}

	return cmd
}

func (w *Widget) OnFocus() {
	w.BaseFocusable.OnFocus()
	bubblehelp.SwitchContext(keybind.ContextMailWidgetNormalMode)
}

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
	if w.tiMoney.Err != nil {
		return 0
	}

	amount, err := strconv.Atoi(w.tiMoney.Value())

	if err != nil {
		return 0
	}

	return amount
}

func (w *Widget) GetPaymentAmount() int {
	if w.tiPayment.Err != nil {
		return 0
	}

	amount, err := strconv.Atoi(w.tiPayment.Value())

	if err != nil {
		return 0
	}

	return amount
}

func (w *Widget) HasAttachments() bool {
	if len(w.attachmentsList.GetItems()) > 0 ||
		w.GetAttachedMoneyAmount() > 0 ||
		w.GetPaymentAmount() > 0 {
		return true
	}

	return false
}

func (w *Widget) AddAttachment(item *api.ItemResponse, amount int) error {
	_, err := w.inventory.AddItem(item, amount)

	w.setListItems()

	return err
}

func (w *Widget) RemoveAttachment(index int) {
	w.inventory.RemoveIndex(index)

	w.setListItems()
}

func (w *Widget) GetAttachments() *data.Inventory {
	return w.inventory
}

func (w *Widget) SetFocusOnAttachmentList() {
	w.focusManager.Focus(2)
}

func (w *Widget) setListItems() {
	w.attachmentsList.SetItems(w.inventory.Stacks)
	w.attachmentsList.FocusFirst()
}
