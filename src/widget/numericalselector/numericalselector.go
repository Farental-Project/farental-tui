package numericalselector

import (
	"farental/internal/helper"
	"farental/internal/keybind"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"strconv"
)

type Widget struct {
	orvyn.BaseWidget

	value int

	minValue int
	maxValue int
	step     int
}

func New(min, max, step int) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.value = min

	w.minValue = min
	w.maxValue = max

	w.step = step

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Left):
			w.value -= w.step

			if w.value < w.minValue {
				w.value = w.minValue
			}

		case key.Matches(msg, keybind.ShiftLeft):
			w.value -= helper.Prev10Inc(w.value)

			if w.value < w.minValue {
				w.value = w.minValue
			}

		case key.Matches(msg, keybind.Right):
			w.value += w.step

			if w.value > w.maxValue {
				w.value = w.maxValue
			}

		case key.Matches(msg, keybind.ShiftRight):
			w.value += helper.Next10Inc(w.value)

			if w.value > w.maxValue {
				w.value = w.maxValue
			}

		}
	}

	return nil
}

func (w *Widget) Resize(size orvyn.Size) {

}

func (w *Widget) Render() string {
	t := orvyn.GetTheme()
	hst := t.Style(theme.HighlightTextStyleID)

	return fmt.Sprintf("%s %s %s",
		hst.Render("<"),
		t.Style(theme.NormalTextStyleID).Bold(true).Render(strconv.Itoa(w.value)),
		hst.Render(">"))
}

func (w *Widget) GetValue() int {
	return w.value
}
