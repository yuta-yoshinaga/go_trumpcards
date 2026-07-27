//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers (cego-prefixed) ---

func cegoTrumpCard(v int) *domain.Card {
	return domain.NewCard(domain.CegoTrumpDesign, v, false)
}

func cegoSkusCard() *domain.Card {
	return domain.NewCard(domain.CegoSkusDesign, domain.CegoSkusValue, false)
}

func cegoSuitCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func cegoKingCard(design int) *domain.Card {
	return domain.NewCard(design, domain.CegoKingValue, false)
}

func cegoNewReset() *domain.Cego {
	g := domain.NewDefaultCego()
	g.Reset()
	return g
}

func cegoSetHand(g *domain.Cego, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- Deck ---

func TestCegoDeckIs54(t *testing.T) {
	deck := domain.BuildCegoDeckPublic()
	require.Len(t, deck, domain.CegoDeckSize)
	suits, trumps, skus := 0, 0, 0
	total := 0
	for _, c := range deck {
		total += domain.CegoCardPointsPublic(c)
		switch c.GetDesign() {
		case domain.CegoSkusDesign:
			skus++
			assert.Equal(t, domain.CegoSkusValue, c.GetValue())
		case domain.CegoTrumpDesign:
			trumps++
			assert.GreaterOrEqual(t, c.GetValue(), 1)
			assert.LessOrEqual(t, c.GetValue(), domain.CegoMaxTrump)
		default:
			suits++
		}
	}
	assert.Equal(t, 32, suits)
	assert.Equal(t, 21, trumps)
	assert.Equal(t, 1, skus)
	// The simplified card-point total must equal exactly the documented deck sum.
	assert.Equal(t, domain.CegoTotalPoints, total)
	assert.Equal(t, 106, total)
}

func TestCegoDealDistribution(t *testing.T) {
	g := cegoNewReset()
	assert.Equal(t, domain.CegoPhaseBid, g.GetPhase())
	assert.Equal(t, domain.CegoBlindSize, g.GetBlindCount())
	for i := 0; i < domain.CegoPlayerCnt; i++ {
		assert.Equal(t, domain.CegoHandSize, g.GetPlayer(i).GetCardsSize())
	}
	// 11*4 + 10 == 54
	assert.Equal(t, domain.CegoDeckSize, domain.CegoHandSize*domain.CegoPlayerCnt+domain.CegoBlindSize)
}

// --- Card classification / points ---

func TestCegoClassification(t *testing.T) {
	assert.True(t, domain.CegoIsTrumpPublic(cegoTrumpCard(5)))
	assert.False(t, domain.CegoIsTrumpPublic(cegoSkusCard()))
	assert.True(t, domain.CegoIsSkusPublic(cegoSkusCard()))
	assert.True(t, domain.CegoIsKingPublic(cegoKingCard(domain.CardDesignHeart)))
	assert.False(t, domain.CegoIsKingPublic(cegoTrumpCard(8)))
	assert.True(t, domain.CegoIsTrullPublic(cegoTrumpCard(domain.CegoPagatValue)))
	assert.True(t, domain.CegoIsTrullPublic(cegoTrumpCard(domain.CegoMaxTrump)))
	assert.True(t, domain.CegoIsTrullPublic(cegoSkusCard()))
	assert.False(t, domain.CegoIsTrullPublic(cegoTrumpCard(10)))
	// nil safety
	assert.False(t, domain.CegoIsTrumpPublic(nil))
	assert.False(t, domain.CegoIsTrullPublic(nil))
	assert.Equal(t, 0, domain.CegoCardPointsPublic(nil))
}

func TestCegoCardPoints(t *testing.T) {
	assert.Equal(t, 5, domain.CegoCardPointsPublic(cegoKingCard(domain.CardDesignSpade)))
	assert.Equal(t, 4, domain.CegoCardPointsPublic(cegoSuitCard(domain.CardDesignSpade, 7)))
	assert.Equal(t, 3, domain.CegoCardPointsPublic(cegoSuitCard(domain.CardDesignSpade, 6)))
	assert.Equal(t, 2, domain.CegoCardPointsPublic(cegoSuitCard(domain.CardDesignSpade, 5)))
	assert.Equal(t, 1, domain.CegoCardPointsPublic(cegoSuitCard(domain.CardDesignSpade, 1)))
	assert.Equal(t, 5, domain.CegoCardPointsPublic(cegoTrumpCard(domain.CegoPagatValue)))
	assert.Equal(t, 5, domain.CegoCardPointsPublic(cegoTrumpCard(domain.CegoMaxTrump)))
	assert.Equal(t, 5, domain.CegoCardPointsPublic(cegoSkusCard()))
	assert.Equal(t, 1, domain.CegoCardPointsPublic(cegoTrumpCard(10)))
}

// --- Scoring helper (pure, zero-sum, 1 vs 3) ---

func TestCegoScoreDealZeroSum(t *testing.T) {
	cases := []int{0, 20, 53, 54, 70, 106}
	for _, pts := range cases {
		bd := domain.CegoScoreDeal(pts, 1)
		// Zero-sum: declarer + 3 opponents == 0.
		total := bd.DeclarerScore + 3*bd.OpponentScore
		assert.Equalf(t, 0, total, "declarerPoints=%d not zero-sum", pts)
		if pts > 53 {
			assert.Truef(t, bd.Won, "pts=%d should win", pts)
			assert.Greater(t, bd.DeclarerScore, 0)
			assert.Less(t, bd.OpponentScore, 0)
		} else {
			assert.Falsef(t, bd.Won, "pts=%d should lose", pts)
			assert.Less(t, bd.DeclarerScore, 0)
			assert.Greater(t, bd.OpponentScore, 0)
		}
		// Declarer swing is 3x an opponent's swing in magnitude.
		assert.Equal(t, 3*(-bd.OpponentScore), bd.DeclarerScore)
	}
}

func TestCegoScoreDealThreshold(t *testing.T) {
	assert.False(t, domain.CegoScoreDeal(53, 1).Won)
	assert.True(t, domain.CegoScoreDeal(54, 1).Won)
	assert.Equal(t, 53, domain.CegoScoreDeal(54, 1).Threshold)
	assert.Equal(t, 1, domain.CegoBidMultPublic(domain.CegoBidPlay))
}

// --- Trick logic: trump priority + Sküs highest ---

func TestCegoTrickWinnerTrumpPriority(t *testing.T) {
	g := domain.NewDefaultCego()
	g.SetPhase(domain.CegoPhasePlay)
	// Led spade king (high), then a low trump beats it.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: cegoKingCard(domain.CardDesignSpade)},
		{PlayerIdx: 1, Card: cegoTrumpCard(2)},
		{PlayerIdx: 2, Card: cegoSuitCard(domain.CardDesignSpade, 7)},
		{PlayerIdx: 3, Card: cegoSuitCard(domain.CardDesignSpade, 1)},
	})
	assert.Equal(t, 1, g.TrickWinnerPublic())
}

func TestCegoTrickWinnerSkusHighest(t *testing.T) {
	g := domain.NewDefaultCego()
	g.SetPhase(domain.CegoPhasePlay)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: cegoTrumpCard(21)},
		{PlayerIdx: 1, Card: cegoTrumpCard(1)},
		{PlayerIdx: 2, Card: cegoSkusCard()},
		{PlayerIdx: 3, Card: cegoTrumpCard(20)},
	})
	assert.Equal(t, 2, g.TrickWinnerPublic())
	assert.Equal(t, domain.CegoTrumpDesign, g.LedSuitPublic())
}

func TestCegoTrickWinnerHighestOfLedSuit(t *testing.T) {
	g := domain.NewDefaultCego()
	g.SetPhase(domain.CegoPhasePlay)
	// No trumps -> highest of led suit (spade) wins.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: cegoSuitCard(domain.CardDesignSpade, 3)},
		{PlayerIdx: 1, Card: cegoKingCard(domain.CardDesignSpade)},
		{PlayerIdx: 2, Card: cegoSuitCard(domain.CardDesignHeart, 8)}, // off-suit
		{PlayerIdx: 3, Card: cegoSuitCard(domain.CardDesignSpade, 6)},
	})
	assert.Equal(t, 1, g.TrickWinnerPublic())
}

// --- Bidding ---

func TestCegoBiddingFirstBidWins(t *testing.T) {
	g := cegoNewReset()
	g.SetBidPlayerIdx(0)
	// Force human seat 0 turn.
	require.NoError(t, g.PlayerBid(domain.CegoBidPlay))
	// Others cannot outbid a single-level auction.
	for g.GetPhase() == domain.CegoPhaseBid {
		g.CpuBid()
	}
	assert.Equal(t, 0, g.GetDeclarerIdx())
	assert.Equal(t, domain.CegoBidPlay, g.GetContract())
	assert.Equal(t, domain.CegoPhaseContract, g.GetPhase())
}

func TestCegoBiddingAllPassForcesNeighbour(t *testing.T) {
	// All players CPU so they can all pass regardless of hand.
	players := make([]*domain.CegoPlayer, domain.CegoPlayerCnt)
	for i := range players {
		players[i] = domain.NewCegoPlayer(false)
	}
	g := domain.NewCego(players, domain.DefaultCegoConfig())
	g.Reset()
	// Clear hands so evalHand=0 -> every CPU passes.
	for i := 0; i < domain.CegoPlayerCnt; i++ {
		g.GetPlayer(i).Reset()
	}
	for g.GetPhase() == domain.CegoPhaseBid {
		g.CpuBid()
	}
	// dealer=0 -> neighbour = 1 forced declarer.
	assert.Equal(t, (g.GetDealerIdx()+1)%domain.CegoPlayerCnt, g.GetDeclarerIdx())
	assert.Equal(t, domain.CegoPhaseContract, g.GetPhase())
}

// --- Contract: Cego exchange ---

func TestCegoExchangeCegoKeepsOne(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhaseContract)
	// Human declarer chooses Cego.
	require.NoError(t, g.PlayerChooseContract(domain.CegoContractCego))
	assert.Equal(t, domain.CegoPhaseExchange, g.GetPhase())
	blindBefore := g.GetBlindCount()
	require.Equal(t, domain.CegoBlindSize, blindBefore)

	require.NoError(t, g.PlayerDiscard([]int{0}))
	// Hand is 1 kept + 10 blind = 11.
	assert.Equal(t, domain.CegoHandSize, g.GetPlayer(0).GetCardsSize())
	// Laid-down 10 go to declarer stash.
	assert.Equal(t, domain.CegoLayDownCount, len(g.GetStash()))
	assert.Equal(t, 0, g.GetStashOwner())
	assert.Equal(t, 0, g.GetBlindCount())
	assert.Equal(t, domain.CegoPhasePlay, g.GetPhase())
}

func TestCegoExchangeInvalidKeepCount(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	g.SetContractType(domain.CegoContractCego)
	g.SetPhase(domain.CegoPhaseExchange)
	assert.Error(t, g.PlayerDiscard([]int{0, 1})) // too many
	assert.Error(t, g.PlayerDiscard([]int{}))     // too few
	assert.Error(t, g.PlayerDiscard([]int{999}))  // out of range
}

// --- Contract: Handspiel ---

func TestCegoHandspielBlindToOpponents(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhaseContract)
	require.NoError(t, g.PlayerChooseContract(domain.CegoContractHandspiel))
	// No exchange: declarer keeps 11 cards, blind goes to opponent stash.
	assert.Equal(t, domain.CegoHandSize, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 1, g.GetStashOwner())
	assert.Equal(t, domain.CegoBlindSize, len(g.GetStash()))
	assert.Equal(t, 0, g.GetBlindCount())
	assert.Equal(t, domain.CegoPhasePlay, g.GetPhase())
}

// --- Captured points conservation ---

func TestCegoCapturedTotalsDeck(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(0)
	g.SetContract(domain.CegoBidPlay)
	g.SetContractType(domain.CegoContractHandspiel)
	// Give all 54 cards out: 11 tricks worth to players + blind stash.
	deck := domain.BuildCegoDeckPublic()
	// Opponent stash = first 10 cards.
	g.SetStash(deck[:domain.CegoBlindSize])
	g.SetStashOwner(1)
	// Distribute remaining 44 cards as tricks across the 4 players (11 each).
	rest := deck[domain.CegoBlindSize:]
	for i := 0; i < domain.CegoPlayerCnt; i++ {
		trick := rest[i*domain.CegoHandSize : (i+1)*domain.CegoHandSize]
		g.GetPlayer(i).AddTrick(trick)
	}
	total := 0
	for i := 0; i < domain.CegoPlayerCnt; i++ {
		total += g.GetCardPoints(i)
	}
	for _, c := range g.GetStash() {
		total += domain.CegoCardPointsPublic(c)
	}
	assert.Equal(t, domain.CegoTotalPoints, total)
}

// --- Full game drive to end (allow tie winnerPlayer == -1) ---

func TestCegoFullGameDrive(t *testing.T) {
	g := domain.NewDefaultCego()
	g.SetConfig(domain.CegoConfig{CpuDifficulty: domain.CegoCpuDifficultyNormal, TargetDeals: 2})
	g.Reset()
	guard := 0
	for !g.GetGameEndFlag() && guard < 10000 {
		guard++
		switch g.GetPhase() {
		case domain.CegoPhaseBid:
			if g.IsHumanBidTurn() {
				if err := g.PlayerBid(domain.CegoBidPlay); err != nil {
					_ = g.PlayerPass()
				}
			} else {
				g.CpuBid()
			}
		case domain.CegoPhaseContract:
			if g.IsHumanContractTurn() {
				require.NoError(t, g.PlayerChooseContract(domain.CegoContractCego))
			} else {
				g.CpuChooseContract()
			}
		case domain.CegoPhaseExchange:
			if g.IsHumanExchangeTurn() {
				require.NoError(t, g.PlayerDiscard([]int{0}))
			} else {
				g.CpuDiscard()
			}
		case domain.CegoPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(idx)
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
		case domain.CegoPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == domain.CegoPhaseTrickEnd {
				g.NextTrick()
			}
		case domain.CegoPhaseRoundEnd:
			g.ScoreRound()
			g.NextRound()
		}
	}
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.CegoPhaseGameEnd, g.GetPhase())
	// Winner may be -1 on a tie.
	assert.GreaterOrEqual(t, g.GetWinnerPlayer(), -1)
}

// --- CPU AI sanity ---

func TestCegoCpuContractChoice(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(1) // a CPU
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhaseContract)
	// Strong hand -> Handspiel.
	cegoSetHand(g, 1,
		cegoSkusCard(), cegoTrumpCard(21), cegoTrumpCard(20), cegoTrumpCard(19),
		cegoTrumpCard(18), cegoTrumpCard(17), cegoTrumpCard(16), cegoTrumpCard(15),
		cegoKingCard(domain.CardDesignSpade), cegoKingCard(domain.CardDesignHeart),
		cegoKingCard(domain.CardDesignDiamond))
	g.CpuChooseContract()
	assert.Equal(t, domain.CegoContractHandspiel, g.GetContractType())
}

func TestCegoCpuContractChoiceWeak(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetDeclarerIdx(1)
	g.SetContract(domain.CegoBidPlay)
	g.SetPhase(domain.CegoPhaseContract)
	cegoSetHand(g, 1,
		cegoSuitCard(domain.CardDesignSpade, 1), cegoSuitCard(domain.CardDesignSpade, 2),
		cegoSuitCard(domain.CardDesignHeart, 1), cegoSuitCard(domain.CardDesignHeart, 2),
		cegoSuitCard(domain.CardDesignDiamond, 1), cegoSuitCard(domain.CardDesignDiamond, 2),
		cegoSuitCard(domain.CardDesignClover, 1), cegoSuitCard(domain.CardDesignClover, 2),
		cegoSuitCard(domain.CardDesignClover, 3), cegoSuitCard(domain.CardDesignClover, 4),
		cegoTrumpCard(3))
	g.CpuChooseContract()
	assert.Equal(t, domain.CegoContractCego, g.GetContractType())
}

// --- Hint ---

func TestCegoHintPhases(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetBidPlayerIdx(0)
	h := g.GetHint()
	require.NotNil(t, h)
	// Contract-phase hint for human declarer.
	g.SetDeclarerIdx(0)
	g.SetPhase(domain.CegoPhaseContract)
	hc := g.GetHint()
	require.NotNil(t, hc)
	assert.NotNil(t, hc.Contract)
}

// --- JSON round-trip + validation ---

func TestCegoJSONRoundTrip(t *testing.T) {
	g := cegoNewReset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var g2 domain.Cego
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetBlindCount(), g2.GetBlindCount())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
}

func TestCegoUnmarshalErrors(t *testing.T) {
	base := func() map[string]any {
		g := cegoNewReset()
		data, _ := json.Marshal(g)
		var m map[string]any
		_ = json.Unmarshal(data, &m)
		return m
	}
	t.Run("bad player count", func(t *testing.T) {
		m := base()
		m["ps"] = []any{}
		data, _ := json.Marshal(m)
		var g domain.Cego
		assert.Error(t, json.Unmarshal(data, &g))
	})
	t.Run("bad phase", func(t *testing.T) {
		m := base()
		m["ph"] = 99
		data, _ := json.Marshal(m)
		var g domain.Cego
		assert.Error(t, json.Unmarshal(data, &g))
	})
	t.Run("bad stashOwner", func(t *testing.T) {
		m := base()
		m["so"] = 5
		data, _ := json.Marshal(m)
		var g domain.Cego
		assert.Error(t, json.Unmarshal(data, &g))
	})
	t.Run("bad contractType", func(t *testing.T) {
		m := base()
		m["cn"] = 9
		data, _ := json.Marshal(m)
		var g domain.Cego
		assert.Error(t, json.Unmarshal(data, &g))
	})
	t.Run("bad bid", func(t *testing.T) {
		m := base()
		m["hb"] = 9
		data, _ := json.Marshal(m)
		var g domain.Cego
		assert.Error(t, json.Unmarshal(data, &g))
	})
	t.Run("nil card in deck", func(t *testing.T) {
		m := base()
		m["dk"] = []any{nil}
		data, _ := json.Marshal(m)
		var g domain.Cego
		assert.Error(t, json.Unmarshal(data, &g))
	})
}

// --- Accessors ---

func TestCegoAccessors(t *testing.T) {
	g := domain.NewDefaultCego()
	g.Reset()
	g.SetRoundNumber(3)
	assert.Equal(t, 3, g.GetRoundNumber())
	g.SetTrickNumber(4)
	assert.Equal(t, 4, g.GetTrickNumber())
	g.SetDealerIdx(2)
	assert.Equal(t, 2, g.GetDealerIdx())
	g.SetContractType(domain.CegoContractCego)
	assert.Equal(t, domain.CegoContractCego, g.GetContractType())
	g.SetPlayerScores([domain.CegoPlayerCnt]int{1, 2, 3, 4})
	assert.Equal(t, [domain.CegoPlayerCnt]int{1, 2, 3, 4}, g.GetPlayerScores())
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetActionLog())
	assert.Equal(t, domain.CegoResultNone, g.GetResult())
}
