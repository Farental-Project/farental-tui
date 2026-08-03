package auctionbid

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/help"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	lblInfo *label.Widget
	tiBid   *textinput.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	auction api.AuctionResponse
	money   int

	submitted bool
}

var _ orvyn.Screen = (*Screen)(nil)

func New(auction api.AuctionResponse, money int) *Screen {
	s := new(Screen)

	s.auction = auction
	s.money = money

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.lblInfo = label.New("")

	s.tiBid = textinput.New()
	s.tiBid.Validate = helper.NumericalValidate
	s.tiBid.Prompt = string(art.CharGrynars)

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(10, t.Size(ftheme.LayoutWidthSizeID),
			orvyn.NewSize(10, 4),
			s.title,
			s.lblInfo,
			s.tiBid,
			s.statusMessage,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextNavEnterEsc)

	s.submitted = false

	s.title.SetValue(fmt.Sprintf("%s x%d", s.auction.Item.Name, s.auction.Quantity))

	s.lblInfo.SetValue(fmt.Sprintf("%s %d%c | %s %d%c",
		lokyn.L("Current bid"), s.auction.CurrentBid, art.CharGrynars,
		lokyn.L("You have"), s.money, art.CharGrynars))

	s.tiBid.SetValue(strconv.Itoa(s.auction.CurrentBid + 1))
	s.tiBid.OnFocus()

	s.statusMessage.Reset()

	return nil
}

// OnExit returns the bid to place, 0 when the dialog was cancelled.
func (s *Screen) OnExit() any {
	if !s.submitted {
		return 0
	}

	bid, err := strconv.Atoi(s.tiBid.Value())

	if err != nil {
		return 0
	}

	return bid
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if k, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.CloseDialog()

		case key.Matches(k, keybind.Enter):
			bid, err := strconv.Atoi(s.tiBid.Value())

			if err != nil || bid <= s.auction.CurrentBid {
				s.statusMessage.SetMessage(
					fmt.Sprintf(lokyn.L("Your bid must be above %d%c"),
						s.auction.CurrentBid, art.CharGrynars),
					statusmessage.WarningMessage)

				return nil
			}

			s.submitted = true

			return orvyn.CloseDialog()
		}
	}

	return s.tiBid.Update(msg)
}
