//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestOmbre returns a fresh, reset Ombre game with the default 3-player setup.
func newTestOmbre() *domain.Ombre {
	g := domain.NewDefaultOmbre()
	g.Reset()
	return g
}

// setOmbreHand replaces player i's hand with the supplied cards deterministically.
func setOmbreHand(g *domain.Ombre, i int, cards ...*domain.Card) {
	p := g.GetPlayer(i)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// ombreCard is a shorthand constructor for a face-up card.
func ombreCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// ombreGiveTricks awards player i exactly n tricks by adding dummy trick groups.
func ombreGiveTricks(g *domain.Ombre, i, n int) {
	p := g.GetPlayer(i)
	for k := 0; k < n; k++ {
		p.AddTrick([]*domain.Card{ombreCard(domain.CardDesignSpade, 2)})
	}
}

func TestOmbre_ResetDeal(t *testing.T) {
	g := newTestOmbre()
	assert.Equal(t, domain.OmbrePhaseBid, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 3, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetOmbreIdx())
	assert.Equal(t, -1, g.GetTrumpSuit())
	assert.Equal(t, domain.OmbreBidNone, g.GetWinningBid())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerPlayer())

	totalHand := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalHand += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, domain.OmbreHandSize*domain.OmbrePlayerCnt, totalHand)

	assert.Equal(t, (g.GetDealerIdx()+1)%domain.OmbrePlayerCnt, g.GetForehandIdx())
	assert.Equal(t, g.GetForehandIdx(), g.GetCurrentBidderIdx())
}

func TestOmbre_DeckIsUnique40(t *testing.T) {
	g := newTestOmbre()
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
	// 27 dealt to hands; 13 remain unused in the stock.
	assert.Equal(t, 27, count)
	valid := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 11: true, 12: true, 13: true}
	for k := range seen {
		assert.True(t, valid[k%100], "unexpected rank %d", k%100)
	}
}

func TestOmbre_Bidding_WrongPhaseAndSuitRequired(t *testing.T) {
	g := newTestOmbre()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.OmbreCpuDifficultyEasy // CPUs always pass
	g.SetConfig(cfg)
	g.SetPhase(domain.OmbrePhaseBid)

	// Drive CPU bids until it is the human's (seat 0) turn.
	guard := 0
	for g.GetPhase() == domain.OmbrePhaseBid && g.GetCurrentBidderIdx() != 0 && guard < 10 {
		guard++
		g.CpuBid()
	}
	if g.GetPhase() == domain.OmbrePhaseBid && g.GetCurrentBidderIdx() == 0 {
		// Entrar without a valid trump suit is rejected.
		assert.Error(t, g.PlayerBid(domain.OmbreBidEntrar, -1))
		// Entrar with a valid trump suit succeeds.
		require.NoError(t, g.PlayerBid(domain.OmbreBidEntrar, domain.CardDesignHeart))
	}

	// Wrong phase -> error.
	g.SetPhase(domain.OmbrePhasePlay)
	assert.ErrorIs(t, g.PlayerBid(domain.OmbreBidEntrar, domain.CardDesignHeart), domain.ErrWrongPhase)
}

func TestOmbre_Bidding_EveryonePasses_DealerForced(t *testing.T) {
	g := newTestOmbre()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.OmbreCpuDifficultyEasy // CPUs always pass
	g.SetConfig(cfg)
	g.SetPhase(domain.OmbrePhaseBid)

	guard := 0
	for g.GetPhase() == domain.OmbrePhaseBid && guard < 50 {
		guard++
		if g.GetPlayer(g.GetCurrentBidderIdx()).GetIsHuman() {
			require.NoError(t, g.PlayerBid(domain.OmbreBidNone, -1)) // human passes
		} else {
			g.CpuBid()
		}
	}
	// Auction resolved -> dealer forced to be Ombre with a chosen trump.
	assert.Equal(t, g.GetDealerIdx(), g.GetOmbreIdx())
	assert.GreaterOrEqual(t, int(g.GetWinningBid()), int(domain.OmbreBidEntrar))
	assert.True(t, g.GetTrumpSuit() >= domain.CardDesignSpade && g.GetTrumpSuit() <= domain.CardDesignDiamond)
	assert.Equal(t, domain.OmbrePhasePlay, g.GetPhase())
}

func TestOmbre_BidOrdering(t *testing.T) {
	assert.Greater(t, int(domain.OmbreBidSolo), int(domain.OmbreBidEntrar))
	assert.Greater(t, int(domain.OmbreBidEntrar), int(domain.OmbreBidNone))
}

func TestOmbre_MatadorRanking(t *testing.T) {
	resolve := func(trump int, trick []*domain.TrickCard) int {
		g := newTestOmbre()
		g.SetOmbreIdx(0)
		g.SetTrumpSuit(trump)
		g.SetTrickNumber(1) // not the last trick, avoid triggering round-end scoring
		g.SetPhase(domain.OmbrePhaseTrickEnd)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		return g.GetLeadPlayerIdx()
	}

	// Spadille (♠A) beats Manille (7 of trump) and Basto (♣A) even when spades isn't trump.
	assert.Equal(t, 1, resolve(domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ombreCard(domain.CardDesignHeart, 7)},  // Manille
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignSpade, 1)},  // Spadille (highest)
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignClover, 1)}, // Basto
	}))

	// Manille (7 of trump) beats Basto, Punto and the trump King.
	assert.Equal(t, 2, resolve(domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ombreCard(domain.CardDesignHeart, 1)},  // Punto
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignHeart, 13)}, // trump K
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignHeart, 7)},  // Manille (highest)
	}))

	// Basto (♣A) beats Punto (red-trump Ace) and lower trumps.
	assert.Equal(t, 0, resolve(domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ombreCard(domain.CardDesignClover, 1)}, // Basto (highest here)
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignHeart, 1)},  // Punto
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignHeart, 13)}, // trump K
	}))

	// Any trump beats any plain card (low trump 2 beats plain King).
	assert.Equal(t, 1, resolve(domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ombreCard(domain.CardDesignSpade, 13)}, // plain K (lead)
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignHeart, 2)},  // trump 2
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignSpade, 12)}, // plain Q
	}))

	// No Punto for a BLACK trump: ♠A is Spadille, trump-suit K ranks below it.
	assert.Equal(t, 0, resolve(domain.CardDesignSpade, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ombreCard(domain.CardDesignSpade, 1)},  // Spadille
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignSpade, 13)}, // trump K
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignSpade, 2)},  // trump 2
	}))
}

func TestOmbre_PlainSuit_AceLowAndOffSuit(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetTrickNumber(1)
	g.SetPhase(domain.OmbrePhaseTrickEnd)
	// Plain red suit (diamonds): K>Q>J>A>2>3>4>5>6>7 — the Ace is the
	// 4th-highest (outranking 2..7), NOT low as in the black plain suits.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: ombreCard(domain.CardDesignDiamond, 2)}, // 2
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignDiamond, 1)}, // A beats 2 and 7
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignDiamond, 7)}, // 7 (lowest red plain)
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx(), "red plain Ace outranks 2 and 7")

	// Off-suit plain cannot win; a higher follower of the led suit does.
	// Red ranking: diamond 3 outranks diamond 7.
	g.SetTrickNumber(1)
	g.SetPhase(domain.OmbrePhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: ombreCard(domain.CardDesignDiamond, 7)}, // lead 7 (lowest red)
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignClover, 13)}, // off-suit K (cannot win, ♣K is plain)
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignDiamond, 3)}, // follows, higher (red 3 > red 7)
	})
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx(), "off-suit plain cannot win")
}

func TestOmbre_MustFollow_TrumpGroupIsASuit(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(0)
	g.SetTrumpSuit(domain.CardDesignDiamond)
	g.SetPhase(domain.OmbrePhasePlay)
	g.SetCurrentPlayerIdx(1)
	// Trump (diamond) is led -> ♠A counts as trump and must be followed.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: ombreCard(domain.CardDesignDiamond, 5)},
	})
	setOmbreHand(g, 1,
		ombreCard(domain.CardDesignSpade, 1), // ♠A is trump -> must follow
		ombreCard(domain.CardDesignHeart, 3)) // plain heart (illegal while holding trump)
	valid := g.GetPlayableIndices(1)
	assert.Equal(t, []int{0}, valid, "must follow trump group with ♠A")

	// Void in trump -> any card playable.
	setOmbreHand(g, 1,
		ombreCard(domain.CardDesignClover, 13),
		ombreCard(domain.CardDesignHeart, 2))
	assert.Len(t, g.GetPlayableIndices(1), 2, "void in trump: all cards playable")
}

func TestOmbre_NextTrick(t *testing.T) {
	g := newTestOmbre()
	g.SetPhase(domain.OmbrePhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: ombreCard(domain.CardDesignSpade, 1)}})
	g.NextTrick()
	assert.Equal(t, domain.OmbrePhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 2, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())

	// Wrong phase is a no-op.
	g.SetPhase(domain.OmbrePhasePlay)
	g.NextTrick()
	assert.Equal(t, domain.OmbrePhasePlay, g.GetPhase())
}

func TestOmbre_Outcomes(t *testing.T) {
	cases := []struct {
		name    string
		tricks  [domain.OmbrePlayerCnt]int
		outcome domain.OmbreOutcome
		scores  [domain.OmbrePlayerCnt]int
	}{
		{"sacar", [domain.OmbrePlayerCnt]int{5, 2, 2}, domain.OmbreOutcomeSacar, [domain.OmbrePlayerCnt]int{2, -1, -1}},
		{"puesta", [domain.OmbrePlayerCnt]int{3, 3, 3}, domain.OmbreOutcomePuesta, [domain.OmbrePlayerCnt]int{-2, 1, 1}},
		{"codille", [domain.OmbrePlayerCnt]int{2, 4, 3}, domain.OmbreOutcomeCodille, [domain.OmbrePlayerCnt]int{-4, 2, 2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := newTestOmbre()
			g.SetOmbreIdx(0)
			g.SetWinningBid(domain.OmbreBidEntrar)
			g.SetPhase(domain.OmbrePhaseRoundEnd)
			for i := 0; i < domain.OmbrePlayerCnt; i++ {
				ombreGiveTricks(g, i, c.tricks[i])
			}
			g.ScoreRound()
			assert.Equal(t, c.outcome, g.GetOutcome())
			assert.Equal(t, c.scores, g.GetPlayerScores())

			// ScoreRound is idempotent (scored flag).
			g.ScoreRound()
			assert.Equal(t, c.scores, g.GetPlayerScores())
		})
	}
}

func TestOmbre_ScoreRound_WrongPhaseNoop(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(0)
	g.SetPhase(domain.OmbrePhasePlay)
	g.ScoreRound()
	assert.Equal(t, [domain.OmbrePlayerCnt]int{0, 0, 0}, g.GetPlayerScores())
}

func TestOmbre_GameEnd_HumanWins(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(0)
	g.SetWinningBid(domain.OmbreBidSolo)
	g.SetRoundNumber(domain.OmbreWinRounds) // final deal
	g.SetPhase(domain.OmbrePhaseRoundEnd)
	ombreGiveTricks(g, 0, 9) // human (Ombre) sweeps -> Sacar
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerPlayer())
	assert.Equal(t, domain.OmbrePhaseGameEnd, g.GetPhase())
	assert.Equal(t, domain.OmbreResultWin, g.GetResult())
}

func TestOmbre_GameEnd_HumanLoses(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(1) // a CPU is Ombre
	g.SetWinningBid(domain.OmbreBidSolo)
	g.SetRoundNumber(domain.OmbreWinRounds)
	g.SetPhase(domain.OmbrePhaseRoundEnd)
	g.SetPlayerScores([domain.OmbrePlayerCnt]int{0, 5, 0})
	ombreGiveTricks(g, 1, 9) // CPU1 sweeps -> Sacar for CPU1
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerPlayer())
	assert.Equal(t, domain.OmbreResultLose, g.GetResult())
}

func TestOmbre_NextRound(t *testing.T) {
	g := newTestOmbre()
	g.SetPhase(domain.OmbrePhaseRoundEnd)
	prevDealer := g.GetDealerIdx()
	prevRound := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, prevRound+1, g.GetRoundNumber())
	assert.Equal(t, (prevDealer+1)%domain.OmbrePlayerCnt, g.GetDealerIdx())
	assert.Equal(t, domain.OmbrePhaseBid, g.GetPhase())

	// Wrong phase -> no-op.
	g.SetPhase(domain.OmbrePhasePlay)
	r := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, r, g.GetRoundNumber())
}

func TestOmbre_PlayerPlay_Errors(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.OmbrePhasePlay)
	g.SetCurrentPlayerIdx(0)
	setOmbreHand(g, 0, ombreCard(domain.CardDesignSpade, 13))

	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(99))

	// Wrong phase.
	g.SetPhase(domain.OmbrePhaseBid)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	// Not human turn.
	g.SetPhase(domain.OmbrePhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)
}

func TestOmbre_PlayerPlay_FollowSuitViolation(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.OmbrePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignDiamond, 13)}, // plain diamond lead
	})
	setOmbreHand(g, 0,
		ombreCard(domain.CardDesignDiamond, 5), // legal (follows diamond)
		ombreCard(domain.CardDesignClover, 3))  // illegal (must follow diamond)
	cloverIdx := -1
	p := g.GetPlayer(0)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignClover {
			cloverIdx = i
		}
	}
	assert.ErrorIs(t, g.PlayerPlay(cloverIdx), domain.ErrInvalidPlay)
}

func TestOmbre_PlayerPlay_SuccessAndTrickComplete(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.OmbrePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: ombreCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignSpade, 6)},
	})
	setOmbreHand(g, 0, ombreCard(domain.CardDesignSpade, 13))
	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, domain.OmbrePhaseTrickEnd, g.GetPhase(), "third card completes the trick")
}

func TestOmbre_GetHint_AllPhases(t *testing.T) {
	// Bid phase hint.
	g := newTestOmbre()
	g.SetPhase(domain.OmbrePhaseBid)
	guard := 0
	for g.GetPhase() == domain.OmbrePhaseBid && g.GetCurrentBidderIdx() != 0 && guard < 10 {
		guard++
		g.CpuBid()
	}
	if g.GetPhase() == domain.OmbrePhaseBid && g.GetCurrentBidderIdx() == 0 {
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Contains(t, []string{"bid_entrar", "bid_solo", "bid_pass"}, h.Reason)
	}

	// Play phase lead hint (human is Ombre -> lead_high).
	g2 := newTestOmbre()
	g2.SetPhase(domain.OmbrePhasePlay)
	g2.SetOmbreIdx(0)
	g2.SetTrumpSuit(domain.CardDesignHeart)
	g2.SetCurrentPlayerIdx(0)
	g2.SetCurrentTrick(nil)
	setOmbreHand(g2, 0, ombreCard(domain.CardDesignHeart, 13), ombreCard(domain.CardDesignSpade, 4))
	ph := g2.GetHint()
	require.NotNil(t, ph)
	assert.Equal(t, "lead_high", ph.Reason)
	assert.Len(t, ph.CardIndices, 1)

	// Play phase lead hint (human is coalition -> lead_low).
	g2.SetOmbreIdx(1)
	g2.SetCurrentPlayerIdx(0)
	g2.SetCurrentTrick(nil)
	lh := g2.GetHint()
	require.NotNil(t, lh)
	assert.Equal(t, "lead_low", lh.Reason)

	// Play phase, not human's turn -> nil.
	g2.SetCurrentPlayerIdx(1)
	assert.Nil(t, g2.GetHint())

	// Unhandled phase -> nil.
	g2.SetPhase(domain.OmbrePhaseTrickEnd)
	assert.Nil(t, g2.GetHint())
}

func TestOmbre_GetHint_PlayReasons(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(2) // player 0 is coalition
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.OmbrePhasePlay)
	g.SetCurrentPlayerIdx(0)
	// Opponent (Ombre seat 2) leads a plain diamond Queen; player 0 can win or duck.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignDiamond, 12)},
	})
	setOmbreHand(g, 0,
		ombreCard(domain.CardDesignDiamond, 13), // K beats Q -> follow_win
		ombreCard(domain.CardDesignDiamond, 4))  // 4 loses -> follow_duck
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Contains(t, []string{"follow_win", "follow_duck"}, h.Reason)

	// discard_low: void in lead suit.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignDiamond, 12)},
	})
	setOmbreHand(g, 0, ombreCard(domain.CardDesignClover, 13)) // off-suit only
	h2 := g.GetHint()
	require.NotNil(t, h2)
	assert.Equal(t, "discard_low", h2.Reason)

	// give_partner: partner (same side) is winning.
	g.SetOmbreIdx(1) // players 0 and 2 are coalition (same side)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: ombreCard(domain.CardDesignDiamond, 13)}, // partner winning with K
	})
	setOmbreHand(g, 0,
		ombreCard(domain.CardDesignDiamond, 11), // J loses to K
		ombreCard(domain.CardDesignDiamond, 7))  // 7 loses to K
	h3 := g.GetHint()
	require.NotNil(t, h3)
	assert.Equal(t, "give_partner", h3.Reason)
}

func TestOmbre_CpuFullRound(t *testing.T) {
	g := newTestOmbre()
	guard := 0
	for !g.GetGameEndFlag() && guard < 8000 {
		guard++
		switch g.GetPhase() {
		case domain.OmbrePhaseBid:
			if g.GetPlayer(g.GetCurrentBidderIdx()).GetIsHuman() {
				require.NoError(t, g.PlayerBid(domain.OmbreBidNone, -1)) // human passes
			} else {
				g.CpuBid()
			}
		case domain.OmbrePhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
		case domain.OmbrePhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == domain.OmbrePhaseTrickEnd {
				g.NextTrick()
			}
		case domain.OmbrePhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case domain.OmbrePhaseGameEnd:
			guard = 8000
		}
	}
	assert.Less(t, guard, 8000, "game flow should progress")
}

func TestOmbre_Getters(t *testing.T) {
	g := newTestOmbre()
	g.SetRoundNumber(4)
	assert.Equal(t, 4, g.GetRoundNumber())
	g.SetTrickNumber(3)
	assert.Equal(t, 3, g.GetTrickNumber())
	g.SetCurrentPlayerIdx(2)
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	g.SetLeadPlayerIdx(1)
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
	g.SetWinningBid(domain.OmbreBidSolo)
	assert.Equal(t, domain.OmbreBidSolo, g.GetWinningBid())
	g.SetOmbreIdx(2)
	assert.Equal(t, 2, g.GetOmbreIdx())
	g.SetTrumpSuit(domain.CardDesignClover)
	assert.Equal(t, domain.CardDesignClover, g.GetTrumpSuit())
	g.SetPlayerScores([domain.OmbrePlayerCnt]int{10, 20, 30})
	assert.Equal(t, [domain.OmbrePlayerCnt]int{10, 20, 30}, g.GetPlayerScores())

	assert.GreaterOrEqual(t, g.GetForehandIdx(), 0)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetPlayer(0))
	assert.NotNil(t, g.GetConfig())
	assert.Equal(t, domain.OmbreOutcomeNone, g.GetOutcome())
	assert.Equal(t, domain.OmbreResultNone, g.GetResult())
	_ = g.GetActionLog()

	assert.Nil(t, g.GetPlayableIndices(-1))
	g.SetPhase(domain.OmbrePhaseBid)
	assert.Nil(t, g.GetPlayableIndices(0), "not play phase -> nil")

	g.SetPhase(domain.OmbrePhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetPhase(domain.OmbrePhaseBid)
	assert.Equal(t, g.GetPlayer(g.GetCurrentBidderIdx()).GetIsHuman(), g.IsHumanBidTurn())
	g.SetPhase(domain.OmbrePhasePlay)
	assert.False(t, g.IsHumanBidTurn())
}

func TestOmbre_JSON_RoundTrip(t *testing.T) {
	g := newTestOmbre()
	g.SetOmbreIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetWinningBid(domain.OmbreBidSolo)
	g.SetPhase(domain.OmbrePhasePlay)
	g.SetLeadPlayerIdx(0)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Ombre
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
	assert.Equal(t, g.GetOmbreIdx(), g2.GetOmbreIdx())
	assert.Equal(t, g.GetTrumpSuit(), g2.GetTrumpSuit())
	assert.Equal(t, g.GetWinningBid(), g2.GetWinningBid())
}

func TestOmbre_JSON_Invalid(t *testing.T) {
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	cases := []string{
		`not json`,
		`{"ph":0,"ps":[null,null],"ts":-1}`,                                           // wrong player count
		`{"ph":0,"ps":[null,null,null],"ts":-1}`,                                      // nil players
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ci":100}`,                            // ci out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ci":-1}`,                             // ci negative
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"di":99}`,                             // di out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"fh":99}`,                             // fh out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"cbi":99}`,                            // cbi out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"li":99}`,                             // li out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"li":-2}`,                             // li below -1
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"om":99}`,                             // ombreIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"lt":99}`,                             // lastTrickWinner out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"wp":99}`,                             // winnerPlayer out of range
		`{"ph":99,"ps":` + okPlayers + `,"ts":-1}`,                                    // bad phase
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"wb":9}`,                              // bad winning bid
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"bd":[9,0,0]}`,                        // bad bid element
		`{"ph":0,"ps":` + okPlayers + `,"ts":9}`,                                      // bad trump suit
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"bt":[9,0,0]}`,                        // bad bidTrump element
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"oc":9}`,                              // bad outcome
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"rs":9}`,                              // bad result
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ct":[null]}`,                         // nil trick card
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ct":[{"pi":99,"c":{"d":1,"v":13}}]}`, // trick card idx out of range
		`{"ph":1,"ps":` + okPlayers + `,"om":-1,"li":0,"ts":3}`,                       // play phase requires ombre set
		`{"ph":1,"ps":` + okPlayers + `,"om":0,"li":-1,"ts":3}`,                       // play phase requires lead set
		`{"ph":1,"ps":` + okPlayers + `,"om":0,"li":0,"ts":-1}`,                       // play phase requires trump set
	}
	for _, c := range cases {
		var g domain.Ombre
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}

	// Config validation failure (bad CPU difficulty) is rejected.
	badCfg := `{"ph":0,"ps":` + okPlayers + `,"ts":-1,"cf":{"cd":99,"tr":5}}`
	var gc domain.Ombre
	assert.Error(t, json.Unmarshal([]byte(badCfg), &gc))

	// Valid restore.
	okJSON := `{"ph":0,"ps":` + okPlayers + `,"wb":0,"cf":{"cd":1,"tr":5},"lt":-1,"wp":-1,"li":-1,"om":-1,"ts":-1}`
	var g2 domain.Ombre
	assert.NoError(t, json.Unmarshal([]byte(okJSON), &g2))
	assert.Equal(t, 3, g2.GetPlayerCnt())
	assert.NotNil(t, g2.GetPlayer(0))
}

func TestOmbrePlayer_JSON_And_ResetRound(t *testing.T) {
	p := domain.NewOmbrePlayer(true)
	p.AddCard(ombreCard(domain.CardDesignSpade, 1))
	p.AddTrick([]*domain.Card{ombreCard(domain.CardDesignHeart, 13)})
	assert.Equal(t, 1, p.GetTrickCount())

	b, err := json.Marshal(p)
	require.NoError(t, err)
	var p2 domain.OmbrePlayer
	require.NoError(t, json.Unmarshal(b, &p2))
	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetCardsSize())
	assert.Equal(t, 1, p2.GetTrickCount())

	p2.ResetRound()
	assert.Equal(t, 0, p2.GetCardsSize())
	assert.Equal(t, 0, p2.GetTrickCount())
	assert.False(t, p2.GetIsFinished())

	assert.Error(t, json.Unmarshal([]byte(`not json`), &p2))
	var p3 domain.OmbrePlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p3))
	assert.False(t, p3.GetIsHuman())
}

func TestOmbreConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultOmbreConfig().Validate())
	assert.Equal(t, domain.OmbreWinRounds, domain.DefaultOmbreConfig().TargetRounds)
	assert.Equal(t, domain.OmbreCpuDifficultyNormal, domain.DefaultOmbreConfig().CpuDifficulty)

	assert.Error(t, domain.OmbreConfig{CpuDifficulty: 99, TargetRounds: 5}.Validate())
	assert.Error(t, domain.OmbreConfig{CpuDifficulty: domain.OmbreCpuDifficultyEasy, TargetRounds: 0}.Validate())
}
