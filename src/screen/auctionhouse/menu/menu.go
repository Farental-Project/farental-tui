package menu

import (
	"farental/internal/keybind"
	"farental/screen"
	"farental/widget/button"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	btSell   *button.Widget
	btBuy    *button.Widget
	btManage *button.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout
}

var _ orvyn.Screen = (*Screen)(nil)

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("auction house")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.btSell = button.New("sell")
	s.btSell.OnClickedCallback = s.btSellOnClicked

	s.btBuy = button.New("buy")
	s.btBuy.OnClickedCallback = s.btBuyOnClicked

	s.btManage = button.New("manage")
	s.btManage.OnClickedCallback = s.btManageOnClicked

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.btSell)
	s.focusManager.Add(s.btBuy)
	s.focusManager.Add(s.btManage)

	s.focusManager.NextFocusKeybind = keybind.Down
	s.focusManager.PreviousFocusKeybind = keybind.Up

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(10, 30, orvyn.NewSize(10, 4),
			s.title,
			orvyn.VGap,
			s.btSell,
			s.btBuy,
			s.btManage),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	s.focusManager.FocusFirst()

	s.title.SetValue(lokyn.L("Auction house"))

	s.btSell.SetLabel(lokyn.L("(S)ell"))
	s.btBuy.SetLabel(lokyn.L("(B)uy"))
	s.btManage.SetLabel(lokyn.L("(M)anage"))

	orvyn.SetPreviousScreen(screen.IDDashBoard)

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
		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(k, keybind.SKey):
			return orvyn.SwitchScreen(screen.IDAuctionHouseSell)

		case key.Matches(k, keybind.BKey):
			return orvyn.SwitchScreen(screen.IDAuctionHouseBuy)

		case key.Matches(k, keybind.MKey):
			return orvyn.SwitchScreen(screen.IDAuctionHouseManage)
		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) btSellOnClicked() tea.Cmd {
	return orvyn.SwitchScreen(screen.IDAuctionHouseSell)
}

func (s *Screen) btBuyOnClicked() tea.Cmd {
	return orvyn.SwitchScreen(screen.IDAuctionHouseBuy)
}

func (s *Screen) btManageOnClicked() tea.Cmd {
	return orvyn.SwitchScreen(screen.IDAuctionHouseManage)
}
