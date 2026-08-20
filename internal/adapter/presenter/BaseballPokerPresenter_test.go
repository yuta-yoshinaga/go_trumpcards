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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newBaseballForPresenter は本物のドメインを返す。
func newBaseballForPresenter(t *testing.T) *domain.BaseballPoker {
	t.Helper()
	g := domain.NewDefaultBaseballPoker()
	g.Reset()
	return g
}

// baseballAtBetting は必ずベット中の卓を返す (配りによっては買い増しから始まる)。
func baseballAtBetting(t *testing.T) *domain.BaseballPoker {
	t.Helper()
	for range 60 {
		g := newBaseballForPresenter(t)
		if g.GetPhase() == domain.BaseballPhaseBetting {
			return g
		}
	}
	t.Fatalf("60 回配ってもベット開始の局面が出なかった")
	return nil
}

// baseballAtBuyIn は必ず買い増しの場面の卓を返す。
func baseballAtBuyIn(t *testing.T) *domain.BaseballPoker {
	t.Helper()
	for range 60 {
		g := newBaseballForPresenter(t)
		for steps := 0; g.GetPhase() != domain.BaseballPhaseShowdown && !g.GetGameEndFlag(); steps++ {
			require.Less(t, steps, 400)
			if g.IsHumanBuying() {
				return g
			}
			if g.GetPhase() == domain.BaseballPhaseBuyIn {
				g.CpuPlay()
				continue
			}
			if g.IsHumanTurn() {
				if err := g.PlayerAction(domain.BaseballActionCheck, 0); err != nil {
					require.NoError(t, g.PlayerAction(domain.BaseballActionCall, 0))
				}
				continue
			}
			g.CpuPlay()
		}
	}
	t.Fatalf("60 回配っても人間が買い増しを迫られる局面が出なかった")
	return nil
}

// baseballSettled はショーダウンまで進めた卓を返す。
func baseballSettled(t *testing.T) *domain.BaseballPoker {
	t.Helper()
	g := newBaseballForPresenter(t)
	for steps := 0; g.GetPhase() != domain.BaseballPhaseShowdown && !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 400)
		switch {
		case g.IsHumanBuying():
			require.NoError(t, g.AnswerBuyIn(domain.BaseballBuyPay))
		case g.IsHumanTurn():
			if err := g.PlayerAction(domain.BaseballActionCheck, 0); err != nil {
				require.NoError(t, g.PlayerAction(domain.BaseballActionCall, 0))
			}
		default:
			g.CpuPlay()
		}
	}
	return g
}

// --- CUI ---

func TestBaseballPokerCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(BaseballPokerCuiPresenter)
	out := cp.Output(newBaseballForPresenter(t), nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ハンド:")
	assert.Contains(t, out, "ストリート:")
	assert.NotContains(t, out, "baseballpoker.", "生の i18n キーが出力に混ざっている")
}

// **表札は全員に見える。それがスタッドの読み合いの材料。** 伏せ札だけを隠す。
func TestBaseballPokerCuiPresenter_ShowsUpCardsAndHidesDownCards(t *testing.T) {
	cp := new(BaseballPokerCuiPresenter)
	g := newBaseballForPresenter(t)
	out := cp.Output(g, nil)

	// **隠す枚数は盤面から数える。** 3rd の表札が 4 だった席は既にボーナスの
	// 伏せ札を 1 枚持っているので、「1 席 2 枚」と決め打つと配り依存で落ちる。
	wantHidden := 0
	for _, p := range g.GetPlayers() {
		if p.GetIsHuman() {
			continue
		}
		for _, up := range p.GetFaceUp() {
			if !up {
				wantHidden++
			}
		}
	}
	require.Positive(t, wantHidden, "CPU の伏せ札が 1 枚も無い")
	assert.Equal(t, wantHidden, countOccurrences(out, "[??]"),
		"隠れている伏せ札の枚数が盤面と合わない")
	for i, p := range g.GetPlayers() {
		if p.GetIsHuman() {
			continue
		}
		up := p.FaceUpCards()
		require.NotEmpty(t, up, "席 %d に表札が無い", i)
		assert.Positive(t, countOccurrences(out, cuiCardStr(up[0])),
			"席 %d の表札が出力に出ていない", i)
	}

	// ショーダウンでは全部開く。
	after := cp.Output(baseballSettled(t), nil)
	assert.NotContains(t, after, "[??]", "ショーダウンでも伏せたままになっている")
}

// **ワイルドと 2 つのイベントを画面に出す。** 知らないと降りどころが決まらない。
func TestBaseballPokerCuiPresenter_ExplainsTheWildsAndEvents(t *testing.T) {
	cp := new(BaseballPokerCuiPresenter)
	out := cp.Output(newBaseballForPresenter(t), nil)
	// **文言もドメインの値から作る** (#5782)。行そのものを組み立てて突き合わせる
	// ——"3" も "4" も盤面の札として出るので、部分一致では素通りする。
	assert.Contains(t, out, i18n.Tf("baseballpoker.wildLine",
		"wilds", baseballPokerWildValuesStr(),
		"bonus", strconv.Itoa(domain.BaseballBonusFour),
		"buyIn", strconv.Itoa(domain.BaseballWildThree)))
	assert.Contains(t, out, "追加札")
	assert.Contains(t, out, "買い増し")
}

// **ワイルドの並びは判定そのものから作る。** 写した定数だと片方だけ古くなる。
func TestBaseballPokerWildValuesStrFollowsTheEvaluator(t *testing.T) {
	got := baseballPokerWildValuesStr()
	want := make([]string, 0, 2)
	for v := 1; v <= 13; v++ {
		if domain.BaseballIsWild(domain.NewCard(domain.CardDesignSpade, v, false)) {
			want = append(want, strconv.Itoa(v))
		}
	}
	require.Len(t, want, 2, "ワイルドは 2 種のはず")
	assert.Equal(t, strings.Join(want, " / "), got)
}

// **買い増しの場面ではそう言い、額まで出す。**
func TestBaseballPokerCuiPresenter_AsksToBuyThePot(t *testing.T) {
	cp := new(BaseballPokerCuiPresenter)
	g := baseballAtBuyIn(t)
	out := cp.Output(g, nil)

	assert.Contains(t, out, "表の3が出ました")
	assert.Contains(t, out, "pay")
	assert.Contains(t, out, "buyfold")
	// ベットの案内は出さない。
	assert.NotContains(t, out, "チェックできます", "買い増しの場面でベットの案内が残っている")
}

func TestBaseballPokerCuiPresenter_ShowsBetGuidanceAndResult(t *testing.T) {
	cp := new(BaseballPokerCuiPresenter)
	out := cp.Output(baseballAtBetting(t), nil)
	assert.True(t,
		countOccurrences(out, "チェックできます") > 0 || countOccurrences(out, "コールに") > 0,
		"賭けの案内が出ていない")

	settled := cp.Output(baseballSettled(t), nil)
	assert.Contains(t, settled, "獲得", "決着の獲得額が出ていない")
	assert.NotContains(t, settled, "baseballpoker.")
}

func TestBaseballPokerCuiPresenter_ErrorsAndHint(t *testing.T) {
	cp := new(BaseballPokerCuiPresenter)
	g := baseballAtBetting(t)
	assert.Contains(t, cp.Output(g, errors.New("賭け金が範囲外です")), "賭け金が範囲外です")
	assert.NotEmpty(t, cp.ActionLogOutput(g))

	hint := cp.HintOutput(g)
	assert.NotEmpty(t, hint)
	assert.NotContains(t, hint, "baseballpoker.", "助言のキーが訳されていない")

	// 買い増しの場面では買い増しの助言になる。
	buying := cp.HintOutput(baseballAtBuyIn(t))
	assert.NotContains(t, buying, "baseballpoker.")
	assert.NotEmpty(t, buying)
}

// --- Web ---

func TestBaseballPokerWebPresenter_ArraysAreNeverNull(t *testing.T) {
	cp := new(BaseballPokerWebPresenter)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newBaseballForPresenter(t), nil)), &out))
	for _, key := range []string{"seats", "wildValues"} {
		assert.NotEqual(t, "null", string(out[key]), "%s が null で返っている", key)
	}
}

// **伏せ札はワイヤに乗せず、表札は乗せる。** どちらか一方だけの検査は、
// 壊れているときに通る。
func TestBaseballPokerWebPresenter_ShipsUpCardsOnly(t *testing.T) {
	cp := new(BaseballPokerWebPresenter)
	g := newBaseballForPresenter(t)

	var got struct {
		Seats []struct {
			IsHuman bool              `json:"isHuman"`
			Cards   []json.RawMessage `json:"cards"`
			FaceUp  []bool            `json:"faceUp"`
		} `json:"seats"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))
	require.Len(t, got.Seats, len(g.GetPlayers()))

	humans := 0
	for i, s := range got.Seats {
		// 向きの列は必ず手札と同じ長さ。
		require.Len(t, s.FaceUp, len(s.Cards), "席 %d の向きの列がずれている", i)
		if s.IsHuman {
			humans++
			for k, c := range s.Cards {
				assert.NotEqual(t, "null", string(c), "自分の %d 枚目が届いていない", k)
			}
			continue
		}
		for k, c := range s.Cards {
			if s.FaceUp[k] {
				assert.NotEqual(t, "null", string(c), "席 %d の表札が届いていない", i)
				continue
			}
			assert.Equal(t, "null", string(c), "席 %d の伏せ札がワイヤに乗っている", i)
		}
	}
	require.Equal(t, 1, humans)
}

// **ショーダウンでは全員の手札と役が載る。**
func TestBaseballPokerWebPresenter_ShowsEveryHandAtShowdown(t *testing.T) {
	cp := new(BaseballPokerWebPresenter)
	g := baseballSettled(t)

	var got struct {
		Seats []struct {
			Cards    []json.RawMessage `json:"cards"`
			BestHand []json.RawMessage `json:"bestHand"`
			Folded   bool              `json:"folded"`
		} `json:"seats"`
		Pot        int `json:"pot"`
		WinnerSeat int `json:"winnerSeat"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(g, nil)), &got))

	assert.Zero(t, got.Pot, "決着後にポットが残っている")
	for i, s := range got.Seats {
		for k, c := range s.Cards {
			assert.NotEqual(t, "null", string(c), "席 %d の %d 枚目が開いていない", i, k)
		}
		if !s.Folded && len(s.Cards) >= domain.BaseballHandSize {
			assert.Len(t, s.BestHand, domain.BaseballHandSize, "席 %d の最良 5 枚が載っていない", i)
		}
	}
	assert.Equal(t, g.WinnerSeat(), got.WinnerSeat)
}

// **ワイルドとイベントの値はサーバが載せる。** ページに 3 と 9 を書き写させると、
// 役の判定と画面の印が別々に育って食い違う。
func TestBaseballPokerWebPresenter_ShipsTheWildAndEventValues(t *testing.T) {
	cp := new(BaseballPokerWebPresenter)

	var got struct {
		WildValues  []int `json:"wildValues"`
		BonusValue  int   `json:"bonusValue"`
		BuyInValue  int   `json:"buyInValue"`
		Street      int   `json:"street"`
		StreetTotal int   `json:"streetTotal"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(newBaseballForPresenter(t), nil)), &got))
	assert.ElementsMatch(t, []int{domain.BaseballWildThree, domain.BaseballWildNine}, got.WildValues)
	assert.Equal(t, domain.BaseballBonusFour, got.BonusValue)
	assert.Equal(t, domain.BaseballWildThree, got.BuyInValue)
	assert.Equal(t, 1, got.Street)
	assert.Equal(t, domain.BaseballUpCards, got.StreetTotal)
}

// **買い増しの状態はサーバが載せる。** ページにフェーズ番号から割り出させない。
func TestBaseballPokerWebPresenter_FlagsTheBuyIn(t *testing.T) {
	cp := new(BaseballPokerWebPresenter)

	var got struct {
		IsBuying  bool `json:"isBuying"`
		BuyerSeat int  `json:"buyerSeat"`
		BuyCost   int  `json:"buyCost"`
		Phase     int  `json:"phase"`
		Seats     []struct {
			IsBuying bool `json:"isBuying"`
		} `json:"seats"`
	}
	require.NoError(t, json.Unmarshal([]byte(cp.Output(baseballAtBuyIn(t), nil)), &got))
	assert.True(t, got.IsBuying, "買い増しの場面が isBuying に出ていない")
	assert.Equal(t, int(domain.BaseballPhaseBuyIn), got.Phase)
	assert.GreaterOrEqual(t, got.BuyerSeat, 0)
	assert.Positive(t, got.BuyCost, "払う額が載っていない")
	assert.True(t, got.Seats[got.BuyerSeat].IsBuying, "席の側に買い増しの印が無い")

	require.NoError(t, json.Unmarshal([]byte(cp.Output(baseballAtBetting(t), nil)), &got))
	assert.False(t, got.IsBuying, "ベット中に買い増しの印が立っている")
	assert.Equal(t, -1, got.BuyerSeat, "買い手がいないのに席番号が立っている")
}

// **賭けの状態はサーバが載せる。** ページに計算し直させない。
func TestBaseballPokerWebPresenter_BettingStateIsOnTheWire(t *testing.T) {
	cp := new(BaseballPokerWebPresenter)
	g := baseballAtBetting(t)

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

func TestBaseballPokerWebPresenter_ErrorAndHint(t *testing.T) {
	cp := new(BaseballPokerWebPresenter)
	g := baseballAtBetting(t)

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

	// 買い増しの場面では pay / fold を薦める。
	require.NoError(t, json.Unmarshal([]byte(cp.HintOutput(baseballAtBuyIn(t))), &hint))
	assert.Contains(t, []string{"pay", "fold"}, hint.Action)
}
