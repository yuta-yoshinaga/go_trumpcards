//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestTysiac returns a fresh, reset Tysiąc game with the default 3-player setup.
func newTestTysiac() *domain.Tysiac {
	g := domain.NewDefaultTysiac()
	g.Reset()
	return g
}

// setTysiacHand replaces player i's hand with the supplied cards deterministically.
func setTysiacHand(g *domain.Tysiac, i int, cards ...*domain.Card) {
	p := g.GetPlayer(i)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// card is a shorthand constructor for a face-up card.
func tysCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func TestTysiac_ResetDeal(t *testing.T) {
	g := newTestTysiac()
	assert.Equal(t, domain.TysiacPhaseBid, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 3, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetDeclarerIdx())
	assert.Equal(t, 0, g.GetTrumpSuit(), "trump starts unset")
	assert.Equal(t, domain.TysiacMinBid, g.GetCurrentBid())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerPlayer())

	// 7 cards each dealt (talon not yet taken).
	totalHand := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalHand += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, domain.TysiacHandSize*domain.TysiacPlayerCnt, totalHand)

	// Forehand is the dealer's left.
	assert.Equal(t, (g.GetDealerIdx()+1)%domain.TysiacPlayerCnt, g.GetForehandIdx())
	assert.Equal(t, g.GetForehandIdx(), g.GetCurrentPlayerIdx())
}

func TestTysiac_DeckIsUnique24(t *testing.T) {
	// Reconstruct the whole round's cards (7*3 hands + 3 talon via marshal) and
	// verify 24 unique cards, each of 9,J,Q,K,10,A across 4 suits.
	g := newTestTysiac()
	seen := map[int]bool{}
	count := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			key := c.GetDesign()*100 + c.GetValue()
			assert.False(t, seen[key], "duplicate card %d", key)
			seen[key] = true
			count++
		}
	}
	// 21 in hands; 3 remain in talon -> distinct hand cards = 21.
	assert.Equal(t, 21, count)
	// Every hand card value is one of the 24-deck ranks.
	valid := map[int]bool{1: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	for k := range seen {
		assert.True(t, valid[k%100], "unexpected rank %d", k%100)
	}
}

func TestTysiac_Bidding_RaiseAndPass(t *testing.T) {
	g := newTestTysiac()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.TysiacPhaseBid)
	// Human (player 0) raises: current bid climbs by the step.
	base := g.GetCurrentBid()
	require.NoError(t, g.PlayerBid(true))
	assert.Equal(t, base+domain.TysiacBidStep, g.GetCurrentBid())

	// Not human turn -> error.
	if !g.GetPlayer(g.GetCurrentPlayerIdx()).GetIsHuman() {
		assert.ErrorIs(t, g.PlayerBid(true), domain.ErrNotHumanTurn)
	}

	// Wrong phase -> error.
	g.SetPhase(domain.TysiacPhasePlay)
	assert.ErrorIs(t, g.PlayerBid(true), domain.ErrWrongPhase)
}

func TestTysiac_Bidding_EveryonePasses_ForehandTakesMin(t *testing.T) {
	g := newTestTysiac()
	// Force all CPUs to easy so they never raise; drive the whole auction.
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.TysiacCpuDifficultyEasy
	g.SetConfig(cfg)
	g.SetPhase(domain.TysiacPhaseBid)
	g.SetCurrentPlayerIdx(g.GetForehandIdx())

	guard := 0
	for g.GetPhase() == domain.TysiacPhaseBid && guard < 50 {
		guard++
		if g.GetPlayer(g.GetCurrentPlayerIdx()).GetIsHuman() {
			require.NoError(t, g.PlayerBid(false)) // human passes
		} else {
			g.CpuBid()
		}
	}
	// Auction resolved -> talon phase, a declarer chosen, contract >= min.
	assert.NotEqual(t, -1, g.GetDeclarerIdx())
	assert.GreaterOrEqual(t, g.GetContract(), domain.TysiacMinBid)
	assert.Contains(t,
		[]domain.TysiacPhase{domain.TysiacPhaseTalon, domain.TysiacPhasePlay},
		g.GetPhase())
}

func TestTysiac_Talon_DeclarerExchange(t *testing.T) {
	// Make the human the declarer so we can drive the discard by hand.
	g := newTestTysiac()
	g.SetDeclarerIdx(0)
	g.SetContract(120)
	// Simulate finalized auction -> talon phase. Manually stage: declarer already
	// took 3 talon cards (10 in hand) and must give one to each opponent.
	setTysiacHand(g, 0,
		tysCard(domain.CardDesignSpade, 1), tysCard(domain.CardDesignSpade, 10),
		tysCard(domain.CardDesignClover, 1), tysCard(domain.CardDesignClover, 10),
		tysCard(domain.CardDesignHeart, 1), tysCard(domain.CardDesignHeart, 10),
		tysCard(domain.CardDesignDiamond, 1), tysCard(domain.CardDesignDiamond, 10),
		tysCard(domain.CardDesignSpade, 9), tysCard(domain.CardDesignClover, 9))
	setTysiacHand(g, 1,
		tysCard(domain.CardDesignHeart, 13), tysCard(domain.CardDesignHeart, 12),
		tysCard(domain.CardDesignHeart, 11), tysCard(domain.CardDesignHeart, 9),
		tysCard(domain.CardDesignDiamond, 13), tysCard(domain.CardDesignDiamond, 12),
		tysCard(domain.CardDesignDiamond, 11))
	setTysiacHand(g, 2,
		tysCard(domain.CardDesignSpade, 13), tysCard(domain.CardDesignSpade, 12),
		tysCard(domain.CardDesignSpade, 11), tysCard(domain.CardDesignClover, 13),
		tysCard(domain.CardDesignClover, 12), tysCard(domain.CardDesignClover, 11),
		tysCard(domain.CardDesignDiamond, 9))
	g.SetPhase(domain.TysiacPhaseTalon)
	g.SetCurrentPlayerIdx(0)

	require.NoError(t, g.PlayerDiscard(0)) // give to first opponent
	require.NoError(t, g.PlayerDiscard(0)) // give to second opponent -> starts play

	// After exchange every player holds 8 cards.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, 8, g.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Equal(t, domain.TysiacPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetLeadPlayerIdx(), "declarer leads")
}

func TestTysiac_Talon_Errors(t *testing.T) {
	g := newTestTysiac()
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.TysiacPhaseTalon)
	setTysiacHand(g, 0, tysCard(domain.CardDesignSpade, 1))

	// Out-of-range index.
	assert.Error(t, g.PlayerDiscard(-1))
	assert.Error(t, g.PlayerDiscard(99))

	// Wrong phase.
	g.SetPhase(domain.TysiacPhasePlay)
	assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrWrongPhase)

	// Not-human declarer.
	g.SetPhase(domain.TysiacPhaseTalon)
	g.SetDeclarerIdx(1)
	assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrNotHumanTurn)
}

func TestTysiac_Marriage_SetsTrumpAndScores(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.TysiacPhasePlay)
	g.SetCurrentPlayerIdx(0)
	// Player 0 leads a Hearts King while holding the Hearts Queen -> marriage.
	setTysiacHand(g, 0,
		tysCard(domain.CardDesignHeart, 13), // K (index 0 after sort? sort by strength within suit)
		tysCard(domain.CardDesignHeart, 12), // Q
		tysCard(domain.CardDesignSpade, 9))
	// Find the King index (sorting may reorder).
	kingIdx := -1
	p := g.GetPlayer(0)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == 13 && p.GetCard(i).GetDesign() == domain.CardDesignHeart {
			kingIdx = i
		}
	}
	require.GreaterOrEqual(t, kingIdx, 0)
	require.NoError(t, g.PlayerPlay(kingIdx))

	assert.Equal(t, domain.CardDesignHeart, g.GetTrumpSuit(), "marriage sets trump")
	rm := g.GetRoundMarriage()
	assert.Equal(t, 100, rm[0], "hearts marriage = 100")
	assert.Len(t, g.GetCurrentTrick(), 1)
}

func TestTysiac_Marriage_SuitPointValues(t *testing.T) {
	cases := []struct {
		suit int
		pts  int
	}{
		{domain.CardDesignSpade, 40},
		{domain.CardDesignClover, 60},
		{domain.CardDesignDiamond, 80},
		{domain.CardDesignHeart, 100},
	}
	for _, c := range cases {
		g := newTestTysiac()
		g.SetTrumpSuit(0)
		g.SetCurrentTrick(nil)
		g.SetPhase(domain.TysiacPhasePlay)
		g.SetCurrentPlayerIdx(0)
		setTysiacHand(g, 0, tysCard(c.suit, 13), tysCard(c.suit, 12))
		p := g.GetPlayer(0)
		kingIdx := 0
		for i := 0; i < p.GetCardsSize(); i++ {
			if p.GetCard(i).GetValue() == 13 {
				kingIdx = i
			}
		}
		require.NoError(t, g.PlayerPlay(kingIdx))
		assert.Equal(t, c.suit, g.GetTrumpSuit())
		assert.Equal(t, c.pts, g.GetRoundMarriage()[0])
	}
}

func TestTysiac_TrickResolution_TrumpBeatsNonTrump(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(domain.CardDesignDiamond)
	g.SetTrickNumber(1)
	g.SetPhase(domain.TysiacPhaseTrickEnd)
	// Lead spade Ace (strongest non-trump); player 2 trumps with a diamond 9.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: tysCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 1, Card: tysCard(domain.CardDesignSpade, 10)},
		{PlayerIdx: 2, Card: tysCard(domain.CardDesignDiamond, 9)},
	})
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx(), "trump wins")
	// Card points collected: A(11)+10(10)+9(0)=21.
	assert.Equal(t, 21, g.GetRoundCardPoints()[2])
}

func TestTysiac_TrickResolution_RankOrder(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(0) // no trump: lead suit highest wins, A>10>K>Q>J>9
	g.SetTrickNumber(1)
	g.SetPhase(domain.TysiacPhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: tysCard(domain.CardDesignSpade, 13)}, // K
		{PlayerIdx: 1, Card: tysCard(domain.CardDesignSpade, 1)},  // A (highest)
		{PlayerIdx: 2, Card: tysCard(domain.CardDesignSpade, 10)}, // 10
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx(), "Ace outranks 10 and K")
}

func TestTysiac_TrickResolution_LastTrickEndsRound(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(0)
	g.SetTrickNumber(domain.TysiacTrickCount) // final trick
	g.SetPhase(domain.TysiacPhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: tysCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 1, Card: tysCard(domain.CardDesignSpade, 9)},
		{PlayerIdx: 2, Card: tysCard(domain.CardDesignSpade, 10)},
	})
	g.ResolveTrick()
	assert.Equal(t, domain.TysiacPhaseRoundEnd, g.GetPhase())
}

func TestTysiac_NextTrick(t *testing.T) {
	g := newTestTysiac()
	g.SetPhase(domain.TysiacPhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: tysCard(domain.CardDesignSpade, 1)}})
	g.NextTrick()
	assert.Equal(t, domain.TysiacPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 2, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())

	// Wrong phase is a no-op.
	g.SetPhase(domain.TysiacPhasePlay)
	g.NextTrick()
	assert.Equal(t, domain.TysiacPhasePlay, g.GetPhase())
}

func TestTysiac_ValidatePlay_FollowSuitAndTrump(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.TysiacPhasePlay)
	g.SetCurrentPlayerIdx(1)
	// Lead is spades; player 1 holds a spade (must follow) plus a heart trump.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: tysCard(domain.CardDesignSpade, 13)},
	})
	setTysiacHand(g, 1,
		tysCard(domain.CardDesignSpade, 1), // must follow suit
		tysCard(domain.CardDesignHeart, 9)) // trump (illegal while holding lead suit)
	spadeIdx := -1
	p := g.GetPlayer(1)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignSpade {
			spadeIdx = i
		}
	}
	// Only the spade is legal (holding the lead suit forbids playing the trump).
	valid := g.GetPlayableIndices(1)
	assert.Equal(t, []int{spadeIdx}, valid, "must follow the lead suit")
}

func TestTysiac_ValidatePlay_MustTrumpWhenVoid(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.TysiacPhasePlay)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: tysCard(domain.CardDesignSpade, 13)},
	})
	// Player 1 is void in spades but holds a trump and an off-suit -> must trump.
	setTysiacHand(g, 1,
		tysCard(domain.CardDesignHeart, 9),  // trump
		tysCard(domain.CardDesignClover, 1)) // off-suit (illegal)
	heartIdx := -1
	p := g.GetPlayer(1)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignHeart {
			heartIdx = i
		}
	}
	valid := g.GetPlayableIndices(1)
	assert.Equal(t, []int{heartIdx}, valid, "void in lead suit must play trump")
}

func TestTysiac_ValidatePlay_Overtrump(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.TysiacPhasePlay)
	// Trick already contains a heart Jack (trump); player 1 void in lead spades,
	// holds a higher heart (King) and a lower heart (9). Must overtrump -> only K.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: tysCard(domain.CardDesignSpade, 13)},
		{PlayerIdx: 2, Card: tysCard(domain.CardDesignHeart, 11)}, // trump J in trick
	})
	setTysiacHand(g, 1,
		tysCard(domain.CardDesignHeart, 13), // trump K (beats J)
		tysCard(domain.CardDesignHeart, 9))  // trump 9 (loses to J)
	kingIdx := -1
	p := g.GetPlayer(1)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == 13 {
			kingIdx = i
		}
	}
	valid := g.GetPlayableIndices(1)
	assert.Equal(t, []int{kingIdx}, valid, "must overtrump with the higher trump")
}

// TestTysiac_ValidatePlay_MustOvertrumpOnTrumpLead guards against the overtrump
// obligation being skipped when the trump suit itself is led. Following suit with
// a lower trump while holding a higher one must be rejected.
func TestTysiac_ValidatePlay_MustOvertrumpOnTrumpLead(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.TysiacPhasePlay)
	// Trump (hearts) is led with a Jack; player 1 follows suit and holds both a
	// higher trump (King) and a lower trump (9). Only the King is legal.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: tysCard(domain.CardDesignHeart, 11)}, // trump J led
	})
	setTysiacHand(g, 1,
		tysCard(domain.CardDesignHeart, 13), // trump K (beats J)
		tysCard(domain.CardDesignHeart, 9))  // trump 9 (loses to J)
	kingIdx := -1
	p := g.GetPlayer(1)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetValue() == 13 {
			kingIdx = i
		}
	}
	valid := g.GetPlayableIndices(1)
	assert.Equal(t, []int{kingIdx}, valid, "must overtrump even when trump is led")
}

func TestTysiac_ContractScoring_MetAndFailed(t *testing.T) {
	// Declarer meets contract -> +contract.
	g := newTestTysiac()
	g.SetDeclarerIdx(0)
	g.SetContract(100)
	g.SetPhase(domain.TysiacPhaseRoundEnd)
	g.SetRoundCardPoints([domain.TysiacPlayerCnt]int{80, 20, 20})
	g.SetRoundMarriage([domain.TysiacPlayerCnt]int{40, 0, 0}) // declarer total 120 >= 100
	g.ScoreRound()
	scores := g.GetPlayerScores()
	assert.Equal(t, 100, scores[0], "met -> +contract")
	assert.Equal(t, 20, scores[1], "opp rounded to 10")
	assert.Equal(t, 20, scores[2])

	// Declarer fails -> -contract.
	g2 := newTestTysiac()
	g2.SetDeclarerIdx(0)
	g2.SetContract(140)
	g2.SetPhase(domain.TysiacPhaseRoundEnd)
	g2.SetRoundCardPoints([domain.TysiacPlayerCnt]int{50, 35, 35})
	g2.SetRoundMarriage([domain.TysiacPlayerCnt]int{0, 0, 0})
	g2.ScoreRound()
	assert.Equal(t, -140, g2.GetPlayerScores()[0], "failed -> -contract")
}

func TestTysiac_ScoreRound_Rounding(t *testing.T) {
	g := newTestTysiac()
	g.SetDeclarerIdx(0)
	g.SetContract(100)
	g.SetPhase(domain.TysiacPhaseRoundEnd)
	// Non-declarer totals get rounded to nearest 10: 24->20, 26->30.
	g.SetRoundCardPoints([domain.TysiacPlayerCnt]int{100, 24, 26})
	g.SetRoundMarriage([domain.TysiacPlayerCnt]int{0, 0, 0})
	g.ScoreRound()
	scores := g.GetPlayerScores()
	assert.Equal(t, 20, scores[1])
	assert.Equal(t, 30, scores[2])
}

func TestTysiac_GameEnd_At1000(t *testing.T) {
	g := newTestTysiac()
	g.SetDeclarerIdx(0)
	g.SetContract(200)
	g.SetPhase(domain.TysiacPhaseRoundEnd)
	g.SetPlayerScores([domain.TysiacPlayerCnt]int{900, 100, 100})
	g.SetRoundCardPoints([domain.TysiacPlayerCnt]int{120, 0, 0})
	g.SetRoundMarriage([domain.TysiacPlayerCnt]int{100, 0, 0}) // total 220 >= 200 -> +200 = 1100
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerPlayer())
	assert.Equal(t, domain.TysiacPhaseGameEnd, g.GetPhase())
}

func TestTysiac_NextRound(t *testing.T) {
	g := newTestTysiac()
	g.SetPhase(domain.TysiacPhaseRoundEnd)
	prevDealer := g.GetDealerIdx()
	prevRound := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, prevRound+1, g.GetRoundNumber())
	assert.Equal(t, (prevDealer+1)%domain.TysiacPlayerCnt, g.GetDealerIdx())
	assert.Equal(t, domain.TysiacPhaseBid, g.GetPhase())

	// Wrong phase -> no-op.
	g.SetPhase(domain.TysiacPhasePlay)
	r := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, r, g.GetRoundNumber())
}

func TestTysiac_PlayerPlay_Errors(t *testing.T) {
	g := newTestTysiac()
	g.SetPhase(domain.TysiacPhasePlay)
	g.SetCurrentPlayerIdx(0)
	setTysiacHand(g, 0, tysCard(domain.CardDesignSpade, 1))

	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(99))

	// Wrong phase.
	g.SetPhase(domain.TysiacPhaseBid)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	// Not human turn.
	g.SetPhase(domain.TysiacPhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)
}

func TestTysiac_GetHint_AllPhases(t *testing.T) {
	// Bid phase hint.
	g := newTestTysiac()
	g.SetPhase(domain.TysiacPhaseBid)
	g.SetCurrentPlayerIdx(0)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Contains(t, []string{"bid_raise", "bid_pass"}, h.Reason)

	// Bid phase, not human's turn -> nil.
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())

	// Talon phase hint (human is declarer).
	g.SetPhase(domain.TysiacPhaseTalon)
	g.SetDeclarerIdx(0)
	setTysiacHand(g, 0, tysCard(domain.CardDesignSpade, 9), tysCard(domain.CardDesignSpade, 1))
	th := g.GetHint()
	require.NotNil(t, th)
	assert.Equal(t, "talon_discard", th.Reason)
	assert.Len(t, th.CardIndices, 1)

	// Talon phase, human not declarer -> nil.
	g.SetDeclarerIdx(1)
	assert.Nil(t, g.GetHint())

	// Play phase lead hint.
	g.SetPhase(domain.TysiacPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetTrumpSuit(0)
	setTysiacHand(g, 0,
		tysCard(domain.CardDesignHeart, 13), tysCard(domain.CardDesignHeart, 12))
	ph := g.GetHint()
	require.NotNil(t, ph)
	assert.Equal(t, "lead_marriage", ph.Reason)
	assert.Len(t, ph.CardIndices, 1)

	// Play phase, not human's turn -> nil.
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())

	// Unhandled phase -> nil.
	g.SetPhase(domain.TysiacPhaseTrickEnd)
	assert.Nil(t, g.GetHint())
}

func TestTysiac_GetHint_PlayReasons(t *testing.T) {
	g := newTestTysiac()
	g.SetPhase(domain.TysiacPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	// Following: player must play spades to a spade lead; higher card -> follow_win.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: tysCard(domain.CardDesignSpade, 12)}, // Q lead
	})
	setTysiacHand(g, 0,
		tysCard(domain.CardDesignSpade, 1), // A beats Q -> follow_win
		tysCard(domain.CardDesignSpade, 9)) // 9 loses -> follow_duck
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Contains(t, []string{"follow_win", "follow_duck"}, h.Reason)

	// Discard-low reason: void in lead suit and no trump held.
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: tysCard(domain.CardDesignSpade, 12)},
	})
	setTysiacHand(g, 0, tysCard(domain.CardDesignClover, 1)) // off-suit, no trump
	h2 := g.GetHint()
	require.NotNil(t, h2)
	assert.Equal(t, "discard_low", h2.Reason)
}

func TestTysiac_CpuBidAndPlay_FullRound(t *testing.T) {
	g := newTestTysiac()
	guard := 0
	for !g.GetGameEndFlag() && guard < 5000 {
		guard++
		switch g.GetPhase() {
		case domain.TysiacPhaseBid:
			if g.GetPlayer(g.GetCurrentPlayerIdx()).GetIsHuman() {
				require.NoError(t, g.PlayerBid(false)) // human always passes
			} else {
				g.CpuBid()
			}
		case domain.TysiacPhaseTalon:
			// If a human is the declarer, discard the first two cards.
			if g.GetPlayer(g.GetDeclarerIdx()).GetIsHuman() {
				require.NoError(t, g.PlayerDiscard(0))
			}
		case domain.TysiacPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
		case domain.TysiacPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == domain.TysiacPhaseTrickEnd {
				g.NextTrick()
			}
		case domain.TysiacPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case domain.TysiacPhaseGameEnd:
			guard = 5000
		}
	}
	assert.Less(t, guard, 5000, "game flow should progress")
}

func TestTysiac_Getters(t *testing.T) {
	g := newTestTysiac()
	g.SetRoundNumber(4)
	assert.Equal(t, 4, g.GetRoundNumber())
	g.SetTrickNumber(3)
	assert.Equal(t, 3, g.GetTrickNumber())
	g.SetCurrentPlayerIdx(2)
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	g.SetLeadPlayerIdx(1)
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
	g.SetContract(130)
	assert.Equal(t, 130, g.GetContract())
	g.SetCurrentBid(150)
	assert.Equal(t, 150, g.GetCurrentBid())
	g.SetTrumpSuit(domain.CardDesignClover)
	assert.Equal(t, domain.CardDesignClover, g.GetTrumpSuit())
	g.SetDeclarerIdx(2)
	assert.Equal(t, 2, g.GetDeclarerIdx())
	g.SetPlayerScores([domain.TysiacPlayerCnt]int{10, 20, 30})
	assert.Equal(t, [domain.TysiacPlayerCnt]int{10, 20, 30}, g.GetPlayerScores())

	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetPlayer(0))
	assert.NotNil(t, g.GetConfig())
	// GetActionLog is exercised (may be nil early); assert it does not panic.
	_ = g.GetActionLog()

	// GetPlayableIndices guards.
	assert.Nil(t, g.GetPlayableIndices(-1))
	g.SetPhase(domain.TysiacPhaseBid)
	assert.Nil(t, g.GetPlayableIndices(0), "not play phase -> nil")

	// IsHumanTurn / IsHumanBidTurn.
	g.SetPhase(domain.TysiacPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetPhase(domain.TysiacPhaseBid)
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanBidTurn())
	g.SetPhase(domain.TysiacPhasePlay)
	assert.False(t, g.IsHumanBidTurn())
}

func TestTysiac_JSON_RoundTrip(t *testing.T) {
	g := newTestTysiac()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetDeclarerIdx(0)
	g.SetContract(120)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Tysiac
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), g2.GetTrumpSuit())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
	assert.Equal(t, g.GetDeclarerIdx(), g2.GetDeclarerIdx())
	assert.Equal(t, g.GetContract(), g2.GetContract())
}

func TestTysiac_JSON_Invalid(t *testing.T) {
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	cases := []string{
		`not json`,                                                            // malformed
		`{"ph":0,"ps":[null,null]}`,                                           // wrong player count
		`{"ph":0,"ps":[null,null,null]}`,                                      // nil players
		`{"ph":0,"ps":` + okPlayers + `,"ci":100}`,                            // currentPlayerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"ci":-1}`,                             // currentPlayerIdx negative
		`{"ph":0,"ps":` + okPlayers + `,"di":99}`,                             // dealerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"fh":99}`,                             // forehandIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"li":99}`,                             // leadPlayerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"li":-2}`,                             // leadPlayerIdx below -1
		`{"ph":0,"ps":` + okPlayers + `,"dc":99}`,                             // declarerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"lt":99}`,                             // lastTrickWinner out of range
		`{"ph":0,"ps":` + okPlayers + `,"wp":99}`,                             // winnerPlayer out of range
		`{"ph":0,"ps":` + okPlayers + `,"dn":99}`,                             // discardCount out of range
		`{"ph":99,"ps":` + okPlayers + `}`,                                    // bad phase
		`{"ph":0,"ps":` + okPlayers + `,"ts":9}`,                              // bad trumpSuit
		`{"ph":0,"ps":` + okPlayers + `,"ct":[null]}`,                         // nil trick card
		`{"ph":0,"ps":` + okPlayers + `,"ct":[{"pi":99,"c":{"d":1,"v":13}}]}`, // trick card PlayerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"ct":[{"pi":-1,"c":{"d":1,"v":13}}]}`, // trick card PlayerIdx negative
		`{"ph":0,"ps":` + okPlayers + `,"tl":[null]}`,                         // nil talon card
	}
	for _, c := range cases {
		var g domain.Tysiac
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}

	// Config validation failure (bad CPU difficulty) is rejected.
	badCfg := `{"ph":0,"ps":` + okPlayers + `,"cf":{"cd":99,"tp":1000}}`
	var gc domain.Tysiac
	assert.Error(t, json.Unmarshal([]byte(badCfg), &gc))

	// Valid restore: trumpSuit=0 (unset) with a valid config and 3 players.
	okJSON := `{"ph":0,"ps":` + okPlayers + `,"ts":0,"cf":{"cd":1,"tp":1000},"dc":-1,"lt":-1,"wp":-1,"li":-1}`
	var g2 domain.Tysiac
	assert.NoError(t, json.Unmarshal([]byte(okJSON), &g2))
	assert.Equal(t, 3, g2.GetPlayerCnt())
	// Nil trumpCards in JSON falls back to a fresh deck (no panic on later use).
	assert.NotNil(t, g2.GetPlayer(0))
}

func TestTysiacPlayer_JSON_And_ResetRound(t *testing.T) {
	p := domain.NewTysiacPlayer(true)
	p.AddCard(tysCard(domain.CardDesignSpade, 1))
	p.AddTrick([]*domain.Card{tysCard(domain.CardDesignHeart, 10)})
	assert.Equal(t, 1, p.GetTrickCount())

	b, err := json.Marshal(p)
	require.NoError(t, err)
	var p2 domain.TysiacPlayer
	require.NoError(t, json.Unmarshal(b, &p2))
	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetCardsSize())
	assert.Equal(t, 1, p2.GetTrickCount())

	p2.ResetRound()
	assert.Equal(t, 0, p2.GetCardsSize())
	assert.Equal(t, 0, p2.GetTrickCount())
	assert.False(t, p2.GetIsFinished())

	// Malformed JSON -> error.
	assert.Error(t, json.Unmarshal([]byte(`not json`), &p2))
	// Empty object -> defaults applied, non-human GamePlayer created.
	var p3 domain.TysiacPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p3))
	assert.False(t, p3.GetIsHuman())
}

func TestTysiacConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultTysiacConfig().Validate())
	assert.Equal(t, domain.TysiacWinTarget, domain.DefaultTysiacConfig().TargetPoints)
	assert.Equal(t, domain.TysiacCpuDifficultyNormal, domain.DefaultTysiacConfig().CpuDifficulty)

	// Bad CPU difficulty.
	assert.Error(t, domain.TysiacConfig{CpuDifficulty: 99, TargetPoints: 1000}.Validate())
	// Non-positive target points.
	assert.Error(t, domain.TysiacConfig{CpuDifficulty: domain.TysiacCpuDifficultyEasy, TargetPoints: 0}.Validate())
}

// #5687: 結婚できるスートの判定は Web ページが自前で複製していた。
// 宣言できるのは K と Q を **両方** 持つスートだけで、点はスートごとに違う。
func TestTysiac_GetMarriageOptions(t *testing.T) {
	g := newTestTysiac()
	setTysiacHand(g, 0,
		tysCard(domain.CardDesignSpade, 13), tysCard(domain.CardDesignSpade, 12),
		tysCard(domain.CardDesignHeart, 13),   // Q が無いので結婚にならない
		tysCard(domain.CardDesignDiamond, 12), // K が無いので結婚にならない
		tysCard(domain.CardDesignClover, 13), tysCard(domain.CardDesignClover, 12),
	)

	got := g.GetMarriageOptions(0)

	assert.Equal(t, []domain.TysiacMarriageOption{
		{Suit: domain.CardDesignSpade, Points: 40},
		{Suit: domain.CardDesignClover, Points: 60},
	}, got)

	t.Run("no pair yields nothing", func(t *testing.T) {
		setTysiacHand(g, 1, tysCard(domain.CardDesignHeart, 13), tysCard(domain.CardDesignSpade, 12))
		assert.Empty(t, g.GetMarriageOptions(1))
	})

	t.Run("out-of-range index is not a panic", func(t *testing.T) {
		assert.Empty(t, g.GetMarriageOptions(-1))
		assert.Empty(t, g.GetMarriageOptions(domain.TysiacPlayerCnt))
	})
}
