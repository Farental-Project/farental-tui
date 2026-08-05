package manage

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/dialog/auctioninspect"
	"farental/screen/dialog/popup"
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

const dialogIDInspect orvyn.ScreenID = "auctionManageInspect"
const dialogIDCancelConfirm orvyn.ScreenID = "auctionManageCancelConfirm"

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

	// SwitchContext above reset every binding back to visible.
	s.updateFocusKeybinds()

	return nil
}

func (s *Screen) OnExit() any {
	s.focusManager.BlurCurrent()

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

// Update wraps the real handler so updateFocusKeybinds runs on every path. The
// dialog exits in particular restore the previous context, which resets every
// binding back to visible, and would otherwise leave the help line offering
// keys the focused pane does not serve.
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

		case key.Matches(k, keybind.IKey):
			list := s.focusedList()

			if list.Length() > 0 {
				return orvyn.OpenDialog(dialogIDInspect,
					auctioninspect.New(list.GetSelectedItem()), nil)
			}

		case key.Matches(k, keybind.RKey):
			s.loadLists()

			return nil

		case key.Matches(k, keybind.CKey):
			if s.ownFocused() && s.ownList.Length() > 0 {
				return s.openCancelConfirm()
			}
		}
	}

	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case dialogIDInspect:
			// auctioninspect switches context on entry and never switches back.
			bubblehelp.SwitchToPreviousContext()

			return nil

		case dialogIDCancelConfirm:
			answer, ok := msg.Param.(uint)

			if ok && answer == 1 {
				s.cancelAuction()
			}

			return nil
		}
	}

	return s.focusManager.Update(msg)
}

// updateFocusKeybinds hides cancelling unless the own-listings pane holds
// focus: the bids pane lists other players' auctions, which this screen has no
// power over.
func (s *Screen) updateFocusKeybinds() {
	bubblehelp.SetKeybindVisible(keybind.CKey, s.ownFocused())
}

// loadLists fetches both panes. They are independent, so one failing leaves the
// other populated rather than blanking the screen.
//
// Reports whether both succeeded, so a caller about to announce success can
// tell it did not: a failed pane has already put an error on the status line,
// and a success message written afterwards would replace it silently.
func (s *Screen) loadLists() bool {
	ownOK := s.loadOwn()
	bidsOK := s.loadBids()

	s.updateLabels()

	return ownOK && bidsOK
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

func (s *Screen) ownFocused() bool {
	return s.focusManager.TabIndex() == ownIndex
}

// focusedList is the pane the keys act on. Both panes hold the same row type,
// so everything read-only can serve either without knowing which is which.
func (s *Screen) focusedList() *widgetlist.Widget[api.AuctionResponse] {
	if s.ownFocused() {
		return s.ownList
	}

	return s.bidsList
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

// openCancelConfirm asks before pulling a listing. Cancelling is the one
// destructive action on this screen and it cannot be undone — the auction is
// gone and the items come back by mail — so the prompt names what it will pull.
//
// The server refuses to cancel a listing that already has a bidder (a business
// 401), so that case is caught here rather than let reach the confirmation:
// on a live house most of a seller's listings have bids, and offering the
// confirm only to refuse it afterwards would make the screen's one destructive
// action misfire on the common case.
func (s *Screen) openCancelConfirm() tea.Cmd {
	auction := s.ownList.GetSelectedItem()

	if auction.CurrentBidderName != "" {
		s.statusMessage.SetMessage(
			lokyn.L("An auction that has received a bid can no longer be pulled"),
			statusmessage.WarningMessage)

		return nil
	}

	return orvyn.OpenDialog(dialogIDCancelConfirm, popup.NewYesNo(
		fmt.Sprintf(lokyn.L("Cancel the auction for %s x%d ?"),
			auction.Item.Name, auction.Quantity)), nil)
}

func (s *Screen) cancelAuction() {
	auction := s.ownList.GetSelectedItem()

	_, err := helper.SendRequest(request.AuctionCancel(api.IDBody{ID: auction.ID}))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	// Both panes are refetched rather than the cancelled row removed locally:
	// the server is the authority on what is still live, and a bid placed on
	// one of the player's other listings meanwhile would otherwise go unseen.
	infoOK := s.updateCharacterInfo()
	listsOK := s.loadLists()

	if !infoOK || !listsOK {
		return
	}

	s.statusMessage.SetMessage(lokyn.L("Auction cancelled, check your mail"),
		statusmessage.SuccessMessage)
}
