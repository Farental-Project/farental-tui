package orvyn

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// FocusManager can be instantiated when needed on a Screen to manage focus
// on multiple widgets. Manage focus and input mode.
type FocusManager struct {
	// NextFocusKeybind holds the key.Binding to loop through the Focusable Widgets.
	// Tab by default.
	NextFocusKeybind key.Binding

	// PreviousFocusKeybind holds the key.Binding to loop through the Focusable Widgets.
	// Shift+Tab by default.
	PreviousFocusKeybind key.Binding

	widgets  []Focusable
	tabIndex int
}

// NewFocusManager creates and return a new *FocusManager
func NewFocusManager() *FocusManager {
	f := new(FocusManager)

	f.widgets = make([]Focusable, 0)
	f.tabIndex = 0

	f.NextFocusKeybind = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next focus"),
	)
	f.PreviousFocusKeybind = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous focus"),
	)

	return f
}

// Add append the given Focusable Widget to the manager.
// Order of append defines the focus order.
func (f *FocusManager) Add(widget Focusable) {
	f.widgets = append(f.widgets, widget)
}

// SetWidgets replaces the manager widget list with the one given.
// Widget order defines the focus order.
func (f *FocusManager) SetWidgets(widgets []Focusable) {
	f.widgets = widgets
}

// Remove the widget from the manager at the given index.
func (f *FocusManager) Remove(index int) {
	if index < 0 || index >= len(f.widgets) {
		return
	}

	f.widgets = append(f.widgets[:index], f.widgets[index+1:]...)
}

// Focus set the focus on the Focusable Widget at the given index.
func (f *FocusManager) Focus(index int) {
	if index < 0 || index >= len(f.widgets) {
		return
	}

	if index != f.tabIndex {
		f.BlurCurrent()
	}

	f.tabIndex = index

	f.focus(f.tabIndex)
}

// BlurCurrent simply blur the currently focused widget
func (f *FocusManager) BlurCurrent() {
	f.blur(f.tabIndex)
}

func (f *FocusManager) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	cmd = nil

	if len(f.widgets) == 0 {
		return nil
	}

	if f.widgets[f.tabIndex].IsInputting() {
		cmd = f.widgets[f.tabIndex].Update(msg)
		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.NextFocusKeybind):
			if f.widgets[f.tabIndex].IsFocused() {
				f.blur(f.tabIndex)
			}
		
			f.tabIndex++
			if f.tabIndex >= len(f.widgets) {
				f.tabIndex = 0
			}

			f.focus(f.tabIndex)

			return nil

		case key.Matches(msg, f.PreviousFocusKeybind):
			if f.widgets[f.tabIndex].IsFocused() {
				f.blur(f.tabIndex)
			}

			f.tabIndex--
			if f.tabIndex < 0 {
				f.tabIndex = len(f.widgets) - 1
			}

			f.focus(f.tabIndex)

			return nil
		}

		inputtingKeybind := f.widgets[f.tabIndex].GetInputKeybind()

		if inputtingKeybind != nil {
			if key.Matches(msg, *inputtingKeybind) {
				f.enterInput(f.tabIndex)

				return nil
			}
		}

		// Checking for specific focus keybind
		for i, widget := range f.widgets {
			keybind := widget.GetFocusKeybind()

			if keybind == nil {
				continue
			}

			if key.Matches(msg, *keybind) {
				if f.widgets[f.tabIndex].IsFocused() {
					f.blur(f.tabIndex)
				}

				f.tabIndex = i

				f.focus(f.tabIndex)

				return nil
			}
		}
	}

	// Call the update on the currently focused widget
	if f.widgets[f.tabIndex].IsFocused() {
		cmd = f.widgets[f.tabIndex].Update(msg)
	}

	return cmd
}

// Hidden functions

// focus is a shorthand to manage the focused state and call OnFocus.
func (f *FocusManager) focus(index int) {
	f.widgets[index].setFocused(true)
	f.widgets[index].OnFocus()
}

// blur is a shorthand to manage the focused state and call OnBlur.
func (f *FocusManager) blur(index int) {
	f.widgets[index].setFocused(false)
	f.widgets[index].OnBlur()
}

// enterInput is a shorthand to manage the inputting state and call OnEnterInput.
func (f *FocusManager) enterInput(index int) {
	f.widgets[index].setInputting(true)
	f.widgets[index].OnEnterInput()
}

// exitInput is a shorthand to manage the inputting state and call OnExitInput.
func (f *FocusManager) exitInput(index int) {
	f.widgets[index].setInputting(false)
	f.widgets[index].OnExitInput()
}
