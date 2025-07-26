package layout

import (
	"farental/internal/orvyn"
	"strings"
)

type VBoxLayout struct {
	orvyn.BaseLayout
}

func NewVBoxLayout(elements []orvyn.Renderable) *VBoxLayout {
	l := new(VBoxLayout)

	l.BaseLayout = orvyn.NewBaseLayout(elements)

	return l
}

func (l *VBoxLayout) Render() string {
	var b strings.Builder
	var s orvyn.Size
	var minSize orvyn.Size
	var prefSize orvyn.Size
	var margin int

	size := l.GetSize()

	margin = 10

	s = orvyn.NewSize(size.Width-margin, 0)

	minSize = l.GetMinSize()
	prefSize = l.GetPreferredSize()

	if s.Width <= minSize.Width {
		s.Width = minSize.Width - margin
	} else if s.Width >= prefSize.Width {
		s.Width = prefSize.Width - margin
	}

	for i, e := range l.GetElements() {
		if i > 0 {
			b.WriteString("\n")
		}

		s.Height = e.GetMinSize().Height

		e.Resize(s)
		b.WriteString(e.Render())
	}

	return b.String()
}

func (l *VBoxLayout) GetMinSize() orvyn.Size {
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

func (l *VBoxLayout) GetPreferredSize() orvyn.Size {
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

func (l *VBoxLayout) GetMaxSize() orvyn.Size {
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
