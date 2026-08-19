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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newCincinnatiForPresenter は本物のドメインを返す。
func newCincinnatiForPresenter(t *testing.T) *domain.Cincinnati {
	t.Helper()
	g := domain.NewDefaultCincinnati()
	g.Reset()
	return g
}

// cincinnatiSettled はショーダウンまで進めた卓を返す。
func cincinnatiSettled(t *testing.T) *domain.Cincinnati {
	t.Helper()
	g := newCincinnatiForPresenter(t)
	for steps := 0; g.GetPhase() == domain.CincinnatiPhaseBetting; steps++ {
		require.Less(t, steps, 200)
		if err := g.PlayerAction(domain.CincinnatiActionCheck, 0); err != nil {
			require.NoError(t, g.PlayerAction(domain.CincinnatiActionCall, 0))
		}
	}
	return g
}

// --- CUI ---

func TestCincinnatiCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(CincinnatiCuiPresenter)
	out := cp.Output(newCincinnatiForPresenter(t), nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ハンド:")
	assert.Contains(t, out, "場:")
	assert.NotContains(t, out, "cincinnati.", "生の i18n キーが出力に混ざっている")
}

// **CPU の手札は伏せたまま。** 5 枚もあるので見えたら勝負にならない。
func TestCincinnatiCuiPresenter_HidesCpuHandsUntilShowdown(t *testing.T) {
	cp := new(CincinnatiCuiPresenter)
	g := newCincinnatiForPresenter(t)
	out := cp.Output(g, nil)

	// 人間以外の席は伏せ表示。
	assert.Contains(t, out, "(伏せ)")
	// 人間の手札は出ている (伏せ表示が席数ぶんは無い)。
	assert.Equal(t, len(g.GetPlayers())-1, countOccurrences(out, "(伏せ)"),
		"伏せている席数が合わない")

	// ショーダウンでは全員開く。
	after := cp.Output(cincinnatiSettled(t), nil)
	assert.NotContains(t, after, "(伏せ)", "ショーダウンでも伏せたままになっている")
}

// countOccurrences は部分文字列の出現回数を返す。
func countOccurrences(s, sub string) int {
	n, idx := 0, 0
	for {
		i := indexFrom(s, sub, idx)
		if i < 0 {
			return n
		}
		n++
		idx = i + len(sub)
	}
}

func indexFrom(s, sub string, from int) int {
	if from >= len(s) {
		return -1
	}
	i := 0
	for i = from; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// **残り何枚めくれるかを出す。** 降りどころの判断に要る。
func TestCincinnatiCuiPresenter_ShowsHowManyCardsRemain(t *testing.T) {
	cp := new(CincinnatiCuiPresenter)
	out := cp.Output(newCincinnatiForPresenter(t), nil)
	assert.Contains(t, out, "0 / 5 枚公開")

	settled := cp.Output(cincinnatiSettled(t), nil)
	assert.Contains(t, settled, "5 / 5 枚公開")
}

func TestCincinnatiCuiPresenter_ShowsBetGuidanceAndResult(t *testing.T) {
	cp := new(CincinnatiCuiPresenter)
	out := cp.Output(newCincinnatiForPresenter(t), nil)
	// 賭けが無ければチェックできると出る。
	assert.True(t,
		countOccurrences(out, "チェックできます") > 0 || countOccurrences(out, "コールに") > 0,
		"賭けの案内が出ていない")

	settled := cp.Output(cincinnatiSettled(t), nil)
	assert.Contains(t, settled, "獲得", "決着の獲得額が出ていない")
	assert.NotContains(t, settled, "cincinnati.")
}

// **なぜその配当になったのかを言う** (#5780)。5 枚の手札だけで成立する役も
// 普通にあるので、金額だけでは読めない。
func TestCincinnatiCuiPresenter_NamesTheWinningHand(t *testing.T) {
	cp := new(CincinnatiCuiPresenter)
	g := cincinnatiSettled(t)
	out := cp.Output(g, nil)

	// 勝った席の役名がそのまま出ている。**盤面から引いて突き合わせる**ので、
	// 表を写した文字列にはならない。
	won := 0
	for i, r := range g.GetResults() {
		if r.WonAmount <= 0 {
			continue
		}
		won++
		rank := g.GetPlayers()[i].GetHandRank()
		require.GreaterOrEqual(t, rank, 0)
		require.Less(t, rank, len(domain.PokerHandNames))
		assert.Contains(t, out, i18n.Tf("cincinnati.wonLine",
			"name", g.GetPlayers()[i].GetName(),
			"amount", strconv.Itoa(r.WonAmount),
			"hand", domain.PokerHandNames[rank]))
	}
	require.Positive(t, won, "獲得した席が 1 つも無い")
}

func TestCincinnatiCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(CincinnatiCuiPresenter)
	g := newCincinnatiForPresenter(t)
	assert.Contains(t, cp.Output(g, errors.New("賭け金が範囲外です")), "賭け金が範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	hint := cp.HintOutput(g)
	assert.NotEmpty(t, hint)
	assert.NotContains(t, hint, "cincinnati.", "助言のキーが訳されていない")

	assert.Equal(t, "いまは助言できる場面ではありません", cp.HintOutput(cincinnatiSettled(t)))
}

// --- Web ---

func TestCincinnatiWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(CincinnatiWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newCincinnatiForPresenter(t), nil)), &out))
	assert.Equal(t, "[]", string(out["community"]), "配る前のコミュニティが配列で返っていない")
	assert.NotEqual(t, "null", string(out["seats"]))
}

// **CPU の手札をワイヤに乗せない。** 画面が出さなければよい、ではなく
// サーバが送らないことで守る。
func TestCincinnatiWebPresenter_DoesNotShipCpuHands(t *testing.T) {
	cp := new(CincinnatiWebPresenter)
	g := newCincinnatiForPresenter(t)

	var got struct {
		Seats []struct {
			IsHuman bool              `json:"isHuman"`
			Cards   []json.RawMessage `json:"cards"`
		} `json:"seats"`
		RevealedCount  int `json:"revealedCount"`
		CommunityTotal int `json:"communityTotal"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	require.Len(t, got.Seats, len(g.GetPlayers()))

	humans := 0
	for i, s := range got.Seats {
		if s.IsHuman {
			humans++
			assert.Len(t, s.Cards, domain.CincinnatiHoleCards, "人間の手札が届いていない")
			continue
		}
		assert.Empty(t, s.Cards, "席 %d (CPU) の手札がワイヤに乗っている", i)
	}
	require.Equal(t, 1, humans)
	assert.Zero(t, got.RevealedCount)
	assert.Equal(t, domain.CincinnatiCommunityCards, got.CommunityTotal)
}

// **ショーダウンでは全員の手札と役が載る。** 伏せる側と開く側の両方を踏む。
func TestCincinnatiWebPresenter_ShowsEveryHandAtShowdown(t *testing.T) {
	cp := new(CincinnatiWebPresenter)
	g := cincinnatiSettled(t)

	var got struct {
		Seats []struct {
			Cards    []json.RawMessage `json:"cards"`
			BestHand []json.RawMessage `json:"bestHand"`
			HandRank int               `json:"handRank"`
			Folded   bool              `json:"folded"`
		} `json:"seats"`
		Community     []json.RawMessage `json:"community"`
		RevealedCount int               `json:"revealedCount"`
		Pot           int               `json:"pot"`
		WinnerSeat    int               `json:"winnerSeat"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Len(t, got.Community, domain.CincinnatiCommunityCards)
	assert.Equal(t, domain.CincinnatiCommunityCards, got.RevealedCount)
	assert.Zero(t, got.Pot, "決着後にポットが残っている")
	for i, s := range got.Seats {
		assert.Len(t, s.Cards, domain.CincinnatiHoleCards, "席 %d の手札が開いていない", i)
		if !s.Folded {
			assert.Len(t, s.BestHand, domain.CincinnatiHandSize, "席 %d の最良 5 枚が載っていない", i)
		}
	}
	assert.Equal(t, g.WinnerSeat(), got.WinnerSeat)
}

// **賭けの状態はサーバが載せる。** ページに計算し直させない。
func TestCincinnatiWebPresenter_BettingStateIsOnTheWire(t *testing.T) {
	cp := new(CincinnatiWebPresenter)
	g := newCincinnatiForPresenter(t)

	var got struct {
		Pot         int  `json:"pot"`
		CurrentBet  int  `json:"currentBet"`
		ToCall      int  `json:"toCall"`
		RaiseCount  int  `json:"raiseCount"`
		CanRaise    bool `json:"canRaise"`
		TurnSeat    int  `json:"turnSeat"`
		HumanSeat   int  `json:"humanSeat"`
		IsHumanTurn bool `json:"isHumanTurn"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	assert.Equal(t, g.GetPot(), got.Pot)
	assert.Equal(t, g.GetToCall(), got.ToCall)
	assert.Equal(t, g.CanRaise(), got.CanRaise)
	assert.Equal(t, g.GetTurnSeat(), got.TurnSeat)
	assert.Equal(t, g.HumanSeat(), got.HumanSeat)
	assert.Equal(t, g.IsHumanTurn(), got.IsHumanTurn)
	assert.Positive(t, got.Pot, "アンティがポットに入っていない")
}

func TestCincinnatiWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(CincinnatiWebPresenter)
	g := newCincinnatiForPresenter(t)

	var withErr struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, errors.New("boom"))), &withErr))
	assert.Equal(t, "boom", withErr.Message)
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	var hint struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(g)), &hint))
	assert.NotEmpty(t, hint.Action)
	assert.NotEmpty(t, hint.Reason)

	assert.JSONEq(t, `{"hint":null}`, cp.HintOutput(cincinnatiSettled(t)))
}
