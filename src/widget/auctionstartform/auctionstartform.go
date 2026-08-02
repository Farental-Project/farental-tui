package auctionstartform

import (
	"farental/art"
	"farental/internal/helper"
	"farental/widget/button"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	btSelectItem *button.Widget

	tiTotalSellQty  *textinput.Widget
	tiSellStackSize *textinput.Widget

	lblStartPrice     *label.Widget
	tiUnitStartPrice  *textinput.Widget
	tiStackStartPrice *textinput.Widget

	lblBuyPrice     *label.Widget
	tiUnitBuyPrice  *textinput.Widget
	tiStackBuyPrice *textinput.Widget

	focusManager *orvyn.FocusManager

	layout *layout.VBoxLayout
}

var _ orvyn.Widget = (*Widget)(nil)
var _ orvyn.Focusable = (*Widget)(nil)

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.btSelectItem = button.New("Select item...")

	w.tiTotalSellQty = textinput.New()
	w.tiTotalSellQty.Placeholder = lokyn.L("Total quantity to sell")
	w.tiTotalSellQty.Validate = helper.NumericalValidate

	w.tiSellStackSize = textinput.New()
	w.tiSellStackSize.Placeholder = lokyn.L("Size of one stack")
	w.tiSellStackSize.Validate = helper.NumericalValidate

	w.tiUnitStartPrice = textinput.New()
	w.tiUnitStartPrice.Prompt = string(art.CharGrynars)
	w.tiUnitStartPrice.Placeholder = lokyn.L("Unit start price")
	w.tiUnitStartPrice.Validate = helper.NumericalValidate

	w.tiStackStartPrice = textinput.New()
	w.tiStackStartPrice.Prompt = string(art.CharGrynars)
	w.tiStackStartPrice.Placeholder = lokyn.L("Stack start price")
	w.tiStackStartPrice.Validate = helper.NumericalValidate

	w.tiUnitBuyPrice = textinput.New()
	w.tiUnitBuyPrice.Prompt = string(art.CharGrynars)
	w.tiUnitBuyPrice.Placeholder = lokyn.L("Unit buy price")
	w.tiUnitBuyPrice.Validate = helper.NumericalValidate

	w.tiStackBuyPrice = textinput.New()
	w.tiStackBuyPrice.Prompt = string(art.CharGrynars)
	w.tiStackBuyPrice.Placeholder = lokyn.L("Unit buy price")
	w.tiStackBuyPrice.Validate = helper.NumericalValidate

	w.lblStartPrice = label.New(lokyn.L("Start price"))
	w.lblBuyPrice = label.New(lokyn.L("Buy price"))

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.btSelectItem)
	w.focusManager.Add(w.tiTotalSellQty)
	w.focusManager.Add(w.tiSellStackSize)
	w.focusManager.Add(w.tiUnitStartPrice)
	w.focusManager.Add(w.tiStackStartPrice)
	w.focusManager.Add(w.tiUnitBuyPrice)
	w.focusManager.Add(w.tiStackBuyPrice)

	sellStack := layout.NewHBoxGrowLayout(1, 0,
		w.tiTotalSellQty, w.tiSellStackSize)

	startPriceStack := layout.NewHBoxGrowLayout(1, 0,
		w.tiUnitStartPrice, w.tiStackStartPrice)

	buyPriceStack := layout.NewHBoxGrowLayout(1, 0,
		w.tiUnitBuyPrice, w.tiStackBuyPrice)

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		w.btSelectItem,
		sellStack,
		w.lblStartPrice,
		startPriceStack,
		w.lblBuyPrice,
		buyPriceStack,
	)

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	cmd := w.focusManager.Update(msg)

	return cmd
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
	w.focusManager.FocusFirst()
}
