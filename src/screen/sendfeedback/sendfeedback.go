package sendfeedback

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/dialog/popup"
	"farental/widget"
	"farental/widget/help"
	"farental/widget/multivalueselector"
	"strings"

	ftheme "farental/internal/theme"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textarea"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type FeedbackKindData api.FeedbackKind

func (f FeedbackKindData) RenderValue() string {
	return string(f)
}

type PlatformNameData api.PlatformName

func (f PlatformNameData) RenderValue() string {
	return string(f)
}

type Screen struct {
	title *orvyn.SimpleRenderable

	mvsKind     *multivalueselector.Widget[FeedbackKindData]
	mvsPlatform *multivalueselector.Widget[PlatformNameData]

	tiSubject *textinput.Widget
	taMessage *textarea.Widget

	help *help.Widget

	statusMessage *statusmessage.Widget

	layout *layout.CenterLayout

	focusManager *orvyn.FocusManager

	selectedPlatform api.PlatformName
	sendedFeedback   bool
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.mvsKind = multivalueselector.New[FeedbackKindData]()

	s.mvsPlatform = multivalueselector.New[PlatformNameData]()

	s.tiSubject = textinput.New()

	s.taMessage = textarea.New()
	s.taMessage.ShowLineNumbers = false
	s.taMessage.KeyMap.InsertNewline = keybind.YKeyCtrl
	s.taMessage.SetMinSize(orvyn.NewSize(10, 8))
	s.taMessage.SetPreferredSize(orvyn.NewSize(10, 15))

	s.help = help.New()

	s.statusMessage = statusmessage.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(
			35,
			t.Size(ftheme.LayoutWidthSizeID),
			10,
			s.title,
			orvyn.VGap,
			s.mvsKind,
			s.mvsPlatform,
			s.tiSubject,
			s.taMessage,
			orvyn.VGap,
			s.statusMessage,
			s.help,
		),
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.mvsKind)
	s.focusManager.Add(s.mvsPlatform)
	s.focusManager.Add(s.tiSubject)
	s.focusManager.Add(s.taMessage)

	s.selectedPlatform = api.PlatformUndefined

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextChat)

	s.title.SetValue(lokyn.L("Send feedback"))

	s.tiSubject.Placeholder = lokyn.L("Subject")
	s.taMessage.Placeholder = lokyn.L("Message")

	s.tiSubject.SetValue("")
	s.taMessage.SetValue("")

	s.loadFeedbackKind()
	s.loadPlatformName()

	s.mvsPlatform.SetActive(false)

	s.focusManager.FocusFirst()

	s.statusMessage.Reset()

	return nil
}

func (s *Screen) OnExit() any {
	if s.sendedFeedback {
		return widget.StatusMessageParam{
			Content: lokyn.L("Thanks for your feedback. It will be analysed by the team."),
			Type:    statusmessage.SuccessMessage,
		}
	}

	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
			s.sendedFeedback = false
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(msg, keybind.Enter):
			if !s.validate() {
				return nil
			}

			orvyn.OpenDialog("feedbackSendConfirm", popup.NewYesNo(
				lokyn.L("Are you sure you want to send this feedback?"),
			), nil)

			return nil
		}

	case orvyn.DialogExitMsg:
		if msg.DialogID == "feedbackSendConfirm" && msg.Param.(uint) == 1 {
			ret := s.submit()

			if ret {
				s.sendedFeedback = true
				return orvyn.SwitchToPreviousScreen()
			}
		}

		return nil
	}

	cmd := s.focusManager.Update(msg)

	s.activatePlatformSelection()

	if s.mvsPlatform.IsActive() {
		s.selectedPlatform = api.PlatformName(s.mvsPlatform.GetSelectedValue())
	}

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadFeedbackKind() {
	values := make(map[string]FeedbackKindData, 3)
	keys := make([]string, 3)

	keys[0] = "question"
	values[keys[0]] = FeedbackKindData(api.KindQuestion)
	keys[1] = "feedback"
	values[keys[1]] = FeedbackKindData(api.KindFeedback)
	keys[2] = "bugreport"
	values[keys[2]] = FeedbackKindData(api.KindBugReport)

	s.mvsKind.SetValues(keys, values)
	s.mvsKind.SetSelected(0)
}

func (s *Screen) loadPlatformName() {
	values := make(map[string]PlatformNameData, 3)
	keys := make([]string, 3)

	keys[0] = "linux"
	values[keys[0]] = PlatformNameData(api.PlatformLinux)
	keys[1] = "windows"
	values[keys[1]] = PlatformNameData(api.PlatformWindows)
	keys[2] = "macos"
	values[keys[2]] = PlatformNameData(api.PlatformMacOS)

	s.mvsPlatform.SetValues(keys, values)
	s.mvsPlatform.SetSelected(0)
	s.selectedPlatform = api.PlatformUndefined
}

func (s *Screen) activatePlatformSelection() {
	kind := s.mvsKind.GetSelectedValue()

	if kind == FeedbackKindData(api.KindBugReport) {
		if !s.mvsPlatform.IsActive() {
			s.mvsPlatform.SetActive(true)
			s.mvsPlatform.SetSelected(0)
		}
	} else {
		s.mvsPlatform.SetActive(false)
		s.selectedPlatform = api.PlatformUndefined
	}
}

// validate checks the form is complete before letting the user confirm sending it.
func (s *Screen) validate() bool {
	if strings.TrimSpace(s.tiSubject.Value()) == "" {
		s.statusMessage.SetMessage(lokyn.L("Subject cannot be empty."), statusmessage.ErrorMessage)
		return false
	}

	if strings.TrimSpace(s.taMessage.Value()) == "" {
		s.statusMessage.SetMessage(lokyn.L("Message cannot be empty."), statusmessage.ErrorMessage)
		return false
	}

	if s.mvsKind.GetSelectedValue() == FeedbackKindData(api.KindBugReport) && s.selectedPlatform == api.PlatformUndefined {
		s.statusMessage.SetMessage(lokyn.L("Please select a platform."), statusmessage.ErrorMessage)
		return false
	}

	return true
}

func (s *Screen) submit() bool {
	body := api.SendFeedbackBody{
		Subject:  s.tiSubject.Value(),
		Message:  s.taMessage.Value(),
		Kind:     api.FeedbackKind(s.mvsKind.GetSelectedValue()),
		Platform: s.selectedPlatform,
	}

	req := request.AuthSendFeedback(body)

	_, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	return true
}
