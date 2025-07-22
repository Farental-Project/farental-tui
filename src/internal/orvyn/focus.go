package orvyn

import (
	"github.com/charmbracelet/bubbles/key"
)

type Focusable interface {
	// Updatable to be able to update a focusable widget.
	Updatable

	// OnFocus is called when the widget gains the focus.
	OnFocus()

	// OnBlur is called when the widget is loosing focus.
	OnBlur()

	// OnEnterInput is called when the widget enters the input mode.
	// Input mode means that all the tea.Msg will be managed by the widget.
	OnEnterInput()

	// OnExitInput is called when the widget exits the input mode.
	OnExitInput()

	// IsFocused return true if the widget is currently focused.
	// A widget can be focused without being in input mode.
	IsFocused() bool

	// IsInputting return true if the widget is in input mode.
	// Input mode means that all the tea.Msg will be managed by the widget.
	IsInputting() bool

	// GetFocusKeybind can return a *key.Binding if the widget should be
	// able to get directly focused with one key.
	GetFocusKeybind() *key.Binding

	// GetInputKeybind can return a *key.Binding if the widget should be
	// able to enter the input mode.
	// Input mode means that all the tea.Msg will be managed by the widget.
	GetInputKeybind() *key.Binding

	// setFocused allows to set the focused value.
	setFocused(bool)

	// setInputting allows to set the inputting value.
	setInputting(bool)
}

// BaseFocusable can be integrated to a Widget to make it focusable.
type BaseFocusable struct {
	focused   bool
	inputting bool
}

func (b *BaseFocusable) IsFocused() bool {
	return b.focused
}

func (b *BaseFocusable) IsInputting() bool {
	return b.inputting
}

func (b *BaseFocusable) GetFocusKeybind() *key.Binding {
	return nil
}

func (b *BaseFocusable) GetInputKeybind() *key.Binding {
	return nil
}

func (b *BaseFocusable) setFocused(focused bool) {
	b.focused = focused
}

func (b *BaseFocusable) setInputting(input bool) {
	b.inputting = input
}
