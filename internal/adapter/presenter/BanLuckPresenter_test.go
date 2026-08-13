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

// newBanLuckForPresenter は本物のドメインを返す。
//
// **モックではなく本物を使う。** アクセサを 20 個モックすると返す値を自分で
// 決めてしまうので、「プレゼンタが盤面を正しく読めているか」の検査にならない。
func newBanLuckForPresenter(t *testing.T) *domain.BanLuck {
	t.Helper()
	g := domain.NewDefaultBanLuck()
	g.Reset()
	return g
}

// banLuckDealt は配り終えた卓を返す。
func banLuckDealt(t *testing.T) *domain.BanLuck {
	t.Helper()
	for range 1000 {
		g := newBanLuckForPresenter(t)
		if err := g.PlaceBet(0); err != nil {
			continue
		}
		if g.GetPhase() == domain.BanLuckPhasePlay {
			return g
		}
	}
	t.Fatalf("1000 回配ってもプレイ待ちの局面が出なかった")
	return nil
}

// banLuckSettled は決着まで進めた卓を返す。
func banLuckSettled(t *testing.T) *domain.BanLuck {
	t.Helper()
	g := banLuckDealt(t)
	for steps := 0; g.GetPhase() == domain.BanLuckPhasePlay; steps++ {
		require.Less(t, steps, 50)
		if !g.IsHumanTurn() {
			g.CpuPlay()
			continue
		}
		if g.MustHit() {
			require.NoError(t, g.Hit())
			continue
		}
		require.NoError(t, g.Stand())
	}
	return g
}

// --- CUI ---

func TestBanLuckCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(BanLuckCuiPresenter)
	out := cp.Output(newBanLuckForPresenter(t), nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ラウンド:")
	assert.Contains(t, out, "親:")
	assert.NotContains(t, out, "banluck.", "生の i18n キーが出力に混ざっている")
}

// **親が誰かは常に見えていないといけない。** 役割が回るゲームなので、
// 「いま自分が親かどうか」で打ち方がまるごと変わる。
func TestBanLuckCuiPresenter_MarksTheBanker(t *testing.T) {
	cp := new(BanLuckCuiPresenter)
	g := banLuckDealt(t)
	out := cp.Output(g, nil)

	banker := g.GetPlayers()[g.GetBankerSeat()].GetName()
	assert.Contains(t, out, "親: "+banker)
	assert.Contains(t, out, "（親）")
	assert.NotContains(t, out, "banluck.")
}

// **親の義務は名指しで出す。** 拒否されたことだけ伝わって規則が伝わらないのを防ぐ。
func TestBanLuckCuiPresenter_NamesTheBankerObligation(t *testing.T) {
	cp := new(BanLuckCuiPresenter)
	// 人間が親で 15 未満、という局面を引き当てる。
	var g *domain.BanLuck
	for range 1000 {
		c := newBanLuckForPresenter(t)
		if err := c.PlaceBet(0); err != nil {
			continue
		}
		if c.GetPhase() != domain.BanLuckPhasePlay {
			continue
		}
		for !c.IsHumanTurn() && c.GetPhase() == domain.BanLuckPhasePlay {
			c.CpuPlay()
		}
		if c.MustHit() {
			g = c
			break
		}
	}
	require.NotNil(t, g, "1000 回配っても親の義務が発生する局面が出なかった")

	out := cp.Output(g, nil)
	assert.Contains(t, out, "15未満")
	assert.NotContains(t, out, "banluck.")

	// 義務が無いときは出さない (両側を踏む)。
	plain := cp.Output(newBanLuckForPresenter(t), nil)
	assert.NotContains(t, plain, "15未満")
}

func TestBanLuckCuiPresenter_ShowsResults(t *testing.T) {
	cp := new(BanLuckCuiPresenter)
	out := cp.Output(banLuckSettled(t), nil)

	assert.Contains(t, out, "→", "収支が出ていない")
	assert.NotContains(t, out, "banluck.")
	// 役の名前が生キーでなく訳されている。
	for _, name := range []string{"バスト", "通常", "ファイブドラゴン", "バンラック", "バンバン"} {
		if assert.NotContains(t, out, "rank."+name) {
			continue
		}
	}
}

func TestBanLuckCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(BanLuckCuiPresenter)
	g := newBanLuckForPresenter(t)
	assert.Contains(t, cp.Output(g, errors.New("賭け金が範囲外です")), "賭け金が範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))
	assert.NotEmpty(t, cp.HintOutput(g))

	inPlay := banLuckDealt(t)
	for !inPlay.IsHumanTurn() && inPlay.GetPhase() == domain.BanLuckPhasePlay {
		inPlay.CpuPlay()
	}
	if inPlay.IsHumanTurn() {
		hint := cp.HintOutput(inPlay)
		assert.NotEmpty(t, hint)
		assert.NotContains(t, hint, "banluck.", "助言のキーが訳されていない")
	}
}

// --- Web ---

func TestBanLuckWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(BanLuckWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newBanLuckForPresenter(t), nil)), &out))
	assert.NotEqual(t, "null", string(out["seats"]))

	var seats []struct {
		Cards []json.RawMessage `json:"cards"`
	}
	require.NoError(t, json.Unmarshal(out["seats"], &seats))
	require.NotEmpty(t, seats)
	for i := range seats {
		assert.NotNil(t, seats[i].Cards, "席 %d の札が null で返っている", i)
	}
}

// **席の役割と手番はサーバが載せる。** ページに計算し直させない。
func TestBanLuckWebPresenter_RolesAreOnTheWire(t *testing.T) {
	cp := new(BanLuckWebPresenter)
	g := banLuckDealt(t)

	var got struct {
		Seats []struct {
			Name     string `json:"name"`
			IsHuman  bool   `json:"isHuman"`
			IsBanker bool   `json:"isBanker"`
			IsTurn   bool   `json:"isTurn"`
			Score    int    `json:"score"`
		} `json:"seats"`
		BankerSeat  int  `json:"bankerSeat"`
		TurnSeat    int  `json:"turnSeat"`
		HumanSeat   int  `json:"humanSeat"`
		IsHumanTurn bool `json:"isHumanTurn"`
		MustHit     bool `json:"mustHit"`
		Phase       int  `json:"phase"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	require.Len(t, got.Seats, len(g.GetPlayers()))
	assert.Equal(t, g.GetBankerSeat(), got.BankerSeat)
	assert.Equal(t, g.GetTurnSeat(), got.TurnSeat)
	assert.Equal(t, g.GetHumanSeat(), got.HumanSeat)
	assert.Equal(t, g.IsHumanTurn(), got.IsHumanTurn)
	assert.Equal(t, g.MustHit(), got.MustHit)

	// **親はちょうど 1 人。** ここが 0 人や 2 人になる盤面は作れないこと。
	bankers := 0
	for i, s := range got.Seats {
		if s.IsBanker {
			bankers++
			assert.Equal(t, got.BankerSeat, i, "親フラグと親の添字が食い違っている")
		}
		assert.Equal(t, g.GetPlayers()[i].GetIsHuman(), s.IsHuman, "席 %d の人間フラグ", i)
		assert.Positive(t, s.Score, "席 %d の点数が載っていない", i)
	}
	assert.Equal(t, 1, bankers, "親が %d 人いる", bankers)
}

func TestBanLuckWebPresenter_ResultsAreOnTheWire(t *testing.T) {
	cp := new(BanLuckWebPresenter)
	g := banLuckSettled(t)

	var got struct {
		Seats []struct {
			Rank     int `json:"rank"`
			Outcome  int `json:"outcome"`
			RoundBet int `json:"roundBet"`
			Delta    int `json:"delta"`
			Chips    int `json:"chips"`
		} `json:"seats"`
		Phase      int `json:"phase"`
		WinnerSeat int `json:"winnerSeat"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Equal(t, int(domain.BanLuckPhaseRoundEnd), got.Phase)
	assert.Equal(t, g.WinnerSeat(), got.WinnerSeat)
	// **収支の合計は 0。** 親と子の間でしか動かない。
	sum := 0
	for i, s := range got.Seats {
		sum += s.Delta
		assert.Equal(t, g.GetPlayers()[i].GetChips(), s.Chips, "席 %d のチップ", i)
		assert.Equal(t, int(g.GetResults()[i].Rank), s.Rank, "席 %d の役", i)
	}
	assert.Zero(t, sum, "席ごとの収支の合計が 0 でない")

	// **精算後は席の bet が 0 に戻る。** 表示に要る額は roundBet 側で運ぶ。
	staked := 0
	for i, s := range got.Seats {
		if i != g.GetBankerSeat() {
			staked += s.RoundBet
		}
	}
	assert.Positive(t, staked, "そのラウンドの賭け金がワイヤに載っていない")
}

func TestBanLuckWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(BanLuckWebPresenter)
	g := newBanLuckForPresenter(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)

	assert.JSONEq(t, `{"hint":null}`, cp.HintOutput(g))
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	inPlay := banLuckDealt(t)
	for !inPlay.IsHumanTurn() && inPlay.GetPhase() == domain.BanLuckPhasePlay {
		inPlay.CpuPlay()
	}
	if inPlay.IsHumanTurn() {
		var hint struct {
			Action string `json:"action"`
			Reason string `json:"reason"`
		}
		require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(inPlay)), &hint))
		assert.NotEmpty(t, hint.Action)
		assert.NotEmpty(t, hint.Reason)
	}
}
