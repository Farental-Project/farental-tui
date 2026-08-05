package manage

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/widget/auctionlistitem"
	"farental/widget/characterinfo"
	"farental/widget/help"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

// ownIndex is the own-listings list's position in the screen's focus manager.
// Cancelling is offered only while it holds focus.
const ownIndex = 0

type Screen struct {
	title *orvyn.SimpleRenderable

	characterInfo *characterinfo.Widget

	lblOwn  *label.Widget
	ownList *widgetlist.Widget[api.AuctionResponse]

	lblBids  *label.Widget
	bidsList *widgetlist.Widget[api.AuctionResponse]

	statusMessage *statusmessage.Widget
	help          *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout
}

var _ orvyn.Screen = (*Screen)(nil)

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("manage")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.characterInfo = characterinfo.New()

	s.lblOwn = label.New("")
	s.ownList = widgetlist.New(auctionlistitem.Constructor)
	s.ownList.SetFilterable(false)
	s.ownList.SetMinSize(orvyn.NewSize(20, 6))

	s.lblBids = label.New("")
	s.bidsList = widgetlist.New(auctionlistitem.Constructor)
	s.bidsList.SetFilterable(false)
	s.bidsList.SetMinSize(orvyn.NewSize(20, 6))

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()

	// Each pane is its own column: a heading naming what it holds, and the
	// list growing into whatever height is left.
	panels := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.5,
			layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
				s.lblOwn, s.ownList)),
		layout.NewFixedRatioRenderable(0.5,
			layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
				s.lblBids, s.bidsList)),
	}

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			2,
			s.title,
			s.characterInfo,
			layout.NewHBoxFixedRatioLayout(0, 1, 1, panels...),
			s.statusMessage,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextAuctionManage)

	s.title.SetValue(lokyn.L("Auction house"))

	s.focusManager.SetWidgets([]orvyn.Focusable{s.ownList, s.bidsList})
	s.focusManager.FocusFirst()

	s.statusMessage.Reset()

	s.updateCharacterInfo()
	s.loadLists()

	return nil
}

func (s *Screen) OnExit() any {
	s.focusManager.BlurCurrent()

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if k, ok := orvyn.GetKeyMsg(msg); ok {
		s.statusMessage.Reset()

		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()
		}
	}

	return s.focusManager.Update(msg)
}

// loadLists fetches both panes. They are independent, so one failing leaves the
// other populated rather than blanking the screen.
//
// Nothing here reports success: the counts live in the headings, so there is no
// status line to weigh against an error an earlier step may have set.
func (s *Screen) loadLists() {
	s.loadOwn()
	s.loadBids()

	s.updateLabels()
}

func (s *Screen) loadOwn() bool {
	res, err := helper.Fetch[[]api.AuctionResponse](request.AuctionGetOwn())

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	s.ownList.SetItems(*res)
	s.ownList.FocusFirst()

	return true
}

func (s *Screen) loadBids() bool {
	res, err := helper.Fetch[[]api.AuctionResponse](request.AuctionGetBids())

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	s.bidsList.SetItems(*res)
	s.bidsList.FocusFirst()

	return true
}

// updateLabels carries each pane's count in its heading, so an empty pane reads
// as empty rather than as a pane that failed to load.
func (s *Screen) updateLabels() {
	s.lblOwn.SetValue(fmt.Sprintf("%s (%d)",
		lokyn.L("Your auctions"), s.ownList.Length()))

	s.lblBids.SetValue(fmt.Sprintf("%s (%d)",
		lokyn.L("Your winning bids"), s.bidsList.Length()))
}

func (s *Screen) updateCharacterInfo() bool {
	characterInfo, currency, err := context.RefreshCharacterInfo(true)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	unreadMail := context.RefreshHaveUnreadMail()

	data := characterinfo.ConvertCharacterInfoResponseToData(
		characterInfo, currency, unreadMail)
	s.characterInfo.UpdateData(data)

	return true
}
