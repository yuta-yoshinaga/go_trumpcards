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
	// Verify substitution works independently across two separate calls.
	result := i18n.Tf("holdem.invalidSmallBlind", "val", "0")
	assert.Equal(t, "Invalid small blind: 0. Please enter 1 or more.", result)
	result2 := i18n.Tf("holdem.invalidBigBlind", "val", "1")
	assert.Equal(t, "Invalid big blind: 1. Please enter 2 or more.", result2)
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

func TestMarkError(t *testing.T) {
	assert.Equal(t, i18n.ErrorPrefix+"boom", i18n.MarkError("boom"))
	assert.Equal(t, "", i18n.MarkError(""))
	// idempotent
	marked := i18n.MarkError("boom")
	assert.Equal(t, marked, i18n.MarkError(marked))
}

func TestStripErrorPrefix(t *testing.T) {
	body, isErr := i18n.StripErrorPrefix(i18n.ErrorPrefix + "boom")
	assert.True(t, isErr)
	assert.Equal(t, "boom", body)

	body, isErr = i18n.StripErrorPrefix("ok")
	assert.False(t, isErr)
	assert.Equal(t, "ok", body)
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

// TestT_CuiCommonKeys_Unprefixed verifies issue #1699 Phase 1: cui_common.json
// keys are loaded into the global namespace (no "cui_common." prefix) so CUI
// presenters can call i18n.T("cuiPlayerYou") directly. The "cui*" prefix on
// every key prevents collisions with common.json.
func TestT_CuiCommonKeys_Unprefixed(t *testing.T) {
	i18n.SetLang("ja")
	assert.Equal(t, "あなた", i18n.T("cuiPlayerYou"))
	assert.Equal(t, "フォールド", i18n.T("cuiBettingActionFold"))
	assert.Equal(t, "リードスートに追随", i18n.T("cuiHintFollowSuit"))

	i18n.SetLang("en")
	assert.Equal(t, "You", i18n.T("cuiPlayerYou"))
	assert.Equal(t, "Fold", i18n.T("cuiBettingActionFold"))
	assert.Equal(t, "Follow the lead suit", i18n.T("cuiHintFollowSuit"))
}

// TestTf_CuiCpuName_BothLangs verifies the {{idx}} substitution path.
// Same English/Japanese template, but exercised through the loader to
// guard against the prefix logic regressing.
func TestTf_CuiCpuName_BothLangs(t *testing.T) {
	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			i18n.SetLang(lang)
			assert.Equal(t, "CPU 2", i18n.Tf("cuiPlayerCpu", "idx", "2"))
		})
	}
}

// Reset to ja after all tests to avoid polluting other packages
func TestSetLang_ResetToJa(t *testing.T) {
	i18n.SetLang("ja")
	assert.Equal(t, "ja", i18n.Lang())
}

// **印がユーザーに届いてはいけない。** 行単位の印は盤面の途中に埋まるので、
// 剥がし忘れると制御文字が画面に出る。
func TestMarkErrorLineRoundTrip(t *testing.T) {
	t.Run("marks and strips", func(t *testing.T) {
		marked := i18n.MarkErrorLine("cannot move")
		assert.NotEqual(t, "cannot move", marked)
		body, had := i18n.StripErrorLines(marked)
		assert.True(t, had)
		assert.Equal(t, "cannot move", body)
	})

	t.Run("empty stays empty", func(t *testing.T) {
		assert.Equal(t, "", i18n.MarkErrorLine(""))
	})

	t.Run("marking twice does not nest", func(t *testing.T) {
		once := i18n.MarkErrorLine("x")
		assert.Equal(t, once, i18n.MarkErrorLine(once))
	})

	t.Run("an unmarked board is reported as clean", func(t *testing.T) {
		body, had := i18n.StripErrorLines("plain board\nsecond line")
		assert.False(t, had)
		assert.Equal(t, "plain board\nsecond line", body)
	})

	// 盤面の途中の 1 行だけが印つき、という本来の使い方。
	t.Run("strips a marker buried in a board", func(t *testing.T) {
		board := "top\n" + i18n.MarkErrorLine("refused") + "\nbottom"
		body, had := i18n.StripErrorLines(board)
		assert.True(t, had)
		assert.Equal(t, "top\nrefused\nbottom", body)
		assert.NotContains(t, body, i18n.ErrorLinePrefix)
	})

	// 行の印と返答全体の印は別物。混ざると盤面が stderr へ流れる。
	t.Run("a line marker is not a reply marker", func(t *testing.T) {
		_, isErr := i18n.StripErrorPrefix(i18n.MarkErrorLine("refused"))
		assert.False(t, isErr)
	})
}
