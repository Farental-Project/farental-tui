package auctioninspect

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/auctionlistitem"
	"farental/widget/help"
	"farental/widget/iteminspect"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	inspector *iteminspect.Widget

	details *orvyn.SimpleRenderable

	help *help.Widget

	layout *layout.CenterLayout

	auction api.AuctionResponse
}

var _ orvyn.Screen = (*Screen)(nil)

func New(auction api.AuctionResponse) *Screen {
	s := new(Screen)

	s.auction = auction

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(auction.Item.Name)
	s.title.Style = t.Style(theme.TitleStyleID)

	s.inspector = iteminspect.New()

	s.details = orvyn.NewSimpleRenderable("")

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(10, t.Size(ftheme.LayoutWidthSizeID),
			orvyn.NewSize(10, 4),
			s.title,
			s.inspector,
			s.details,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextBackAndQuit)

	s.title.SetValue(s.auction.Item.Name)
	s.inspector.UpdateData(&s.auction.Item)
	s.details.SetValue(s.detailLines())

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

func (s *Screen) detailLines() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s: %d\n", lokyn.L("Quantity"), s.auction.Quantity)
	fmt.Fprintf(&b, "%s: %d%c\n", lokyn.L("Current bid"), s.auction.CurrentBid, art.CharGrynars)

	if s.auction.DirectBuyPrice > 0 {
		fmt.Fprintf(&b, "%s: %d%c\n", lokyn.L("Direct buy"), s.auction.DirectBuyPrice, art.CharGrynars)
	}

	bidder := s.auction.CurrentBidderName

	if bidder == "" {
		bidder = lokyn.L("nobody")
	}

	fmt.Fprintf(&b, "%s: %s\n", lokyn.L("Current bidder"), bidder)
	fmt.Fprintf(&b, "%s: %s\n", lokyn.L("Seller"), s.auction.SellerName)
	fmt.Fprintf(&b, "%s: %s", lokyn.L("Ends in"), auctionlistitem.EndsIn(s.auction.EndTimestamp, time.Now()))

	return b.String()
}
