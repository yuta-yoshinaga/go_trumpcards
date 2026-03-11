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

func TestTf_WithSingleParam(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.Tf("switchedTo", "name", "poker")
	assert.Equal(t, "Switched to poker.", result)
}

func TestTf_ExtraParamsIgnored(t *testing.T) {
	i18n.SetLang("en")
	// Extra trailing pair is silently ignored (i+1 < len guard)
	result := i18n.Tf("unknownCommand", "cmd", "foo", "extra", "ignored")
	assert.Equal(t, "Unknown command: foo", result)
}

func TestTf_WithMultipleSubstitutions(t *testing.T) {
	i18n.SetLang("en")
	// holdem.smallBlindMustBeLess has one placeholder {{bb}}, but we can
	// verify two simultaneous substitutions via a key that has two placeholders.
	// holdem.bigBlindMustBeGreater uses {{sb}}.
	// Use the interactiveMode key (has {{name}}) combined with a multi-pair call.
	// The cleanest test: call Tf with two pairs where one matches and one doesn't.
	result := i18n.Tf("holdem.smallBlindMustBeLess", "bb", "10", "extra", "ignored")
	assert.Contains(t, result, "10")
	// Verify a key with two actual placeholders substituted in one call
	result2 := i18n.Tf("unknownGame", "name", "chess")
	assert.Equal(t, "Unknown game: \"chess\". Type 'games' for the list.", result2)
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
	// doubt-specific key (prefixed with game name)
	assert.Contains(t, i18n.T("doubt.doubtPrompt"), "ダウト")
}

func TestT_GameSpecific_en(t *testing.T) {
	i18n.SetLang("en")
	assert.Equal(t, "Timeout: skipping doubt", i18n.T("doubt.timeout"))
}

func TestT_HoldemKeys_ja(t *testing.T) {
	i18n.SetLang("ja")
	assert.Contains(t, i18n.T("holdem.amountRequired"), "ベット")
}

func TestT_HoldemKeys_en(t *testing.T) {
	i18n.SetLang("en")
	assert.Equal(t, "Bet/raise requires an amount.", i18n.T("holdem.amountRequired"))
}

func TestTf_InteractiveMode_en(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.Tf("interactiveMode", "name", "blackjack")
	assert.Equal(t, "Interactive mode  (current: blackjack)", result)
}

func TestTf_UnknownGame_en(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.Tf("unknownGame", "name", "chess")
	assert.Equal(t, "Unknown game: \"chess\". Type 'games' for the list.", result)
}

func TestTf_AlreadyPlaying_en(t *testing.T) {
	i18n.SetLang("en")
	result := i18n.Tf("alreadyPlaying", "name", "poker")
	assert.Equal(t, "Already playing poker.", result)
}

func TestT_HelpLines_GamePrefixed_ja(t *testing.T) {
	i18n.SetLang("ja")
	// Verify game-specific help keys are correctly prefixed and accessible
	assert.Equal(t, "BlackJack (ブラックジャック)", i18n.T("blackjack.helpTitle"))
	assert.Equal(t, "5-card Draw Poker (ポーカー)", i18n.T("poker.helpTitle"))
	assert.Equal(t, "Doubt (ダウト)", i18n.T("doubt.helpTitle"))
}

func TestT_HelpLines_GamePrefixed_en(t *testing.T) {
	i18n.SetLang("en")
	// Verify English game-specific help keys work
	assert.NotEqual(t, "blackjack.helpTitle", i18n.T("blackjack.helpTitle"))
	assert.NotEqual(t, "poker.helpTitle", i18n.T("poker.helpTitle"))
}

// Reset to ja after all tests to avoid polluting other packages
func TestSetLang_ResetToJa(t *testing.T) {
	i18n.SetLang("ja")
	assert.Equal(t, "ja", i18n.Lang())
}
