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
)

type Model struct {
	List                 list.Model
	Items                []list.Item
	Help                 help.Model
	Keymap               config.ModularKeyMap
	showIncreaseDecrease bool
	showPageUpDown       bool

	Width int

	Title string

	ErrMsg error

	loadData func(m *Model) []list.Item
	submit   func(m *Model) bool
}

func New(title string, listItemDelegate list.ItemDelegate, loadData func(m *Model) []list.Item, submit func(m *Model) bool) Model {
	m := Model{}

	m.Width = style.LayoutWidth

	m.Help = help.New()

	m.Items = make([]list.Item, 0)

	m.List = list.New(m.Items,
		listItemDelegate,
		m.Width, 30)
	m.List.SetShowHelp(false)
	m.List.SetShowTitle(false)
	m.List.DisableQuitKeybindings()

	m.Title = title
	m.loadData = loadData
	m.submit = submit
	m.showIncreaseDecrease = false
	m.showPageUpDown = false

	m.Keymap = config.ModularKeyMap{}

	m.updateKeymap()

	return m
}

func (m *Model) SetShowExtraKeybinds(showIncreaseDecrease, showPageUpDown bool) {
	m.showIncreaseDecrease = showIncreaseDecrease
	m.showPageUpDown = showPageUpDown
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
		m.ErrMsg = nil
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit

		case key.Matches(msg, keybind.Submit):
			if m.List.FilterState() != list.Filtering {
				ok := m.submit(&m)

				if ok {
					return context.ContentManager.
						SwitchContent(m, model.ContentGameDashboard)
				}

				return m, nil
			}

		case key.Matches(msg, keybind.HelpMore):
			m.Help.ShowAll = !m.Help.ShowAll
		}
	}

	context.ContentManager.Update(msg)

	m.List, cmd = m.List.Update(msg)

	m.updateKeymap()

	return m, cmd
}

func (m Model) View() string {
	return style.ContainerStyle.Width(m.Width).Render(m.List.View())
}

func (m Model) ViewHelp() string {
	return m.Help.View(m.Keymap)
}

func (m Model) ViewTitle() string {
	return style.TitleStyle.Render(m.Title)
}

func (m Model) ViewError() string {
	var err string

	if m.ErrMsg != nil {
		err = m.ErrMsg.Error()
	}

	return err
}

func (m *Model) UpdateData() {
	m.Items = m.loadData(m)
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

	fullKeys = append(fullKeys, leftColumn, rightColumn)

	m.Keymap.SetBindings(fullKeys)
	m.Keymap.SetEssentialBindings(essentialsKeys)
}
