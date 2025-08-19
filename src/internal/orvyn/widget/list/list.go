package widget

import (
	"farental/internal/orvyn"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ItemConstructor defines the signature of the item constructor.
// T type represents the type of the item data.
type ItemConstructor[T any] func(T) orvyn.Focusable

// Widget defines a list widget.
// T type represents the type of the item data.
type Widget[T any] struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	listItems []orvyn.Focusable
	items     []T

	focusManager *orvyn.FocusManager

	itemConstructor ItemConstructor[T]
}

// New creates a new *Widget list and takes an itemConstructor as parameter.
// T type represents the type of the item data.
func New[T any](itemConstructor ItemConstructor[T]) *Widget[T] {
	w := new(Widget[T])

	w.itemConstructor = itemConstructor

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.PreviousFocusKeybind = key.NewBinding(key.WithKeys("up"))
	w.focusManager.NextFocusKeybind = key.NewBinding(key.WithKeys("down"))

	return w
}

func (w *Widget[T]) Update(msg tea.Msg) tea.Cmd {
	cmd := w.focusManager.Update(msg)

	return cmd
}

func (w *Widget[T]) Render() string {
	return ""
}

func (w *Widget[T]) OnFocus() {}

func (w *Widget[T]) OnBlur() {}

func (w *Widget[T]) OnEnterInput() {}

func (w *Widget[T]) OnExitInput() {}

// SetItems takes a []T (slice of data) and instantiate all items
// based on it.
func (w *Widget[T]) SetItems(items []T) {
	w.items = items

	w.listItems = make([]orvyn.Focusable, 0)
	w.focusManager.SetWidgets([]orvyn.Focusable{})

	for _, i := range w.items {
		w.listItems = append(w.listItems,
			w.itemConstructor(i))
	}

	// TODO: Test - Order should be good ?
	w.focusManager.SetWidgets(w.listItems)
}
