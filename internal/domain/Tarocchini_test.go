//go:build test

package domain

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tarocchiniTrick(entries ...[2]int) []*TrickCard {
	trick := make([]*TrickCard, 0, len(entries))
	for seat, e := range entries {
		trick = append(trick, &TrickCard{PlayerIdx: seat, Card: NewCard(e[0], e[1], false)})
	}
	return trick
}

func TestTarocchini_DeckIsSixtyTwoDistinctCards(t *testing.T) {
	deck := buildTarocchiniDeck()
	assert.Len(t, deck, TarocchiniDeckSize)
	assert.Equal(t, 62, TarocchiniDeckSize)

	seen := map[[2]int]bool{}
	suits, trumps, matto := 0, 0, 0
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "duplicate card %v", key)
		seen[key] = true
		switch {
		case tarocchiniIsMatto(c):
			matto++
		case tarocchiniIsTrump(c):
			trumps++
		default:
			suits++
			// **2..5 は抜かれている。**52 枚デッキの感覚で 1..10 を作ると枚数が合わない。
			assert.NotContains(t, []int{2, 3, 4, 5}, c.GetValue(),
				"low pips must not be in a Bolognese deck")
		}
	}
	assert.Equal(t, 40, suits)
	assert.Equal(t, TarocchiniMaxTrump, trumps)
	assert.Equal(t, 1, matto)
}

// 62 は 4 で割り切れないので、余りが出ることを構造として固定しておく。
func TestTarocchini_DealLeavesASurplus(t *testing.T) {
	assert.Equal(t, 2, TarocchiniSurplus)
	assert.Equal(t, TarocchiniDeckSize, TarocchiniPlayerCnt*TarocchiniHandSize+TarocchiniSurplus)
}

func TestTarocchini_TeamsAreOpposite(t *testing.T) {
	assert.Equal(t, TarocchiniTeamOf(0), TarocchiniTeamOf(2))
	assert.Equal(t, TarocchiniTeamOf(1), TarocchiniTeamOf(3))
	assert.NotEqual(t, TarocchiniTeamOf(0), TarocchiniTeamOf(1))
}

// **これがこのゲームの核心。**同格のパパが複数出たら後から出した方が勝つ。
// 共通の「厳密に強い札だけが勝者を更新する」判定では先に出した方が勝ってしまう。
func TestTarocchini_LaterPapaBeatsEarlierPapa(t *testing.T) {
	led := 1
	cases := []struct {
		name  string
		trick []*TrickCard
		want  int
	}{
		{"two papi: the later one wins", tarocchiniTrick(
			[2]int{TarocchiniTrumpDesign, 2},
			[2]int{led, 14},
			[2]int{TarocchiniTrumpDesign, 5},
			[2]int{led, 13},
		), 2},
		{"the papi order on the table does not matter, only lateness", tarocchiniTrick(
			[2]int{TarocchiniTrumpDesign, 5},
			[2]int{led, 14},
			[2]int{TarocchiniTrumpDesign, 2},
			[2]int{led, 13},
		), 2},
		{"all four papi: the last one wins", tarocchiniTrick(
			[2]int{TarocchiniTrumpDesign, 2},
			[2]int{TarocchiniTrumpDesign, 3},
			[2]int{TarocchiniTrumpDesign, 4},
			[2]int{TarocchiniTrumpDesign, 5},
		), 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tarocchiniTrickWinnerOf(tc.trick, led))
		})
	}
}

// 同格なのはパパ同士だけ。通常の切り札は数字順で、後出しでは勝てない。
func TestTarocchini_OrdinaryTrumpsKeepStrictOrder(t *testing.T) {
	led := 1
	assert.Equal(t, 0, tarocchiniTrickWinnerOf(tarocchiniTrick(
		[2]int{TarocchiniTrumpDesign, 20},
		[2]int{TarocchiniTrumpDesign, 19},
		[2]int{led, 14},
		[2]int{led, 13},
	), led), "a later but lower trump must not win")
}

// パパは中位の同格札であって最強札ではない。上位切り札には負ける。
func TestTarocchini_PapaLosesToAHigherTrump(t *testing.T) {
	led := 1
	assert.Equal(t, 1, tarocchiniTrickWinnerOf(tarocchiniTrick(
		[2]int{TarocchiniTrumpDesign, tarocchiniPapiHigh},
		[2]int{TarocchiniTrumpDesign, TarocchiniMaxTrump},
		[2]int{led, 14},
		[2]int{led, 13},
	), led), "the highest trump must beat a papa played earlier")

	// 席 0 が最強切り札、席 1 が後からパパ。後出しでも上位切り札には届かない。
	assert.Equal(t, 0, tarocchiniTrickWinnerOf(tarocchiniTrick(
		[2]int{TarocchiniTrumpDesign, TarocchiniMaxTrump},
		[2]int{TarocchiniTrumpDesign, tarocchiniPapiLow},
		[2]int{led, 14},
		[2]int{led, 13},
	), led), "a papa played later must not beat a higher trump")
}

func TestTarocchini_TrickBasics(t *testing.T) {
	led := 2
	t.Run("highest of the led suit wins with no trump", func(t *testing.T) {
		assert.Equal(t, 1, tarocchiniTrickWinnerOf(tarocchiniTrick(
			[2]int{led, 9}, [2]int{led, 14}, [2]int{led, 1}, [2]int{3, 14},
		), led))
	})
	t.Run("any trump beats any plain card", func(t *testing.T) {
		assert.Equal(t, 2, tarocchiniTrickWinnerOf(tarocchiniTrick(
			[2]int{led, 14}, [2]int{led, 13}, [2]int{TarocchiniTrumpDesign, 1}, [2]int{led, 12},
		), led))
	})
	t.Run("an off-suit discard never wins", func(t *testing.T) {
		assert.Equal(t, 0, tarocchiniTrickWinnerOf(tarocchiniTrick(
			[2]int{led, 6}, [2]int{4, 14}, [2]int{3, 14}, [2]int{1, 14},
		), led))
	})
	t.Run("the Matto never takes the trick", func(t *testing.T) {
		assert.Equal(t, 1, tarocchiniTrickWinnerOf(tarocchiniTrick(
			[2]int{TarocchiniMattoDesign, TarocchiniMattoValue},
			[2]int{led, 6}, [2]int{4, 14}, [2]int{3, 14},
		), led))
	})
	t.Run("an empty trick does not panic", func(t *testing.T) {
		assert.Equal(t, 0, tarocchiniTrickWinnerOf(nil, led))
	})
}

// newTestTarocchini は配り札を固定した Tarocchini を返す。
//
// **Reset 直後の手札に依存する主張は必ずシードを固定する。**固定しないと低確率で
// 落ちるテストになり develop を赤くする (#4467)。ただしコンストラクタが乱数源を
// 入れているかは別に確かめる —— helper が SetRand で上書きすると、既定が
// 壊れていても全テストが通ってしまう (Ganjifa #4661)。
func newTestTarocchini(t *testing.T) *Tarocchini {
	t.Helper()
	g := NewDefaultTarocchini()
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()
	return g
}

// **SetRand を呼ばずに作る。**この一本だけは意図的にシードを固定しない。
func TestTarocchini_ProductionConstructorShufflesTheDeck(t *testing.T) {
	handOf := func(g *Tarocchini, seat int) string {
		p := g.GetPlayer(seat)
		var b strings.Builder
		for i := 0; i < p.GetCardsSize(); i++ {
			fmt.Fprintf(&b, "%d-%d,", p.GetCard(i).GetDesign(), p.GetCard(i).GetValue())
		}
		return b.String()
	}
	first, second := NewDefaultTarocchini(), NewDefaultTarocchini()
	first.Reset()
	second.Reset()
	assert.NotEqual(t, handOf(first, 0), handOf(second, 0),
		"two fresh games dealt identical hands -- the constructor did not seed rng")
}

func TestTarocchini_ResetDealsFifteenEachAndTwoToTheDealer(t *testing.T) {
	g := newTestTarocchini(t)
	assert.Equal(t, TarocchiniPhaseScarto, g.GetPhase())
	for i := 0; i < TarocchiniPlayerCnt; i++ {
		want := TarocchiniHandSize
		if i == g.GetDealerIdx() {
			want += TarocchiniSurplus
		}
		assert.Equal(t, want, g.GetPlayer(i).GetCardsSize(), "seat %d", i)
	}
}

func TestTarocchini_ScartoReturnsTheDealerToFifteen(t *testing.T) {
	g := newTestTarocchini(t)
	require.True(t, g.IsHumanScartoTurn(), "seat 0 deals first and is the human")
	dealer := g.GetPlayer(g.GetDealerIdx())

	idx := make([]int, 0, TarocchiniSurplus)
	for i := 0; i < dealer.GetCardsSize() && len(idx) < TarocchiniSurplus; i++ {
		if tarocchiniCanDiscard(dealer.GetCard(i)) {
			idx = append(idx, i)
		}
	}
	require.Len(t, idx, TarocchiniSurplus)
	require.NoError(t, g.PlayerScarto(idx))

	assert.Equal(t, TarocchiniHandSize, dealer.GetCardsSize())
	assert.Equal(t, TarocchiniSurplus, g.GetScartoSize())
	assert.Equal(t, TarocchiniPhasePlay, g.GetPhase())
}

// 捨札はそのまま得点になるので、切り札とマットは伏せられない。
func TestTarocchini_ScartoRefusesTrumpsAndTheMatto(t *testing.T) {
	g := newTestTarocchini(t)
	dealer := g.GetPlayer(g.GetDealerIdx())
	for dealer.GetCardsSize() > 0 {
		dealer.RemoveCard(0)
	}
	dealer.AddCard(NewCard(TarocchiniTrumpDesign, 7, false))
	dealer.AddCard(NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false))
	dealer.AddCard(NewCard(1, 6, false))

	assert.Error(t, g.PlayerScarto([]int{0, 1}), "a trump and the Matto must be refused")
	assert.Error(t, g.PlayerScarto([]int{0, 2}), "a trump must be refused")
	assert.Error(t, g.PlayerScarto([]int{2, 2}), "the same card twice must be refused")
	assert.Error(t, g.PlayerScarto([]int{2}), "the wrong count must be refused")
	assert.Error(t, g.PlayerScarto([]int{2, 99}), "an out-of-range index must be refused")
}

func TestTarocchini_ScartoGuards(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhasePlay)
	assert.ErrorIs(t, g.PlayerScarto([]int{0, 1}), ErrWrongPhase)

	g.SetPhase(TarocchiniPhaseScarto)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerScarto([]int{0, 1}), ErrGameEnded)
}

func TestTarocchini_CpuScartoDiscardsTwo(t *testing.T) {
	g := newTestTarocchini(t)
	g.dealerIdx = 1 // a CPU seat
	g.startRound()
	require.False(t, g.IsHumanScartoTurn())
	g.CpuScarto()
	assert.Equal(t, TarocchiniHandSize, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, TarocchiniPhasePlay, g.GetPhase())
}

// 全 15 トリックを回し、トリック数とチーム得点が保存されることを確かめる。
func TestTarocchini_FullRoundAccountsForEveryTrick(t *testing.T) {
	g := newTestTarocchini(t)
	g.CpuScarto()
	if g.GetPhase() == TarocchiniPhaseScarto {
		dealer := g.GetPlayer(g.GetDealerIdx())
		idx := make([]int, 0, TarocchiniSurplus)
		for i := 0; i < dealer.GetCardsSize() && len(idx) < TarocchiniSurplus; i++ {
			if tarocchiniCanDiscard(dealer.GetCard(i)) {
				idx = append(idx, i)
			}
		}
		require.NoError(t, g.PlayerScarto(idx))
	}

	for range TarocchiniTrickCount {
		for range TarocchiniPlayerCnt {
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
				continue
			}
			g.CpuPlay()
		}
		require.Equal(t, TarocchiniPhaseTrickEnd, g.GetPhase())
		g.ResolveTrick()
		if g.GetPhase() == TarocchiniPhaseTrickEnd {
			g.NextTrick()
		}
	}
	require.Equal(t, TarocchiniPhaseRoundEnd, g.GetPhase())

	tricks := 0
	for _, n := range g.GetRoundTricks() {
		tricks += n
	}
	assert.Equal(t, TarocchiniTrickCount, tricks, "every trick must land with someone")
	for i := 0; i < TarocchiniPlayerCnt; i++ {
		assert.Zero(t, g.GetPlayer(i).GetCardsSize(), "seat %d should be out of cards", i)
	}

	// チーム得点 = トリック数 + 最終トリックボーナス + スカルト 2 枚。
	total := g.GetTeamScores()[0] + g.GetTeamScores()[1]
	assert.Equal(t, TarocchiniTrickCount+TarocchiniLastTrickBonus+TarocchiniSurplus, total)
}

func TestTarocchini_MustFollowTheLedSuit(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhasePlay)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(1, 9, false))                     // follows
	p.AddCard(NewCard(2, 14, false))                    // off-suit
	p.AddCard(NewCard(TarocchiniTrumpDesign, 8, false)) // trump
	g.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(1, 6, false)}}

	assert.Equal(t, []int{0}, g.GetValidPlayIndices(0), "holding the led suit forces it")
	assert.Error(t, g.PlayerPlay(1), "an off-suit card must be refused")
	assert.NoError(t, g.PlayerPlay(0))
}

// リードスートを持たないときは切り札を出す義務がある (タロー系の共通ルール)。
func TestTarocchini_VoidInTheLedSuitForcesATrump(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhasePlay)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(2, 14, false))
	p.AddCard(NewCard(TarocchiniTrumpDesign, 8, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(1, 6, false)}}

	assert.Equal(t, []int{1}, g.GetValidPlayIndices(0), "void in the led suit forces the trump")
	assert.Error(t, g.PlayerPlay(0))
}

// マットはフォロー義務を免除される。
func TestTarocchini_MattoIsAlwaysPlayable(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhasePlay)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(1, 9, false))
	p.AddCard(NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(1, 6, false)}}

	assert.Equal(t, []int{0, 1}, g.GetValidPlayIndices(0), "the Matto is legal alongside a follow")
}

func TestTarocchini_PlayGuards(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(999))

	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerPlay(0), ErrNotHumanTurn)

	g.SetPhase(TarocchiniPhaseScarto)
	assert.ErrorIs(t, g.PlayerPlay(0), ErrWrongPhase)

	g.SetPhase(TarocchiniPhasePlay)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPlay(0), ErrGameEnded)
}

// 手札が空の CPU に出させると RemoveCard が nil を返す。それを渡すと落ちる (#4606)。
func TestTarocchini_CpuPlayWithAnEmptyHandIsANoOp(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhasePlay)
	g.SetCurrentPlayerIdx(1)
	p := g.GetPlayer(1)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	assert.NotPanics(t, func() { g.CpuPlay() })
	assert.Empty(t, g.GetCurrentTrick())
}

func TestTarocchini_MatchWinnerIsATeam(t *testing.T) {
	cases := map[string]struct {
		scores [2]int
		want   int
	}{
		"team 0 ahead":         {[2]int{40, 30}, 0},
		"team 1 ahead":         {[2]int{30, 40}, 1},
		"a draw has no winner": {[2]int{35, 35}, -1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			g := newTestTarocchini(t)
			g.teamScores = tc.scores
			g.finishMatch()
			assert.Equal(t, tc.want, g.GetWinnerTeam())
			assert.True(t, g.GetGameEndFlag())
			assert.Equal(t, TarocchiniPhaseGameEnd, g.GetPhase())
		})
	}
}

func TestTarocchini_ConfigValidation(t *testing.T) {
	assert.NoError(t, DefaultTarocchiniConfig().Validate())
	assert.Error(t, TarocchiniConfig{CpuDifficulty: 99, TargetRounds: 4}.Validate())
	assert.Error(t, TarocchiniConfig{TargetRounds: 0}.Validate())
	// ディーラーが一巡しないとスカルトの回数が不平等になる。
	assert.Error(t, TarocchiniConfig{TargetRounds: 6}.Validate(),
		"rounds must be a multiple of the player count")
	assert.NoError(t, TarocchiniConfig{TargetRounds: 8}.Validate())
}

// 全状態が非公開フィールドなので、専用コーデックが無いと Worker のセッションが
// 毎リクエスト空に戻る。Ganjifa (#4661) と Vira (#4660) で同じ穴を開けている。
func TestTarocchini_JSONRoundTripPreservesState(t *testing.T) {
	src := NewDefaultTarocchini()
	src.SetRand(rand.New(rand.NewSource(9)))
	src.Reset()
	src.CpuScarto()
	src.teamScores = [2]int{7, 5}

	data, err := src.MarshalJSON()
	require.NoError(t, err)

	var got Tarocchini
	require.NoError(t, got.UnmarshalJSON(data))
	assert.Equal(t, src.GetPhase(), got.GetPhase())
	assert.Equal(t, src.GetDealerIdx(), got.GetDealerIdx())
	assert.Equal(t, src.GetTeamScores(), got.GetTeamScores())
	assert.Equal(t, TarocchiniPlayerCnt, got.GetPlayerCnt())
	for i := 0; i < TarocchiniPlayerCnt; i++ {
		assert.Equal(t, src.GetPlayer(i).GetCardsSize(), got.GetPlayer(i).GetCardsSize())
	}
}

// 切り札 (5) とマット (6) を含む局面が復元できること。52 枚デッキ用の 1..4 を
// 持ち込むと、このゲームのセッションは軒並み復元不能になる。
func TestTarocchini_UnmarshalAcceptsTarotDesigns(t *testing.T) {
	src := NewDefaultTarocchini()
	src.SetRand(rand.New(rand.NewSource(9)))
	src.Reset()
	src.SetPhase(TarocchiniPhasePlay)
	src.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(TarocchiniTrumpDesign, 21, false)},
		{PlayerIdx: 2, Card: NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false)},
	}
	data, err := src.MarshalJSON()
	require.NoError(t, err)

	var got Tarocchini
	require.NoError(t, got.UnmarshalJSON(data))
	require.Len(t, got.GetCurrentTrick(), 2)
	assert.Equal(t, TarocchiniTrumpDesign, got.GetCurrentTrick()[0].Card.GetDesign())
	assert.Equal(t, TarocchiniMattoDesign, got.GetCurrentTrick()[1].Card.GetDesign())
}

func TestTarocchini_UnmarshalRejectsBadState(t *testing.T) {
	cases := map[string]string{
		"not json":                 `{`,
		"wrong player count":       `{"pl":[],"rn":1,"tn":1}`,
		"trick number too high":    `{"pl":[{},{},{},{}],"rn":1,"tn":99,"cfg":{"cd":1,"tr":4}}`,
		"scarto too large":         `{"pl":[{},{},{},{}],"rn":1,"tn":1,"sc":[{},{},{}],"cfg":{"cd":1,"tr":4}}`,
		"nil trick card":           `{"pl":[{},{},{},{}],"rn":1,"tn":1,"ct":[null],"cfg":{"cd":1,"tr":4}}`,
		"winner team out of range": `{"pl":[{},{},{},{}],"rn":1,"tn":1,"wt":5,"cfg":{"cd":1,"tr":4}}`,
		"rounds not a multiple":    `{"pl":[{},{},{},{}],"rn":1,"tn":1,"cfg":{"cd":1,"tr":6}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var got Tarocchini
			assert.Error(t, got.UnmarshalJSON([]byte(payload)))
		})
	}
}

func TestTarocchini_NextRoundRotatesTheDealer(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhaseRoundEnd)
	dealer, round := g.GetDealerIdx(), g.GetRoundNumber()

	g.NextRound()
	assert.Equal(t, round+1, g.GetRoundNumber())
	assert.Equal(t, (dealer+1)%TarocchiniPlayerCnt, g.GetDealerIdx())
	assert.Equal(t, TarocchiniPhaseScarto, g.GetPhase())
	assert.Equal(t, 1, g.GetTrickNumber())
	for i := 0; i < TarocchiniPlayerCnt; i++ {
		assert.Zero(t, g.GetRoundTricks()[i])
	}
}

func TestTarocchini_NextRoundEndsTheMatchOnTheLastRound(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhaseRoundEnd)
	g.roundNumber = g.GetConfig().TargetRounds
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
}

func TestTarocchini_NextRoundAndScoreRoundAreNoOpsOutsideRoundEnd(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhasePlay)
	g.roundNumber = g.GetConfig().TargetRounds
	g.NextRound()
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, g.GetConfig().TargetRounds, g.GetRoundNumber(), "neither call advanced the round")
}

func TestTarocchini_ScoreRoundEndsTheMatchOnTheLastRound(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhaseRoundEnd)

	g.roundNumber = g.GetConfig().TargetRounds - 1
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag())

	g.roundNumber = g.GetConfig().TargetRounds
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
}

func TestTarocchini_ResolveAndNextTrickAreNoOpsOutsideTrickEnd(t *testing.T) {
	g := newTestTarocchini(t)
	g.SetPhase(TarocchiniPhasePlay)
	g.ResolveTrick()
	g.NextTrick()
	assert.Equal(t, TarocchiniPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetTrickNumber())
}

func TestTarocchini_Accessors(t *testing.T) {
	g := newTestTarocchini(t)
	assert.Equal(t, TarocchiniPlayerCnt, g.GetPlayerCnt())
	assert.Len(t, g.GetPlayers(), TarocchiniPlayerCnt)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(TarocchiniPlayerCnt))
	assert.Equal(t, -1, g.GetLastTrickWinner())
	assert.Equal(t, (g.GetDealerIdx()+1)%TarocchiniPlayerCnt, g.GetLeadPlayerIdx())
	// 配っただけでは棋譜は空。スカルトを実行して初めて 1 件載る。
	assert.Empty(t, g.GetActionLog())
	g.CpuScarto()
	if g.GetPhase() == TarocchiniPhasePlay {
		assert.NotEmpty(t, g.GetActionLog())
	}

	cfg := TarocchiniConfig{CpuDifficulty: TarocchiniCpuDifficultyHard, TargetRounds: 8}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())

	assert.Nil(t, g.GetValidPlayIndices(-1))
	assert.Nil(t, g.GetValidPlayIndices(TarocchiniPlayerCnt))
	assert.NotEmpty(t, g.GetPlayableIndices(0))
}

func TestTarocchini_GetHintOnlyOnTheHumanPlayTurn(t *testing.T) {
	g := newTestTarocchini(t)
	assert.Nil(t, g.GetHint(), "no play hint during the scarto phase")

	g.SetPhase(TarocchiniPhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint(), "no hint on a CPU turn")

	g.SetCurrentPlayerIdx(0)
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Len(t, hint.CardIndices, 1)
	assert.NotEmpty(t, hint.Reason)

	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	assert.Nil(t, g.GetHint(), "an empty hand has nothing to suggest")
}

// パパとマットには専用の理由を返す。「後出しが勝つ」を知らないと同じ助言が
// 真逆の手に読めるため、キーで区別する。
func TestTarocchini_HintReasons(t *testing.T) {
	cases := []struct {
		name       string
		card       *Card
		trickCards int
		want       string
	}{
		{"leading a plain card", NewCard(1, 6, false), 0, "lead_low"},
		{"leading a trump", NewCard(TarocchiniTrumpDesign, 12, false), 0, "lead_trump"},
		{"a papa is called out by name", NewCard(TarocchiniTrumpDesign, tarocchiniPapiLow, false), 0, "play_papa"},
		{"the Matto is called out by name", NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false), 1, "play_matto"},
		{"following with a trump", NewCard(TarocchiniTrumpDesign, 12, false), 1, "follow_trump"},
		{"following with a plain card", NewCard(1, 6, false), 1, "follow_low"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestTarocchini(t)
			p := g.GetPlayer(0)
			for p.GetCardsSize() > 0 {
				p.RemoveCard(0)
			}
			p.AddCard(tc.card)
			g.currentTrick = nil
			for i := 0; i < tc.trickCards; i++ {
				g.currentTrick = append(g.currentTrick,
					&TrickCard{PlayerIdx: i + 1, Card: NewCard(2, 9, false)})
			}
			assert.Equal(t, tc.want, g.playHintReason(0, 0))
		})
	}
}

func TestTarocchini_HintReasonHandlesAMissingCard(t *testing.T) {
	g := newTestTarocchini(t)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	assert.Equal(t, "lead_low", g.playHintReason(0, 0), "an out-of-range index must not panic")
}
