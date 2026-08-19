//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers (koenigrufen-prefixed) ---

func koenigrufenTrumpCard(v int) *domain.Card {
	return domain.NewCard(domain.KoenigrufenTrumpDesign, v, false)
}

func koenigrufenSkusCard() *domain.Card {
	return domain.NewCard(domain.KoenigrufenSkusDesign, domain.KoenigrufenSkusValue, false)
}

func koenigrufenSuitCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func koenigrufenKingCard(design int) *domain.Card {
	return domain.NewCard(design, domain.KoenigrufenKingValue, false)
}

func koenigrufenNewReset() *domain.Koenigrufen {
	g := domain.NewDefaultKoenigrufen()
	g.Reset()
	return g
}

func koenigrufenSetHand(g *domain.Koenigrufen, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func koenigrufenTrickCards(cards ...*domain.TrickCard) []*domain.TrickCard {
	return cards
}

// --- Deck ---

func TestKoenigrufenDeckIs54(t *testing.T) {
	deck := domain.BuildKoenigrufenDeckPublic()
	require.Len(t, deck, domain.KoenigrufenDeckSize)
	suits, trumps, skus := 0, 0, 0
	total := 0
	for _, c := range deck {
		total += domain.KoenigrufenCardPointsPublic(c)
		switch c.GetDesign() {
		case domain.KoenigrufenSkusDesign:
			skus++
			assert.Equal(t, domain.KoenigrufenSkusValue, c.GetValue())
		case domain.KoenigrufenTrumpDesign:
			trumps++
			assert.GreaterOrEqual(t, c.GetValue(), 1)
			assert.LessOrEqual(t, c.GetValue(), domain.KoenigrufenMaxTrump)
		default:
			suits++
		}
	}
	assert.Equal(t, 32, suits)
	assert.Equal(t, 21, trumps)
	assert.Equal(t, 1, skus)
	assert.Equal(t, domain.KoenigrufenTotalPoints, total)
}

func TestKoenigrufenDealDistribution(t *testing.T) {
	g := koenigrufenNewReset()
	assert.Equal(t, domain.KoenigrufenPhaseBid, g.GetPhase())
	assert.Equal(t, domain.KoenigrufenTalonSize, g.GetTalonCount())
	for i := 0; i < domain.KoenigrufenPlayerCnt; i++ {
		assert.Equal(t, domain.KoenigrufenHandSize, g.GetPlayer(i).GetCardsSize())
	}
}

// --- Card classification / points ---

func TestKoenigrufenClassification(t *testing.T) {
	assert.True(t, domain.KoenigrufenIsTrumpPublic(koenigrufenTrumpCard(5)))
	assert.False(t, domain.KoenigrufenIsTrumpPublic(koenigrufenSuitCard(1, 5)))
	assert.True(t, domain.KoenigrufenIsSkusPublic(koenigrufenSkusCard()))
	assert.False(t, domain.KoenigrufenIsSkusPublic(koenigrufenTrumpCard(1)))
	assert.True(t, domain.KoenigrufenIsKingPublic(koenigrufenKingCard(domain.CardDesignHeart)))
	assert.False(t, domain.KoenigrufenIsKingPublic(koenigrufenTrumpCard(8)))

	// trull: pagat (trump 1), 21, skus.
	assert.True(t, domain.KoenigrufenIsTrullPublic(koenigrufenTrumpCard(1)))
	assert.True(t, domain.KoenigrufenIsTrullPublic(koenigrufenTrumpCard(21)))
	assert.True(t, domain.KoenigrufenIsTrullPublic(koenigrufenSkusCard()))
	assert.False(t, domain.KoenigrufenIsTrullPublic(koenigrufenTrumpCard(10)))
	assert.False(t, domain.KoenigrufenIsTrullPublic(koenigrufenKingCard(1)))
}

func TestKoenigrufenCardPoints(t *testing.T) {
	cases := []struct {
		card *domain.Card
		want int
	}{
		{koenigrufenSuitCard(1, 8), 5}, // King
		{koenigrufenSuitCard(1, 7), 4}, // Queen
		{koenigrufenSuitCard(1, 6), 3}, // Cavalier
		{koenigrufenSuitCard(1, 5), 2}, // Jack
		{koenigrufenSuitCard(1, 4), 1}, // pip
		{koenigrufenTrumpCard(1), 5},   // pagat trull
		{koenigrufenTrumpCard(21), 5},  // 21 trull
		{koenigrufenSkusCard(), 5},     // skus trull
		{koenigrufenTrumpCard(10), 1},  // plain trump
		{nil, 0},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, domain.KoenigrufenCardPointsPublic(c.card))
	}
}

// --- Scoring helper (pure) ---

func TestKoenigrufenBidMult(t *testing.T) {
	assert.Equal(t, 1, domain.KoenigrufenBidMultPublic(domain.KoenigrufenBidRufer))
	assert.Equal(t, 1, domain.KoenigrufenBidMultPublic(domain.KoenigrufenBidPass))
}

func TestKoenigrufenScoreDealPartnershipWinZeroSum(t *testing.T) {
	// team captured 60 (> 53) → won. diff 7, base 17.
	bd := domain.KoenigrufenScoreDeal(60, false, 1)
	assert.True(t, bd.Won)
	assert.False(t, bd.Solo)
	assert.Equal(t, 53, bd.Threshold)
	assert.Equal(t, 7, bd.Diff)
	assert.Equal(t, 17, bd.Base)
	assert.Equal(t, 17, bd.DeclarerScore)
	assert.Equal(t, 17, bd.PartnerScore)
	assert.Equal(t, -17, bd.OpponentScore)
	// zero-sum: declarer + partner + 2 opponents.
	assert.Equal(t, 0, bd.DeclarerScore+bd.PartnerScore+2*bd.OpponentScore)
}

func TestKoenigrufenScoreDealPartnershipLoss(t *testing.T) {
	bd := domain.KoenigrufenScoreDeal(40, false, 1)
	assert.False(t, bd.Won)
	assert.Equal(t, 13, bd.Diff)
	assert.Equal(t, 23, bd.Base)
	assert.Equal(t, -23, bd.DeclarerScore)
	assert.Equal(t, -23, bd.PartnerScore)
	assert.Equal(t, 23, bd.OpponentScore)
	assert.Equal(t, 0, bd.DeclarerScore+bd.PartnerScore+2*bd.OpponentScore)
}

func TestKoenigrufenScoreDealSoloZeroSum(t *testing.T) {
	bd := domain.KoenigrufenScoreDeal(60, true, 1)
	assert.True(t, bd.Won)
	assert.True(t, bd.Solo)
	assert.Equal(t, 51, bd.DeclarerScore) // 3 × 17
	assert.Equal(t, 0, bd.PartnerScore)
	assert.Equal(t, -17, bd.OpponentScore)
	// zero-sum: declarer + 3 opponents.
	assert.Equal(t, 0, bd.DeclarerScore+3*bd.OpponentScore)
}

func TestKoenigrufenScoreDealExactHalfIsLoss(t *testing.T) {
	// exactly half (53 of 106) is NOT more than half → loss.
	bd := domain.KoenigrufenScoreDeal(53, false, 1)
	assert.False(t, bd.Won)
	assert.Equal(t, 0, bd.Diff)
	assert.Equal(t, 10, bd.Base)
}

// --- Bidding ---

func TestKoenigrufenBiddingRufer(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetBidPlayerIdx(0)
	require.True(t, g.IsHumanBidTurn())
	require.NoError(t, g.PlayerBid(domain.KoenigrufenBidRufer))
	assert.Equal(t, domain.KoenigrufenBidRufer, g.GetHighestBid())
	assert.Equal(t, 0, g.GetHighestBidder())
}

func TestKoenigrufenBidInvalid(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetBidPlayerIdx(0)
	require.Error(t, g.PlayerBid(domain.KoenigrufenBidPass))
}

func TestKoenigrufenBidMustExceed(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetBidPlayerIdx(0)
	g.SetHighestBid(domain.KoenigrufenBidRufer)
	require.Error(t, g.PlayerBid(domain.KoenigrufenBidRufer))
}

func TestKoenigrufenAllPassForcedDeclarer(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetBidPlayerIdx(0)
	dealer := g.GetDealerIdx()
	for i := 0; i < domain.KoenigrufenPlayerCnt; i++ {
		if g.GetPhase() != domain.KoenigrufenPhaseBid {
			break
		}
		if g.IsHumanBidTurn() {
			_ = g.PlayerPass()
		} else {
			// force a weak CPU hand so it passes deterministically.
			g.GetPlayer(g.GetBidPlayerIdx()).Reset()
			g.CpuBid()
		}
	}
	// All passed → dealer's left neighbour is forced declarer, no redeal.
	assert.Equal(t, (dealer+1)%domain.KoenigrufenPlayerCnt, g.GetDeclarerIdx())
	assert.Equal(t, domain.KoenigrufenBidRufer, g.GetContract())
	// phase is Call or (auto-solo) Talon.
	assert.Contains(t, []domain.KoenigrufenPhase{
		domain.KoenigrufenPhaseCall, domain.KoenigrufenPhaseTalon,
	}, g.GetPhase())
}

// --- Call a king (secret partner) ---

func TestKoenigrufenCallKingPartnerHidden(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetPhase(domain.KoenigrufenPhaseCall)
	// Declarer (0) has no heart king; opponent (2) holds it.
	koenigrufenSetHand(g, 0, koenigrufenSuitCard(domain.CardDesignSpade, 2))
	koenigrufenSetHand(g, 1, koenigrufenSuitCard(domain.CardDesignClover, 2))
	koenigrufenSetHand(g, 2, koenigrufenKingCard(domain.CardDesignHeart))
	koenigrufenSetHand(g, 3, koenigrufenSuitCard(domain.CardDesignDiamond, 2))

	require.NoError(t, g.PlayerCallKing(domain.CardDesignHeart))
	// Partner is player 2 (server-side) but NOT yet revealed.
	assert.Equal(t, 2, g.GetPartnerIdx())
	assert.False(t, g.GetPartnerRevealed())
	assert.Equal(t, domain.CardDesignHeart, g.GetCalledKing())
	// Talon exchange follows.
	assert.Equal(t, domain.KoenigrufenPhaseTalon, g.GetPhase())
}

func TestKoenigrufenCannotCallOwnKing(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.KoenigrufenPhaseCall)
	koenigrufenSetHand(g, 0, koenigrufenKingCard(domain.CardDesignHeart))
	require.Error(t, g.PlayerCallKing(domain.CardDesignHeart))
	require.Error(t, g.PlayerCallKing(9)) // out of range
}

func TestKoenigrufenCalledKingInTalonIsSolo(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetPhase(domain.KoenigrufenPhaseCall)
	// Nobody holds the heart king → partner in talon → solo (partnerIdx -1).
	for i := 0; i < domain.KoenigrufenPlayerCnt; i++ {
		koenigrufenSetHand(g, i, koenigrufenSuitCard(domain.CardDesignSpade, i+1))
	}
	require.NoError(t, g.PlayerCallKing(domain.CardDesignHeart))
	assert.Equal(t, -1, g.GetPartnerIdx())
}

func TestKoenigrufenAllKingsAutoSolo(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetBidPlayerIdx(0)
	// Give the human all four kings, then take the contract → auto solo, skip call.
	koenigrufenSetHand(g, 0,
		koenigrufenKingCard(domain.CardDesignSpade),
		koenigrufenKingCard(domain.CardDesignClover),
		koenigrufenKingCard(domain.CardDesignHeart),
		koenigrufenKingCard(domain.CardDesignDiamond),
	)
	require.NoError(t, g.PlayerBid(domain.KoenigrufenBidRufer))
	for g.GetPhase() == domain.KoenigrufenPhaseBid {
		g.CpuBid()
	}
	if g.GetDeclarerIdx() == 0 {
		assert.Equal(t, -1, g.GetPartnerIdx())
		assert.Equal(t, domain.KoenigrufenPhaseTalon, g.GetPhase())
	}
}

func TestKoenigrufenCalledKingPlayRevealsPartner(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetDeclarerIdx(0)
	g.SetPartnerIdx(2)
	g.SetCalledKing(domain.CardDesignHeart)
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetPhase(domain.KoenigrufenPhasePlay)
	g.SetCurrentPlayerIdx(2)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(1)
	// Player 2 leads the called king → partner revealed.
	koenigrufenSetHand(g, 2, koenigrufenKingCard(domain.CardDesignHeart))
	require.False(t, g.GetPartnerRevealed())
	// Cast to drive a CPU play (player 2 is CPU).
	g.CpuPlay()
	assert.True(t, g.GetPartnerRevealed())
}

// --- Discard validation ---

func TestKoenigrufenDiscardValidation(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetPhase(domain.KoenigrufenPhaseTalon)
	koenigrufenSetHand(g, 0,
		koenigrufenKingCard(1), // king (illegal)
		koenigrufenSkusCard(),  // skus / trull (illegal)
		koenigrufenSuitCard(1, 1),
		koenigrufenSuitCard(1, 2),
		koenigrufenSuitCard(1, 3),
		koenigrufenSuitCard(1, 4),
		koenigrufenSuitCard(2, 1),
		koenigrufenSuitCard(2, 2),
	)
	require.Error(t, g.PlayerDiscard([]int{0, 1, 2}))          // wrong count
	require.Error(t, g.PlayerDiscard([]int{0, 2, 3, 4, 5, 6})) // king in discard
	require.Error(t, g.PlayerDiscard([]int{1, 2, 3, 4, 5, 6})) // skus in discard
	require.Error(t, g.PlayerDiscard([]int{2, 2, 3, 4, 5, 6})) // duplicate
	require.NoError(t, g.PlayerDiscard([]int{2, 3, 4, 5, 6, 7}))
	assert.Equal(t, domain.KoenigrufenPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetStashOwner())
}

// --- Follow / trump priority / overtrump / Sküs ---

func TestKoenigrufenFollowSuit(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetPhase(domain.KoenigrufenPhasePlay)
	g.SetDeclarerIdx(0)
	koenigrufenSetHand(g, 1,
		koenigrufenSuitCard(domain.CardDesignHeart, 5),
		koenigrufenSuitCard(domain.CardDesignHeart, 3),
		koenigrufenSuitCard(domain.CardDesignSpade, 2),
		koenigrufenTrumpCard(4),
	)
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenSuitCard(domain.CardDesignHeart, 7)},
	))
	g.SetCurrentPlayerIdx(1)
	assert.ElementsMatch(t, []int{0, 1}, g.GetPlayableIndices(1))
}

func TestKoenigrufenVoidMustTrump(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetPhase(domain.KoenigrufenPhasePlay)
	g.SetDeclarerIdx(0)
	// void of hearts; has trumps + spade + skus (skus is trump-like).
	koenigrufenSetHand(g, 1,
		koenigrufenSuitCard(domain.CardDesignSpade, 2),
		koenigrufenTrumpCard(4),
		koenigrufenSkusCard(),
	)
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenSuitCard(domain.CardDesignHeart, 7)},
	))
	g.SetCurrentPlayerIdx(1)
	assert.ElementsMatch(t, []int{1, 2}, g.GetPlayableIndices(1)) // trump + skus
}

func TestKoenigrufenOvertrumpObligation(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetPhase(domain.KoenigrufenPhasePlay)
	g.SetDeclarerIdx(0)
	// hearts led, trump 10 played; void, trumps 5 & 15 → must overtrump (only 15).
	koenigrufenSetHand(g, 2,
		koenigrufenTrumpCard(5),
		koenigrufenTrumpCard(15),
	)
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenSuitCard(domain.CardDesignHeart, 7)},
		&domain.TrickCard{PlayerIdx: 1, Card: koenigrufenTrumpCard(10)},
	))
	g.SetCurrentPlayerIdx(2)
	assert.ElementsMatch(t, []int{1}, g.GetPlayableIndices(2))
}

func TestKoenigrufenSkusOvertrumpsAll(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetPhase(domain.KoenigrufenPhasePlay)
	g.SetDeclarerIdx(0)
	// trump 21 played (the highest numbered trump); player has skus + trump 5.
	// skus (rank 22) can overtrump 21 → both would satisfy? skus>21, trump5<21.
	koenigrufenSetHand(g, 2,
		koenigrufenTrumpCard(5),
		koenigrufenSkusCard(),
	)
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenSuitCard(domain.CardDesignHeart, 7)},
		&domain.TrickCard{PlayerIdx: 1, Card: koenigrufenTrumpCard(21)},
	))
	g.SetCurrentPlayerIdx(2)
	// Only the skus can overtrump the 21.
	assert.ElementsMatch(t, []int{1}, g.GetPlayableIndices(2))
}

func TestKoenigrufenLeadAllValid(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetPhase(domain.KoenigrufenPhasePlay)
	g.SetDeclarerIdx(0)
	koenigrufenSetHand(g, 0,
		koenigrufenSuitCard(domain.CardDesignHeart, 5),
		koenigrufenTrumpCard(4),
		koenigrufenSkusCard(),
	)
	g.SetCurrentTrick(nil)
	g.SetCurrentPlayerIdx(0)
	assert.ElementsMatch(t, []int{0, 1, 2}, g.GetPlayableIndices(0))
}

// --- Trick winner ---

func TestKoenigrufenTrickWinnerHighestTrump(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenSuitCard(domain.CardDesignHeart, 8)},
		&domain.TrickCard{PlayerIdx: 1, Card: koenigrufenTrumpCard(3)},
		&domain.TrickCard{PlayerIdx: 2, Card: koenigrufenTrumpCard(9)},
		&domain.TrickCard{PlayerIdx: 3, Card: koenigrufenSuitCard(domain.CardDesignHeart, 2)},
	))
	assert.Equal(t, 2, g.TrickWinnerPublic())
}

func TestKoenigrufenSkusWinsAsTopTrump(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenTrumpCard(21)},
		&domain.TrickCard{PlayerIdx: 1, Card: koenigrufenSkusCard()},
		&domain.TrickCard{PlayerIdx: 2, Card: koenigrufenTrumpCard(20)},
		&domain.TrickCard{PlayerIdx: 3, Card: koenigrufenKingCard(domain.CardDesignHeart)},
	))
	assert.Equal(t, 1, g.TrickWinnerPublic()) // Sküs is the top trump
}

func TestKoenigrufenTrickWinnerLedSuit(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenSuitCard(domain.CardDesignHeart, 6)},
		&domain.TrickCard{PlayerIdx: 1, Card: koenigrufenSuitCard(domain.CardDesignHeart, 8)},
		&domain.TrickCard{PlayerIdx: 2, Card: koenigrufenSuitCard(domain.CardDesignSpade, 8)},
		&domain.TrickCard{PlayerIdx: 3, Card: koenigrufenSuitCard(domain.CardDesignHeart, 3)},
	))
	assert.Equal(t, domain.CardDesignHeart, g.LedSuitPublic())
	assert.Equal(t, 1, g.TrickWinnerPublic())
}

func TestKoenigrufenSkusLedIsTrumpSuit(t *testing.T) {
	g := koenigrufenNewReset()
	// Sküs leads → led suit is trump; a trump wins.
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenSkusCard()},
		&domain.TrickCard{PlayerIdx: 1, Card: koenigrufenSuitCard(domain.CardDesignSpade, 8)},
	))
	assert.Equal(t, domain.KoenigrufenTrumpDesign, g.LedSuitPublic())
	assert.Equal(t, 0, g.TrickWinnerPublic()) // skus is highest trump
}

// --- ResolveTrick / round-end scoring ---

func TestKoenigrufenResolveTrickCapturesAll(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetDeclarerIdx(0)
	g.SetTrickNumber(1)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.KoenigrufenPhaseTrickEnd)
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenSuitCard(domain.CardDesignSpade, 2)},
		&domain.TrickCard{PlayerIdx: 1, Card: koenigrufenSuitCard(domain.CardDesignSpade, 3)},
		&domain.TrickCard{PlayerIdx: 2, Card: koenigrufenTrumpCard(4)},
		&domain.TrickCard{PlayerIdx: 3, Card: koenigrufenSkusCard()},
	))
	g.ResolveTrick()
	// Sküs (top trump) wins for player 3; the whole trick (4 cards) goes to player 3.
	assert.Equal(t, 1, g.GetPlayer(3).GetTrickCount())
	// captured points: spade2(1)+spade3(1)+trump4(1)+skus(5) = 8.
	assert.Equal(t, 8, g.GetCardPoints(3))
}

func TestKoenigrufenEnterRoundEndZeroSum(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetDeclarerIdx(0)
	g.SetPartnerIdx(2)
	g.SetContract(domain.KoenigrufenBidRufer)
	g.SetTrickNumber(domain.KoenigrufenTrickCount)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.KoenigrufenPhaseTrickEnd)
	g.SetCurrentTrick(koenigrufenTrickCards(
		&domain.TrickCard{PlayerIdx: 1, Card: koenigrufenSuitCard(domain.CardDesignSpade, 2)},
		&domain.TrickCard{PlayerIdx: 2, Card: koenigrufenSuitCard(domain.CardDesignSpade, 3)},
		&domain.TrickCard{PlayerIdx: 3, Card: koenigrufenSuitCard(domain.CardDesignSpade, 4)},
		&domain.TrickCard{PlayerIdx: 0, Card: koenigrufenTrumpCard(1)},
	))
	g.ResolveTrick()
	assert.Equal(t, domain.KoenigrufenPhaseRoundEnd, g.GetPhase())
	assert.True(t, g.GetPartnerRevealed())
	scores := g.GetPlayerScores()
	assert.Equal(t, 0, scores[0]+scores[1]+scores[2]+scores[3])
	assert.Contains(t, []domain.KoenigrufenOutcome{
		domain.KoenigrufenOutcomeWin, domain.KoenigrufenOutcomeLoss,
	}, g.GetOutcome())
}

// --- Full game drive (smoke) ---

func TestKoenigrufenFullGameDrive(t *testing.T) {
	g := koenigrufenNewReset()
	guard := 0
	for !g.GetGameEndFlag() && guard < 5000 {
		guard++
		switch g.GetPhase() {
		case domain.KoenigrufenPhaseBid:
			if g.IsHumanBidTurn() {
				if err := g.PlayerBid(domain.KoenigrufenBidRufer); err != nil {
					_ = g.PlayerPass()
				}
			} else {
				g.CpuBid()
			}
		case domain.KoenigrufenPhaseCall:
			if g.IsHumanCallTurn() {
				_ = g.PlayerCallKing(koenigrufenFirstCallableSuit(g))
			} else {
				g.CpuCallKing()
			}
		case domain.KoenigrufenPhaseTalon:
			if g.IsHumanDiscardTurn() {
				_ = g.PlayerDiscard(koenigrufenFirstLegalDiscards(g))
			} else {
				g.CpuDiscard()
			}
		case domain.KoenigrufenPhasePlay:
			if g.IsHumanTurn() {
				idx := g.GetPlayableIndices(0)
				require.NotEmpty(t, idx)
				_ = g.PlayerPlay(idx[0])
				if g.GetPhase() == domain.KoenigrufenPhaseTrickEnd {
					g.ResolveTrick()
				}
			} else {
				g.CpuPlay()
				if g.GetPhase() == domain.KoenigrufenPhaseTrickEnd {
					g.ResolveTrick()
				}
			}
		case domain.KoenigrufenPhaseTrickEnd:
			g.NextTrick()
		case domain.KoenigrufenPhaseRoundEnd:
			g.ScoreRound()
			g.NextRound()
		case domain.KoenigrufenPhaseGameEnd:
		}
	}
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetWinnerPlayer(), -1)
	assert.Less(t, g.GetWinnerPlayer(), domain.KoenigrufenPlayerCnt)
}

func koenigrufenFirstCallableSuit(g *domain.Koenigrufen) int {
	p := g.GetPlayer(g.GetDeclarerIdx())
	held := map[int]bool{}
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if domain.KoenigrufenIsKingPublic(c) {
			held[c.GetDesign()] = true
		}
	}
	for suit := 1; suit <= domain.KoenigrufenSuitCnt; suit++ {
		if !held[suit] {
			return suit
		}
	}
	return 1
}

func koenigrufenFirstLegalDiscards(g *domain.Koenigrufen) []int {
	p := g.GetPlayer(g.GetDeclarerIdx())
	legal := make([]int, 0)
	fallback := make([]int, 0)
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if domain.KoenigrufenIsTrullPublic(c) {
			continue
		}
		if domain.KoenigrufenIsTrumpPublic(c) || domain.KoenigrufenIsSkusPublic(c) {
			fallback = append(fallback, i)
			continue
		}
		if domain.KoenigrufenIsKingPublic(c) {
			continue
		}
		legal = append(legal, i)
	}
	legal = append(legal, fallback...)
	if len(legal) > domain.KoenigrufenTalonSize {
		legal = legal[:domain.KoenigrufenTalonSize]
	}
	return legal
}

// --- JSON round-trip & validation ---

func TestKoenigrufenJSONRoundTrip(t *testing.T) {
	g := koenigrufenNewReset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var restored domain.Koenigrufen
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetTalonCount(), restored.GetTalonCount())
	assert.Equal(t, g.GetDealerIdx(), restored.GetDealerIdx())
}

func TestKoenigrufenUnmarshalErrors(t *testing.T) {
	cases := []string{
		`{"ps":[]}`,                    // wrong player count
		`{"ps":[null,null,null,null]}`, // nil players
		`{"ph":99,"ps":` + koenigrufenFourPlayers() + `}`,       // bad phase
		`{"ck":9,"ps":` + koenigrufenFourPlayers() + `,"ph":0}`, // bad called king
		`{"pn":9,"ps":` + koenigrufenFourPlayers() + `,"ph":0}`, // bad partner index
	}
	for _, c := range cases {
		var g domain.Koenigrufen
		assert.Error(t, json.Unmarshal([]byte(c), &g), "input: %s", c)
	}
}

func TestKoenigrufenUnmarshalBadCard(t *testing.T) {
	in := `{"ps":` + koenigrufenFourPlayers() + `,"ph":0,"ct":[{"pi":0,"c":{"design":5,"value":99}}]}`
	var g domain.Koenigrufen
	assert.Error(t, json.Unmarshal([]byte(in), &g))
}

func TestKoenigrufenUnmarshalBadIndex(t *testing.T) {
	in := `{"ps":` + koenigrufenFourPlayers() + `,"ph":0,"ci":9}`
	var g domain.Koenigrufen
	assert.Error(t, json.Unmarshal([]byte(in), &g))
}

func koenigrufenFourPlayers() string {
	one := `{"gp":{"isHuman":false,"cards":[]},"th":{}}`
	return "[" + one + "," + one + "," + one + "," + one + "]"
}

// --- Config ---

func TestKoenigrufenConfigValidate(t *testing.T) {
	cfg := domain.DefaultKoenigrufenConfig()
	assert.NoError(t, cfg.Validate())
	bad := cfg
	bad.TargetDeals = 0
	assert.Error(t, bad.Validate())
	bad2 := cfg
	bad2.CpuDifficulty = domain.KoenigrufenCpuDifficulty(99)
	assert.Error(t, bad2.Validate())
}

// --- Accessors / misc ---

func TestKoenigrufenAccessors(t *testing.T) {
	g := koenigrufenNewReset()
	g.SetRoundNumber(2)
	assert.Equal(t, 2, g.GetRoundNumber())
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	g.SetContract(domain.KoenigrufenBidRufer)
	assert.Equal(t, domain.KoenigrufenBidRufer, g.GetContract())
	g.SetCalledKing(3)
	assert.Equal(t, 3, g.GetCalledKing())
	assert.Nil(t, g.GetPlayer(99))
	assert.Nil(t, g.GetPlayableIndices(99))
	assert.Equal(t, 0, g.GetCardPoints(99))
	assert.NotNil(t, g.GetActionLog())
}

// --- Hint ---

func TestKoenigrufenHintPhases(t *testing.T) {
	// bid hint.
	g := koenigrufenNewReset()
	g.SetBidPlayerIdx(0)
	assert.NotNil(t, g.GetHint())

	// call hint.
	g2 := koenigrufenNewReset()
	g2.SetDeclarerIdx(0)
	g2.SetPhase(domain.KoenigrufenPhaseCall)
	h := g2.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "call_king", h.Reason)
	assert.NotNil(t, h.CallSuit)

	// discard hint.
	g3 := koenigrufenNewReset()
	g3.SetDeclarerIdx(0)
	g3.SetPhase(domain.KoenigrufenPhaseTalon)
	assert.NotNil(t, g3.GetHint())

	// play hint.
	g4 := koenigrufenNewReset()
	g4.SetDeclarerIdx(0)
	g4.SetPhase(domain.KoenigrufenPhasePlay)
	g4.SetCurrentPlayerIdx(0)
	assert.NotNil(t, g4.GetHint())
}

// #5713: 呼びスートの王を持っているかは、その本人に伝えてよい情報 (自分の手札と
// 公開済みの呼びスートだけから分かる)。宣言者は自分が持つ王を呼べないので常に false。
func TestKoenigrufenHoldsCalledKing(t *testing.T) {
	king := func(suit int) *domain.Card { return domain.NewCard(suit, domain.KoenigrufenKingValue, false) }

	g := koenigrufenNewReset()
	g.SetDeclarerIdx(0)
	g.SetCalledKing(domain.CardDesignHeart)
	koenigrufenSetHand(g, 0, king(domain.CardDesignHeart))                     // 宣言者 (呼べないはずの状態)
	koenigrufenSetHand(g, 1, king(domain.CardDesignHeart))                     // 呼ばれた王の持ち主
	koenigrufenSetHand(g, 2, king(domain.CardDesignSpade))                     // 別スートの王
	koenigrufenSetHand(g, 3, domain.NewCard(domain.CardDesignHeart, 1, false)) // 王ではない

	assert.False(t, g.KoenigrufenHoldsCalledKing(0), "the declarer is never told")
	assert.True(t, g.KoenigrufenHoldsCalledKing(1))
	assert.False(t, g.KoenigrufenHoldsCalledKing(2), "another suit's King does not count")
	assert.False(t, g.KoenigrufenHoldsCalledKing(3))

	// 範囲外の添字は false (呼び出し側が席を取り違えても panic しない)。
	assert.False(t, g.KoenigrufenHoldsCalledKing(-1))
	assert.False(t, g.KoenigrufenHoldsCalledKing(domain.KoenigrufenPlayerCnt))

	// 王を呼んでいないラウンド (単独) では誰も該当しない。
	g.SetCalledKing(-1)
	assert.False(t, g.KoenigrufenHoldsCalledKing(1))
}
