// Package widgetfocusmanager contains everything needed to automate focus management.
package widgetfocusmanager

import (
	"farental/internal/keybind"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// WidgetFocusManager helps managing focusable widgets.
type WidgetFocusManager struct {
	widgets  []FocusableWidget
	tabIndex int
}

// New creates and return a new WidgetFocusManager
func New() *WidgetFocusManager {
	w := &WidgetFocusManager{}

	w.widgets = make([]FocusableWidget, 0)
	w.tabIndex = 0

	return w
}

// Add adds up the given widget to the widgets list. Order of addition defines tab order.
func (w *WidgetFocusManager) Add(widget FocusableWidget) {
	w.widgets = append(w.widgets, widget)
}

// Remove removes the widget at the given index.
func (w *WidgetFocusManager) Remove(index int) {
	if index < 0 || index >= len(w.widgets) {
		return
	}

	w.widgets = append(w.widgets[:index], w.widgets[index+1:]...)
}

// Focus set the focus on the given index.
func (w *WidgetFocusManager) Focus(index int) {
	if index != w.tabIndex {
		w.widgets[w.tabIndex].Blur()
	}

	w.tabIndex = index

	w.widgets[w.tabIndex].Focus()
}

// BlurCurrent blurs the currently focused widget.
func (w *WidgetFocusManager) BlurCurrent() {
	w.widgets[w.tabIndex].Blur()
}

// Update must be called from the screen Model, all key events should be managed by widgets.
// It returns a tea.Cmd, that should be used by the screen Model.
// The widget currently in EditMode have total control of the update messages.
// If the currently focused widget is not in EditMode, the standard keybinds are managed.
// The update of the currently focused widget is executed, for example to manage the back action.
// The switch to EditMode is managed here.
func (w *WidgetFocusManager) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if w.widgets[w.tabIndex].IsInEditMode() {
		_, cmd = w.widgets[w.tabIndex].Update(msg)
		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Tab):
			w.widgets[w.tabIndex].Blur()

			w.tabIndex++
			if w.tabIndex >= len(w.widgets) {
				w.tabIndex = 0
			}

			w.widgets[w.tabIndex].Focus()

			return nil

		case key.Matches(msg, keybind.ShiftTab):
			w.widgets[w.tabIndex].Blur()

			w.tabIndex--

			if w.tabIndex < 0 {
				w.tabIndex = len(w.widgets) - 1
			}

			w.widgets[w.tabIndex].Focus()

			return nil

		}

		editModeKeybind := w.widgets[w.tabIndex].GetEditModeKeybind()

		if editModeKeybind != nil {
			if key.Matches(msg, *editModeKeybind) {
				w.widgets[w.tabIndex].EnterEditMode()

				return nil
			}
		}

		// Specific focus keybind
		for i, widget := range w.widgets {
			keybind := widget.GetFocusKeybind()

			if keybind == nil {
				continue
			}

			if key.Matches(msg, *keybind) {
				w.widgets[w.tabIndex].Blur()

				w.tabIndex = i

				w.widgets[w.tabIndex].Focus()

				return nil
			}
		}
	}

	_, cmd = w.widgets[w.tabIndex].Update(msg)

	return cmd
}
