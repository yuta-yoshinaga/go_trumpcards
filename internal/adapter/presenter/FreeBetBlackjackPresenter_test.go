//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newFreeBetForPresenter は本物のドメインを返す。
//
// **モックではなく本物を使う。** アクセサを 25 個モックすると返す値を自分で決めて
// しまうので、「プレゼンタが盤面を正しく読めているか」の検査にならない。
func newFreeBetForPresenter(t *testing.T) *domain.FreeBetBlackjack {
	t.Helper()
	g := domain.NewDefaultFreeBetBlackjack()
	g.Reset()
	return g
}

// freeBetDealtUntil は cond が満たされる配りに当たるまでラウンドを回す。
//
// **配りを積むのではなく引き当てる。** プレゼンタのテストからはドメインの内部を
// 触れないので、無料操作が可能な局面は素直に配って探す。
func freeBetDealtUntil(t *testing.T, cond func(*domain.FreeBetBlackjack) bool) *domain.FreeBetBlackjack {
	t.Helper()
	for range 1000 {
		g := newFreeBetForPresenter(t)
		require.NoError(t, g.PlaceBet(50))
		if cond(g) {
			return g
		}
	}
	t.Fatalf("1000 回配っても条件を満たす局面が出なかった")
	return nil
}

// --- CUI ---

func TestFreeBetCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(FreeBetBlackjackCuiPresenter)
	out := cp.Output(newFreeBetForPresenter(t), nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ラウンド:")
	assert.NotContains(t, out, "freebet.", "生の i18n キーが出力に混ざっている")
}

// **ハウス持ちの額は自分の賭け金と別に出す。**
//
// 「いくら失うのか」がこのゲームでいちばん見せたい数字なので、合算した 1 本の額
// しか出さないと、無料ダブルした手札が「100 賭けている」ようにしか見えなくなる。
func TestFreeBetCuiPresenter_ShowsTheHouseShareSeparately(t *testing.T) {
	cp := new(FreeBetBlackjackCuiPresenter)
	g := freeBetDealtUntil(t, func(g *domain.FreeBetBlackjack) bool { return g.CanFreeDouble() })

	before := cp.Output(g, nil)
	assert.Contains(t, before, "無料ダブル", "できる操作が案内されていない")
	assert.NotContains(t, before, "ハウス", "まだ出資していないのにハウス持ちの額が出ている")

	require.NoError(t, g.FreeDouble())
	after := cp.Output(g, nil)
	assert.Contains(t, after, "賭け 50", "自分の賭け金が倍に化けている")
	assert.Contains(t, after, "ハウス 50", "ハウス持ちの額が出ていない")
	assert.NotContains(t, after, "freebet.")
}

func TestFreeBetCuiPresenter_ShowsResultAndNet(t *testing.T) {
	cp := new(FreeBetBlackjackCuiPresenter)
	g := newFreeBetForPresenter(t)
	require.NoError(t, g.PlaceBet(50))
	for g.GetPhase() == domain.FreeBetPhasePlay {
		require.NoError(t, g.Stand())
	}

	out := cp.Output(g, nil)
	assert.Contains(t, out, "アンティ: 50")
	assert.Contains(t, out, "収支:")
	assert.Contains(t, out, "手札1:")
	assert.NotContains(t, out, "freebet.")
}

// **ディーラーの 22 は画面で名指しする。** 無料ダブル / 無料スプリットの対価が
// これなので、黙って引き分けにすると規則が伝わらない。
func TestFreeBetCuiPresenter_NamesTheDealer22Push(t *testing.T) {
	cp := new(FreeBetBlackjackCuiPresenter)
	g := freeBetDealtUntil(t, func(g *domain.FreeBetBlackjack) bool {
		for g.GetPhase() == domain.FreeBetPhasePlay {
			if err := g.Stand(); err != nil {
				return false
			}
		}
		return g.IsDealerPushed22()
	})

	out := cp.Output(g, nil)
	assert.Contains(t, out, "22でバスト")
	assert.Contains(t, out, "ディーラー22")
	assert.NotContains(t, out, "freebet.")
}

// **収支は自分が出した金だけで測る。** ハウスの出資を賭け金に数えると、
// 無料ダブルして負けたラウンドが実際の 2 倍の損に見える。
func TestFreeBetCuiPresenter_NetIgnoresTheHouseShare(t *testing.T) {
	cp := new(FreeBetBlackjackCuiPresenter)
	g := freeBetDealtUntil(t, func(g *domain.FreeBetBlackjack) bool { return g.CanFreeDouble() })
	// **アンティを置く前ではなく、置いた後のチップを基準にする。**
	// PlaceBet は済んでいるので、ここからの増減がそのまま収支になる。
	base := g.GetChips()
	require.NoError(t, g.FreeDouble())
	for g.GetPhase() == domain.FreeBetPhasePlay {
		require.NoError(t, g.Stand())
	}

	want := g.GetChips() - base - 50
	assert.Contains(t, cp.Output(g, nil), "収支: "+strconv.Itoa(want),
		"表示された収支がチップの増減と食い違っている")
}

func TestFreeBetCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(FreeBetBlackjackCuiPresenter)
	g := newFreeBetForPresenter(t)
	assert.Contains(t, cp.Output(g, errors.New("アンティが範囲外です")), "アンティが範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))
	assert.NotEmpty(t, cp.HintOutput(g))

	// **配ったら必ず助言できる、ではない。** ナチュラルはその場で決着するので、
	// プレイ待ちになる配りを引き当ててから見る。
	inPlay := freeBetDealtUntil(t, func(g *domain.FreeBetBlackjack) bool {
		return g.GetPhase() == domain.FreeBetPhasePlay
	})
	hint := cp.HintOutput(inPlay)
	assert.NotEmpty(t, hint)
	assert.NotContains(t, hint, "freebet.", "助言のキーが訳されていない")
}

// --- Web ---

func TestFreeBetWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(FreeBetBlackjackWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newFreeBetForPresenter(t), nil)), &out))
	for _, key := range []string{"hands", "dealerCards"} {
		assert.Equal(t, "[]", string(out[key]), "%s が配列で返っていない", key)
	}
}

// **2 つの金の出どころが別々の欄でワイヤに乗る。**
func TestFreeBetWebPresenter_BothMoneySourcesAreOnTheWire(t *testing.T) {
	cp := new(FreeBetBlackjackWebPresenter)
	g := freeBetDealtUntil(t, func(g *domain.FreeBetBlackjack) bool { return g.CanFreeDouble() })
	require.NoError(t, g.FreeDouble())

	var got struct {
		Hands []struct {
			Bet     int `json:"bet"`
			FreeBet int `json:"freeBet"`
			Score   int `json:"score"`
		} `json:"hands"`
		AnteBet int `json:"anteBet"`
		Chips   int `json:"chips"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	require.NotEmpty(t, got.Hands)
	assert.Equal(t, 50, got.Hands[0].Bet, "自分の賭け金が倍に化けている")
	assert.Equal(t, 50, got.Hands[0].FreeBet, "ハウス持ちの額が載っていない")
	assert.Equal(t, 50, got.AnteBet)
	assert.Equal(t, g.GetChips(), got.Chips)
}

// **可否はサーバが載せる。** ページに規則を作り直させない。
func TestFreeBetWebPresenter_FreeActionFlagsAreOnTheWire(t *testing.T) {
	cp := new(FreeBetBlackjackWebPresenter)
	g := freeBetDealtUntil(t, func(g *domain.FreeBetBlackjack) bool {
		return g.GetPhase() == domain.FreeBetPhasePlay
	})

	var got struct {
		CanFreeDouble bool `json:"canFreeDouble"`
		CanFreeSplit  bool `json:"canFreeSplit"`
		Phase         int  `json:"phase"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	assert.Equal(t, g.CanFreeDouble(), got.CanFreeDouble)
	assert.Equal(t, g.CanFreeSplit(), got.CanFreeSplit)
	assert.Equal(t, int(domain.FreeBetPhasePlay), got.Phase)
}

func TestFreeBetWebPresenter_HandsCarryTheirState(t *testing.T) {
	cp := new(FreeBetBlackjackWebPresenter)
	g := newFreeBetForPresenter(t)
	require.NoError(t, g.PlaceBet(50))
	for g.GetPhase() == domain.FreeBetPhasePlay {
		require.NoError(t, g.Stand())
	}

	var got struct {
		Hands []struct {
			Cards  []json.RawMessage `json:"cards"`
			Score  int               `json:"score"`
			Bet    int               `json:"bet"`
			Result int               `json:"result"`
		} `json:"hands"`
		DealerCards    []json.RawMessage `json:"dealerCards"`
		DealerScore    int               `json:"dealerScore"`
		DealerPushed22 bool              `json:"dealerPushed22"`
		Payout         int               `json:"payout"`
		RemainingCards int               `json:"remainingCards"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	require.NotEmpty(t, got.Hands)
	assert.Positive(t, got.Hands[0].Score)
	assert.Equal(t, 50, got.Hands[0].Bet)
	assert.NotZero(t, got.Hands[0].Result, "決着が載っていない")
	assert.GreaterOrEqual(t, len(got.DealerCards), 2)
	assert.Positive(t, got.DealerScore)
	assert.Equal(t, g.IsDealerPushed22(), got.DealerPushed22)
	assert.Equal(t, g.GetPayout(), got.Payout)
	assert.Equal(t, g.GetRemainingCards(), got.RemainingCards)
}

func TestFreeBetWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(FreeBetBlackjackWebPresenter)
	g := newFreeBetForPresenter(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)

	assert.JSONEq(t, `{"hint":null}`, cp.HintOutput(g))

	// **ナチュラルはその場で決着する。** プレイ待ちの配りを引き当ててから見る。
	inPlay := freeBetDealtUntil(t, func(g *domain.FreeBetBlackjack) bool {
		return g.GetPhase() == domain.FreeBetPhasePlay
	})
	var hint struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(inPlay)), &hint))
	assert.NotEmpty(t, hint.Action)
	assert.NotEmpty(t, hint.Reason)

	assert.NotEmpty(t, cp.ActionLogOutput(g))
}

// **チップが尽きたら messageCode で伝える。**
func TestFreeBetWebPresenter_BrokeCarriesAMessageCode(t *testing.T) {
	cp := new(FreeBetBlackjackWebPresenter)
	g := newFreeBetForPresenter(t)
	g.SetChips(50)
	require.NoError(t, g.PlaceBet(50))
	for g.GetPhase() == domain.FreeBetPhasePlay {
		require.NoError(t, g.Stand())
	}
	if !g.GetGameEndFlag() {
		t.Skip("この配りでは破産しなかった")
	}

	var got struct {
		MessageCode string `json:"messageCode"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	assert.Equal(t, "freebet.result.broke", got.MessageCode)
}
