package theme

import (
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type ThemeData struct {
	Code string
	Name string
}

func (t ThemeData) RenderValue() string {
	return t.Name
}

func GetThemeData() ([]string, map[string]ThemeData) {
	keys := make([]string, 0)
	data := make(map[string]ThemeData)

	keys = append(keys, "dark")
	data["dark"] = ThemeData{
		Code: "dark",
		Name: "Dark",
	}

	keys = append(keys, "light")
	data["light"] = ThemeData{
		Code: "light",
		Name: "Light",
	}

	return keys, data
}

func GetTheme(code string) theme.Theme {
	switch code {
	case "dark":
		return NewFarentalDarkTheme()
	case "light":
		return NewFarentalLightTheme()
	default:
		return NewFarentalDarkTheme()
	}
}

const (
	TitleUnderlinedTextStyleID theme.StyleID = iota + 9999
	DimUnderlinedTextStyleID
)

const (
	HPBarColorID theme.ColorID = iota + 9999
	MPBarColorID

	RacePrataarColorID
	RaceGarnothColorID
	RaceEltrysColorID
	RaceKaarColorID
	RaceNymanColorID
	RaceKrynamColorID

	BiomeMountainColorID
	BiomeFieldColorID
	BiomeHillColorID
	BiomeDesertColorID
	BiomeTropicalColorID
	BiomeForestColorID
	BiomeSwampColorID
	BiomeUndergroundColorID
	BiomeBeachColorID
)

const (
	LayoutWidthSizeID theme.SizeID = iota + 9999
)

func SetUnfocusableTextinputStyle(ti *textinput.Widget) {
	t := orvyn.GetTheme()

	ti.TextStyle = t.Style(theme.DimTextStyleID)
	ti.Cursor.TextStyle = t.Style(theme.DimTextStyleID)
}
