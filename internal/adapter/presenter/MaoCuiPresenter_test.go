//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupMaoCuiMock() *interfaces.MockMaoGame {
	m := new(interfaces.MockMaoGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(30)
	m.On("GetDirection").Return(1)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetChosenSuit").Return(-1)
	m.On("GetPenaltyDrawCount").Return(0)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MaoPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetAwaitingWord").Return(false)
	m.On("GetHintUnlocked").Return(false)
	m.On("GetPlayerCorrectCount").Return(1)
	m.On("GetRuleHintKey").Return("")
	m.On("GetRulePenaltyFlag").Return(false)
	return m
}

func setupMaoCuiMockWithPlayers() (*interfaces.MockMaoGame, []*domain.MaoPlayer) {
	m := setupMaoCuiMock()
	players := makeMaoPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestMaoCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.MaoCuiPresenter)

	t.Run("initial state header and players", func(t *testing.T) {
		m, players := setupMaoCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Mao")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "手番: あなた")
		// Before the hint unlocks, the compliance progress (1/3) is shown.
		assert.Contains(t, result, "ルール適合: 1/3")
	})

	t.Run("unlocked hint replaces the compliance progress", func(t *testing.T) {
		m, _ := setupMaoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHintUnlocked")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRuleHintKey")
		m.On("GetHintUnlocked").Return(true)
		m.On("GetRuleHintKey").Return("hintNumber")
		result := p.Output(m, nil)
		assert.NotContains(t, result, "ルール適合:")
		// **キーではなく訳文が出る。**キーがそのまま出たら翻訳を通していない (#4917)。
		assert.Contains(t, result, "ある数字を出したときに言葉が必要です。")
		assert.NotContains(t, result, "hintNumber")
	})

	t.Run("awaiting word prompt and no hidden rule leak", func(t *testing.T) {
		m, _ := setupMaoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetAwaitingWord")
		m.On("GetAwaitingWord").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "秘密のルール")
		assert.Contains(t, result, "dw <word>")
	})

	t.Run("hint shown when unlocked", func(t *testing.T) {
		m, _ := setupMaoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHintUnlocked")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRuleHintKey")
		m.On("GetHintUnlocked").Return(true)
		m.On("GetRuleHintKey").Return("hintSuit")
		result := p.Output(m, nil)
		assert.Contains(t, result, "あるスートを出したときに言葉が必要です。")
	})

	t.Run("rule penalty notice", func(t *testing.T) {
		m, _ := setupMaoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRulePenaltyFlag")
		m.On("GetRulePenaltyFlag").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ペナルティ")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupMaoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("penalty stack shown", func(t *testing.T) {
		m, _ := setupMaoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPenaltyDrawCount")
		m.On("GetPenaltyDrawCount").Return(4)
		result := p.Output(m, nil)
		assert.Contains(t, result, "4")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupMaoCuiMockWithPlayers()
		result := p.Output(m, errors.New("invalid card index"))
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended shows winner", func(t *testing.T) {
		m, _ := setupMaoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
	})

	t.Run("phase prompts", func(t *testing.T) {
		for phase, want := range map[domain.MaoPhase]string{
			domain.MaoPhaseChooseSuit:  "スート選択フェーズ",
			domain.MaoPhaseMustDeclare: "宣言フェーズ",
			domain.MaoPhaseRoundEnd:    "ラウンド終了",
		} {
			m, _ := setupMaoCuiMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			result := p.Output(m, nil)
			assert.Contains(t, result, want)
		}
	})
}

func TestMaoCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MaoCuiPresenter)
	m := new(interfaces.MockMaoGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays SPADE 5"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewMaoPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "You plays SPADE 5")
}

// MaoRuleHintKeys は domain.maoRuleSet が使うヒントキー。ここを増やしたら
// TestMao_RuleHintKeySet も落ちるので、片方だけ更新して訳し忘れることはない。
var MaoRuleHintKeys = []string{"hintSuit", "hintNumber", "hintFace", "hintRank"}

// **ja / en 両方に訳がある。**片方だけだと `--lang en` で日本語のまま出る (#4917)。
func TestMaoRuleHintKeys_TranslatedInBothLanguages(t *testing.T) {
	defer i18n.SetLang("ja")
	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		for _, key := range MaoRuleHintKeys {
			full := "mao." + key
			got := i18n.T(full)
			assert.NotEqual(t, full, got, "%s is missing from %s", full, lang)
			assert.NotEmpty(t, got, "%s is empty in %s", full, lang)
		}
	}
}

// 訳が言語ごとに違うことも見る。両方に同じ日本語を入れても上のテストは通る。
func TestMaoRuleHintKeys_DifferPerLanguage(t *testing.T) {
	defer i18n.SetLang("ja")
	for _, key := range MaoRuleHintKeys {
		i18n.SetLang("ja")
		ja := i18n.T("mao." + key)
		i18n.SetLang("en")
		en := i18n.T("mao." + key)
		assert.NotEqual(t, ja, en, "mao.%s is the same string in both languages", key)
	}
}
