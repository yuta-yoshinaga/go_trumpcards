package presenter

import (
	"sort"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// translateMessageCode resolves a message code with its map parameters.
func translateMessageCode(code string, params map[string]string) string {
	if code == "" {
		return ""
	}
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	args := make([]string, 0, len(keys)*2)
	for _, key := range keys {
		args = append(args, key, params[key])
	}
	return i18n.Tf(code, args...)
}

func withMessageParams(params map[string]string, extra ...string) map[string]string {
	result := make(map[string]string, len(params)+len(extra)/2)
	for key, value := range params {
		result[key] = value
	}
	for i := 0; i+1 < len(extra); i += 2 {
		result[extra[i]] = extra[i+1]
	}
	return result
}
