package chat

import (
	"errors"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/widget/simplelogviewer"
	"farental/model/widgetcontainer"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

const (
	tick time.Duration = 15
)

type Model struct {
	ErrMsg       error
	Log          simplelogviewer.Model
	LogContainer widgetcontainer.Model
	Input        textarea.Model
	Help         help.Model
	Keymap       config.ModularKeyMap

	tickTag uint
}

func New() Model {
	m := Model{}

	m.ErrMsg = nil

	m.Log = simplelogviewer.New(style.LayoutWidth, 40)

	m.LogContainer = widgetcontainer.New(
		m.Log, lang.L("Chat"), style.LayoutWidth, 42)

	m.Input = textarea.New()
	m.Input.CharLimit = 350
	m.Input.SetWidth(style.LayoutWidth)
	m.Input.SetHeight(3)
	m.Input.Placeholder = lang.L("Enter a message...")
	m.Input.Focus()
	m.Input.Prompt = ""
	m.Input.ShowLineNumbers = false
	m.Input.FocusedStyle.Base = style.ContainerStyle

	m.Input.KeyMap.InsertNewline = keybind.NewLine

	m.Help = help.New()

	m.Keymap = config.ModularKeyMap{}

	m.Keymap.SetEssentialBindings([]key.Binding{
		keybind.SendMessage,
		keybind.NewLine,
		keybind.ScrollUp,
		keybind.ScrollDown,
		keybind.Back,
		keybind.Quit,
	})

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(model.InitCmd, textarea.Blink,
		model.TickCmd(m, tick, m.tickTag))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.Input, cmd = m.Input.Update(msg)

	defer context.ContentManager.UpdateCurrentContent(m)

	switch msg := msg.(type) {
	case model.InitMsg:
		m.loadChat()
		m.Input.SetValue("")
		m.LogContainer.Title = fmt.Sprintf("%s - %s",
			lang.L("Chat"), context.CharacterInfo.Location.Name)
	case model.TickMsg:
		if msg.Tag != m.tickTag {
			return m, nil
		}

		m.loadChat()

		m.tickTag++
		return m, model.TickCmd(m, tick, m.tickTag)
	case tea.KeyMsg:
		m.ErrMsg = nil

		switch {
		case key.Matches(msg, keybind.SendMessage):
			m.sendMessage()
		case key.Matches(msg, keybind.ScrollUp):
			m.Log.Viewport.ScrollUp(1)
			m.LogContainer.UpdateContent(m.Log)
		case key.Matches(msg, keybind.ScrollDown):
			m.Log.Viewport.ScrollDown(1)
			m.LogContainer.UpdateContent(m.Log)
		case key.Matches(msg, keybind.Back):
			return context.ContentManager.
				SwitchContent(m, model.ContentGameDashboard)
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		}
	}

	context.ContentManager.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	var errorMessage string

	errorMessage = ""

	if m.ErrMsg != nil {
		errorMessage = fmt.Sprintf("\n%s\n",
			style.ErrorStyle.Render(m.ErrMsg.Error()))
	}

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			m.LogContainer.View(),
			m.Input.View(),
			errorMessage,
			m.Help.View(m.Keymap)))
}

func (m *Model) sendMessage() {
	var message string

	message = m.Input.Value()

	if len(message) == 0 {
		m.ErrMsg = errors.New(lang.L("Can't send empty messages"))
		return
	}

	message = helper.RemoveEmptyLines(message, 5)

	req := request.ChatSendMessage()

	req.Body = api.ChatMessageBody{
		Message: message,
	}

	resp, err := req.Send()

	if err != nil {
		m.ErrMsg = helper.ConnectionError()
		return
	}

	m.ErrMsg = helper.ExtractError(resp)

	if m.ErrMsg != nil {
		return
	}

	m.Input.SetValue("")
	m.loadChat()
}

func (m *Model) loadChat() {
	context.UpdateChat()

	if len(context.ChatContent) > len(m.Log.Content) {
		m.Log.SetContent(context.ChatContent)
		m.LogContainer.UpdateContent(m.Log)
	}
}
