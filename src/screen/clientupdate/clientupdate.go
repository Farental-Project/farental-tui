package clientupdate

import (
	"errors"
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
	stateUnrecoverable
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

	s.title.SetValue(lokyn.L("A new version is available"))

	// Latest is empty when the manifest fetch failed; showing "1.1.0  →  "
	// with a blank right side is worse than showing nothing.
	if s.result.Latest != "" {
		s.subtitle.SetValue(fmt.Sprintf("%s  →  %s", s.result.Current, s.result.Latest))
	} else {
		s.subtitle.SetValue("")
	}

	s.refreshNotes()

	s.bar.SetActive(false)
	s.status.SetValue("")

	// PreflightWritable is a cheap local check (create+remove one temp file),
	// so it always runs; decideEntry only consults it once the fetch and
	// platform checks already passed.
	preflightErr := updater.PreflightWritable()

	entryState, reason := decideEntry(s.result, preflightErr)
	s.state = entryState

	switch reason {
	case reasonFetchFailed:
		s.enterManual(lokyn.L("Could not reach the update server."))
	case reasonNoFile:
		s.enterManual(lokyn.L("No build is published for your platform."))
	case reasonNotWritable:
		s.enterManual(fmt.Sprintf("%s\n%v",
			lokyn.L("Farental cannot write to its own directory."), preflightErr))
	}

	return nil
}

// entryReason identifies why OnEnter chose stateManualRequired, so OnEnter can
// pick the right localized message without decideEntry depending on lokyn.
// That keeps decideEntry a pure function of its inputs, trivial to unit test.
type entryReason int

const (
	reasonNone entryReason = iota
	reasonFetchFailed
	reasonNoFile
	reasonNotWritable
)

// decideEntry is the pure decision made at screen entry: given the update
// check result and whatever the preflight check already found (so this itself
// needs no filesystem access), it picks the state to enter and, when that
// state is stateManualRequired, why.
//
// Ordered most specific cause first: a failed fetch also leaves File empty,
// so checking HasFile first would blame the platform for what is really a
// network problem.
func decideEntry(result updater.Result, preflightErr error) (state, entryReason) {
	if result.Err != nil {
		return stateManualRequired, reasonFetchFailed
	}

	if !result.HasFile() {
		return stateManualRequired, reasonNoFile
	}

	if preflightErr != nil {
		return stateManualRequired, reasonNotWritable
	}

	return statePrompt, reasonNone
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
	case tea.WindowSizeMsg:
		// orvyn.SwitchScreen runs OnEnter synchronously from App.Init(),
		// before bubbletea's loop starts and so before the first resize
		// message arrives; refreshNotes there wraps to a hardcoded
		// default. This case is what makes it wrap to the real terminal
		// width, both at startup and on every later resize.
		s.refreshNotes()

		return s.notes.Update(msg)

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

	case stateManualRequired:
		switch {
		case key.Matches(m, keybind.Esc):
			if s.result.Mode == updater.ModeOptional {
				return orvyn.SwitchScreen(screen.IDLogin), true
			}
		}

	case stateUnrecoverable:
		switch {
		case key.Matches(m, keybind.Esc):
			// No retry key here on purpose: the target binary is gone
			// and a retry cannot succeed.
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

	// A retry from stateFailed left the title reading "Update failed"; a
	// fresh attempt needs it back to the neutral prompt title.
	s.title.SetValue(lokyn.L("A new version is available"))

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
		s.bar.SetActive(false)
		s.title.SetValue(lokyn.L("Update failed"))

		// RollbackFailedError means the swap failed *and* selfupdate's own
		// rollback failed: there is no file at the target path at all.
		// Retrying can only fail again the same way, so the hint must not
		// be "press r" — it has to send the user to the saved old binary.
		var rerr *updater.RollbackFailedError

		if errors.As(msg.err, &rerr) {
			s.state = stateUnrecoverable
			s.status.SetValue(fmt.Sprintf("%s\n%v\n\n%s",
				lokyn.L("The update failed."), msg.err,
				fmt.Sprintf(lokyn.L("The previous version was saved at %s. Reinstall it manually."), rerr.OldPath)))

			return nil
		}

		s.state = stateFailed
		// The retry key is not in the help keymap, which is fixed per context,
		// so the hint goes in the message the user is already reading.
		s.status.SetValue(fmt.Sprintf("%s\n%v\n\n%s",
			lokyn.L("The update failed."), msg.err, lokyn.L("Press r to retry.")))

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

	message := fmt.Sprintf("%s\n\n%s\n%s/clienttui",
		reason, lokyn.L("Download the new version here:"), config.WebURL)

	// Only a compatible client can skip past this screen; a mandatory update
	// has no exit besides quitting, so telling it about esc would be a lie.
	if s.result.Mode == updater.ModeOptional {
		message = fmt.Sprintf("%s\n\n%s", message, lokyn.L("Press esc to continue without updating."))
	}

	s.status.SetValue(message)
}

func (s *Screen) refreshNotes() {
	// The pane never grows past the theme's layout width even on a very
	// wide terminal (DefinedWidthVerticalLayout clamps to it), so the wrap
	// width has to follow the same min, not the raw terminal width. -10 is
	// that layout's margin; -2 is the notes pane's own border.
	layoutWidth := orvyn.GetTheme().Size(ftheme.LayoutWidthSizeID)
	width := min(orvyn.WindowSize.Width, layoutWidth) - 10 - 2

	lines := updater.RenderNotes(s.result.Notes, width)

	if len(lines) == 0 {
		s.notes.SetActive(false)
		return
	}

	s.notes.SetActive(true)
	s.notes.SetContent(lines)
}
