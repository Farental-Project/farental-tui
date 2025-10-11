package craftlistitem

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/numericalselector"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Data struct {
	api.RecipeResponse
	CraftAmount int
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	amountSelector *numericalselector.Widget
	paginator      paginator.Model

	data Data

	style lipgloss.Style

	contentSize orvyn.Size
}

func Constructor(data Data) list.ListItem[Data] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.amountSelector = numericalselector.New(0, 100, 1)
	w.paginator = paginator.New()
	w.paginator.Type = paginator.Dots
	w.paginator.PerPage = 4
	widget.UpdatePaginatorTheme(&w.paginator)
	w.paginator.SetTotalPages(len(data.Ingredients))
	w.paginator.KeyMap.NextPage = keybind.NextPage
	w.paginator.KeyMap.PrevPage = keybind.PrevPage

	w.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	w.amountSelector.Update(msg)
	w.paginator, _ = w.paginator.Update(msg)

	w.data.CraftAmount = w.amountSelector.GetValue()

	return nil
}

func (w *Widget) UpdateData(data Data) {
	w.data = data
}

func (w *Widget) GetData() Data {
	return w.data
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 7

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) Render() string {
	var s lipgloss.Style
	var left strings.Builder
	var right strings.Builder
	var b *strings.Builder
	var ingredientsList strings.Builder
	var top string
	var width int

	size := w.contentSize
	t := orvyn.GetTheme()
	ds := t.Style(theme.DimTextStyleID)
	hs := t.Style(theme.HighlightTextStyleID)
	ns := lipgloss.NewStyle()

	width = size.Width

	s = w.style

	left.WriteString(hs.Render(w.data.Name))
	left.WriteString("\n")
	left.WriteString(ds.Render(w.data.Description))

	right.WriteString(ds.Render(w.data.Skill.Name))
	right.WriteString("\n")
	right.WriteString(helper.HoursDecFormat(w.data.Duration.Duration))
	right.WriteString("\n")

	// Amount control
	right.WriteString(w.amountSelector.Render())

	right.WriteString("\n")
	right.WriteString(ds.Render(
		fmt.Sprintf("%dx %s",
			w.data.Amount, w.data.Item.Name)))

	width1, width2 := orvyn.DivideSizeFull(width)

	top = t.Style(ftheme.DimUnderlinedTextStyleID).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width1).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(width2).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	perPage := w.paginator.PerPage
	count := 0
	start, end := w.paginator.GetSliceBounds(len(w.data.Ingredients))

	left.Reset()
	right.Reset()

	for i, ingredient := range w.data.Ingredients[start:end] {
		if i%2 == 0 {
			b = &left
		} else {
			b = &right
		}

		if i > 1 {
			b.WriteString("\n")
		}

		b.WriteString(fmt.Sprintf("%dx %s",
			ingredient.Amount,
			ingredient.Item.Name))

		count++
	}

	ingredientsList.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		ns.Width(width1).
			AlignHorizontal(lipgloss.Left).
			Render(left.String()),
		ns.Width(width2).
			AlignHorizontal(lipgloss.Left).
			Render(right.String())))

	if count < 2 {
		ingredientsList.WriteString(strings.Repeat("\n", 2-count))
	}

	if len(w.data.Ingredients) > perPage {
		ingredientsList.WriteString("\n")
		ingredientsList.WriteString(hs.Render("< "))
		ingredientsList.WriteString(w.paginator.View())
		ingredientsList.WriteString(hs.Render(" >"))
	} else {
		ingredientsList.WriteString("\n")
	}

	tui := s.Width(width).Height(size.Height).Render(
		lipgloss.JoinVertical(lipgloss.Top,
			top, ingredientsList.String()))

	return tui
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}

func (w *Widget) FilterValue() string {
	var b strings.Builder

	b.WriteString(w.data.Name)
	b.WriteString(" ")
	b.WriteString(w.data.Description)
	b.WriteString(" ")
	b.WriteString(w.data.Skill.Name)
	b.WriteString(" ")

	return b.String()
}
