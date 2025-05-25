package lang

import (
	"embed"
	"encoding/json"
	"github.com/jeandeaual/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"log"
	"sync"
)

var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	once      sync.Once

	currentLang language.Tag
	translated  []language.Tag
)

// Init inits the package, only once
func Init() {
	once.Do(initBundle)
}

func AddTranslationFS(fs embed.FS, dir string) error {
	files, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		name := f.Name()
		data, err := fs.ReadFile(dir + "/" + name)
		if err != nil {
			continue
		}

		err = addLanguage(data, name)
		if err != nil {
			continue
		}
	}

	initLanguage()

	return nil
}

func GetCurrentLanguage() string {
	return currentLang.String()
}

func SetLanguage(lang string) {
	setupLang(lang)
}

// L returns translation
func L(key string) string {
	return getKey(key, key)
}

// P returns translation with plural management
func P(key string, count int) string {
	return getPluralKey(key, key, count)
}

func initBundle() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	translated = []language.Tag{language.Make("en")}
}

func initLanguage() {
	all, err := locale.GetLocales()
	if err != nil {
		all = []string{"en"}
	}

	setupLang(closestSupportedLocale(all).String())
}

func setupLang(lang string) {
	currentLang = language.Make(lang)
	localizer = i18n.NewLocalizer(bundle, lang)
}

func addLanguage(data []byte, name string) error {
	f, err := bundle.ParseMessageFileBytes(data, name)
	if err != nil {
		return err
	}

	translated = append(translated, f.Tag)
	return nil
}

func getKey(key, fallback string) string {
	ret, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    key,
			Other: fallback,
		},
	})

	if err != nil {
		log.Println("Error in translation")
	}

	return ret
}

func getPluralKey(key, fallback string, count int) string {
	ret, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    key,
			Other: fallback,
		},
		PluralCount: count,
	})

	if err != nil {
		log.Println("Error in translation")
	}

	return ret
}

func closestSupportedLocale(locs []string) language.Tag {
	matcher := language.NewMatcher(translated)

	tags := make([]language.Tag, len(locs))
	for i, loc := range locs {
		tag, err := language.Parse(loc)
		if err != nil {
			log.Println("Error in parsing tags")
		}
		tags[i] = tag
	}
	best, _, _ := matcher.Match(tags...)
	return best
}
