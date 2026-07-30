package clientupdate

import (
	"errors"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/internal/ticker"
	"farental/internal/updater"
	"farental/screen"
	"farental/widget/help"
	"farental/widget/simplelogviewer"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/progressbar"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/spf13/viper"

	"sync/atomic"
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

	// The three states below only exist in consultation mode (see
	// OpenConsultation): the startup path never has to represent "still
	// checking" or "already up to date" because main only routes to this
	// screen at all when updater.Pending.Mode != ModeNone.
	stateConsultChecking
	stateConsultUpToDate
	stateConsultFailed
)

// bytesPerMB is used only for the human-readable size line.
const bytesPerMB = 1024 * 1024

type finishedMsg struct {
	err error
}

// consultCheckedMsg is the result of the fresh, on-demand check consultation
// mode runs on entry. err is the version-compat request's own failure
// (distinct from a populated result whose Err field is set, which is
// updater.Check's internal manifest-fetch failure); either means there is no
// usable Result to show.
//
// gen ties the message back to the specific enterConsultation call that
// launched the check that produced it (see Screen.generation). checkForConsultation
// is an in-flight network command; if the screen is left and re-entered -
// on either path - before it returns, a stale message must not be able to
// rewrite whatever the new entry already set up.
type consultCheckedMsg struct {
	result updater.Result
	err    error
	gen    uint64
}

// consultPending tells the next OnEnter to run in consultation mode:
// fetch fresh data instead of consuming updater.Pending, and let esc leave
// unconditionally. It is the same hand-off shape as updater.Pending and
// session.Expired() - a package-level flag set by the caller and consumed by
// OnEnter - because orvyn.SwitchScreen only ever hands OnEnter the caller's
// OnExit() return value, and login and usersettings both return nil from
// OnExit unconditionally: neither can tell OnEnter "this transition is a
// consultation" any other way.
var consultPending bool

// consultFrom and consultRestorePrevious ride alongside consultPending,
// carrying the extra hand-off OpenConsultation needs: the screen esc must
// return to, and the previousScreenID orvyn had before OpenConsultation's own
// SwitchScreen overwrote it. See OpenConsultation and Screen.exitCmd.
var (
	consultFrom            orvyn.ScreenID
	consultRestorePrevious orvyn.ScreenID
)

// OpenConsultation opens the client-update screen read-only: it always
// fetches fresh version info rather than reusing updater.Pending, shows
// current/latest version and release notes, and lets the user start the
// update from here if a newer release exists. Unlike the startup path, esc
// always returns to whichever screen called this, even if the check comes
// back reporting a mandatory update - the user opened this voluntarily and
// must never be trapped by it.
//
// from is the ScreenID of the caller. orvyn keeps a single previousScreenID
// slot, not a stack: the SwitchScreen call below is about to overwrite it
// with from, so relying on orvyn.SwitchToPreviousScreen to leave consultation
// would send esc back into whichever screen last held that slot, not
// necessarily the caller - and if the caller is itself reached via
// SwitchToPreviousScreen (as user settings' own esc handler is), the two
// ping-pong forever. Capturing from explicitly, and capturing what
// previousScreenID was *before* this call so it can be restored on the way
// out, are both required to break that: see exitCmd.
func OpenConsultation(from orvyn.ScreenID) tea.Cmd {
	consultPending = true
	consultFrom = from
	consultRestorePrevious = orvyn.GetPreviousScreen()

	return orvyn.SwitchScreen(screen.IDClientUpdate)
}

// Screen drives the whole update: prompt, download, swap, restart. It also
// drives the read-only consultation flow entered via OpenConsultation.
type Screen struct {
	title    *orvyn.SimpleRenderable
	subtitle *orvyn.SimpleRenderable

	// detail carries plain, non-alert text the statusMessage widget
	// shouldn't own: the manual-download URL and the "press esc" hint. Those
	// need to stay plainly readable regardless of whatever warning/error is
	// showing in statusMessage above them.
	detail *orvyn.SimpleRenderable

	statusMessage *statusmessage.Widget

	notes *simplelogviewer.Widget

	bar *progressbar.Widget

	help *help.Widget

	layout *layout.CenterLayout

	state state

	// consultation is true for the remainder of this screen's lifetime once
	// OnEnter finds consultPending set. It changes esc's behavior (always
	// exits) and where esc goes (back to the caller, not to login).
	consultation bool

	// consultFrom and consultRestorePrevious are OnEnter's copy of the
	// package-level consultFrom/consultRestorePrevious vars, taken at the
	// same time and for the same reason s.consultation is a copy of
	// consultPending: exitCmd needs them long after OnEnter returns, and the
	// package vars could have been overwritten by then (a second
	// OpenConsultation call, from a different screen, while this one is
	// still on screen). See exitCmd.
	consultFrom            orvyn.ScreenID
	consultRestorePrevious orvyn.ScreenID

	result updater.Result

	progress atomic.Int64

	ticker *ticker.Ticker

	// progressCmd carries refreshProgress's returned tea.Cmd from the
	// ticker's onFire callback (which cannot itself return one) to Update's
	// orvyn.TickMsg case, which must batch it in - dropping it stops the
	// progress bar's animation dead.
	progressCmd tea.Cmd

	// generation counts OnEnter calls. It is bumped unconditionally, on both
	// paths, so any consultCheckedMsg tied to an older generation - a check
	// still in flight when the screen was left and re-entered, on either
	// path - is recognized as stale and dropped rather than clobbering
	// whatever the new entry just set up. See handleConsultChecked.
	generation uint64
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.subtitle = orvyn.NewSimpleRenderable("")
	s.subtitle.Style = t.Style(theme.DimTextStyleID)

	s.detail = orvyn.NewSimpleRenderable("")
	s.detail.Style = t.Style(theme.NormalTextStyleID)

	s.statusMessage = statusmessage.New()

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
			s.statusMessage,
			s.detail,
			orvyn.VGap,
			s.help,
		),
	)

	s.ticker = ticker.New(1, s.onProgressTick)

	return s
}

func (s *Screen) OnEnter(_ any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextClientUpdate)

	// Bumped unconditionally, before branching, so a consultCheckedMsg from
	// any earlier entry - consultation or startup - is stale from this point
	// on. See the doc comment on the generation field.
	s.generation++

	s.consultation = consultPending
	consultPending = false

	if s.consultation {
		s.consultFrom = consultFrom
		s.consultRestorePrevious = consultRestorePrevious

		return s.enterConsultation()
	}

	return s.enterPending()
}

// enterPending is the startup path: OnEnter's original behavior, consuming
// updater.Pending exactly as main.go left it.
func (s *Screen) enterPending() tea.Cmd {
	s.result = updater.Pending

	// main only ever switches to this screen on the startup path after
	// finding updater.Pending.Mode != ModeNone; reaching here with ModeNone
	// means something upstream is broken (e.g. a stale re-entry ping-ponging
	// through user settings - see OpenConsultation). Rendering the prompt
	// anyway would offer to reinstall the exact version already running,
	// with canEscape false (Mode is never ModeOptional/ModeMandatory here)
	// and so no way out but ctrl+c. Bail to login instead of ever showing
	// that.
	if s.result.Mode == updater.ModeNone {
		return orvyn.SwitchScreen(screen.IDLogin)
	}

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
	s.statusMessage.Reset()
	s.detail.SetValue("")

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

// decideEntry is the pure decision made at startup entry: given the update
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

// enterConsultation starts the read-only flow: it never touches
// updater.Pending, always re-fetches, and shows a checking state while the
// fetch (a tea.Cmd, not run inline here) is in flight so the UI stays
// responsive.
func (s *Screen) enterConsultation() tea.Cmd {
	s.result = updater.Result{}
	s.progress.Store(0)

	s.state = stateConsultChecking
	s.refreshHelpContext()

	s.title.SetValue(lokyn.L("Checking for updates..."))
	s.subtitle.SetValue("")
	s.detail.SetValue("")
	s.notes.SetActive(false)
	s.bar.SetActive(false)
	s.statusMessage.Reset()

	// gen is captured now, not read from s inside the closure: by the time
	// this command runs (and, worse, by the time its result comes back),
	// s.generation may have moved on to a later entry.
	gen := s.generation

	return func() tea.Msg {
		msg := checkForConsultation()

		if m, ok := msg.(consultCheckedMsg); ok {
			m.gen = gen
			return m
		}

		return msg
	}
}

// checkForConsultation is consultation mode's fresh check, run as a tea.Cmd
// so OnEnter returns immediately. It mirrors main.go's startup sequence -
// fetch the server's ClientTui compat string via request.VersionGet, then
// updater.Check - but reports the version-compat fetch's own failure back
// through consultCheckedMsg instead of printing and returning, since there is
// no os.Exit available (or wanted) from inside a running screen.
func checkForConsultation() tea.Msg {
	version, err := helper.Fetch[api.DbVersion](request.VersionGet())

	if err != nil {
		return consultCheckedMsg{err: err}
	}

	result := updater.Check(config.VERSION, version.ClientTui, viper.GetString("language"))

	return consultCheckedMsg{result: result}
}

// decideConsultEntry is decideEntry's counterpart for consultation mode,
// reached once a fresh check has come back (or failed to come back at all).
// checkErr is checkForConsultation's own failure - distinct from
// result.Err, which is set when Check's internal manifest fetch failed -
// checked first since without it there is no Result to look at.
//
// Unlike decideEntry, "already up to date" is a state of its own here: the
// startup path only ever reaches this screen when Pending.Mode != ModeNone,
// so decideEntry never had anywhere to send that case. Consultation is
// reachable at any time and must handle it. A newer release being available
// (ModeOptional or ModeMandatory) falls through to decideEntry unchanged, so
// a platform with no published file or an unwritable install directory gets
// exactly the same manual-required guidance the startup path would give it.
func decideConsultEntry(result updater.Result, checkErr, preflightErr error) (state, entryReason) {
	if checkErr != nil {
		return stateConsultFailed, reasonFetchFailed
	}

	if result.Err != nil {
		return stateConsultFailed, reasonFetchFailed
	}

	if result.Mode == updater.ModeNone {
		return stateConsultUpToDate, reasonNone
	}

	return decideEntry(result, preflightErr)
}

func (s *Screen) handleConsultChecked(msg consultCheckedMsg) tea.Cmd {
	// checkForConsultation is an in-flight network command; if the screen
	// was left and re-entered - on either path - before it returned, this
	// message belongs to a check OnEnter has already superseded. Applying it
	// now would rewrite s.result/s.state/title/subtitle out from under
	// whatever the new entry set up, including moving a startup-path re-entry
	// into a consultation state whose esc handler jumps to login instead of
	// the startup path's own exit.
	if msg.gen != s.generation {
		return nil
	}

	s.result = msg.result
	s.detail.SetValue("")

	preflightErr := updater.PreflightWritable()

	entryState, reason := decideConsultEntry(msg.result, msg.err, preflightErr)
	s.state = entryState
	s.refreshHelpContext()

	if entryState == stateConsultFailed {
		s.title.SetValue(lokyn.L("Could not check for updates"))
		s.subtitle.SetValue("")
		s.notes.SetActive(false)

		checkErr := msg.err
		if checkErr == nil {
			checkErr = msg.result.Err
		}

		s.statusMessage.SetError(checkErr)

		return nil
	}

	// Every other consult outcome has a real Latest to show: the only case
	// where Check leaves it empty (its own manifest fetch failing) was
	// handled by the stateConsultFailed branch above.
	s.subtitle.SetValue(fmt.Sprintf("%s  →  %s", msg.result.Current, msg.result.Latest))
	s.refreshNotes()

	switch entryState {
	case stateConsultUpToDate:
		s.title.SetValue(lokyn.L("Farental is up to date"))
		s.subtitle.SetValue(msg.result.Current)
		s.statusMessage.SetMessage(lokyn.L("You are running the latest version."), statusmessage.SuccessMessage)

	case statePrompt:
		s.title.SetValue(lokyn.L("A new version is available"))
		s.statusMessage.Reset()

	case stateManualRequired:
		switch reason {
		case reasonNoFile:
			s.enterManual(lokyn.L("No build is published for your platform."))
		case reasonNotWritable:
			s.enterManual(fmt.Sprintf("%s\n%v",
				lokyn.L("Farental cannot write to its own directory."), preflightErr))
		}
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
	case tea.WindowSizeMsg:
		// orvyn.SwitchScreen runs OnEnter synchronously from App.Init(),
		// before bubbletea's loop starts and so before the first resize
		// message arrives; refreshNotes there wraps to a hardcoded
		// default. This case is what makes it wrap to the real terminal
		// width, both at startup and on every later resize.
		//
		// Preserve the visibility state that the current state intends:
		// some states (e.g., downloading, manual-required) deliberately hide
		// the pane, but refreshNotes would force it visible if there's
		// content to wrap. Capture and restore to let resizes re-wrap without
		// re-activating.
		wasActive := s.notes.IsActive()
		s.refreshNotes()
		// refreshNotes sets active to true if there's content, false if not.
		// If there's content (active is now true), restore the original state.
		// If there's no content (active is now false), keep it inactive.
		if s.notes.IsActive() {
			s.notes.SetActive(wasActive)
		}

		return s.notes.Update(msg)

	case progress.FrameMsg:
		return s.bar.Update(msg)

	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return tea.Batch(cmd, s.progressCmd)

	case finishedMsg:
		return s.handleFinished(msg)

	case consultCheckedMsg:
		return s.handleConsultChecked(msg)
	}

	return s.notes.Update(msg)
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

// canEscape reports whether esc is allowed to leave the screen from the
// current state: always in consultation mode (the user opened this
// voluntarily and must never be trapped by it, even by a mandatory-update
// result), and - for the startup path, unchanged from before - only when the
// update itself is optional. A mandatory update at startup has no exit
// besides quitting.
func (s *Screen) canEscape() bool {
	return s.consultation || s.result.Mode == updater.ModeOptional
}

// exitCmd is where esc goes once canEscape allows it: back to whichever
// screen opened a consultation, or to login for the startup path, exactly as
// before.
//
// The consultation branch switches to s.consultFrom directly rather than
// orvyn.SwitchToPreviousScreen: orvyn keeps a single previousScreenID slot,
// which OpenConsultation's own SwitchScreen(IDClientUpdate) already
// overwrote with the caller's ID, so by now it holds no information exitCmd
// doesn't already have in s.consultFrom. Switching there also overwrites
// previousScreenID a second time - to this screen's own ID - which would
// otherwise make the caller's *next* esc bounce right back here; restoring
// it to whatever it was before the consultation (captured in
// s.consultRestorePrevious, see OpenConsultation) undoes that.
func (s *Screen) exitCmd() tea.Cmd {
	if s.consultation {
		cmd := orvyn.SwitchScreen(s.consultFrom)
		orvyn.SetPreviousScreen(s.consultRestorePrevious)

		return cmd
	}

	return orvyn.SwitchScreen(screen.IDLogin)
}

// refreshHelpContext keeps the help bar's advertised keys honest across the
// three read-only consultation states (still checking, already up to date,
// fetch failed): unlike every other state, enter does nothing there, and esc
// means "back to whichever screen opened this" rather than "skip". Every
// other state - including statePrompt reached via consultation, where enter
// does start the update - keeps the ordinary ContextClientUpdate keymap.
func (s *Screen) refreshHelpContext() {
	switch s.state {
	case stateConsultChecking, stateConsultUpToDate, stateConsultFailed:
		bubblehelp.SwitchContext(keybind.ContextClientUpdateConsult)
	default:
		bubblehelp.SwitchContext(keybind.ContextClientUpdate)
	}
}

func (s *Screen) handleKey(m tea.KeyMsg) (tea.Cmd, bool) {
	switch s.state {
	case statePrompt:
		switch {
		case key.Matches(m, keybind.Enter):
			return s.startUpdate(), true
		case key.Matches(m, keybind.Esc):
			if s.canEscape() {
				return s.exitCmd(), true
			}
		}

	case stateFailed:
		switch {
		case key.Matches(m, keybind.RKey):
			return s.startUpdate(), true
		case key.Matches(m, keybind.Esc):
			if s.canEscape() {
				return s.exitCmd(), true
			}
		}

	case stateManualRequired:
		switch {
		case key.Matches(m, keybind.Esc):
			if s.canEscape() {
				return s.exitCmd(), true
			}
		}

	case stateUnrecoverable:
		switch {
		case key.Matches(m, keybind.Esc):
			// No retry key here on purpose: the target binary is gone
			// and a retry cannot succeed.
			if s.canEscape() {
				return s.exitCmd(), true
			}
		}

	case stateConsultChecking, stateConsultUpToDate, stateConsultFailed:
		// These three states are only ever entered from enterConsultation,
		// so s.consultation is always true here: esc always leaves, no
		// canEscape gating needed.
		switch {
		case key.Matches(m, keybind.Esc):
			return s.exitCmd(), true
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
	s.statusMessage.SetMessage(lokyn.L("Downloading..."), statusmessage.InformationMessage)
	s.detail.SetValue("")

	file := s.result.File

	download := func() tea.Msg {
		return finishedMsg{err: updater.Apply(config.WebURL, file, &s.progress)}
	}

	// Restart (not Start) bumps the ticker's tag, invalidating any tick
	// still in flight from a previous attempt - the same thing the old
	// hand-rolled s.tickTag++ did on every call, including a retry.
	//
	// refreshProgress is also called directly, once, right here: Restart
	// only arms the *next* tick after a full interval, it never fires
	// onFire itself, so without this the bar would sit at 0% with no MB
	// counter for the first second of every download.
	return tea.Batch(download, s.refreshProgress(), s.ticker.Restart())
}

// onProgressTick is the ticker's onFire callback. A Ticker's onFire is a
// plain func() with no return value, so refreshProgress's tea.Cmd (which
// must keep reaching bubbletea or the progress bar's animation stalls) is
// stashed here for Update's orvyn.TickMsg case to pick up and batch in.
func (s *Screen) onProgressTick() {
	s.progressCmd = s.refreshProgress()
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

	s.statusMessage.SetMessage(fmt.Sprintf("%.1f / %.1f MB",
		float64(done)/bytesPerMB, float64(total)/bytesPerMB), statusmessage.InformationMessage)

	// The whole body is read before the swap, so a full bar means the
	// download finished and verification is under way.
	if done >= total {
		s.state = stateApplying
		s.statusMessage.SetMessage(lokyn.L("Verifying and installing..."), statusmessage.InformationMessage)
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
			s.statusMessage.SetMessage(fmt.Sprintf("%s\n%v\n\n%s",
				lokyn.L("The update failed."), msg.err,
				fmt.Sprintf(lokyn.L("The previous version was saved at %s. Reinstall it manually."), rerr.OldPath)),
				statusmessage.ErrorMessage)

			return nil
		}

		s.state = stateFailed
		// The retry key is not in the help keymap, which is fixed per context,
		// so the hint goes in the message the user is already reading.
		s.statusMessage.SetMessage(fmt.Sprintf("%s\n%v\n\n%s",
			lokyn.L("The update failed."), msg.err, lokyn.L("Press r to retry.")),
			statusmessage.ErrorMessage)

		return nil
	}

	s.state = stateRestarting
	s.statusMessage.SetMessage(lokyn.L("Restarting..."), statusmessage.InformationMessage)

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

	s.statusMessage.SetMessage(reason, statusmessage.WarningMessage)

	// Kept out of statusMessage, which renders one styled (here, warning-
	// colored) message: the URL needs to stay plainly readable rather than
	// tinted the same as the warning above it, and the hint is guidance, not
	// part of the warning itself.
	detail := fmt.Sprintf("%s\n%s/clienttui", lokyn.L("Download the new version here:"), config.WebURL)

	if s.canEscape() {
		detail = fmt.Sprintf("%s\n\n%s", detail, lokyn.L("Press esc to continue without updating."))
	}

	s.detail.SetValue(detail)
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
