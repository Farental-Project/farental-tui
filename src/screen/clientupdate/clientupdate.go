package clientupdate

import (
	"farental/internal/config"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/internal/updater"
	"farental/screen"
	"farental/widget/help"
	"farental/widget/simplelogviewer"
	"fmt"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/progressbar"
)

type state int

const (
	statePrompt state = iota
	stateDownloading
	stateApplying
	stateRestarting
	stateFailed
	stateManualRequired
)

// bytesPerMB is used only for the human-readable size line.
const bytesPerMB = 1024 * 1024

type finishedMsg struct {
	err error
}

// Screen drives the whole update: prompt, download, swap, restart.
type Screen struct {
	title    *orvyn.SimpleRenderable
	subtitle *orvyn.SimpleRenderable
	status   *orvyn.SimpleRenderable

	notes *simplelogviewer.Widget

	bar *progressbar.Widget

	help *help.Widget

	layout *layout.CenterLayout

	state state

	result updater.Result

	progress atomic.Int64

	tickTag uint
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.subtitle = orvyn.NewSimpleRenderable("")
	s.subtitle.Style = t.Style(theme.DimTextStyleID)

	s.status = orvyn.NewSimpleRenderable("")
	s.status.Style = t.Style(theme.NormalTextStyleID)

	s.notes = simplelogviewer.New(lokyn.L("release notes"))
	s.notes.Style = simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}
	s.notes.SetAutoScroll(false)

	// simplelogviewer resolves its border and title styles in OnFocus/OnBlur;
	// without one of them both stay zero-valued and the widget renders bare.
	// This is the only interactive widget on the screen, so it reads focused.
	s.notes.OnFocus()

	s.bar = progressbar.New("")
	s.bar.SetTitleProgressVisibility(false)
	s.bar.SetPercentageVisibility(true)
	s.bar.SetActive(false)

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(
			35,
			t.Size(ftheme.LayoutWidthSizeID),
			10,
			s.title,
			s.subtitle,
			orvyn.VGap,
			s.notes,
			orvyn.VGap,
			s.bar,
			s.status,
			orvyn.VGap,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(_ any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextClientUpdate)

	s.result = updater.Pending
	s.state = statePrompt

	s.title.SetValue(lokyn.L("A new version is available"))
	s.subtitle.SetValue(fmt.Sprintf("%s  →  %s", s.result.Current, s.result.Latest))

	s.refreshNotes()

	s.bar.SetActive(false)
	s.status.SetValue("")

	// Ordered most specific cause first: a failed fetch also leaves File
	// empty, so checking HasFile first would blame the platform for what is
	// really a network problem.
	if s.result.Err != nil {
		s.enterManual(lokyn.L("Could not reach the update server."))
		return nil
	}

	if !s.result.HasFile() {
		s.enterManual(lokyn.L("No build is published for your platform."))
		return nil
	}

	// Failing the preflight before downloading spares the user a 20 MB wait
	// they cannot use.
	if err := updater.PreflightWritable(); err != nil {
		s.enterManual(fmt.Sprintf("%s\n%v",
			lokyn.L("Farental cannot write to its own directory."), err))
		return nil
	}

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		if cmd, handled := s.handleKey(m); handled {
			return cmd
		}
	}

	switch msg := msg.(type) {
	case progress.FrameMsg:
		return s.bar.Update(msg)

	case orvyn.TickMsg:
		if msg.Tag != s.tickTag {
			return nil
		}

		s.tickTag++

		return tea.Batch(s.refreshProgress(), orvyn.TickCmd(1, s.tickTag))

	case finishedMsg:
		return s.handleFinished(msg)
	}

	return s.notes.Update(msg)
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) handleKey(m tea.KeyMsg) (tea.Cmd, bool) {
	switch s.state {
	case statePrompt:
		switch {
		case key.Matches(m, keybind.Enter):
			return s.startUpdate(), true
		case key.Matches(m, keybind.Esc):
			if s.result.Mode == updater.ModeOptional {
				return orvyn.SwitchScreen(screen.IDLogin), true
			}
		}

	case stateFailed:
		switch {
		case key.Matches(m, keybind.RKey):
			return s.startUpdate(), true
		case key.Matches(m, keybind.Esc):
			if s.result.Mode == updater.ModeOptional {
				return orvyn.SwitchScreen(screen.IDLogin), true
			}
		}
	}

	return nil, false
}

func (s *Screen) startUpdate() tea.Cmd {
	s.state = stateDownloading
	s.progress.Store(0)

	s.bar.SetActive(true)
	s.notes.SetActive(false)
	s.status.SetValue(lokyn.L("Downloading..."))

	file := s.result.File

	download := func() tea.Msg {
		return finishedMsg{err: updater.Apply(config.WebURL, file, &s.progress)}
	}

	s.tickTag++

	return tea.Batch(download, orvyn.TickCmd(0, s.tickTag))
}

func (s *Screen) refreshProgress() tea.Cmd {
	if s.state != stateDownloading {
		return nil
	}

	done := s.progress.Load()
	total := s.result.File.SizeBytes

	if total <= 0 {
		return nil
	}

	percent := float64(done) / float64(total)

	s.status.SetValue(fmt.Sprintf("%.1f / %.1f MB",
		float64(done)/bytesPerMB, float64(total)/bytesPerMB))

	// The whole body is read before the swap, so a full bar means the
	// download finished and verification is under way.
	if done >= total {
		s.state = stateApplying
		s.status.SetValue(lokyn.L("Verifying and installing..."))
	}

	return s.bar.SetPercent(percent)
}

func (s *Screen) handleFinished(msg finishedMsg) tea.Cmd {
	if msg.err != nil {
		s.state = stateFailed
		s.bar.SetActive(false)
		// The retry key is not in the help keymap, which is fixed per context,
		// so the hint goes in the message the user is already reading.
		s.status.SetValue(fmt.Sprintf("%s\n%v\n\n%s",
			lokyn.L("The update failed."), msg.err, lokyn.L("Press r to retry.")))
		s.title.SetValue(lokyn.L("Update failed"))

		return nil
	}

	s.state = stateRestarting
	s.status.SetValue(lokyn.L("Restarting..."))

	// The exec happens in main, after bubbletea has left the alt screen and
	// restored the cursor; doing it here would hand the new process a
	// terminal still in raw mode.
	updater.RestartPending = true

	return tea.Quit
}

func (s *Screen) enterManual(reason string) {
	s.state = stateManualRequired
	s.title.SetValue(lokyn.L("Update required"))
	s.bar.SetActive(false)
	s.notes.SetActive(false)

	s.status.SetValue(fmt.Sprintf("%s\n\n%s\n%s/clienttui",
		reason, lokyn.L("Download the new version here:"), config.WebURL))
}

func (s *Screen) refreshNotes() {
	lines := updater.RenderNotes(s.result.Notes, orvyn.WindowSize.Width-10)

	if len(lines) == 0 {
		s.notes.SetActive(false)
		return
	}

	s.notes.SetActive(true)
	s.notes.SetContent(lines)
}
