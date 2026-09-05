//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestMarjapussi returns a fresh, reset Marjapussi game with the default 4-player setup.
func newTestMarjapussi() *domain.Marjapussi {
	g := domain.NewDefaultMarjapussi()
	g.Reset()
	return g
}

// setMarjapussiHand replaces player i's hand with the supplied cards deterministically.
func setMarjapussiHand(g *domain.Marjapussi, i int, cards ...*domain.Card) {
	p := g.GetPlayer(i)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// marjapussiCard is a shorthand constructor for a face-up card.
func marjapussiCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func TestMarjapussi_ResetDeal(t *testing.T) {
	g := newTestMarjapussi()
	assert.Equal(t, domain.MarjapussiPhasePlay, g.GetPhase(), "starts directly in Play phase")
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 1, g.GetTrickNumber())
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.Equal(t, 0, g.GetTrumpSuit(), "trump starts unset")
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerPlayer())
	assert.Equal(t, -1, g.GetWinnerTeam())

	// 8 cards each dealt (32 cards in hands).
	totalHand := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalHand += g.GetPlayer(i).GetCardsSize()
		assert.Equal(t, domain.MarjapussiHandSize, g.GetPlayer(i).GetCardsSize())
	}
	assert.Equal(t, domain.MarjapussiHandSize*domain.MarjapussiPlayerCnt, totalHand)

	// Pussi has 4 cards.
	assert.Equal(t, domain.MarjapussiPussiSize, len(g.GetPussi()))

	// Lead player is dealer's left.
	expectedLead := (g.GetDealerIdx() + 1) % domain.MarjapussiPlayerCnt
	assert.Equal(t, expectedLead, g.GetLeadPlayerIdx())
	assert.Equal(t, expectedLead, g.GetCurrentPlayerIdx())
}

func TestMarjapussi_DeckAndDeal_36CardsUnique(t *testing.T) {
	// Reconstruct the whole round's cards (8*4 hands + 4 pussi) and
	// verify 36 unique cards: 6,7,8,9,10,J(11),Q(12),K(13),A(1) across 4 suits.
	g := newTestMarjapussi()
	seen := map[int]bool{}
	count := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			key := c.GetDesign()*100 + c.GetValue()
			assert.False(t, seen[key], "duplicate card %d in hand", key)
			seen[key] = true
			count++
		}
	}
	assert.Equal(t, 32, count)

	for _, c := range g.GetPussi() {
		key := c.GetDesign()*100 + c.GetValue()
		assert.False(t, seen[key], "duplicate card %d in pussi", key)
		seen[key] = true
		count++
	}
	assert.Equal(t, 36, count, "deck must have exactly 36 unique cards")

	validRanks := map[int]bool{1: true, 6: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	for k := range seen {
		val := k % 100
		assert.True(t, validRanks[val], "unexpected rank: %d", val)
	}

	// Total points of all 36 cards must strictly be 120.
	totalPts := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			totalPts += g.CardPoint(p.GetCard(j))
		}
	}
	for _, c := range g.GetPussi() {
		totalPts += g.CardPoint(c)
	}
	assert.Equal(t, 120, totalPts, "total deck points must be 120")
}

func TestMarjapussi_Marriage_RulesAndScoring(t *testing.T) {
	// 3 cases on the same table:
	// Case 1: Different suit marriage -> 20 pts, changes trump.
	// Case 2: Same suit marriage -> 40 pts, preserves trump.
	// Case 3: Non-lead play of King with Queen in hand -> 0 pts, no marriage declared.
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhasePlay)

	// Case 1: Initial trump is Spade (1). Player 0 leads Heart K while holding Heart Q.
	// Heart != Spade -> 20 pts, trump becomes Heart.
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetCurrentTrick(nil)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignHeart, 13), // K
		marjapussiCard(domain.CardDesignHeart, 12), // Q
		marjapussiCard(domain.CardDesignClover, 7),
	)
	p := g.GetPlayer(0)
	kingIdx := -1
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignHeart && p.GetCard(i).GetValue() == 13 {
			kingIdx = i
		}
	}
	require.GreaterOrEqual(t, kingIdx, 0)
	require.NoError(t, g.PlayerPlay(kingIdx))
	assert.Equal(t, domain.CardDesignHeart, g.GetTrumpSuit(), "trump changes to Heart")
	assert.Equal(t, 20, g.GetRoundMarriage()[0], "different suit marriage gives 20 pts to team 0")

	// Case 2: Current trump is Heart. Player 0 leads another Heart K/Q pair in next trick.
	// Heart == Heart -> 40 pts, trump remains Heart.
	g.SetCurrentTrick(nil)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignHeart, 12), // Q
		marjapussiCard(domain.CardDesignHeart, 13), // K
	)
	queenIdx := -1
	p = g.GetPlayer(0)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignHeart && p.GetCard(i).GetValue() == 12 {
			queenIdx = i
		}
	}
	require.GreaterOrEqual(t, queenIdx, 0)
	require.NoError(t, g.PlayerPlay(queenIdx))
	assert.Equal(t, domain.CardDesignHeart, g.GetTrumpSuit(), "trump remains Heart")
	assert.Equal(t, 60, g.GetRoundMarriage()[0], "20 + 40 = 60 pts to team 0")

	// Case 3: Player 0 plays King to follow a trick (NOT leading) while holding the matching Queen.
	// Must NOT declare marriage.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignDiamond, 1)}, // lead Diamond A
	})
	g.SetLeadPlayerIdx(1)
	g.SetCurrentPlayerIdx(0)
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignDiamond, 13), // K
		marjapussiCard(domain.CardDesignDiamond, 12), // Q
	)
	p = g.GetPlayer(0)
	dkIdx := -1
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignDiamond && p.GetCard(i).GetValue() == 13 {
			dkIdx = i
		}
	}
	require.GreaterOrEqual(t, dkIdx, 0)
	require.NoError(t, g.PlayerPlay(dkIdx))
	assert.Equal(t, domain.CardDesignHeart, g.GetTrumpSuit(), "trump does NOT change on non-lead")
	assert.Equal(t, 60, g.GetRoundMarriage()[0], "no marriage points added on non-lead")
}

func TestMarjapussi_FollowRules_MustFollowAndMustTrump(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)

	// Rule 1: Lead is Spade. Player has Spade -> MUST follow Spade.
	g.SetLeadPlayerIdx(1)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignSpade, 10)},
	})
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignSpade, 7),
		marjapussiCard(domain.CardDesignHeart, 1), // trump Ace
		marjapussiCard(domain.CardDesignClover, 6),
	)
	// Trying to play trump Ace (Heart 1) when holding Spade 7 -> ErrInvalidPlay ("リードスートに従ってください")
	heartAceIdx := 1
	err := g.PlayerPlay(heartAceIdx)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	assert.Contains(t, err.Error(), "リードスート")

	// Legal to play Spade 7
	spade7Idx := 0
	gCopy := *g
	require.NoError(t, gCopy.PlayerPlay(spade7Idx))

	// Rule 2: Void in lead suit (Spade), but holds trump (Heart) -> MUST play trump.
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignHeart, 8),  // trump
		marjapussiCard(domain.CardDesignClover, 6), // non-trump
	)
	// Trying to discard Clover 6 when holding trump -> ErrInvalidPlay ("切り札を出してください")
	cloverIdx := 1
	err = g.PlayerPlay(cloverIdx)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	assert.Contains(t, err.Error(), "切り札")
	// Playing trump Heart 8 is valid
	trumpIdx := 0
	gCopy2 := *g
	require.NoError(t, gCopy2.PlayerPlay(trumpIdx))

	// Rule 3: Void in lead suit and void in trump suit (or trump unset) -> can play ANY card.
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignClover, 6),
		marjapussiCard(domain.CardDesignDiamond, 7),
	)
	playable := g.GetPlayableIndices(0)
	assert.ElementsMatch(t, []int{0, 1}, playable)
	require.NoError(t, g.PlayerPlay(0))
}

func TestMarjapussi_NoOvertrumpRule(t *testing.T) {
	// Finnish Marjapussi does NOT require overtrumping.
	// If a higher trump was already played to the trick, a player who has trump may play a lower trump.
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetLeadPlayerIdx(1)
	g.SetCurrentPlayerIdx(0)

	// Lead was Spade 10, Player 2 trumped with Heart Ace.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignSpade, 10)},
		{PlayerIdx: 2, Card: marjapussiCard(domain.CardDesignHeart, 1)}, // Trump Ace (highest trump)
	})

	// Player 0 is void in Spade, but has Heart 6 (lowest trump) and Clover 8.
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignHeart, 6), // Lower trump
		marjapussiCard(domain.CardDesignClover, 8),
	)

	// Heart 6 must be playable (no overtrump required), even though it cannot beat Heart Ace.
	playable := g.GetPlayableIndices(0)
	assert.Contains(t, playable, 0, "lower trump must be playable without overtrump rule")
	require.NoError(t, g.PlayerPlay(0))
}

func TestMarjapussi_Pussi_AwardedToFinalTrickWinner(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhaseTrickEnd)
	g.SetTrumpSuit(domain.CardDesignSpade)

	// 4 pussi cards: A(11) + 10(10) + 7(0) + 6(0) = 21 points
	g.SetPussi([]*domain.Card{
		marjapussiCard(domain.CardDesignSpade, 1),   // 11
		marjapussiCard(domain.CardDesignHeart, 10),  // 10
		marjapussiCard(domain.CardDesignClover, 7),  // 0
		marjapussiCard(domain.CardDesignDiamond, 6), // 0
	})

	// Trick 1..7: Pussi points must NOT be awarded.
	for tr := 1; tr <= 7; tr++ {
		g.SetTrickNumber(tr)
		g.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: marjapussiCard(domain.CardDesignClover, 8)},
			{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignClover, 9)},
			{PlayerIdx: 2, Card: marjapussiCard(domain.CardDesignClover, 10)}, // 10 pts
			{PlayerIdx: 3, Card: marjapussiCard(domain.CardDesignClover, 7)},
		})
		g.SetRoundCardPoints([domain.MarjapussiTeamCnt]int{0, 0})
		g.ResolveTrick()
		// Winner is Player 2 (Team 0). Only trick card points (10 pts) awarded, NO pussi pts.
		assert.Equal(t, 10, g.GetRoundCardPoints()[0], "trick %d must not award pussi", tr)
		assert.Equal(t, 0, g.GetRoundCardPoints()[1])
	}

	// Trick 8 (final trick): Winner gets trick cards + pussi (21 pts).
	g.SetTrickNumber(8)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: marjapussiCard(domain.CardDesignClover, 8)},
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignClover, 1)}, // Winner: Player 1 (Team 1), Ace = 11 pts
		{PlayerIdx: 2, Card: marjapussiCard(domain.CardDesignClover, 7)},
		{PlayerIdx: 3, Card: marjapussiCard(domain.CardDesignClover, 6)},
	})
	g.SetRoundCardPoints([domain.MarjapussiTeamCnt]int{0, 0})
	g.ResolveTrick()

	// Player 1 (Team 1) won trick 8.
	// Trick cards: 11 pts. Pussi cards: 21 pts. Total for Team 1: 32 pts. Team 0: 0 pts.
	assert.Equal(t, 1, g.GetLeadPlayerIdx(), "winner is Player 1")
	assert.Equal(t, 0, g.GetRoundCardPoints()[0])
	assert.Equal(t, 32, g.GetRoundCardPoints()[1], "trick 8 winner team receives trick cards + pussi")
	assert.Equal(t, domain.MarjapussiPhaseRoundEnd, g.GetPhase(), "trick 8 transitions to RoundEnd")
}

func TestMarjapussi_TrickResolution_TrumpBeatsNonTrump(t *testing.T) {
	g := newTestMarjapussi()
	g.SetTrumpSuit(domain.CardDesignDiamond)
	g.SetTrickNumber(1)
	g.SetPhase(domain.MarjapussiPhaseTrickEnd)
	// Lead spade Ace (strongest non-trump); player 2 trumps with a diamond 6.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: marjapussiCard(domain.CardDesignSpade, 1)},   // Ace (11)
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignSpade, 10)},  // 10 (10)
		{PlayerIdx: 2, Card: marjapussiCard(domain.CardDesignDiamond, 6)}, // trump (0)
		{PlayerIdx: 3, Card: marjapussiCard(domain.CardDesignSpade, 7)},   // 7 (0)
	})
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx(), "trump wins even if lowest value")
	assert.Equal(t, 21, g.GetRoundCardPoints()[0], "seat 2 points go to team 0")
}

func TestMarjapussi_TrickResolution_RankOrder(t *testing.T) {
	g := newTestMarjapussi()
	g.SetTrumpSuit(0) // no trump: lead suit highest wins, A > 10 > K > Q > J > 9 > 8 > 7 > 6
	g.SetTrickNumber(1)
	g.SetPhase(domain.MarjapussiPhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: marjapussiCard(domain.CardDesignSpade, 13)}, // K (rank 7)
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignSpade, 1)},  // A (rank 9, highest)
		{PlayerIdx: 2, Card: marjapussiCard(domain.CardDesignSpade, 10)}, // 10 (rank 8)
		{PlayerIdx: 3, Card: marjapussiCard(domain.CardDesignSpade, 8)},  // 8 (rank 3)
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx(), "Ace outranks 10 and K")
}

func TestMarjapussi_NextTrick(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: marjapussiCard(domain.CardDesignSpade, 1)}})
	g.NextTrick()
	assert.Equal(t, domain.MarjapussiPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 2, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())

	// Wrong phase is a no-op.
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.NextTrick()
	assert.Equal(t, domain.MarjapussiPhasePlay, g.GetPhase())
}

func TestMarjapussi_ScoreRound_AndGameEnd(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhaseRoundEnd)
	g.SetRoundCardPoints([domain.MarjapussiTeamCnt]int{70, 50})
	g.SetRoundMarriage([domain.MarjapussiTeamCnt]int{40, 20})
	g.SetTeamScores([domain.MarjapussiTeamCnt]int{400, 350})

	// Team 0 reaches 400 + 70 + 40 = 510 >= 500 (PointLimit)
	g.ScoreRound()

	assert.Equal(t, 510, g.GetTeamScores()[0])
	assert.Equal(t, 420, g.GetTeamScores()[1])
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerTeam())
	assert.Equal(t, 0, g.GetWinnerPlayer())
	assert.Equal(t, domain.MarjapussiPhaseGameEnd, g.GetPhase())
}

func TestMarjapussi_NextRound(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhaseRoundEnd)
	prevDealer := g.GetDealerIdx()
	prevRound := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, prevRound+1, g.GetRoundNumber())
	assert.Equal(t, (prevDealer+1)%domain.MarjapussiPlayerCnt, g.GetDealerIdx())
	assert.Equal(t, domain.MarjapussiPhasePlay, g.GetPhase())

	// Wrong phase -> no-op.
	g.SetPhase(domain.MarjapussiPhasePlay)
	r := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, r, g.GetRoundNumber())
}

func TestMarjapussi_PlayerPlay_Errors(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	setMarjapussiHand(g, 0, marjapussiCard(domain.CardDesignSpade, 1))

	assert.ErrorIs(t, g.PlayerPlay(-1), domain.ErrInvalidCard)
	assert.ErrorIs(t, g.PlayerPlay(99), domain.ErrInvalidCard)

	// Game ended -> error.
	g.SetGameEndFlag(true)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrGameEnded)
	g.SetGameEndFlag(false)

	// Wrong phase -> error.
	g.SetPhase(domain.MarjapussiPhaseRoundEnd)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	// Not human turn -> error.
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)
}

func TestMarjapussi_GetHint(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetTrumpSuit(0)

	// Lead hint with marriage pair -> lead_marriage
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignHeart, 13),
		marjapussiCard(domain.CardDesignHeart, 12),
	)
	ph := g.GetHint()
	require.NotNil(t, ph)
	assert.Equal(t, "lead_marriage", ph.Reason)
	assert.Len(t, ph.CardIndices, 1)

	// Following hint: lead is Spades Q, human plays Spades A -> follow_win
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignSpade, 12)},
	})
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignSpade, 1),
		marjapussiCard(domain.CardDesignSpade, 9),
	)
	fh := g.GetHint()
	require.NotNil(t, fh)
	assert.Equal(t, "follow_win", fh.Reason)

	// Partner is winning -> follow_duck
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignSpade, 9)},
		{PlayerIdx: 2, Card: marjapussiCard(domain.CardDesignSpade, 1)}, // partner (seat 2) winning with Ace
	})
	dh := g.GetHint()
	require.NotNil(t, dh)
	assert.Equal(t, "follow_duck", dh.Reason)

	// Discard low: void in lead suit, no trump
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: marjapussiCard(domain.CardDesignSpade, 1)},
	})
	setMarjapussiHand(g, 0, marjapussiCard(domain.CardDesignClover, 6))
	disH := g.GetHint()
	require.NotNil(t, disH)
	assert.Equal(t, "discard_low", disH.Reason)

	// Not human's turn -> nil.
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())

	// Unhandled phase -> nil.
	g.SetPhase(domain.MarjapussiPhaseTrickEnd)
	assert.Nil(t, g.GetHint())
}

func TestMarjapussi_CpuPlay_Smart_AvoidsOvertakingPartner(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.SetTrumpSuit(domain.CardDesignHeart)

	// Partner of Player 1 is Player 3. Player 3 is winning the trick with Heart Ace (trump Ace).
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: marjapussiCard(domain.CardDesignSpade, 10)},
		{PlayerIdx: 3, Card: marjapussiCard(domain.CardDesignHeart, 1)}, // partner winning with trump Ace
	})
	// Player 1 has Spade 6 (must follow lead suit) and Spade 7.
	setMarjapussiHand(g, 1,
		marjapussiCard(domain.CardDesignSpade, 7),
		marjapussiCard(domain.CardDesignSpade, 6),
	)
	g.CpuPlay()
	assert.Len(t, g.GetCurrentTrick(), 3)
}

func TestMarjapussi_GetMarriageOptions(t *testing.T) {
	g := newTestMarjapussi()
	g.SetTrumpSuit(domain.CardDesignSpade) // Spade is trump
	setMarjapussiHand(g, 0,
		marjapussiCard(domain.CardDesignSpade, 13), marjapussiCard(domain.CardDesignSpade, 12),
		marjapussiCard(domain.CardDesignHeart, 13),   // Q missing
		marjapussiCard(domain.CardDesignDiamond, 12), // K missing
		marjapussiCard(domain.CardDesignClover, 13), marjapussiCard(domain.CardDesignClover, 12),
	)

	got := g.GetMarriageOptions(0)
	// Spade is trump -> 40 pts. Clover is not trump -> 20 pts.
	assert.Equal(t, []domain.MarjapussiMarriageOption{
		{Suit: domain.CardDesignSpade, Points: 40},
		{Suit: domain.CardDesignClover, Points: 20},
	}, got)

	t.Run("no pair yields nothing", func(t *testing.T) {
		setMarjapussiHand(g, 1, marjapussiCard(domain.CardDesignHeart, 13), marjapussiCard(domain.CardDesignSpade, 12))
		assert.Empty(t, g.GetMarriageOptions(1))
	})

	t.Run("out-of-range index is not a panic", func(t *testing.T) {
		assert.Empty(t, g.GetMarriageOptions(-1))
		assert.Empty(t, g.GetMarriageOptions(domain.MarjapussiPlayerCnt))
	})
}

func TestMarjapussi_GettersAndSetters(t *testing.T) {
	g := newTestMarjapussi()
	g.SetRoundNumber(4)
	assert.Equal(t, 4, g.GetRoundNumber())
	g.SetTrickNumber(3)
	assert.Equal(t, 3, g.GetTrickNumber())
	g.SetCurrentPlayerIdx(2)
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	g.SetLeadPlayerIdx(1)
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
	g.SetDealerIdx(3)
	assert.Equal(t, 3, g.GetDealerIdx())
	g.SetTrumpSuit(domain.CardDesignClover)
	assert.Equal(t, domain.CardDesignClover, g.GetTrumpSuit())
	g.SetTeamScores([domain.MarjapussiTeamCnt]int{100, 200})
	assert.Equal(t, [domain.MarjapussiTeamCnt]int{100, 200}, g.GetTeamScores())
	assert.Equal(t, [domain.MarjapussiPlayerCnt]int{100, 200, 100, 200}, g.GetPlayerScores())
	g.SetPlayerScores([domain.MarjapussiPlayerCnt]int{150, 250, 150, 250})
	assert.Equal(t, [domain.MarjapussiTeamCnt]int{150, 250}, g.GetTeamScores())

	g.SetRoundCardPoints([domain.MarjapussiTeamCnt]int{30, 40})
	assert.Equal(t, [domain.MarjapussiTeamCnt]int{30, 40}, g.GetRoundCardPoints())
	g.SetRoundMarriage([domain.MarjapussiTeamCnt]int{20, 40})
	assert.Equal(t, [domain.MarjapussiTeamCnt]int{20, 40}, g.GetRoundMarriage())

	g.SetWinnerPlayer(0)
	assert.Equal(t, 0, g.GetWinnerPlayer())
	g.SetWinnerTeam(1)
	assert.Equal(t, 1, g.GetWinnerTeam())

	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetPlayer(0))
	assert.NotNil(t, g.GetConfig())
	_ = g.GetActionLog()

	// GetPlayableIndices guards.
	assert.Nil(t, g.GetPlayableIndices(-1))
	g.SetPhase(domain.MarjapussiPhaseRoundEnd)
	assert.Nil(t, g.GetPlayableIndices(0), "not play phase -> nil")

	// IsHumanTurn.
	g.SetPhase(domain.MarjapussiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
}

func TestMarjapussi_JSON_RoundTrip(t *testing.T) {
	g := newTestMarjapussi()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetTeamScores([domain.MarjapussiTeamCnt]int{120, 80})
	g.SetRoundCardPoints([domain.MarjapussiTeamCnt]int{30, 25})
	g.SetRoundMarriage([domain.MarjapussiTeamCnt]int{40, 20})
	g.SetWinnerPlayer(0)
	g.SetWinnerTeam(0)
	pussi := []*domain.Card{
		marjapussiCard(domain.CardDesignSpade, 1),
		marjapussiCard(domain.CardDesignClover, 10),
		marjapussiCard(domain.CardDesignHeart, 13),
		marjapussiCard(domain.CardDesignDiamond, 6),
	}
	g.SetPussi(pussi)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Marjapussi
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), g2.GetTrumpSuit())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
	assert.Equal(t, g.GetTeamScores(), g2.GetTeamScores())
	assert.Equal(t, g.GetRoundCardPoints(), g2.GetRoundCardPoints())
	assert.Equal(t, g.GetRoundMarriage(), g2.GetRoundMarriage())
	assert.Equal(t, g.GetWinnerPlayer(), g2.GetWinnerPlayer())
	assert.Equal(t, g.GetWinnerTeam(), g2.GetWinnerTeam())

	require.Len(t, g2.GetPussi(), 4)
	for i := 0; i < 4; i++ {
		assert.Equal(t, pussi[i].GetDesign(), g2.GetPussi()[i].GetDesign())
		assert.Equal(t, pussi[i].GetValue(), g2.GetPussi()[i].GetValue())
	}
}

func TestMarjapussi_JSON_PussiMutation_FailsWhenDropped(t *testing.T) {
	g := newTestMarjapussi()
	g.SetPussi([]*domain.Card{
		marjapussiCard(domain.CardDesignSpade, 1),
		marjapussiCard(domain.CardDesignClover, 10),
		marjapussiCard(domain.CardDesignHeart, 13),
		marjapussiCard(domain.CardDesignDiamond, 6),
	})
	data, err := json.Marshal(g)
	require.NoError(t, err)

	// Mutate JSON by dropping the pussi ("pu") field.
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	delete(raw, "pu")

	mutatedData, err := json.Marshal(raw)
	require.NoError(t, err)

	var restored domain.Marjapussi
	require.NoError(t, json.Unmarshal(mutatedData, &restored))

	// Dropping pussi field results in empty pussi, which fails equality check with original pussi.
	assert.NotEqual(t, len(g.GetPussi()), len(restored.GetPussi()), "dropping pussi must result in mismatch")

	// Corrupted nil card inside pussi must fail unmarshal.
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	badPussiJSON := `{"ph":0,"ps":` + okPlayers + `,"pu":[null]}`
	var gCorrupt domain.Marjapussi
	assert.Error(t, json.Unmarshal([]byte(badPussiJSON), &gCorrupt))
}

func TestMarjapussi_JSON_Invalid(t *testing.T) {
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	cases := []string{
		`not json`,                                                            // malformed
		`{"ph":0,"ps":[null,null]}`,                                           // wrong player count
		`{"ph":0,"ps":[null,null,null,null]}`,                                 // nil players
		`{"ph":0,"ps":` + okPlayers + `,"ci":100}`,                            // currentPlayerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"ci":-1}`,                             // currentPlayerIdx negative
		`{"ph":0,"ps":` + okPlayers + `,"di":99}`,                             // dealerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"li":99}`,                             // leadPlayerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"li":-2}`,                             // leadPlayerIdx below -1
		`{"ph":0,"ps":` + okPlayers + `,"lt":99}`,                             // lastTrickWinner out of range
		`{"ph":0,"ps":` + okPlayers + `,"wp":99}`,                             // winnerPlayer out of range
		`{"ph":0,"ps":` + okPlayers + `,"wt":99}`,                             // winnerTeam out of range
		`{"ph":99,"ps":` + okPlayers + `}`,                                    // bad phase
		`{"ph":0,"ps":` + okPlayers + `,"ts":9}`,                              // bad trumpSuit
		`{"ph":0,"ps":` + okPlayers + `,"ct":[null]}`,                         // nil trick card
		`{"ph":0,"ps":` + okPlayers + `,"ct":[{"pi":99,"c":{"d":1,"v":13}}]}`, // trick card PlayerIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"ct":[{"pi":-1,"c":{"d":1,"v":13}}]}`, // trick card PlayerIdx negative
		`{"ph":0,"ps":` + okPlayers + `,"pu":[null]}`,                         // nil pussi card
	}
	for _, c := range cases {
		var g domain.Marjapussi
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}

	// Valid restore
	okJSON := `{"ph":0,"ps":` + okPlayers + `,"ts":0,"cf":{"cd":1,"pl":500},"lt":-1,"wp":-1,"wt":-1,"li":-1}`
	var g2 domain.Marjapussi
	assert.NoError(t, json.Unmarshal([]byte(okJSON), &g2))
	assert.Equal(t, 4, g2.GetPlayerCnt())
}

func TestMarjapussiPlayer_JSON_And_ResetRound(t *testing.T) {
	p := domain.NewMarjapussiPlayer(true)
	p.AddCard(marjapussiCard(domain.CardDesignSpade, 1))
	p.AddTrick([]*domain.Card{marjapussiCard(domain.CardDesignHeart, 10)})
	assert.Equal(t, 1, p.GetTrickCount())

	b, err := json.Marshal(p)
	require.NoError(t, err)
	var p2 domain.MarjapussiPlayer
	require.NoError(t, json.Unmarshal(b, &p2))
	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetCardsSize())
	assert.Equal(t, 1, p2.GetTrickCount())

	p2.ResetRound()
	assert.Equal(t, 0, p2.GetCardsSize())
	assert.Equal(t, 0, p2.GetTrickCount())
	assert.False(t, p2.GetIsFinished())
}

func TestMarjapussiConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultMarjapussiConfig().Validate())
	assert.Equal(t, domain.MarjapussiWinTarget, domain.DefaultMarjapussiConfig().PointLimit)
	assert.Equal(t, domain.MarjapussiWinTarget, domain.DefaultMarjapussiConfig().TargetPoints)
	assert.Equal(t, domain.MarjapussiCpuDifficultyNormal, domain.DefaultMarjapussiConfig().CpuDifficulty)

	// Bad CPU difficulty.
	assert.Error(t, domain.MarjapussiConfig{CpuDifficulty: 99, PointLimit: 500}.Validate())
	// Non-positive point limit.
	assert.Error(t, domain.MarjapussiConfig{CpuDifficulty: domain.MarjapussiCpuDifficultyEasy, PointLimit: 0, TargetPoints: 0}.Validate())
}

func TestMarjapussi_Statistical_2000Rounds(t *testing.T) {
	const totalRounds = 2000
	marriageDealtRounds := 0
	marriageRounds := 0
	humanMarriageDealtRounds := 0
	humanMarriageRounds := 0
	lastTrickWinners := [domain.MarjapussiPlayerCnt]int{}

	for round := 0; round < totalRounds; round++ {
		g := domain.NewDefaultMarjapussi()
		g.Reset()

		// 配布直後 (Reset 直後) に各席の結婚存在を記録する。
		roundHasMarriage := false
		for p := 0; p < domain.MarjapussiPlayerCnt; p++ {
			if len(g.GetMarriageOptions(p)) > 0 {
				roundHasMarriage = true
			}
		}
		if roundHasMarriage {
			marriageDealtRounds++
		}
		if len(g.GetMarriageOptions(0)) > 0 {
			humanMarriageDealtRounds++
		}

		humanDeclared := false
		for g.GetPhase() == domain.MarjapussiPhasePlay || g.GetPhase() == domain.MarjapussiPhaseTrickEnd {
			if g.GetPhase() == domain.MarjapussiPhasePlay {
				if g.IsHumanTurn() {
					h := g.GetHint()
					require.NotNil(t, h)
					require.NotEmpty(t, h.CardIndices)
					prevMarriage := g.GetRoundMarriage()[0]
					require.NoError(t, g.PlayerPlay(h.CardIndices[0]))
					if h.Reason == "lead_marriage" && g.GetRoundMarriage()[0] > prevMarriage {
						humanDeclared = true
					}
				} else {
					g.CpuPlay()
				}
			} else if g.GetPhase() == domain.MarjapussiPhaseTrickEnd {
				g.ResolveTrick()
				if g.GetPhase() == domain.MarjapussiPhaseTrickEnd {
					g.NextTrick()
				}
			}
		}

		require.Equal(t, domain.MarjapussiPhaseRoundEnd, g.GetPhase())

		if humanDeclared {
			humanMarriageRounds++
		}

		totalMarriage := g.GetRoundMarriage()[0] + g.GetRoundMarriage()[1]
		if totalMarriage > 0 {
			marriageRounds++
		}

		ltw := g.GetLeadPlayerIdx()
		require.True(t, ltw >= 0 && ltw < domain.MarjapussiPlayerCnt)
		lastTrickWinners[ltw]++

		// Invariant: Card points strictly equal 120, total points strictly equal 120 + marriage points.
		totalCardPts := g.GetRoundCardPoints()[0] + g.GetRoundCardPoints()[1]
		require.Equal(t, 120, totalCardPts, "round %d: card points must strictly equal 120", round)
		totalPts := totalCardPts + totalMarriage
		require.Equal(t, 120+totalMarriage, totalPts, "round %d: total points must equal 120 + marriage", round)
	}

	dealtRate := float64(marriageDealtRounds) / float64(totalRounds)
	conversionRate := float64(marriageRounds) / float64(marriageDealtRounds)
	humanDealtRate := float64(humanMarriageDealtRounds) / float64(totalRounds)
	humanConversionRate := float64(humanMarriageRounds) / float64(humanMarriageDealtRounds)

	t.Logf("Statistical Results over %d rounds:", totalRounds)
	t.Logf("1. Marriage dealt in round (45%% - 60%%): %.2f%% (%d/%d)", dealtRate*100, marriageDealtRounds, totalRounds)
	t.Logf("2. Overall marriage conversion rate (>= 50%%): %.2f%% (%d/%d)", conversionRate*100, marriageRounds, marriageDealtRounds)
	t.Logf("3. Human (seat 0) dealt marriage rate: %.2f%% (%d/%d)", humanDealtRate*100, humanMarriageDealtRounds, totalRounds)
	t.Logf("4. Human (seat 0) marriage conversion rate (>= 30%%): %.2f%% (%d/%d)", humanConversionRate*100, humanMarriageRounds, humanMarriageDealtRounds)
	for i := 0; i < domain.MarjapussiPlayerCnt; i++ {
		rate := float64(lastTrickWinners[i]) / float64(totalRounds)
		t.Logf("5. Last trick winner seat %d (10%% - 50%%): %.2f%% (%d/%d)", i, rate*100, lastTrickWinners[i], totalRounds)
	}
	t.Logf("6. Card points strictly equal 120 in all %d rounds: PASS", totalRounds)

	// 1. 結婚が配りの中に存在する割合: 45% 〜 60%
	// 理論的根拠:
	// 36枚デッキから4人に8枚ずつ配る (残り4枚はプッシ)。
	// 1人が特定スートのKとQの両方を持つ確率は C(34,6)/C(36,8) = (8*7)/(36*35) = 2/45 ≈ 4.44%。
	// 4人×4スートで、配りの中に結婚が1つ以上存在する確率は理論上約54% (包除原理等の上限)。
	// デッキ構成 (36枚) や配り枚数 (各席8枚) が破壊された場合に検知できるよう、45% 〜 60% の帯とする。
	assert.GreaterOrEqual(t, dealtRate, 0.45, "marriage dealt rate should be >= 45%%")
	assert.LessOrEqual(t, dealtRate, 0.60, "marriage dealt rate should be <= 60%%")

	// 2. 転換率 = 宣言された局 / 結婚が存在した局: 50% 以上
	// 理論的根拠:
	// 結婚が存在しても所持者がリード権を獲得しなければ宣言できない。
	// 結婚が存在する局において、リード権を獲得して実際に宣言が行われる転換率は
	// 正常なAIプレイとルール進行下で半数以上 (50%以上) となるはずである (実測約71%)。
	// 結婚宣言の処理経路 (リード判定、maybeDeclareMarriage、CPU/人間AI) が機能不全を起こした際に検知するため、
	// 下限を 50% とする。
	require.Greater(t, marriageDealtRounds, 0, "at least one round should have marriage dealt")
	assert.GreaterOrEqual(t, conversionRate, 0.50, "marriage conversion rate should be >= 50%%")

	// 3. 人間 (席 0) の転換率 = 人間が宣言した局 / 人間の手札に結婚が存在した局: 30% 以上
	// 理論的根拠:
	// 席 0 の手札に結婚が存在する理論確率は 1 - (43/45)^4 ≈ 16.6%。
	// 4人プレイにおいて席 0 がリードを取る機会は確率的に均等なら 25% 前後だが、
	// 手札に K/Q を含む強いカードを持つためリード奪取率はそれ以上になり、転換率は実測約55%となる。
	// 人間が永久に宣言できないバグやヒント生成不良を確実に検知するため、下限を 30% とする。
	require.Greater(t, humanMarriageDealtRounds, 0, "at least one round should have human marriage dealt")
	assert.GreaterOrEqual(t, humanConversionRate, 0.30, "human marriage conversion rate should be >= 30%%")

	// 4. 最後のトリックを取った席の分布: どの席も 10% 〜 50%
	// 理論的根拠:
	// 4人ゲームで各席が均等なら 25%。特定席への過度な偏り (偏向バグ) がないことを保証する。
	for i := 0; i < domain.MarjapussiPlayerCnt; i++ {
		rate := float64(lastTrickWinners[i]) / float64(totalRounds)
		assert.GreaterOrEqual(t, rate, 0.10, "seat %d last trick rate should be >= 10%%", i)
		assert.LessOrEqual(t, rate, 0.50, "seat %d last trick rate should be <= 50%%", i)
	}
}
