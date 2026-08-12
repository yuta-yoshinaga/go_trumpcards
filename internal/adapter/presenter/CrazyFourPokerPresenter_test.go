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

// newCrazyFourPokerForPresenter は本物のドメインを返す。
//
// **モックではなく本物を使う。** アクセサを 25 個モックすると返す値を自分で決めて
// しまうので、「プレゼンタが盤面を正しく読めているか」の検査にならない。
func newCrazyFourPokerForPresenter(t *testing.T) *domain.CrazyFourPoker {
	t.Helper()
	g := domain.NewDefaultCrazyFourPoker()
	g.Reset()
	return g
}

// crazyFourPokerDecided は決着済みの盤面を返す。
func crazyFourPokerDecided(t *testing.T) *domain.CrazyFourPoker {
	t.Helper()
	g := newCrazyFourPokerForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 20))
	require.NoError(t, g.Play(g.MaxPlayMultiplier()))
	return g
}

// --- CUI ---

// **生のキーが画面に出ていないこと。**
//
// ロケールが 1 か所でもネストしていると Go 側は map[string]string に読めず、
// そのゲームの訳が丸ごと落ちる。日本語の実文字列で確かめる。
func TestCrazyFourPokerCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(CrazyFourPokerCuiPresenter)
	out := cp.Output(newCrazyFourPokerForPresenter(t), nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ラウンド:")
	assert.NotContains(t, out, "crazyfourpoker.", "生の i18n キーが出力に混ざっている")
}

// **決着するまでディーラーの手は伏せる。**
//
// 5 枚から最良の 4 枚を選ぶゲームなので、見えていると判断がまるごと変わる。
func TestCrazyFourPokerCuiPresenter_HidesTheDealerUntilTheShowdown(t *testing.T) {
	cp := new(CrazyFourPokerCuiPresenter)

	g := newCrazyFourPokerForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 0))
	require.Equal(t, domain.CrazyFourPokerPhaseDecide, g.GetPhase())

	out := cp.Output(g, nil)
	assert.Contains(t, out, "あなた:")
	assert.NotContains(t, out, "ディーラー:", "判断前にディーラーの手が見えている")
	assert.Contains(t, out, "置ける倍率", "選べる倍率が画面に出ていない")

	require.NoError(t, g.Play(1))
	assert.Contains(t, cp.Output(g, nil), "ディーラー:", "決着後もディーラーの手が伏せられている")
}

func TestCrazyFourPokerCuiPresenter_ShowsBetsAndResult(t *testing.T) {
	cp := new(CrazyFourPokerCuiPresenter)
	g := crazyFourPokerDecided(t)

	out := cp.Output(g, nil)
	assert.Contains(t, out, "アンティ 50")
	assert.Contains(t, out, "Super Bonus 50")
	assert.Contains(t, out, "Queens Up 20")
	assert.Contains(t, out, "決着:")
	assert.NotContains(t, out, "crazyfourpoker.")
}

func TestCrazyFourPokerCuiPresenter_ShowsErrorsAndUnknownPhase(t *testing.T) {
	cp := new(CrazyFourPokerCuiPresenter)
	g := newCrazyFourPokerForPresenter(t)
	assert.Contains(t, cp.Output(g, errors.New("アンティが範囲外です")), "アンティが範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))
}

func TestCrazyFourPokerCuiPresenter_HintOutput(t *testing.T) {
	cp := new(CrazyFourPokerCuiPresenter)

	t.Run("判断どころでなければその旨", func(t *testing.T) {
		g := newCrazyFourPokerForPresenter(t)
		assert.NotEmpty(t, cp.HintOutput(g))
		assert.NotContains(t, cp.HintOutput(g), "crazyfourpoker.")
	})

	t.Run("判断中は倍率か降りるを薦める", func(t *testing.T) {
		g := newCrazyFourPokerForPresenter(t)
		require.NoError(t, g.PlaceBet(50, 0))
		out := cp.HintOutput(g)
		assert.NotEmpty(t, out)
		assert.NotContains(t, out, "crazyfourpoker.", "理由キーが訳されていない")
	})
}

// --- Web ---

// **配る前も配列は配列。**
func TestCrazyFourPokerWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(CrazyFourPokerWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newCrazyFourPokerForPresenter(t), nil)), &out))

	for _, key := range []string{"playerHand", "dealerHand", "playerBest", "dealerBest"} {
		assert.Equal(t, "[]", string(out[key]), "%s が配列で返っていない", key)
	}
}

// **判断中はディーラーの手をワイヤに載せない。**
//
// 画面で隠しても、レスポンスに入っていれば開発者ツールで見える。
func TestCrazyFourPokerWebPresenter_DealerHandIsNotOnTheWireBeforeTheShowdown(t *testing.T) {
	cp := new(CrazyFourPokerWebPresenter)
	g := newCrazyFourPokerForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 0))

	var got struct {
		Phase          int               `json:"phase"`
		PlayerHand     []json.RawMessage `json:"playerHand"`
		DealerHand     []json.RawMessage `json:"dealerHand"`
		DealerBest     []json.RawMessage `json:"dealerBest"`
		DealerHandRank int               `json:"dealerHandRank"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Equal(t, int(domain.CrazyFourPokerPhaseDecide), got.Phase)
	assert.Len(t, got.PlayerHand, domain.CrazyFourPokerHandSize)
	assert.Empty(t, got.DealerHand, "判断前にディーラーの手がワイヤに乗っている")
	assert.Empty(t, got.DealerBest)
	assert.Zero(t, got.DealerHandRank, "判断前にディーラーの役が漏れている")
}

func TestCrazyFourPokerWebPresenter_ShowdownRevealsTheDealer(t *testing.T) {
	cp := new(CrazyFourPokerWebPresenter)
	g := crazyFourPokerDecided(t)

	var got struct {
		DealerHand      []json.RawMessage `json:"dealerHand"`
		DealerBest      []json.RawMessage `json:"dealerBest"`
		DealerHandRank  int               `json:"dealerHandRank"`
		Result          int               `json:"result"`
		AnteBet         int               `json:"anteBet"`
		SuperBet        int               `json:"superBet"`
		QueensUpBet     int               `json:"queensUpBet"`
		MaxMultiplier   int               `json:"maxMultiplier"`
		HasAcesOrBetter bool              `json:"hasAcesOrBetter"`
		Chips           int               `json:"chips"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Len(t, got.DealerHand, domain.CrazyFourPokerHandSize)
	assert.Len(t, got.DealerBest, domain.CrazyFourPokerBestSize)
	assert.Positive(t, got.DealerHandRank)
	assert.NotZero(t, got.Result)
	assert.Equal(t, 50, got.AnteBet)
	assert.Equal(t, 50, got.SuperBet, "Super Bonus がアンティと違う")
	assert.Equal(t, 20, got.QueensUpBet)
	assert.Equal(t, g.GetChips(), got.Chips)
}

// **上限倍率はサーバが載せる。** ページに計算し直させない。
func TestCrazyFourPokerWebPresenter_MaxMultiplierIsOnTheWire(t *testing.T) {
	cp := new(CrazyFourPokerWebPresenter)
	g := newCrazyFourPokerForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 0))

	var got struct {
		MaxMultiplier   int  `json:"maxMultiplier"`
		HasAcesOrBetter bool `json:"hasAcesOrBetter"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Equal(t, g.MaxPlayMultiplier(), got.MaxMultiplier)
	assert.Equal(t, g.PlayerHasAcesOrBetter(), got.HasAcesOrBetter)
	// 上限は必ず 1 か 3 のどちらか。2 が出るなら規則が壊れている。
	assert.Contains(t, []int{domain.CrazyFourPokerPlayNormalMax, domain.CrazyFourPokerPlayAcesMax},
		got.MaxMultiplier)
}

func TestCrazyFourPokerWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(CrazyFourPokerWebPresenter)
	g := newCrazyFourPokerForPresenter(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)

	assert.JSONEq(t, `{"hint":null}`, cp.HintOutput(g))

	require.NoError(t, g.PlaceBet(50, 0))
	var hint struct {
		Multiplier int    `json:"multiplier"`
		Reason     string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(g)), &hint))
	assert.NotEmpty(t, hint.Reason)

	assert.NotEmpty(t, cp.ActionLogOutput(g))
}
