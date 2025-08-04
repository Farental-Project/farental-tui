package style

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

func SetTextInputStyle(ti *textinput.Model) {
	ti.TextStyle = NormalStyle
	ti.Cursor.Style = HighlightStyle
	ti.Cursor.TextStyle = NormalStyle
	ti.Prompt = ""
}

func SetTextAreaStyle(ta *textarea.Model) {
	ta.BlurredStyle.Text = DimTextStyle
	ta.BlurredStyle.Base = BlurredStyle
	ta.FocusedStyle.Text = NormalStyle
	ta.FocusedStyle.Base = FocusedStyle
	ta.FocusedStyle.CursorLine = NormalStyle
	ta.Cursor.TextStyle = NormalStyle
	ta.Cursor.Style = HighlightStyle
	ta.Prompt = ""
}
