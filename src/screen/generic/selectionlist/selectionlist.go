package selectionlist

import (
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/help"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Screen[T any] struct {
	title *orvyn.SimpleRenderable

	list *widgetlist.Widget[T]

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	submitScreenID orvyn.ScreenID

	loadDataCallback func()
	submitCallback   func() bool
}

func new[T any](title string, constructor widgetlist.ItemConstructor[T],
	loadDataCallback func(), submitCallback func() bool) Screen[T] {
	s := Screen[T]{}

	s.submitScreenID = ""

	s.loadDataCallback = loadDataCallback
	s.submitCallback = submitCallback

	s.title = orvyn.NewSimpleRenderable(title)
	s.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	s.list = widgetlist.New(constructor)

	s.list.SetPreferredSize(orvyn.NewSize(orvyn.GetTheme().Size(ftheme.LayoutWidthSizeID), 6))
	s.list.SetMinSize(orvyn.NewSize(6, 6))

	s.list.Filter = widgetlist.BasicFilter

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	return s
}

func New[T any](title string, constructor widgetlist.ItemConstructor[T],
	loadDataCallback func(), submitCallback func() bool) Screen[T] {
	s := new(title, constructor, loadDataCallback, submitCallback)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4),
			2,
			s.title,
			orvyn.VGap,
			s.list,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func NewWithHeader[T any](title string, constructor widgetlist.ItemConstructor[T],
	loadDataCallback func(), submitCallback func() bool, headerLayout orvyn.Layout) Screen[T] {

	s := new(title, constructor, loadDataCallback, submitCallback)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4),
			2,
			s.title,
			headerLayout,
			s.list,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen[T]) OnEnter(i any) tea.Cmd {
	s.loadDataCallback()
	s.list.FocusFirst()
	s.list.Init()

	s.statusMessage.Reset()

	err, ok := i.(error)

	if ok {
		s.SetStatusError(err)
	}

	return nil
}

func (s *Screen[T]) OnExit() any {
	return nil
}

func (s *Screen[T]) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s.statusMessage.Reset()

		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			if s.list.FilterState() == widgetlist.Unfiltered {
				return orvyn.SwitchToPreviousScreen()
			}

		case key.Matches(msg, keybind.Enter):
			if s.list.FilterState() != widgetlist.Filtering {
				if s.submitCallback() {
					if len(s.submitScreenID) > 0 {
						return orvyn.SwitchScreen(s.submitScreenID)
					}

					return orvyn.SwitchToPreviousScreen()
				}

				return nil
			}

		case key.Matches(msg, keybind.Help):
			if s.list.FilterState() != widgetlist.Filtering {
				bubblehelp.ShowAll = !bubblehelp.ShowAll

				return nil
			}
		}
	}

	cmd := s.list.Update(msg)

	return cmd
}

func (s *Screen[T]) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen[T]) SetItems(items []T) {
	// Reset any active filter before swapping items: widgetlist.Widget.SetItems
	// replaces the item list but keeps the old filtered indices, which then
	// point past the new (possibly shorter or empty) list and panic on render.
	if s.list.FilterState() != widgetlist.Unfiltered {
		s.list.Init()
	}

	s.list.SetItems(items)
}

func (s *Screen[T]) SetStatusError(err error) {
	s.statusMessage.SetError(err)
}

func (s *Screen[T]) SetStatusSuccess(str string) {
	s.statusMessage.SetMessage(str, statusmessage.SuccessMessage)
}

func (s *Screen[T]) GetSelectedItem() T {
	return s.list.GetSelectedItem()
}

func (s *Screen[T]) GetFilteringState() widgetlist.FilterState {
	return s.list.FilterState()
}

func (s *Screen[T]) SetSubmitScreenID(id orvyn.ScreenID) {
	s.submitScreenID = id
}

func (s *Screen[T]) SetTitle(t string) {
	s.title.SetValue(t)
}

func (s *Screen[T]) FocusFirst() {
	s.list.FocusFirst()
}
