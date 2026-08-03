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

// newTestMinchiate は配り札を固定した Minchiate を返す。
//
// **Reset 直後の手札に依存する主張は必ずシードを固定する** (#4467)。ただし
// コンストラクタが乱数源を入れているかは別に確かめる —— helper が SetRand で
// 上書きすると、既定が壊れていても全テストが通る (Ganjifa #4661)。
func newTestMinchiate(t *testing.T) *Minchiate {
	t.Helper()
	g := NewDefaultMinchiate()
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()
	return g
}

func minchiateTrick(entries ...[2]int) []*TrickCard {
	trick := make([]*TrickCard, 0, len(entries))
	for seat, e := range entries {
		trick = append(trick, &TrickCard{PlayerIdx: seat, Card: NewCard(e[0], e[1], false)})
	}
	return trick
}

func TestMinchiate_DeckIsNinetySevenDistinctCards(t *testing.T) {
	deck := buildMinchiateDeck()
	assert.Len(t, deck, MinchiateDeckSize)
	assert.Equal(t, 97, MinchiateDeckSize)

	seen := map[[2]int]bool{}
	suits, trumps, matto := 0, 0, 0
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "duplicate card %v", key)
		seen[key] = true
		switch {
		case minchiateIsMatto(c):
			matto++
		case minchiateIsTrump(c):
			trumps++
		default:
			suits++
		}
	}
	assert.Equal(t, 56, suits)
	// **切札は 21 枚ではなく 40 枚。**78 枚タローの感覚で作ると 19 枚足りない。
	assert.Equal(t, 40, trumps)
	assert.Equal(t, 1, matto)
}

// 97 は 4 で割り切れない。余りが構造として残ることを固定しておく。
func TestMinchiate_DealLeavesThirteenForTheDealer(t *testing.T) {
	assert.Equal(t, 13, MinchiateSurplus)
	assert.Equal(t, MinchiateDeckSize, MinchiatePlayerCnt*MinchiateHandSize+MinchiateSurplus)
}

// 切札 40 枚それぞれに固有の呼び名がある。番号だけだと「35 と 36 のどちらが強いか」
// 以外の情報が画面から失われる。
func TestMinchiate_EveryTrumpHasADistinctName(t *testing.T) {
	seen := map[string]bool{}
	for v := 1; v <= MinchiateMaxTrump; v++ {
		name := MinchiateTrumpName(v)
		assert.NotEmpty(t, name, "trump %d", v)
		assert.False(t, seen[name], "trump name %q is reused", name)
		seen[name] = true
	}
	assert.Len(t, seen, MinchiateMaxTrump)
	for _, bad := range []int{0, -1, MinchiateMaxTrump + 1} {
		assert.Empty(t, MinchiateTrumpName(bad))
	}
}

func TestMinchiate_TeamsAreOpposite(t *testing.T) {
	assert.Equal(t, MinchiateTeamOf(0), MinchiateTeamOf(2))
	assert.Equal(t, MinchiateTeamOf(1), MinchiateTeamOf(3))
	assert.NotEqual(t, MinchiateTeamOf(0), MinchiateTeamOf(1))
}

func TestMinchiate_TrickWinner(t *testing.T) {
	led := 1
	cases := []struct {
		name  string
		trick []*TrickCard
		want  int
	}{
		{"highest of the led suit wins", minchiateTrick(
			[2]int{led, 5}, [2]int{led, 14}, [2]int{led, 9}, [2]int{2, 14},
		), 1},
		{"any trump beats any plain card", minchiateTrick(
			[2]int{led, 14}, [2]int{MinchiateTrumpDesign, 1}, [2]int{led, 13}, [2]int{led, 12},
		), 1},
		// 切札 40 枚はすべて序列が異なる。Tarocchini のような同格札は無い。
		{"the higher trump wins, and later play does not help", minchiateTrick(
			[2]int{MinchiateTrumpDesign, 40}, [2]int{MinchiateTrumpDesign, 39},
			[2]int{MinchiateTrumpDesign, 2}, [2]int{led, 14},
		), 0},
		{"an off-suit discard never wins", minchiateTrick(
			[2]int{led, 6}, [2]int{2, 14}, [2]int{3, 14}, [2]int{4, 14},
		), 0},
		{"the Matto never takes the trick", minchiateTrick(
			[2]int{MinchiateMattoDesign, MinchiateMattoValue},
			[2]int{led, 6}, [2]int{2, 14}, [2]int{3, 14},
		), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, minchiateTrickWinnerOf(tc.trick, led))
		})
	}
	t.Run("an empty trick does not panic", func(t *testing.T) {
		assert.Equal(t, 0, minchiateTrickWinnerOf(nil, led))
	})
}

// **SetRand を呼ばずに作る。**この一本だけは意図的にシードを固定しない。
func TestMinchiate_ProductionConstructorShufflesTheDeck(t *testing.T) {
	handOf := func(g *Minchiate, seat int) string {
		p := g.GetPlayer(seat)
		var b strings.Builder
		for i := 0; i < p.GetCardsSize(); i++ {
			fmt.Fprintf(&b, "%d-%d,", p.GetCard(i).GetDesign(), p.GetCard(i).GetValue())
		}
		return b.String()
	}
	first, second := NewDefaultMinchiate(), NewDefaultMinchiate()
	first.Reset()
	second.Reset()
	assert.NotEqual(t, handOf(first, 0), handOf(second, 0),
		"two fresh games dealt identical hands -- the constructor did not seed rng")
}

func TestMinchiate_ResetDealsTwentyOneEachAndThirteenToTheDealer(t *testing.T) {
	g := newTestMinchiate(t)
	assert.Equal(t, MinchiatePhaseScarto, g.GetPhase())
	for i := 0; i < MinchiatePlayerCnt; i++ {
		want := MinchiateHandSize
		if i == g.GetDealerIdx() {
			want += MinchiateSurplus
		}
		assert.Equal(t, want, g.GetPlayer(i).GetCardsSize(), "seat %d", i)
	}
}

func TestMinchiate_ScartoReturnsTheDealerToTwentyOne(t *testing.T) {
	g := newTestMinchiate(t)
	require.True(t, g.IsHumanScartoTurn(), "seat 0 deals first and is the human")
	dealer := g.GetPlayer(g.GetDealerIdx())

	idx := make([]int, 0, MinchiateSurplus)
	for i := 0; i < dealer.GetCardsSize() && len(idx) < MinchiateSurplus; i++ {
		if minchiateCanDiscard(dealer.GetCard(i)) {
			idx = append(idx, i)
		}
	}
	require.Len(t, idx, MinchiateSurplus)
	require.NoError(t, g.PlayerScarto(idx))

	assert.Equal(t, MinchiateHandSize, dealer.GetCardsSize())
	assert.Equal(t, MinchiateSurplus, g.GetScartoSize())
	assert.Equal(t, MinchiatePhasePlay, g.GetPhase())
}

func TestMinchiate_ScartoRefusesTrumpsAndTheMatto(t *testing.T) {
	g := newTestMinchiate(t)
	dealer := g.GetPlayer(g.GetDealerIdx())
	for dealer.GetCardsSize() > 0 {
		dealer.RemoveCard(0)
	}
	dealer.AddCard(NewCard(MinchiateTrumpDesign, 7, false))
	dealer.AddCard(NewCard(MinchiateMattoDesign, MinchiateMattoValue, false))
	for i := 1; i <= MinchiateSurplus; i++ {
		dealer.AddCard(NewCard(1, i, false))
	}
	plain := make([]int, 0, MinchiateSurplus)
	for i := 2; i < 2+MinchiateSurplus; i++ {
		plain = append(plain, i)
	}

	withTrump := append([]int{0}, plain[:MinchiateSurplus-1]...)
	assert.Error(t, g.PlayerScarto(withTrump), "a trump must be refused")
	withMatto := append([]int{1}, plain[:MinchiateSurplus-1]...)
	assert.Error(t, g.PlayerScarto(withMatto), "the Matto must be refused")
	assert.Error(t, g.PlayerScarto(plain[:MinchiateSurplus-1]), "the wrong count must be refused")
	dup := append([]int{plain[0]}, plain[:MinchiateSurplus-1]...)
	assert.Error(t, g.PlayerScarto(dup), "the same card twice must be refused")
	assert.NoError(t, g.PlayerScarto(plain))
}

func TestMinchiate_ScartoGuards(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhasePlay)
	assert.ErrorIs(t, g.PlayerScarto(make([]int, MinchiateSurplus)), ErrWrongPhase)

	g.SetPhase(MinchiatePhaseScarto)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerScarto(make([]int, MinchiateSurplus)), ErrGameEnded)
}

func TestMinchiate_CpuScartoDiscardsTheSurplus(t *testing.T) {
	g := newTestMinchiate(t)
	g.dealerIdx = 1 // a CPU seat
	g.startRound()
	require.False(t, g.IsHumanScartoTurn())
	g.CpuScarto()
	assert.Equal(t, MinchiateHandSize, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, MinchiatePhasePlay, g.GetPhase())
}

// scartoDealt は親のスカルトを済ませた局面を返す。
func minchiateAfterScarto(t *testing.T, g *Minchiate) {
	t.Helper()
	g.CpuScarto()
	if g.GetPhase() != MinchiatePhaseScarto {
		return
	}
	dealer := g.GetPlayer(g.GetDealerIdx())
	idx := make([]int, 0, MinchiateSurplus)
	for i := 0; i < dealer.GetCardsSize() && len(idx) < MinchiateSurplus; i++ {
		if minchiateCanDiscard(dealer.GetCard(i)) {
			idx = append(idx, i)
		}
	}
	require.NoError(t, g.PlayerScarto(idx))
}

func TestMinchiate_FullRoundAccountsForEveryTrick(t *testing.T) {
	g := newTestMinchiate(t)
	minchiateAfterScarto(t, g)

	for range MinchiateTrickCount {
		for range MinchiatePlayerCnt {
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
				continue
			}
			g.CpuPlay()
		}
		require.Equal(t, MinchiatePhaseTrickEnd, g.GetPhase())
		g.ResolveTrick()
		if g.GetPhase() == MinchiatePhaseTrickEnd {
			g.NextTrick()
		}
	}
	require.Equal(t, MinchiatePhaseRoundEnd, g.GetPhase())

	tricks := 0
	for _, n := range g.GetRoundTricks() {
		tricks += n
	}
	assert.Equal(t, MinchiateTrickCount, tricks, "every trick must land with someone")
	for i := 0; i < MinchiatePlayerCnt; i++ {
		assert.Zero(t, g.GetPlayer(i).GetCardsSize(), "seat %d should be out of cards", i)
	}
	total := g.GetTeamScores()[0] + g.GetTeamScores()[1]
	assert.Equal(t, MinchiateTrickCount+MinchiateLastTrickBonus+MinchiateSurplus, total)
}

func TestMinchiate_MustFollowThenMustTrump(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhasePlay)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(1, 9, false))
	p.AddCard(NewCard(2, 14, false))
	p.AddCard(NewCard(MinchiateTrumpDesign, 8, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(1, 6, false)}}

	assert.Equal(t, []int{0}, g.GetValidPlayIndices(0), "holding the led suit forces it")
	assert.Error(t, g.PlayerPlay(1))
	require.NoError(t, g.PlayerPlay(0))

	// リードスートを持たなければ切札を出す義務。
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(2, 14, false))
	p.AddCard(NewCard(MinchiateTrumpDesign, 8, false))
	g.SetCurrentPlayerIdx(0)
	g.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(1, 6, false)}}
	assert.Equal(t, []int{1}, g.GetValidPlayIndices(0), "void in the led suit forces the trump")
}

func TestMinchiate_MattoIsAlwaysPlayable(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhasePlay)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(1, 9, false))
	p.AddCard(NewCard(MinchiateMattoDesign, MinchiateMattoValue, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(1, 6, false)}}
	assert.Equal(t, []int{0, 1}, g.GetValidPlayIndices(0), "the Matto is legal alongside a follow")
}

func TestMinchiate_PlayGuards(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(999))

	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerPlay(0), ErrNotHumanTurn)

	g.SetPhase(MinchiatePhaseScarto)
	assert.ErrorIs(t, g.PlayerPlay(0), ErrWrongPhase)

	g.SetPhase(MinchiatePhasePlay)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.PlayerPlay(0), ErrGameEnded)
}

// 手札が空の CPU に出させると RemoveCard が nil を返す (#4606)。
func TestMinchiate_CpuPlayWithAnEmptyHandIsANoOp(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhasePlay)
	g.SetCurrentPlayerIdx(1)
	p := g.GetPlayer(1)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	assert.NotPanics(t, func() { g.CpuPlay() })
	assert.Empty(t, g.GetCurrentTrick())
}

func TestMinchiate_CpuPlayIgnoresAHumanSeatAndAFinishedGame(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.CpuPlay()
	assert.Empty(t, g.GetCurrentTrick())

	g.SetCurrentPlayerIdx(1)
	g.gameEndFlag = true
	g.CpuPlay()
	assert.Empty(t, g.GetCurrentTrick())
}

func TestMinchiate_ResolveAndNextTrickAreNoOpsOutsideTrickEnd(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhasePlay)
	g.ResolveTrick()
	g.NextTrick()
	assert.Equal(t, MinchiatePhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetTrickNumber())
}

func TestMinchiate_NextRoundRotatesTheDealer(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhaseRoundEnd)
	dealer, round := g.GetDealerIdx(), g.GetRoundNumber()

	g.NextRound()
	assert.Equal(t, round+1, g.GetRoundNumber())
	assert.Equal(t, (dealer+1)%MinchiatePlayerCnt, g.GetDealerIdx())
	assert.Equal(t, MinchiatePhaseScarto, g.GetPhase())
	for i := 0; i < MinchiatePlayerCnt; i++ {
		assert.Zero(t, g.GetRoundTricks()[i])
	}
}

func TestMinchiate_NextRoundAndScoreRoundEndTheMatch(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhaseRoundEnd)
	g.roundNumber = g.GetConfig().TargetRounds
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())

	h := newTestMinchiate(t)
	h.SetPhase(MinchiatePhaseRoundEnd)
	h.roundNumber = h.GetConfig().TargetRounds - 1
	h.ScoreRound()
	assert.False(t, h.GetGameEndFlag())
	h.roundNumber = h.GetConfig().TargetRounds
	h.ScoreRound()
	assert.True(t, h.GetGameEndFlag())
}

func TestMinchiate_NextRoundAndScoreRoundAreNoOpsOutsideRoundEnd(t *testing.T) {
	g := newTestMinchiate(t)
	g.SetPhase(MinchiatePhasePlay)
	g.roundNumber = g.GetConfig().TargetRounds
	g.NextRound()
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, g.GetConfig().TargetRounds, g.GetRoundNumber())
}

func TestMinchiate_MatchWinnerIsATeam(t *testing.T) {
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
			g := newTestMinchiate(t)
			g.teamScores = tc.scores
			g.finishMatch()
			assert.Equal(t, tc.want, g.GetWinnerTeam())
			assert.True(t, g.GetGameEndFlag())
		})
	}
}

func TestMinchiate_HintOnlyOnTheHumanPlayTurn(t *testing.T) {
	g := newTestMinchiate(t)
	assert.Nil(t, g.GetHint(), "no play hint during the scarto phase")

	g.SetPhase(MinchiatePhasePlay)
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
	assert.Nil(t, g.GetHint())
}

func TestMinchiate_HintReasons(t *testing.T) {
	cases := []struct {
		name       string
		card       *Card
		trickCards int
		want       string
	}{
		{"leading a plain card", NewCard(1, 6, false), 0, "lead_low"},
		{"leading a trump", NewCard(MinchiateTrumpDesign, 33, false), 0, "lead_trump"},
		{"the Matto is called out by name", NewCard(MinchiateMattoDesign, MinchiateMattoValue, false), 1, "play_matto"},
		{"following with a trump", NewCard(MinchiateTrumpDesign, 12, false), 1, "follow_trump"},
		{"following with a plain card", NewCard(1, 6, false), 1, "follow_low"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestMinchiate(t)
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
	t.Run("an out-of-range index does not panic", func(t *testing.T) {
		g := newTestMinchiate(t)
		p := g.GetPlayer(0)
		for p.GetCardsSize() > 0 {
			p.RemoveCard(0)
		}
		assert.Equal(t, "lead_low", g.playHintReason(0, 0))
	})
}

func TestMinchiate_Accessors(t *testing.T) {
	g := newTestMinchiate(t)
	assert.Equal(t, MinchiatePlayerCnt, g.GetPlayerCnt())
	assert.Len(t, g.GetPlayers(), MinchiatePlayerCnt)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(MinchiatePlayerCnt))
	assert.Equal(t, -1, g.GetLastTrickWinner())
	assert.Equal(t, (g.GetDealerIdx()+1)%MinchiatePlayerCnt, g.GetLeadPlayerIdx())
	assert.Empty(t, g.GetActionLog())

	cfg := MinchiateConfig{CpuDifficulty: MinchiateCpuDifficultyHard, TargetRounds: 8}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())

	assert.Nil(t, g.GetValidPlayIndices(-1))
	assert.Nil(t, g.GetValidPlayIndices(MinchiatePlayerCnt))
	assert.NotEmpty(t, g.GetPlayableIndices(0))
}

func TestMinchiate_ConfigValidation(t *testing.T) {
	assert.NoError(t, DefaultMinchiateConfig().Validate())
	assert.Error(t, MinchiateConfig{CpuDifficulty: 99, TargetRounds: 4}.Validate())
	assert.Error(t, MinchiateConfig{TargetRounds: 0}.Validate())
	assert.Error(t, MinchiateConfig{TargetRounds: 6}.Validate(),
		"rounds must be a multiple of the player count")
	assert.NoError(t, MinchiateConfig{TargetRounds: 8}.Validate())
}

// Worker のセッションは毎リクエスト JSON を往復する。1 手ごとに往復しても状態が
// 壊れないことを確かめる —— 1 フィールドの取りこぼしでも局面が狂う。
func TestMinchiate_SurvivesRoundTrippingEveryRequest(t *testing.T) {
	g := NewDefaultMinchiate()
	g.SetRand(rand.New(rand.NewSource(5)))
	g.Reset()

	roundTrip := func(src *Minchiate) *Minchiate {
		t.Helper()
		data, err := src.MarshalJSON()
		require.NoError(t, err)
		var out Minchiate
		require.NoError(t, out.UnmarshalJSON(data))
		out.SetRand(rand.New(rand.NewSource(5)))
		return &out
	}

	minchiateAfterScarto(t, g)
	g = roundTrip(g)

	for trick := 0; trick < MinchiateTrickCount; trick++ {
		for range MinchiatePlayerCnt {
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid, "trick %d", trick)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
			g = roundTrip(g)
		}
		require.Equal(t, MinchiatePhaseTrickEnd, g.GetPhase(), "trick %d", trick)
		g.ResolveTrick()
		g = roundTrip(g)
		if g.GetPhase() == MinchiatePhaseTrickEnd {
			g.NextTrick()
			g = roundTrip(g)
		}
	}
	require.Equal(t, MinchiatePhaseRoundEnd, g.GetPhase())
	total := 0
	for _, n := range g.GetRoundTricks() {
		total += n
	}
	require.Equal(t, MinchiateTrickCount, total, "tricks were lost across the round trips")
}

func TestMinchiate_UnmarshalRejectsBadState(t *testing.T) {
	cases := map[string]string{
		"not json":              `{`,
		"wrong player count":    `{"pl":[],"rn":1,"tn":1}`,
		"trick number too high": `{"pl":[{},{},{},{}],"rn":1,"tn":99,"cfg":{"cd":1,"tr":4}}`,
		"nil trick card":        `{"pl":[{},{},{},{}],"rn":1,"tn":1,"ct":[null],"cfg":{"cd":1,"tr":4}}`,
		"winner team out of range": `{"pl":[{},{},{},{}],"rn":1,"tn":1,"wt":5,` +
			`"cfg":{"cd":1,"tr":4}}`,
		"rounds not a multiple": `{"pl":[{},{},{},{}],"rn":1,"tn":1,"cfg":{"cd":1,"tr":6}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var got Minchiate
			assert.Error(t, got.UnmarshalJSON([]byte(payload)))
		})
	}
}

// 切札 (5) とマット (6) を含む局面が復元できること。
func TestMinchiate_UnmarshalAcceptsTarotDesigns(t *testing.T) {
	src := NewDefaultMinchiate()
	src.SetRand(rand.New(rand.NewSource(5)))
	src.Reset()
	src.SetPhase(MinchiatePhasePlay)
	src.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(MinchiateTrumpDesign, MinchiateMaxTrump, false)},
		{PlayerIdx: 2, Card: NewCard(MinchiateMattoDesign, MinchiateMattoValue, false)},
	}
	data, err := src.MarshalJSON()
	require.NoError(t, err)

	var got Minchiate
	require.NoError(t, got.UnmarshalJSON(data))
	require.Len(t, got.GetCurrentTrick(), 2)
	assert.Equal(t, MinchiateMaxTrump, got.GetCurrentTrick()[0].Card.GetValue())
}

// マットはどのトリックも取らない。**リードされた場合も同じ。**
// ledSuit() を経由する本番の経路で確かめる —— 直接 led を渡すと、リードスートの
// 決め方の誤りをテストが迂回してしまう (Tarocchini #4662 で実際に迂回した)。
func TestMinchiate_MattoLedNeverTakesTheTrick(t *testing.T) {
	g := NewDefaultMinchiate()
	g.Reset()
	g.SetPhase(MinchiatePhaseTrickEnd)
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(MinchiateMattoDesign, MinchiateMattoValue, false)},
		{PlayerIdx: 1, Card: NewCard(1, 6, false)},
		{PlayerIdx: 2, Card: NewCard(1, 9, false)},
		{PlayerIdx: 3, Card: NewCard(2, 14, false)},
	}
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLastTrickWinner(),
		"the Matto must not take the trick; the highest card of the suit actually led should")
}

// マットがリードされても切札を出す義務は生じない。
func TestMinchiate_MattoLedImposesNoObligation(t *testing.T) {
	g := NewDefaultMinchiate()
	g.Reset()
	g.SetPhase(MinchiatePhasePlay)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(1, 6, false))
	p.AddCard(NewCard(MinchiateTrumpDesign, 9, false))
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 3, Card: NewCard(MinchiateMattoDesign, MinchiateMattoValue, false)},
	}
	assert.Equal(t, []int{0, 1}, g.GetValidPlayIndices(0))
	require.NoError(t, g.PlayerPlay(0), "a plain card must be legal after a led Matto")
}

// 味方がマットをリードした局面は「味方が勝っている」ではない。
func TestMinchiate_ALoneLedMattoLeavesTheTrickOpen(t *testing.T) {
	g := NewDefaultMinchiate()
	g.Reset()
	g.SetPhase(MinchiatePhasePlay)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(1, 6, false))
	p.AddCard(NewCard(MinchiateTrumpDesign, 38, false))
	require.Equal(t, MinchiateTeamOf(0), MinchiateTeamOf(2))
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 2, Card: NewCard(MinchiateMattoDesign, MinchiateMattoValue, false)},
	}
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, []int{1}, hint.CardIndices,
		"the trick is still open, so the strong card should be suggested rather than a duck")
}

// **スート札が 13 枚に満たない配りは実在する。**親は 97 枚中 34 枚を持ち、
// 捨てられないのは切札 40 + マット 1 の 41 枚。34 枚中スート札が 13 枚に届かない
// 裾は約 0.3% で起こる —— 以前ここに「構造上ありえない」と書いて捨てずに進めて
// いたが、それだと親の 13 枚が一度も場に出ないまま消え、97 = 21x4 + 13 の勘定が
// 崩れる (#4665 のレビュー指摘)。足りない分は切札を弱い順に開放する。
func TestMinchiate_ScartoFallsBackToTrumpsWhenSuitCardsAreShort(t *testing.T) {
	g := newTestMinchiate(t)
	// **親は CPU 席にする。**席 0 は人間なので CpuScarto は正しく何もしない。
	g.dealerIdx = 1
	g.startRound()
	dealer := g.GetPlayer(g.GetDealerIdx())
	require.False(t, dealer.GetIsHuman())
	for dealer.GetCardsSize() > 0 {
		dealer.RemoveCard(0)
	}
	// スート札を 3 枚だけ持たせ、残りは切札で埋める (計 34 枚)。
	const suitCards = 3
	for i := 1; i <= suitCards; i++ {
		dealer.AddCard(NewCard(1, i, false))
	}
	for i := 1; i <= MinchiateHandSize+MinchiateSurplus-suitCards; i++ {
		dealer.AddCard(NewCard(MinchiateTrumpDesign, i, false))
	}
	require.Equal(t, MinchiateHandSize+MinchiateSurplus, dealer.GetCardsSize())

	allowed := minchiateDiscardable(dealer)
	require.GreaterOrEqual(t, len(allowed), MinchiateSurplus,
		"the allowed set must still reach the surplus by opening up trumps")

	g.CpuScarto()
	// **枚数の勘定が閉じること。**捨てずに進めると親だけ 34 枚のまま残る。
	assert.Equal(t, MinchiateHandSize, dealer.GetCardsSize())
	assert.Equal(t, MinchiateSurplus, g.GetScartoSize())
	assert.Equal(t, MinchiatePhasePlay, g.GetPhase())
}

// 開放は必要な分だけ。スート札が足りている通常の配りでは切札は候補に入らない。
func TestMinchiate_ScartoPrefersSuitCardsWhenThereAreEnough(t *testing.T) {
	g := newTestMinchiate(t)
	dealer := g.GetPlayer(g.GetDealerIdx())
	for dealer.GetCardsSize() > 0 {
		dealer.RemoveCard(0)
	}
	for i := 0; i < MinchiateSurplus; i++ {
		dealer.AddCard(NewCard(1, (i%MinchiateSuitMax)+1, false))
	}
	dealer.AddCard(NewCard(MinchiateTrumpDesign, 40, false))

	for _, idx := range minchiateDiscardable(dealer) {
		assert.True(t, minchiateCanDiscard(dealer.GetCard(idx)),
			"a trump must not be offered while suit cards still suffice")
	}
}

// 人間の親も同じ許可集合で検証される。スート札が足りていれば切札は拒否される。
func TestMinchiate_HumanScartoUsesTheSameAllowedSet(t *testing.T) {
	g := newTestMinchiate(t)
	dealer := g.GetPlayer(g.GetDealerIdx())
	require.True(t, dealer.GetIsHuman())
	for dealer.GetCardsSize() > 0 {
		dealer.RemoveCard(0)
	}
	dealer.AddCard(NewCard(MinchiateTrumpDesign, 40, false)) // 位置 0 = 切札
	for i := 1; i <= MinchiateSurplus; i++ {
		dealer.AddCard(NewCard(1, i, false))
	}
	withTrump := []int{0}
	for i := 1; i < MinchiateSurplus; i++ {
		withTrump = append(withTrump, i)
	}
	assert.Error(t, g.PlayerScarto(withTrump), "a trump must be refused while suit cards suffice")

	plain := make([]int, 0, MinchiateSurplus)
	for i := 1; i <= MinchiateSurplus; i++ {
		plain = append(plain, i)
	}
	assert.NoError(t, g.PlayerScarto(plain))
}
