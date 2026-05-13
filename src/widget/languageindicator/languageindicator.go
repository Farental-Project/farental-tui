package languageindicator

import (
	"farental/internal/config"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/spf13/viper"
)

var languages [3]string = [3]string{"en", "fr", "de"}

type Widget struct {
	orvyn.BaseWidget

	currentLangIndex int
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.SetStyle(lipgloss.NewStyle())

	w.setCurrentLangIndex()

	return w
}

func (w *Widget) Render() string {
	return w.GetStyle().Render(w.renderLanguages())
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(14, 1)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(20, 1)
}

func (w *Widget) renderLanguages() string {
	var b strings.Builder
	var s *lipgloss.Style

	t := orvyn.GetTheme()
	ts := t.Style(theme.TitleStyleID)
	ds := t.Style(theme.DimTextStyleID)

	for i, l := range languages {
		if i > 0 {
			b.WriteString(" ")
		}

		if i == w.currentLangIndex {
			s = &ts
		} else {
			s = &ds
		}

		elem := fmt.Sprintf("[%s]", strings.ToUpper(l))

		b.WriteString(s.Render(elem))
	}

	return b.String()
}

func (w *Widget) SwitchLanguage() {
	w.currentLangIndex++

	if w.currentLangIndex == len(languages) {
		w.currentLangIndex = 0
	}

	config.ChangeLanguage(languages[w.currentLangIndex])
}

func (w *Widget) setCurrentLangIndex() {
	lang := strings.ToLower(viper.GetString("language"))

	for i, l := range languages {
		if l == lang {
			w.currentLangIndex = i
		}
	}
}
