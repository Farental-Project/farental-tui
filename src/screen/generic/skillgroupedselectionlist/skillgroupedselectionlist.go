package skillgroupedselectionlist

import (
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/help"
	"farental/widget/skillgrouplistitem"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

// TODO : Function to get active list and refactor all "if-else" statements.

type Screen[T any] struct {
	title *orvyn.SimpleRenderable

	groupList   *widgetlist.Widget[skillgrouplistitem.Data[T]]
	elementList *widgetlist.Widget[T]

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	submitScreenID orvyn.ScreenID

	groupSelection bool

	loadDataCallback func()
	submitCallback   func() bool
}

func New[T any](title string, constructor widgetlist.ItemConstructor[T],
	loadDataCallback func(), submitCallback func() bool) Screen[T] {
	s := Screen[T]{}

	s.submitScreenID = ""

	s.loadDataCallback = loadDataCallback
	s.submitCallback = submitCallback

	s.title = orvyn.NewSimpleRenderable(title)
	s.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	s.groupList = widgetlist.New(skillgrouplistitem.Constructor[T])

	s.groupList.SetPreferredSize(orvyn.NewSize(orvyn.GetTheme().Size(ftheme.LayoutWidthSizeID), 13))
	s.groupList.SetMinSize(orvyn.NewSize(6, 13))

	s.elementList = widgetlist.New(constructor)

	s.elementList.SetPreferredSize(orvyn.NewSize(orvyn.GetTheme().Size(ftheme.LayoutWidthSizeID), 13))
	s.elementList.SetMinSize(orvyn.NewSize(6, 13))

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.SetGroupSelection(true)

	stack := layout.NewPileLayout(
		s.groupList,
		s.elementList,
	)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4),
			2,
			s.title,
			orvyn.VGap,
			stack,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen[T]) OnEnter(i any) tea.Cmd {
	s.loadDataCallback()

	// loadDataCallback calls SetItems, which may shrink the group list.
	// widgetlist.SetItems does not reset the focus manager's tabIndex, so
	// reset the cursor/filter here (as done for elementList below) to avoid
	// a stale tabIndex indexing past the new, shorter group list on the next
	// Update.
	s.groupList.FocusFirst()
	s.groupList.Init()

	s.elementList.FocusFirst()
	s.elementList.Init()

	s.statusMessage.Reset()

	s.SetGroupSelection(true)

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
			if s.GetFilteringState() == widgetlist.Unfiltered {
				if !s.groupSelection {
					s.SetGroupSelection(true)
					return nil
				}

				return orvyn.SwitchToPreviousScreen()
			}

		case key.Matches(msg, keybind.Enter):
			if s.GetFilteringState() != widgetlist.Filtering {
				if s.groupSelection {
					if s.groupList.GetSelectedItem().SkillName != "" {
						s.elementList.SetItems(s.groupList.GetSelectedItem().Items)
						s.elementList.FocusFirst()

						s.SetGroupSelection(!s.groupSelection)
					}

					return nil
				}

				if s.submitCallback() {
					if len(s.submitScreenID) > 0 {
						return orvyn.SwitchScreen(s.submitScreenID)
					}

					return orvyn.SwitchToPreviousScreen()
				}

				return nil
			}

		case key.Matches(msg, keybind.Help):
			if s.GetFilteringState() != widgetlist.Filtering {
				bubblehelp.ShowAll = !bubblehelp.ShowAll

				return nil
			}
		}
	}

	if s.groupSelection {
		return s.groupList.Update(msg)
	}

	return s.elementList.Update(msg)
}

func (s *Screen[T]) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen[T]) SetItems(items []skillgrouplistitem.Data[T]) {
	s.groupList.SetItems(items)
}

func (s *Screen[T]) SetStatusError(err error) {
	s.statusMessage.SetError(err)
}

func (s *Screen[T]) SetStatusSuccess(str string) {
	s.statusMessage.SetMessage(str, statusmessage.SuccessMessage)
}

func (s *Screen[T]) GetSelectedItem() T {
	return s.elementList.GetSelectedItem()
}

func (s *Screen[T]) GetFilteringState() widgetlist.FilterState {
	if s.groupSelection {
		return s.groupList.FilterState()
	}

	return s.elementList.FilterState()
}

func (s *Screen[T]) SetSubmitScreenID(id orvyn.ScreenID) {
	s.submitScreenID = id
}

func (s *Screen[T]) SetTitle(t string) {
	s.title.SetValue(t)
}

func (s *Screen[T]) FocusFirst() {
	if s.groupSelection {
		s.groupList.FocusFirst()
		return
	}

	s.elementList.FocusFirst()
}

func (s *Screen[T]) SetGroupSelection(b bool) {
	if b {
		s.groupSelection = true
		s.groupList.SetActive(true)
		s.elementList.SetActive(false)
	} else {
		s.groupSelection = false
		s.groupList.SetActive(false)
		s.elementList.SetActive(true)
	}
}
