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

	b.BaseRenderable = NewBaseRenderable()
	b.elements = elements

	return b
}

func (b *BaseLayout) GetElements() []Renderable {
	var visibleElements []Renderable

	for _, e := range b.elements {
		if !e.IsActive() {
			continue
		}

		visibleElements = append(visibleElements, e)
	}

	return visibleElements
}
