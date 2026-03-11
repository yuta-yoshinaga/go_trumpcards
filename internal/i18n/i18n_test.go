package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestSetLang_Japanese(t *testing.T) {
	i18n.SetLang("ja")
	assert.Equal(t, "ja", i18n.Lang())
	// Japanese "bye" key should be present
	assert.Equal(t, "さようなら。", i18n.T("bye"))
}

func TestSetLang_English(t *testing.T) {
	i18n.SetLang("en")
	assert.Equal(t, "en", i18n.Lang())
	assert.Equal(t, "Goodbye.", i18n.T("bye"))
}

func TestSetLang_Unknown_DefaultsToJa(t *testing.T) {
	i18n.SetLang("fr")
	assert.Equal(t, "ja", i18n.Lang())
	assert.Equal(t, "さようなら。", i18n.T("bye"))
}

func TestSetLang_Empty_DefaultsToJa(t *testing.T) {
	i18n.SetLang("")
	assert.Equal(t, "ja", i18n.Lang())
}

func TestT_KeyNotFound_ReturnsKey(t *testing.T) {
	i18n.SetLang("ja")
	result := i18n.T("nonexistent.key.xyz")
	assert.Equal(t, "nonexistent.key.xyz", result)
}

func TestT_KeyFound_ReturnsValue(t *testing.T) {
	i18n.SetLang("ja")
	result := i18n.T("unknownCommand")
	assert.Equal(t, "コマンドが不明です: {{cmd}}", result)
}

func TestTf_WithParams(t *testing.T) {
	i18n.SetLang("ja")
	result := i18n.Tf("unknownCommand", "cmd", "foo")
	assert.Equal(t, "コマンドが不明です: foo", result)
}

func TestTf_WithMultipleParams(t *testing.T) {
	i18n.SetLang("ja")
	result := i18n.Tf("switchedTo", "name", "poker")
	assert.Equal(t, "Switched to poker.", result)
}

func TestTf_WithNoParams_ReturnsT(t *testing.T) {
	i18n.SetLang("ja")
	// Odd number of params (last one ignored)
	result := i18n.Tf("bye")
	assert.Equal(t, "さようなら。", result)
}

func TestTf_OddParams_LastIgnored(t *testing.T) {
	i18n.SetLang("ja")
	// Even loop: i+1 < len(params) → with 1 param, loop never runs
	result := i18n.Tf("unknownCommand", "cmd")
	// No substitution since params count is odd (only 1 element)
	assert.Equal(t, "コマンドが不明です: {{cmd}}", result)
}

func TestQuitSentinel(t *testing.T) {
	assert.Equal(t, "bye.", i18n.QuitSentinel)
}

func TestT_English_unknownCommand(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.T("unknownCommand")
	assert.Equal(t, "Unknown command: {{cmd}}", result)
}

func TestTf_English_unknownCommand(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.Tf("unknownCommand", "cmd", "xyz")
	assert.Equal(t, "Unknown command: xyz", result)
}

func TestT_GameSpecific_ja(t *testing.T) {
	i18n.SetLang("ja")
	// doubt-specific key
	assert.Contains(t, i18n.T("doubtPrompt"), "ダウト")
}

func TestT_GameSpecific_en(t *testing.T) {
	i18n.SetLang("en")
	assert.Equal(t, "Timeout: skipping doubt", i18n.T("timeout"))
}

func TestT_HoldemKeys_ja(t *testing.T) {
	i18n.SetLang("ja")
	assert.Contains(t, i18n.T("amountRequired"), "ベット")
}

func TestT_HoldemKeys_en(t *testing.T) {
	i18n.SetLang("en")
	assert.Equal(t, "Bet/raise requires an amount.", i18n.T("amountRequired"))
}

func TestTf_InteractiveMode_en(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.Tf("interactiveMode", "name", "blackjack")
	assert.Equal(t, "Interactive mode  (current: blackjack)", result)
}

func TestTf_UnknownGame_en(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.Tf("unknownGame", "name", "chess")
	assert.Equal(t, "Unknown game: chess. Type 'games' for the list.", result)
}

func TestTf_AlreadyPlaying_en(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.Tf("alreadyPlaying", "name", "poker")
	assert.Equal(t, "Already playing poker.", result)
}

// Reset to ja after all tests to avoid polluting other packages
func TestSetLang_ResetToJa(t *testing.T) {
	i18n.SetLang("ja")
	assert.Equal(t, "ja", i18n.Lang())
}
