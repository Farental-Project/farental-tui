package charactercreation

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/widget/multivalueselector"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"strings"
)

type DataRaceValue struct {
	data api.RaceResponse
}

func (d DataRaceValue) RenderValue() string {
	return d.data.Name
}

type Model struct {
	FirstnameInput textinput.Model
	LastnameInput  textinput.Model
	RaceInput      multivalueselector.Model[DataRaceValue]

	Keymap config.ModularKeyMap

	Title string

	ErrMsg error

	tabIndex              int
	currentlySelectedRace DataRaceValue
}

func New() Model {
	m := Model{}

	m.FirstnameInput = textinput.New()
	m.FirstnameInput.Placeholder = lang.L("Firstname")
	m.FirstnameInput.Focus()
	m.FirstnameInput.Width = 31
	m.FirstnameInput.Prompt = ""
	style.SetTextInputStyle(&m.FirstnameInput)

	m.LastnameInput = textinput.New()
	m.LastnameInput.Placeholder = lang.L("Lastname")
	m.LastnameInput.Width = 31
	m.LastnameInput.Prompt = ""
	style.SetTextInputStyle(&m.LastnameInput)

	m.RaceInput = multivalueselector.New[DataRaceValue]()
	m.RaceInput.Width = 32
	m.RaceInput.Style.BlurredControl = style.DimTextStyle
	m.RaceInput.Style.BlurredValue = style.DimTextStyle
	m.RaceInput.Style.FocusedControl = style.TitleStyle
	m.RaceInput.Style.FocusedValue = style.HighlightStyle

	m.Title = lang.L("Character creation")

	m.tabIndex = 0

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var mod tea.Model
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case model.InitMsg:
		m.FirstnameInput.SetValue("")
		m.LastnameInput.SetValue("")

		m.loadRaces()

		m.RaceInput.SetSelectedIndex(0)
		m.currentlySelectedRace = m.RaceInput.GetSelectedValue()

		bubblehelp.SwitchContext(model.ContextCharacterCreation)

		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		case key.Matches(msg, keybind.Esc):
			return context.ContentManager.SwitchContent(m, model.ContentCharacterSelection)
		case key.Matches(msg, keybind.Enter):
			ok := m.submit()

			if ok {
				return context.ContentManager.SwitchContent(m, model.ContentCharacterSelection)
			}
		case key.Matches(msg, keybind.Tab, keybind.ShiftTab):
			if key.Matches(msg, keybind.Tab) {
				m.tabIndex++
			} else if key.Matches(msg, keybind.ShiftTab) {
				m.tabIndex--
			}

			if m.tabIndex > 2 {
				m.tabIndex = 0
			} else if m.tabIndex < 0 {
				m.tabIndex = 2
			}

			var cmd tea.Cmd

			m.updateFocus()

			return m, cmd
		}
	}

	m.FirstnameInput, cmd = m.FirstnameInput.Update(msg)
	cmds = append(cmds, cmd)

	m.LastnameInput, cmd = m.LastnameInput.Update(msg)
	cmds = append(cmds, cmd)

	mod, cmd = m.RaceInput.Update(msg)
	cmds = append(cmds, cmd)

	m.RaceInput = mod.(multivalueselector.Model[DataRaceValue])

	m.currentlySelectedRace = m.RaceInput.GetSelectedValue()

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var form strings.Builder
	var tui strings.Builder

	form.WriteString(m.renderFocus(
		m.FirstnameInput.Focused(),
		m.FirstnameInput.View()))
	form.WriteString("\n")
	form.WriteString(m.renderFocus(
		m.LastnameInput.Focused(),
		m.LastnameInput.View()))
	form.WriteString("\n")
	form.WriteString(m.renderFocus(
		m.RaceInput.Focused(),
		m.RaceInput.View()))
	form.WriteString("\n")
	form.WriteString(style.DimTextStyle.Width(style.LayoutWidth).
		AlignHorizontal(lipgloss.Center).
		Render(m.currentlySelectedRace.data.Description))

	tui.WriteString(style.TitleStyle.Render(m.Title))
	tui.WriteString("\n\n\n")
	tui.WriteString(form.String())

	if m.ErrMsg != nil {
		tui.WriteString("\n\n")
		tui.WriteString(style.ErrorStyle.Render(m.ErrMsg.Error()))
	}

	tui.WriteString("\n\n\n")
	tui.WriteString(bubblehelp.View(style.LayoutWidth))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth, context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui.String())
}

func (m *Model) updateFocus() {
	switch m.tabIndex {
	case 0:
		m.FirstnameInput.Focus()
		m.LastnameInput.Blur()
		m.RaceInput.Blur()
	case 1:
		m.FirstnameInput.Blur()
		m.LastnameInput.Focus()
		m.RaceInput.Blur()
	case 2:
		m.FirstnameInput.Blur()
		m.LastnameInput.Blur()
		m.RaceInput.Focus()
	}
}

func (m *Model) renderFocus(focus bool, view string) string {
	var s lipgloss.Style

	if focus {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	return s.Render(view)
}

func (m *Model) loadRaces() {
	resp, err := helper.SendRequest(request.DataGetAllRace())

	if err != nil {
		m.ErrMsg = err
		return
	}

	races := *resp.Result().(*[]api.RaceResponse)

	raceValues := make(map[string]DataRaceValue, len(races))
	keys := make([]string, len(races))

	for k, v := range races {
		keys[k] = v.Name
		raceValues[keys[k]] = DataRaceValue{
			data: v,
		}
	}

	m.RaceInput.SetValues(keys, raceValues)

	m.currentlySelectedRace = m.RaceInput.GetSelectedValue()
}

func (m *Model) submit() bool {
	req := request.CharacterCreate()

	req.SetBody(
		api.CharacterCreateBody{
			FirstName: m.FirstnameInput.Value(),
			LastName:  m.LastnameInput.Value(),
			RaceID:    m.currentlySelectedRace.data.ID,
		})

	_, err := helper.SendRequest(req)

	if err != nil {
		m.ErrMsg = err
		return false
	}

	return true
}
