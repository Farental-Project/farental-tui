package layout

import (
	"farental/internal/orvyn"
	"github.com/charmbracelet/lipgloss"
	"math"
)

type GrowHBoxLayout struct {
	orvyn.BaseLayout

	gap int

	compensatorIndex int
}

func NewGrowHBoxLayout(gap, compensatorIndex int, elements []orvyn.Renderable) *GrowHBoxLayout {
	l := new(GrowHBoxLayout)

	l.BaseLayout = orvyn.NewBaseLayout(elements)
	l.gap = gap
	l.compensatorIndex = compensatorIndex

	return l
}

func (l *GrowHBoxLayout) Render() string {
	var view []string
	var s orvyn.Size

	size := l.GetSize()

	minSize := l.GetMinSize()
	prefSize := l.GetPreferredSize()

	s.Height = size.Height

	if size.Height <= minSize.Height {
		s.Height = minSize.Height
	} else if size.Height >= prefSize.Height {
		s.Height = prefSize.Height
	}

	availableWidth := size.Width - (l.gap*len(l.GetElements()) - 1)
	s.Width = int(math.Floor(float64(availableWidth / len(l.GetElements()))))

	// calculate the compensation
	compensatorSize := orvyn.NewSize(s.Width, s.Height)
	totalWidth := 0

	for i := range l.GetElements() {
		if i > 0 {
			totalWidth += l.gap
		}

		totalWidth += s.Width
	}

	compensation := size.Width - totalWidth

	compensatorSize.Width += compensation

	view = make([]string, 0)

	for i, e := range l.GetElements() {
		if i > 0 {
			view = append(view, " ")
		}

		if i == l.compensatorIndex {
			e.Resize(compensatorSize)
		} else {
			e.Resize(s)
		}
		view = append(view, e.Render())
	}

	return lipgloss.JoinHorizontal(lipgloss.Center,
		view...)
}

func (l *GrowHBoxLayout) GetMinSize() orvyn.Size {
	var size orvyn.Size

	for _, e := range l.GetElements() {
		eSize := e.GetMinSize()

		if eSize.Width == 0 && eSize.Height == 0 {
			eSize = e.GetSize()
		}

		size.Width = max(eSize.Width, size.Width)
		size.Height = max(eSize.Height, size.Height)
	}

	size.Width *= len(l.GetElements())

	return size
}

func (l *GrowHBoxLayout) GetPreferredSize() orvyn.Size {
	var size orvyn.Size

	for _, e := range l.GetElements() {
		eSize := e.GetPreferredSize()

		if eSize.Width == 0 && eSize.Height == 0 {
			eSize = e.GetSize()
		}

		size.Width = max(eSize.Width, size.Width)
		size.Height = max(eSize.Height, size.Height)
	}

	size.Width *= len(l.GetElements())

	return size
}

func (l *GrowHBoxLayout) GetMaxSize() orvyn.Size {
	var size orvyn.Size

	for _, e := range l.GetElements() {
		eSize := e.GetMaxSize()

		if eSize.Width == 0 && eSize.Height == 0 {
			eSize = e.GetSize()
		}

		size.Width = max(eSize.Width, size.Width)
		size.Height = max(eSize.Height, size.Height)
	}

	size.Width *= len(l.GetElements())

	return size
}
