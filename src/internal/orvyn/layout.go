package orvyn

type Layout interface {
	Renderable

	GetElements() []Renderable
}

type BaseLayout struct {
	elements []Renderable
	visible  bool
}

func NewBaseLayout(elements []Renderable) BaseLayout {
	b := BaseLayout{}

	b.elements = elements

	return b
}

func (b *BaseLayout) Resize(_ Size) {}

func (b *BaseLayout) GetSize() Size {
	return NewSize(0, 0)
}

func (b *BaseLayout) GetElements() []Renderable {
	return b.elements
}

func (b *BaseLayout) SetVisible(visible bool) {
	b.visible = visible
}

func (b *BaseLayout) IsVisible() bool {
	return b.visible
}
