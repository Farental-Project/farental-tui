package updater

import (
	"context"
	"encoding/json"
	"farental/core/request"
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

// maxManifestBodySize bounds the manifest response, symmetric with the
// binary download's size bound: a JSON document describing a release has no
// business being anywhere near this large.
const maxManifestBodySize = 1 * 1024 * 1024

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
	Type    string   `json:"type"`
	Spans   []Span   `json:"spans"`
	Ordered bool     `json:"ordered"`
	Items   []Item   `json:"items"`
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
	if err := requireSecureURL(baseURL); err != nil {
		return nil, err
	}

	endpoint, err := url.JoinPath(strings.TrimSuffix(baseURL, "/"), "clienttui", "latest")

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	// SetResponseBodyLimit bounds the read from the network itself — resty
	// aborts the request once it has read past the limit — rather than only
	// bounding the eventual json.Unmarshal. Symmetric with the binary
	// download's size bound: a hostile or misbehaving host cannot make this
	// buffer more than maxManifestBodySize before the request fails.
	req := request.ManifestGet(endpoint, lang).
		SetContext(ctx).
		SetResponseBodyLimit(maxManifestBodySize)

	resp, err := req.Send()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("manifest request failed with status %d", resp.StatusCode())
	}

	var m manifest

	if err := json.Unmarshal(resp.Body(), &m); err != nil {
		return nil, err
	}

	return &m, nil
}
