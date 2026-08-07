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
	// search is the applied local search, held for the same reason: a reload
	// must use what was committed, not an unconfirmed edit.
	search string
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
	bubblehelp.SwitchContext(keybind.ContextAuctionBuy)

	s.title.SetValue(lokyn.L("Auction house"))

	s.focusManager.SetWidgets([]orvyn.Focusable{s.auctionFilter, s.auctionList})
	s.focusManager.FocusFirst()

	// SwitchContext above reset every binding back to visible.
	s.updateFocusKeybinds()

	s.statusMessage.Reset()

	s.auctionFilter.Init()

	infoOK := s.updateCharacterInfo()
	optionsOK := s.loadFilterOptions()

	// Reporting the count would overwrite whichever of the two calls above set
	// an error, so it only happens when both succeeded (FINDING 3).
	s.applyFilter(infoOK && optionsOK)

	return nil
}

func (s *Screen) OnExit() any {
	s.focusManager.BlurCurrent()

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

// Update wraps the real handler so updateFocusKeybinds runs on every path.
// Several branches return early - the dialog exits in particular, which restore
// the previous context and so reset every binding back to visible - and each of
// them can leave the help line advertising keys the focused pane does not serve.
func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.update(msg)

	s.updateFocusKeybinds()

	return cmd
}

func (s *Screen) update(msg tea.Msg) tea.Cmd {
	if k, ok := orvyn.GetKeyMsg(msg); ok {
		s.statusMessage.Reset()

		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(k, keybind.MKey):
			if s.listFocused() {
				s.statusMessage.Reset()
				s.loadMore()

				return nil
			}

		case key.Matches(k, keybind.RKeyCtrl):
			s.statusMessage.Reset()
			s.auctionFilter.Reset()
			s.applyFilter(true)

			return nil

		case key.Matches(k, keybind.RKey):
			if !s.searchFocused() {
				// Reloads with the applied filter/search (s.filter, s.search),
				// not applyFilter: an unconfirmed panel edit must not be committed.
				s.statusMessage.Reset()
				s.reload(s.filter, s.search, true)

				return nil
			}

		case key.Matches(k, keybind.IKey):
			// Inspecting is read-only, so it stays available with the panel
			// focused - except over the search box, where it's typed text.
			if !s.searchFocused() && s.auctionList.VisibleLength() > 0 {
				return orvyn.OpenDialog(dialogIDInspect,
					auctioninspect.New(s.auctionList.GetSelectedItem()), nil)
			}

		case key.Matches(k, keybind.BKey):
			if s.listFocused() && s.auctionList.VisibleLength() > 0 {
				return s.openBuyConfirm()
			}

		case key.Matches(k, keybind.Enter):
			if s.listFocused() && s.auctionList.VisibleLength() > 0 {
				return orvyn.OpenDialog(dialogIDBid,
					auctionbid.New(s.auctionList.GetSelectedItem(), s.money), nil)
			}
		}
	}

	switch msg := msg.(type) {
	case auctionfilter.AppliedMsg:
		s.statusMessage.Reset()
		s.applyFilter(true)

		return nil

	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case dialogIDInspect:
			// auctioninspect switches context on entry and never switches back.
			bubblehelp.SwitchToPreviousContext()

			return nil

		case dialogIDBid:
			// auctionbid switches context on entry and never switches back.
			bubblehelp.SwitchToPreviousContext()

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

// searchFocused reports whether the panel's search box has the cursor: the
// screen claims several letters before delegating, and there they're text.
func (s *Screen) searchFocused() bool {
	return s.auctionFilter.SearchFocused()
}

// updateFocusKeybinds re-advertises the keys whose availability or meaning
// depends on which pane holds focus, so the help line only ever offers what the
// focused pane will actually do.
//
// bubblehelp clears visibility and custom descriptions whenever a context is
// switched (Keymap.Reset), so this cannot be done once at setup - it has to run
// after anything that may have changed focus or context.
//
// Space is deliberately absent: auctionfilter owns it, tying it to its own
// stat/skill button's focus.
//
// m and b need no search-box handling of their own: both are already gated on
// the auction list holding focus, so the panel never reaches them.
func (s *Screen) updateFocusKeybinds() {
	listFocused := s.listFocused()

	// Both act on the selected auction and do nothing from the filter panel.
	bubblehelp.SetKeybindVisible(keybind.MKey, listFocused)
	bubblehelp.SetKeybindVisible(keybind.BKey, listFocused)

	// The panel holds a free-text search box, and these two are single letters:
	// over that box they are text, not commands.
	searchFocused := s.searchFocused()

	bubblehelp.SetKeybindVisible(keybind.RKey, !searchFocused)
	bubblehelp.SetKeybindVisible(keybind.IKey, !searchFocused)

	// Enter serves both panes, but not with the same meaning.
	desc := lokyn.L("apply filter")

	if listFocused {
		desc = lokyn.L("place bid")
	}

	bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, desc)
}

// loadFilterOptions fetches the vocabulary on every entry: the labels are
// localized server-side, so a language change has to be picked up.
func (s *Screen) loadFilterOptions() bool {
	options, err := helper.Fetch[api.AuctionFilterOptionsResponse](
		request.AuctionGetFilterOptions())

	if err != nil {
		// Browsing unfiltered still works, so this degrades rather than blocks.
		s.statusMessage.SetError(err)
		return false
	}

	s.auctionFilter.SetOptions(options)

	return true
}

// applyFilter runs the filter currently in the panel from page 1. mayReport
// gates reportCount: a caller that already put an error on the status line
// (updateCharacterInfo or loadFilterOptions failing) passes false so the count
// does not overwrite it.
func (s *Screen) applyFilter(mayReport bool) bool {
	return s.reload(s.auctionFilter.GetFilter(), s.auctionFilter.GetSearch(),
		mayReport)
}

// reload fetches with filter and search, committing both along with s.page,
// s.total and the list only on success, so a failed fetch leaves them untouched.
func (s *Screen) reload(filter api.AuctionFilter, search string, mayReport bool) bool {
	resp, err := helper.Fetch[api.AuctionListResponse](
		request.AuctionGetAll(1, filter))

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	s.filter = filter
	s.search = search
	s.page = 1
	s.total = resp.Total

	s.auctionList.SetItems(resp.Auctions)
	s.auctionList.SetFilter(search)
	s.auctionList.FocusFirst()

	if mayReport {
		s.reportCount()
	}

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
	if s.auctionList.VisibleLength() == 0 {
		s.statusMessage.SetMessage(lokyn.L("No auction matches these filters"),
			statusmessage.InformationMessage)

		return
	}

	// The loaded count stays in the line while a search narrows it: without it,
	// a short list reads as if there were nothing left for load more to fetch.
	if s.search != "" {
		s.statusMessage.SetMessage(fmt.Sprintf("%d/%d/%d %s",
			s.auctionList.VisibleLength(), s.auctionList.Length(), s.total,
			lokyn.L("auction(s)")),
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

// afterAction reloads from page 1 with the applied filter and search, not
// the panel, since a sold listing shifts every page after it.
func (s *Screen) afterAction(message string) {
	infoOK := s.updateCharacterInfo()
	filterOK := s.reload(s.filter, s.search, infoOK)

	if !infoOK || !filterOK {
		return
	}

	s.statusMessage.SetMessage(message, statusmessage.SuccessMessage)
}
