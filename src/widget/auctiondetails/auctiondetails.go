package auctiondetails

import (
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

// labelColumnRatio is the share of the box width given to the labels; the rest
// holds the right-flushed values. 0.55 is sized off the longest translated
// label rather than the English one - "Current bidder" is 14 characters but the
// French "Enchérisseur actuel" is 19.
const labelColumnRatio = 0.55

// Widget renders the auction facts as a bordered box of aligned label/value
// rows. It is a peer of iteminspect: the border comes from the default widget
// style, so the two boxes match without any extra styling.
type Widget struct {
	orvyn.BaseWidget

	rows []row
}

var _ orvyn.Widget = (*Widget)(nil)

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	return w
}

func (w *Widget) UpdateData(auction *api.AuctionResponse) {
	w.rows = buildRows(auction, time.Now())
}

func (w *Widget) Render() string {
	var b strings.Builder

	t := orvyn.GetTheme()
	contentSize := w.GetContentSize()

	b.WriteString(headerStyle().Width(contentSize.Width).
		Render(lokyn.L("Auction")))

	labelWidth := int(float64(contentSize.Width) * labelColumnRatio)
	valueWidth := max(contentSize.Width-labelWidth, 0)

	labelStyle := t.Style(theme.DimTextStyleID)
	valueStyle := t.Style(theme.NormalTextStyleID)
	moneyStyle := t.Style(theme.HighlightTextStyleID)

	ns := lipgloss.NewStyle()

	for _, r := range w.rows {
		vs := valueStyle

		if r.highlight {
			vs = moneyStyle
		}

		// Width(n) wraps rather than truncates, and a wrapped row would push the
		// box past the height GetMinSize reports. Cut both columns to their
		// budget before styling so neither can wrap.
		label := ansi.Truncate(r.label, labelWidth, "")
		value := ansi.Truncate(r.value, valueWidth, "")

		b.WriteString("\n")
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(labelWidth).Render(labelStyle.Render(label)),
			ns.Width(valueWidth).AlignHorizontal(lipgloss.Right).
				Render(vs.Render(value))))
	}

	return w.GetStyle().
		Width(contentSize.Width).
		Height(max(contentSize.Height, w.contentHeight())).
		Render(b.String())
}

// headerStyle is the box title treatment: the same underlined style iteminspect
// uses for its Stats / Equip conditions / Effects sections.
func headerStyle() lipgloss.Style {
	return orvyn.GetTheme().Style(ftheme.DimUnderlinedTextStyleID)
}

// GetMinSize reports the height the box actually needs: the header plus one
// line per row, plus the frame. GetPreferredSize reports this same value
// unless a caller has set an explicit preferred size to stretch the box - but
// the minimum itself never moves, and Render enforces it as a floor regardless
// of what height the layout hands the box. Width stays at 1 so the widget
// never drives the layout width, matching iteminspect.GetMinSize.
func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(1, w.height())
}

// GetPreferredSize reports the same height as GetMinSize, so the layout treats
// the box as a fixed-height element and hands it exactly its rows - unless a
// caller sets a preferred size explicitly, which is how the auction inspect
// dialog asks the box to stretch and fill the height beside the item inspector.
// Mirrors the override orvyn's SimpleRenderable uses for the same purpose.
func (w *Widget) GetPreferredSize() orvyn.Size {
	if size := w.BaseRenderable.GetPreferredSize(); size != orvyn.NewSize(1, 1) {
		return size
	}

	return orvyn.NewSize(1, w.height())
}

// contentHeight is what the box's content needs: the underlined header plus
// one line per row. The header height is measured rather than assumed so a
// theme that drops the underline does not leave a dead line behind. Render
// uses this as a floor so the box never shrinks below the height it actually
// holds, even when the layout hands it less.
func (w *Widget) contentHeight() int {
	header := lipgloss.Height(headerStyle().Render("X"))

	return header + len(w.rows)
}

// height is what the box needs including its frame.
func (w *Widget) height() int {
	return w.contentHeight() + w.GetStyle().GetVerticalFrameSize()
}
