package locationinfo

import (
	"farental/core/data/api"
	"farental/internal/context"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/card"
	"farental/widget/help"
	"farental/widget/simplelogviewer"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	description *orvyn.SimpleRenderable

	featuresList *simplelogviewer.Widget

	continentCard *card.Widget

	typeCard *card.Widget

	biomeCard *card.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()
	ts := t.Style(theme.TitleStyleID)
	ws := t.Style(theme.BlurredWidgetStyleID)
	ns := t.Style(theme.NormalTextStyleID)

	logStyle := simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: ws,
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = ts

	s.description = orvyn.NewSimpleRenderable("")
	s.description.SizeConstraint = true
	s.description.Style = ns.Border(ws.GetBorderStyle()).
		BorderForeground(ws.GetBorderTopForeground()).
		AlignHorizontal(lipgloss.Center)

	s.featuresList = simplelogviewer.New(lokyn.L("Features"))
	s.featuresList.Style = logStyle
	s.featuresList.SetAutoScroll(false)

	s.continentCard = card.New("", "")

	s.typeCard = card.New("", "")

	s.biomeCard = card.New("", "")

	s.help = help.New()

	cardLayout := layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 0,
		s.continentCard,
		s.typeCard,
		s.biomeCard,
	)

	infoLayout := layout.NewHBoxGrowFullHeightLayout(1, 0,
		s.featuresList,
		cardLayout,
	)
	infoLayout.Align = lipgloss.Top

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(
			35,
			t.Size(ftheme.LayoutWidthSizeID),
			10,
			s.title,
			orvyn.VGap,
			s.description,
			infoLayout,
			orvyn.VGap,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(_ any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextBackAndQuit)

	s.updateData(&context.CharacterInfo.Location)

	s.featuresList.OnFocus()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()
		}
	}

	cmd := s.featuresList.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) updateData(data *api.LocationResponse) {
	var featureLine strings.Builder

	t := orvyn.GetTheme()
	ts := t.Style(theme.TitleStyleID)
	ds := t.Style(theme.DimTextStyleID)

	s.title.SetValue(data.Name)
	s.description.SetValue(data.Description)

	s.continentCard.Title = data.Continent.Name
	s.continentCard.Content = data.Continent.Description

	s.typeCard.Title = data.Type.Name
	s.typeCard.Content = data.Type.Description

	s.biomeCard.Title = data.Biome.Name
	s.biomeCard.Content = data.Biome.Description

	s.featuresList.SetContent([]string{})

	if len(data.Features) == 0 {
		s.featuresList.AppendContent(ds.Render(lokyn.L("No features in this location")))
	} else {
		for i, f := range data.Features {
			featureLine.Reset()

			if i > 0 {
				featureLine.WriteString("\n")
			}

			fmt.Fprintf(&featureLine, "%s\n", ts.Render(f.Name))
			featureLine.WriteString(ds.Render(f.Description))

			s.featuresList.AppendContent(featureLine.String())
		}
	}
}
