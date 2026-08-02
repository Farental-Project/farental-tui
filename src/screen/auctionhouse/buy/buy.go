package buy

import (
	"farental/core/data/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Screen struct {
	auctionList *widgetlist.Widget[api.AuctionResponse]
}

var _ orvyn.Screen = (*Screen)(nil)

func (s *Screen) OnEnter(any) tea.Cmd {
	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return nil
}

func (s *Screen) Update(tea.Msg) tea.Cmd {
	return nil
}
