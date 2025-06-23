package style

import (
	"github.com/charmbracelet/bubbles/textinput"
)

func SetTextInputStyle(ti *textinput.Model) {
	ti.TextStyle = NormalStyle
	ti.Cursor.Style = HighlightStyle
	ti.Cursor.TextStyle = NormalStyle
}
