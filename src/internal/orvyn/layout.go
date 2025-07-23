package orvyn

type Layout interface {
	Renderable

	GetElements() []Renderable
}

type BaseLayout struct {
	elements []Renderable
}

func NewBaseLayout(elements []Renderable) *BaseLayout {
	b := new(BaseLayout)

	b.elements = elements

	return b
}

func (b *BaseLayout) GetElements() []Renderable {
	return b.elements
}
