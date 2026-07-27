package helper

import (
	"fmt"
	"os"
	"strings"
)

// Desktop notification protocols. Terminals disagree on how a notification is
// requested, and an unknown OSC number is parsed and discarded rather than
// printed, so exactly one of these is sent — sending several would pop several
// notifications on the terminals that understand more than one.
const (
	notifyNone = ""
	// notifyOSC9 carries a single message field. iTerm2, WezTerm, Ghostty.
	notifyOSC9 = "osc9"
	// notifyOSC777 carries a title and a body. foot, urxvt.
	notifyOSC777 = "osc777"
	// notifyOSC99 is kitty's own chunked protocol.
	notifyOSC99 = "osc99"
)

// notifyEnv names the environment variable forcing a notification protocol,
// bypassing the detection below. Needed inside multiplexers, where TERM
// describes the multiplexer instead of the terminal underneath it. Accepts the
// protocol constants above, or "off" to keep the bell as the only signal.
const notifyEnv = "FARENTAL_NOTIFY"

// SetTerminalTitle sets the terminal window/tab title through the OSC 2 escape
// sequence. The sequence occupies no screen cell and moves no cursor, so it is
// safe to write while the TUI renders; terminals without support ignore it.
func SetTerminalTitle(title string) {
	writeSequence(fmt.Sprintf("\x1b]2;%s\x1b\\", sanitizeEscape(title)))
}

// Bell rings the terminal bell, which terminals surface as an audible ping
// and/or a taskbar highlight.
func Bell() {
	fmt.Fprint(os.Stdout, "\a")
}

// Notify raises a desktop notification carrying title and body, on the
// terminals that support one. Terminals without support get nothing at all, so
// callers that need to be noticed should ring the Bell as well.
func Notify(title, body string) {
	title = sanitizeEscape(title)
	body = sanitizeEscape(body)

	switch detectNotifyProtocol() {
	case notifyOSC9:
		// A single field, so the two parts are joined into one line.
		writeSequence(fmt.Sprintf("\x1b]9;%s - %s\a", title, body))

	case notifyOSC777:
		writeSequence(fmt.Sprintf("\x1b]777;notify;%s;%s\a", title, body))

	case notifyOSC99:
		// d=0 marks the title chunk as not the last one; the body chunk
		// closes the notification.
		writeSequence(fmt.Sprintf("\x1b]99;i=1:d=0;%s\x1b\\", title) +
			fmt.Sprintf("\x1b]99;i=1:p=body;%s\x1b\\", body))
	}
}

// detectNotifyProtocol reports which notification protocol the terminal
// understands, or notifyNone when it understands none.
func detectNotifyProtocol() string {
	switch forced := os.Getenv(notifyEnv); forced {
	case notifyOSC9, notifyOSC777, notifyOSC99:
		return forced
	case "off":
		return notifyNone
	}

	term := os.Getenv("TERM")

	if os.Getenv("KITTY_WINDOW_ID") != "" || strings.Contains(term, "kitty") {
		return notifyOSC99
	}

	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm", "ghostty":
		return notifyOSC9
	}

	if os.Getenv("WEZTERM_PANE") != "" || os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return notifyOSC9
	}

	if strings.HasPrefix(term, "foot") || strings.HasPrefix(term, "rxvt-unicode") {
		return notifyOSC777
	}

	return notifyNone
}

// writeSequence writes an escape sequence to the terminal, wrapping it for tmux
// when running inside it. tmux swallows sequences it does not handle itself
// unless they come through its passthrough, which also needs
// "set -g allow-passthrough on".
func writeSequence(sequence string) {
	if os.Getenv("TMUX") != "" {
		sequence = "\x1bPtmux;" +
			strings.ReplaceAll(sequence, "\x1b", "\x1b\x1b") +
			"\x1b\\"
	}

	fmt.Fprint(os.Stdout, sequence)
}

// sanitizeEscape drops the control characters from text, which can come from
// the server: they would otherwise terminate the escape sequence early and
// leak the rest of the text onto the screen.
func sanitizeEscape(text string) string {
	return strings.Map(func(r rune) rune {
		if r < ' ' || r == 0x7f {
			return -1
		}

		return r
	}, text)
}
