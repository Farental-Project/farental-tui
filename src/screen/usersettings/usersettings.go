package usersettings

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/help"
	"farental/widget/multivalueselector"
	"net/http"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/checkbox"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/spf13/viper"
)

type LanguageData struct {
	api.LanguageResponse
}

func (l LanguageData) RenderValue() string {
	return l.Name
}

type Screen struct {
	title *orvyn.SimpleRenderable

	labelLangage     *label.Widget
	mvsLanguage      *multivalueselector.Widget[LanguageData]
	chkbxNewsletters *checkbox.Widget

	labelTheme    *label.Widget
	mvsTheme      *multivalueselector.Widget[ftheme.ThemeData]
	statusMessage *statusmessage.Widget
	help          *help.Widget

	layout *layout.CenterLayout

	focusManager *orvyn.FocusManager
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(lokyn.L("User settings"))
	s.title.Style = t.Style(theme.TitleStyleID)

	s.labelLangage = label.New(lokyn.L("Language"))
	s.mvsLanguage = multivalueselector.New[LanguageData]()
	s.mvsLanguage.OnBlur()
	s.mvsLanguage.Looping = true

	s.labelTheme = label.New(lokyn.L("Theme (need restart)"))
	s.mvsTheme = multivalueselector.New[ftheme.ThemeData]()
	s.mvsTheme.OnBlur()
	s.mvsTheme.Looping = true

	s.mvsTheme.SetValues(ftheme.GetThemeData())

	s.chkbxNewsletters = checkbox.New(lokyn.L("Receive newsletters ?"))

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxLayout(10,
			s.title,
			orvyn.VGap,
			s.labelLangage,
			s.mvsLanguage,
			orvyn.VGap,
			s.labelTheme,
			s.mvsTheme,
			orvyn.VGap,
			s.chkbxNewsletters,
			orvyn.VGap,
			s.statusMessage,
			s.help,
		),
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.mvsLanguage)
	s.focusManager.Add(s.mvsTheme)
	s.focusManager.Add(s.chkbxNewsletters)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterCreation)

	s.statusMessage.Reset()

	s.focusManager.FocusFirst()

	s.loadData()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()
		case key.Matches(m, keybind.Enter):
			ok := s.submit()

			if ok {
				return orvyn.SwitchToPreviousScreen()
			}

			return nil
		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadData() {
	resp, err := helper.SendRequest(request.LangGetAll())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	languages, ok := resp.Result().(*[]api.LanguageResponse)

	if !ok {
		return
	}

	keys := make([]string, 0)
	data := make(map[string]LanguageData, 0)

	for _, l := range *languages {
		keys = append(keys, l.Code)
		data[l.Code] = LanguageData{
			LanguageResponse: l,
		}
	}

	s.mvsLanguage.SetValues(keys, data)

	resp, err = helper.SendRequest(request.AuthInfo())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	info, ok := resp.Result().(*api.UserResponse)

	if !ok {
		return
	}

	s.mvsLanguage.SetSelected(slices.Index(keys, info.LanguageCode))
	s.chkbxNewsletters.SetChecked(info.WantsNewsletter)

	currentTheme := viper.GetString("theme")
	if currentTheme == "" {
		currentTheme = "dark"
	}

	s.mvsTheme.SetSelectedKey(currentTheme)
}

func (s *Screen) submit() bool {
	body := api.UserSettingsBody{
		LanguageCode:    s.mvsLanguage.GetSelectedValue().Code,
		WantsNewsletter: s.chkbxNewsletters.IsChecked(),
	}

	resp, err := helper.SendRequest(request.AuthSetSettings(body))

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	if resp.StatusCode() == http.StatusOK {
		viper.Set("theme", s.mvsTheme.GetSelectedValue().Code)
		config.ChangeLanguage(body.LanguageCode)
		return true
	}

	return false
}
