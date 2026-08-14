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

// newMonteBankForPresenter は本物のドメインを返す。
//
// **モックではなく本物を使う。** アクセサをモックすると返す値を自分で決めて
// しまうので、「プレゼンタが盤面を正しく読めているか」の検査にならない。
func newMonteBankForPresenter(t *testing.T) *domain.MonteBank {
	t.Helper()
	g := domain.NewDefaultMonteBank()
	g.Reset()
	return g
}

// monteBankWithDup は場札に同じスートが 2 枚以上並ぶ局面を引き当てる。
func monteBankWithDup(t *testing.T) *domain.MonteBank {
	t.Helper()
	for range 1000 {
		g := newMonteBankForPresenter(t)
		for _, c := range g.GetLayout() {
			if g.SuitCountInLayout(c.GetDesign()) > 1 {
				return g
			}
		}
	}
	t.Fatalf("1000 回配っても重複したスートの場札が出なかった")
	return nil
}

// --- CUI ---

func TestMonteBankCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(MonteBankCuiPresenter)
	out := cp.Output(newMonteBankForPresenter(t), nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ラウンド:")
	assert.NotContains(t, out, "montebank.", "生の i18n キーが出力に混ざっている")
}

// **同じスートが何枚出ているかを必ず添える。** それが賭けの良し悪しを決める
// 唯一の数字なので、伏せると勝負が運任せになる。
func TestMonteBankCuiPresenter_AnnotatesEachLayoutCard(t *testing.T) {
	cp := new(MonteBankCuiPresenter)
	g := monteBankWithDup(t)
	out := cp.Output(g, nil)

	assert.Contains(t, out, "不利", "重複したスートに注意書きが無い")
	assert.NotContains(t, out, "montebank.")

	// 1 枚だけのスートがある局面では「互角」も出る。
	for range 1000 {
		c := newMonteBankForPresenter(t)
		lone := false
		for _, card := range c.GetLayout() {
			if c.SuitCountInLayout(card.GetDesign()) == 1 {
				lone = true
				break
			}
		}
		if lone {
			assert.Contains(t, cp.Output(c, nil), "互角")
			return
		}
	}
	t.Fatalf("1 枚だけのスートがある局面が出なかった")
}

func TestMonteBankCuiPresenter_ShowsGateAndResult(t *testing.T) {
	cp := new(MonteBankCuiPresenter)
	g := newMonteBankForPresenter(t)
	require.NoError(t, g.PlaceBet(0, 50))

	out := cp.Output(g, nil)
	assert.Contains(t, out, "ゲート:")
	assert.Contains(t, out, "収支")
	assert.NotContains(t, out, "montebank.")
}

func TestMonteBankCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(MonteBankCuiPresenter)
	g := newMonteBankForPresenter(t)
	assert.Contains(t, cp.Output(g, errors.New("賭け金が範囲外です")), "賭け金が範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	hint := cp.HintOutput(g)
	assert.NotEmpty(t, hint)
	assert.NotContains(t, hint, "montebank.", "助言のキーが訳されていない")

	require.NoError(t, g.PlaceBet(0, 50))
	assert.Equal(t, "いまは助言できる場面ではありません", cp.HintOutput(g))
}

// --- Web ---

func TestMonteBankWebPresenter_ArraysAndPick(t *testing.T) {
	cp := new(MonteBankWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newMonteBankForPresenter(t), nil)), &out))

	assert.NotEqual(t, "null", string(out["layout"]), "場札が null で返っている")
	assert.Equal(t, "-1", string(out["pick"]), "賭ける前の pick が -1 でない")
	_, hasGate := out["gate"]
	assert.False(t, hasGate, "賭ける前にゲートがワイヤに乗っている")
}

// **スートの枚数と「互角か」はサーバが載せる。** ページに数え直させない。
func TestMonteBankWebPresenter_SuitCountsAreOnTheWire(t *testing.T) {
	cp := new(MonteBankWebPresenter)
	g := monteBankWithDup(t)

	var got struct {
		Layout []struct {
			SuitCount       int  `json:"suitCount"`
			RemainingOfSuit int  `json:"remainingOfSuit"`
			IsEven          bool `json:"isEven"`
			IsPicked        bool `json:"isPicked"`
		} `json:"layout"`
		PayoutMultiplier int `json:"payoutMultiplier"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	require.Len(t, got.Layout, domain.MonteBankLayoutSize)
	assert.Equal(t, domain.MonteBankPayout, got.PayoutMultiplier)

	sawDup := false
	for i, card := range got.Layout {
		want := g.SuitCountInLayout(g.GetLayout()[i].GetDesign())
		assert.Equal(t, want, card.SuitCount, "場札 %d のスート枚数", i)
		assert.Equal(t, domain.MonteBankSuitSize-want, card.RemainingOfSuit, "場札 %d の残り", i)
		// **isEven は「1 枚だけ」と同値。** ここがずれると画面が嘘をつく。
		assert.Equal(t, want == 1, card.IsEven, "場札 %d の互角判定", i)
		assert.False(t, card.IsPicked, "賭ける前に印が付いている")
		if want > 1 {
			sawDup = true
		}
	}
	require.True(t, sawDup, "重複の無い局面を掴んでいる — 検査になっていない")
}

func TestMonteBankWebPresenter_MarksThePickAndGate(t *testing.T) {
	cp := new(MonteBankWebPresenter)
	g := newMonteBankForPresenter(t)
	require.NoError(t, g.PlaceBet(2, 50))

	var got struct {
		Layout []struct {
			IsPicked bool `json:"isPicked"`
		} `json:"layout"`
		Gate   *json.RawMessage `json:"gate"`
		Pick   int              `json:"pick"`
		Bet    int              `json:"bet"`
		Result int              `json:"result"`
		Payout int              `json:"payout"`
		Chips  int              `json:"chips"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Equal(t, 2, got.Pick)
	assert.True(t, got.Layout[2].IsPicked, "賭けた札に印が付いていない")
	assert.False(t, got.Layout[0].IsPicked)
	assert.NotNil(t, got.Gate, "決着後もゲートが載っていない")
	assert.Equal(t, 50, got.Bet)
	assert.NotEqual(t, int(domain.MonteBankResultNone), got.Result)
	assert.Equal(t, g.GetChips(), got.Chips)
	assert.Equal(t, g.GetPayout(), got.Payout)
}

func TestMonteBankWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(MonteBankWebPresenter)
	g := newMonteBankForPresenter(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	var hint struct {
		PickIdx int    `json:"pickIdx"`
		Reason  string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(g)), &hint))
	assert.GreaterOrEqual(t, hint.PickIdx, 0)
	assert.NotEmpty(t, hint.Reason)

	require.NoError(t, g.PlaceBet(0, 50))
	assert.JSONEq(t, `{"hint":null}`, cp.HintOutput(g))
}
