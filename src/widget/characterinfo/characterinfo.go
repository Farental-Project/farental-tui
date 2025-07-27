package characterinfo

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/style"
	"farental/widget/characterbar"
	"farental/widget/label"
	"fmt"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget

	info  *label.Widget
	barHp *characterbar.Widget
	barMp *characterbar.Widget

	layout *layout.GrowHBoxLayout
}

func New() *Widget {
	w := new(Widget)

	w.info = label.New("")
	w.barHp = characterbar.New(style.ColorHpBar)
	w.barMp = characterbar.New(style.ColorMpBar)

	w.layout = layout.NewGrowHBoxLayout(2,
		[]orvyn.Renderable{
			w.barHp,
			w.info,
			w.barMp,
		})

	return w
}

func (w *Widget) Render() string {
	return w.layout.Render()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.layout.Resize(size)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(30, 4)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(45, 4)
}

func (w *Widget) GetMaxSize() orvyn.Size {
	return orvyn.NewSize(100, 5)
}

func (w *Widget) UpdateData(info *api.CharacterInfoResponse, money int) {
	w.constructInfo(info, money)

	for _, stat := range info.Stats {
		if stat.Code == "hp" {
			w.barHp.MaxValue = stat.MaxValue
			w.barHp.CurrentValue = stat.Value
			continue
		}

		if stat.Code == "mp" {
			w.barMp.MaxValue = stat.MaxValue
			w.barMp.CurrentValue = stat.Value
			continue
		}
	}
}

func (w *Widget) constructInfo(info *api.CharacterInfoResponse, money int) {
	var b strings.Builder

	fullName := style.BoldTextStyle.Render(
		fmt.Sprintf("%s %s", info.FirstName, info.LastName))
	raceName := info.RaceName
	raceStyle := style.RaceStyle(raceName)
	power := info.Power

	b.WriteString(fullName)
	b.WriteString("\n")
	b.WriteString(raceStyle.Render(raceName))
	b.WriteString("\n")
	b.WriteString(style.NormalStyle.Render(
		fmt.Sprintf("%d %c", money, art.CharGrynars),
	))
	b.WriteString("\n")
	b.WriteString(style.SpecialHighlightStyle.Render(
		fmt.Sprintf("%s : %d", lang.L("Power"), power),
	))

	w.info.SetValue(b.String())
}
