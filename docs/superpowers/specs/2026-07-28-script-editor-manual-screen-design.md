# Script editor manual

## Goal

Give the script editor an in-app manual: a full-view title plus a scrollable
text viewer, holding a guide the user writes by hand. The component is generic
so other topics (fight, craft, bank) can reuse it later.

## Scope

- New embedded manual store with per-language plain-text files.
- New `manual` dialog (title + `simplelogviewer` + help bar).
- `F1` entry point from the script editor and the script explorer.
- Placeholder content only. Writing the guide itself is out of scope.

## Content storage

```
src/internal/manual/manual.go
src/internal/manual/docs/scripteditor_en.txt
src/internal/manual/docs/scripteditor_fr.txt
src/internal/manual/docs/scripteditor_de.txt
```

`manual.go` embeds the `docs` directory:

```go
//go:embed docs
var docs embed.FS
```

Public API:

```go
// Get returns the title and body lines of the manual for the given topic in
// the current language, falling back to English.
func Get(topic string) (title string, lines []string, err error)
```

Resolution:

1. Language comes from `lokyn.GetCurrentLanguage()`, reduced to its base tag
   (`fr-FR` -> `fr`).
2. Read `docs/<topic>_<lang>.txt`. On any read error, read
   `docs/<topic>_en.txt`.
3. If the English file is missing too, return an error.

Parsing: the first non-empty line of the file is the title, the remaining lines
are the body. Leading blank lines between the title and the body are dropped.
This keeps the title translated alongside the content, with no extra `lokyn`
keys.

Files are plain text, not markdown: nothing in the dependency set renders
markdown, so the author formats with blank lines, indentation, and ASCII
separators.

## Dialog, not a screen

Orvyn passes screen parameters through the *caller's* `OnExit()` return value,
not through `SwitchScreen` (`orvyn.go:121`). Reaching a manual *screen* from
the script editor would therefore re-enter `scripteditor.OnEnter` on the way
back with a parameter that is not a `*api.ScriptBasicResponse`, which resets
`s.new = true` and clears `s.data` - the user's unsaved rules would be
destroyed every time they read the manual.

A dialog avoids this. `orvyn.Render` (`orvyn.go:79`) replaces the layout with
the dialog's own layout rather than overlaying it, so a dialog is visually a
full screen. `OpenDialog(id, screen, param)` passes the topic directly, and the
calling screen is never re-entered, so editor state survives untouched.

Both call sites already re-arm the ticker and the running-task spinner
unconditionally on `orvyn.DialogExitMsg` (`scripteditor.go:269`,
`scriptexplorer.go:198`), so the manual inherits that handling for free. Their
`msg.Param.(uint)` assertions are guarded by a `msg.DialogID ==` comparison
first, so a distinct `"manual"` dialog ID cannot panic them.

## Dialog implementation

New package `src/screen/dialog/manual/manual.go`, alongside the existing
`screen/dialog/popup` and `screen/dialog/ruletypeselection`.

`manual.New()` returns an `orvyn.Screen` for use with `orvyn.OpenDialog`, and
`manual.Open(topic)` wraps that call with the package's own `DialogID`
(`"manual"`), so call sites never repeat the dialog ID. The topic constants
(`manual.TopicScriptEditor`) live here too, next to `Open`, so a caller needs a
single import.

Widgets, top to bottom, in a `CenterLayout` / `MaxWidthVBoxFullLayout` matching
`screen/logfull`:

- `title` - `orvyn.SimpleRenderable` styled with `theme.TitleStyleID`
- `logViewer` - `widget/simplelogviewer`, focused/blurred styles copied from
  `logfull`, focused on creation, empty viewer title (the heading lives in the
  title renderable, so the manual reads as a document rather than a boxed log)
- `help` - `widget/help`

`SetAutoScroll(false)` on the viewer so the manual opens at the top. `logfull`
relies on the default auto-scroll to jump to the newest event; the manual must
not.

Lifecycle:

- `OnEnter(i any)`: `i` must be a `string` topic. Store
  `bubblehelp.CurrentContext`, then switch to `keybind.ContextManual`. On a
  failed type assertion or a `manual.Get` error, return `orvyn.CloseDialog()`
  immediately.
- `OnExit()`: restore the stored bubblehelp context and return `nil`. Callers
  need no cleanup of their own.
- `Update`: `Esc` closes the dialog. `PrevPage` / `NextPage` scroll ten lines
  through `logViewer.ScrollUp` / `ScrollDown`. Everything else is forwarded to
  `logViewer.Update`, which handles up/down by one line.
- `Render`: returns the layout.

## Entry point

New binding in `internal/keybind/keybind.go`:

```go
F1Key = key.NewBinding(
    key.WithKeys("f1"),
    key.WithHelp("f1", lokyn.L("manual")))
```

`internal/keybind/context.go`:

- New context `ContextManual` with a keymap covering `Up`, `Down`, `PrevPage`,
  `NextPage`, `Esc`, `Quit`.
- `F1Key` added to `scriptExplorerKeymap` and to every non-input script editor
  keymap - `ScriptEditorWidgetNormalModeKeymap`,
  `ScriptEditorRulesListKeymap`, `ScriptEditorRuleInspectorNormalKeymap` -
  since the editor handles `F1` at screen level in all of those states.
- `lokyn` key `manual` added to `translations/{en,fr,de}.json`.

`screen/scripteditor/scripteditor.go`: handle `F1Key` in the key switch next to
`SKeyCtrl`, guarded the same way - only when
`!s.focusManager.IsInputting() && !s.list.IsInputting()`, behind the same
`bubblehelp.IsKeybindVisible` check - and return
`manual.Open(manual.TopicScriptEditor)`. Read-only scripts get the same entry.

`screen/scriptexplorer/scriptexplorer.go`: same call, guarded by
`s.GetFilteringState() != widgetlist.Filtering` like the screen's other keys,
so the manual can be read before opening a script.

`F1` is free in both screens: the script editor normal mode uses `1`, `2`, `3`,
`e`, `ctrl+s`, `tab`, `shift+tab`, `esc`, and `?` already belongs to
`bubblehelp.ShowAll`.

No new `orvyn.ScreenID` and no `main.go` registration: dialogs are constructed
at their call site, like `popup.NewYesNo`.

## Shipped content

Each of the three `.txt` files contains a translated title line, a blank line,
and one placeholder line. The user replaces the placeholders with the real
guide.

## Error handling

- Wrong `OnEnter` parameter type, unknown topic, missing English fallback:
  close the dialog silently. A missing file is a build-time mistake, not a
  runtime condition, so no status message is warranted.
- `Get` does not panic on an empty file: it yields an empty title and no body
  lines, which still renders.

## Verification

No TTY is available in this environment, so verification is `go build ./...`,
`go vet ./...`, and `gofmt -l` over the changed packages. Interactive checks
(F1 from both screens, scrolling, language fallback) are left to the user.
