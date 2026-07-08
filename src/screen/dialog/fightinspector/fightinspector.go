package fightinspector

import (
	"farental/core/data/api"
	"farental/widget/help"
	"farental/widget/simplelogviewer"
	"sort"
	"strings"

	"farental/internal/keybind"
	ftheme "farental/internal/theme"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Screen struct {
	actors     []string
	actorsDesc map[string]string

	title *orvyn.SimpleRenderable

	list *widgetlist.Widget[string]

	description *simplelogviewer.Widget

	help *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout
}

func New(actors []api.FightActorResponse) *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Fight information"))
	s.title.Style = t.Style(theme.TitleStyleID)

	s.list = widgetlist.New(widgetlist.SimpleListItemConstructor)
	s.list.SetFilterable(false)

	logStyle := simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}

	s.description = simplelogviewer.New("Description")
	s.description.Style = logStyle
	s.description.OnBlur()

	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.list)
	s.focusManager.Add(s.description)
	s.focusManager.FocusFirst()

	elements := make([]layout.FixedRatioRenderable, 2)
	elements[0] = layout.NewFixedRatioRenderable(0.30, s.list)
	elements[1] = layout.NewFixedRatioRenderable(0.70, s.description)

	layoutList := layout.NewHBoxFixedRatioLayout(0, 1, 0, elements...)

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(35,
			t.Size(ftheme.LayoutWidthSizeID),
			10,
			s.title,
			orvyn.VGap,
			layoutList,
			orvyn.VGap,
			s.help,
		),
	)

	s.initData(actors)

	return s
}

func (s *Screen) initData(actors []api.FightActorResponse) {
	s.actorsDesc = make(map[string]string, 0)

	for _, a := range actors {
		if a.SpecialDescription {
			_, ok := s.actorsDesc[a.Name]

			if ok {
				continue
			}

			s.actorsDesc[a.Name] = a.Description
			s.actors = append(s.actors, a.Name)
			continue
		}

		_, ok := s.actorsDesc[a.RaceName]

		if ok {
			continue
		}

		s.actorsDesc[a.RaceName] = a.Description
		s.actors = append(s.actors, a.RaceName)
	}

	sort.Strings(s.actors)

	s.list.SetItems(s.actors)
	s.updateDescription()
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextFightInspector)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			return orvyn.CloseDialog()
		}
	}

	cmd := s.focusManager.Update(msg)

	s.updateDescription()

	return cmd
}

func (s *Screen) updateDescription() {
	description, ok := s.actorsDesc[s.list.GetSelectedItem()]
	if ok {
		s.description.SetContent(strings.Split(description, "/n"))
	} else {
		s.description.SetContent([]string{})
	}
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
