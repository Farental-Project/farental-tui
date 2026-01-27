package skillgrouplistitem

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Data[T any] struct {
	Items     []T
	SkillName string
}

type Widget[T any] struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data Data[T]
}

func Constructor[T any](data Data[T]) widgetlist.ListItem[Data[T]] {
	w := new(Widget[T])

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget[T]) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (w *Widget[T]) UpdateData(data Data[T]) {
	w.data = data
}

func (w *Widget[T]) GetData() Data[T] {
	return w.data
}

func (w *Widget[T]) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)
}

func (w *Widget[T]) Render() string {
	size := w.GetContentSize()
	t := orvyn.GetTheme()
	s := w.GetStyle()

	return s.Width(size.Width).Height(size.Height).
		Render(t.Style(theme.TitleStyleID).Render(w.data.SkillName))
}

func (w *Widget[T]) FilterValue() string {
	return w.data.SkillName
}
