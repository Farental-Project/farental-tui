package updater

import (
	"farental/core/request"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/go-resty/resty/v2"
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

	// fetchManifest and Apply build requests through core/request rather
	// than raw net/http, so the package-level web client must be wired the
	// way main.go wires it in production (request.InitWeb(context.Web)).
	// The client's own base URL doesn't matter here: every test in this
	// package points fetchManifest/Apply at an httptest server via an
	// absolute endpoint URL, which resty uses as-is regardless of the
	// client's configured base — exactly how production points the same
	// code at config.WebURL.
	request.InitWeb(resty.New())

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

// Ordered numbering keeps a separate counter per indent level: nesting under
// an item starts that level at 1, and returning to the parent level resumes
// the parent's own sequence rather than continuing a single flat count.
func TestRenderNotesOrderedListNestedNumbersResetPerLevel(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "list", Ordered: true, Items: []Item{
			{Indent: 0, Spans: []Span{{Text: "first"}}},
			{Indent: 1, Spans: []Span{{Text: "sub-a"}}},
			{Indent: 1, Spans: []Span{{Text: "sub-b"}}},
			{Indent: 0, Spans: []Span{{Text: "second"}}},
		}},
	}, 40))

	want := []string{"1. first", "  1. sub-a", "  2. sub-b", "2. second"}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// A second nested run, after returning to the shallower level in between,
// must restart its own counter at 1 rather than continuing from the first
// nested run's count.
func TestRenderNotesOrderedListNestedNumbersRestartAfterReturn(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "list", Ordered: true, Items: []Item{
			{Indent: 0, Spans: []Span{{Text: "first"}}},
			{Indent: 1, Spans: []Span{{Text: "sub-a"}}},
			{Indent: 1, Spans: []Span{{Text: "sub-b"}}},
			{Indent: 0, Spans: []Span{{Text: "second"}}},
			{Indent: 1, Spans: []Span{{Text: "sub-c"}}},
		}},
	}, 40))

	want := []string{
		"1. first",
		"  1. sub-a",
		"  2. sub-b",
		"2. second",
		"  1. sub-c",
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
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

// The manifest is server-supplied and untrusted: a negative Indent must not
// panic strings.Repeat or index a counters slice out of range.
func TestRenderNotesClampsNegativeIndent(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "list", Items: []Item{
			{Indent: -1, Spans: []Span{{Text: "top"}}},
		}},
	}, 40))

	if len(got) != 1 || got[0] != "• top" {
		t.Errorf("got %q, want a single unindented bullet", got)
	}
}

// An ordered list with a negative Indent must not panic either, since the
// counters slice is grown by "len(counters) <= indent" which never runs for
// a negative indent unless it is clamped first.
func TestRenderNotesClampsNegativeIndentOrdered(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "list", Ordered: true, Items: []Item{
			{Indent: -5, Spans: []Span{{Text: "first"}}},
		}},
	}, 40))

	if len(got) != 1 || got[0] != "1. first" {
		t.Errorf("got %q, want a single numbered entry", got)
	}
}

// A huge Indent must not allocate an unbounded number of spaces or grow the
// ordered counters slice without limit.
func TestRenderNotesClampsHugeIndent(t *testing.T) {
	got := stripAll(RenderNotes([]Block{
		{Type: "list", Items: []Item{
			{Indent: 1_000_000, Spans: []Span{{Text: "deep"}}},
		}},
	}, 60))

	want := strings.Repeat(" ", maxListIndent*listIndentWidth) + "• deep"

	if len(got) != 1 || got[0] != want {
		t.Errorf("got %q, want indent clamped to %d levels", got, maxListIndent)
	}
}
