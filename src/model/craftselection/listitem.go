package craftselection

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strconv"
	"strings"
)

type ListItem struct {
	CraftRecipe api.RecipeResponse
	Amount      int
	Paginator   paginator.Model
}

func NewListItem(recipe *api.RecipeResponse) ListItem {
	li := ListItem{}

	li.CraftRecipe = *recipe
	li.Amount = 1

	li.Paginator = paginator.New()
	li.Paginator.Type = paginator.Dots
	li.Paginator.PerPage = 4
	li.Paginator.ActiveDot = style.TitleStyle.Render("•")
	li.Paginator.InactiveDot = style.DimTextStyle.Render("•")
	li.Paginator.SetTotalPages(len(li.CraftRecipe.Ingredients))
	li.Paginator.KeyMap.NextPage = keybind.NextPage
	li.Paginator.KeyMap.PrevPage = keybind.PrevPage

	return li
}

func (i ListItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.CraftRecipe.Name)
	b.WriteString(i.CraftRecipe.Description)
	b.WriteString(i.CraftRecipe.Skill.Name)

	return b.String()
}

type ListItemDelegate struct{}

func (l ListItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(ListItem)

	if !ok {
		return
	}

	var s lipgloss.Style
	var left strings.Builder
	var right strings.Builder
	var b *strings.Builder
	var ingredientsList strings.Builder
	var top string
	var width int

	width = m.Width() - 2

	if index == m.Index() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	left.WriteString(i.CraftRecipe.Name)
	left.WriteString("\n")
	left.WriteString(style.DimTextStyle.Render(i.CraftRecipe.Description))

	right.WriteString(style.DimTextStyle.Render(i.CraftRecipe.Skill.Name))
	right.WriteString("\n")
	right.WriteString(helper.HoursDecFormat(i.CraftRecipe.Duration.Duration))
	right.WriteString("\n")

	// Amount control
	right.WriteString(style.HighlightStyle.Render("< "))
	right.WriteString(style.BoldTextStyle.Render(strconv.Itoa(i.Amount)))
	right.WriteString(style.HighlightStyle.Render(" >"))

	right.WriteString("\n")
	right.WriteString(style.DimTextStyle.Render(
		fmt.Sprintf("%dx %s",
			i.CraftRecipe.Amount, i.CraftRecipe.Item.Name)))

	top = style.DimBottomBorderStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			style.TextStyle.Width(width/2).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			style.TextStyle.Width(width/2).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	perPage := i.Paginator.PerPage
	count := 0
	start, end := i.Paginator.GetSliceBounds(len(i.CraftRecipe.Ingredients))

	left.Reset()
	right.Reset()

	for i, ingredient := range i.CraftRecipe.Ingredients[start:end] {
		if i%2 == 0 {
			b = &left
		} else {
			b = &right
		}

		if i > 1 {
			b.WriteString("\n")
		}

		b.WriteString(
			fmt.Sprintf("%dx %s",
				ingredient.Amount,
				ingredient.Item.Name))

		count++
	}

	ingredientsList.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		style.TextStyle.Width(width/2).
			AlignHorizontal(lipgloss.Left).
			Render(left.String()),
		style.TextStyle.Width(width/2).
			AlignHorizontal(lipgloss.Left).
			Render(right.String())))

	if count < 2 {
		ingredientsList.WriteString(strings.Repeat("\n", 2-count))
	}

	if len(i.CraftRecipe.Ingredients) > perPage {
		ingredientsList.WriteString("\n")
		ingredientsList.WriteString(style.HighlightStyle.Render("< "))
		ingredientsList.WriteString(i.Paginator.View())
		ingredientsList.WriteString(style.HighlightStyle.Render(" >"))
	} else {
		ingredientsList.WriteString("\n")
	}

	tui := s.Width(m.Width() - 2).Height(l.Height()).Render(
		lipgloss.JoinVertical(lipgloss.Top,
			top, ingredientsList.String()))

	fmt.Fprint(w, tui)
}

func (l ListItemDelegate) Height() int {
	return 5
}

func (l ListItemDelegate) Spacing() int {
	return 0
}

func (l ListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	selectedIndex := m.GlobalIndex()
	selectedItem, ok := m.SelectedItem().(ListItem)

	if !ok {
		return nil
	}

	switch msgType := msg.(type) {
	case tea.KeyMsg:

		switch {
		case key.Matches(msgType, keybind.Left):
			selectedItem.Amount--

			if selectedItem.Amount < 1 {
				selectedItem.Amount = 1
			}

			return updateItem(m, selectedIndex, selectedItem)
		case key.Matches(msgType, keybind.Right):
			selectedItem.Amount++

			if selectedItem.Amount > 100 {
				selectedItem.Amount = 100
			}

			return updateItem(m, selectedIndex, selectedItem)
		}
	}

	selectedItem.Paginator, _ = selectedItem.Paginator.Update(msg)

	return updateItem(m, selectedIndex, selectedItem)
}

func updateItem(m *list.Model, index int, item ListItem) tea.Cmd {
	return m.SetItem(index, item)
}
