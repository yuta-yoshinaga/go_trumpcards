//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- CPU bid selection via GetHint (evalHand + difficulty tiers) ---

func frenchTarotBidHintGame(t *testing.T, diff domain.FrenchTarotCpuDifficulty, cards ...*domain.Card) *domain.FrenchTarot {
	t.Helper()
	g := frenchTarotNewReset()
	cfg := domain.DefaultFrenchTarotConfig()
	cfg.CpuDifficulty = diff
	g.SetConfig(cfg)
	g.SetBidPlayerIdx(0)
	frenchTarotSetHand(g, 0, cards...)
	return g
}

func TestFrenchTarotHintBidTiers(t *testing.T) {
	k := func(d int) *domain.Card { return frenchTarotSuitCard(d, 14) } // king
	q := func(d int) *domain.Card { return frenchTarotSuitCard(d, 13) } // dame

	// Pass tier: weak hand -> bid_pass.
	weak := frenchTarotBidHintGame(t, domain.FrenchTarotCpuDifficultyNormal,
		frenchTarotSuitCard(domain.CardDesignHeart, 2),
		frenchTarotSuitCard(domain.CardDesignHeart, 3),
		frenchTarotSuitCard(domain.CardDesignHeart, 4))
	h := weak.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "bid_pass", h.Reason)
	assert.Nil(t, h.Bid)

	// Petite tier (~24): 4 kings + 2 queens + 4 low trumps.
	petite := frenchTarotBidHintGame(t, domain.FrenchTarotCpuDifficultyNormal,
		k(domain.CardDesignSpade), k(domain.CardDesignClover), k(domain.CardDesignHeart), k(domain.CardDesignDiamond),
		q(domain.CardDesignSpade), q(domain.CardDesignClover),
		frenchTarotTrumpCard(5), frenchTarotTrumpCard(6), frenchTarotTrumpCard(7), frenchTarotTrumpCard(8))
	if h := petite.GetHint(); assert.NotNil(t, h) {
		assert.NotNil(t, h.Bid)
	}

	// Garde tier (~30): Petite hand + a bout.
	garde := frenchTarotBidHintGame(t, domain.FrenchTarotCpuDifficultyNormal,
		k(domain.CardDesignSpade), k(domain.CardDesignClover), k(domain.CardDesignHeart), k(domain.CardDesignDiamond),
		q(domain.CardDesignSpade), q(domain.CardDesignClover),
		frenchTarotTrumpCard(5), frenchTarotTrumpCard(6), frenchTarotTrumpCard(7), frenchTarotTrumpCard(8),
		frenchTarotTrumpCard(21))
	assert.NotNil(t, garde.GetHint())

	// Garde Sans tier (~31): 4 kings + 2 queens + 3 low trumps + 1 bout + 1 high trump.
	gardeSans := frenchTarotBidHintGame(t, domain.FrenchTarotCpuDifficultyNormal,
		k(domain.CardDesignSpade), k(domain.CardDesignClover), k(domain.CardDesignHeart), k(domain.CardDesignDiamond),
		q(domain.CardDesignSpade), q(domain.CardDesignClover),
		frenchTarotTrumpCard(5), frenchTarotTrumpCard(6), frenchTarotTrumpCard(7),
		frenchTarotTrumpCard(21), frenchTarotTrumpCard(15))
	assert.NotNil(t, gardeSans.GetHint())

	// Garde Contre tier (~48): 4 kings + 3 bouts + 6 high trumps.
	strong := frenchTarotBidHintGame(t, domain.FrenchTarotCpuDifficultyNormal,
		k(domain.CardDesignSpade), k(domain.CardDesignClover), k(domain.CardDesignHeart), k(domain.CardDesignDiamond),
		frenchTarotTrumpCard(1), frenchTarotTrumpCard(21), frenchTarotExcuseCard(),
		frenchTarotTrumpCard(15), frenchTarotTrumpCard(16), frenchTarotTrumpCard(17),
		frenchTarotTrumpCard(18), frenchTarotTrumpCard(19), frenchTarotTrumpCard(20))
	if h := strong.GetHint(); assert.NotNil(t, h) {
		assert.NotNil(t, h.Bid)
	}

	// Easy difficulty (higher threshold) still bids on a strong-ish hand.
	easy := frenchTarotBidHintGame(t, domain.FrenchTarotCpuDifficultyEasy,
		k(domain.CardDesignSpade), k(domain.CardDesignClover), k(domain.CardDesignHeart), k(domain.CardDesignDiamond),
		q(domain.CardDesignSpade), q(domain.CardDesignClover),
		frenchTarotTrumpCard(5), frenchTarotTrumpCard(6), frenchTarotTrumpCard(7),
		frenchTarotTrumpCard(21), frenchTarotTrumpCard(15))
	assert.NotNil(t, easy.GetHint())

	// Hard difficulty (lower threshold) bids on a medium hand.
	hard := frenchTarotBidHintGame(t, domain.FrenchTarotCpuDifficultyHard,
		k(domain.CardDesignSpade), k(domain.CardDesignClover), k(domain.CardDesignHeart), k(domain.CardDesignDiamond),
		q(domain.CardDesignSpade), q(domain.CardDesignClover),
		frenchTarotTrumpCard(5), frenchTarotTrumpCard(6), frenchTarotTrumpCard(7), frenchTarotTrumpCard(8))
	assert.NotNil(t, hard.GetHint())
}

func TestFrenchTarotHintBidWantBelowNext(t *testing.T) {
	// Even a strong hand cannot exceed a Garde Contre already on the table.
	g := frenchTarotBidHintGame(t, domain.FrenchTarotCpuDifficultyNormal,
		frenchTarotSuitCard(domain.CardDesignSpade, 14), frenchTarotSuitCard(domain.CardDesignClover, 14),
		frenchTarotTrumpCard(1), frenchTarotTrumpCard(21), frenchTarotExcuseCard(),
		frenchTarotTrumpCard(15), frenchTarotTrumpCard(16), frenchTarotTrumpCard(17))
	g.SetHighestBid(domain.FrenchTarotBidGardeContre)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "bid_pass", h.Reason)
}

func TestFrenchTarotHintBidNotHumanSeat(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetBidPlayerIdx(1) // CPU seat -> hint is nil for the human.
	assert.Nil(t, g.GetHint())
}

// --- CPU discard selection via GetHint ---

func TestFrenchTarotHintDiscardNormal(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetPhase(domain.FrenchTarotPhaseChien)
	frenchTarotSetHand(g, 0,
		frenchTarotSuitCard(domain.CardDesignHeart, 2), frenchTarotSuitCard(domain.CardDesignHeart, 3),
		frenchTarotSuitCard(domain.CardDesignHeart, 4), frenchTarotSuitCard(domain.CardDesignHeart, 5),
		frenchTarotSuitCard(domain.CardDesignHeart, 6), frenchTarotSuitCard(domain.CardDesignHeart, 7),
		frenchTarotSuitCard(domain.CardDesignSpade, 8), frenchTarotSuitCard(domain.CardDesignSpade, 9))
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "discard_weak", h.Reason)
	assert.Len(t, h.CardIndices, domain.FrenchTarotChienSize)
}

func TestFrenchTarotHintDiscardTrumpFallback(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetPhase(domain.FrenchTarotPhaseChien)
	// Nothing discardable except trumps -> the fallback path fills the écart with trumps.
	frenchTarotSetHand(g, 0,
		frenchTarotSuitCard(domain.CardDesignSpade, 14), frenchTarotSuitCard(domain.CardDesignHeart, 14),
		frenchTarotExcuseCard(),
		frenchTarotTrumpCard(5), frenchTarotTrumpCard(6), frenchTarotTrumpCard(7),
		frenchTarotTrumpCard(8), frenchTarotTrumpCard(9), frenchTarotTrumpCard(10))
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Len(t, h.CardIndices, domain.FrenchTarotChienSize)
}

// --- CPU play selection via GetHint (cpuPlaySmart branches + playHintReason) ---

func frenchTarotPlayHint(t *testing.T, declarer, current int, trick []*domain.TrickCard, hand ...*domain.Card) *domain.FrenchTarotHint {
	t.Helper()
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetDeclarerIdx(declarer)
	frenchTarotSetHand(g, 0, hand...)
	g.SetCurrentTrick(trick)
	g.SetCurrentPlayerIdx(current)
	h := g.GetHint()
	require.NotNil(t, h)
	require.Len(t, h.CardIndices, 1)
	return h
}

func TestFrenchTarotHintPlayLeadHigh(t *testing.T) {
	h := frenchTarotPlayHint(t, 0, 0, nil,
		frenchTarotSuitCard(domain.CardDesignHeart, 5), frenchTarotTrumpCard(10),
		frenchTarotSuitCard(domain.CardDesignSpade, 3))
	assert.Equal(t, "lead_high", h.Reason)
}

func TestFrenchTarotHintPlayLeadLow(t *testing.T) {
	h := frenchTarotPlayHint(t, 1, 0, nil,
		frenchTarotSuitCard(domain.CardDesignHeart, 5), frenchTarotTrumpCard(10),
		frenchTarotSuitCard(domain.CardDesignSpade, 3))
	assert.Equal(t, "lead_low", h.Reason)
}

func TestFrenchTarotHintPlayFollowWin(t *testing.T) {
	// Declarer (seat 1) is winning; the human defender can overtake with a higher heart.
	trick := frenchTarotTrickCards(&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotSuitCard(domain.CardDesignHeart, 7)})
	h := frenchTarotPlayHint(t, 1, 0, trick,
		frenchTarotSuitCard(domain.CardDesignHeart, 10), frenchTarotSuitCard(domain.CardDesignHeart, 5))
	assert.Equal(t, "follow_win", h.Reason)
}

func TestFrenchTarotHintPlayFollowDuck(t *testing.T) {
	// Declarer (seat 1) overtrumps and cannot be beaten; the defender ducks low.
	trick := frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotSuitCard(domain.CardDesignHeart, 7)},
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotTrumpCard(21)},
	)
	h := frenchTarotPlayHint(t, 1, 0, trick,
		frenchTarotSuitCard(domain.CardDesignHeart, 5), frenchTarotSuitCard(domain.CardDesignHeart, 9))
	assert.Equal(t, "follow_duck", h.Reason)
}

func TestFrenchTarotHintPlaySameSideWinning(t *testing.T) {
	// A fellow defender (seat 2) is already winning; the human defender feeds points.
	trick := frenchTarotTrickCards(&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotSuitCard(domain.CardDesignHeart, 14)})
	h := frenchTarotPlayHint(t, 1, 0, trick,
		frenchTarotSuitCard(domain.CardDesignHeart, 2), frenchTarotSuitCard(domain.CardDesignHeart, 5))
	assert.NotEmpty(t, h.Reason)
}

func TestFrenchTarotHintPlayExcuse(t *testing.T) {
	// The human is void of the led suit and holds only the Excuse.
	trick := frenchTarotTrickCards(&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotSuitCard(domain.CardDesignHeart, 7)})
	h := frenchTarotPlayHint(t, 1, 0, trick, frenchTarotExcuseCard())
	assert.Equal(t, "play_excuse", h.Reason)
}

// --- Discard validation: petit/21 rejection + allowTrump fallback ---

func TestFrenchTarotDiscardRejectsBoutTrump(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetPhase(domain.FrenchTarotPhaseChien)
	frenchTarotSetHand(g, 0,
		frenchTarotTrumpCard(1), // petit (bout) at index 0
		frenchTarotSuitCard(domain.CardDesignHeart, 2), frenchTarotSuitCard(domain.CardDesignHeart, 3),
		frenchTarotSuitCard(domain.CardDesignHeart, 4), frenchTarotSuitCard(domain.CardDesignHeart, 5),
		frenchTarotSuitCard(domain.CardDesignHeart, 6), frenchTarotSuitCard(domain.CardDesignHeart, 7))
	require.Error(t, g.PlayerDiscard([]int{0, 1, 2, 3, 4, 5}))
}

func TestFrenchTarotDiscardAllowsTrumpWhenForced(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetPhase(domain.FrenchTarotPhaseChien)
	// Only one discardable pip; the rest are (non-bout) trumps -> trumps become legal.
	frenchTarotSetHand(g, 0,
		frenchTarotSuitCard(domain.CardDesignHeart, 2),
		frenchTarotTrumpCard(5), frenchTarotTrumpCard(6), frenchTarotTrumpCard(7),
		frenchTarotTrumpCard(8), frenchTarotTrumpCard(9), frenchTarotTrumpCard(10),
		frenchTarotSuitCard(domain.CardDesignSpade, 14))
	require.NoError(t, g.PlayerDiscard([]int{1, 2, 3, 4, 5, 6}))
	assert.Equal(t, domain.FrenchTarotPhasePlay, g.GetPhase())
}

// --- Petit-au-bout sign via ResolveTrick (defender wins / no petit) ---

func TestFrenchTarotPetitAuBoutAgainstDeclarer(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidGarde)
	g.SetTrickNumber(domain.FrenchTarotTrickCount)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.FrenchTarotPhaseTrickEnd)
	// Last trick contains the petit but a defender (seat 1) wins it.
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotTrumpCard(1)},  // petit, declarer
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotTrumpCard(21)}, // defender wins
		&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotSuitCard(domain.CardDesignSpade, 3)},
		&domain.TrickCard{PlayerIdx: 3, Card: frenchTarotSuitCard(domain.CardDesignSpade, 4)},
	))
	g.ResolveTrick()
	assert.Equal(t, domain.FrenchTarotPhaseRoundEnd, g.GetPhase())
	scores := g.GetPlayerScores()
	assert.Equal(t, 0, scores[0]+scores[1]+scores[2]+scores[3])
}

func TestFrenchTarotPetitAuBoutNone(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.FrenchTarotBidGarde)
	g.SetTrickNumber(domain.FrenchTarotTrickCount)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.FrenchTarotPhaseTrickEnd)
	// Last trick without the petit -> petitAuBoutSign is 0.
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotSuitCard(domain.CardDesignSpade, 5)},
		&domain.TrickCard{PlayerIdx: 1, Card: frenchTarotSuitCard(domain.CardDesignSpade, 9)},
		&domain.TrickCard{PlayerIdx: 2, Card: frenchTarotSuitCard(domain.CardDesignSpade, 3)},
		&domain.TrickCard{PlayerIdx: 3, Card: frenchTarotSuitCard(domain.CardDesignSpade, 4)},
	))
	g.ResolveTrick()
	bd := g.ComputeBreakdownPublic()
	assert.Equal(t, 0, bd.PetitDelta)
}

// --- checkGameEnd: solo human win, solo CPU win (human loses), and a draw ---

// frenchTarotForceGameEnd builds a game where declarer captures nothing (a heavy
// contract loss) at the final deal, with pre-loaded scores so the post-deal totals
// resolve to a specific leadership outcome, then triggers scoring.
func frenchTarotForceGameEnd(declarer int, preset [domain.FrenchTarotPlayerCnt]int) *domain.FrenchTarot {
	g := frenchTarotNewReset()
	cfg := domain.DefaultFrenchTarotConfig()
	cfg.TargetDeals = 1
	g.SetConfig(cfg)
	g.SetRoundNumber(1)
	g.SetDeclarerIdx(declarer)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetPlayerScores(preset)
	g.SetPhase(domain.FrenchTarotPhaseRoundEnd)
	g.ScoreRound()
	return g
}

func TestFrenchTarotGameEndHumanWin(t *testing.T) {
	// declarer is a CPU (seat 1) and loses; the human (seat 0) ends sole leader.
	g := frenchTarotForceGameEnd(1, [domain.FrenchTarotPlayerCnt]int{300, 0, 0, 0})
	require.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerPlayer())
	assert.Equal(t, domain.FrenchTarotResultWin, g.GetResult())
}

func TestFrenchTarotGameEndHumanLose(t *testing.T) {
	// Human (seat 0) is declarer and loses; a CPU ends sole leader.
	g := frenchTarotForceGameEnd(0, [domain.FrenchTarotPlayerCnt]int{0, 10, 0, 0})
	require.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerPlayer())
	assert.Equal(t, domain.FrenchTarotResultLose, g.GetResult())
}

func TestFrenchTarotGameEndDraw(t *testing.T) {
	// Pre-load so the post-deal totals are all equal -> a tie, winnerPlayer -1.
	g := frenchTarotForceGameEnd(0, [domain.FrenchTarotPlayerCnt]int{324, 0, 0, 0})
	require.True(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerPlayer())
	assert.Equal(t, domain.FrenchTarotResultNone, g.GetResult())
}

// --- Phase / turn guards ---

func TestFrenchTarotActionGuards(t *testing.T) {
	// Wrong-phase bid/pass (game is in Play).
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	assert.ErrorIs(t, g.PlayerBid(domain.FrenchTarotBidPetite), domain.ErrWrongPhase)
	assert.ErrorIs(t, g.PlayerPass(), domain.ErrWrongPhase)

	// Invalid bid value.
	g2 := frenchTarotNewReset()
	g2.SetBidPlayerIdx(0)
	require.Error(t, g2.PlayerBid(domain.FrenchTarotBidPass))

	// Not the human's bid turn.
	g3 := frenchTarotNewReset()
	g3.SetBidPlayerIdx(1)
	assert.ErrorIs(t, g3.PlayerBid(domain.FrenchTarotBidPetite), domain.ErrNotHumanTurn)
	assert.ErrorIs(t, g3.PlayerPass(), domain.ErrNotHumanTurn)

	// Discard in the wrong phase.
	g4 := frenchTarotNewReset()
	assert.ErrorIs(t, g4.PlayerDiscard([]int{0, 1, 2, 3, 4, 5}), domain.ErrWrongPhase)

	// Discard when the declarer is a CPU.
	g5 := frenchTarotNewReset()
	g5.SetDeclarerIdx(1)
	g5.SetContract(domain.FrenchTarotBidPetite)
	g5.SetPhase(domain.FrenchTarotPhaseChien)
	assert.ErrorIs(t, g5.PlayerDiscard([]int{0, 1, 2, 3, 4, 5}), domain.ErrNotHumanTurn)

	// Play in the wrong phase and with a bad index.
	g6 := frenchTarotNewReset()
	assert.ErrorIs(t, g6.PlayerPlay(0), domain.ErrWrongPhase)
	g6.SetPhase(domain.FrenchTarotPhasePlay)
	g6.SetCurrentPlayerIdx(0)
	require.Error(t, g6.PlayerPlay(999))

	// Play when it is a CPU's turn.
	g7 := frenchTarotNewReset()
	g7.SetPhase(domain.FrenchTarotPhasePlay)
	g7.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g7.PlayerPlay(0), domain.ErrNotHumanTurn)
}

func TestFrenchTarotGameEndedGuards(t *testing.T) {
	g := frenchTarotForceGameEnd(0, [domain.FrenchTarotPlayerCnt]int{324, 0, 0, 0})
	require.True(t, g.GetGameEndFlag())
	assert.ErrorIs(t, g.PlayerBid(domain.FrenchTarotBidPetite), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerPass(), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDiscard([]int{0, 1, 2, 3, 4, 5}), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrGameEnded)
	// CPU actions are no-ops once the game has ended.
	g.CpuBid()
	g.CpuDiscard()
	g.CpuPlay()
	assert.True(t, g.GetGameEndFlag())
}

func TestFrenchTarotCpuActionWrongPhaseNoop(t *testing.T) {
	g := frenchTarotNewReset() // Bid phase
	g.CpuDiscard()             // wrong phase -> no-op
	g.CpuPlay()                // wrong phase -> no-op
	assert.Equal(t, domain.FrenchTarotPhaseBid, g.GetPhase())

	// CpuBid on a human seat is a no-op.
	g.SetBidPlayerIdx(0)
	g.CpuBid()
	assert.Equal(t, domain.FrenchTarotBidPass, g.GetHighestBid())
}

func TestFrenchTarotTransitionGuards(t *testing.T) {
	g := frenchTarotNewReset() // Bid phase
	g.NextRound()              // not RoundEnd -> no-op
	g.ScoreRound()             // not RoundEnd -> no-op
	g.ResolveTrick()           // not TrickEnd -> no-op
	g.NextTrick()              // not TrickEnd -> no-op
	assert.Equal(t, domain.FrenchTarotPhaseBid, g.GetPhase())
}

// --- Turn predicates (false branches) ---

func TestFrenchTarotTurnPredicates(t *testing.T) {
	g := frenchTarotNewReset()              // Bid phase, seat 0 human
	assert.False(t, g.IsHumanTurn())        // not Play phase
	assert.False(t, g.IsHumanDiscardTurn()) // not Chien phase
	g.SetBidPlayerIdx(1)
	assert.False(t, g.IsHumanBidTurn()) // CPU seat
	g.SetPhase(domain.FrenchTarotPhasePlay)
	assert.False(t, g.IsHumanBidTurn()) // wrong phase
}

// --- Empty-trick winner / all-excuse led suit ---

func TestFrenchTarotTrickWinnerEmpty(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetCurrentTrick(nil)
	assert.Equal(t, 0, g.TrickWinnerPublic())
	assert.Equal(t, -1, g.LedSuitPublic())
}

func TestFrenchTarotExcuseLedNotYetResolved(t *testing.T) {
	g := frenchTarotNewReset()
	g.SetPhase(domain.FrenchTarotPhasePlay)
	g.SetContract(domain.FrenchTarotBidPetite)
	g.SetDeclarerIdx(0)
	frenchTarotSetHand(g, 1,
		frenchTarotSuitCard(domain.CardDesignHeart, 5),
		frenchTarotSuitCard(domain.CardDesignSpade, 3),
		frenchTarotTrumpCard(4))
	// Only the Excuse has been led so far: led suit is undetermined -> any card is legal.
	g.SetCurrentTrick(frenchTarotTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: frenchTarotExcuseCard()}))
	g.SetCurrentPlayerIdx(1)
	assert.ElementsMatch(t, []int{0, 1, 2}, g.GetPlayableIndices(1))
}

// --- More JSON unmarshal validation branches ---

func TestFrenchTarotUnmarshalValidationBranches(t *testing.T) {
	four := frenchTarotFourPlayers()
	base := func(extra string) string { return `{"ps":` + four + `,"ph":0` + extra + `}` }
	cases := map[string]string{
		"badChienCard":      base(`,"ch":[{"design":9,"value":9}]`),
		"badStashCard":      base(`,"st":[{"design":5,"value":99}]`),
		"badLastTrickCard":  base(`,"lc":[{"design":5,"value":99}]`),
		"nilTrickCard":      base(`,"ct":[null]`),
		"trickBadPlayerIdx": base(`,"ct":[{"pi":9,"c":{"design":1,"value":1}}]`),
		"badDealerIdx":      base(`,"di":9`),
		"badLeadIdx":        base(`,"li":9`),
		"badStashOwner":     base(`,"so":5`),
		"badHighestBid":     base(`,"hb":9`),
		"badContract":       base(`,"co":9`),
		"badOutcome":        base(`,"oc":9`),
		"badResult":         base(`,"rs":9`),
		"playNeedsDeclarer": `{"ps":` + four + `,"ph":2,"co":0}`,
		"badConfig":         `{"ps":` + four + `,"ph":0,"cf":{"cd":99,"td":4}}`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.FrenchTarot
			assert.Error(t, json.Unmarshal([]byte(in), &g))
		})
	}
}

func TestFrenchTarotUnmarshalOversized(t *testing.T) {
	// A deck array beyond the max slice length is rejected before element checks.
	oversized := `{"dk":[` + strings.TrimSuffix(strings.Repeat("null,", 5001), ",") + `]}`
	var g domain.FrenchTarot
	assert.Error(t, json.Unmarshal([]byte(oversized), &g))
}

func TestFrenchTarotUnmarshalNullSlicesNormalized(t *testing.T) {
	// Valid minimal state with null slices: they are normalized to empty slices.
	in := `{"ps":` + frenchTarotFourPlayers() + `,"ph":0,"cf":{"cd":1,"td":4}}`
	var g domain.FrenchTarot
	require.NoError(t, json.Unmarshal([]byte(in), &g))
	assert.Equal(t, 0, g.GetChienCount())
	assert.NotNil(t, g.GetChien())
	assert.NotNil(t, g.GetActionLog())
}
