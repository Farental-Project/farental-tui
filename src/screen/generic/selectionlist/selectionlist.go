package selectionlist

import (
	"farental/internal/keybind"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/style"
	"farental/widget/filterablelist"
	"farental/widget/help"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
)

type Screen struct {
	orvyn.BaseScreen

	title *orvyn.SimpleRenderable

	list *filterablelist.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	loadDataCallback func()
	submitCallback   func() bool
}

func New(title string, delegate tealist.ItemDelegate,
	loadDataCallback func(), submitCallback func() bool) Screen {
	s := Screen{}

	s.loadDataCallback = loadDataCallback
	s.submitCallback = submitCallback

	s.title = orvyn.NewSimpleRenderable(
		style.TitleStyle.Render(title),
	)

	s.list = filterablelist.New(delegate, []tealist.Item{})

	s.list.PreferredSize.Width = style.LayoutWidth - 2 // border
	s.list.MinSize.Height = 13

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxFullLayout(10, 2,
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
	s.loadDataCallback()
	s.list.Select(0)

	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s.statusMessage.Reset()

		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			if s.list.FilterState() == tealist.Unfiltered {
				return orvyn.SwitchToPreviousScreen()
			}

		case key.Matches(msg, keybind.Enter):
			if s.list.FilterState() != tealist.Filtering {
				if s.submitCallback() {
					return orvyn.SwitchToPreviousScreen()
				}

				return nil
			}

		case key.Matches(msg, keybind.Help):
			if s.list.FilterState() != tealist.Filtering {
				bubblehelp.ShowAll = !bubblehelp.ShowAll

				return nil
			}
		}
	}

	cmd := s.list.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) SetItems(items []tealist.Item) {
	s.list.SetItems(items)
}

func (s *Screen) SetStatusError(err error) {
	s.statusMessage.SetError(err)
}

func (s *Screen) GetSelectedItem() tealist.Item {
	return s.list.SelectedItem()
}
