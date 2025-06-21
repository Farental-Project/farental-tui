package filterselectionlist

import (
	"farental/internal/context"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	List                 list.Model
	Items                []list.Item
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

		case key.Matches(msg, keybind.Enter):
			if m.List.FilterState() != list.Filtering {
				ok := m.submit(&m)

				if ok {
					return context.ContentManager.
						SwitchContent(m, model.ContentGameDashboard)
				}

				return m, nil
			}

		case key.Matches(msg, keybind.Help):
			context.KeymapManager.ShowAll = !context.KeymapManager.ShowAll
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
	switch m.List.FilterState() {
	case list.Filtering:
		context.KeymapManager.UpdateKeybindHelpDesc(keybind.Esc, lang.L("cancel"))
		context.KeymapManager.UpdateKeybindHelpDesc(keybind.Enter, lang.L("apply"))
		context.KeymapManager.SetKeybindVisible(keybind.Filter, false)
		context.KeymapManager.SetKeybindVisible(keybind.Help, false)
	case list.FilterApplied:
		context.KeymapManager.UpdateKeybindHelpDesc(keybind.Esc, lang.L("clear filter"))
		context.KeymapManager.UpdateKeybindHelpDesc(keybind.Enter, "")
		context.KeymapManager.SetKeybindVisible(keybind.Filter, true)
		context.KeymapManager.SetKeybindVisible(keybind.Help, true)
	case list.Unfiltered:
		context.KeymapManager.UpdateKeybindHelpDesc(keybind.Esc, "")
		context.KeymapManager.UpdateKeybindHelpDesc(keybind.Enter, "")
		context.KeymapManager.SetKeybindVisible(keybind.Filter, true)
		context.KeymapManager.SetKeybindVisible(keybind.Help, true)
	}
}
