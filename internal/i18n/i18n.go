package i18n

import (
	"embed"
	"encoding/json"
	"io/fs"
	"strings"
)

//go:embed locales/ja
//go:embed locales/en
var localesFS embed.FS

// QuitSentinel is the internal protocol value returned by controllers on quit.
const QuitSentinel = "bye."

// ErrorPrefix marks a CUI result as an error to be routed to stderr.
// The prefix itself is stripped before display. Using ASCII Record Separator (0x1E)
// keeps it invisible to users and unlikely to collide with any real content.
const ErrorPrefix = "\x1eERR\x1e"

// MarkError prefixes msg so the CUI runner routes it to stderr.
// Returns msg unchanged if it is empty or already marked.
func MarkError(msg string) string {
	if msg == "" || strings.HasPrefix(msg, ErrorPrefix) {
		return msg
	}
	return ErrorPrefix + msg
}

// StripErrorPrefix removes ErrorPrefix from msg and reports whether it was present.
func StripErrorPrefix(msg string) (body string, isError bool) {
	if strings.HasPrefix(msg, ErrorPrefix) {
		return msg[len(ErrorPrefix):], true
	}
	return msg, false
}

var currentLang = "ja"
var translations = map[string]string{}

// SetLang sets the active language. Only "ja" and "en" are supported; anything else defaults to "ja".
func SetLang(lang string) {
	if lang != "ja" && lang != "en" {
		lang = "ja"
	}
	currentLang = lang
	translations = loadTranslations(localesFS, lang)
}

// Lang returns the currently active language.
func Lang() string { return currentLang }

// T returns the translation for key, falling back to the key itself if not found.
func T(key string) string {
	if v, ok := translations[key]; ok {
		return v
	}
	return key
}

// Tf returns the translation for key with {{param}} substitutions.
// params should be alternating key/value pairs: Tf("key", "name", "Alice", "age", "30")
func Tf(key string, params ...string) string {
	s := T(key)
	for i := 0; i+1 < len(params); i += 2 {
		s = strings.ReplaceAll(s, "{{"+params[i]+"}}", params[i+1])
	}
	return s
}

func loadTranslations(fsys fs.FS, lang string) map[string]string {
	result := map[string]string{}
	games := []string{"common", "blackjack", "poker", "oldmaid", "daifugo", "sevens", "doubt", "holdem", "omaha", "shortdeck", "hearts", "memory", "klondike", "freecell", "baccarat", "spades", "crazyeights", "ginrummy", "spider", "napoleon", "indianpoker", "videopoker", "euchre", "pyramid", "cribbage", "tripeaks", "threecard", "ohhell", "pineapple", "crazypineapple", "speed", "pigtail", "sevencardstud", "clocksolitaire", "twotenjack", "caribbeanstud", "war", "canfield", "fiftyone", "yukon", "whist", "pageone", "reddog", "razz", "badugi", "scorpion", "accordion", "trash", "spanish21"}
	for _, game := range games {
		path := "locales/" + lang + "/" + game + ".json"
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			continue
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		for k, v := range m {
			if game == "common" {
				result[k] = v
			} else {
				result[game+"."+k] = v
			}
		}
	}
	return result
}

func init() {
	// Pre-load Japanese translations so T/Tf work correctly even if SetLang
	// is never called (e.g. in tests that don't set a language).
	translations = loadTranslations(localesFS, "ja")
}
