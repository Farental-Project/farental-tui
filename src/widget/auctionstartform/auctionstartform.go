package auctionstartform

import (
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/screen/dialog/itemselection"
	"farental/widget/button"
	"farental/widget/multivalueselector"
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	btSelectItem *button.Widget

	lblTotalSellQty *label.Widget
	lblStackSize    *label.Widget
	tiTotalSellQty  *textinput.Widget
	tiSellStackSize *textinput.Widget

	lblStartPrice     *label.Widget
	tiUnitStartPrice  *textinput.Widget
	tiStackStartPrice *textinput.Widget

	lblBuyPrice     *label.Widget
	tiUnitBuyPrice  *textinput.Widget
	tiStackBuyPrice *textinput.Widget

	mvsDuration *multivalueselector.Widget[api.AuctionDuration]

	lblUnit  *label.Widget
	lblStack *label.Widget

	statusMessage *statusmessage.Widget

	focusManager *orvyn.FocusManager

	data *api.AuctionStartBody

	layout *layout.VBoxLayout
}

var _ orvyn.Widget = (*Widget)(nil)
var _ orvyn.Focusable = (*Widget)(nil)

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.btSelectItem = button.New("")
	w.btSelectItem.OnClickedCallback = w.btSelectItemOnClicked
	w.btSelectItem.OnFocusCallback = w.btOnFocus
	w.btSelectItem.OnBlurCallback = w.btOnBlur

	w.tiTotalSellQty = textinput.New()
	w.tiTotalSellQty.Validate = helper.NumericalValidate

	w.tiSellStackSize = textinput.New()
	w.tiSellStackSize.Validate = helper.NumericalValidate

	w.tiUnitStartPrice = textinput.New()
	w.tiUnitStartPrice.Prompt = string(art.CharGrynars)
	w.tiUnitStartPrice.Validate = helper.NumericalValidate

	w.tiStackStartPrice = textinput.New()
	w.tiStackStartPrice.Prompt = string(art.CharGrynars)
	ftheme.SetUnfocusableTextinputStyle(w.tiStackStartPrice)

	w.tiUnitBuyPrice = textinput.New()
	w.tiUnitBuyPrice.Prompt = string(art.CharGrynars)
	w.tiUnitBuyPrice.Validate = helper.NumericalValidate

	w.tiStackBuyPrice = textinput.New()
	w.tiStackBuyPrice.Prompt = string(art.CharGrynars)
	w.tiStackBuyPrice.Validate = helper.NumericalValidate
	ftheme.SetUnfocusableTextinputStyle(w.tiStackBuyPrice)

	w.mvsDuration = multivalueselector.New[api.AuctionDuration]()
	w.mvsDuration.OnBlur()

	w.lblTotalSellQty = label.New("")
	w.lblStackSize = label.New("")
	w.lblStartPrice = label.New("")
	w.lblBuyPrice = label.New("")
	w.lblUnit = label.New("")
	w.lblStack = label.New("")

	w.statusMessage = statusmessage.New()

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.btSelectItem)
	w.focusManager.Add(w.tiTotalSellQty)
	w.focusManager.Add(w.tiSellStackSize)
	w.focusManager.Add(w.tiUnitStartPrice)
	w.focusManager.Add(w.tiUnitBuyPrice)
	w.focusManager.Add(w.mvsDuration)

	qtyLabels := layout.NewHBoxGrowLayout(0, 0,
		w.lblTotalSellQty, w.lblStackSize)

	priceLabels := layout.NewHBoxGrowLayout(0, 0,
		w.lblUnit, w.lblStack)

	sellStack := layout.NewHBoxGrowLayout(0, 0,
		w.tiTotalSellQty, w.tiSellStackSize)

	startPriceStack := layout.NewHBoxGrowLayout(0, 0,
		w.tiUnitStartPrice, w.tiStackStartPrice)

	buyPriceStack := layout.NewHBoxGrowLayout(0, 0,
		w.tiUnitBuyPrice, w.tiStackBuyPrice)

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		w.btSelectItem,
		qtyLabels,
		sellStack,
		w.lblStartPrice,
		priceLabels,
		startPriceStack,
		w.lblBuyPrice,
		priceLabels,
		buyPriceStack,
		w.mvsDuration,
		orvyn.VGap,
		w.statusMessage,
	)

	w.Reset()
	w.updateData()
	w.autoCalculateFields()

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.Reset()

	w.tiStackBuyPrice.Placeholder = lokyn.L("Unit buy price")
	w.tiUnitBuyPrice.Placeholder = lokyn.L("Unit buy price")
	w.tiStackStartPrice.Placeholder = lokyn.L("Stack start price")
	w.tiUnitStartPrice.Placeholder = lokyn.L("Unit start price")
	w.tiSellStackSize.Placeholder = lokyn.L("Size of one stack")
	w.tiTotalSellQty.Placeholder = lokyn.L("Total quantity to sell")

	w.lblTotalSellQty.SetValue(lokyn.L("Total quantity to sell"))
	w.lblStackSize.SetValue(lokyn.L("Stack size"))
	w.lblStartPrice.SetValue(lokyn.L("Start price"))
	w.lblBuyPrice.SetValue(lokyn.L("Buy price"))
	w.lblUnit.SetValue(lokyn.L("Unit"))
	w.lblStack.SetValue(lokyn.L("Total"))

	w.updateEstimation()

	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "selectItem":
			val, ok := msg.Param.(api.StackResponse)

			if ok {
				w.itemSelected(val)
				bubblehelp.SwitchToPreviousContext()
			}
		}
	}

	cmd := w.focusManager.Update(msg)

	cmds = append(cmds, cmd)

	w.updateData()
	w.autoCalculateFields()

	return tea.Batch(cmds...)
}

// updateData is used to keep the w.data up-to-date with the widget.
func (w *Widget) updateData() {
	w.data.TotalQuantity, _ = strconv.Atoi(w.tiTotalSellQty.Value())
	w.data.StackSize, _ = strconv.Atoi(w.tiSellStackSize.Value())

	w.data.UnitStartBid, _ = strconv.Atoi(w.tiUnitStartPrice.Value())
	w.data.UnitDirectBuyPrice, _ = strconv.Atoi(w.tiUnitBuyPrice.Value())

	w.data.Duration = w.mvsDuration.GetSelectedValue()

	w.updateEstimation()
}

func (w *Widget) autoCalculateFields() {
	w.tiStackStartPrice.SetValue(fmt.Sprintf("%d", w.data.UnitStartBid*w.data.StackSize))
	w.tiStackBuyPrice.SetValue(fmt.Sprintf("%d", w.data.UnitDirectBuyPrice*w.data.StackSize))

	w.fieldValidationErrors()
}
func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	return w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) Reset() {
	w.statusMessage.Reset()

	w.focusManager.FocusFirst()
	w.data = new(api.AuctionStartBody)
	w.data.Duration = api.AuctionDurationLong

	w.btSelectItem.SetLabel(lokyn.L("Select Item"))
	w.tiTotalSellQty.SetValue("0")
	w.tiSellStackSize.SetValue("0")
	w.tiUnitStartPrice.SetValue("0")
	w.tiUnitBuyPrice.SetValue("0")

	w.loadDurations()
	w.mvsDuration.SetSelectedKey(w.data.Duration.RenderValue())

	w.autoCalculateFields()
}

func (w *Widget) btOnFocus() {
	bubblehelp.SetKeybindVisible(keybind.Space, true)
}

func (w *Widget) btOnBlur() {
	bubblehelp.SetKeybindVisible(keybind.Space, false)
}

func (w *Widget) btSelectItemOnClicked() tea.Cmd {
	orvyn.OpenDialog("selectItem", itemselection.New(), nil)

	return nil
}

func (w *Widget) loadDurations() {
	durationValues := make(map[string]api.AuctionDuration, 3)
	keys := make([]string, 3)

	keys[0] = api.AuctionDurationShort.RenderValue()
	durationValues[keys[0]] = api.AuctionDurationShort
	keys[1] = api.AuctionDurationLong.RenderValue()
	durationValues[keys[1]] = api.AuctionDurationLong
	keys[2] = api.AuctionDurationVeryLong.RenderValue()
	durationValues[keys[2]] = api.AuctionDurationVeryLong

	w.mvsDuration.SetValues(keys, durationValues)
}

func (w *Widget) updateEstimation() {
	if w.data == nil || w.data.ItemID <= 0 {
		w.statusMessage.SetMessage(lokyn.L("Please select an item"),
			statusmessage.InformationMessage)
		return
	}

	if w.data.TotalQuantity <= 0 || w.data.StackSize <= 0 {
		w.statusMessage.SetMessage(lokyn.L("Please enter valid quantities"),
			statusmessage.WarningMessage)
		return
	}

	if w.data.UnitStartBid <= 0 {
		w.statusMessage.SetMessage(lokyn.L("Unit start price is mandatory"),
			statusmessage.WarningMessage)
		return
	}

	est, err := helper.Fetch[api.AuctionPlanResponse](request.AuctionEstimate(*w.data))

	if err != nil {
		w.statusMessage.SetError(err)
		return
	}

	estText := fmt.Sprintf("%d %s | %d %s | %c%d %s",
		est.RequestedAuctions, lokyn.L("requested auction(s)"),
		est.Auctions, lokyn.L("auction(s) will be created"),
		art.CharGrynars, est.TotalTax, lokyn.L("total tax"))

	w.statusMessage.SetMessage(estText,
		statusmessage.InformationMessage)
}

func (w *Widget) fieldValidationErrors() {
	if w.tiTotalSellQty.Err != nil {
		w.statusMessage.SetError(w.tiTotalSellQty.Err)
		return
	}

	if w.tiSellStackSize.Err != nil {
		w.statusMessage.SetError(w.tiSellStackSize.Err)
		return
	}

	if w.tiUnitStartPrice.Err != nil {
		w.statusMessage.SetError(w.tiUnitStartPrice.Err)
		return
	}

	if w.tiUnitBuyPrice.Err != nil {
		w.statusMessage.SetError(w.tiUnitBuyPrice.Err)
		return
	}
}

func (w *Widget) itemSelected(stack api.StackResponse) {
	w.data.ItemID = stack.ItemID
	w.btSelectItem.SetLabel(stack.Item.Name)
	w.tiTotalSellQty.SetValue(fmt.Sprintf("%d", stack.Count))
	w.tiSellStackSize.SetValue(fmt.Sprintf("%d", stack.Item.MaxStackCount))
	startPrice := max(1, stack.Item.SellPrice-20)
	w.tiUnitStartPrice.SetValue(fmt.Sprintf("%d", startPrice))
	w.tiUnitBuyPrice.SetValue(fmt.Sprintf("%d", stack.Item.SellPrice))
	w.lblStackSize.SetValue(fmt.Sprintf("%s (max: %d)",
		lokyn.L("Stack size"), stack.Item.MaxStackCount))
	w.lblTotalSellQty.SetValue(fmt.Sprintf("%s (max: %d)",
		lokyn.L("Total quantity to sell"), stack.Count))

	w.updateEstimation()
}

func (w *Widget) GetData() *api.AuctionStartBody {
	return w.data
}
