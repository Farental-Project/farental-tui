package auctionlistitem

import (
	"farental/art"
	"farental/core/data/api"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data api.AuctionResponse

	ownBid bool
}

func Constructor(data api.AuctionResponse) widgetlist.ListItem[api.AuctionResponse] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.UpdateData(data)

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data api.AuctionResponse) {
	w.data = data
	w.ownBid = IsOwnBid(data.CurrentBidderName)
}

func (w *Widget) GetData() api.AuctionResponse {
	return w.data
}

func (w *Widget) FilterValue() string {
	return w.data.Item.Name
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	t := orvyn.GetTheme()
	ns := lipgloss.NewStyle()

	left := fmt.Sprintf("%s x%d", w.data.Item.Name, w.data.Quantity)

	bid := fmt.Sprintf("%s %d%c", lokyn.L("bid"), w.data.CurrentBid, art.CharGrynars)

	if w.ownBid {
		bid = fmt.Sprintf("%s (%s)", bid, lokyn.L("you"))
	}

	buy := fmt.Sprintf("%s —", lokyn.L("buy"))

	if w.data.DirectBuyPrice > 0 {
		buy = fmt.Sprintf("%s %d%c", lokyn.L("buy"), w.data.DirectBuyPrice, art.CharGrynars)
	}

	right := strings.Join([]string{
		bid,
		buy,
		EndsIn(w.data.EndTimestamp, time.Now()),
		w.data.SellerName,
	}, "  ")

	width1, width2 := orvyn.DivideSizeFull(contentSize.Width)

	// Width(n) wraps rather than truncates (MaxWidth would then cut the
	// already-wrapped lines, not the text), so a long name or a full right
	// block would push the row past its declared height of 3. Cut both blocks
	// to their column budget before styling so neither can wrap.
	left = ansi.Truncate(left, width1, "")
	right = ansi.Truncate(right, width2, "")

	return w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).
		Render(lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width1).
				AlignHorizontal(lipgloss.Left).
				Render(left),
			ns.Width(width2).
				AlignHorizontal(lipgloss.Right).
				Render(t.Style(theme.DimTextStyleID).Render(right))))
}
