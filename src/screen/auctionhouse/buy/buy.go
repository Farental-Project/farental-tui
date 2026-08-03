package buy

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
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
		}
	}

	switch msg.(type) {
	case auctionfilter.AppliedMsg:
		s.statusMessage.Reset()
		s.applyFilter()

		return nil
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
func (s *Screen) applyFilter() {
	filter := s.auctionFilter.GetFilter()

	resp, err := helper.Fetch[api.AuctionListResponse](
		request.AuctionGetAll(1, filter))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.filter = filter
	s.page = 1
	s.total = resp.Total

	s.auctionList.SetItems(resp.Auctions)
	s.auctionList.FocusFirst()

	s.reportCount()
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

func (s *Screen) updateCharacterInfo() {
	characterInfo, currency, err := context.RefreshCharacterInfo(true)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.money = currency

	unreadMail := context.RefreshHaveUnreadMail()

	data := characterinfo.ConvertCharacterInfoResponseToData(
		characterInfo, currency, unreadMail)
	s.characterInfo.UpdateData(data)
}
