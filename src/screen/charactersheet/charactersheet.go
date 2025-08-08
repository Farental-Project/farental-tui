package charactersheet

import (
	"farental/internal/orvyn"
	"farental/widget/characterinfo"
)

type Screen struct {
	orvyn.BaseScreen

	characterInfo *characterinfo.Widget
}

func New() *Screen {
	s := new(Screen)

	return s
}
