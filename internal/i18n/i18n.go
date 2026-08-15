package i18n

import (
	"encoding/json"
	"io/fs"
	"strings"
)

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

// ErrorLinePrefix marks one line inside an otherwise ordinary reply as a
// rejection. The board games render a refusal as a red line inside the board,
// so the reply as a whole is a board and must stay on stdout -- marking it with
// ErrorPrefix would send the whole board to stderr in red. This says "there is
// a refusal in here" without claiming the reply is one.
const ErrorLinePrefix = "\x1eERRLN\x1e"

// MarkErrorLine marks a single line as a rejection. Returns line unchanged if
// it is empty or already marked.
func MarkErrorLine(line string) string {
	if line == "" || strings.HasPrefix(line, ErrorLinePrefix) {
		return line
	}
	return ErrorLinePrefix + line
}

// StripErrorLines removes every ErrorLinePrefix from msg and reports whether
// any was present. Callers display the result; the marker never reaches a user.
func StripErrorLines(msg string) (body string, hadError bool) {
	if !strings.Contains(msg, ErrorLinePrefix) {
		return msg, false
	}
	return strings.ReplaceAll(msg, ErrorLinePrefix, ""), true
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
	translations = loadLocale(lang)
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

// globalNamespaces lists locale files whose keys are merged into the
// translation map without a "<file>." prefix. Game files are namespaced
// instead (e.g. "blackjack.helpTitle") so distinct games can reuse the
// same short key. cui_common is shared by every CUI presenter and lives
// alongside common.json — its keys all begin with "cui*" so collisions
// with common.json are impossible.
var globalNamespaces = map[string]bool{
	"common":     true,
	"cui_common": true,
}

// loadTranslations reads every *.json file under locales/<lang>/ and merges
// the entries into a single map. Per-game files are namespaced as
// "<file>.<key>"; files in globalNamespaces (common, cui_common) are merged
// flat so their keys are reusable across the codebase.
func loadTranslations(fsys fs.FS, lang string) map[string]string {
	result := map[string]string{}
	entries, err := fs.ReadDir(fsys, "locales/"+lang)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		file := strings.TrimSuffix(name, ".json")
		data, err := fs.ReadFile(fsys, "locales/"+lang+"/"+name)
		if err != nil {
			continue
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		for k, v := range m {
			if globalNamespaces[file] {
				result[k] = v
			} else {
				result[file+"."+k] = v
			}
		}
	}
	return result
}

func init() {
	// Pre-load Japanese translations so T/Tf work correctly even if SetLang
	// is never called (e.g. in tests that don't set a language).
	translations = loadLocale("ja")
}
