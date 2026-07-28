//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestUlti returns a fresh, reset Ulti game with the default 3-player setup.
func newTestUlti() *domain.Ulti {
	g := domain.NewDefaultUlti()
	g.Reset()
	return g
}

// setUltiHand replaces player i's hand with the supplied cards deterministically.
func setUltiHand(g *domain.Ulti, i int, cards ...*domain.Card) {
	p := g.GetPlayer(i)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// ultiCard is a shorthand constructor for a face-up card.
func ultiCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// ultiGiveTricks awards player i exactly n zero-point tricks.
func ultiGiveTricks(g *domain.Ulti, i, n int) {
	p := g.GetPlayer(i)
	for k := 0; k < n; k++ {
		p.AddTrick([]*domain.Card{ultiCard(domain.CardDesignSpade, 7)}) // 7 = 0 points
	}
}

// ultiGivePointTricks awards player i n tricks each holding one card of the given value.
func ultiGivePointTricks(g *domain.Ulti, i, n, value int) {
	p := g.GetPlayer(i)
	for k := 0; k < n; k++ {
		p.AddTrick([]*domain.Card{ultiCard(domain.CardDesignSpade, value)})
	}
}

func TestUlti_ResetDeal(t *testing.T) {
	g := newTestUlti()
	assert.Equal(t, domain.UltiPhaseBid, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 3, g.GetPlayerCnt())
	assert.Equal(t, 0, g.GetDeclarerIdx())
	assert.Equal(t, domain.UltiContractNone, g.GetContract())
	assert.Equal(t, -1, g.GetTrumpSuit())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerPlayer())
	assert.Equal(t, domain.UltiTalonSize, g.GetTalonCount())
	assert.False(t, g.GetTalonTaken())

	totalHand := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalHand += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, domain.UltiHandSize*domain.UltiPlayerCnt, totalHand)
	assert.True(t, g.IsHumanTurn())
	assert.True(t, g.IsHumanBidTurn())
}

func TestUlti_DeckIsUnique32(t *testing.T) {
	g := newTestUlti()
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
	// 30 dealt + 2 talon = 32.
	assert.Equal(t, 30, count)
	valid := map[int]bool{1: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	for k := range seen {
		assert.True(t, valid[k%100], "unexpected rank %d", k%100)
	}
}

func TestUlti_Bid_Declare(t *testing.T) {
	t.Run("party takes trump and talon", func(t *testing.T) {
		g := newTestUlti()
		require.NoError(t, g.PlayerBid(domain.UltiContractParty, domain.CardDesignHeart))
		assert.Equal(t, domain.UltiPhaseDiscard, g.GetPhase())
		assert.Equal(t, domain.UltiContractParty, g.GetContract())
		assert.Equal(t, domain.CardDesignHeart, g.GetTrumpSuit())
		assert.Equal(t, domain.UltiHandSize+domain.UltiTalonSize, g.GetPlayer(0).GetCardsSize())
		assert.True(t, g.GetTalonTaken())
		assert.Equal(t, 0, g.GetTalonCount())
		assert.True(t, g.IsHumanTurn(), "discard phase is a human turn")
	})

	t.Run("betli no trump", func(t *testing.T) {
		g := newTestUlti()
		require.NoError(t, g.PlayerBid(domain.UltiContractBetli, -1))
		assert.Equal(t, domain.UltiContractBetli, g.GetContract())
		assert.Equal(t, -1, g.GetTrumpSuit())
	})

	t.Run("durchmarsch no trump", func(t *testing.T) {
		g := newTestUlti()
		require.NoError(t, g.PlayerBid(domain.UltiContractDurchmarsch, -1))
		assert.Equal(t, domain.UltiContractDurchmarsch, g.GetContract())
	})

	t.Run("errors", func(t *testing.T) {
		g := newTestUlti()
		assert.Error(t, g.PlayerBid(domain.UltiContractNone, -1))  // invalid contract
		assert.Error(t, g.PlayerBid(domain.UltiContractParty, -1)) // party requires trump
		g.SetPhase(domain.UltiPhasePlay)
		assert.ErrorIs(t, g.PlayerBid(domain.UltiContractBetli, -1), domain.ErrWrongPhase)
	})
}

func TestUlti_Discard(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newTestUlti()
		require.NoError(t, g.PlayerBid(domain.UltiContractParty, domain.CardDesignHeart))
		require.NoError(t, g.PlayerDiscard([]int{0, 1}))
		assert.Equal(t, domain.UltiPhasePlay, g.GetPhase())
		assert.Equal(t, domain.UltiHandSize, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.UltiDiscardSize, g.GetDiscardCount())
		assert.Equal(t, 0, g.GetLeadPlayerIdx(), "declarer leads")
		assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	})

	t.Run("errors", func(t *testing.T) {
		g := newTestUlti()
		require.NoError(t, g.PlayerBid(domain.UltiContractBetli, -1))
		assert.Error(t, g.PlayerDiscard([]int{0}))     // wrong count
		assert.Error(t, g.PlayerDiscard([]int{0, 99})) // out of range
		assert.Error(t, g.PlayerDiscard([]int{3, 3}))  // duplicate
		g.SetPhase(domain.UltiPhasePlay)
		assert.ErrorIs(t, g.PlayerDiscard([]int{0, 1}), domain.ErrWrongPhase)
	})
}

func TestUlti_TrickWinner(t *testing.T) {
	resolve := func(contract domain.UltiContract, trump int, trick []*domain.TrickCard) int {
		g := newTestUlti()
		g.SetDeclarerIdx(0)
		g.SetContract(contract)
		g.SetTrumpSuit(trump)
		g.SetTrickNumber(1)
		g.SetPhase(domain.UltiPhaseTrickEnd)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		return g.GetLeadPlayerIdx()
	}

	// Party: any trump beats plain; heart 7 (lowest trump) beats spade Ace.
	assert.Equal(t, 2, resolve(domain.UltiContractParty, domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ultiCard(domain.CardDesignSpade, 1)},  // spade A (plain)
		{PlayerIdx: 1, Card: ultiCard(domain.CardDesignSpade, 10)}, // spade 10 (plain)
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignHeart, 7)},  // heart 7 (trump)
	}))

	// Party: no trump in trick -> highest of led suit (A > 10).
	assert.Equal(t, 0, resolve(domain.UltiContractParty, domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ultiCard(domain.CardDesignSpade, 1)},   // spade A wins
		{PlayerIdx: 1, Card: ultiCard(domain.CardDesignSpade, 10)},  // spade 10
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignClover, 13)}, // off-suit K cannot win
	}))

	// Betli (no trump): Ace beats 10 and K of led suit; off-suit cannot win.
	assert.Equal(t, 1, resolve(domain.UltiContractBetli, -1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ultiCard(domain.CardDesignHeart, 7)}, // led heart 7
		{PlayerIdx: 1, Card: ultiCard(domain.CardDesignHeart, 1)}, // heart A wins
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignSpade, 1)}, // off-suit A cannot win
	}))

	// Trick rank order: 10 > K (10 outranks King).
	assert.Equal(t, 1, resolve(domain.UltiContractBetli, -1, []*domain.TrickCard{
		{PlayerIdx: 0, Card: ultiCard(domain.CardDesignHeart, 13)}, // led K
		{PlayerIdx: 1, Card: ultiCard(domain.CardDesignHeart, 10)}, // 10 beats K
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignHeart, 12)}, // Q
	}))
}

func TestUlti_MustFollow(t *testing.T) {
	g := newTestUlti()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.UltiContractParty)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.UltiPhasePlay)
	g.SetCurrentPlayerIdx(1)

	// Must follow led suit (spade) when holding it.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: ultiCard(domain.CardDesignSpade, 13)},
	})
	setUltiHand(g, 1,
		ultiCard(domain.CardDesignSpade, 9),
		ultiCard(domain.CardDesignHeart, 8))
	assert.Equal(t, []int{0}, g.GetPlayableIndices(1))

	// Void in led, no trump in trick yet -> any card (reduced rule: no forced ruff).
	setUltiHand(g, 1,
		ultiCard(domain.CardDesignHeart, 8),
		ultiCard(domain.CardDesignClover, 9))
	assert.Len(t, g.GetPlayableIndices(1), 2)

	// Overtrump obligation: a trump is already in the trick and player can beat it.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: ultiCard(domain.CardDesignSpade, 13)}, // led spade
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignHeart, 7)},  // trump 7 (rank 0)
	})
	setUltiHand(g, 1,
		ultiCard(domain.CardDesignHeart, 1), // trump A beats trump 7 -> forced
		ultiCard(domain.CardDesignClover, 9))
	assert.Equal(t, []int{0}, g.GetPlayableIndices(1), "must overtrump")

	// Cannot beat the highest trump -> any card allowed.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: ultiCard(domain.CardDesignSpade, 13)},
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignHeart, 1)}, // trump A (top)
	})
	setUltiHand(g, 1,
		ultiCard(domain.CardDesignHeart, 8), // trump 8 cannot beat trump A
		ultiCard(domain.CardDesignClover, 9))
	assert.Len(t, g.GetPlayableIndices(1), 2)
}

func TestUlti_Outcomes(t *testing.T) {
	cases := []struct {
		name     string
		contract domain.UltiContract
		setup    func(g *domain.Ulti)
		outcome  domain.UltiOutcome
		coins    [domain.UltiPlayerCnt]int
	}{
		{"party win", domain.UltiContractParty, func(g *domain.Ulti) {
			ultiGivePointTricks(g, 0, 7, 1) // 7 aces = 70 points >= 61
		}, domain.UltiOutcomeWin, [domain.UltiPlayerCnt]int{4, -2, -2}},
		{"party loss", domain.UltiContractParty, func(g *domain.Ulti) {
			ultiGivePointTricks(g, 0, 5, 1) // 50 points < 61
		}, domain.UltiOutcomeLoss, [domain.UltiPlayerCnt]int{-4, 2, 2}},
		{"betli win", domain.UltiContractBetli, func(g *domain.Ulti) {
			ultiGiveTricks(g, 1, 5) // declarer takes 0 tricks
		}, domain.UltiOutcomeWin, [domain.UltiPlayerCnt]int{10, -5, -5}},
		{"betli loss", domain.UltiContractBetli, func(g *domain.Ulti) {
			ultiGiveTricks(g, 0, 1) // declarer took a trick
		}, domain.UltiOutcomeLoss, [domain.UltiPlayerCnt]int{-10, 5, 5}},
		{"durchmarsch win", domain.UltiContractDurchmarsch, func(g *domain.Ulti) {
			ultiGiveTricks(g, 0, 10) // declarer sweeps
		}, domain.UltiOutcomeWin, [domain.UltiPlayerCnt]int{12, -6, -6}},
		{"durchmarsch loss", domain.UltiContractDurchmarsch, func(g *domain.Ulti) {
			ultiGiveTricks(g, 0, 9)
		}, domain.UltiOutcomeLoss, [domain.UltiPlayerCnt]int{-12, 6, 6}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := newTestUlti()
			g.SetDeclarerIdx(0)
			g.SetContract(c.contract)
			if c.contract == domain.UltiContractParty {
				g.SetTrumpSuit(domain.CardDesignHeart)
			}
			g.SetPhase(domain.UltiPhaseRoundEnd)
			c.setup(g)
			g.ScoreRound()
			assert.Equal(t, c.outcome, g.GetOutcome())
			assert.Equal(t, c.coins, g.GetPlayerCoins())

			// ScoreRound is idempotent (scored flag).
			g.ScoreRound()
			assert.Equal(t, c.coins, g.GetPlayerCoins())
		})
	}
}

func TestUlti_PartyLastTrickBonus(t *testing.T) {
	g := newTestUlti()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.UltiContractParty)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.UltiPhaseRoundEnd)
	ultiGivePointTricks(g, 0, 5, 1) // 50 points
	assert.Equal(t, 50, g.GetCardPoints(0))
	// Without the last-trick bonus this would be a loss; the test just verifies
	// GetCardPoints excludes the bonus and outcome uses points only.
	g.ScoreRound()
	assert.Equal(t, domain.UltiOutcomeLoss, g.GetOutcome())
}

func TestUlti_ScoreRound_WrongPhaseNoop(t *testing.T) {
	g := newTestUlti()
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.UltiPhasePlay)
	g.ScoreRound()
	assert.Equal(t, [domain.UltiPlayerCnt]int{0, 0, 0}, g.GetPlayerCoins())
}

func TestUlti_GameEnd_HumanWins(t *testing.T) {
	g := newTestUlti()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.UltiContractDurchmarsch)
	g.SetRoundNumber(domain.UltiWinRounds)
	g.SetPhase(domain.UltiPhaseRoundEnd)
	ultiGiveTricks(g, 0, 10) // human declarer sweeps -> win
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerPlayer())
	assert.Equal(t, domain.UltiPhaseGameEnd, g.GetPhase())
	assert.Equal(t, domain.UltiResultWin, g.GetResult())
}

func TestUlti_GameEnd_HumanLoses(t *testing.T) {
	g := newTestUlti()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.UltiContractBetli)
	g.SetRoundNumber(domain.UltiWinRounds)
	g.SetPhase(domain.UltiPhaseRoundEnd)
	g.SetPlayerCoins([domain.UltiPlayerCnt]int{0, 8, 0})
	ultiGiveTricks(g, 0, 3) // human declarer took tricks -> betli lost, defenders gain
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerPlayer())
	assert.Equal(t, domain.UltiResultLose, g.GetResult())
}

func TestUlti_NextRoundAndTrick(t *testing.T) {
	g := newTestUlti()
	g.SetPhase(domain.UltiPhaseRoundEnd)
	prevDealer := g.GetDealerIdx()
	prevRound := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, prevRound+1, g.GetRoundNumber())
	assert.Equal(t, (prevDealer+1)%domain.UltiPlayerCnt, g.GetDealerIdx())
	assert.Equal(t, domain.UltiPhaseBid, g.GetPhase())

	// Wrong phase -> no-op.
	g.SetPhase(domain.UltiPhasePlay)
	r := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, r, g.GetRoundNumber())

	// NextTrick.
	g.SetPhase(domain.UltiPhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: ultiCard(domain.CardDesignSpade, 1)}})
	g.NextTrick()
	assert.Equal(t, domain.UltiPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 2, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())

	// NextTrick wrong phase -> no-op.
	g.SetPhase(domain.UltiPhasePlay)
	g.NextTrick()
	assert.Equal(t, domain.UltiPhasePlay, g.GetPhase())
}

func TestUlti_PlayerPlay_Errors(t *testing.T) {
	g := newTestUlti()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.UltiContractParty)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.UltiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	setUltiHand(g, 0, ultiCard(domain.CardDesignSpade, 13))

	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(99))

	// Wrong phase.
	g.SetPhase(domain.UltiPhaseBid)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	// Not human turn.
	g.SetPhase(domain.UltiPhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)
}

func TestUlti_PlayerPlay_FollowViolationAndComplete(t *testing.T) {
	g := newTestUlti()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.UltiContractParty)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.UltiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: ultiCard(domain.CardDesignClover, 13)}, // led club
	})
	setUltiHand(g, 0,
		ultiCard(domain.CardDesignClover, 9), // legal (follows club)
		ultiCard(domain.CardDesignSpade, 8))  // illegal (must follow club)
	spadeIdx := -1
	p := g.GetPlayer(0)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignSpade {
			spadeIdx = i
		}
	}
	assert.ErrorIs(t, g.PlayerPlay(spadeIdx), domain.ErrInvalidPlay)

	// Completing the trick moves to TrickEnd.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: ultiCard(domain.CardDesignSpade, 9)},
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignSpade, 8)},
	})
	setUltiHand(g, 0, ultiCard(domain.CardDesignSpade, 13))
	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, domain.UltiPhaseTrickEnd, g.GetPhase())
}

func TestUlti_GetHint_AllPhases(t *testing.T) {
	// Bid phase.
	g := newTestUlti()
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Contains(t, []string{"bid_party", "bid_betli", "bid_durchmarsch", "bid_ulti"}, h.Reason)

	// Discard phase recommends 2 cards.
	require.NoError(t, g.PlayerBid(domain.UltiContractParty, domain.CardDesignHeart))
	dh := g.GetHint()
	require.NotNil(t, dh)
	assert.Equal(t, "discard_weak", dh.Reason)
	assert.Len(t, dh.CardIndices, domain.UltiDiscardSize)

	// Play phase lead hint (declarer -> lead_high).
	g2 := newTestUlti()
	g2.SetPhase(domain.UltiPhasePlay)
	g2.SetDeclarerIdx(0)
	g2.SetContract(domain.UltiContractParty)
	g2.SetTrumpSuit(domain.CardDesignHeart)
	g2.SetCurrentPlayerIdx(0)
	g2.SetCurrentTrick(nil)
	setUltiHand(g2, 0, ultiCard(domain.CardDesignHeart, 13), ultiCard(domain.CardDesignSpade, 9))
	ph := g2.GetHint()
	require.NotNil(t, ph)
	assert.Equal(t, "lead_high", ph.Reason)

	// Coalition lead -> lead_low.
	g2.SetDeclarerIdx(1)
	g2.SetCurrentPlayerIdx(0)
	g2.SetCurrentTrick(nil)
	lh := g2.GetHint()
	require.NotNil(t, lh)
	assert.Equal(t, "lead_low", lh.Reason)

	// Not the human's turn -> nil.
	g2.SetCurrentPlayerIdx(1)
	assert.Nil(t, g2.GetHint())

	// Unhandled phase -> nil.
	g2.SetPhase(domain.UltiPhaseTrickEnd)
	assert.Nil(t, g2.GetHint())
}

func TestUlti_GetHint_PlayReasons(t *testing.T) {
	g := newTestUlti()
	g.SetDeclarerIdx(2) // player 0 is coalition
	g.SetContract(domain.UltiContractParty)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.UltiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	// Declarer (seat 2) leads a plain club Queen; coalition can win or duck.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignClover, 12)},
	})
	setUltiHand(g, 0,
		ultiCard(domain.CardDesignClover, 13), // K beats Q -> follow_win
		ultiCard(domain.CardDesignClover, 9))  // 9 loses
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Contains(t, []string{"follow_win", "follow_duck"}, h.Reason)

	// discard_low: void in lead suit.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: ultiCard(domain.CardDesignClover, 12)},
	})
	setUltiHand(g, 0, ultiCard(domain.CardDesignSpade, 13)) // off-suit only (no trump either)
	h2 := g.GetHint()
	require.NotNil(t, h2)
	assert.Equal(t, "discard_low", h2.Reason)
}

func TestUlti_CpuFullRound(t *testing.T) {
	g := newTestUlti()
	guard := 0
	for !g.GetGameEndFlag() && guard < 20000 {
		guard++
		switch g.GetPhase() {
		case domain.UltiPhaseBid:
			require.NoError(t, g.PlayerBid(domain.UltiContractParty, domain.CardDesignHeart))
		case domain.UltiPhaseDiscard:
			require.NoError(t, g.PlayerDiscard([]int{0, 1}))
		case domain.UltiPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
		case domain.UltiPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == domain.UltiPhaseTrickEnd {
				g.NextTrick()
			}
		case domain.UltiPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case domain.UltiPhaseGameEnd:
			guard = 20000
		}
	}
	assert.Less(t, guard, 20000, "game flow should progress")
}

func TestUlti_CpuContracts(t *testing.T) {
	// Exercise the CPU AI under each contract with a full deterministic drive.
	for _, contract := range []domain.UltiContract{domain.UltiContractBetli, domain.UltiContractDurchmarsch} {
		g := newTestUlti()
		cfg := g.GetConfig()
		cfg.CpuDifficulty = domain.UltiCpuDifficultyHard
		g.SetConfig(cfg)
		require.NoError(t, g.PlayerBid(contract, -1))
		require.NoError(t, g.PlayerDiscard([]int{0, 1}))
		guard := 0
		for g.GetPhase() != domain.UltiPhaseRoundEnd && guard < 2000 {
			guard++
			switch g.GetPhase() {
			case domain.UltiPhasePlay:
				if g.IsHumanTurn() {
					valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
					require.NoError(t, g.PlayerPlay(valid[0]))
				} else {
					g.CpuPlay()
				}
			case domain.UltiPhaseTrickEnd:
				g.ResolveTrick()
				if g.GetPhase() == domain.UltiPhaseTrickEnd {
					g.NextTrick()
				}
			}
		}
		assert.Equal(t, domain.UltiPhaseRoundEnd, g.GetPhase())
	}
}

func TestUlti_Getters(t *testing.T) {
	g := newTestUlti()
	g.SetRoundNumber(4)
	assert.Equal(t, 4, g.GetRoundNumber())
	g.SetTrickNumber(3)
	assert.Equal(t, 3, g.GetTrickNumber())
	g.SetCurrentPlayerIdx(2)
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	g.SetLeadPlayerIdx(1)
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
	g.SetContract(domain.UltiContractBetli)
	assert.Equal(t, domain.UltiContractBetli, g.GetContract())
	g.SetDeclarerIdx(2)
	assert.Equal(t, 2, g.GetDeclarerIdx())
	g.SetTrumpSuit(domain.CardDesignClover)
	assert.Equal(t, domain.CardDesignClover, g.GetTrumpSuit())
	g.SetPlayerCoins([domain.UltiPlayerCnt]int{10, 20, 30})
	assert.Equal(t, [domain.UltiPlayerCnt]int{10, 20, 30}, g.GetPlayerCoins())

	assert.GreaterOrEqual(t, g.GetDealerIdx(), 0)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetPlayer(0))
	assert.NotNil(t, g.GetConfig())
	assert.Equal(t, domain.UltiResultNone, g.GetResult())
	assert.Equal(t, 0, g.GetCardPoints(-1))
	_ = g.GetActionLog()

	assert.Nil(t, g.GetPlayableIndices(-1))
	g.SetPhase(domain.UltiPhaseBid)
	assert.Nil(t, g.GetPlayableIndices(0), "not play phase -> nil")

	g.SetPhase(domain.UltiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetPhase(domain.UltiPhaseTrickEnd)
	assert.False(t, g.IsHumanTurn())
	assert.False(t, g.IsHumanBidTurn())

	// CpuBid is a no-op.
	g.CpuBid()
}

func TestUlti_JSON_RoundTrip(t *testing.T) {
	g := newTestUlti()
	require.NoError(t, g.PlayerBid(domain.UltiContractParty, domain.CardDesignHeart))
	require.NoError(t, g.PlayerDiscard([]int{0, 1}))
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Ulti
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
	assert.Equal(t, g.GetDeclarerIdx(), g2.GetDeclarerIdx())
	assert.Equal(t, g.GetTrumpSuit(), g2.GetTrumpSuit())
	assert.Equal(t, g.GetContract(), g2.GetContract())
	assert.Equal(t, g.GetDiscardCount(), g2.GetDiscardCount())
}

func TestUlti_JSON_Invalid(t *testing.T) {
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	cases := []string{
		`not json`,
		`{"ph":0,"ps":[null,null],"ts":-1}`,                                           // wrong player count
		`{"ph":0,"ps":[null,null,null],"ts":-1}`,                                      // nil players
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ci":100}`,                            // ci out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ci":-1}`,                             // ci negative
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"di":99}`,                             // di out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"dc":99}`,                             // declarer out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"li":99}`,                             // li out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"li":-2}`,                             // li below -1
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"lt":99}`,                             // lastTrickWinner out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"wp":99}`,                             // winnerPlayer out of range
		`{"ph":99,"ps":` + okPlayers + `,"ts":-1}`,                                    // bad phase
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"co":9}`,                              // bad contract
		`{"ph":0,"ps":` + okPlayers + `,"ts":9}`,                                      // bad trump suit
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"oc":9}`,                              // bad outcome
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"rs":9}`,                              // bad result
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ct":[null]}`,                         // nil trick card
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ct":[{"pi":99,"c":{"d":1,"v":13}}]}`, // trick idx out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"tl":[null]}`,                         // nil talon element
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ds":[null]}`,                         // nil discard element
		`{"ph":2,"ps":` + okPlayers + `,"li":-1,"co":1,"ts":3}`,                       // play requires lead set
		`{"ph":2,"ps":` + okPlayers + `,"li":0,"co":0,"ts":3}`,                        // play requires contract set
		`{"ph":2,"ps":` + okPlayers + `,"li":0,"co":1,"ts":-1}`,                       // party play requires trump
	}
	for _, c := range cases {
		var g domain.Ulti
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}

	// Config validation failure (bad CPU difficulty).
	badCfg := `{"ph":0,"ps":` + okPlayers + `,"ts":-1,"cf":{"cd":99,"tr":5}}`
	var gc domain.Ulti
	assert.Error(t, json.Unmarshal([]byte(badCfg), &gc))

	// Valid restore.
	okJSON := `{"ph":0,"ps":` + okPlayers + `,"co":0,"cf":{"cd":1,"tr":5},"lt":-1,"wp":-1,"li":-1,"dc":0,"ts":-1}`
	var g2 domain.Ulti
	assert.NoError(t, json.Unmarshal([]byte(okJSON), &g2))
	assert.Equal(t, 3, g2.GetPlayerCnt())
	assert.NotNil(t, g2.GetPlayer(0))
}

func TestUltiPlayer_JSON_And_ResetRound(t *testing.T) {
	p := domain.NewUltiPlayer(true)
	p.AddCard(ultiCard(domain.CardDesignSpade, 1))
	p.AddTrick([]*domain.Card{ultiCard(domain.CardDesignHeart, 13)})
	assert.Equal(t, 1, p.GetTrickCount())

	b, err := json.Marshal(p)
	require.NoError(t, err)
	var p2 domain.UltiPlayer
	require.NoError(t, json.Unmarshal(b, &p2))
	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetCardsSize())
	assert.Equal(t, 1, p2.GetTrickCount())

	p2.ResetRound()
	assert.Equal(t, 0, p2.GetCardsSize())
	assert.Equal(t, 0, p2.GetTrickCount())
	assert.False(t, p2.GetIsFinished())

	assert.Error(t, json.Unmarshal([]byte(`not json`), &p2))
	var p3 domain.UltiPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p3))
	assert.False(t, p3.GetIsHuman())
}

func TestUlti_UltiContract(t *testing.T) {
	// driveFinalTrick sets up the 10th (final) trick and resolves it, which sets
	// lastTrickWinner and runs the Ulti round-end scoring.
	driveFinalTrick := func(trump int, trick []*domain.TrickCard) *domain.Ulti {
		g := newTestUlti()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.UltiContractUlti)
		g.SetTrumpSuit(trump)
		g.SetTrickNumber(domain.UltiTrickCount)
		g.SetPhase(domain.UltiPhaseTrickEnd)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		return g
	}

	t.Run("win: declarer takes the final trick with the trump 7", func(t *testing.T) {
		g := driveFinalTrick(domain.CardDesignHeart, []*domain.TrickCard{
			{PlayerIdx: 0, Card: ultiCard(domain.CardDesignHeart, 7)},  // trump 7 leads and wins (only trump)
			{PlayerIdx: 1, Card: ultiCard(domain.CardDesignSpade, 1)},  // off-suit A cannot win
			{PlayerIdx: 2, Card: ultiCard(domain.CardDesignSpade, 10)}, // off-suit 10 cannot win
		})
		assert.Equal(t, domain.UltiPhaseRoundEnd, g.GetPhase())
		assert.Equal(t, domain.UltiOutcomeWin, g.GetOutcome())
		// stake 4 -> declarer +8, each opponent -4.
		assert.Equal(t, [domain.UltiPlayerCnt]int{8, -4, -4}, g.GetPlayerCoins())
	})

	t.Run("loss: final trick won with a trump other than the 7", func(t *testing.T) {
		g := driveFinalTrick(domain.CardDesignHeart, []*domain.TrickCard{
			{PlayerIdx: 0, Card: ultiCard(domain.CardDesignHeart, 1)}, // trump A wins, but it is not the 7
			{PlayerIdx: 1, Card: ultiCard(domain.CardDesignSpade, 1)},
			{PlayerIdx: 2, Card: ultiCard(domain.CardDesignSpade, 10)},
		})
		assert.Equal(t, domain.UltiOutcomeLoss, g.GetOutcome())
		// Double payment on failure: stake 4*2=8 -> declarer -16, each opponent +8.
		assert.Equal(t, [domain.UltiPlayerCnt]int{-16, 8, 8}, g.GetPlayerCoins())
	})

	t.Run("loss: declarer plays the trump 7 but a coalition trump overtakes it", func(t *testing.T) {
		g := driveFinalTrick(domain.CardDesignHeart, []*domain.TrickCard{
			{PlayerIdx: 0, Card: ultiCard(domain.CardDesignHeart, 7)},  // trump 7
			{PlayerIdx: 1, Card: ultiCard(domain.CardDesignHeart, 1)},  // trump A overtakes
			{PlayerIdx: 2, Card: ultiCard(domain.CardDesignSpade, 10)}, // off-suit
		})
		assert.Equal(t, domain.UltiOutcomeLoss, g.GetOutcome())
		assert.Equal(t, [domain.UltiPlayerCnt]int{-16, 8, 8}, g.GetPlayerCoins())
	})

	t.Run("bid requires a trump suit and takes the talon", func(t *testing.T) {
		g := newTestUlti()
		assert.Error(t, g.PlayerBid(domain.UltiContractUlti, -1)) // ulti requires a trump suit
		require.NoError(t, g.PlayerBid(domain.UltiContractUlti, domain.CardDesignSpade))
		assert.Equal(t, domain.UltiContractUlti, g.GetContract())
		assert.Equal(t, domain.CardDesignSpade, g.GetTrumpSuit())
		assert.Equal(t, domain.UltiPhaseDiscard, g.GetPhase())
		assert.True(t, g.GetTalonTaken())
	})

	t.Run("hint recommends ulti with a long trump suit holding the 7", func(t *testing.T) {
		g := newTestUlti()
		setUltiHand(g, 0,
			ultiCard(domain.CardDesignSpade, 7),
			ultiCard(domain.CardDesignSpade, 1),
			ultiCard(domain.CardDesignSpade, 13),
			ultiCard(domain.CardDesignSpade, 12),
			ultiCard(domain.CardDesignSpade, 11),
			ultiCard(domain.CardDesignHeart, 9),
			ultiCard(domain.CardDesignClover, 8),
		)
		hint := g.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "bid_ulti", hint.Reason)
	})
}

func TestUltiConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultUltiConfig().Validate())
	assert.Equal(t, domain.UltiWinRounds, domain.DefaultUltiConfig().TargetRounds)
	assert.Equal(t, domain.UltiCpuDifficultyNormal, domain.DefaultUltiConfig().CpuDifficulty)

	assert.Error(t, domain.UltiConfig{CpuDifficulty: 99, TargetRounds: 5}.Validate())
	assert.Error(t, domain.UltiConfig{CpuDifficulty: domain.UltiCpuDifficultyEasy, TargetRounds: 0}.Validate())
}
