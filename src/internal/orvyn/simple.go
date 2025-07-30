package orvyn

type SimpleRenderable struct {
	BaseRenderable

	view string
}

var VGap = NewSimpleRenderable("\n")

func NewSimpleRenderable(view string) *SimpleRenderable {
	s := new(SimpleRenderable)

	s.BaseRenderable = NewBaseRenderable()
	s.view = view

	return s
}

func (s *SimpleRenderable) SetView(view string) {
	s.view = view
}

func (s *SimpleRenderable) Render() string {
	return s.view
}
