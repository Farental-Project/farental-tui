package layout

import (
	"farental/internal/orvyn"
	"github.com/charmbracelet/lipgloss"
)

type CenterLayout struct {
	orvyn.BaseLayout
}

func NewCenterLayout(element orvyn.Renderable) *CenterLayout {
	l := new(CenterLayout)

	l.BaseLayout = orvyn.NewBaseLayout([]orvyn.Renderable{element})

	return l
}

func (l *CenterLayout) Render() string {
	size := l.GetSize()

	l.GetElements()[0].Resize(size)

	return lipgloss.Place(
		size.Width, size.Height,
		lipgloss.Center, lipgloss.Center,
		l.GetElements()[0].Render(),
	)
}

func (l *CenterLayout) GetMinSize() orvyn.Size {
	return l.GetElements()[0].GetMinSize()
}

func (l *CenterLayout) GetPreferredSize() orvyn.Size {
	return l.GetElements()[0].GetPreferredSize()
}

func (l *CenterLayout) GetMaxSize() orvyn.Size {
	return l.GetElements()[0].GetMaxSize()
}
