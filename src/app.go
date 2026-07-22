package main

import (
	"farental/internal/context"
	"farental/internal/session"
	"farental/screen"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
)

// App is the main model to run the Orvyn application
type App struct{}

func (a App) Init() tea.Cmd {
	return orvyn.SwitchScreen(screen.IDLogin)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := orvyn.Update(msg)

	if session.Expired() {
		if orvyn.GetCurrentScreenID() == screen.IDLogin {
			// Already on the login screen, which has shown the error
			// itself; the flag only needs to be cleared.
			session.TakeExpired()
		} else {
			context.Logout()
			closeAnyDialog()

			// The login screen consumes the flag in its OnEnter to
			// display the session expired message.
			cmd = orvyn.SwitchScreen(screen.IDLogin)
		}
	}

	return a, cmd
}

func (a App) View() string {
	return orvyn.Render()
}

// closeAnyDialog dismisses an open orvyn dialog so it does not keep
// overlaying the login screen after a forced logout. Orvyn has no way to
// query whether a dialog is open and CloseDialog panics when none is, so the
// panic is absorbed here. The DialogExitMsg command is dropped on purpose.
func closeAnyDialog() {
	defer func() {
		_ = recover()
	}()

	orvyn.CloseDialog()
}
