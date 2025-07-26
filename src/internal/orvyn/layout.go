package orvyn

type Layout interface {
	Renderable

	GetElements() []Renderable
}

type BaseLayout struct {
	BaseRenderable

	elements []Renderable
}

func NewBaseLayout(elements []Renderable) BaseLayout {
	b := BaseLayout{}

	b.elements = elements

	return b
}

func (b *BaseLayout) GetElements() []Renderable {
	return b.elements
}
