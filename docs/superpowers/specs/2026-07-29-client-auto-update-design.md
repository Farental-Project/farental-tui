# Client auto-update

## Goal

Let the TUI client update itself in place from a single executable — no
launcher, no wrapper script. On startup the client learns which release the
server publishes, and either forces an update (its version is incompatible with
the server) or offers one (a newer patch exists), downloads the binary for its
platform, verifies it, swaps itself, and relaunches.

## Scope

- New JSON manifest route on the website (`farental-cli/serverweb`).
- New `linux-arm64` platform in the server's release model.
- New `internal/updater` package in the TUI.
- New `clientupdate` screen in the TUI.
- Corrected version compatibility comparison.
- `docs/VERSION_MANAGEMENT.md` release checklist update.

Out of scope: signing releases, delta updates, filling in
`scripts/install.sh`, updating the CLI (server-side) the same way.

## Why no launcher is needed

A running executable cannot be *written to* on Unix (`ETXTBSY`), but it can be
*renamed*: the running process holds the inode, so replacing the directory
entry is safe and atomic. Windows refuses to delete a running `.exe` but allows
renaming it aside. Both cases are handled without a second process:

- Unix: rename new binary over the target, then `syscall.Exec` — same PID, the
  process image is replaced in place.
- Windows: rename the running `.exe` to `.exe.old`, move the new one in, start
  it, exit; the `.old` is removed on the next launch.

## Current state

Server, `farental-cli`:

- `GET /version` (api.farental.ch) returns
  `DbVersion{Structure, Content, ClientTui, Cli}`. `ClientTui` is a
  `Major.Minor` compatibility string.
- `GET /clienttui` (farental.ch) renders the latest published release as HTML.
- `GET /clienttui/download/:fileID` streams a release binary, refusing
  unpublished releases.
- `ClientRelease{Version, IsPublished, PublishedAt}` with
  `ClientReleaseFile{Platform, Filename, SizeBytes, SHA256}` and
  `ClientReleaseTranslation{LanguageID, Notes}`. Admin UI at
  `/admin/releases`.
- `data.ClientPlatforms` covers `windows-amd64`, `linux-amd64`,
  `darwin-amd64`, `darwin-arm64`.

Client, `farental-tui`:

- `config.VERSION = "1.1.0"`, a hand-bumped const.
- `main.go` fetches `/version` before any UI exists and, on mismatch, prints
  "Please update it" and returns.
- Taskfile builds five targets, including `linux-arm64`.

Two gaps block a client-side updater: the download URL is keyed by a database
`fileID` the client cannot guess, and `linux-arm64` binaries are built but have
no platform to be uploaded under.

## Server changes (`farental-cli`)

### `linux-arm64` platform

Add to `src/model/data/clientrelease.go`:

```go
PlatformLinuxArm64 = "linux-arm64"
```

Append it to `ClientPlatforms` and add its case to `PlatformLabel`
("Linux (arm64)"). The admin upload form and the download page both iterate
`ClientPlatforms`, so the upload field and the download row appear with no
further change.

### Manifest route

`src/serverweb/routes.go`, alongside the existing two:

```go
app.Get("/clienttui/latest.json", controller.ClientTuiLatestJSON)
```

Controller in `src/serverweb/controller/clienttui.go`, reusing
`ctx.ClientReleases.FindLatestPublished()` — no new repository method, no new
column:

```json
{
  "version": "1.2.0",
  "published_at": "2026-07-29T10:00:00Z",
  "notes": "Fixed the fight timer.\nAdded arm64 builds.",
  "files": {
    "linux-amd64": {
      "filename": "Farental",
      "size_bytes": 15499426,
      "sha256": "9f2b...",
      "url": "/clienttui/download/42"
    }
  }
}
```

`files` is keyed by platform string. Those keys match
`runtime.GOOS + "-" + runtime.GOARCH` exactly, so the client needs no
translation table. `url` is relative; the client joins it to its web base URL.
No published release returns `404 {"error": "no published release"}`.

The existing `/clienttui/download/:fileID` handler is untouched — its
`IsPublished` guard already covers updater traffic.

### Release notes in the manifest

`notes` is the *existing* `ClientReleaseTranslation.Notes`, the same text the
`/clienttui` page shows. Release notes stay written in one place,
`/admin/releases/edit/:id`, and gain a second consumer.

Language resolution reuses what `ClientTuiView` already does — exact language,
then English, then lowest language ID — via `TranslationFor`. The client passes
its language as a query parameter:

```
GET /clienttui/latest.json?lang=fr
```

Notes are stored as HTML (both views render them with `templ.Raw`). A TUI
cannot render HTML, so the controller converts them to plain text before
serialising:

```go
notesPlain := srvutil.HTMLToText(t.Notes)
```

`HTMLToText` maps `<br>`, `</p>`, and `</li>` to newlines, drops the remaining
tags, decodes entities, and collapses runs of blank lines. It uses
`golang.org/x/net/html`, already present in `go.sum` as an indirect dependency;
promoting it to direct costs no new download.

## Client changes (`farental-tui`)

### Web base URL

The existing resty client is bound to `baseurl` (the API host). The manifest
and the binaries live on the website host, a different service, so a second
base URL is needed. In `internal/config/config.go`, next to `BaseURL`:

```go
var WebURL = "http://127.0.0.1:3001" // dev default = serverweb WEB_PORT
```

Injected like the existing two, with `WEB_URL: https://www.farental.ch` added
to the Taskfile vars:

```
-X 'farental/internal/config.WebURL={{.WEB_URL}}'
```

`internal/updater` uses plain `net/http` rather than resty: it streams a 22 MB
body and needs a byte counter, which resty's buffering does not suit.

### Version comparison

The current compatibility gate is wrong:

```go
!strings.HasPrefix(config.VERSION, version.ClientTui)
```

`strings.HasPrefix("1.10.0", "1.1")` is `true`, so a 1.10.x client would pass a
server demanding 1.1.x. Replace it with a parsed comparison: split on `.`,
compare major and minor as integers. The same parser answers "is the published
release newer than mine", which additionally needs patch.

Malformed versions on either side are treated as incompatible, not as equal.

### `internal/updater`

| File | Contents |
| --- | --- |
| `updater.go` | `Mode` (`None`/`Optional`/`Mandatory`), `Result`, `Check(current, compat, lang string) Result`, semver parse and compare, package-level `Pending Result` and `RestartPending bool` |
| `manifest.go` | Manifest structs, `fetch(lang string)` over `net/http` with a 10s timeout |
| `download.go` | HTTP GET, progress-counting reader over an `atomic.Int64` |
| `apply.go` | Writability preflight, `selfupdate.Apply`, `CleanupOld()` |
| `restart_unix.go` | `//go:build !windows` — `syscall.Exec` |
| `restart_windows.go` | `//go:build windows` — spawn inheriting stdio, then exit |

`Pending` as a package variable mirrors how `session.Expired()` /
`session.TakeExpired()` already hands state to a screen: the update screen
reads it in `OnEnter` instead of taking constructor arguments.

```go
type Result struct {
    Mode   Mode
    Latest string      // manifest version, empty when the fetch failed
    Notes  string
    File   FileInfo    // zero when this platform has no published file
    Err    error       // manifest fetch or decode failure
}
```

`Check` decides the mode:

- major/minor of `config.VERSION` differ from `DbVersion.ClientTui` →
  `Mandatory`, whether or not the manifest could be fetched.
- compatible, and the manifest version is greater than `config.VERSION` →
  `Optional`.
- otherwise `None`, including every case where the client is compatible and the
  manifest could not be read.

`Optional` therefore cannot occur without a usable manifest. `Mandatory` can:
the incompatibility is known from `/version` alone, and the screen falls back
to `ManualRequired` when `Err != nil` or `File` is zero.

### Download and swap

New dependency: `github.com/minio/selfupdate` (pure Go, no cgo).

```go
selfupdate.Apply(reader, selfupdate.Options{
    TargetPath: exePath,
    Checksum:   sha256Bytes,
})
```

`Apply` creates the new file next to the target itself, verifies the SHA-256,
performs the rename dance for the platform, and rolls back if the swap fails
partway. It is bought rather than built because the rollback path cannot be
exercised in this environment, and a bug there leaves users with no working
binary.

The HTTP body is passed straight to `Apply`, with no intermediate file of our
own:

```go
io.LimitReader(resp.Body, file.SizeBytes)
```

wrapped in the progress-counting reader. The bound is the exact advertised
size: a longer body is cut short and fails the checksum, a shorter one fails it
too. Between the two, a truncated, oversized, or tampered response can never be
swapped in. `Apply` buffers the payload in memory, so expect a ~22 MB spike
during the install.

`CleanupOld()` removes a leftover `<exe>.old` at startup and ignores any error.
It matters on Windows, where the old binary cannot be deleted by the process
replacing it, and is harmless elsewhere.

### `clientupdate` screen

`screen/clientupdate`, registered as
`IDClientUpdate orvyn.ScreenID = "clientupdate"`.

| State | Shows | Keys |
| --- | --- | --- |
| `Prompt` | current → new version, release notes | `enter` update · `esc` skip (`Optional` only) · `q` quit |
| `Downloading` | progress bar, `12.4 / 22.1 MB` | none |
| `Applying` | "Verifying and installing…" | none |
| `Restarting` | "Restarting…", then quits | — |
| `Failed` | error text | `r` retry · `esc` (`Optional` only) · `q` |
| `ManualRequired` | reason, `https://www.farental.ch/clienttui`, target version | `q` quit |

Progress uses `orvyn/widget/progressbar`, already available at the pinned orvyn
version (`e054d1a`), following the pattern in `orvyn/dialog/progress.go`:
forward `progress.FrameMsg` to the widget's `Update`, and on each tick set
`MaxValue`/`CurrentValue` and return the `Cmd` from `SetPercent`.

```go
cmd := w.progressBar.SetPercent(percent)
return tea.Batch(cmd, orvyn.TickCmd(1, w.tickTag))
```

`orvyn.TickCmd` takes whole seconds, so a fast download produces only a handful
of ticks; the harmonica spring inside `progress.Model` interpolates between
them, exactly as the progress dialog does. If it reads as chunky in practice,
that one call becomes a 200 ms `tea.Tick`.

`dialog.Progress` is not reused: it is built for `OpenDialog` over a live
screen and closes itself at 100%, while this is a standalone screen that also
carries version, notes, and keybinds.

The download runs as a `tea.Cmd` returning `updateFinishedMsg{err}`, batched
with the tick that reads the byte counter.

There is no cancel key during the download. Nothing is swapped until `Apply`
runs at the very end, so `ctrl+c` is already safe and leaves the installed
binary untouched.

All strings go through `lokyn.L` and are added to
`translations/{en,fr,de}.json`.

### Startup flow

`main.go` is reordered. The version check currently runs before any UI exists;
it has to move after screen registration so the update screen can be the first
screen shown.

1. `config.Init()`, log rotation, `context.Init()`, `request.Init()`,
   `lokyn` — unchanged.
2. `updater.CleanupOld()`.
3. Fetch `/version`. On error, print and exit, as today.
4. `updater.Pending = updater.Check(config.VERSION, version.ClientTui)`.
5. Remaining init: `keybind`, `bubblehelp`, `data.InitTargets`, `orvyn.Init`,
   theme, styles, screen registration — plus `IDClientUpdate`.
6. Start bubbletea with `App{startScreen: id}`, where `id` is
   `IDClientUpdate` when `Pending.Mode != None`, otherwise `IDLogin`.
7. After `p.Run()` returns, if `updater.RestartPending` then
   `updater.Restart()`.

`App.Init()` returns `orvyn.SwitchScreen(a.startScreen)` instead of a hardcoded
`screen.IDLogin`.

Step 7 is why the restart lives outside bubbletea: by the time `p.Run()`
returns, the alt screen has been left and the cursor restored. Calling
`syscall.Exec` from inside `Update` would hand the new process a terminal still
in raw mode.

## Failure handling

The columns are the two situations the client can be in, not the two modes: a
compatible client can be running with `Mode` `Optional` or `None`.

| Failure | Incompatible client | Compatible client |
| --- | --- | --- |
| `/version` unreachable | print and exit (unchanged) | same |
| `latest.json` unreachable or 404 | `ManualRequired`, blocks login | log only, straight to login |
| No file for this platform | `ManualRequired` | log only |
| Binary directory not writable | `ManualRequired`, names the path | log only |
| Download aborts, or SHA-256 mismatch | `Failed`, `r` retries | `Failed`, `esc` continues to login |
| `Apply` fails mid-swap | rollback, then `Failed` | same |
| Exec fails after a successful swap | print "Updated to X, relaunch Farental", exit 0 | same |

A compatible client is never blocked by a failed update check: it logs and goes
to login. Only a client the server will refuse anyway gets stopped.

The writability preflight creates and removes a temporary file in the
directory holding the executable, because the swap is a rename of a directory
entry, not a write to the binary itself. It runs before the download, so a
user who cannot install never waits for 22 MB.

No privilege escalation is attempted. A client installed into a root-owned
path reports the path and points at the download page.

## Release process

`docs/VERSION_MANAGEMENT.md`, Client section:

- "Upload the clients on the FTP for download from the website" becomes
  "Upload all five platform binaries via `/admin/releases/new`", including
  `linux-arm64`.
- Setting `IsPublished` is the point of no return: it is what pushes the update
  to every running client, not the upload.
- `config.VERSION` stays a hand-bumped const, already covered by "Commit the
  version change".

## Verification

`go test` cannot run in the development environment and the TUI needs a real
terminal, so verification splits in three.

Static, both repositories: `go build ./...`, `go vet ./...`, `gofmt -l`.

Unit tests, on logic that needs no terminal:

- semver parse and compare, including the `1.10.0` against `1.1` case that
  the current `HasPrefix` gets wrong
- platform key mapping for all five platforms
- manifest JSON decoding, including the 404 and missing-platform paths
- `HTMLToText` against representative stored notes

The server repository's tests run normally and cover the new route and
`HTMLToText`.

Manual, by hand on a real terminal: build 1.1.0, publish 1.1.1, run the client,
take the update, confirm it relaunches on the new version. Worth doing on both
Linux and Windows, since the swap and restart take different code paths.
