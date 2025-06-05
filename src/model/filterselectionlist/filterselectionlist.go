package filterselectionlist

import (
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/keybind"
	"farental/model"
	"farental/style"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	List                 list.Model
	Items                []list.Item
	Help                 help.Model
	Keymap               config.ModularKeyMap
	ShowIncreaseDecrease bool
	ShowPageUpDown       bool

	Title string

	loadData func() []list.Item
	submit   func(m *Model) bool
}

func New(title string, listItemDelegate list.ItemDelegate, loadData func() []list.Item, submit func(m *Model) bool) Model {
	m := Model{}

	m.Help = help.New()

	m.Items = make([]list.Item, 0)

	m.List = list.New(m.Items,
		listItemDelegate,
		style.LayoutWidth, 30)
	m.List.SetShowHelp(false)
	m.List.SetShowTitle(false)
	m.List.DisableQuitKeybindings()

	m.Title = title
	m.loadData = loadData
	m.submit = submit
	m.ShowIncreaseDecrease = false
	m.ShowPageUpDown = false

	m.Keymap = config.ModularKeyMap{}

	m.updateKeymap()

	return m
}

func (m *Model) SetShowExtraKeybinds(showIncreaseDecrease, showPageUpDown bool) {
	m.ShowIncreaseDecrease = showIncreaseDecrease
	m.ShowPageUpDown = showPageUpDown
	m.updateKeymap()
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case model.InitMsg:
		m.UpdateData()

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit

		case key.Matches(msg, keybind.Submit):
			if m.List.FilterState() != list.Filtering {
				ok := m.submit(&m)

				if ok {
					return context.ContentManager.SwitchContent(
						model.ContentGameDashboard)
				}

				return m, nil
			}

		case key.Matches(msg, keybind.HelpMore):
			m.Help.ShowAll = !m.Help.ShowAll

		case key.Matches(msg, keybind.Back):
			if m.List.FilterState() == list.Unfiltered {
				return context.ContentManager.SwitchContent(
					model.ContentGameDashboard)
			}
		}
	}

	context.ContentManager.Update(msg)

	m.List, cmd = m.List.Update(msg)

	m.updateKeymap()

	return m, cmd
}

func (m Model) View() string {
	var tui strings.Builder

	helpText := m.Help.View(m.Keymap)

	title := style.TitleStyle.Render(m.Title)

	tui.WriteString(title)
	tui.WriteString("\n\n")
	tui.WriteString(style.ContainerStyle.Width(style.LayoutWidth).Render(m.List.View()))
	tui.WriteString("\n\n")
	tui.WriteString(helpText)

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui.String())
}

func (m *Model) UpdateData() {
	m.Items = m.loadData()
	m.List.SetItems(m.Items)
}

func (m *Model) updateKeymap() {
	var fullKeys [][]key.Binding
	var leftColumn []key.Binding
	var rightColumn []key.Binding
	var essentialsKeys []key.Binding
	var escKey key.Binding
	var enterKey key.Binding

	keybind.HelpMore.SetEnabled(true)

	switch m.List.FilterState() {
	case list.Filtering:
		escKey = keybind.Cancel
		enterKey = keybind.Apply
	case list.FilterApplied:
		escKey = keybind.ClearFilter
		enterKey = keybind.Submit
	case list.Unfiltered:
		escKey = keybind.Back
		enterKey = keybind.Submit
	}

	if m.List.FilterState() == list.Filtering {
		essentialsKeys = append(essentialsKeys, enterKey, escKey)
		keybind.HelpMore.SetEnabled(false)
		m.Keymap.SetEssentialBindings(essentialsKeys)
		return
	}

	essentialsKeys = append(essentialsKeys,
		keybind.Up, keybind.Down, keybind.Filter, escKey, enterKey, keybind.HelpMore)

	leftColumn = append(leftColumn,
		keybind.Up,
		keybind.Down,
		keybind.Decrease,
		keybind.Increase,
		keybind.GotoListStart,
		keybind.GotoListEnd,
	)

	rightColumn = append(rightColumn,
		keybind.PrevPage,
		keybind.NextPage,
		keybind.Filter,
		keybind.Submit,
		keybind.Back,
		keybind.Quit,
		keybind.HelpClose,
	)

	if !m.ShowIncreaseDecrease {
		keybind.Decrease.SetEnabled(false)
		keybind.Increase.SetEnabled(false)
	} else {
		keybind.Decrease.SetEnabled(true)
		keybind.Increase.SetEnabled(true)
	}

	if !m.ShowPageUpDown {
		keybind.PrevPage.SetEnabled(false)
		keybind.NextPage.SetEnabled(false)
	} else {
		keybind.PrevPage.SetEnabled(true)
		keybind.NextPage.SetEnabled(true)
	}

	fullKeys = append(fullKeys, leftColumn, rightColumn)

	m.Keymap.SetBindings(fullKeys)
	m.Keymap.SetEssentialBindings(essentialsKeys)
}
