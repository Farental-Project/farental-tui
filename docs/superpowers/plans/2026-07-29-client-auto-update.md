# Client Auto-Update Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let the Farental TUI update itself in place from a single executable — detect, download, verify, swap, relaunch — with no launcher process.

**Architecture:** The website (`farental-cli/serverweb`) gains a JSON manifest route describing the latest published release, including per-platform SHA-256 and a download URL, plus release notes converted from Quill HTML into structured blocks. The TUI gains an `internal/updater` package that compares versions, fetches the manifest, downloads and applies the binary via `minio/selfupdate`, and a `clientupdate` screen that drives it. Restart happens after bubbletea has torn down, so the new process inherits a clean terminal.

**Tech Stack:** Go 1.25.0, bubbletea/lipgloss, orvyn (TUI framework), lokyn (i18n), resty (existing API client), `github.com/minio/selfupdate` (new), `golang.org/x/net/html` (server, promote to direct), Fiber + GORM (server).

**Spec:** `docs/superpowers/specs/2026-07-29-client-auto-update-design.md`

## Global Constraints

- **Two repositories.** Tasks 1–3 are in `/home/halsten/Dev/Farental/farental-cli` (server). Tasks 4–11 are in `/home/halsten/Dev/Farental/farental-tui` (client). Commit in the repo you are working in. Never mix a server change and a client change in one commit.
- **Go module roots are `src/`** in both repos. Run all `go` commands from `<repo>/src`.
- **Verification bar, both repos:** `go build ./...`, `go vet ./...` (exit 0, no output), `gofmt -l .` (no output), and `go test ./<touched packages>/...` passing. All four are currently clean on `main` — a failure is yours.
- **Unit tests run fine.** Do not skip tests claiming "no TTY". Only the *interactive* TUI (`go run .`) needs a real terminal and a reachable backend; unit tests on logic packages do not.
- **All user-facing client strings go through `lokyn.L("...")`** and must be added to all three of `src/translations/en.json`, `fr.json`, `de.json`. Those files are single-line JSON objects keyed by the English string.
- **No new client dependencies** beyond `github.com/minio/selfupdate`.
- **Platform strings are exactly** `linux-amd64`, `linux-arm64`, `windows-amd64`, `darwin-amd64`, `darwin-arm64` — identical on both sides, and identical to `runtime.GOOS + "-" + runtime.GOARCH`.
- **Version format** is `Major.Minor.Patch` (client `config.VERSION`), compared against the server's `Major.Minor` compatibility string (`DbVersion.ClientTui`).

---

## File Structure

**Server (`farental-cli/src`):**

| File | Responsibility |
| --- | --- |
| `model/data/clientrelease.go` (modify) | Add `linux-arm64` platform |
| `srvutil/notes.go` (create) | Quill HTML → `[]NoteBlock` |
| `srvutil/notes_test.go` (create) | Conversion tests |
| `serverweb/controller/clienttui.go` (modify) | `buildLatestJSON` + `ClientTuiLatestJSON` handler |
| `serverweb/controller/clienttui_test.go` (create) | `buildLatestJSON` tests, no DB |
| `serverweb/routes.go` (modify) | Register the route |

**Client (`farental-tui/src`):**

| File | Responsibility |
| --- | --- |
| `internal/config/config.go` (modify) | `WebURL` var |
| `Taskfile.yml` (modify) | Inject `WebURL` |
| `internal/updater/version.go` (create) | Semver parse/compare |
| `internal/updater/manifest.go` (create) | Manifest types + fetch |
| `internal/updater/updater.go` (create) | `Mode`, `Result`, `Check`, package state |
| `internal/updater/notes.go` (create) | `RenderNotes` |
| `internal/updater/apply.go` (create) | Preflight, download, apply, cleanup |
| `internal/updater/restart_unix.go` (create) | `syscall.Exec` restart |
| `internal/updater/restart_windows.go` (create) | Spawn-and-exit restart |
| `internal/updater/*_test.go` (create) | Logic tests |
| `screen/clientupdate/clientupdate.go` (create) | The update screen |
| `screen/screen.go` (modify) | `IDClientUpdate` |
| `internal/keybind/context.go` (modify) | Help context |
| `main.go`, `app.go` (modify) | Startup wiring, start screen, restart hook |
| `translations/{en,fr,de}.json` (modify) | New strings |

---

## Task 1: `linux-arm64` platform (server)

**Files:**
- Modify: `farental-cli/src/model/data/clientrelease.go:10-40`
- Test: `farental-cli/src/serverweb/controller/admin_releases_test.go` (existing, extend)

**Interfaces:**
- Consumes: nothing.
- Produces: `data.PlatformLinuxArm64 = "linux-arm64"`, present in `data.ClientPlatforms`.

- [ ] **Step 1: Write the failing test**

Add to `farental-cli/src/serverweb/controller/admin_releases_test.go`, and add the new row to the existing `TestReleasePathUsesDistLayout` table too:

```go
func TestLinuxArm64IsAKnownPlatform(t *testing.T) {
	found := false
	for _, p := range data.ClientPlatforms {
		if p == data.PlatformLinuxArm64 {
			found = true
		}
	}
	if !found {
		t.Fatalf("ClientPlatforms missing %q, got %v", data.PlatformLinuxArm64, data.ClientPlatforms)
	}

	if got := data.PlatformLabel(data.PlatformLinuxArm64); got != "Linux (arm64)" {
		t.Errorf("PlatformLabel = %q, want %q", got, "Linux (arm64)")
	}

	if got := data.PlatformExtension(data.PlatformLinuxArm64); got != "" {
		t.Errorf("PlatformExtension = %q, want empty", got)
	}
}
```

In `TestReleasePathUsesDistLayout`, add:

```go
		{data.PlatformLinuxArm64, "/srv/releases/1.0.0/linux-arm64/Farental"},
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go test ./serverweb/controller/ -run 'TestLinuxArm64IsAKnownPlatform|TestReleasePathUsesDistLayout' -v`
Expected: FAIL — `undefined: data.PlatformLinuxArm64`

- [ ] **Step 3: Write minimal implementation**

In `model/data/clientrelease.go`, add the constant, the slice entry, and the label case:

```go
const (
	PlatformWindowsAmd64 = "windows-amd64"
	PlatformLinuxAmd64   = "linux-amd64"
	PlatformLinuxArm64   = "linux-arm64"
	PlatformDarwinAmd64  = "darwin-amd64"
	PlatformDarwinArm64  = "darwin-arm64"
)

var ClientPlatforms = []string{
	PlatformWindowsAmd64,
	PlatformLinuxAmd64,
	PlatformLinuxArm64,
	PlatformDarwinAmd64,
	PlatformDarwinArm64,
}
```

And in `PlatformLabel`:

```go
	case PlatformLinuxArm64:
		return "Linux (arm64)"
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go test ./serverweb/controller/ -run 'TestLinuxArm64IsAKnownPlatform|TestReleasePathUsesDistLayout' -v`
Expected: PASS

- [ ] **Step 5: Verify nothing else broke**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go build ./... && go vet ./... && gofmt -l .`
Expected: no output from any of the three.

- [ ] **Step 6: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git add src/model/data/clientrelease.go src/serverweb/controller/admin_releases_test.go
git commit -m "feat: add linux-arm64 client release platform"
```

---

## Task 2: Quill HTML to note blocks (server)

The admin release-notes editor is Quill 2 with a closed toolbar (`admin_releases_edit.templ:41-49`): headers 2 and 3, bold/italic/underline/strike, ordered and bullet lists, blockquote, code-block, link. That fixes the tag set.

Three Quill 2 output shapes are easy to get wrong and are what the tests pin down:
- Bullet lists are `<ol><li data-list="bullet">`, **not** `<ul>`. The `data-list` attribute decides the kind, not the parent tag.
- Code blocks are `<div class="ql-code-block-container"><div class="ql-code-block">line</div>…</div>`, **not** `<pre>`.
- Nesting is `class="ql-indent-N"` on the `<li>`.

**Files:**
- Create: `farental-cli/src/srvutil/notes.go`
- Test: `farental-cli/src/srvutil/notes_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  ```go
  type NoteSpan struct{ Text string; Bold, Italic, Underline, Strike bool; Href string }
  type NoteItem struct{ Indent int; Spans []NoteSpan }
  type NoteBlock struct{ Type string; Spans []NoteSpan; Ordered bool; Items []NoteItem; Lines []string }
  func QuillToBlocks(source string) []NoteBlock
  ```
  Block `Type` is one of `"h2"`, `"h3"`, `"p"`, `"list"`, `"quote"`, `"code"`.

- [ ] **Step 1: Write the failing test**

Create `farental-cli/src/srvutil/notes_test.go`:

```go
package srvutil

import (
	"reflect"
	"testing"
)

func TestQuillToBlocksHeadingsAndParagraph(t *testing.T) {
	got := QuillToBlocks(`<h2>Fixes</h2><p>Fixed the <strong>fight timer</strong>.</p>`)

	want := []NoteBlock{
		{Type: "h2", Spans: []NoteSpan{{Text: "Fixes"}}},
		{Type: "p", Spans: []NoteSpan{
			{Text: "Fixed the "},
			{Text: "fight timer", Bold: true},
			{Text: "."},
		}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// Quill 2 emits bullet lists as <ol> with data-list="bullet", so the parent tag
// must not decide whether the list is ordered.
func TestQuillToBlocksBulletListUsesDataList(t *testing.T) {
	got := QuillToBlocks(`<ol><li data-list="bullet">one</li><li data-list="bullet">two</li></ol>`)

	want := []NoteBlock{{
		Type:    "list",
		Ordered: false,
		Items: []NoteItem{
			{Indent: 0, Spans: []NoteSpan{{Text: "one"}}},
			{Indent: 0, Spans: []NoteSpan{{Text: "two"}}},
		},
	}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestQuillToBlocksOrderedListAndIndent(t *testing.T) {
	got := QuillToBlocks(`<ol><li data-list="ordered">one</li>` +
		`<li data-list="ordered" class="ql-indent-1">nested</li></ol>`)

	want := []NoteBlock{{
		Type:    "list",
		Ordered: true,
		Items: []NoteItem{
			{Indent: 0, Spans: []NoteSpan{{Text: "one"}}},
			{Indent: 1, Spans: []NoteSpan{{Text: "nested"}}},
		},
	}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// A run of bullets followed by a run of numbers inside one <ol> must split into
// two blocks, because a single block carries one Ordered flag.
func TestQuillToBlocksSplitsMixedListRuns(t *testing.T) {
	got := QuillToBlocks(`<ol><li data-list="bullet">a</li><li data-list="ordered">b</li></ol>`)

	if len(got) != 2 {
		t.Fatalf("got %d blocks, want 2: %+v", len(got), got)
	}
	if got[0].Ordered || !got[1].Ordered {
		t.Errorf("got Ordered %v then %v, want false then true", got[0].Ordered, got[1].Ordered)
	}
}

func TestQuillToBlocksCodeBlock(t *testing.T) {
	got := QuillToBlocks(`<div class="ql-code-block-container">` +
		`<div class="ql-code-block">go build ./...</div>` +
		`<div class="ql-code-block">go vet ./...</div></div>`)

	want := []NoteBlock{{Type: "code", Lines: []string{"go build ./...", "go vet ./..."}}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestQuillToBlocksQuoteAndLink(t *testing.T) {
	got := QuillToBlocks(`<blockquote>See <a href="https://farental.ch">the site</a></blockquote>`)

	want := []NoteBlock{{Type: "quote", Spans: []NoteSpan{
		{Text: "See "},
		{Text: "the site", Href: "https://farental.ch"},
	}}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// Quill writes an empty line as <p><br></p>; it must not become a blank block.
func TestQuillToBlocksDropsEmptyParagraphs(t *testing.T) {
	got := QuillToBlocks(`<p>text</p><p><br></p>`)

	if len(got) != 1 {
		t.Fatalf("got %d blocks, want 1: %+v", len(got), got)
	}
}

func TestQuillToBlocksDecodesEntities(t *testing.T) {
	got := QuillToBlocks(`<p>a &amp; b</p>`)

	if got[0].Spans[0].Text != "a & b" {
		t.Errorf("got %q, want %q", got[0].Spans[0].Text, "a & b")
	}
}

// Content written before this feature, or pasted, may use tags outside the
// toolbar. Their text must survive rather than vanish.
func TestQuillToBlocksKeepsUnknownTagText(t *testing.T) {
	got := QuillToBlocks(`<section><p>kept</p></section>`)

	if len(got) != 1 || got[0].Spans[0].Text != "kept" {
		t.Errorf("got %+v, want a single block containing %q", got, "kept")
	}
}

func TestQuillToBlocksEmptyInput(t *testing.T) {
	if got := QuillToBlocks(""); got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go test ./srvutil/ -run TestQuillToBlocks -v`
Expected: FAIL — `undefined: QuillToBlocks`

- [ ] **Step 3: Write the implementation**

Create `farental-cli/src/srvutil/notes.go`:

```go
package srvutil

import (
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Note block types produced by QuillToBlocks.
const (
	NoteBlockH2    = "h2"
	NoteBlockH3    = "h3"
	NoteBlockP     = "p"
	NoteBlockList  = "list"
	NoteBlockQuote = "quote"
	NoteBlockCode  = "code"
)

// NoteSpan is a run of text sharing one set of inline marks.
type NoteSpan struct {
	Text      string `json:"text"`
	Bold      bool   `json:"bold,omitempty"`
	Italic    bool   `json:"italic,omitempty"`
	Underline bool   `json:"underline,omitempty"`
	Strike    bool   `json:"strike,omitempty"`
	Href      string `json:"href,omitempty"`
}

// NoteItem is one entry of a list block. Indent is the nesting level.
type NoteItem struct {
	Indent int        `json:"indent"`
	Spans  []NoteSpan `json:"spans"`
}

// NoteBlock is one structural element of a release note.
type NoteBlock struct {
	Type    string     `json:"type"`
	Spans   []NoteSpan `json:"spans,omitempty"`
	Ordered bool       `json:"ordered,omitempty"`
	Items   []NoteItem `json:"items,omitempty"`
	Lines   []string   `json:"lines,omitempty"`
}

// QuillToBlocks converts the HTML stored by the Quill editor into structured
// blocks. The client renders them; this function decides nothing about looks.
// Unknown tags degrade to their text content rather than being dropped.
func QuillToBlocks(source string) []NoteBlock {
	if strings.TrimSpace(source) == "" {
		return nil
	}

	doc, err := html.Parse(strings.NewReader(source))

	if err != nil {
		return nil
	}

	body := findElement(doc, "body")

	if body == nil {
		return nil
	}

	var blocks []NoteBlock

	for n := body.FirstChild; n != nil; n = n.NextSibling {
		blocks = append(blocks, blocksFromNode(n)...)
	}

	return blocks
}

type spanStyle struct {
	bold      bool
	italic    bool
	underline bool
	strike    bool
	href      string
}

func (s spanStyle) span(text string) NoteSpan {
	return NoteSpan{
		Text:      text,
		Bold:      s.bold,
		Italic:    s.italic,
		Underline: s.underline,
		Strike:    s.strike,
		Href:      s.href,
	}
}

func blocksFromNode(n *html.Node) []NoteBlock {
	if n.Type == html.TextNode {
		if strings.TrimSpace(n.Data) == "" {
			return nil
		}
		return []NoteBlock{{Type: NoteBlockP, Spans: []NoteSpan{{Text: n.Data}}}}
	}

	if n.Type != html.ElementNode {
		return nil
	}

	switch n.Data {
	case "h2":
		return textBlock(NoteBlockH2, n)
	case "h3":
		return textBlock(NoteBlockH3, n)
	case "p":
		return textBlock(NoteBlockP, n)
	case "blockquote":
		return textBlock(NoteBlockQuote, n)
	case "ol", "ul":
		return listBlocks(n)
	case "pre":
		return codeBlock(strings.Split(textContent(n), "\n"))
	case "div":
		if hasClass(n, "ql-code-block-container") {
			return codeBlock(codeLines(n))
		}
		return childBlocks(n)
	default:
		return childBlocks(n)
	}
}

func childBlocks(n *html.Node) []NoteBlock {
	var blocks []NoteBlock

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		blocks = append(blocks, blocksFromNode(c)...)
	}

	return blocks
}

func textBlock(kind string, n *html.Node) []NoteBlock {
	spans := inlineSpans(n, spanStyle{})

	if spansAreBlank(spans) {
		return nil
	}

	return []NoteBlock{{Type: kind, Spans: spans}}
}

func codeBlock(lines []string) []NoteBlock {
	if len(lines) == 0 {
		return nil
	}

	return []NoteBlock{{Type: NoteBlockCode, Lines: lines}}
}

func codeLines(container *html.Node) []string {
	var lines []string

	for c := container.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && hasClass(c, "ql-code-block") {
			lines = append(lines, textContent(c))
		}
	}

	return lines
}

// listBlocks starts a new block whenever the ordered/bullet kind changes,
// because one block carries a single Ordered flag.
func listBlocks(list *html.Node) []NoteBlock {
	var blocks []NoteBlock

	current := -1

	for li := list.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.Data != "li" {
			continue
		}

		kind := attr(li, "data-list")

		ordered := kind == "ordered"

		// Legacy content has no data-list; fall back to the parent tag.
		if kind == "" {
			ordered = list.Data == "ol"
		}

		item := NoteItem{
			Indent: indentLevel(li),
			Spans:  inlineSpans(li, spanStyle{}),
		}

		if current == -1 || blocks[current].Ordered != ordered {
			blocks = append(blocks, NoteBlock{Type: NoteBlockList, Ordered: ordered})
			current = len(blocks) - 1
		}

		blocks[current].Items = append(blocks[current].Items, item)
	}

	return blocks
}

func inlineSpans(n *html.Node, style spanStyle) []NoteSpan {
	var spans []NoteSpan

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch {
		case c.Type == html.TextNode:
			if c.Data == "" {
				continue
			}
			spans = append(spans, style.span(c.Data))

		case c.Type == html.ElementNode:
			next := style

			switch c.Data {
			case "br":
				spans = append(spans, style.span("\n"))
				continue
			case "strong", "b":
				next.bold = true
			case "em", "i":
				next.italic = true
			case "u":
				next.underline = true
			case "s", "strike", "del":
				next.strike = true
			case "a":
				next.href = attr(c, "href")
			}

			spans = append(spans, inlineSpans(c, next)...)
		}
	}

	return spans
}

func spansAreBlank(spans []NoteSpan) bool {
	for _, s := range spans {
		if strings.TrimSpace(s.Text) != "" {
			return false
		}
	}

	return true
}

func indentLevel(n *html.Node) int {
	for _, class := range strings.Fields(attr(n, "class")) {
		if !strings.HasPrefix(class, "ql-indent-") {
			continue
		}

		level, err := strconv.Atoi(strings.TrimPrefix(class, "ql-indent-"))

		if err == nil {
			return level
		}
	}

	return 0
}

func hasClass(n *html.Node, want string) bool {
	for _, class := range strings.Fields(attr(n, "class")) {
		if class == want {
			return true
		}
	}

	return false
}

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}

	return ""
}

func textContent(n *html.Node) string {
	var b strings.Builder

	var walk func(*html.Node)

	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)

	return b.String()
}

func findElement(n *html.Node, name string) *html.Node {
	if n.Type == html.ElementNode && n.Data == name {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findElement(c, name); found != nil {
			return found
		}
	}

	return nil
}
```

- [ ] **Step 4: Promote `x/net` to a direct dependency**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go mod tidy`
Expected: `golang.org/x/net` moves out of the `// indirect` block in `go.mod`. No new module is downloaded — it was already in `go.sum`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go test ./srvutil/ -run TestQuillToBlocks -v`
Expected: PASS, 10 tests.

- [ ] **Step 6: Verify the repo is clean**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go build ./... && go vet ./... && gofmt -l .`
Expected: no output.

- [ ] **Step 7: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git add src/srvutil/notes.go src/srvutil/notes_test.go src/go.mod src/go.sum
git commit -m "feat: convert Quill release notes to structured blocks"
```

---

## Task 3: `latest.json` manifest route (server)

**Files:**
- Modify: `farental-cli/src/serverweb/controller/clienttui.go`
- Modify: `farental-cli/src/serverweb/routes.go:44-45`
- Test: `farental-cli/src/serverweb/controller/clienttui_test.go` (create)

**Interfaces:**
- Consumes: `srvutil.QuillToBlocks`, `srvutil.NoteBlock` (Task 2); `data.PlatformLinuxArm64` (Task 1).
- Produces: `GET /clienttui/latest.json?lang=<code>` returning the JSON shape below, and `buildLatestJSON(release *data.ClientRelease, langID, fallbackID uint) latestReleaseJSON`.

The handler stays a thin wrapper so the payload logic is testable without a database — the controller package's existing tests use no DB, and this must not be the first to need one.

- [ ] **Step 1: Write the failing test**

Create `farental-cli/src/serverweb/controller/clienttui_test.go`:

```go
package controller

import (
	"farental/model/data"
	"testing"
	"time"
)

func testRelease() *data.ClientRelease {
	published := time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC)

	release := &data.ClientRelease{
		Version:     "1.2.0",
		IsPublished: true,
		PublishedAt: &published,
		Files: []data.ClientReleaseFile{
			{Platform: data.PlatformLinuxAmd64, Filename: "Farental", SizeBytes: 15499426, SHA256: "abc"},
			{Platform: data.PlatformWindowsAmd64, Filename: "Farental.exe", SizeBytes: 16000000, SHA256: "def"},
		},
		Translations: []data.ClientReleaseTranslation{
			{LanguageID: 1, Notes: "<p>English notes</p>"},
			{LanguageID: 2, Notes: "<p>Notes en francais</p>"},
		},
	}

	release.Files[0].ID = 42
	release.Files[1].ID = 43

	return release
}

func TestBuildLatestJSONFilesKeyedByPlatform(t *testing.T) {
	got := buildLatestJSON(testRelease(), 1, 1)

	if got.Version != "1.2.0" {
		t.Errorf("Version = %q, want %q", got.Version, "1.2.0")
	}

	linux, ok := got.Files[data.PlatformLinuxAmd64]

	if !ok {
		t.Fatalf("missing %q key, got %v", data.PlatformLinuxAmd64, got.Files)
	}

	if linux.URL != "/clienttui/download/42" {
		t.Errorf("URL = %q, want %q", linux.URL, "/clienttui/download/42")
	}

	if linux.SHA256 != "abc" || linux.SizeBytes != 15499426 || linux.Filename != "Farental" {
		t.Errorf("file metadata wrong: %+v", linux)
	}
}

func TestBuildLatestJSONUsesRequestedLanguage(t *testing.T) {
	got := buildLatestJSON(testRelease(), 2, 1)

	if len(got.NotesBlocks) != 1 || got.NotesBlocks[0].Spans[0].Text != "Notes en francais" {
		t.Errorf("got %+v, want the French notes", got.NotesBlocks)
	}
}

func TestBuildLatestJSONFallsBackToEnglish(t *testing.T) {
	got := buildLatestJSON(testRelease(), 99, 1)

	if len(got.NotesBlocks) != 1 || got.NotesBlocks[0].Spans[0].Text != "English notes" {
		t.Errorf("got %+v, want the English notes", got.NotesBlocks)
	}
}

func TestBuildLatestJSONWithoutTranslations(t *testing.T) {
	release := testRelease()
	release.Translations = nil

	got := buildLatestJSON(release, 1, 1)

	if got.NotesBlocks != nil {
		t.Errorf("NotesBlocks = %+v, want nil", got.NotesBlocks)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go test ./serverweb/controller/ -run TestBuildLatestJSON -v`
Expected: FAIL — `undefined: buildLatestJSON`

- [ ] **Step 3: Write the implementation**

Append to `farental-cli/src/serverweb/controller/clienttui.go` (and add `"strings"`, `"time"`, `"farental/srvutil"` to its imports — `fmt`, `os`, `strconv`, `errors`, `fiber`, `gorm`, `lokyn`, `views`, `util` are already there):

```go
type releaseFileJSON struct {
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
	URL       string `json:"url"`
}

type latestReleaseJSON struct {
	Version     string                     `json:"version"`
	PublishedAt *time.Time                 `json:"published_at"`
	NotesBlocks []srvutil.NoteBlock        `json:"notes_blocks"`
	Files       map[string]releaseFileJSON `json:"files"`
}

// buildLatestJSON shapes a release into the manifest the TUI updater reads.
// Kept separate from the handler so it is testable without a database.
func buildLatestJSON(release *data.ClientRelease, langID, fallbackID uint) latestReleaseJSON {
	payload := latestReleaseJSON{
		Version:     release.Version,
		PublishedAt: release.PublishedAt,
		Files:       make(map[string]releaseFileJSON, len(release.Files)),
	}

	if t := release.TranslationFor(langID, fallbackID); t != nil {
		payload.NotesBlocks = srvutil.QuillToBlocks(t.Notes)
	}

	for _, f := range release.Files {
		payload.Files[f.Platform] = releaseFileJSON{
			Filename:  f.Filename,
			SizeBytes: f.SizeBytes,
			SHA256:    f.SHA256,
			URL:       fmt.Sprintf("/clienttui/download/%d", f.ID),
		}
	}

	return payload
}

// ClientTuiLatestJSON serves the machine-readable manifest of the latest
// published client release, used by the in-client updater.
func ClientTuiLatestJSON(c *fiber.Ctx) error {
	ctx := newWebCtx()
	release, err := ctx.ClientReleases.FindLatestPublished()

	// A real database failure must not masquerade as "no release published",
	// or an outage looks like "no update available" to every client. Mirrors
	// how ClientTuiView separates the two.
	if errors.Is(err, gorm.ErrRecordNotFound) || release == nil {
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"error": "no published release"})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "internal error"})
	}

	// The client sends its lokyn code ("en"); language codes are stored
	// uppercase, as ClientTuiView's "EN" fallback shows.
	lang := strings.ToUpper(c.Query("lang", "EN"))

	fallbackID := srvutil.GetLanguageIDByCode("EN")

	langID := srvutil.GetLanguageIDByCode(lang)

	return c.JSON(buildLatestJSON(release, langID, fallbackID))
}
```

Add `"farental/model/data"` to the import block too — `buildLatestJSON` takes a `*data.ClientRelease`.

- [ ] **Step 4: Register the route**

In `farental-cli/src/serverweb/routes.go`, directly after the existing `/clienttui/download/:fileID` line:

```go
	app.Get("/clienttui/latest.json", optAuth, controller.ClientTuiLatestJSON)
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go test ./serverweb/controller/ -run TestBuildLatestJSON -v`
Expected: PASS, 4 tests.

- [ ] **Step 6: Verify the whole server repo**

Run: `cd /home/halsten/Dev/Farental/farental-cli/src && go build ./... && go vet ./... && gofmt -l . && go test ./srvutil/... ./serverweb/...`
Expected: builds clean, no vet or gofmt output, tests `ok`.

- [ ] **Step 7: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git add src/serverweb/controller/clienttui.go src/serverweb/controller/clienttui_test.go src/serverweb/routes.go
git commit -m "feat: serve client release manifest at /clienttui/latest.json"
```

---

## Task 4: Web base URL in the client

**Files:**
- Modify: `farental-tui/src/internal/config/config.go:17-20`
- Modify: `farental-tui/src/Taskfile.yml:6-14`

**Interfaces:**
- Consumes: nothing.
- Produces: `config.WebURL string` — the website origin, no trailing slash.

The existing resty client is bound to `baseurl` (the API host, `api.farental.ch`). The manifest and binaries live on the website host, a different service, so a second base URL is required.

- [ ] **Step 1: Add the variable**

In `internal/config/config.go`, beside the existing injected vars:

```go
var BaseURL = "http://127.0.0.1:3000"  // valeur par défaut (dev)
var WebURL = "http://127.0.0.1:3001"   // site web, valeur par défaut (dev)
var ConfigFileName = "farental_dev"    // valeur par défaut (dev)
```

`3001` is the server's `WEB_PORT` default (`farental-cli/src/config/env.go:131-134`).

- [ ] **Step 2: Inject it at build time**

In `Taskfile.yml`, add the var and extend `LDFLAGS`:

```yaml
vars:
  OUTPUT_NAME: Farental
  BASE_URL: https://api.farental.ch
  WEB_URL: https://www.farental.ch
  CONFIG_FILE_NAME: farental
  DIST_DIR: dist
  LDFLAGS: >-
    -X 'farental/internal/config.BaseURL={{.BASE_URL}}'
    -X 'farental/internal/config.WebURL={{.WEB_URL}}'
    -X 'farental/internal/config.ConfigFileName={{.CONFIG_FILE_NAME}}'
```

- [ ] **Step 3: Verify**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go build ./... && go vet ./... && gofmt -l .`
Expected: no output.

- [ ] **Step 4: Verify the injection actually works**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go build -o /tmp/farental-ldflags-check -ldflags "-X 'farental/internal/config.WebURL=https://example.test'" . && strings /tmp/farental-ldflags-check | grep -c 'https://example.test' && rm /tmp/farental-ldflags-check`
Expected: a count of at least 1 — the injected string is present in the binary.

- [ ] **Step 5: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/internal/config/config.go src/Taskfile.yml
git commit -m "feat: add injectable web base URL"
```

---

## Task 5: Version comparison

The current gate in `main.go` is `!strings.HasPrefix(config.VERSION, version.ClientTui)`. `strings.HasPrefix("1.10.0", "1.1")` is `true`, so a 1.10.x client passes a server demanding 1.1.x. This task replaces it with parsed comparison.

**Files:**
- Create: `farental-tui/src/internal/updater/version.go`
- Test: `farental-tui/src/internal/updater/version_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  ```go
  func Compatible(clientVersion, serverCompat string) bool
  func Newer(candidate, current string) bool
  ```
  Both return `false` when either argument is malformed.

- [ ] **Step 1: Write the failing test**

Create `farental-tui/src/internal/updater/version_test.go`:

```go
package updater

import "testing"

func TestCompatible(t *testing.T) {
	tests := []struct {
		client string
		compat string
		want   bool
	}{
		{"1.1.0", "1.1", true},
		{"1.1.7", "1.1", true},
		// The bug the old strings.HasPrefix gate had: "1.10.0" starts with
		// "1.1" but is a different minor version.
		{"1.10.0", "1.1", false},
		{"1.2.0", "1.1", false},
		{"2.1.0", "1.1", false},
		{"1.1.0", "1.1.0", true},
		{"", "1.1", false},
		{"1.1.0", "", false},
		{"nonsense", "1.1", false},
		{"1.1.0", "nonsense", false},
	}

	for _, tt := range tests {
		if got := Compatible(tt.client, tt.compat); got != tt.want {
			t.Errorf("Compatible(%q, %q) = %v, want %v", tt.client, tt.compat, got, tt.want)
		}
	}
}

func TestNewer(t *testing.T) {
	tests := []struct {
		candidate string
		current   string
		want      bool
	}{
		{"1.1.1", "1.1.0", true},
		{"1.2.0", "1.1.9", true},
		{"2.0.0", "1.9.9", true},
		{"1.1.0", "1.1.0", false},
		{"1.1.0", "1.1.1", false},
		{"1.9.0", "1.10.0", false},
		{"1.10.0", "1.9.0", true},
		{"", "1.1.0", false},
		{"1.1.0", "", false},
	}

	for _, tt := range tests {
		if got := Newer(tt.candidate, tt.current); got != tt.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", tt.candidate, tt.current, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ -v`
Expected: FAIL — `undefined: Compatible`

- [ ] **Step 3: Write the implementation**

Create `farental-tui/src/internal/updater/version.go`:

```go
package updater

import (
	"strconv"
	"strings"
)

type version struct {
	major int
	minor int
	patch int
}

// parseVersion reads "Major.Minor" or "Major.Minor.Patch". A missing patch is
// zero. Anything else fails, and callers treat failure as incompatible rather
// than as a match.
func parseVersion(s string) (version, bool) {
	parts := strings.Split(strings.TrimSpace(s), ".")

	if len(parts) < 2 || len(parts) > 3 {
		return version{}, false
	}

	numbers := make([]int, 3)

	for i, p := range parts {
		n, err := strconv.Atoi(p)

		if err != nil || n < 0 {
			return version{}, false
		}

		numbers[i] = n
	}

	return version{major: numbers[0], minor: numbers[1], patch: numbers[2]}, true
}

// Compatible reports whether the client may talk to a server advertising
// serverCompat. Major and minor must match exactly; the patch is ignored.
func Compatible(clientVersion, serverCompat string) bool {
	client, ok := parseVersion(clientVersion)

	if !ok {
		return false
	}

	server, ok := parseVersion(serverCompat)

	if !ok {
		return false
	}

	return client.major == server.major && client.minor == server.minor
}

// Newer reports whether candidate is a strictly later version than current.
func Newer(candidate, current string) bool {
	c, ok := parseVersion(candidate)

	if !ok {
		return false
	}

	base, ok := parseVersion(current)

	if !ok {
		return false
	}

	switch {
	case c.major != base.major:
		return c.major > base.major
	case c.minor != base.minor:
		return c.minor > base.minor
	default:
		return c.patch > base.patch
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ -v`
Expected: PASS, 2 tests.

- [ ] **Step 5: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/internal/updater/version.go src/internal/updater/version_test.go
git commit -m "feat: add version comparison for client updates"
```

---

## Task 6: Manifest fetch and update check

**Files:**
- Create: `farental-tui/src/internal/updater/manifest.go`
- Create: `farental-tui/src/internal/updater/updater.go`
- Test: `farental-tui/src/internal/updater/manifest_test.go`

**Interfaces:**
- Consumes: `Compatible`, `Newer` (Task 5); `config.WebURL` (Task 4).
- Produces:
  ```go
  type Span struct{ Text string; Bold, Italic, Underline, Strike bool; Href string }
  type Item struct{ Indent int; Spans []Span }
  type Block struct{ Type string; Spans []Span; Ordered bool; Items []Item; Lines []string }
  type FileInfo struct{ Filename string; SizeBytes int64; SHA256 string; URL string }

  type Mode int
  const (ModeNone Mode = iota; ModeOptional; ModeMandatory)

  type Result struct{ Mode Mode; Current, Latest string; Notes []Block; File FileInfo; Err error }

  func PlatformKey() string
  func Check(currentVersion, serverCompat, lang string) Result
  var Pending Result
  var RestartPending bool
  ```
  `Result.File` is the zero value when this platform has no published file. Block `Type` values match the server's: `"h2"`, `"h3"`, `"p"`, `"list"`, `"quote"`, `"code"`.

- [ ] **Step 1: Write the failing test**

Create `farental-tui/src/internal/updater/manifest_test.go`:

```go
package updater

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

const sampleManifest = `{
  "version": "1.2.0",
  "published_at": "2026-07-29T10:00:00Z",
  "notes_blocks": [
    {"type": "h2", "spans": [{"text": "Fixes"}]},
    {"type": "p", "spans": [{"text": "Fixed the "}, {"text": "timer", "bold": true}]},
    {"type": "list", "items": [{"indent": 0, "spans": [{"text": "one"}]}]}
  ],
  "files": {
    "PLATFORM": {
      "filename": "Farental",
      "size_bytes": 15499426,
      "sha256": "abc",
      "url": "/clienttui/download/42"
    }
  }
}`

func manifestServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clienttui/latest.json" {
			t.Errorf("path = %q, want /clienttui/latest.json", r.URL.Path)
		}

		if r.URL.Query().Get("lang") != "fr" {
			t.Errorf("lang = %q, want fr", r.URL.Query().Get("lang"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))

	t.Cleanup(srv.Close)

	return srv
}

func TestPlatformKeyMatchesServerConstants(t *testing.T) {
	got := PlatformKey()
	want := runtime.GOOS + "-" + runtime.GOARCH

	if got != want {
		t.Errorf("PlatformKey() = %q, want %q", got, want)
	}

	known := map[string]bool{
		"linux-amd64": true, "linux-arm64": true, "windows-amd64": true,
		"darwin-amd64": true, "darwin-arm64": true,
	}

	if !known[got] {
		t.Logf("running on unsupported platform %q; updater will report no file", got)
	}
}

func TestFetchManifestDecodesBlocks(t *testing.T) {
	body := replacePlatform(sampleManifest)
	srv := manifestServer(t, http.StatusOK, body)

	m, err := fetchManifest(srv.URL, "fr")

	if err != nil {
		t.Fatalf("fetchManifest returned %v", err)
	}

	if m.Version != "1.2.0" {
		t.Errorf("Version = %q, want 1.2.0", m.Version)
	}

	if len(m.NotesBlocks) != 3 {
		t.Fatalf("got %d blocks, want 3", len(m.NotesBlocks))
	}

	if m.NotesBlocks[0].Type != "h2" || m.NotesBlocks[0].Spans[0].Text != "Fixes" {
		t.Errorf("first block = %+v", m.NotesBlocks[0])
	}

	if !m.NotesBlocks[1].Spans[1].Bold {
		t.Errorf("second span should be bold: %+v", m.NotesBlocks[1].Spans[1])
	}

	if m.NotesBlocks[2].Items[0].Spans[0].Text != "one" {
		t.Errorf("list item = %+v", m.NotesBlocks[2].Items)
	}

	f, ok := m.Files[PlatformKey()]

	if !ok {
		t.Fatalf("no entry for %q in %v", PlatformKey(), m.Files)
	}

	if f.SHA256 != "abc" || f.SizeBytes != 15499426 || f.URL != "/clienttui/download/42" {
		t.Errorf("file = %+v", f)
	}
}

func TestFetchManifestNotFound(t *testing.T) {
	srv := manifestServer(t, http.StatusNotFound, `{"error":"no published release"}`)

	if _, err := fetchManifest(srv.URL, "fr"); err == nil {
		t.Fatal("expected an error for 404, got nil")
	}
}

func TestCheckMandatoryWhenIncompatible(t *testing.T) {
	body := replacePlatform(sampleManifest)
	srv := manifestServer(t, http.StatusOK, body)

	got := checkAt(srv.URL, "1.1.0", "1.2", "fr")

	if got.Mode != ModeMandatory {
		t.Errorf("Mode = %v, want ModeMandatory", got.Mode)
	}

	if got.Latest != "1.2.0" || got.File.SHA256 != "abc" {
		t.Errorf("Result = %+v", got)
	}
}

// An unreachable manifest must not hide the fact that the server refuses this
// client: the mode stays mandatory and the error is carried for the screen.
func TestCheckMandatoryWithUnreachableManifest(t *testing.T) {
	got := checkAt("http://127.0.0.1:1", "1.1.0", "1.2", "fr")

	if got.Mode != ModeMandatory {
		t.Errorf("Mode = %v, want ModeMandatory", got.Mode)
	}

	if got.Err == nil {
		t.Error("Err = nil, want the fetch failure")
	}
}

func TestCheckOptionalWhenNewerPatchExists(t *testing.T) {
	body := replacePlatform(sampleManifest)
	srv := manifestServer(t, http.StatusOK, body)

	got := checkAt(srv.URL, "1.2.0", "1.2", "fr")

	if got.Mode != ModeNone {
		t.Errorf("Mode = %v, want ModeNone for an up-to-date client", got.Mode)
	}

	got = checkAt(srv.URL, "1.1.0", "1.1", "fr")

	if got.Mode != ModeOptional {
		t.Errorf("Mode = %v, want ModeOptional", got.Mode)
	}
}

// A compatible client is never blocked by a failed update check.
func TestCheckCompatibleWithUnreachableManifest(t *testing.T) {
	got := checkAt("http://127.0.0.1:1", "1.1.0", "1.1", "fr")

	if got.Mode != ModeNone {
		t.Errorf("Mode = %v, want ModeNone", got.Mode)
	}
}
```

Add this helper at the bottom of the same file:

```go
func replacePlatform(body string) string {
	return strings.Replace(body, "PLATFORM", PlatformKey(), 1)
}
```

and add `"strings"` to the test file's imports.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ -run 'TestPlatformKey|TestFetchManifest|TestCheck' -v`
Expected: FAIL — `undefined: PlatformKey`

- [ ] **Step 3: Write the manifest layer**

Create `farental-tui/src/internal/updater/manifest.go`:

```go
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

// fetchTimeout bounds the manifest request. The client blocks on it at
// startup, so it must fail fast on a dead or slow host.
const fetchTimeout = 10 * time.Second

// Span is a run of note text sharing one set of inline marks.
type Span struct {
	Text      string `json:"text"`
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	Underline bool   `json:"underline"`
	Strike    bool   `json:"strike"`
	Href      string `json:"href"`
}

// Item is one entry of a list block.
type Item struct {
	Indent int    `json:"indent"`
	Spans  []Span `json:"spans"`
}

// Block is one structural element of a release note. Type is one of
// "h2", "h3", "p", "list", "quote", "code".
type Block struct {
	Type    string `json:"type"`
	Spans   []Span `json:"spans"`
	Ordered bool   `json:"ordered"`
	Items   []Item `json:"items"`
	Lines   []string `json:"lines"`
}

// FileInfo describes one downloadable binary.
type FileInfo struct {
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
	URL       string `json:"url"`
}

type manifest struct {
	Version     string              `json:"version"`
	PublishedAt time.Time           `json:"published_at"`
	NotesBlocks []Block             `json:"notes_blocks"`
	Files       map[string]FileInfo `json:"files"`
}

// PlatformKey returns this build's platform string, matching the server's
// data.Platform* constants exactly.
func PlatformKey() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

func fetchManifest(baseURL, lang string) (*manifest, error) {
	endpoint, err := url.JoinPath(strings.TrimSuffix(baseURL, "/"), "clienttui", "latest.json")

	if err != nil {
		return nil, err
	}

	endpoint += "?lang=" + url.QueryEscape(lang)

	client := &http.Client{Timeout: fetchTimeout}

	resp, err := client.Get(endpoint)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest request failed with status %d", resp.StatusCode)
	}

	var m manifest

	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}
```

- [ ] **Step 4: Write the check layer**

Create `farental-tui/src/internal/updater/updater.go`:

```go
package updater

import (
	"farental/internal/config"
	"log"
)

// Mode says what the client must or may do about its version.
type Mode int

const (
	// ModeNone means the client is compatible and up to date, or the check
	// failed harmlessly.
	ModeNone Mode = iota

	// ModeOptional means a newer release exists but this client still works.
	ModeOptional

	// ModeMandatory means the server refuses this client's version.
	ModeMandatory
)

// Result is the outcome of Check, consumed by the clientupdate screen.
type Result struct {
	Mode    Mode
	Current string
	Latest  string
	Notes   []Block
	File    FileInfo
	Err     error
}

// HasFile reports whether a binary exists for this platform.
func (r Result) HasFile() bool {
	return r.File.URL != "" && r.File.SHA256 != ""
}

// Pending carries the startup check to the update screen, the way
// session.Expired() carries session state to the login screen.
var Pending Result

// RestartPending tells main to exec the new binary once bubbletea has torn
// down the terminal.
var RestartPending bool

// Check compares this client against the server and the published release.
// It never returns an error: a failure is reported through Result.
func Check(currentVersion, serverCompat, lang string) Result {
	return checkAt(config.WebURL, currentVersion, serverCompat, lang)
}

func checkAt(baseURL, currentVersion, serverCompat, lang string) Result {
	result := Result{Current: currentVersion}

	mandatory := !Compatible(currentVersion, serverCompat)

	m, err := fetchManifest(baseURL, lang)

	if err != nil {
		result.Err = err

		// An incompatible client is refused by the server regardless, so it
		// still stops; a compatible one carries on to the login screen.
		if mandatory {
			result.Mode = ModeMandatory
		}

		log.Println("update check failed:", err)

		return result
	}

	result.Latest = m.Version
	result.Notes = m.NotesBlocks
	result.File = m.Files[PlatformKey()]

	switch {
	case mandatory:
		result.Mode = ModeMandatory
	case Newer(m.Version, currentVersion):
		result.Mode = ModeOptional
	default:
		result.Mode = ModeNone
	}

	return result
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ -v`
Expected: PASS, all tests including Task 5's.

- [ ] **Step 6: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/internal/updater/manifest.go src/internal/updater/updater.go src/internal/updater/manifest_test.go
git commit -m "feat: fetch client release manifest and decide update mode"
```

---

## Task 7: Release notes rendering

**Files:**
- Create: `farental-tui/src/internal/updater/notes.go`
- Test: `farental-tui/src/internal/updater/notes_test.go`

**Interfaces:**
- Consumes: `Block`, `Item`, `Span` (Task 6).
- Produces: `func RenderNotes(blocks []Block, width int) []string`

Wrapping uses `ansi.Wordwrap` from `github.com/charmbracelet/x/ansi`, already a direct dependency. Plain `lipgloss` width math counts a styled span's escape sequences as visible characters and wraps short.

Tests compare `ansi.Strip(line)` so they assert on layout, not on the active theme's escape codes.

- [ ] **Step 1: Write the failing test**

Create `farental-tui/src/internal/updater/notes_test.go`:

```go
package updater

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/orvyn"
)

func stripAll(lines []string) []string {
	out := make([]string, len(lines))

	for i, l := range lines {
		out[i] = strings.TrimRight(ansi.Strip(l), " ")
	}

	return out
}

func TestMain(m *testing.M) {
	// RenderNotes reads styles from the active theme, so orvyn must be
	// initialized. This needs no terminal.
	orvyn.Init()
	m.Run()
}

func TestRenderNotesHeadingGetsRule(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "h2", Spans: []Span{{Text: "Fixes"}}},
	}, 40))

	if len(got) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(got), got)
	}

	if got[0] != "Fixes" {
		t.Errorf("title line = %q, want %q", got[0], "Fixes")
	}

	if got[1] != strings.Repeat("─", 5) {
		t.Errorf("rule = %q, want 5 rule characters", got[1])
	}
}

func TestRenderNotesH3HasNoRule(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "h3", Spans: []Span{{Text: "Details"}}},
	}, 40))

	if len(got) != 1 || got[0] != "Details" {
		t.Errorf("got %q, want a single %q line", got, "Details")
	}
}

func TestRenderNotesWrapsStyledParagraph(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "p", Spans: []Span{
			{Text: "Fixed the "},
			{Text: "fight timer", Bold: true},
			{Text: " not being initialized correctly."},
		}},
	}, 20))

	if len(got) < 2 {
		t.Fatalf("expected the text to wrap, got %q", got)
	}

	for _, line := range got {
		if len(line) > 20 {
			t.Errorf("line %q exceeds width 20", line)
		}
	}

	joined := strings.Join(got, " ")

	if !strings.Contains(joined, "fight timer") {
		t.Errorf("bold text lost: %q", joined)
	}
}

func TestRenderNotesBulletsAndIndent(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "list", Items: []Item{
			{Indent: 0, Spans: []Span{{Text: "top"}}},
			{Indent: 1, Spans: []Span{{Text: "nested"}}},
		}},
	}, 40))

	want := []string{"• top", "  • nested"}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRenderNotesOrderedListNumbers(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "list", Ordered: true, Items: []Item{
			{Indent: 0, Spans: []Span{{Text: "first"}}},
			{Indent: 0, Spans: []Span{{Text: "second"}}},
		}},
	}, 40))

	if got[0] != "1. first" || got[1] != "2. second" {
		t.Errorf("got %q, want numbered entries", got)
	}
}

// A wrapped list item must align under its text, not under its marker.
func TestRenderNotesListContinuationAligns(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "list", Items: []Item{
			{Indent: 0, Spans: []Span{{Text: "a fairly long bullet that wraps"}}},
		}},
	}, 20))

	if len(got) < 2 {
		t.Fatalf("expected wrapping, got %q", got)
	}

	if !strings.HasPrefix(got[0], "• ") {
		t.Errorf("first line = %q, want a bullet marker", got[0])
	}

	if !strings.HasPrefix(got[1], "  ") || strings.HasPrefix(strings.TrimSpace(got[1]), "•") {
		t.Errorf("continuation = %q, want alignment under the text", got[1])
	}
}

func TestRenderNotesQuotePrefix(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "quote", Spans: []Span{{Text: "re-save your scripts"}}},
	}, 40))

	if !strings.HasPrefix(got[0], "│ ") {
		t.Errorf("got %q, want a quote prefix", got[0])
	}
}

func TestRenderNotesCodeIsVerbatim(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "code", Lines: []string{"go build ./...", "go vet ./..."}},
	}, 10))

	if got[0] != "  go build ./..." || got[1] != "  go vet ./..." {
		t.Errorf("got %q, want indented verbatim lines despite width 10", got)
	}
}

func TestRenderNotesLinkShowsURL(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "p", Spans: []Span{{Text: "the site", Href: "https://farental.ch"}}},
	}, 60))

	if !strings.Contains(got[0], "the site (https://farental.ch)") {
		t.Errorf("got %q, want the URL beside the text", got[0])
	}
}

func TestRenderNotesBlankLineBetweenBlocks(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "p", Spans: []Span{{Text: "one"}}},
		{Type: "p", Spans: []Span{{Text: "two"}}},
	}, 40))

	if len(got) != 3 || got[1] != "" {
		t.Errorf("got %q, want a blank line between blocks", got)
	}
}

func TestRenderNotesEmpty(t *testing.T) {
	if got := RenderNotes(nil, 40); got != nil {
		t.Errorf("got %q, want nil", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ -run TestRenderNotes -v`
Expected: FAIL — `undefined: RenderNotes`

- [ ] **Step 3: Write the implementation**

Create `farental-tui/src/internal/updater/notes.go`:

```go
package updater

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

// minNotesWidth keeps rendering sane on a very narrow terminal.
const minNotesWidth = 20

// listIndentWidth is the number of spaces one nesting level adds.
const listIndentWidth = 2

// RenderNotes turns release note blocks into styled terminal lines wrapped to
// width. The server decides what the notes say; this decides how they look,
// because only the client knows the width and the active theme.
func RenderNotes(blocks []Block, width int) []string {
	if len(blocks) == 0 {
		return nil
	}

	if width < minNotesWidth {
		width = minNotesWidth
	}

	t := orvyn.GetTheme()
	titleStyle := t.Style(theme.TitleStyleID)
	highlightStyle := t.Style(theme.HighlightTextStyleID)
	dimStyle := t.Style(theme.DimTextStyleID)

	var lines []string

	for i, b := range blocks {
		if i > 0 {
			lines = append(lines, "")
		}

		switch b.Type {
		case "h2":
			text := spansText(b.Spans)
			lines = append(lines, titleStyle.Render(text))
			lines = append(lines, titleStyle.Render(strings.Repeat("─", lipgloss.Width(text))))

		case "h3":
			lines = append(lines, highlightStyle.Render(spansText(b.Spans)))

		case "quote":
			for _, l := range wrapStyled(renderSpans(b.Spans), width-2) {
				lines = append(lines, dimStyle.Render("│ ")+l)
			}

		case "code":
			for _, l := range b.Lines {
				lines = append(lines, dimStyle.Render("  "+l))
			}

		case "list":
			lines = append(lines, renderList(b, width)...)

		default:
			lines = append(lines, wrapStyled(renderSpans(b.Spans), width)...)
		}
	}

	return lines
}

func renderList(b Block, width int) []string {
	var lines []string

	number := 0

	for _, item := range b.Items {
		number++

		indent := strings.Repeat(" ", item.Indent*listIndentWidth)

		marker := "• "

		if b.Ordered {
			marker = fmt.Sprintf("%d. ", number)
		}

		first := indent + marker
		rest := indent + strings.Repeat(" ", lipgloss.Width(marker))

		wrapped := wrapStyled(renderSpans(item.Spans), width-lipgloss.Width(first))

		for i, l := range wrapped {
			if i == 0 {
				lines = append(lines, first+l)
				continue
			}

			lines = append(lines, rest+l)
		}
	}

	return lines
}

// renderSpans applies inline marks. A link becomes "text (url)" because
// terminal hyperlink escapes are not portable enough to rely on here.
func renderSpans(spans []Span) string {
	var b strings.Builder

	for _, s := range spans {
		text := s.Text

		if s.Href != "" {
			text = fmt.Sprintf("%s (%s)", text, s.Href)
		}

		style := lipgloss.NewStyle().
			Bold(s.Bold).
			Italic(s.Italic).
			Underline(s.Underline).
			Strikethrough(s.Strike)

		b.WriteString(style.Render(text))
	}

	return b.String()
}

func spansText(spans []Span) string {
	var b strings.Builder

	for _, s := range spans {
		b.WriteString(s.Text)
	}

	return b.String()
}

// wrapStyled wraps text that already contains style escapes. ansi.Wordwrap
// measures visible width; lipgloss width math would count the escapes.
func wrapStyled(s string, width int) []string {
	if width < 1 {
		width = 1
	}

	return strings.Split(ansi.Wordwrap(s, width, ""), "\n")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ -run TestRenderNotes -v`
Expected: PASS, 11 tests.

If `TestRenderNotesHeadingGetsRule` fails on the rule length, the cause is `lipgloss.Width` versus rune count on the *unstyled* title text — `spansText` deliberately returns the raw text so the rule matches the visible title.

- [ ] **Step 5: Run the whole package**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ && go vet ./... && gofmt -l .`
Expected: `ok`, then no output.

- [ ] **Step 6: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/internal/updater/notes.go src/internal/updater/notes_test.go
git commit -m "feat: render release notes as styled terminal lines"
```

---

## Task 8: Download, verify, swap, restart

**Files:**
- Create: `farental-tui/src/internal/updater/apply.go`
- Create: `farental-tui/src/internal/updater/restart_unix.go`
- Create: `farental-tui/src/internal/updater/restart_windows.go`
- Test: `farental-tui/src/internal/updater/apply_test.go`

**Interfaces:**
- Consumes: `FileInfo`, `Result`, `RestartPending` (Task 6).
- Produces:
  ```go
  func ExecutablePath() (string, error)
  func PreflightWritable() error
  func Apply(baseURL string, f FileInfo, progress *atomic.Int64) error
  func CleanupOld()
  func Restart() error
  ```
  `progress` receives cumulative downloaded bytes; pass a non-nil pointer.

- [ ] **Step 1: Add the dependency**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go get github.com/minio/selfupdate@latest && go mod tidy`
Expected: `go.mod` gains `github.com/minio/selfupdate`.

- [ ] **Step 2: Write the failing test**

Create `farental-tui/src/internal/updater/apply_test.go`:

```go
package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func binaryServer(t *testing.T, payload []byte) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))

	t.Cleanup(srv.Close)

	return srv
}

func sum(payload []byte) string {
	h := sha256.Sum256(payload)
	return hex.EncodeToString(h[:])
}

// applyTo installs into a temp file rather than the test binary itself.
func applyTo(t *testing.T, target, baseURL string, f FileInfo, counter *atomic.Int64) error {
	t.Helper()

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = target

	return Apply(baseURL, f, counter)
}

func TestApplyReplacesTargetAndReportsProgress(t *testing.T) {
	payload := []byte("new binary contents")
	srv := binaryServer(t, payload)

	target := filepath.Join(t.TempDir(), "Farental")

	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	var counter atomic.Int64

	err := applyTo(t, target, srv.URL, FileInfo{
		Filename:  "Farental",
		SizeBytes: int64(len(payload)),
		SHA256:    sum(payload),
		URL:       "/clienttui/download/42",
	}, &counter)

	if err != nil {
		t.Fatalf("Apply returned %v", err)
	}

	got, err := os.ReadFile(target)

	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(payload) {
		t.Errorf("target contents = %q, want %q", got, payload)
	}

	if counter.Load() != int64(len(payload)) {
		t.Errorf("progress = %d, want %d", counter.Load(), len(payload))
	}
}

func TestApplyRejectsChecksumMismatch(t *testing.T) {
	srv := binaryServer(t, []byte("tampered"))

	target := filepath.Join(t.TempDir(), "Farental")

	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	var counter atomic.Int64

	err := applyTo(t, target, srv.URL, FileInfo{
		SizeBytes: int64(len("tampered")),
		SHA256:    sum([]byte("expected something else")),
		URL:       "/clienttui/download/42",
	}, &counter)

	if err == nil {
		t.Fatal("expected a checksum error, got nil")
	}

	got, _ := os.ReadFile(target)

	if string(got) != "old" {
		t.Errorf("target was modified despite the bad checksum: %q", got)
	}
}

func TestPreflightWritableRejectsUnwritableDir(t *testing.T) {
	dir := t.TempDir()

	if err := os.Chmod(dir, 0o500); err != nil {
		t.Skip("cannot make the directory read-only here")
	}

	t.Cleanup(func() { os.Chmod(dir, 0o700) })

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = filepath.Join(dir, "Farental")

	if err := PreflightWritable(); err == nil {
		t.Error("expected an error for a read-only directory, got nil")
	}
}

func TestPreflightWritableLeavesNoFiles(t *testing.T) {
	dir := t.TempDir()

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = filepath.Join(dir, "Farental")

	if err := PreflightWritable(); err != nil {
		t.Fatalf("PreflightWritable returned %v", err)
	}

	entries, err := os.ReadDir(dir)

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Errorf("left %d files behind, want none", len(entries))
	}
}

func TestCleanupOldRemovesLeftover(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Farental")

	if err := os.WriteFile(target+".old", []byte("stale"), 0o755); err != nil {
		t.Fatal(err)
	}

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = target

	CleanupOld()

	if _, err := os.Stat(target + ".old"); !os.IsNotExist(err) {
		t.Errorf("leftover .old still present (err = %v)", err)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ -run 'TestApply|TestPreflight|TestCleanup' -v`
Expected: FAIL — `undefined: targetPathOverride`

- [ ] **Step 4: Write the implementation**

Create `farental-tui/src/internal/updater/apply.go`:

```go
package updater

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/minio/selfupdate"
)

// downloadTimeout bounds the whole transfer, not just the connection.
const downloadTimeout = 10 * time.Minute

// targetPathOverride replaces the running executable's path in tests. Empty in
// production, where os.Executable is used.
var targetPathOverride string

// ExecutablePath returns the path of the binary to replace, with symlinks
// resolved so the real file is swapped rather than a link to it.
func ExecutablePath() (string, error) {
	if targetPathOverride != "" {
		return targetPathOverride, nil
	}

	exe, err := os.Executable()

	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(exe)
}

// PreflightWritable reports whether the update can be installed at all. The
// swap renames a directory entry, so the *directory* must be writable, not the
// binary. Checked before downloading, so a user who cannot install never waits
// for the transfer.
func PreflightWritable() error {
	exe, err := ExecutablePath()

	if err != nil {
		return err
	}

	dir := filepath.Dir(exe)

	probe, err := os.CreateTemp(dir, ".farental-update-*")

	if err != nil {
		return fmt.Errorf("%s: %w", dir, err)
	}

	name := probe.Name()

	probe.Close()

	return os.Remove(name)
}

// progressReader counts bytes as they are read.
type progressReader struct {
	inner   io.Reader
	counter *atomic.Int64
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.inner.Read(b)

	if n > 0 && p.counter != nil {
		p.counter.Add(int64(n))
	}

	return n, err
}

// Apply downloads the release binary and swaps it in. selfupdate.Apply writes
// the new file beside the target, verifies the checksum, performs the rename
// dance for the platform, and rolls back if the swap fails partway.
func Apply(baseURL string, f FileInfo, progress *atomic.Int64) error {
	exe, err := ExecutablePath()

	if err != nil {
		return err
	}

	checksum, err := hex.DecodeString(f.SHA256)

	if err != nil {
		return fmt.Errorf("invalid checksum in manifest: %w", err)
	}

	endpoint, err := url.JoinPath(strings.TrimSuffix(baseURL, "/"), f.URL)

	if err != nil {
		return err
	}

	client := &http.Client{Timeout: downloadTimeout}

	resp, err := client.Get(endpoint)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// The bound is the exact advertised size: a longer body is cut short and
	// fails the checksum, a shorter one fails it too.
	body := &progressReader{
		inner:   io.LimitReader(resp.Body, f.SizeBytes),
		counter: progress,
	}

	return selfupdate.Apply(body, selfupdate.Options{
		TargetPath: exe,
		Checksum:   checksum,
	})
}

// CleanupOld removes the previous binary left beside the target. Windows
// cannot delete a running executable, so the swap renames it aside and the
// next launch clears it. Harmless elsewhere.
func CleanupOld() {
	exe, err := ExecutablePath()

	if err != nil {
		return
	}

	os.Remove(exe + ".old")
}
```

- [ ] **Step 5: Write the restart implementations**

Create `farental-tui/src/internal/updater/restart_unix.go`:

```go
//go:build !windows

package updater

import (
	"os"
	"syscall"
)

// Restart replaces the current process image with the new binary. The PID is
// kept and no second process appears, so the shell sees one continuous
// program. Must be called only after bubbletea has restored the terminal.
func Restart() error {
	exe, err := ExecutablePath()

	if err != nil {
		return err
	}

	return syscall.Exec(exe, os.Args, os.Environ())
}
```

Create `farental-tui/src/internal/updater/restart_windows.go`:

```go
//go:build windows

package updater

import (
	"os"
	"os/exec"
)

// Restart launches the new binary and lets this process exit. Windows has no
// exec that replaces a process image, and the running executable has already
// been renamed aside by the swap.
func Restart() error {
	exe, err := ExecutablePath()

	if err != nil {
		return err
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./internal/updater/ -v`
Expected: PASS, all tests.

Note: `TestPreflightWritableRejectsUnwritableDir` skips itself when run as root, since root ignores directory permissions.

- [ ] **Step 7: Verify the Windows build compiles**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && GOOS=windows GOARCH=amd64 go build ./... && GOOS=darwin GOARCH=arm64 go build ./...`
Expected: no output. This is the only check that the build-tagged files are correct — the tests only exercise the host platform.

- [ ] **Step 8: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/internal/updater/apply.go src/internal/updater/restart_unix.go src/internal/updater/restart_windows.go src/internal/updater/apply_test.go src/go.mod src/go.sum
git commit -m "feat: download, verify and swap the client binary"
```

---

## Task 9: The update screen

**Files:**
- Create: `farental-tui/src/screen/clientupdate/clientupdate.go`
- Modify: `farental-tui/src/screen/screen.go:6-33`
- Modify: `farental-tui/src/internal/keybind/context.go`
- Modify: `farental-tui/src/translations/{en,fr,de}.json`

**Interfaces:**
- Consumes: `updater.Pending`, `updater.Result`, `updater.Mode*`, `updater.RenderNotes`, `updater.Apply`, `updater.PreflightWritable`, `updater.RestartPending`, `config.WebURL`.
- Produces: `clientupdate.New() *Screen`, `screen.IDClientUpdate`, `keybind.ContextClientUpdate`.

- [ ] **Step 1: Add the screen ID**

In `screen/screen.go`, add to the const block:

```go
	IDClientUpdate          orvyn.ScreenID = "clientupdate"
```

- [ ] **Step 2: Add the help context**

In `internal/keybind/context.go`, add to the context const block:

```go
	ContextClientUpdate                        bubblehelp.KeymapContext = "clientUpdate"
```

and inside `InitContexts()`:

```go
	clientUpdateKeymap := bubblehelp.NewKeymap(2)
	clientUpdateKeymap.Style = mainHelpStyle
	clientUpdateKeymap.NewKeyBinding(Enter, true)
	clientUpdateKeymap.SetHelpDesc(Enter, lokyn.L("update now"))
	clientUpdateKeymap.NewKeyBinding(Esc, true)
	clientUpdateKeymap.SetHelpDesc(Esc, lokyn.L("skip"))
	clientUpdateKeymap.NewKeyBinding(Quit, true)

	bubblehelp.RegisterContext(ContextClientUpdate, clientUpdateKeymap)
```

- [ ] **Step 3: Write the screen**

Create `farental-tui/src/screen/clientupdate/clientupdate.go`:

```go
package clientupdate

import (
	"farental/internal/config"
	"farental/internal/keybind"
	"farental/internal/updater"
	ftheme "farental/internal/theme"
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
```

Import `"farental/screen"` alongside the others and use `screen.IDLogin` directly — there is no import cycle, because `farental/screen` imports nothing but orvyn, and every sibling screen does the same (`screen/login/login.go:12`).

- [ ] **Step 4: Add the translation strings**

Add these keys to `translations/en.json`, `fr.json`, and `de.json`. In `en.json` the value equals the key. Suggested translations:

| key | fr | de |
| --- | --- | --- |
| `A new version is available` | `Une nouvelle version est disponible` | `Eine neue Version ist verfügbar` |
| `release notes` | `notes de version` | `Versionshinweise` |
| `update now` | `mettre à jour` | `jetzt aktualisieren` |
| `skip` | `ignorer` | `überspringen` |
| `Downloading...` | `Téléchargement...` | `Wird heruntergeladen...` |
| `Verifying and installing...` | `Vérification et installation...` | `Wird geprüft und installiert...` |
| `Restarting...` | `Redémarrage...` | `Neustart...` |
| `The update failed.` | `La mise à jour a échoué.` | `Die Aktualisierung ist fehlgeschlagen.` |
| `Update failed` | `Échec de la mise à jour` | `Aktualisierung fehlgeschlagen` |
| `Press r to retry.` | `Appuyez sur r pour réessayer.` | `Drücken Sie r, um es erneut zu versuchen.` |
| `Update required` | `Mise à jour requise` | `Aktualisierung erforderlich` |
| `No build is published for your platform.` | `Aucune version n'est publiée pour votre plateforme.` | `Für Ihre Plattform ist keine Version veröffentlicht.` |
| `Farental cannot write to its own directory.` | `Farental ne peut pas écrire dans son propre dossier.` | `Farental kann nicht in sein eigenes Verzeichnis schreiben.` |
| `Could not reach the update server.` | `Impossible de contacter le serveur de mise à jour.` | `Der Update-Server ist nicht erreichbar.` |
| `Download the new version here:` | `Téléchargez la nouvelle version ici :` | `Laden Sie die neue Version hier herunter:` |

- [ ] **Step 5: Verify it compiles on every target**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go build ./... && GOOS=windows GOARCH=amd64 go build ./... && go vet ./... && gofmt -l .`
Expected: no output.

`SetActive`/`IsActive` reach both widgets through `orvyn.BaseWidget` → `BaseRenderable` → `BaseActivable` (`orvyn/activable.go:18`), and layouts skip inactive elements (`orvyn/layout.go:34`) — that is how a widget is hidden here.

- [ ] **Step 6: Verify the translation files are valid JSON**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && for f in translations/*.json; do python3 -m json.tool "$f" > /dev/null && echo "$f ok"; done`
Expected: three `ok` lines.

- [ ] **Step 7: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/clientupdate src/screen/screen.go src/internal/keybind/context.go src/translations
git commit -m "feat: add client update screen"
```

---

## Task 10: Startup wiring

The version check currently runs before any UI exists (`main.go:110-124`) and exits on mismatch. It moves after screen registration so the update screen can be the first screen shown.

**Files:**
- Modify: `farental-tui/src/main.go:96-190`
- Modify: `farental-tui/src/app.go:13-17`

**Interfaces:**
- Consumes: everything from Tasks 5–9.
- Produces: a client that shows `IDClientUpdate` when an update is pending and `IDLogin` otherwise, and execs the new binary after bubbletea exits.

- [ ] **Step 1: Make the start screen configurable**

In `app.go`:

```go
// App is the main model to run the Orvyn application
type App struct {
	StartScreen orvyn.ScreenID
}

func (a App) Init() tea.Cmd {
	return orvyn.SwitchScreen(a.StartScreen)
}
```

- [ ] **Step 2: Rewrite the startup sequence**

In `main.go`, replace the version-check block:

```go
	// Check version
	reqVer := request.VersionGet()
	version, err := helper.Fetch[api.DbVersion](reqVer)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if !strings.HasPrefix(config.VERSION, version.ClientTui) {
		fmt.Println(lokyn.L("Your client version is not aligned with the server. Please update it."))
		fmt.Println(lokyn.L("Visit https://www.farental.ch for more information."))
		return
	}
```

with:

```go
	// Clear a binary left behind by a previous update. Windows cannot delete
	// the running executable during the swap, so it lands here.
	updater.CleanupOld()

	// Check version
	reqVer := request.VersionGet()
	version, err := helper.Fetch[api.DbVersion](reqVer)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	updater.Pending = updater.Check(config.VERSION, version.ClientTui,
		viper.GetString("language"))
```

Then register the screen alongside the others:

```go
	orvyn.RegisterScreen(screen.IDClientUpdate, clientupdate.New())
```

And replace the program startup:

```go
	context.ResetTerminalTitle()

	startScreen := screen.IDLogin

	if updater.Pending.Mode != updater.ModeNone {
		startScreen = screen.IDClientUpdate
	}

	p := tea.NewProgram(&App{StartScreen: startScreen}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// Run after bubbletea has left the alt screen and restored the cursor, so
	// the replacement process inherits a sane terminal.
	if updater.RestartPending {
		if err := updater.Restart(); err != nil {
			fmt.Println(lokyn.L("Update installed. Please start Farental again."))
		}
	}
```

Add `"farental/internal/updater"` and `"farental/screen/clientupdate"` to the imports, and drop `"strings"` if nothing else in the file uses it (the compiler will say).

- [ ] **Step 3: Add the remaining translation string**

Add `Update installed. Please start Farental again.` to all three translation files:

- fr: `Mise à jour installée. Veuillez relancer Farental.`
- de: `Aktualisierung installiert. Bitte starten Sie Farental erneut.`

- [ ] **Step 4: Verify**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go build ./... && GOOS=windows GOARCH=amd64 go build ./... && go vet ./... && gofmt -l . && go test ./...`
Expected: no output from build/vet/gofmt, tests `ok`.

- [ ] **Step 5: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/main.go src/app.go src/translations
git commit -m "feat: run the update check at startup"
```

---

## Task 11: Release process documentation

**Files:**
- Modify: `farental-cli/docs/VERSION_MANAGEMENT.md`

**Interfaces:**
- Consumes: the behavior built in Tasks 1–10.
- Produces: nothing code depends on.

- [ ] **Step 1: Update the client release checklist**

In the `#### Client` section, replace:

```
- [ ] Create a release on github with the tag and join all client executables
- [ ] Upload the clients on the FTP for download from the website
```

with:

```
- [ ] Upload all five platform binaries via /admin/releases/new
      (windows-amd64, linux-amd64, linux-arm64, darwin-amd64, darwin-arm64)
- [ ] Write the release notes for each language in the same form
- [ ] Publish the release
```

- [ ] **Step 2: Document what publishing now means**

Add this after the Client checklist:

```markdown
### Publishing pushes the update

Setting `IsPublished` on a release is the point of no return: `/clienttui/latest.json`
starts advertising it immediately, and every client that starts up from then on
offers or forces the update. Uploading is safe; publishing is the switch.

A client whose Major.Minor does not match `db_version.client_tui` cannot reach the
login screen at all — it can only update or quit. Publish a release for every
platform before bumping `client_tui`, or users on an unpublished platform are
locked out with only the manual download page to fall back on.
```

- [ ] **Step 3: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git add docs/VERSION_MANAGEMENT.md
git commit -m "docs: update client release process for in-client updates"
```

---

## Manual verification (after Task 11)

These cannot be automated here — they need a real terminal and a running server.

- [ ] Run the server locally (`WEB_PORT=3001`), upload a build as version `1.1.1`, publish it.
- [ ] Build the client with `config.VERSION` at `1.1.0` and `WebURL=http://127.0.0.1:3001`.
- [ ] Start it: the update screen appears with the release notes formatted (heading rule, indented bullets, quote bar).
- [ ] Press enter: the progress bar advances, then the client relaunches on `1.1.1` — confirm with the version shown at login.
- [ ] Set `db_version.client_tui` to `1.2` and confirm a `1.1.x` client cannot reach login by pressing esc.
- [ ] Repeat the update on Windows, which uses the other swap and restart path, and confirm no `Farental.exe.old` survives the next launch.
- [ ] Install the binary into a root-owned directory and confirm the manual-download screen appears instead of a crash.
