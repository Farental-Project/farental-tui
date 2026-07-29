// Package manual holds the in-app manuals, embedded as one plain text file per
// topic and language.
package manual

import (
	"embed"
	"fmt"
	"strings"

	"github.com/halsten-dev/lokyn"
)

//go:embed docs
var docs embed.FS

// defaultLang is the language used when the current one has no file for the
// requested topic.
const defaultLang = "en"

// Get returns the title and the body lines of the manual of the given topic,
// in the current language, falling back to the English version.
func Get(topic string) (string, []string, error) {
	data, err := docs.ReadFile(filePath(topic, baseLanguage()))

	if err != nil {
		data, err = docs.ReadFile(filePath(topic, defaultLang))
	}

	if err != nil {
		return "", nil, fmt.Errorf("no manual available for the topic %s", topic)
	}

	title, lines := split(string(data))

	return title, lines, nil
}

// filePath returns the embedded path of the manual of the given topic and
// language.
func filePath(topic, lang string) string {
	return fmt.Sprintf("docs/%s_%s.txt", topic, lang)
}

// baseLanguage returns the base tag of the current language, so that a
// regional locale like fr-FR resolves to the fr manual.
func baseLanguage() string {
	base, _, _ := strings.Cut(lokyn.GetCurrentLanguage(), "-")

	return base
}

// split separates the title from the body: the first non-empty line is the
// title, the rest is the body without its leading blank lines.
func split(content string) (string, []string) {
	var title string

	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")

	i := 0

	for ; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			title = strings.TrimSpace(lines[i])
			i++
			break
		}
	}

	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	return title, lines[i:]
}
