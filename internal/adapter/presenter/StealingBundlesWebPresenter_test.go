//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newStealingBundlesForWeb(t *testing.T) *domain.StealingBundles {
	t.Helper()
	s := domain.NewDefaultStealingBundles()
	s.Reset()
	return s
}

func decodeStealingBundles(t *testing.T, str string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(str), &m))
	return m
}

func TestStealingBundlesWebPresenterOutput(t *testing.T) {
	p := new(StealingBundlesWebPresenter)
	s := newStealingBundlesForWeb(t)
	m := decodeStealingBundles(t, p.Output(s, nil))

	assert.Equal(t, float64(domain.StealingBundlesPhasePlay), m["phase"])
	assert.Equal(t, float64(-1), m["winnerIdx"])
	assert.Equal(t, float64(-1), m["lastCaptureIdx"], "まだ誰も取っていない")
	assert.Len(t, m["tableCards"].([]any), domain.StealingBundlesTableSize)
	assert.Equal(t, float64(1), m["packsDealt"])
	n := domain.StealingBundlesDefaultPlayerCnt
	assert.Equal(t, float64(domain.StealingBundlesDeckSize-domain.StealingBundlesTableSize-n*domain.StealingBundlesHandSize),
		m["deckRemaining"])

	players := m["players"].([]any)
	require.Len(t, players, n)
	human := players[0].(map[string]any)
	assert.True(t, human["isHuman"].(bool))
	assert.Equal(t, float64(domain.StealingBundlesHandSize), human["cardCount"])
	assert.Zero(t, human["bundleSize"])
	assert.Nil(t, human["bundleTop"], "束が空なら一番上は無い")
	assert.Empty(t, players[1].(map[string]any)["cards"], "CPU の手札は伏せる")
}

// **束の一番上は全員に見えます。** そこが狙われる場所だからです。
func TestStealingBundlesWebPresenterExposesEveryBundleTop(t *testing.T) {
	p := new(StealingBundlesWebPresenter)
	s := newStealingBundlesForWeb(t)
	s.GetPlayer(1).SetBundle([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
	})

	cpu := decodeStealingBundles(t, p.Output(s, nil))["players"].([]any)[1].(map[string]any)
	assert.Equal(t, float64(2), cpu["bundleSize"])
	require.NotNil(t, cpu["bundleTop"])
	assert.Equal(t, float64(9), cpu["bundleTop"].(map[string]any)["value"])
	assert.Empty(t, cpu["cards"], "手札は伏せたまま")
}

// **どの札で何ができるかは盤面から読み切れません。**
func TestStealingBundlesWebPresenterMapsTheAvailableMoves(t *testing.T) {
	p := new(StealingBundlesWebPresenter)
	s := newStealingBundlesForWeb(t)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
	})
	s.GiveHandForTest(0,
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 4, false))
	s.GetPlayer(2).SetBundle([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 9, false)})

	m := decodeStealingBundles(t, p.Output(s, nil))
	assert.Len(t, m["tableMatches"].(map[string]any)["0"], 2, "7 は場の 2 枚を取れる")
	assert.Equal(t, []any{float64(2)}, m["stealTargets"].(map[string]any)["1"], "9 は席 2 の束を奪える")
	assert.NotContains(t, m["tableMatches"], "2", "4 では何も取れない")
	assert.True(t, m["canCapture"].(bool))
	assert.Equal(t, "stealingbundles.mustCapture", m["messageCode"])
}

// **取れないときだけ場に置けます。**
func TestStealingBundlesWebPresenterFlagsWhenOnlyTrailingIsLeft(t *testing.T) {
	p := new(StealingBundlesWebPresenter)
	s := newStealingBundlesForWeb(t)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	s.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 4, false))
	for i := 1; i < s.GetPlayerCnt(); i++ {
		s.GetPlayer(i).SetBundle(nil)
	}

	m := decodeStealingBundles(t, p.Output(s, nil))
	assert.False(t, m["canCapture"].(bool))
	assert.Empty(t, m["tableMatches"])
	assert.Empty(t, m["stealTargets"])
	assert.Equal(t, "stealingbundles.trail", m["messageCode"])
}

func TestStealingBundlesWebPresenterMessages(t *testing.T) {
	p := new(StealingBundlesWebPresenter)

	t.Run("エラーはそのまま返す", func(t *testing.T) {
		m := decodeStealingBundles(t, p.Output(newStealingBundlesForWeb(t), errors.New("boom")))
		assert.Equal(t, "boom", m["message"])
		assert.Empty(t, m["messageCode"])
	})

	t.Run("相手の番はそう言う", func(t *testing.T) {
		s := newStealingBundlesForWeb(t)
		s.SetCurrentPlayerIdxForTest(1)
		assert.Equal(t, "stealingbundles.waiting", decodeStealingBundles(t, p.Output(s, nil))["messageCode"])
	})

	t.Run("投了すると相手の勝ち", func(t *testing.T) {
		s := newStealingBundlesForWeb(t)
		s.GiveUp()
		m := decodeStealingBundles(t, p.Output(s, nil))
		assert.Equal(t, "stealingbundles.result.cpu", m["messageCode"])
	})

	t.Run("いちばん多く集めたら勝ち", func(t *testing.T) {
		s := newStealingBundlesForWeb(t)
		s.DrainDeckForTest()
		s.SetCurrentPlayerIdxForTest(0)
		s.SetTableCardsForTest([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
		s.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 7, false))
		for i := 1; i < s.GetPlayerCnt(); i++ {
			s.GiveHandForTest(i)
		}
		require.NoError(t, s.PlayerTake(0))
		require.True(t, s.GetGameEndFlag())

		m := decodeStealingBundles(t, p.Output(s, nil))
		assert.Equal(t, "stealingbundles.result.you", m["messageCode"])
		assert.Equal(t, "2", m["messageParams"].(map[string]any)["n"])
	})
}

func TestStealingBundlesWebPresenterHintOutput(t *testing.T) {
	p := new(StealingBundlesWebPresenter)
	s := newStealingBundlesForWeb(t)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
	s.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 5, false))

	hint := decodeStealingBundles(t, p.HintOutput(s))["hint"].(map[string]any)
	assert.Equal(t, "stealingbundlesTake", hint["reason"])
	assert.Equal(t, float64(-1), hint["victimIdx"])

	// **束を奪えるならそちらを勧めます。** 相手の席も一緒に返します。
	s.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 9, false))
	s.GetPlayer(1).SetBundle([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
	})
	hint = decodeStealingBundles(t, p.HintOutput(s))["hint"].(map[string]any)
	assert.Equal(t, "stealingbundlesSteal", hint["reason"])
	assert.Equal(t, float64(1), hint["victimIdx"])

	// 受動ヒントは Output() でも埋まる (#4483)。
	assert.NotNil(t, decodeStealingBundles(t, p.Output(s, nil))["hint"])

	s.GiveUp()
	assert.Nil(t, decodeStealingBundles(t, p.HintOutput(s))["hint"], "終局後は助言しない")
}

// **棋譜は終局まで伏せる。**
func TestStealingBundlesWebPresenterActionLogOutput(t *testing.T) {
	p := new(StealingBundlesWebPresenter)
	s := newStealingBundlesForWeb(t)
	assert.Empty(t, decodeStealingBundles(t, p.ActionLogOutput(s))["entries"])

	s.GiveUp()
	assert.NotEmpty(t, decodeStealingBundles(t, p.ActionLogOutput(s))["entries"])
}
