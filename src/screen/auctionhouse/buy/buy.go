package buy

import (
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/dialog/auctionbid"
	"farental/screen/dialog/auctioninspect"
	"farental/screen/dialog/popup"
	"farental/widget/auctionfilter"
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
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

// listIndex is the auction list's position in the screen's focus manager.
const listIndex = 1

const (
	dialogIDInspect    orvyn.ScreenID = "auctionInspect"
	dialogIDBid        orvyn.ScreenID = "auctionBid"
	dialogIDBuyConfirm orvyn.ScreenID = "auctionBuyConfirm"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	characterInfo *characterinfo.Widget

	auctionFilter *auctionfilter.Widget
	auctionList   *widgetlist.Widget[api.AuctionResponse]

	statusMessage *statusmessage.Widget
	help          *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	// filter is the applied query, not the one being edited: load more has to
	// page the query that produced the rows on screen.
	filter api.AuctionFilter
	page   int
	total  int64

	money int
}

var _ orvyn.Screen = (*Screen)(nil)

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("buy")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.characterInfo = characterinfo.New()

	s.auctionFilter = auctionfilter.New()

	s.auctionList = widgetlist.New(auctionlistitem.Constructor)
	s.auctionList.SetFilterable(false)
	s.auctionList.SetMinSize(orvyn.NewSize(20, 6))

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()

	panels := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.25, s.auctionFilter),
		layout.NewFixedRatioRenderable(0.75, s.auctionList),
	}

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			3,
			s.title,
			s.characterInfo,
			orvyn.VGap,
			layout.NewHBoxFixedRatioLayout(0, 1, 1, panels...),
			s.statusMessage,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextAuctionBuy)

	s.title.SetValue(lokyn.L("Auction house"))

	s.focusManager.SetWidgets([]orvyn.Focusable{s.auctionFilter, s.auctionList})
	s.focusManager.FocusFirst()

	s.statusMessage.Reset()

	s.auctionFilter.Init()

	s.updateCharacterInfo()
	s.loadFilterOptions()
	s.applyFilter()

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
		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(k, keybind.MKey):
			if s.listFocused() {
				s.statusMessage.Reset()
				s.loadMore()

				return nil
			}

		case key.Matches(k, keybind.IKey):
			if s.listFocused() && s.auctionList.Length() > 0 {
				return orvyn.OpenDialog(dialogIDInspect,
					auctioninspect.New(s.auctionList.GetSelectedItem()), nil)
			}

		case key.Matches(k, keybind.BKey):
			if s.listFocused() && s.auctionList.Length() > 0 {
				return s.openBuyConfirm()
			}

		case key.Matches(k, keybind.Enter):
			if s.listFocused() && s.auctionList.Length() > 0 {
				return orvyn.OpenDialog(dialogIDBid,
					auctionbid.New(s.auctionList.GetSelectedItem(), s.money), nil)
			}
		}
	}

	switch msg := msg.(type) {
	case auctionfilter.AppliedMsg:
		s.statusMessage.Reset()
		s.applyFilter()

		return nil

	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case dialogIDBid:
			bid, ok := msg.Param.(int)

			if ok && bid > 0 {
				s.placeBid(bid)
			}

			return nil

		case dialogIDBuyConfirm:
			answer, ok := msg.Param.(uint)

			if ok && answer == 1 {
				s.directBuy()
			}

			return nil
		}
	}

	return s.focusManager.Update(msg)
}

func (s *Screen) listFocused() bool {
	return s.focusManager.TabIndex() == listIndex
}

// loadFilterOptions fetches the vocabulary on every entry: the labels are
// localized server-side, so a language change has to be picked up.
func (s *Screen) loadFilterOptions() {
	options, err := helper.Fetch[api.AuctionFilterOptionsResponse](
		request.AuctionGetFilterOptions())

	if err != nil {
		// Browsing unfiltered still works, so this degrades rather than blocks.
		s.statusMessage.SetError(err)
		return
	}

	s.auctionFilter.SetOptions(options)
}

// applyFilter runs the filter currently in the panel from page 1, dropping
// whatever was accumulated. The candidate filter is fetched before anything
// is committed: on error, s.filter/s.page/s.total and the list must stay
// exactly as they were, so a later load more keeps paging the query that is
// actually on screen rather than the rejected one.
func (s *Screen) applyFilter() bool {
	filter := s.auctionFilter.GetFilter()

	resp, err := helper.Fetch[api.AuctionListResponse](
		request.AuctionGetAll(1, filter))

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	s.filter = filter
	s.page = 1
	s.total = resp.Total

	s.auctionList.SetItems(resp.Auctions)
	s.auctionList.FocusFirst()

	s.reportCount()

	return true
}

// loadMore appends the next server page, keeping the applied filter.
func (s *Screen) loadMore() {
	if int64(s.auctionList.Length()) >= s.total {
		s.statusMessage.SetMessage(lokyn.L("Everything is already loaded"),
			statusmessage.InformationMessage)

		return
	}

	resp, err := helper.Fetch[api.AuctionListResponse](
		request.AuctionGetAll(s.page+1, s.filter))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.page++
	s.total = resp.Total

	s.auctionList.SetItems(append(s.auctionList.GetItems(), resp.Auctions...))

	s.reportCount()
}

func (s *Screen) reportCount() {
	if s.auctionList.Length() == 0 {
		s.statusMessage.SetMessage(lokyn.L("No auction matches these filters"),
			statusmessage.InformationMessage)

		return
	}

	s.statusMessage.SetMessage(fmt.Sprintf("%d/%d %s",
		s.auctionList.Length(), s.total, lokyn.L("auction(s)")),
		statusmessage.InformationMessage)
}

func (s *Screen) updateCharacterInfo() bool {
	characterInfo, currency, err := context.RefreshCharacterInfo(true)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	s.money = currency

	unreadMail := context.RefreshHaveUnreadMail()

	data := characterinfo.ConvertCharacterInfoResponseToData(
		characterInfo, currency, unreadMail)
	s.characterInfo.UpdateData(data)

	return true
}

func (s *Screen) openBuyConfirm() tea.Cmd {
	auction := s.auctionList.GetSelectedItem()

	if auction.DirectBuyPrice <= 0 {
		s.statusMessage.SetMessage(lokyn.L("This auction has no direct buy price"),
			statusmessage.WarningMessage)

		return nil
	}

	return orvyn.OpenDialog(dialogIDBuyConfirm, popup.NewYesNo(
		fmt.Sprintf(lokyn.L("Buy %s x%d for %d%c ?"),
			auction.Item.Name, auction.Quantity,
			auction.DirectBuyPrice, art.CharGrynars)), nil)
}

func (s *Screen) placeBid(bid int) {
	auction := s.auctionList.GetSelectedItem()

	_, err := helper.SendRequest(request.AuctionMakeBid(api.AuctionBidBody{
		ID:  auction.ID,
		Bid: bid,
	}))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.afterAction(lokyn.L("Bid placed"))
}

func (s *Screen) directBuy() {
	auction := s.auctionList.GetSelectedItem()

	_, err := helper.SendRequest(request.AuctionDirectBuy(api.IDBody{ID: auction.ID}))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.afterAction(lokyn.L("Item bought, check your mail"))
}

// afterAction reloads from page 1: a sold listing is gone and every later page
// shifts under it, so refetching the accumulated pages would show duplicates
// or holes.
func (s *Screen) afterAction(message string) {
	infoOK := s.updateCharacterInfo()
	filterOK := s.applyFilter()

	if !infoOK || !filterOK {
		return
	}

	s.statusMessage.SetMessage(message, statusmessage.SuccessMessage)
}
