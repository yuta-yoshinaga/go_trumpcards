//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestGermanWhist() *GermanWhist {
	g := NewDefaultGermanWhist()
	g.Reset()
	return g
}

// setGermanWhistHands installs exact hands so a test can state the position it
// cares about. Never assert on a freshly Reset deal -- it is shuffled.
func setGermanWhistHands(g *GermanWhist, hands [][]*Card) {
	for i, cards := range hands {
		p := g.players[i]
		p.Reset()
		for _, c := range cards {
			p.AddCard(c)
		}
	}
}

func TestNewGermanWhist(t *testing.T) {
	assert.NotNil(t, NewGermanWhist(NewTrumpCards(0), []*GermanWhistPlayer{
		NewGermanWhistPlayer(true), NewGermanWhistPlayer(false),
	}))
	assert.NotNil(t, NewDefaultGermanWhist())
}

func TestGermanWhist_Reset(t *testing.T) {
	g := newTestGermanWhist()

	assert.Equal(t, GermanWhistPhaseDraw, g.GetPhase())
	assert.Equal(t, 0, g.GetTrickNumber())
	assert.Equal(t, GermanWhistPlayerCnt, g.GetPlayerCnt())
	for i := range GermanWhistPlayerCnt {
		assert.Equal(t, GermanWhistHandSize, g.GetPlayer(i).GetCardsSize(), "player %d", i)
		assert.Equal(t, 0, g.GetPlayer(i).GetScoringTricks(), "player %d", i)
	}

	// 13 + 13 dealt, one turned up, so 25 remain face down.
	require.NotNil(t, g.GetUpCard())
	assert.Equal(t, CardCnt-GermanWhistHandSize*2-1, g.GetStockCount())
	assert.Equal(t, 25, g.GetStockCount())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
}

// **The rule #5232 omits.** The suit of the first exposed card is trump for the
// whole hand; without it this is a plain follow-suit game.
func TestGermanWhist_Reset_TrumpComesFromTheFirstUpCard(t *testing.T) {
	for range 20 {
		g := newTestGermanWhist()
		require.NotNil(t, g.GetUpCard())
		assert.Equal(t, g.GetUpCard().GetDesign(), g.GetTrumpSuit())
	}
}

func TestGermanWhist_ResetDealsEveryCard(t *testing.T) {
	for range 20 {
		g := newTestGermanWhist()
		total := g.GetStockCount()
		if g.GetUpCard() != nil {
			total++
		}
		for i := range GermanWhistPlayerCnt {
			total += g.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, CardCnt, total)
	}
}

func TestGermanWhist_ResetTwiceIsClean(t *testing.T) {
	g := newTestGermanWhist()
	require.NoError(t, g.PlayerPlay(0))
	g.Reset()
	assert.Equal(t, 0, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())
	assert.Equal(t, GermanWhistHandSize, g.GetPlayer(0).GetCardsSize())
}

// --- Trick resolution ---

func TestGermanWhist_Beats(t *testing.T) {
	const trump = CardDesignSpade
	lead := CardDesignHeart

	// Higher of the led suit wins.
	assert.True(t, germanWhistBeats(NewCard(lead, 10, true), NewCard(lead, 5, true), lead, trump))
	assert.False(t, germanWhistBeats(NewCard(lead, 5, true), NewCard(lead, 10, true), lead, trump))
	// An off-suit card that is not trump never wins.
	assert.False(t, germanWhistBeats(NewCard(CardDesignClover, 13, true), NewCard(lead, 2, true), lead, trump))
	// Trump beats any non-trump.
	assert.True(t, germanWhistBeats(NewCard(trump, 2, true), NewCard(lead, 13, true), lead, trump))
	assert.False(t, germanWhistBeats(NewCard(lead, 13, true), NewCard(trump, 2, true), lead, trump))
	// Between trumps, the higher wins.
	assert.True(t, germanWhistBeats(NewCard(trump, 5, true), NewCard(trump, 3, true), lead, trump))
}

// The Ace is the highest card, not the lowest.
func TestGermanWhist_Rank_AceIsHigh(t *testing.T) {
	assert.Equal(t, CardValueMax+1, germanWhistRank(NewCard(CardDesignSpade, 1, true)))
	assert.Greater(t, germanWhistRank(NewCard(CardDesignSpade, 1, true)),
		germanWhistRank(NewCard(CardDesignSpade, CardValueMax, true)))
	assert.True(t, germanWhistBeats(
		NewCard(CardDesignHeart, 1, true), NewCard(CardDesignHeart, CardValueMax, true),
		CardDesignHeart, CardDesignSpade))
}

// --- Follow suit ---

func TestGermanWhist_CanPlay_MustFollowSuit(t *testing.T) {
	g := newTestGermanWhist()
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignHeart, 5, true), NewCard(CardDesignClover, 9, true)},
		{NewCard(CardDesignHeart, 7, true), NewCard(CardDesignClover, 3, true)},
	})
	g.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 4, true)}}

	assert.True(t, g.canPlay(1, NewCard(CardDesignHeart, 7, true)))
	assert.False(t, g.canPlay(1, NewCard(CardDesignClover, 3, true)), "holds a heart, so must follow")

	// With no heart left, anything goes.
	g.players[1].Reset()
	g.players[1].AddCard(NewCard(CardDesignClover, 3, true))
	assert.True(t, g.canPlay(1, NewCard(CardDesignClover, 3, true)))
}

func TestGermanWhist_CanPlay_LeadIsFree(t *testing.T) {
	g := newTestGermanWhist()
	g.currentTrick = nil
	assert.True(t, g.canPlay(0, NewCard(CardDesignClover, 3, true)))
}

func TestGermanWhist_PlayerPlay_Rejections(t *testing.T) {
	g := newTestGermanWhist()

	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(GermanWhistHandSize))

	g.currentPlayerIdx = 1
	err := g.PlayerPlay(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not your turn")

	g.currentPlayerIdx = 0
	g.gameEndFlag = true
	err = g.PlayerPlay(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ended")
}

func TestGermanWhist_PlayerPlay_RejectsARevoke(t *testing.T) {
	g := newTestGermanWhist()
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignClover, 9, true), NewCard(CardDesignHeart, 5, true)},
		{NewCard(CardDesignHeart, 7, true)},
	})
	g.trumpSuit = CardDesignSpade
	g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 4, true)}}
	g.currentPlayerIdx = 0

	err := g.PlayerPlay(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "follow suit")
	assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize(), "the card stays in hand")
}

// --- The draw stage ---

// The winner takes the FACE-UP card and the loser the next face-down one.
func TestGermanWhist_DrawStage_WinnerTakesTheUpCard(t *testing.T) {
	g := newTestGermanWhist()
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignHeart, 10, true)},
		{NewCard(CardDesignHeart, 4, true)},
	})
	g.trumpSuit = CardDesignSpade
	g.upCard = NewCard(CardDesignDiamond, 1, true)
	g.stock = []*Card{NewCard(CardDesignClover, 2, true), NewCard(CardDesignClover, 3, true)}
	g.currentPlayerIdx = 0
	g.leadPlayerIdx = 0

	require.NoError(t, g.PlayerPlay(0))
	require.NoError(t, g.play(1, 0))

	// Player 0 won with the ten.
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 1, g.GetPlayer(0).GetCard(0).GetValue(), "the winner got the exposed Ace")
	assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, 2, g.GetPlayer(1).GetCard(0).GetValue(), "the loser got the hidden card")

	// A fresh card is exposed for the next trick.
	require.NotNil(t, g.GetUpCard())
	assert.Equal(t, 3, g.GetUpCard().GetValue())
	assert.Equal(t, 0, g.GetStockCount())
}

// No trick in the first stage counts towards the score.
func TestGermanWhist_DrawStage_TricksDoNotScore(t *testing.T) {
	g := newTestGermanWhist()
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignHeart, 10, true)},
		{NewCard(CardDesignHeart, 4, true)},
	})
	g.trumpSuit = CardDesignSpade
	g.upCard = NewCard(CardDesignDiamond, 1, true)
	g.stock = []*Card{NewCard(CardDesignClover, 2, true)}

	require.NoError(t, g.PlayerPlay(0))
	require.NoError(t, g.play(1, 0))

	assert.Equal(t, 1, g.GetPlayer(0).GetTrickCount(), "the trick is recorded")
	assert.Equal(t, 0, g.GetPlayer(0).GetScoringTricks(), "but it does not score")
	assert.Equal(t, 0, g.GetPlayer(1).GetScoringTricks())
}

// Trump is fixed by the FIRST exposed card and does not follow later ones.
func TestGermanWhist_TrumpDoesNotChangeWithTheNewUpCard(t *testing.T) {
	g := newTestGermanWhist()
	trumpBefore := g.GetTrumpSuit()
	first := g.GetUpCard()
	require.NotNil(t, first)

	// Play a whole trick so a new card is exposed.
	require.NoError(t, g.PlayerPlay(0))
	g.CpuPlay()
	require.NotNil(t, g.GetUpCard())

	assert.Equal(t, trumpBefore, g.GetTrumpSuit(), "trump is set once, at the deal")
}

// --- Phase transition and scoring ---

func TestGermanWhist_EntersScoringAfterThirteenTricks(t *testing.T) {
	g := newTestGermanWhist()
	g.trickNumber = GermanWhistStageTricks - 1
	g.phase = GermanWhistPhaseDraw
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignHeart, 10, true)},
		{NewCard(CardDesignHeart, 4, true)},
	})
	g.trumpSuit = CardDesignSpade
	g.upCard = NewCard(CardDesignDiamond, 1, true)
	g.stock = nil

	require.NoError(t, g.PlayerPlay(0))
	require.NoError(t, g.play(1, 0))

	assert.Equal(t, GermanWhistStageTricks, g.GetTrickNumber())
	assert.Equal(t, GermanWhistPhaseScoring, g.GetPhase())
	assert.Equal(t, 0, g.GetPlayer(0).GetScoringTricks(), "the 13th trick was still a draw trick")
}

func TestGermanWhist_ScoringStageCountsTricks(t *testing.T) {
	g := newTestGermanWhist()
	g.phase = GermanWhistPhaseScoring
	g.trickNumber = GermanWhistStageTricks
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignHeart, 10, true)},
		{NewCard(CardDesignHeart, 4, true)},
	})
	g.trumpSuit = CardDesignSpade
	g.currentPlayerIdx = 0

	require.NoError(t, g.PlayerPlay(0))
	require.NoError(t, g.play(1, 0))

	assert.Equal(t, 1, g.GetPlayer(0).GetScoringTricks())
	assert.Equal(t, 0, g.GetPlayer(1).GetScoringTricks())
	// No drawing happens any more.
	assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
}

func TestGermanWhist_FinishesAfterTwentySixTricks(t *testing.T) {
	g := newTestGermanWhist()
	g.phase = GermanWhistPhaseScoring
	g.trickNumber = GermanWhistStageTricks*2 - 1
	g.players[0].SetScoringTricks(6)
	g.players[1].SetScoringTricks(6)
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignHeart, 10, true)},
		{NewCard(CardDesignHeart, 4, true)},
	})
	g.trumpSuit = CardDesignSpade
	g.currentPlayerIdx = 0

	require.NoError(t, g.PlayerPlay(0))
	require.NoError(t, g.play(1, 0))

	assert.Equal(t, GermanWhistPhaseGameEnd, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 7, g.GetPlayer(0).GetScoringTricks())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, GermanWhistWinTricks, 7, "seven of thirteen is the majority")
}

func TestGermanWhist_WinnerIsWhoeverTookMoreScoringTricks(t *testing.T) {
	g := newTestGermanWhist()
	g.players[0].SetScoringTricks(5)
	g.players[1].SetScoringTricks(8)
	g.finish()
	assert.Equal(t, 1, g.GetWinnerIdx())
	assert.True(t, g.GetGameEndFlag())
}

// Thirteen is odd, so the second stage cannot actually tie -- but a restored
// snapshot could claim one, and that must not be reported as a win.
func TestGermanWhist_FinishLeavesNoWinnerOnATie(t *testing.T) {
	g := newTestGermanWhist()
	g.players[0].SetScoringTricks(6)
	g.players[1].SetScoringTricks(6)
	g.finish()
	assert.Equal(t, -1, g.GetWinnerIdx())
}

// --- CPU ---

// In the draw stage the CPU only fights for a card worth having: that choice is
// the whole game, so it must not simply always play its best card.
func TestGermanWhist_Cpu_FightsForAGoodUpCard(t *testing.T) {
	g := newTestGermanWhist()
	g.phase = GermanWhistPhaseDraw
	g.trumpSuit = CardDesignSpade
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignHeart, 2, true)},
		{NewCard(CardDesignHeart, 3, true), NewCard(CardDesignHeart, 13, true)},
	})
	g.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 2, true)}}
	g.currentPlayerIdx = 1

	g.upCard = NewCard(CardDesignSpade, 5, true) // a trump: worth taking
	assert.Equal(t, 1, g.chooseCpuCard(1), "plays the King to win a trump")

	g.upCard = NewCard(CardDesignDiamond, 3, true) // junk: not worth a good card
	assert.Equal(t, 0, g.chooseCpuCard(1), "ducks with the three")
}

// In the scoring stage every trick counts, so it always tries to win.
func TestGermanWhist_Cpu_AlwaysFightsInTheScoringStage(t *testing.T) {
	g := newTestGermanWhist()
	g.phase = GermanWhistPhaseScoring
	g.trumpSuit = CardDesignSpade
	setGermanWhistHands(g, [][]*Card{
		{NewCard(CardDesignHeart, 2, true)},
		{NewCard(CardDesignHeart, 3, true), NewCard(CardDesignHeart, 13, true)},
	})
	g.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 2, true)}}
	g.upCard = NewCard(CardDesignDiamond, 3, true)

	assert.Equal(t, 1, g.chooseCpuCard(1), "the up card is irrelevant now")
}

func TestGermanWhist_UpCardIsWorthTaking(t *testing.T) {
	g := newTestGermanWhist()
	g.trumpSuit = CardDesignSpade

	g.upCard = nil
	assert.False(t, g.upCardIsWorthTaking())
	g.upCard = NewCard(CardDesignSpade, 2, true)
	assert.True(t, g.upCardIsWorthTaking(), "any trump is worth it")
	g.upCard = NewCard(CardDesignHeart, 1, true)
	assert.True(t, g.upCardIsWorthTaking(), "an Ace is worth it")
	g.upCard = NewCard(CardDesignHeart, 10, true)
	assert.True(t, g.upCardIsWorthTaking())
	g.upCard = NewCard(CardDesignHeart, 4, true)
	assert.False(t, g.upCardIsWorthTaking())
}

func TestGermanWhist_CpuPlay_DoesNothingOutOfTurn(t *testing.T) {
	g := newTestGermanWhist()
	g.currentPlayerIdx = 0
	before := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before, g.GetPlayer(1).GetCardsSize())

	g.gameEndFlag = true
	g.currentPlayerIdx = 1
	g.CpuPlay()
	assert.Equal(t, before, g.GetPlayer(1).GetCardsSize())
}

// A full playthrough must reach the end without panicking, and the two stages
// must each account for exactly thirteen tricks.
func TestGermanWhist_FullGame(t *testing.T) {
	for range 10 {
		g := newTestGermanWhist()
		for range GermanWhistStageTricks * 2 * GermanWhistPlayerCnt {
			if g.GetGameEndFlag() {
				break
			}
			if g.GetCurrentPlayerIdx() == 0 {
				require.NoError(t, g.PlayerPlay(g.chooseCpuCard(0)))
			} else {
				g.CpuPlay()
			}
		}
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, GermanWhistStageTricks*2, g.GetTrickNumber())
		total := g.GetPlayer(0).GetScoringTricks() + g.GetPlayer(1).GetScoringTricks()
		assert.Equal(t, GermanWhistStageTricks, total, "only the second stage scores")
		assert.NotEqual(t, -1, g.GetWinnerIdx())
	}
}

func TestGermanWhist_GiveUp(t *testing.T) {
	g := newTestGermanWhist()
	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, GermanWhistPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 1, g.GetWinnerIdx())
	require.NotEmpty(t, g.GetActionLog())

	before := len(g.GetActionLog())
	g.GiveUp()
	assert.Len(t, g.GetActionLog(), before, "a second give-up adds nothing")
}

func TestGermanWhist_GetPlayer_OutOfRange(t *testing.T) {
	g := newTestGermanWhist()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(GermanWhistPlayerCnt))
}

func TestGermanWhist_ActionLog(t *testing.T) {
	g := newTestGermanWhist()
	require.NoError(t, g.PlayerPlay(0))
	log := g.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "play", log[len(log)-1].ActionType)
}

// --- JSON round-trip ---

func TestGermanWhist_JSONRoundTrip(t *testing.T) {
	g := newTestGermanWhist()
	require.NoError(t, g.PlayerPlay(0))
	g.CpuPlay()

	data, err := json.Marshal(g)
	require.NoError(t, err)

	restored := NewDefaultGermanWhist()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetTrickNumber(), restored.GetTrickNumber())
	assert.Equal(t, g.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, g.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, g.GetPlayer(0).GetCardsSize(), restored.GetPlayer(0).GetCardsSize())
}

// The per-player scoring count has to survive: the Worker rebuilds the game
// from KV on every request, so without it the score resets mid-hand (#4478).
func TestGermanWhist_JSONRoundTripKeepsScoringTricks(t *testing.T) {
	g := newTestGermanWhist()
	g.phase = GermanWhistPhaseScoring
	g.players[0].SetScoringTricks(4)
	g.players[1].SetScoringTricks(3)

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := NewDefaultGermanWhist()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, 4, restored.GetPlayer(0).GetScoringTricks())
	assert.Equal(t, 3, restored.GetPlayer(1).GetScoringTricks())
}

func TestGermanWhist_UnmarshalJSON_Rejections(t *testing.T) {
	for _, tc := range []struct{ name, data string }{
		{"broken json", `{`},
		{"phase too low", `{"ph":-1}`},
		{"phase too high", `{"ph":99}`},
		{"negative trick number", `{"tn":-1}`},
		{"trick number past the hand", `{"tn":27}`},
		{"current player out of range", `{"cp":5}`},
		{"lead player out of range", `{"lp":-2}`},
		{"winner out of range", `{"wi":7}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(tc.data), NewDefaultGermanWhist()))
		})
	}
}

func TestGermanWhist_UnmarshalJSON_RejectsOversizedArrays(t *testing.T) {
	big := make([]*Card, CardCnt+1)
	for i := range big {
		big[i] = NewCard(CardDesignSpade, 1, true)
	}

	t.Run("stock", func(t *testing.T) {
		data, err := json.Marshal(&germanWhistJSON{Stock: big})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultGermanWhist()))
	})
	t.Run("action log", func(t *testing.T) {
		data, err := json.Marshal(&germanWhistJSON{ActionLog: make([]*ActionLogEntry, germanWhistMaxSliceLen+1)})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultGermanWhist()))
	})
	t.Run("current trick too long", func(t *testing.T) {
		trick := make([]*TrickCard, GermanWhistPlayerCnt+1)
		for i := range trick {
			trick[i] = &TrickCard{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, true)}
		}
		data, err := json.Marshal(&germanWhistJSON{CurrentTrick: trick})
		require.NoError(t, err)
		assert.Error(t, json.Unmarshal(data, NewDefaultGermanWhist()))
	})

	// Negative control: a snapshot inside the bounds is accepted.
	t.Run("valid snapshot", func(t *testing.T) {
		data, err := json.Marshal(&germanWhistJSON{
			Phase: GermanWhistPhaseScoring, TrickNumber: 13, WinnerIdx: -1,
		})
		require.NoError(t, err)
		restored := NewDefaultGermanWhist()
		require.NoError(t, json.Unmarshal(data, restored))
		assert.Equal(t, GermanWhistPhaseScoring, restored.GetPhase())
	})
}

// --- GetValidPlayIndices / IsHumanTurn / GetHint ---

func TestGermanWhist_GetValidPlayIndices_LeadAllowsEverything(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	g.currentPlayerIdx = 0
	g.currentTrick = nil

	got := g.GetValidPlayIndices(0)
	assert.Len(t, got, g.GetPlayer(0).GetCardsSize(),
		"リードなら手札のどれでも出せる")
}

func TestGermanWhist_GetValidPlayIndices_MustFollowSuit(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	p := g.GetPlayer(1)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	// リードのスート 2 枚と、別スート 1 枚。
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	p.AddCard(NewCard(CardDesignSpade, 9, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 12, false)}}
	g.currentPlayerIdx = 1

	// スペードを持っているのでスペードの 2 枚だけが合法。
	assert.Equal(t, []int{0, 2}, g.GetValidPlayIndices(1))
}

func TestGermanWhist_GetValidPlayIndices_VoidAllowsEverything(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	p := g.GetPlayer(1)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	p.AddCard(NewCard(CardDesignClover, 9, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 12, false)}}
	g.currentPlayerIdx = 1

	assert.Equal(t, []int{0, 1}, g.GetValidPlayIndices(1),
		"リードのスートが無ければ何を出してもよい")
}

func TestGermanWhist_GetValidPlayIndices_OutOfRange(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	assert.Nil(t, g.GetValidPlayIndices(-1))
	assert.Nil(t, g.GetValidPlayIndices(GermanWhistPlayerCnt))
}

func TestGermanWhist_IsHumanTurn(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	g.currentPlayerIdx = 0
	assert.True(t, g.IsHumanTurn())
	g.currentPlayerIdx = 1
	assert.False(t, g.IsHumanTurn())
	g.currentPlayerIdx = 0
	g.gameEndFlag = true
	assert.False(t, g.IsHumanTurn(), "終局後は手番ではない")
}

func TestGermanWhist_GetHint_NilWhenNotHumanTurn(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	g.currentPlayerIdx = 1
	assert.Nil(t, g.GetHint())

	g.currentPlayerIdx = 0
	g.gameEndFlag = true
	assert.Nil(t, g.GetHint(), "終局後はヒント無し")
}

func TestGermanWhist_GetHint_SuggestsALegalCard(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	g.currentPlayerIdx = 0

	h := g.GetHint()
	if assert.NotNil(t, h) && assert.NotNil(t, h.CardIndex) {
		assert.Contains(t, g.GetValidPlayIndices(0), *h.CardIndex,
			"ヒントは必ず合法手を指す")
		assert.NotEmpty(t, h.Reason)
	}
}

// 前半の狙いは表向きの札。**それが欲しいかどうかで理由が入れ替わる**のが
// このゲームの肝なので、両方向を踏む。
func TestGermanWhist_GetHint_ReasonFollowsTheUpCard(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	g.phase = GermanWhistPhaseDraw
	g.currentPlayerIdx = 0
	g.currentTrick = nil
	g.trumpSuit = CardDesignSpade

	g.upCard = NewCard(CardDesignSpade, 1, false) // 切り札のA = 絶対欲しい
	win := g.GetHint()
	if assert.NotNil(t, win) {
		assert.Equal(t, "germanWhistTakeUpCard", win.Reason)
	}

	g.upCard = NewCard(CardDesignHeart, 2, false) // 2 = 要らない
	duck := g.GetHint()
	if assert.NotNil(t, duck) {
		assert.Equal(t, "germanWhistDuck", duck.Reason)
	}
}

func TestGermanWhist_GetHint_ScoringPhaseReason(t *testing.T) {
	g := NewDefaultGermanWhist()
	g.Reset()
	g.phase = GermanWhistPhaseScoring
	g.currentPlayerIdx = 0
	g.currentTrick = nil
	g.upCard = nil

	h := g.GetHint()
	if assert.NotNil(t, h) {
		assert.Equal(t, "germanWhistWinTrick", h.Reason,
			"後半は取れるだけ取る。表向きの札はもう無い")
	}
}
