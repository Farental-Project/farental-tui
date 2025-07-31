package travel

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/model"
	"farental/style"
	"farental/widget/help"
	"farental/widget/list"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
)

type Screen struct {
	orvyn.BaseScreen

	title *orvyn.SimpleRenderable

	travels []tealist.Item
	list    *list.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(
		style.TitleStyle.Render(lang.L("Travels")),
	)

	s.list = list.New(ItemDelegate{},
		[]tealist.Item{})

	s.list.SetShowStatusBar(false)
	s.list.SetShowHelp(false)
	s.list.SetShowTitle(false)
	s.list.SetShowPagination(false)

	s.list.PreferredSize.Width = 45
	s.list.MinSize.Height = 13

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxLayout(10,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.list,
				s.statusMessage,
				orvyn.VGap,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	bubblehelp.SwitchContext(model.ContextFilterSelectionListBasic)

	s.loadTravels()
	s.list.Select(0)

	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()
		}
	}

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadTravels() {
	var travels *[]api.TravelResponse
	var ok bool

	resp, err := helper.SendRequest(request.TravelGetAvailable())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	travels, ok = resp.Result().(*[]api.TravelResponse)

	if !ok {
		return
	}

	s.travels = make([]tealist.Item, 0)

	for _, t := range *travels {
		s.travels = append(s.travels, NewItem(&t))
	}

	s.list.SetItems(s.travels)
}
