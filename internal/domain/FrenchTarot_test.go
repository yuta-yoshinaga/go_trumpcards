//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers (frenchtarot-prefixed) ---

func frenchTarotTrumpCard(v int) *domain.Card {
	return domain.NewCard(domain.FrenchTarotTrumpDesign, v, false)
}

func frenchTarotExcuseCard() *domain.Card {
	return domain.NewCard(domain.FrenchTarotExcuseDesign, domain.FrenchTarotExcuseValue, false)
}

func frenchTarotSuitCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func frenchTarotNewReset() *domain.FrenchTarot {
	g := domain.NewDefaultFrenchTarot()
	g.Reset()
	return g
}

func frenchTarotSetHand(g *domain.FrenchTarot, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func frenchTarotTrickCards(cards ...*domain.TrickCard) []*domain.TrickCard {
	return cards
}

// --- Deck ---

func TestFrenchTarotDeckIs78(t *testing.T) {
	deck := domain.BuildFrenchTarotDeckPublic()
	require.Len(t, deck, domain.FrenchTarotDeckSize)
	suits, trumps, excuse := 0, 0, 0
	total := 0
	for _, c := range deck {
		total += domain.FrenchTarotCardHalfPointsPublic(c)
		switch c.GetDesign() {
		case domain.FrenchTarotExcuseDesign:
			excuse++
			assert.Equal(t, domain.FrenchTarotExcuseValue, c.GetValue())
		case domain.FrenchTarotTrumpDesign:
			trumps++
			assert.GreaterOrEqual(t, c.GetValue(), 1)
			assert.LessOrEqual(t, c.GetValue(), domain.FrenchTarotMaxTrump)
		default:
			suits++
		}
	}
	assert.Equal(t, 56, suits)
	assert.Equal(t, 21, trumps)
	assert.Equal(t, 1, excuse)
	// Total = 91 points = 182 half-points.
	assert.Equal(t, 182, total)
}

func TestFrenchTarotDealDistribution(t *testing.T) {
	g := frenchTarotNewReset()
	assert.Equal(t, domain.FrenchTarotPhaseBid, g.GetPhase())
	assert.Equal(t, domain.FrenchTarotChienSize, g.GetChienCount())
	for i := 0; i < domain.FrenchTarotPlayerCnt; i++ {
		assert.Equal(t, domain.FrenchTarotHandSize, g.GetPlayer(i).GetCardsSize())
	}
}

// --- Card classification / points ---

func TestFrenchTarotClassification(t *testing.T) {
	assert.True(t, domain.FrenchTarotIsTrumpPublic(frenchTarotTrumpCard(5)))
	assert.False(t, domain.FrenchTarotIsTrumpPublic(frenchTarotSuitCard(1, 5)))
	assert.True(t, domain.FrenchTarotIsExcusePublic(frenchTarotExcuseCard()))
	assert.False(t, domain.FrenchTarotIsExcusePublic(frenchTarotTrumpCard(1)))

	// bouts: petit (trump 1), 21, excuse.
	assert.True(t, domain.FrenchTarotIsBoutPublic(frenchTarotTrumpCard(1)))
	assert.True(t, domain.FrenchTarotIsBoutPublic(frenchTarotTrumpCard(21)))
	assert.True(t, domain.FrenchTarotIsBoutPublic(frenchTarotExcuseCard()))
	assert.False(t, domain.FrenchTarotIsBoutPublic(frenchTarotTrumpCard(10)))
	assert.False(t, domain.FrenchTarotIsBoutPublic(frenchTarotSuitCard(1, 14)))
}

func TestFrenchTarotHalfPoints(t *testing.T) {
	cases := []struct {
		card *domain.Card
		want int
	}{
		{frenchTarotSuitCard(1, 14), 9}, // Roi
		{frenchTarotSuitCard(1, 13), 7}, // Dame
		{frenchTarotSuitCard(1, 12), 5}, // Cavalier
		{frenchTarotSuitCard(1, 11), 3}, // Valet
		{frenchTarotSuitCard(1, 5), 1},  // pip
		{frenchTarotTrumpCard(1), 9},    // petit bout
		{frenchTarotTrumpCard(21), 9},   // 21 bout
		{frenchTarotExcuseCard(), 9},    // excuse bout
		{frenchTarotTrumpCard(10), 1},   // plain trump
		{nil, 0},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, domain.FrenchTarotCardHalfPointsPublic(c.card))
	}
}

// --- Scoring helper (pure) ---

func TestFrenchTarotTargetForBouts(t *testing.T) {
	assert.Equal(t, 56, domain.FrenchTarotTargetForBouts(0))
	assert.Equal(t, 51, domain.FrenchTarotTargetForBouts(1))
	assert.Equal(t, 41, domain.FrenchTarotTargetForBouts(2))
	assert.Equal(t, 36, domain.FrenchTarotTargetForBouts(3))
	assert.Equal(t, 36, domain.FrenchTarotTargetForBouts(3)) // clamp
}

func TestFrenchTarotBidMult(t *testing.T) {
	assert.Equal(t, 1, domain.FrenchTarotBidMultPublic(domain.FrenchTarotBidPetite))
	assert.Equal(t, 2, domain.FrenchTarotBidMultPublic(domain.FrenchTarotBidGarde))
	assert.Equal(t, 4, domain.FrenchTarotBidMultPublic(domain.FrenchTarotBidGardeSans))
	assert.Equal(t, 6, domain.FrenchTarotBidMultPublic(domain.FrenchTarotBidGardeContre))
	assert.Equal(t, 1, domain.FrenchTarotBidMultPublic(domain.FrenchTarotBidPass))
}

func TestFrenchTarotScoreDealWinZeroSum(t *testing.T) {
	// 1 bout → target 51 (102 half). declarer captured 60 pts (120 half). won.
	bd := domain.FrenchTarotScoreDeal(120, 1, 0, 1)
	assert.True(t, bd.Won)
	assert.Equal(t, 51, bd.Target)
	assert.Equal(t, 9, bd.DiffPoints) // (120-102)/2 = 9
	assert.Equal(t, (25+9)*1, bd.Base)
	assert.Equal(t, bd.Base, bd.PerDefender)
	assert.Equal(t, 3*bd.Base, bd.DeclarerScore)
	assert.Equal(t, -bd.Base, bd.DefenderScore)
	// zero-sum
	assert.Equal(t, 0, bd.DeclarerScore+3*bd.DefenderScore)
}

func TestFrenchTarotScoreDealLoss(t *testing.T) {
	// 0 bouts → target 56 (112 half). declarer 40 pts (80 half). loss.
	bd := domain.FrenchTarotScoreDeal(80, 0, 0, 2)
	assert.False(t, bd.Won)
	assert.Equal(t, 56, bd.Target)
	assert.Equal(t, 16, bd.DiffPoints) // (112-80)/2 = 16
	assert.Equal(t, (25+16)*2, bd.Base)
	assert.Equal(t, -bd.Base, bd.PerDefender) // declarer loses
	assert.Equal(t, -3*bd.Base, bd.DeclarerScore)
	assert.Equal(t, bd.Base, bd.DefenderScore)
	assert.Equal(t, 0, bd.DeclarerScore+3*bd.DefenderScore)
}

func TestFrenchTarotScoreDealPetitAuBout(t *testing.T) {
	// declarer won, petit au bout in declarer's favor.
	bd := domain.FrenchTarotScoreDeal(120, 1, 1, 2)
	base := (25 + 9) * 2
	petit := 10 * 2
	assert.Equal(t, base+petit, bd.PerDefender)
	assert.Equal(t, 3*(base+petit), bd.DeclarerScore)
	assert.Equal(t, 0, bd.DeclarerScore+3*bd.DefenderScore)

	// petit au bout against declarer.
	bd2 := domain.FrenchTarotScoreDeal(120, 1, -1, 1)
	base2 := (25 + 9)
	assert.Equal(t, base2-10, bd2.PerDefender)
}

func TestFrenchTarotScoreDealExactTarget(t *testing.T) {
	// exactly on target (declHalf == targetHalf) counts as won, diff 0.
	bd := domain.FrenchTarotScoreDeal(112, 0, 0, 1)
	assert.True(t, bd.Won)
	assert.Equal(t, 0, bd.DiffPoints)
	assert.Equal(t, 25, bd.Base)
}

func TestFrenchTarotScoreDealOddHalfRoundsUp(t *testing.T) {
	// declHalf 103 vs targetHalf 102 → diffHalf 1 → diffPoints (1+1)/2 = 1.
	bd := domain.FrenchTarotScoreDeal(103, 1, 0, 1)
	assert.True(t, bd.Won)
	assert.Equal(t, 1, bd.DiffPoints)
}

// --- Bidding ---

func TestFrenchTarotBiddingAscension(t *testing.T) {
	g := frenchTarotNewReset()
	// force human (seat 0) to be the bidder.
	g.SetBidPlayerIdx(0)
	require.True(t, g.IsHumanBidTurn())
	// A pass then bid should work; a non-ascending bid should fail.
	require.NoError(t, g.PlayerBid(domain.FrenchTarotBidGarde))
	assert.Equal(t, domain.FrenchTarotBidGarde, g.GetHighestBid())
	assert.Equal(t, 0, g.GetHighestBidder())
}

func TestFrenchTarotBidMustExceed(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetBidPlayerIdx(0)
	g.SetHighestBid(domain.FrenchTarotBidGarde)
	// Petite <= Garde → rejected.
	err := g.PlayerBid(domain.FrenchTarotBidPetite)
	require.Error(t, err)
}

func TestFrenchTarotAllPassRedeal(t *testing.T) {
	g := frenchTarotNewReset()
	dealer0 := g.GetDealerIdx()
	// Everyone passes: drive CPU passes and human pass.
	for i := 0; i < domain.FrenchTarotPlayerCnt*2; i++ {
		if g.GetPhase() != domain.FrenchTarotPhaseBid {
			break
		}
		if g.IsHumanBidTurn() {
			_ = g.PlayerPass()
		} else {
			g.CpuBid()
		}
	}
	// After the auction resolves it is either a redeal (still bid phase, dealer
	// advanced) or a contract was taken (chien/play). Both are valid; assert no panic
	// and a sane phase.
	assert.Contains(t, []domain.FrenchTarotPhase{
		domain.FrenchTarotPhaseBid, domain.FrenchTarotPhaseChien, domain.FrenchTarotPhasePlay,
	}, g.GetPhase())
	_ = dealer0
}

func TestFrenchTarotForcedAllPassRedeal(t *testing.T) {
	g := frenchTarotNewReset()
	// Drive a full all-pass auction: 3 CPUs then the human all pass. Because CPU
	// bids are hand-dependent, loop the auction until it either redeals or a
	// contract is taken, but assert that an explicit all-pass triggers a redeal by
	// checking the dealer advances when nobody bid.
	g.SetBidPlayerIdx(0)
	dealer := g.GetDealerIdx()
	// Force every seat to pass by passing as human when it is our turn and making
	// CPUs pass via CpuBid only when their hand is weak is not deterministic, so we
	// directly exercise the redeal path: set all others passed, human passes last.
	for i := 0; i < domain.FrenchTarotPlayerCnt; i++ {
		if g.GetPhase() != domain.FrenchTarotPhaseBid {
			break
		}
		if g.IsHumanBidTurn() {
			_ = g.PlayerPass()
		} else {
			// make the CPU pass deterministically by emptying its evaluation:
			// clear its hand so evalHand returns 0 → pass.
			g.GetPlayer(g.GetBidPlayerIdx()).Reset()
			g.CpuBid()
		}
	}
	// All four passed → redeal happened: still bid phase and dealer advanced.
	assert.Equal(t, domain.FrenchTarotPhaseBid, g.GetPhase())
	assert.Equal(t, (dealer+1)%domain.FrenchTarotPlayerCnt, g.GetDealerIdx())
}

// --- Chien exchange per bid ---

func TestFrenchTarotFinalizePetiteEntersChien(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetBidPlayerIdx(0)
	require.NoError(t, g.PlayerBid(domain.FrenchTarotBidPetite))
	// Drive remaining CPU passes.
	for g.GetPhase() == domain.FrenchTarotPhaseBid {
		g.CpuBid()
	}
	if g.GetDeclarerIdx() == 0 {
		// Human took Petite → chien revealed, declarer holds 24 cards.
		assert.Equal(t, domain.FrenchTarotPhaseChien, g.GetPhase())
		assert.True(t, g.GetChienRevealed())
		assert.Equal(t, domain.FrenchTarotHandSize+domain.FrenchTarotChienSize, g.GetPlayer(0).GetCardsSize())
	}
}

func TestFrenchTarotGardeSansStashToDeclarer(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetBidPlayerIdx(0)
	require.NoError(t, g.PlayerBid(domain.FrenchTarotBidGardeSans))
	for g.GetPhase() == domain.FrenchTarotPhaseBid {
		g.CpuBid()
	}
	if g.GetDeclarerIdx() == 0 {
		assert.Equal(t, domain.FrenchTarotPhasePlay, g.GetPhase())
		assert.Equal(t, 0, g.GetStashOwner()) // declarer
		assert.Equal(t, 0, g.GetChienCount())
	}
}

func TestFrenchTarotGardeContreStashToDefenders(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetBidPlayerIdx(0)
	require.NoError(t, g.PlayerBid(domain.FrenchTarotBidGardeContre))
	for g.GetPhase() == domain.FrenchTarotPhaseBid {
		g.CpuBid()
	}
	if g.GetDeclarerIdx() == 0 {
		assert.Equal(t, domain.FrenchTarotPhasePlay, g.GetPhase())
		assert.Equal(t, 1, g.GetStashOwner()) // defenders
	}
}

// --- Discard validation ---

func TestFrenchTarotDiscardValidation(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetPhase(domain.FrenchTarotPhaseChien)
	// Build a 24-card-ish hand with a king, an excuse, plenty of pips.
	frenchTarotSetHand(g, 0,
		frenchTarotSuitCard(1, 14), // king (illegal discard)
		frenchTarotExcuseCard(),    // excuse (illegal discard)
		frenchTarotSuitCard(1, 2),
		frenchTarotSuitCard(1, 3),
		frenchTarotSuitCard(1, 4),
		frenchTarotSuitCard(1, 5),
		frenchTarotSuitCard(1, 6),
		frenchTarotSuitCard(1, 7),
	)
	// Wrong count.
	require.Error(t, g.PlayerDiscard([]int{0, 1, 2}))
	// King in discard → error.
	require.Error(t, g.PlayerDiscard([]int{0, 2, 3, 4, 5, 6}))
	// Excuse in discard → error.
	require.Error(t, g.PlayerDiscard([]int{1, 2, 3, 4, 5, 6}))
	// Duplicate index → error.
	require.Error(t, g.PlayerDiscard([]int{2, 2, 3, 4, 5, 6}))
	// Legal discard of 6 pips.
	require.NoError(t, g.PlayerDiscard([]int{2, 3, 4, 5, 6, 7}))
	assert.Equal(t, domain.FrenchTarotPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetStashOwner())
}

// --- Follow / trump priority / overtrump ---

func TestFrenchTarotFollowSuit(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetDeclarerIdx(0)
	// Player 1 has hearts and a spade; hearts led → must follow hearts.
	frenchTarotSetHand(g, 1,
		frenchTarotSuitCard(domain.CardDesignHeart, 5),
		frenchTarotSuitCard(domain.CardDesignHeart, 9),
		frenchTarotSuitCard(domain.CardDesignSpade, 3),
		frenchTarotTrumpCard(4),
	)
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignHeart, 7)},
	))
	g.SetCurrentPlayerIdx(1)
	valid := g.GetPlayableIndices(1)
	assert.ElementsMatch(t, []int{0, 1}, valid) // only the two hearts
}

func TestFrenchTarotVoidMustTrump(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetDeclarerIdx(0)
	// Player 1 void of hearts, has trumps + spades → must trump.
	frenchTarotSetHand(g, 1,
		frenchTarotSuitCard(domain.CardDesignSpade, 3),
		frenchTarotTrumpCard(4),
		frenchTarotTrumpCard(9),
	)
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignHeart, 7)},
	))
	g.SetCurrentPlayerIdx(1)
	valid := g.GetPlayableIndices(1)
	assert.ElementsMatch(t, []int{1, 2}, valid) // only the trumps
}

func TestFrenchTarotOvertrumpObligation(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetDeclarerIdx(0)
	// Hearts led, a trump 10 already played. Player 2 void of hearts, trumps 5 and 15.
	// Must overtrump → only trump 15 (>10) valid.
	frenchTarotSetHand(g, 2,
		frenchTarotTrumpCard(5),
		frenchTarotTrumpCard(15),
	)
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignHeart, 7)},
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotTrumpCard(10)},
	))
	g.SetCurrentPlayerIdx(2)
	valid := g.GetPlayableIndices(2)
	assert.ElementsMatch(t, []int{1}, valid)
}

func TestFrenchTarotCannotOvertrumpPlaysAnyTrump(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetDeclarerIdx(0)
	// Trump 18 already played; player only has lower trumps → may play any trump.
	frenchTarotSetHand(g, 2,
		frenchTarotTrumpCard(5),
		frenchTarotTrumpCard(9),
	)
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignHeart, 7)},
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotTrumpCard(18)},
	))
	g.SetCurrentPlayerIdx(2)
	valid := g.GetPlayableIndices(2)
	assert.ElementsMatch(t, []int{0, 1}, valid)
}

func TestFrenchTarotExcuseAlwaysPlayable(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetDeclarerIdx(0)
	// Hearts led; player has a heart + the excuse. Both heart and excuse valid.
	frenchTarotSetHand(g, 1,
		frenchTarotSuitCard(domain.CardDesignHeart, 5),
		frenchTarotExcuseCard(),
		frenchTarotSuitCard(domain.CardDesignSpade, 3),
	)
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignHeart, 7)},
	))
	g.SetCurrentPlayerIdx(1)
	valid := g.GetPlayableIndices(1)
	assert.ElementsMatch(t, []int{0, 1}, valid) // heart + excuse (not the spade)
}

func TestFrenchTarotLeadAllValid(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	g.SetDeclarerIdx(0)
	frenchTarotSetHand(g, 0,
		frenchTarotSuitCard(domain.CardDesignHeart, 5),
		frenchTarotTrumpCard(4),
		frenchTarotExcuseCard(),
	)
	g.SetCurrentTrick(nil)
	g.SetCurrentPlayerIdx(0)
	valid := g.GetPlayableIndices(0)
	assert.ElementsMatch(t, []int{0, 1, 2}, valid)
}

// --- Trick winner ---

func TestFrenchTarotTrickWinnerHighestTrump(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignHeart, 14)},
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotTrumpCard(3)},
		&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotTrumpCard(9)},
		&domain.TrickCard{PlayerIdx: 3, Card: frenchTarotSuitCard(domain.CardDesignHeart, 2)},
	))
	assert.Equal(t, 2, g.TrickWinnerPublic())
}

func TestFrenchTarotTrickWinnerLedSuit(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignHeart, 8)},
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotSuitCard(domain.CardDesignHeart, 14)},
		&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotSuitCard(domain.CardDesignSpade, 14)},
		&domain.TrickCard{PlayerIdx: 3, Card: frenchTarotSuitCard(domain.CardDesignHeart, 3)},
	))
	assert.Equal(t, 1, g.TrickWinnerPublic()) // heart 14
}

func TestFrenchTarotExcuseNeverWinsAndLedSuit(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetContract(domain.FrenchTarotBidPetite)
	// Excuse led; the led suit becomes the next card (spade).
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotExcuseCard()},
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotSuitCard(domain.CardDesignSpade, 5)},
		&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotSuitCard(domain.CardDesignSpade, 9)},
		&domain.TrickCard{PlayerIdx: 3, Card: frenchTarotSuitCard(domain.CardDesignHeart, 14)},
	))
	assert.Equal(t, domain.CardDesignSpade, g.LedSuitPublic())
	assert.Equal(t, 2, g.TrickWinnerPublic()) // spade 9, excuse never wins
}

// --- ResolveTrick / excuse ownership ---

func TestFrenchTarotResolveTrickExcuseKeptByOwner(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetDeclarerIdx(0)
	g.SetTrickNumber(1)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.FrenchTarotPhaseTrickEnd)
	// Player 3 plays the excuse; player 2 wins with a trump.
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignSpade, 5)},
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotSuitCard(domain.CardDesignSpade, 9)},
		&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotTrumpCard(4)},
		&domain.TrickCard{PlayerIdx: 3, Card: frenchTarotExcuseCard()},
	))
	g.ResolveTrick()
	// Winner (player 2) keeps 3 cards; excuse owner (player 3) keeps the excuse.
	assert.Equal(t, 1, g.GetPlayer(2).GetTrickCount())
	assert.Equal(t, 1, g.GetPlayer(3).GetTrickCount())
	// Player 3's captured half-points include the excuse (9).
	assert.Equal(t, 9, g.GetCardPoints(3))
}

// --- Round-end scoring zero-sum & petit au bout ---

func TestFrenchTarotEnterRoundEndZeroSum(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidGarde)
	g.SetTrickNumber(domain.FrenchTarotTrickCount)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.FrenchTarotPhaseTrickEnd)
	// Give the declarer a strong last trick (contains the petit).
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotSuitCard(domain.CardDesignSpade, 5)},
		&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotSuitCard(domain.CardDesignSpade, 9)},
		&domain.TrickCard{PlayerIdx: 3, Card: frenchTarotSuitCard(domain.CardDesignSpade, 3)},
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotTrumpCard(1)}, // petit, declarer wins
	))
	g.ResolveTrick() // last trick → RoundEnd + enterRoundEnd
	assert.Equal(t, domain.FrenchTarotPhaseRoundEnd, g.GetPhase())
	scores := g.GetPlayerScores()
	// zero-sum across all four seats.
	assert.Equal(t, 0, scores[0]+scores[1]+scores[2]+scores[3])
	assert.Contains(t, []domain.FrenchTarotOutcome{
		domain.FrenchTarotOutcomeWin, domain.FrenchTarotOutcomeLoss,
	}, g.GetOutcome())
}

// --- Full game drive (smoke) ---

func TestFrenchTarotFullGameDrive(t *testing.T) {
	g := frenchTarotNewReset()
	guard := 0
	for !g.GetGameEndFlag() && guard < 5000 {
		guard++
		switch g.GetPhase() {
		case domain.FrenchTarotPhaseBid:
			if g.IsHumanBidTurn() {
				if err := g.PlayerBid(domain.FrenchTarotBidGardeContre); err != nil {
					_ = g.PlayerPass()
				}
			} else {
				g.CpuBid()
			}
		case domain.FrenchTarotPhaseChien:
			if g.IsHumanDiscardTurn() {
				// discard the first 6 legal cards.
				_ = g.PlayerDiscard(frenchTarotFirstLegalDiscards(g))
			} else {
				g.CpuDiscard()
			}
		case domain.FrenchTarotPhasePlay:
			if g.IsHumanTurn() {
				idx := g.GetPlayableIndices(0)
				require.NotEmpty(t, idx)
				_ = g.PlayerPlay(idx[0])
				if g.GetPhase() == domain.FrenchTarotPhaseTrickEnd {
					g.ResolveTrick()
				}
			} else {
				g.CpuPlay()
				if g.GetPhase() == domain.FrenchTarotPhaseTrickEnd {
					g.ResolveTrick()
				}
			}
		case domain.FrenchTarotPhaseTrickEnd:
			g.NextTrick()
		case domain.FrenchTarotPhaseRoundEnd:
			g.ScoreRound()
			g.NextRound()
		case domain.FrenchTarotPhaseGameEnd:
		}
	}
	assert.True(t, g.GetGameEndFlag())
	// -1 は同点トップ (引き分け)、0..3 は単独勝者。どちらも有効な終局。
	assert.GreaterOrEqual(t, g.GetWinnerPlayer(), -1)
	assert.Less(t, g.GetWinnerPlayer(), domain.FrenchTarotPlayerCnt)
}

// frenchTarotFirstLegalDiscards returns 6 indices of legal écart cards from the
// declarer's revealed hand (avoiding kings, trumps, and the excuse when possible).
func frenchTarotFirstLegalDiscards(g *domain.FrenchTarot) []int {
	p := g.GetPlayer(g.GetDeclarerIdx())
	legal := make([]int, 0)
	fallback := make([]int, 0)
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if domain.FrenchTarotIsExcusePublic(c) {
			continue
		}
		if domain.FrenchTarotIsTrumpPublic(c) {
			fallback = append(fallback, i)
			continue
		}
		if c.GetValue() == domain.FrenchTarotKingValue {
			continue
		}
		legal = append(legal, i)
	}
	legal = append(legal, fallback...)
	if len(legal) > domain.FrenchTarotChienSize {
		legal = legal[:domain.FrenchTarotChienSize]
	}
	return legal
}

// --- JSON round-trip & validation ---

func TestFrenchTarotJSONRoundTrip(t *testing.T) {
	g := frenchTarotNewReset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var restored domain.FrenchTarot
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetChienCount(), restored.GetChienCount())
	assert.Equal(t, g.GetDealerIdx(), restored.GetDealerIdx())
}

func TestFrenchTarotUnmarshalErrors(t *testing.T) {
	cases := []string{
		`{"ps":[]}`,                    // wrong player count
		`{"ps":[null,null,null,null]}`, // nil players
		`{"ph":99,"ps":` + frenchTarotFourPlayers() + `}`, // bad phase
	}
	for _, c := range cases {
		var g domain.FrenchTarot
		assert.Error(t, json.Unmarshal([]byte(c), &g), "input: %s", c)
	}
}

func TestFrenchTarotUnmarshalBadCard(t *testing.T) {
	// A trick card with an out-of-range trump value.
	in := `{"ps":` + frenchTarotFourPlayers() + `,"ph":0,"ct":[{"pi":0,"c":{"design":5,"value":99}}]}`
	var g domain.FrenchTarot
	assert.Error(t, json.Unmarshal([]byte(in), &g))
}

func TestFrenchTarotUnmarshalBadIndex(t *testing.T) {
	in := `{"ps":` + frenchTarotFourPlayers() + `,"ph":0,"ci":9}`
	var g domain.FrenchTarot
	assert.Error(t, json.Unmarshal([]byte(in), &g))
}

func frenchTarotFourPlayers() string {
	one := `{"gp":{"isHuman":false,"cards":[]},"th":{}}`
	return "[" + one + "," + one + "," + one + "," + one + "]"
}

// --- Config ---

func TestFrenchTarotConfigValidate(t *testing.T) {
	cfg := domain.DefaultFrenchTarotConfig()
	assert.NoError(t, cfg.Validate())
	bad := cfg
	bad.TargetDeals = 0
	assert.Error(t, bad.Validate())
	bad2 := cfg
	bad2.CpuDifficulty = domain.FrenchTarotCpuDifficulty(99)
	assert.Error(t, bad2.Validate())
}

// --- Accessors / misc ---

func TestFrenchTarotAccessors(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetRoundNumber(2)
	assert.Equal(t, 2, g.GetRoundNumber())
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	g.SetContract(domain.FrenchTarotBidGarde)
	assert.Equal(t, domain.FrenchTarotBidGarde, g.GetContract())
	assert.Nil(t, g.GetPlayer(99))
	assert.Nil(t, g.GetPlayableIndices(99))
	assert.Equal(t, 0, g.GetCardPoints(99))
	assert.NotNil(t, g.GetActionLog())
}

// #5712: エカルトの合法性判定と、CUI/Web が出す「捨てられる札」の案内は
// FrenchTarotUnburiableReason ただ一つから来る。ここは**弾く側**が各理由を
// 実際に弾くことを固定する (案内だけ直って検証が古いままになるのを防ぐ)。
func TestFrenchTarotDiscardRejectsEveryUnburiableReason(t *testing.T) {
	cases := []struct {
		name   string
		card   *domain.Card
		reason string
	}{
		{"king", frenchTarotSuitCard(1, domain.FrenchTarotKingValue), domain.FrenchTarotUnburiableKing},
		{"excuse", frenchTarotExcuseCard(), domain.FrenchTarotUnburiableExcuse},
		{"petit", frenchTarotTrumpCard(1), domain.FrenchTarotUnburiableBout},
		{"twenty-one", frenchTarotTrumpCard(domain.FrenchTarotMaxTrump), domain.FrenchTarotUnburiableBout},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// 分類と検証が同じ理由を指していること。
			assert.Equal(t, c.reason, domain.FrenchTarotUnburiableReason(c.card))

			g := frenchTarotNewReset()
			g.SetDeclarerIdx(0)
			g.SetContract(domain.FrenchTarotBidPetite)
			g.SetPhase(domain.FrenchTarotPhaseChien)
			// 捨てられる札を 7 枚 + 対象の 1 枚。切り札を許す条件には掛からない。
			frenchTarotSetHand(g, 0, c.card,
				frenchTarotSuitCard(1, 2), frenchTarotSuitCard(1, 3), frenchTarotSuitCard(1, 4),
				frenchTarotSuitCard(1, 5), frenchTarotSuitCard(1, 6), frenchTarotSuitCard(1, 7),
				frenchTarotSuitCard(2, 2),
			)

			// 対象を含む 6 枚は拒否され、含まない 6 枚は通る。
			require.Error(t, g.PlayerDiscard([]int{0, 1, 2, 3, 4, 5}))
			require.NoError(t, g.PlayerDiscard([]int{1, 2, 3, 4, 5, 6}))
		})
	}
}

// #5712: 案内 (CUI/Web) と検証 (validateDiscards) が同じ判定を使うので、その判定
// 自体をここで直接固定する。
func TestFrenchTarotUnburiableReason(t *testing.T) {
	cases := []struct {
		name string
		card *domain.Card
		want string
	}{
		{"pip is free", frenchTarotSuitCard(1, 5), ""},
		{"king", frenchTarotSuitCard(1, domain.FrenchTarotKingValue), domain.FrenchTarotUnburiableKing},
		{"excuse", frenchTarotExcuseCard(), domain.FrenchTarotUnburiableExcuse},
		{"petit", frenchTarotTrumpCard(1), domain.FrenchTarotUnburiableBout},
		{"twenty-one", frenchTarotTrumpCard(domain.FrenchTarotMaxTrump), domain.FrenchTarotUnburiableBout},
		{"plain trump", frenchTarotTrumpCard(9), domain.FrenchTarotUnburiableTrump},
		{"nil", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, domain.FrenchTarotUnburiableReason(c.card))
		})
	}
}

// 切り札は「自由に捨てられる札が 6 枚に満たないとき」だけ候補に入る (validateDiscards と同条件)。
func TestFrenchTarotBuriableIndices(t *testing.T) {
	build := func(cards ...*domain.Card) *domain.FrenchTarotPlayer {
		g := frenchTarotNewReset()
		frenchTarotSetHand(g, 0, cards...)
		return g.GetPlayer(0)
	}

	t.Run("skips kings, the Excuse and the bouts", func(t *testing.T) {
		p := build(
			frenchTarotSuitCard(1, 2), frenchTarotSuitCard(1, domain.FrenchTarotKingValue),
			frenchTarotExcuseCard(), frenchTarotTrumpCard(1), frenchTarotTrumpCard(9),
			frenchTarotSuitCard(1, 3), frenchTarotSuitCard(1, 4), frenchTarotSuitCard(1, 5),
			frenchTarotSuitCard(1, 6), frenchTarotSuitCard(1, 7),
		)

		assert.Equal(t, []int{0, 5, 6, 7, 8, 9}, domain.FrenchTarotBuriableIndices(p))
	})

	t.Run("allows trumps only when the free cards run short", func(t *testing.T) {
		p := build(
			frenchTarotSuitCard(1, 2), frenchTarotSuitCard(1, 3),
			frenchTarotTrumpCard(9), frenchTarotTrumpCard(10),
		)

		assert.Equal(t, []int{0, 1, 2, 3}, domain.FrenchTarotBuriableIndices(p))
	})

	t.Run("nil player yields nothing", func(t *testing.T) {
		assert.Empty(t, domain.FrenchTarotBuriableIndices(nil))
	})
}
