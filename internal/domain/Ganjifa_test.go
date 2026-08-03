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

// newTestGanjifa は配り札を固定した Ganjifa を返す。
//
// **Reset 直後の手札に依存する主張は必ずシードを固定する。**固定しないと
// 低確率で落ちるテストになり、develop を赤くする (#4467)。
func newTestGanjifa(t *testing.T) *Ganjifa {
	t.Helper()
	g := NewDefaultGanjifa()
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()
	return g
}

func TestGanjifa_StrongSuitRanksAscending(t *testing.T) {
	// 強いスート群 (1..4) は数字が大きいほど強い。
	for suit := 1; suit <= GanjifaStrongSuitMax; suit++ {
		assert.Less(t, GanjifaCardStrength(suit, 3), GanjifaCardStrength(suit, 9), "スート %d", suit)
		assert.Equal(t, 12, GanjifaCardStrength(suit, 12), "12 が最強")
		assert.Equal(t, 1, GanjifaCardStrength(suit, 1), "1 が最弱")
	}
}

func TestGanjifa_WeakSuitRanksDescending(t *testing.T) {
	// **弱いスート群 (5..8) は逆。**1 が最強で 12 が最弱。
	for suit := GanjifaStrongSuitMax + 1; suit <= GanjifaSuitCnt; suit++ {
		assert.Greater(t, GanjifaCardStrength(suit, 3), GanjifaCardStrength(suit, 9), "スート %d", suit)
		assert.Equal(t, 12, GanjifaCardStrength(suit, 1), "1 が最強")
		assert.Equal(t, 1, GanjifaCardStrength(suit, 12), "12 が最弱")
	}
}

func TestGanjifa_SameValueDiffersByGroup(t *testing.T) {
	// 同じ「3」が群によって下から 3 番目にも、上から 3 番目にもなる。
	// 生の value で比べると必ずどちらかを取り違える。
	assert.Equal(t, 3, GanjifaCardStrength(1, 3), "強い群では 3 のまま")
	assert.Equal(t, 10, GanjifaCardStrength(5, 3), "弱い群では 10 相当")
}

func TestGanjifa_GroupBoundaryIsFour(t *testing.T) {
	assert.True(t, GanjifaIsStrongSuit(GanjifaStrongSuitMax), "境界は強い群に含む")
	assert.False(t, GanjifaIsStrongSuit(GanjifaStrongSuitMax+1), "その次から弱い群")
}

func TestGanjifa_DeckHasNinetySixDistinctCards(t *testing.T) {
	deck := buildGanjifaDeck()
	require.Len(t, deck, GanjifaDeckSize)
	seen := map[[2]int]bool{}
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		require.False(t, seen[key], "重複した札 %v", key)
		seen[key] = true
	}
	assert.Len(t, seen, GanjifaDeckSize)
}

func TestGanjifa_ResetDealsThirtyTwoEach(t *testing.T) {
	g := newTestGanjifa(t)
	assert.Equal(t, GanjifaPhasePlay, g.GetPhase())
	for i := 0; i < GanjifaPlayerCnt; i++ {
		assert.Equal(t, GanjifaHandSize, g.GetPlayer(i).GetCardsSize(), "席 %d", i)
	}
	// 96 枚は 3 人で割り切れるので余り札は出ない。
	assert.Equal(t, GanjifaDeckSize, GanjifaHandSize*GanjifaPlayerCnt)
}

func TestGanjifa_TrumpIsChosenFromTheDealersHand(t *testing.T) {
	g := newTestGanjifa(t)
	trump := g.GetTrumpSuit()
	assert.GreaterOrEqual(t, trump, 1)
	assert.LessOrEqual(t, trump, GanjifaSuitCnt)
}

func TestGanjifa_TrumpBeatsAnyPlainCard(t *testing.T) {
	g := newTestGanjifa(t)
	trump := g.GetTrumpSuit()
	plain := 1
	if plain == trump {
		plain = 2
	}
	// 切り札の最弱でも、非切り札の最強より強い。
	weakTrump := NewCard(trump, 1, false)
	strongPlain := NewCard(plain, 12, false)
	assert.Greater(t, g.ganjifaRank(weakTrump), g.ganjifaRank(strongPlain))
}

func TestGanjifa_ValidPlayIndicesEnforceFollow(t *testing.T) {
	g := newTestGanjifa(t)
	lead := g.GetLeadPlayerIdx()
	// リード前は全部出せる。
	assert.Len(t, g.GetValidPlayIndices(lead), GanjifaHandSize)
}

func TestGanjifa_ConfigRejectsOutOfRangeDifficulty(t *testing.T) {
	assert.Error(t, GanjifaConfig{CpuDifficulty: 99, TargetRounds: 3}.Validate())
	assert.NoError(t, GanjifaConfig{CpuDifficulty: GanjifaCpuDifficultyNormal, TargetRounds: 3}.Validate())
}

func TestGanjifa_ConfigRejectsZeroRounds(t *testing.T) {
	assert.Error(t, GanjifaConfig{CpuDifficulty: GanjifaCpuDifficultyNormal, TargetRounds: 0}.Validate())
}

func TestGanjifa_HintNamesTheWeakSuitWhenItPicksOne(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetCurrentPlayerIdx(0)
	hint := g.GetHint()
	require.NotNil(t, hint)
	require.Len(t, hint.CardIndices, 1)
	card := g.GetPlayer(0).GetCard(hint.CardIndices[0])
	require.NotNil(t, card)
	// **弱い群の札を勧めるときは理由でそう言う。**強い群の感覚で読むと逆手になる。
	if GanjifaIsStrongSuit(card.GetDesign()) {
		assert.NotContains(t, hint.Reason, "weak_suit")
	} else {
		assert.Contains(t, hint.Reason, "weak_suit")
	}
}

// The whole game state lives in unexported fields, so a session that round-trips
// through KV on the Worker only survives if the custom codec carries it. Plain
// json.Marshal on the struct would emit "{}" and silently reset every hand.
func TestGanjifa_JSONRoundTripPreservesHands(t *testing.T) {
	src := NewDefaultGanjifa()
	src.SetRand(rand.New(rand.NewSource(7)))
	src.Reset()
	src.SetPhase(GanjifaPhaseTrickEnd)

	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	var got Ganjifa
	assert.NoError(t, got.UnmarshalJSON(data))
	assert.Equal(t, GanjifaPhaseTrickEnd, got.GetPhase())
	assert.Equal(t, src.GetTrumpSuit(), got.GetTrumpSuit())
	assert.Equal(t, src.GetDealerIdx(), got.GetDealerIdx())
	assert.Equal(t, GanjifaPlayerCnt, got.GetPlayerCnt())
	for i := 0; i < GanjifaPlayerCnt; i++ {
		assert.Equal(t, GanjifaHandSize, got.GetPlayer(i).GetCardsSize())
		assert.Equal(t, src.GetPlayer(i).GetCard(0).GetDesign(), got.GetPlayer(i).GetCard(0).GetDesign())
		assert.Equal(t, src.GetPlayer(i).GetCard(0).GetValue(), got.GetPlayer(i).GetCard(0).GetValue())
	}
}

// A weak-group trump is a legal state (designs 5-8), so borrowing the standard
// 52-card bound of 1-4 here would make half of all sessions unrestorable.
func TestGanjifa_UnmarshalAcceptsWeakGroupTrump(t *testing.T) {
	src := NewDefaultGanjifa()
	src.Reset()
	for suit := 1; suit <= GanjifaSuitCnt; suit++ {
		src.trumpSuit = suit
		data, err := src.MarshalJSON()
		assert.NoError(t, err)
		var got Ganjifa
		assert.NoError(t, got.UnmarshalJSON(data), "trump suit %d must round-trip", suit)
		assert.Equal(t, suit, got.GetTrumpSuit())
	}
}

func TestGanjifa_UnmarshalRejectsBadState(t *testing.T) {
	cases := map[string]string{
		"not json":   `{`,
		"no players": `{"pl":[],"rn":1,"tn":1}`,
		"trump out of range": `{"pl":[{},{},{}],"rn":1,"tn":1,"ts":9,` +
			`"cfg":{"cd":1,"tr":3}}`,
		"trick number too high": `{"pl":[{},{},{}],"rn":1,"tn":99,"cfg":{"cd":1,"tr":3}}`,
		"nil trick card":        `{"pl":[{},{},{}],"rn":1,"tn":1,"ct":[null],"cfg":{"cd":1,"tr":3}}`,
		"invalid config":        `{"pl":[{},{},{}],"rn":1,"tn":1,"cfg":{"cd":1,"tr":0}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var got Ganjifa
			assert.Error(t, got.UnmarshalJSON([]byte(payload)))
		})
	}
}

func TestGanjifa_SuitNamesAndGlyphs(t *testing.T) {
	seenNames := map[string]bool{}
	seenGlyphs := map[string]bool{}
	for suit := 1; suit <= GanjifaSuitCnt; suit++ {
		name, glyph := GanjifaSuitName(suit), GanjifaSuitGlyph(suit)
		assert.NotEmpty(t, name)
		assert.NotEmpty(t, glyph)
		assert.False(t, seenNames[name], "suit name %q is reused", name)
		assert.False(t, seenGlyphs[glyph], "suit glyph %q is reused", glyph)
		seenNames[name], seenGlyphs[glyph] = true, true
	}
	for _, bad := range []int{0, -1, GanjifaSuitCnt + 1} {
		assert.Empty(t, GanjifaSuitName(bad))
		assert.Empty(t, GanjifaSuitGlyph(bad))
	}
}

// setHand replaces a player's hand with exactly the given cards.
func setGanjifaHand(t *testing.T, g *Ganjifa, idx int, cards ...*Card) {
	t.Helper()
	p := g.GetPlayer(idx)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playOneTrick drives a full three-card trick from the current lead.
func playOneTrick(t *testing.T, g *Ganjifa) {
	t.Helper()
	for range GanjifaPlayerCnt {
		if g.IsHumanTurn() {
			idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
			require.NotEmpty(t, idx)
			require.NoError(t, g.PlayerPlay(idx[0]))
			continue
		}
		g.CpuPlay()
	}
}

// The trick winner is the crux of the inversion: in a weak suit the 1 beats the
// 12, which is the opposite of what a raw value comparison would say.
func TestGanjifa_TrickWinnerUsesGroupAwareOrder(t *testing.T) {
	cases := []struct {
		name      string
		lead      int // design of the led suit
		values    [3]int
		wantOrder int // index of the seat that must win
	}{
		{"strong suit: the 12 wins", 1, [3]int{3, 12, 7}, 1},
		{"weak suit: the 1 wins, not the 12", 5, [3]int{12, 1, 7}, 1},
		{"weak suit: 2 beats 11", 6, [3]int{11, 2, 9}, 1},
		{"strong suit: 1 is the weakest", 2, [3]int{1, 2, 3}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestGanjifa(t)
			g.trumpSuit = 0 // no trump, so the led suit alone decides
			g.currentTrick = nil
			for seat, v := range tc.values {
				g.currentTrick = append(g.currentTrick,
					&TrickCard{PlayerIdx: seat, Card: NewCard(tc.lead, v, false)})
			}
			assert.Equal(t, tc.wantOrder, g.trickWinner())
		})
	}
}

func TestGanjifa_TrumpBeatsTheLedSuitInEitherGroup(t *testing.T) {
	for _, trump := range []int{2, 7} { // one strong, one weak
		g := newTestGanjifa(t)
		g.trumpSuit = trump
		g.currentTrick = []*TrickCard{
			// Seat 0 leads the strongest possible card of a non-trump suit...
			{PlayerIdx: 0, Card: NewCard(1, 12, false)},
			// ...and seat 1 ruffs with the weakest card of the trump suit.
			{PlayerIdx: 1, Card: NewCard(trump, weakestValue(trump), false)},
			{PlayerIdx: 2, Card: NewCard(1, 11, false)},
		}
		assert.Equal(t, 1, g.trickWinner(), "trump %d must beat any plain card", trump)
	}
}

// weakestValue returns the lowest-ranking value in the given suit's group.
func weakestValue(design int) int {
	if GanjifaIsStrongSuit(design) {
		return 1
	}
	return GanjifaRankCnt
}

// A card of neither trump nor the led suit cannot win, however high its value.
func TestGanjifa_OffSuitDiscardNeverWins(t *testing.T) {
	g := newTestGanjifa(t)
	g.trumpSuit = 4
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(1, 2, false)},
		{PlayerIdx: 1, Card: NewCard(3, 12, false)},
		{PlayerIdx: 2, Card: NewCard(1, 1, false)},
	}
	assert.Equal(t, 0, g.trickWinner(), "seat 1 discarded off-suit and must not win")
}

func TestGanjifa_PlayerPlayRejectsRenege(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(GanjifaPhasePlay)
	setGanjifaHand(t, g, 0, NewCard(1, 5, false), NewCard(3, 9, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 2, Card: NewCard(1, 8, false)}}

	assert.Error(t, g.PlayerPlay(1), "holding the led suit, an off-suit card must be refused")
	assert.NoError(t, g.PlayerPlay(0))
}

func TestGanjifa_PlayerPlayGuards(t *testing.T) {
	t.Run("index out of range", func(t *testing.T) {
		g := newTestGanjifa(t)
		g.SetCurrentPlayerIdx(0)
		assert.Error(t, g.PlayerPlay(-1))
		assert.Error(t, g.PlayerPlay(GanjifaHandSize))
	})
	t.Run("not the human turn", func(t *testing.T) {
		g := newTestGanjifa(t)
		g.SetCurrentPlayerIdx(1)
		assert.ErrorIs(t, g.PlayerPlay(0), ErrNotHumanTurn)
	})
	t.Run("wrong phase", func(t *testing.T) {
		g := newTestGanjifa(t)
		g.SetPhase(GanjifaPhaseRoundEnd)
		assert.ErrorIs(t, g.PlayerPlay(0), ErrWrongPhase)
	})
	t.Run("game already ended", func(t *testing.T) {
		g := newTestGanjifa(t)
		g.gameEndFlag = true
		assert.ErrorIs(t, g.PlayerPlay(0), ErrGameEnded)
	})
}

// A CPU with an empty hand must not be asked to play: RemoveCard returns nil and
// passing that to playCard would nil-deref the whole request (#4606).
func TestGanjifa_CpuPlayWithAnEmptyHandIsANoOp(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetPhase(GanjifaPhasePlay)
	g.SetCurrentPlayerIdx(1)
	setGanjifaHand(t, g, 1)
	assert.NotPanics(t, func() { g.CpuPlay() })
	assert.Empty(t, g.GetCurrentTrick())
}

func TestGanjifa_CpuPlayIgnoresAHumanSeatAndAFinishedGame(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetPhase(GanjifaPhasePlay)
	g.SetCurrentPlayerIdx(0) // the human seat
	g.CpuPlay()
	assert.Empty(t, g.GetCurrentTrick())

	g.SetCurrentPlayerIdx(1)
	g.gameEndFlag = true
	g.CpuPlay()
	assert.Empty(t, g.GetCurrentTrick())
}

func TestGanjifa_TrickFlowAwardsAndLeadsFromTheWinner(t *testing.T) {
	g := newTestGanjifa(t)
	before := g.GetTrickNumber()
	playOneTrick(t, g)

	require.Equal(t, GanjifaPhaseTrickEnd, g.GetPhase())
	g.ResolveTrick()

	won := 0
	for i := range GanjifaPlayerCnt {
		won += g.GetRoundTricks()[i]
	}
	assert.Equal(t, 1, won, "exactly one seat takes the trick")
	assert.Equal(t, g.GetLeadPlayerIdx(), g.GetCurrentPlayerIdx(), "the winner leads next")

	g.NextTrick()
	assert.Equal(t, GanjifaPhasePlay, g.GetPhase())
	assert.Equal(t, before+1, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())
}

func TestGanjifa_ResolveAndNextTrickAreNoOpsOutsideTrickEnd(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetPhase(GanjifaPhasePlay)
	g.ResolveTrick()
	g.NextTrick()
	assert.Equal(t, GanjifaPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetTrickNumber())
}

// One point per trick, and the 32 tricks of a round are all accounted for.
func TestGanjifa_RoundSettlementAddsEveryTrick(t *testing.T) {
	g := newTestGanjifa(t)
	for range GanjifaTrickCount {
		playOneTrick(t, g)
		g.ResolveTrick()
		if g.GetPhase() == GanjifaPhaseTrickEnd {
			g.NextTrick()
		}
	}
	require.Equal(t, GanjifaPhaseRoundEnd, g.GetPhase())

	total := 0
	for i := range GanjifaPlayerCnt {
		assert.Equal(t, g.GetRoundTricks()[i], g.GetPlayerScores()[i])
		total += g.GetPlayerScores()[i]
	}
	assert.Equal(t, GanjifaTrickCount, total, "every trick must land in someone's score")
}

func TestGanjifa_NextRoundRedealsAndRotatesTheDealer(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetPhase(GanjifaPhaseRoundEnd)
	dealer, round := g.GetDealerIdx(), g.GetRoundNumber()

	g.NextRound()
	assert.Equal(t, round+1, g.GetRoundNumber())
	assert.Equal(t, (dealer+1)%GanjifaPlayerCnt, g.GetDealerIdx(), "the deal must move on")
	assert.Equal(t, GanjifaPhasePlay, g.GetPhase())
	for i := range GanjifaPlayerCnt {
		assert.Equal(t, GanjifaHandSize, g.GetPlayer(i).GetCardsSize())
		assert.Zero(t, g.GetRoundTricks()[i], "the per-round tally resets")
	}
}

func TestGanjifa_NextRoundIsANoOpOutsideRoundEnd(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetPhase(GanjifaPhasePlay)
	g.NextRound()
	assert.Equal(t, 1, g.GetRoundNumber())
}

func TestGanjifa_ScoreRoundEndsTheMatchOnTheLastRound(t *testing.T) {
	g := newTestGanjifa(t)
	cfg := g.GetConfig()
	g.SetPhase(GanjifaPhaseRoundEnd)

	g.roundNumber = cfg.TargetRounds - 1
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag(), "an earlier round must not end the match")

	g.roundNumber = cfg.TargetRounds
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, GanjifaPhaseGameEnd, g.GetPhase())
}

func TestGanjifa_ScoreRoundIsANoOpOutsideRoundEnd(t *testing.T) {
	g := newTestGanjifa(t)
	g.SetPhase(GanjifaPhasePlay)
	g.roundNumber = g.GetConfig().TargetRounds
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag())
}

func TestGanjifa_MatchWinner(t *testing.T) {
	cases := map[string]struct {
		scores [GanjifaPlayerCnt]int
		want   int
	}{
		"clear leader":        {[GanjifaPlayerCnt]int{40, 30, 26}, 0},
		"leader in last seat": {[GanjifaPlayerCnt]int{26, 30, 40}, 2},
		// A tie must leave no winner: breaking it by seat order would hand the
		// match to the lowest seat every time.
		"two-way tie at the top": {[GanjifaPlayerCnt]int{40, 40, 16}, -1},
		"three-way tie":          {[GanjifaPlayerCnt]int{32, 32, 32}, -1},
		"tie below the leader":   {[GanjifaPlayerCnt]int{40, 28, 28}, 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			g := newTestGanjifa(t)
			g.playerScores = tc.scores
			g.finishMatch()
			assert.Equal(t, tc.want, g.GetWinnerPlayer())
			assert.True(t, g.GetGameEndFlag())
		})
	}
}

func TestGanjifa_ValidPlayIndicesOutOfRange(t *testing.T) {
	g := newTestGanjifa(t)
	assert.Nil(t, g.GetValidPlayIndices(-1))
	assert.Nil(t, g.GetValidPlayIndices(GanjifaPlayerCnt))
}

func TestGanjifa_ValidPlayIndicesFallBackToTheWholeHandWhenVoid(t *testing.T) {
	g := newTestGanjifa(t)
	setGanjifaHand(t, g, 0, NewCard(2, 4, false), NewCard(3, 9, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 2, Card: NewCard(1, 8, false)}}
	assert.Equal(t, []int{0, 1}, g.GetValidPlayIndices(0), "void in the led suit frees the whole hand")
}

func TestGanjifa_ActionLogRecordsPlay(t *testing.T) {
	g := newTestGanjifa(t)
	playOneTrick(t, g)
	g.ResolveTrick()
	types := map[string]int{}
	for _, e := range g.GetActionLog() {
		types[e.ActionType]++
	}
	assert.Equal(t, GanjifaPlayerCnt, types["play"])
	assert.Equal(t, 1, types["trickwin"])
}

func TestGanjifa_ConfigAccessors(t *testing.T) {
	g := newTestGanjifa(t)
	cfg := GanjifaConfig{CpuDifficulty: GanjifaCpuDifficultyHard, TargetRounds: 9}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
	assert.Len(t, g.GetPlayers(), GanjifaPlayerCnt)
}

// **SetRand を呼ばずに作る。**この一本だけは意図的にシードを固定しない。
// 他のテストが全部 newTestGanjifa 経由で SetRand しているせいで、コンストラクタ
// が乱数源を入れ忘れていても誰も気づかなかった —— shuffle() の nil 分岐は
// 各札を自分自身と交換するだけの no-op で、本番の全対局が同じ配りになっていた。
func TestGanjifa_ProductionConstructorShufflesTheDeck(t *testing.T) {
	handOf := func(g *Ganjifa, seat int) string {
		p := g.GetPlayer(seat)
		var b strings.Builder
		for i := 0; i < p.GetCardsSize(); i++ {
			c := p.GetCard(i)
			fmt.Fprintf(&b, "%d-%d,", c.GetDesign(), c.GetValue())
		}
		return b.String()
	}

	// 96 枚から 32 枚を配る組み合わせは天文学的なので、独立した 2 局が一致
	// するのは実質ゼロ。一致したらシャッフルが効いていない。
	first := NewDefaultGanjifa()
	first.Reset()
	second := NewDefaultGanjifa()
	second.Reset()

	assert.NotEqual(t, handOf(first, 0), handOf(second, 0),
		"two fresh games dealt identical hands -- the constructor did not seed rng")
}

// Reset を繰り返しても同じ配りに戻らないこと。
func TestGanjifa_ResetReshuffles(t *testing.T) {
	g := NewDefaultGanjifa()
	seen := map[string]bool{}
	for range 5 {
		g.Reset()
		var b strings.Builder
		p := g.GetPlayer(0)
		for i := 0; i < p.GetCardsSize(); i++ {
			fmt.Fprintf(&b, "%d-%d,", p.GetCard(i).GetDesign(), p.GetCard(i).GetValue())
		}
		seen[b.String()] = true
	}
	assert.Greater(t, len(seen), 1, "every Reset produced the same deal")
}
