package vbox

import (
	"farental/internal/orvyn"
	"strings"
)

type Layout struct {
	orvyn.BaseLayout
}

func New(elements []orvyn.Renderable) *Layout {
	l := new(Layout)

	l.BaseLayout = *orvyn.NewBaseLayout(elements)

	return l
}

func (l *Layout) ResizeLayout(size *orvyn.Size) {
	var b strings.Builder

	for i, e := range l.GetElements() {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString(e.Render(orvyn.NewSize(l.GetMinSize().Width, e.GetMinSize().Height)))
	}
}

func (l *Layout) GetMinSize() orvyn.Size {
	var size orvyn.Size

	for _, e := range l.GetElements() {
		eSize := e.GetMinSize()
		size.Height += eSize.Height

		if eSize.Width > size.Width {
			size.Width = eSize.Width
		}
	}

	return size
}

func (l *Layout) GetPreferredSize() orvyn.Size {
	var size orvyn.Size

	for _, e := range l.GetElements() {
		eSize := e.GetPreferredSize()
		size.Height += eSize.Height

		if eSize.Width > size.Width {
			size.Width = eSize.Width
		}
	}

	return size
}

func (l *Layout) GetMaxSize() orvyn.Size {
	var size orvyn.Size

	for _, e := range l.GetElements() {
		eSize := e.GetPreferredSize()
		size.Height += eSize.Height

		if eSize.Width > size.Width {
			size.Width = eSize.Width
		}
	}

	return size
}
