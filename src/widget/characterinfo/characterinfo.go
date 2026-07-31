package characterinfo

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/style"
	ftheme "farental/internal/theme"
	"farental/widget/characterbar"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type CharacterInfoData struct {
	FirstName  string
	LastName   string
	RaceName   string
	Gender     string
	Power      int
	Money      int
	UnreadMail bool

	Stats []api.CharacterStatResponse
}

type Widget struct {
	orvyn.BaseWidget

	info  *orvyn.SimpleRenderable
	barHp *characterbar.Widget
	barMp *characterbar.Widget

	layout *layout.HBoxGrowLayout

	ShowMoney bool
	ShowPower bool
	ShowMail  bool
}

func New() *Widget {
	w := new(Widget)

	t := orvyn.GetTheme()

	w.BaseWidget = orvyn.NewBaseWidget()

	w.ShowMoney = true
	w.ShowPower = true
	w.ShowMail = true

	w.info = orvyn.NewSimpleRenderable("")
	w.info.Style = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center)
	w.info.SizeConstraint = true
	w.barHp = characterbar.New(lokyn.L("HP"), t.Color(ftheme.HPBarColorID))
	w.barMp = characterbar.New(lokyn.L("MP"), t.Color(ftheme.MPBarColorID))

	w.layout = layout.NewHBoxGrowLayout(1, 1,
		w.barHp,
		w.info,
		w.barMp,
	)

	// Keep the bars pinned to the top so they do not shift down when the info
	// column grows a line.
	w.layout.Align = lipgloss.Top

	return w
}

func (w *Widget) Render() string {
	return w.GetStyle().Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
}

// getHeight measures an actual render rather than guessing. The height follows the
// info column, which grows with the money/power/mail rows and with how the name
// wraps, so a constant was always going to drift from what is drawn.
func (w *Widget) getHeight() int {
	return lipgloss.Height(w.Render())
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(30, w.getHeight())
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(45, w.getHeight())
}

func (w *Widget) UpdateData(info *CharacterInfoData) {
	w.constructInfo(info)

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

func (w *Widget) constructInfo(info *CharacterInfoData) {
	var b strings.Builder

	t := orvyn.GetTheme()

	fullName := t.Style(theme.TitleStyleID).Render(
		fmt.Sprintf("%s %s", info.FirstName, info.LastName))
	raceName := info.RaceName
	raceStyle := style.RaceStyle(raceName)
	power := info.Power
	unreadMail := ""

	if info.UnreadMail {
		unreadMail = t.Style(theme.TitleStyleID).Render(
			lokyn.L("You have unread mail"))
	} else {
		unreadMail = t.Style(theme.DimTextStyleID).Render(
			lokyn.L("You have no unread mail"))
	}

	b.WriteString(fullName)
	b.WriteString("\n")
	fmt.Fprintf(&b, "%s - %s", raceStyle.Render(raceName), info.Gender)
	b.WriteString("\n")

	if w.ShowMoney {
		b.WriteString(t.Style(theme.NormalTextStyleID).Render(
			fmt.Sprintf("%d %c", info.Money, art.CharGrynars),
		))
	}

	b.WriteString("\n")

	if w.ShowPower {
		b.WriteString(t.Style(theme.HighlightTextStyleID).Render(
			fmt.Sprintf("%s : %d", lokyn.L("Power"), power),
		))
	}

	if w.ShowMail {
		b.WriteString("\n")
		b.WriteString(unreadMail)
	}

	w.info.SetValue(b.String())
}

func ConvertCharacterInfoResponseToData(character *api.CharacterInfoResponse, money int, unreadMail bool) *CharacterInfoData {
	return &CharacterInfoData{
		FirstName:  character.FirstName,
		LastName:   character.LastName,
		RaceName:   character.RaceName,
		Gender:     character.Gender,
		Power:      character.Power,
		Money:      money,
		Stats:      character.Stats,
		UnreadMail: unreadMail,
	}
}

func ConvertCharacterInspectResponseToData(character *api.CharacterInspectResponse) *CharacterInfoData {
	return &CharacterInfoData{
		FirstName:  character.FirstName,
		LastName:   character.LastName,
		RaceName:   character.RaceName,
		Gender:     character.Gender,
		Power:      0,
		Money:      0,
		Stats:      character.Stats,
		UnreadMail: false,
	}
}
