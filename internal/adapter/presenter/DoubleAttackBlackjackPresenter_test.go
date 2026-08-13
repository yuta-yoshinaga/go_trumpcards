//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newDoubleAttackForPresenter は本物のドメインを返す。
//
// **モックではなく本物を使う。** アクセサを 25 個モックすると返す値を自分で決めて
// しまうので、「プレゼンタが盤面を正しく読めているか」の検査にならない。
func newDoubleAttackForPresenter(t *testing.T) *domain.DoubleAttackBlackjack {
	t.Helper()
	g := domain.NewDefaultDoubleAttackBlackjack()
	g.Reset()
	return g
}

// --- CUI ---

func TestDoubleAttackCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(DoubleAttackBlackjackCuiPresenter)
	out := cp.Output(newDoubleAttackForPresenter(t), nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ラウンド:")
	assert.NotContains(t, out, "doubleattack.", "生の i18n キーが出力に混ざっている")
}

// **追加ベットの前はアップカードだけを見せ、そう明示する。**
func TestDoubleAttackCuiPresenter_ShowsOnlyTheUpCardBeforeTheAttack(t *testing.T) {
	cp := new(DoubleAttackBlackjackCuiPresenter)
	g := newDoubleAttackForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 0))

	out := cp.Output(g, nil)
	assert.Contains(t, out, "2枚目は追加ベットの後", "情報の順序が画面で説明されていない")
	assert.Contains(t, out, "追加ベットの上限")
	assert.NotContains(t, out, "doubleattack.")

	require.NoError(t, g.Attack(0))
	after := cp.Output(g, nil)
	assert.NotContains(t, after, "2枚目は追加ベットの後", "配った後も伏せ字のまま")
}

func TestDoubleAttackCuiPresenter_ShowsBetsAndResult(t *testing.T) {
	cp := new(DoubleAttackBlackjackCuiPresenter)
	g := newDoubleAttackForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 20))
	require.NoError(t, g.Attack(50))
	for g.GetPhase() == domain.DoubleAttackPhasePlay {
		require.NoError(t, g.Stand())
	}

	out := cp.Output(g, nil)
	assert.Contains(t, out, "アンティ 50")
	assert.Contains(t, out, "追加ベット 50")
	assert.Contains(t, out, "Bust It 20")
	assert.Contains(t, out, "収支:")
	assert.NotContains(t, out, "doubleattack.")
}

func TestDoubleAttackCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(DoubleAttackBlackjackCuiPresenter)
	g := newDoubleAttackForPresenter(t)
	assert.Contains(t, cp.Output(g, errors.New("アンティが範囲外です")), "アンティが範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))
	assert.NotEmpty(t, cp.HintOutput(g))

	require.NoError(t, g.PlaceBet(50, 0))
	hint := cp.HintOutput(g)
	assert.NotEmpty(t, hint)
	assert.NotContains(t, hint, "doubleattack.", "助言のキーが訳されていない")
}

// --- Web ---

func TestDoubleAttackWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(DoubleAttackBlackjackWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newDoubleAttackForPresenter(t), nil)), &out))
	for _, key := range []string{"hands", "dealerCards"} {
		assert.Equal(t, "[]", string(out[key]), "%s が配列で返っていない", key)
	}
}

// **追加ベットの前は、ディーラーの札が 1 枚しかワイヤに乗らない。**
//
// サーバがそもそも 2 枚目を持っていないので伏せ字ではなく不在。点数も出さない
// (1 枚だけの点数はホールカードの手掛かりになる)。
func TestDoubleAttackWebPresenter_OnlyTheUpCardBeforeTheAttack(t *testing.T) {
	cp := new(DoubleAttackBlackjackWebPresenter)
	g := newDoubleAttackForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 0))

	var got struct {
		Phase           int               `json:"phase"`
		DealerCards     []json.RawMessage `json:"dealerCards"`
		DealerScore     int               `json:"dealerScore"`
		DealerHoleDealt bool              `json:"dealerHoleDealt"`
		MaxAttackBet    int               `json:"maxAttackBet"`
		Hands           []json.RawMessage `json:"hands"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Equal(t, int(domain.DoubleAttackPhaseAttack), got.Phase)
	assert.Len(t, got.DealerCards, 1, "追加ベットの前に 2 枚目がワイヤに乗っている")
	assert.False(t, got.DealerHoleDealt)
	assert.Zero(t, got.DealerScore, "アップカードだけの点数が漏れている")
	assert.Equal(t, 50, got.MaxAttackBet)
	assert.Len(t, got.Hands, 1)

	require.NoError(t, g.Attack(0))
	var after struct {
		DealerCards     []json.RawMessage `json:"dealerCards"`
		DealerHoleDealt bool              `json:"dealerHoleDealt"`
		DealerScore     int               `json:"dealerScore"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &after))
	assert.True(t, after.DealerHoleDealt)
	assert.GreaterOrEqual(t, len(after.DealerCards), 2)
	assert.Positive(t, after.DealerScore)
}

// **上限と可否はサーバが載せる。** ページに計算し直させない。
func TestDoubleAttackWebPresenter_LimitsAreOnTheWire(t *testing.T) {
	cp := new(DoubleAttackBlackjackWebPresenter)
	g := newDoubleAttackForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 0))

	var got struct {
		MaxAttackBet int  `json:"maxAttackBet"`
		CanDouble    bool `json:"canDouble"`
		CanSplit     bool `json:"canSplit"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	assert.Equal(t, g.MaxAttackBet(), got.MaxAttackBet)
	assert.Equal(t, g.CanDouble(), got.CanDouble)
	assert.Equal(t, g.CanSplit(), got.CanSplit)
	// 追加ベットの上限はアンティを超えない。
	assert.LessOrEqual(t, got.MaxAttackBet, g.GetAnteBet())
}

func TestDoubleAttackWebPresenter_HandsCarryTheirState(t *testing.T) {
	cp := new(DoubleAttackBlackjackWebPresenter)
	g := newDoubleAttackForPresenter(t)
	require.NoError(t, g.PlaceBet(50, 20))
	require.NoError(t, g.Attack(50))
	for g.GetPhase() == domain.DoubleAttackPhasePlay {
		require.NoError(t, g.Stand())
	}

	var got struct {
		Hands []struct {
			Cards  []json.RawMessage `json:"cards"`
			Score  int               `json:"score"`
			Bet    int               `json:"bet"`
			Result int               `json:"result"`
		} `json:"hands"`
		AnteBet   int `json:"anteBet"`
		AttackBet int `json:"attackBet"`
		BustItBet int `json:"bustItBet"`
		Chips     int `json:"chips"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	require.NotEmpty(t, got.Hands)
	assert.Positive(t, got.Hands[0].Score)
	assert.Equal(t, 100, got.Hands[0].Bet, "アンティ + 追加ベットが手札の賭け金になっていない")
	assert.NotZero(t, got.Hands[0].Result, "決着が載っていない")
	assert.Equal(t, 50, got.AnteBet)
	assert.Equal(t, 50, got.AttackBet)
	assert.Equal(t, 20, got.BustItBet)
	assert.Equal(t, g.GetChips(), got.Chips)
}

func TestDoubleAttackWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(DoubleAttackBlackjackWebPresenter)
	g := newDoubleAttackForPresenter(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)

	assert.JSONEq(t, `{"hint":null}`, cp.HintOutput(g))

	require.NoError(t, g.PlaceBet(50, 0))
	var hint struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(g)), &hint))
	assert.NotEmpty(t, hint.Action)
	assert.NotEmpty(t, hint.Reason)

	assert.NotEmpty(t, cp.ActionLogOutput(g))
}

// **スプリットしても収支が正しい。**
//
// 以前は「アンティ + 追加ベット」を土台にして手札ごとの差分を足していたため、
// 分割すると両方の手札が土台と同額になって差分が 0 になり、**2 つ目の手札を
// 作るのに払ったぶんが消えていた**。アンティ 50 で分割して両方プッシュ
// (= 収支 0) が **+50 の黒字**として表示される、という嘘をつく。
func TestDoubleAttackCuiPresenter_NetIsCorrectAfterASplit(t *testing.T) {
	cp := new(DoubleAttackBlackjackCuiPresenter)
	g := newDoubleAttackForPresenter(t)
	// **アンティを置く前のチップを基準にする。** PlaceBet の後で取ると、
	// アンティぶんが基準に織り込まれてしまい、収支の定義とずれる。
	before := g.GetChips()
	require.NoError(t, g.PlaceBet(50, 0))
	require.NoError(t, g.Attack(0))

	// 8 のペア対ディーラーのハード 17。配りに依存させず局面を組む。
	g.SetupSplitForTest(8)
	require.True(t, g.CanSplit())
	require.NoError(t, g.Split())
	for g.GetPhase() == domain.DoubleAttackPhasePlay {
		require.NoError(t, g.Stand())
	}
	require.Equal(t, domain.DoubleAttackPhaseResult, g.GetPhase())

	// **画面の収支は、実際のチップの増減と一致しなければならない。**
	// Split で引かれた分も含む。
	wantNet := g.GetChips() - before
	out := cp.Output(g, nil)

	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "収支:") {
			found = true
			assert.Contains(t, line, strconv.Itoa(wantNet),
				"画面の収支が実際のチップの増減 (%d) と違う: %s", wantNet, line)
		}
	}
	assert.True(t, found, "収支の行が出ていない")
}

// 分割していないときも収支はチップの増減と一致する (負のコントロール)。
func TestDoubleAttackCuiPresenter_NetMatchesChipsWithoutASplit(t *testing.T) {
	cp := new(DoubleAttackBlackjackCuiPresenter)
	g := newDoubleAttackForPresenter(t)
	before := g.GetChips()
	require.NoError(t, g.PlaceBet(50, 20))
	require.NoError(t, g.Attack(50))
	for g.GetPhase() == domain.DoubleAttackPhasePlay {
		require.NoError(t, g.Stand())
	}

	wantNet := g.GetChips() - before
	out := cp.Output(g, nil)
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "収支:") {
			found = true
			assert.Contains(t, line, strconv.Itoa(wantNet),
				"画面の収支が実際のチップの増減 (%d) と違う: %s", wantNet, line)
		}
	}
	assert.True(t, found, "収支の行が出ていない")
}
