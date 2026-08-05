package auctioninspect

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/auctiondetails"
	"farental/widget/help"
	"farental/widget/iteminspect"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

// stretchedHeight is the preferred height the auction box declares so the
// layout treats the box row as flexible rather than fixed. Any value above the
// box's own row count has the same effect - the row is the only flexible
// element, so it receives all of the leftover height and the layout clamps it
// to what the terminal actually has. It is deliberately larger than any
// plausible terminal so the row always reads as "wants more".
const stretchedHeight = 200

type Screen struct {
	title *orvyn.SimpleRenderable

	details   *auctiondetails.Widget
	inspector *iteminspect.Widget

	help *help.Widget

	layout *layout.CenterLayout

	auction api.AuctionResponse
}

var _ orvyn.Screen = (*Screen)(nil)

func New(auction api.AuctionResponse) *Screen {
	s := new(Screen)

	s.auction = auction

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.details = auctiondetails.New()
	s.inspector = iteminspect.New()

	// The box needs only its rows, but the dialog wants both outlines to fill
	// the height it was given instead of sitting content-sized in the middle.
	s.details.SetPreferredSize(orvyn.NewSize(1, stretchedHeight))

	s.help = help.New()

	// The auction box holds a label and a right-flushed value; 0.40 of the 109
	// columns left after the layout margin and the gap gives it 43, which fits
	// the longest translated label. The compensator index is 1, so the rounding
	// remainder widens the item box rather than the auction one.
	elements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.40, s.details),
		layout.NewFixedRatioRenderable(0.60, s.inspector),
	}

	inspectLayout := layout.NewHBoxFixedRatioLayout(0, 1, 1, elements...)

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(10, t.Size(ftheme.LayoutWidthSizeID),
			orvyn.NewSize(10, 4),
			s.title,
			orvyn.VGap,
			inspectLayout,
			orvyn.VGap,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextBackAndQuit)

	s.title.SetValue(fmt.Sprintf("%s x%d",
		s.auction.Item.Name, s.auction.Quantity))

	s.inspector.UpdateData(&s.auction.Item)
	s.details.UpdateData(&s.auction)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if k, ok := orvyn.GetKeyMsg(msg); ok {
		if key.Matches(k, keybind.Esc) {
			return orvyn.CloseDialog()
		}
	}

	return nil
}
