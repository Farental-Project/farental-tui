package widgetfocusmanager

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// FocusableWidget describes a widget that can be managed by the WidgetFocusManager.
type FocusableWidget interface {
	tea.Model
	Focus()
	Blur()
	GetFocusKeybind() *key.Binding
	GetEditModeKeybind() *key.Binding
	EnterEditMode()
	ExitEditMode()
	IsInEditMode() bool
}

// BaseFocusableWidget is a base struct to help build FocusableWidget.
type BaseFocusableWidget struct {
	Focused  bool
	EditMode bool
}

// Init from the tea.Model interface
func (b BaseFocusableWidget) Init() tea.Cmd {
	return nil
}

// Update from the tea.Model interface. But use a pointer on the model.
func (b *BaseFocusableWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

// View from the tea.Model interface.
func (b BaseFocusableWidget) View() string {
	return ""
}

// Focus called when the widget gain focus.
func (b *BaseFocusableWidget) Focus() {
	b.Focused = true
}

// Blur called when the widget looses focus.
func (b *BaseFocusableWidget) Blur() {
	b.Focused = false
}

// GetFocusKeybind returns the key.Binding that should allow for instant focus of the widget.
func (b *BaseFocusableWidget) GetFocusKeybind() *key.Binding {
	return nil
}

// GetEditModeKeybind returns the key.Binding that should allow to enter in EditMode.
func (b *BaseFocusableWidget) GetEditModeKeybind() *key.Binding {
	return nil
}

// IsInEditMode returns the EditMode value
func (b *BaseFocusableWidget) IsInEditMode() bool {
	return b.EditMode
}

// EnterEditMode is called to enter in edit mode. The widget should call the base code.
func (b *BaseFocusableWidget) EnterEditMode() {
	b.EditMode = true
}

// ExitEditMode is called to exit the edit mode. The widget should call the base code.
func (b *BaseFocusableWidget) ExitEditMode() {
	b.EditMode = false
}
