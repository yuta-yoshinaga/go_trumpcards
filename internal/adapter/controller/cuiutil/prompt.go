package cuiutil

import "strings"

// PromptPrefix is the sentinel prefix indicating a prompt request.
// When a controller returns a string starting with this prefix,
// the game loop should prompt the user for the missing argument
// and re-dispatch the command with the user's input.
const PromptPrefix = "PROMPT:"

// promptSep separates the prompt message from the command template.
const promptSep = "\t"

// PromptRequest builds a prompt return string.
// promptMsg is displayed to the user (e.g., "Enter bet amount:").
// commandTemplate is the command to re-dispatch, with {0} as a placeholder
// for the user's input (e.g., "b {0}").
func PromptRequest(promptMsg, commandTemplate string) string {
	return PromptPrefix + promptMsg + promptSep + commandTemplate
}

// IsPromptRequest checks if s starts with PromptPrefix.
func IsPromptRequest(s string) bool {
	return strings.HasPrefix(s, PromptPrefix)
}

// ParsePromptRequest extracts the prompt message and command template
// from a prompt request string. Returns empty strings if malformed.
func ParsePromptRequest(s string) (prompt, template string) {
	if !IsPromptRequest(s) {
		return "", ""
	}
	body := s[len(PromptPrefix):]
	before, after, found := strings.Cut(body, promptSep)
	if !found {
		return before, ""
	}
	return before, after
}

// FillTemplate replaces the {0} placeholder in template with value.
func FillTemplate(template, value string) string {
	return strings.Replace(template, "{0}", value, 1)
}
