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
	"github.com/halsten-dev/bubblehelp"
)

type Model struct {
	List  list.Model
	Items []list.Item

	Width int

	Title string

	ErrMsg error

	CustomEnterDesc string

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
	m.CustomEnterDesc = ""
	m.loadData = loadData
	m.submit = submit

	m.updateKeymap()

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case model.InitMsg:
		m.UpdateData()
		m.updateKeymap()

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
					return m, model.SwitchContentCmd("")
				}

				return m, nil
			}

		case key.Matches(msg, keybind.Help):
			bubblehelp.ShowAll = !bubblehelp.ShowAll
		}
	}

	context.ContentManager.Update(msg)

	m.List, cmd = m.List.Update(msg)

	m.updateKeymap()

	return m, cmd
}

func (m Model) View() string {
	return style.FocusedStyle.Width(m.Width).Render(m.List.View())
}

func (m Model) ViewTitle() string {
	return style.TitleStyle.Render(m.Title)
}

func (m Model) ViewError() string {
	var err string

	err = ""

	if m.ErrMsg != nil {
		err = m.ErrMsg.Error()
	}

	return style.ErrorStyle.Render(err)
}

func (m *Model) UpdateData() {
	m.Items = m.loadData(m)
	m.List.SetItems(m.Items)
}

func (m *Model) updateKeymap() {
	switch m.List.FilterState() {
	case list.Filtering:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, lang.L("cancel"))
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, lang.L("apply"))
		bubblehelp.SetKeybindVisible(keybind.Filter, false)
		bubblehelp.SetKeybindVisible(keybind.Help, false)
	case list.FilterApplied:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, lang.L("clear filter"))
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, m.CustomEnterDesc)
		bubblehelp.SetKeybindVisible(keybind.Filter, true)
		bubblehelp.SetKeybindVisible(keybind.Help, true)
	case list.Unfiltered:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, "")
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, m.CustomEnterDesc)
		bubblehelp.SetKeybindVisible(keybind.Filter, true)
		bubblehelp.SetKeybindVisible(keybind.Help, true)
	}
}
