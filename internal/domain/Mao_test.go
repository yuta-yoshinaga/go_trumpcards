//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// maoCard はテスト用のカード生成ヘルパー。
func maoCard(d, v int) *Card {
	return NewCard(d, v, false)
}

// newTestMao はデフォルト設定の4人対戦Maoを生成する。
func newTestMao() *Mao {
	players := []*MaoPlayer{
		NewMaoPlayer(true),
		NewMaoPlayer(false),
		NewMaoPlayer(false),
		NewMaoPlayer(false),
	}
	return NewMao(NewTrumpCards(0), players, DefaultMaoConfig())
}

// maoSetHand はプレイヤーの手札を指定カードに置き換える。
func maoSetHand(p *MaoPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestMao_ResetInitialState(t *testing.T) {
	g := newTestMao()
	g.Reset()

	assert.Equal(t, MaoPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, MaoPlayerCnt, g.GetPlayerCnt())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.NotNil(t, g.GetDiscardTop())
	assert.False(t, g.GetAwaitingWord())
	assert.Equal(t, 0, g.GetPlayerCorrectCount())
	assert.False(t, g.GetHintUnlocked())
	// Each player has 5 cards
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, MaoHandSize, g.GetPlayer(i).GetCardsSize())
	}
}

func TestMao_SelectHiddenRule_Deterministic(t *testing.T) {
	// Same deck order => same hidden rule (deterministic selection).
	g1 := newTestMao()
	g1.SetDrawPile([]*Card{maoCard(CardDesignSpade, 3), maoCard(CardDesignHeart, 9), maoCard(CardDesignClover, 5)})
	g1.SetDiscardPile([]*Card{maoCard(CardDesignDiamond, 4)})
	g1.selectHiddenRule()
	rule1 := g1.GetHiddenRule()

	g2 := newTestMao()
	g2.SetDrawPile([]*Card{maoCard(CardDesignSpade, 3), maoCard(CardDesignHeart, 9), maoCard(CardDesignClover, 5)})
	g2.SetDiscardPile([]*Card{maoCard(CardDesignDiamond, 4)})
	g2.selectHiddenRule()
	rule2 := g2.GetHiddenRule()

	assert.Equal(t, rule1, rule2)
	assert.NotEmpty(t, rule1.RequiredWord)
}

func TestMao_RuleTriggered(t *testing.T) {
	g := newTestMao()

	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade"})
	assert.True(t, g.ruleTriggered(maoCard(CardDesignSpade, 5)))
	assert.False(t, g.ruleTriggered(maoCard(CardDesignHeart, 5)))
	assert.False(t, g.ruleTriggered(nil))

	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerValue, TriggerValue: 7, RequiredWord: "seven"})
	assert.True(t, g.ruleTriggered(maoCard(CardDesignHeart, 7)))
	assert.False(t, g.ruleTriggered(maoCard(CardDesignHeart, 8)))
}

func TestMao_PlayTriggersAwaitingWord(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade", HintKey: "hintSuit"})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	g.SetChosenSuit(-1)
	g.SetDiscardPile([]*Card{maoCard(CardDesignSpade, 5)})
	// Human plays a Spade => trigger fires
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, 9), maoCard(CardDesignHeart, 4))

	require.NoError(t, g.PlayerPlay(0))
	assert.True(t, g.GetAwaitingWord())
}

func TestMao_DeclareWordCorrect(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade", HintKey: "hintSuit"})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	g.SetChosenSuit(-1)
	g.SetDiscardPile([]*Card{maoCard(CardDesignSpade, 5)})
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, 9), maoCard(CardDesignHeart, 4))

	require.NoError(t, g.PlayerPlay(0))
	require.True(t, g.GetAwaitingWord())

	// Correct word (case-insensitive) => correct count up, no penalty
	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDeclareWord("SPADE"))
	assert.False(t, g.GetAwaitingWord())
	assert.Equal(t, 1, g.GetPlayerCorrectCount())
	assert.False(t, g.GetRulePenaltyFlag())
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize())
}

func TestMao_DeclareWordWrong_AppliesPenalty(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade", HintKey: "hintSuit"})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	g.SetChosenSuit(-1)
	g.SetDiscardPile([]*Card{maoCard(CardDesignSpade, 5)})
	g.SetDrawPile([]*Card{maoCard(CardDesignClover, 6)})
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, 9), maoCard(CardDesignHeart, 4))

	require.NoError(t, g.PlayerPlay(0))
	require.True(t, g.GetAwaitingWord())

	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDeclareWord("wrong"))
	assert.False(t, g.GetAwaitingWord())
	assert.Equal(t, 0, g.GetPlayerCorrectCount())
	assert.True(t, g.GetRulePenaltyFlag())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
}

func TestMao_DeclareWordWhenNotAwaiting_Penalty(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade", HintKey: "hintSuit"})
	g.SetCurrentPlayerIdx(0)
	g.SetAwaitingWord(false)
	g.SetDrawPile([]*Card{maoCard(CardDesignClover, 6)})
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, 9))

	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDeclareWord("spade"))
	assert.True(t, g.GetRulePenaltyFlag())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
}

func TestMao_HintUnlocksAfterThreeCorrect(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade", HintKey: "hintSuit"})

	for i := 0; i < MaoHintThreshold; i++ {
		g.SetAwaitingWord(true)
		require.NoError(t, g.PlayerDeclareWord("spade"))
	}
	assert.True(t, g.GetHintUnlocked())
	assert.Equal(t, "hintSuit", g.GetRuleHintKey())
	assert.Equal(t, MaoHintThreshold, g.GetPlayerCorrectCount())
}

func TestMao_RuleHintHiddenBeforeUnlock(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade", HintKey: "hintSuit"})
	assert.False(t, g.GetHintUnlocked())
	assert.Equal(t, "", g.GetRuleHintKey())
}

func TestMao_PlayWhileAwaitingWord_Penalizes(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerValue, TriggerValue: 99, RequiredWord: "x", HintKey: "hintSuit"})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	g.SetChosenSuit(-1)
	g.SetAwaitingWord(true) // a pending word from a prior turn
	g.SetDiscardPile([]*Card{maoCard(CardDesignHeart, 5)})
	g.SetDrawPile([]*Card{maoCard(CardDesignClover, 6)})
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignHeart, 9))

	require.NoError(t, g.PlayerPlay(0))
	assert.True(t, g.GetRulePenaltyFlag())
	assert.False(t, g.GetAwaitingWord())
}

func TestMao_isValidPlay(t *testing.T) {
	g := newTestMao()
	g.Reset()

	g.SetChosenSuit(-1)
	g.SetPenaltyDrawCount(0)
	g.SetDiscardPile([]*Card{maoCard(CardDesignSpade, 5)})
	assert.True(t, g.isValidPlay(maoCard(CardDesignHeart, 8)), "wild 8 always valid")
	assert.True(t, g.isValidPlay(maoCard(CardDesignSpade, 3)), "suit match")
	assert.True(t, g.isValidPlay(maoCard(CardDesignHeart, 5)), "rank match")
	assert.False(t, g.isValidPlay(maoCard(CardDesignHeart, 7)), "no match")

	g.SetPenaltyDrawCount(2)
	assert.True(t, g.isValidPlay(maoCard(CardDesignHeart, MaoDrawTwoValue)), "only 2 stacks under penalty")
	assert.False(t, g.isValidPlay(maoCard(CardDesignSpade, 5)))

	g.SetPenaltyDrawCount(0)
	g.SetChosenSuit(CardDesignHeart)
	assert.True(t, g.isValidPlay(maoCard(CardDesignHeart, 9)), "chosen-suit match")
	assert.False(t, g.isValidPlay(maoCard(CardDesignSpade, 9)))
}

func TestMao_PlayDrawTwoStacks(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerValue, TriggerValue: 99, RequiredWord: "x"})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	g.SetChosenSuit(-1)
	g.SetDiscardPile([]*Card{maoCard(CardDesignSpade, 5)})
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, MaoDrawTwoValue), maoCard(CardDesignHeart, 9))

	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, MaoDrawTwoAmount, g.GetPenaltyDrawCount())
}

func TestMao_PlaySkipAdvancesTwo(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerValue, TriggerValue: 99, RequiredWord: "x"})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	g.SetChosenSuit(-1)
	g.SetDirection(1)
	g.SetDiscardPile([]*Card{maoCard(CardDesignSpade, 5)})
	// Three cards so that after playing the skip the hand still has 2 cards —
	// otherwise dropping to 1 card would enter MustDeclare and the turn would
	// not advance, masking the skip behaviour under test.
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, MaoSkipValue), maoCard(CardDesignHeart, 9), maoCard(CardDesignClover, 4))

	require.NoError(t, g.PlayerPlay(0))
	// player 0 -> skip player 1 -> player 2
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
}

func TestMao_PlayWildEntersChooseSuit(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerValue, TriggerValue: 99, RequiredWord: "x"})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	g.SetChosenSuit(-1)
	g.SetDiscardPile([]*Card{maoCard(CardDesignSpade, 5)})
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignHeart, MaoWildValue), maoCard(CardDesignHeart, 9))

	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, MaoPhaseChooseSuit, g.GetPhase())
	require.NoError(t, g.PlayerChooseSuit(CardDesignClover))
	assert.Equal(t, CardDesignClover, g.GetChosenSuit())
}

func TestMao_MustDeclareThenDeclare(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerValue, TriggerValue: 99, RequiredWord: "x"})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	g.SetChosenSuit(-1)
	g.SetDiscardPile([]*Card{maoCard(CardDesignSpade, 5)})
	g.GetPlayer(0).SetHasDeclared(false)
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, 3), maoCard(CardDesignHeart, 9))

	require.NoError(t, g.PlayerPlay(0)) // down to 1 card
	assert.Equal(t, MaoPhaseMustDeclare, g.GetPhase())
	require.NoError(t, g.PlayerDeclareMao())
	assert.True(t, g.GetPlayer(0).GetHasDeclared())
}

func TestMao_SkipDeclarePenalty(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhaseMustDeclare)
	g.SetDrawPile([]*Card{maoCard(CardDesignClover, 6), maoCard(CardDesignClover, 7)})
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, 3))

	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerSkipDeclare())
	assert.Equal(t, before+MaoForgotPenalty, g.GetPlayer(0).GetCardsSize())
}

func TestMao_WrongPhaseAndTurnErrors(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetPhase(MaoPhaseChooseSuit)
	assert.Error(t, g.PlayerPlay(0))
	g.SetPhase(MaoPhasePlay)
	g.SetCurrentPlayerIdx(1) // CPU
	assert.Error(t, g.PlayerPlay(0))
	assert.Error(t, g.PlayerDraw())

	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPlay(0), ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDeclareWord("x"), ErrGameEnded)
}

func TestMao_InvalidCardIndex(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(MaoPhasePlay)
	maoSetHand(g.GetPlayer(0), maoCard(CardDesignSpade, 3))
	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(99))
}

func TestMao_FullCpuGameTerminates(t *testing.T) {
	for _, diff := range []MaoCpuDifficulty{MaoCpuDifficultyEasy, MaoCpuDifficultyNormal, MaoCpuDifficultyHard} {
		cfg := DefaultMaoConfig()
		cfg.CpuDifficulty = diff
		cfg.PointLimit = 30 // low limit so the game ends quickly
		// All players CPU so the game runs unattended.
		players := []*MaoPlayer{
			NewMaoPlayer(false),
			NewMaoPlayer(false),
			NewMaoPlayer(false),
			NewMaoPlayer(false),
		}
		g := NewMao(NewTrumpCards(0), players, cfg)
		g.Reset()

		for step := 0; step < 200000 && !g.GetGameEndFlag(); step++ {
			switch g.GetPhase() {
			case MaoPhasePlay:
				g.CpuPlay()
			case MaoPhaseChooseSuit:
				g.CpuChooseSuit()
			case MaoPhaseMustDeclare:
				g.CpuDeclare()
			case MaoPhaseRoundEnd:
				g.ScoreRound()
				if !g.GetGameEndFlag() {
					g.NextRound()
				}
			case MaoPhaseGameEnd:
				// loop exits via GetGameEndFlag
			}
		}
		assert.True(t, g.GetGameEndFlag(), "difficulty %d should terminate", diff)
		assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
	}
}

func TestMao_JSONRoundTripIncludesHiddenRule(t *testing.T) {
	g := newTestMao()
	g.Reset()
	rule := MaoHiddenRule{TriggerKind: MaoTriggerValue, TriggerValue: 7, RequiredWord: "seven", HintKey: "hintNumber"}
	g.SetHiddenRule(rule)
	g.SetAwaitingWord(true)
	g.playerCorrectCount = 2
	g.hintUnlocked = false

	data, err := json.Marshal(g)
	require.NoError(t, err)
	// The marshalled domain JSON MUST contain the hidden rule (KV round-trip).
	assert.Contains(t, string(data), "seven")

	var restored Mao
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, rule, restored.GetHiddenRule())
	assert.True(t, restored.GetAwaitingWord())
	assert.Equal(t, 2, restored.GetPlayerCorrectCount())
	assert.Equal(t, MaoPlayerCnt, restored.GetPlayerCnt())
}

func TestMao_UnmarshalRejectsBadInput(t *testing.T) {
	// invalid JSON
	var g Mao
	assert.Error(t, json.Unmarshal([]byte("{"), &g))

	// wrong player count
	bad := `{"pl":[],"ci":0,"ps":0,"cf":{"cd":1,"pl":200}}`
	assert.Error(t, json.Unmarshal([]byte(bad), &g))

	// out-of-range phase via a valid 4-player marshal then tampering is awkward;
	// instead build from a real game and tamper the phase.
	base := newTestMao()
	base.Reset()
	data, err := json.Marshal(base)
	require.NoError(t, err)
	var jm map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &jm))
	jm["ps"] = json.RawMessage("99")
	tampered, err := json.Marshal(jm)
	require.NoError(t, err)
	assert.Error(t, json.Unmarshal(tampered, &g))

	// bad currentPlayerIdx
	jm["ps"] = json.RawMessage("0")
	jm["ci"] = json.RawMessage("9")
	tampered2, err := json.Marshal(jm)
	require.NoError(t, err)
	assert.Error(t, json.Unmarshal(tampered2, &g))
}

func TestMao_NextRoundKeepsHiddenRuleProgress(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetHiddenRule(MaoHiddenRule{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade", HintKey: "hintSuit"})
	g.playerCorrectCount = 3
	g.hintUnlocked = true
	g.SetPhase(MaoPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 3, g.GetPlayerCorrectCount())
	assert.True(t, g.GetHintUnlocked())
	assert.Equal(t, 2, g.GetRoundNumber())
}

func TestMao_ScoreRoundAndConfig(t *testing.T) {
	g := newTestMao()
	g.Reset()
	g.SetPhase(MaoPhaseRoundEnd)
	// Player 0 empties hand; others keep cards.
	maoSetHand(g.GetPlayer(0))
	maoSetHand(g.GetPlayer(1), maoCard(CardDesignSpade, 9))
	maoSetHand(g.GetPlayer(2), maoCard(CardDesignHeart, 10))
	maoSetHand(g.GetPlayer(3), maoCard(CardDesignClover, 3))
	g.ScoreRound()
	assert.Greater(t, g.GetPlayer(0).GetRoundScore(), 0)

	cfg := DefaultMaoConfig()
	assert.NoError(t, cfg.Validate())
	cfg.PointLimit = 0
	assert.Error(t, cfg.Validate())
}

// **ヒント文は i18n キーで持つ。**Go に日本語を直書きすると `--lang en` でも
// 日本語のまま出る (#4917)。
func TestMao_RuleHintIsAnI18nKeyNotJapaneseText(t *testing.T) {
	for i, r := range maoRuleSet {
		assert.NotEmpty(t, r.HintKey, "rule %d", i)
		assert.True(t, strings.HasPrefix(r.HintKey, "hint"), "rule %d: %q is not a hint key", i, r.HintKey)
		for _, ru := range r.HintKey {
			assert.Less(t, ru, rune(128), "rule %d: %q must be ASCII, not a translated string", i, r.HintKey)
		}
	}
}

// **キー集合を固定する。**presenter 側の翻訳確認テストと突き合わせる相手が
// 無いと、キーを増やしたときに訳し忘れても誰も気づかない (#4917)。
func TestMao_RuleHintKeySet(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range maoRuleSet {
		seen[r.HintKey] = true
	}
	assert.Equal(t, map[string]bool{"hintSuit": true, "hintNumber": true, "hintFace": true, "hintRank": true}, seen)
}
