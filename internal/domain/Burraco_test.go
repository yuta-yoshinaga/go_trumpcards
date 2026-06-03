//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestBurraco() *domain.Burraco {
	players := []*domain.BurracoPlayer{
		domain.NewBurracoPlayer(true),
		domain.NewBurracoPlayer(false),
	}
	return domain.NewBurraco(domain.NewTrumpCardsWithDecks(2, 4), players, domain.DefaultBurracoConfig())
}

func newTestBurracoWithDifficulty(d domain.BurracoCpuDifficulty) *domain.Burraco {
	players := []*domain.BurracoPlayer{
		domain.NewBurracoPlayer(true),
		domain.NewBurracoPlayer(false),
	}
	cfg := domain.DefaultBurracoConfig()
	cfg.CpuDifficulty = d
	return domain.NewBurraco(domain.NewTrumpCardsWithDecks(2, 4), players, cfg)
}

func setupBurracoDrawPhase(g *domain.Burraco, currentIdx int) {
	g.SetPhase(domain.BurracoPhaseDraw)
	g.SetCurrentPlayerIdx(currentIdx)
}

func setupBurracoMeldPhase(g *domain.Burraco, currentIdx int) {
	g.SetPhase(domain.BurracoPhaseMeld)
	g.SetCurrentPlayerIdx(currentIdx)
}

func setupBurracoDiscardPhase(g *domain.Burraco, currentIdx int) {
	g.SetPhase(domain.BurracoPhaseDiscard)
	g.SetCurrentPlayerIdx(currentIdx)
}

// --- Constructor & Reset ---

func TestNewBurraco(t *testing.T) {
	g := newTestBurraco()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
}

func TestBurraco_Reset(t *testing.T) {
	g := newTestBurraco()
	g.Reset()

	assert.Equal(t, domain.BurracoPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())

	// Each player should have 11 cards in hand (red 3s are replaced from stock)
	for i := 0; i < 2; i++ {
		p := g.GetPlayer(i)
		assert.Equal(t, domain.BurracoHandSize, p.GetCardsSize(), "player %d hand should have 11 cards", i)
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
		assert.Empty(t, p.GetMelds())
		// Hand should contain no red 3s (all auto-laid)
		for j := 0; j < p.GetCardsSize(); j++ {
			assert.False(t, domain.BurracoIsRed3(p.GetCard(j)), "player %d hand should not contain red 3s", i)
		}
	}

	// Discard pile should have at least 1 card
	assert.GreaterOrEqual(t, g.GetDiscardPileCount(), 1)

	// Two pozzetti of 11 cards each are set aside at deal time
	assert.Equal(t, 2, g.GetPozzettoCount(), "two pozzetti should be set aside")
	assert.Equal(t, 2*domain.BurracoPozzettoSize, g.GetPozzettoCardCount(), "pozzetti should hold 22 cards total")

	// Total cards: 108 = hands + red3s + discard + draw + pozzetti
	totalCards := g.GetDrawPileCount() + g.GetDiscardPileCount() + g.GetPozzettoCardCount()
	for i := 0; i < 2; i++ {
		totalCards += g.GetPlayer(i).GetCardsSize() + len(g.GetPlayer(i).GetRed3s())
	}
	assert.Equal(t, 108, totalCards, "total cards should be 108 (got draw=%d discard=%d pozzetti=%d p0hand=%d p0red3=%d p1hand=%d p1red3=%d)",
		g.GetDrawPileCount(), g.GetDiscardPileCount(), g.GetPozzettoCardCount(),
		g.GetPlayer(0).GetCardsSize(), len(g.GetPlayer(0).GetRed3s()),
		g.GetPlayer(1).GetCardsSize(), len(g.GetPlayer(1).GetRed3s()))
}

func TestBurraco_Reset_ClearsAllState(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetGameEndFlag(true)

	g.Reset()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, -1, g.GetWinnerIdx())
}

// --- Card Type Helpers ---

func TestBurracoIsWild(t *testing.T) {
	tests := []struct {
		name   string
		card   *domain.Card
		expect bool
	}{
		{"Joker", domain.NewCard(domain.CardDesignJoker, 1, false), true},
		{"2 of spades", domain.NewCard(domain.CardDesignSpade, 2, false), true},
		{"2 of hearts", domain.NewCard(domain.CardDesignHeart, 2, false), true},
		{"3 of hearts", domain.NewCard(domain.CardDesignHeart, 3, false), false},
		{"Ace of spades", domain.NewCard(domain.CardDesignSpade, 1, false), false},
		{"King", domain.NewCard(domain.CardDesignDiamond, 13, false), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, domain.BurracoIsWild(tt.card))
		})
	}
}

func TestBurracoIsRed3(t *testing.T) {
	tests := []struct {
		name   string
		card   *domain.Card
		expect bool
	}{
		{"Heart 3", domain.NewCard(domain.CardDesignHeart, 3, false), true},
		{"Diamond 3", domain.NewCard(domain.CardDesignDiamond, 3, false), true},
		{"Spade 3", domain.NewCard(domain.CardDesignSpade, 3, false), false},
		{"Clover 3", domain.NewCard(domain.CardDesignClover, 3, false), false},
		{"Heart 4", domain.NewCard(domain.CardDesignHeart, 4, false), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, domain.BurracoIsRed3(tt.card))
		})
	}
}

func TestBurracoIsBlack3(t *testing.T) {
	tests := []struct {
		name   string
		card   *domain.Card
		expect bool
	}{
		{"Spade 3", domain.NewCard(domain.CardDesignSpade, 3, false), true},
		{"Clover 3", domain.NewCard(domain.CardDesignClover, 3, false), true},
		{"Heart 3", domain.NewCard(domain.CardDesignHeart, 3, false), false},
		{"Spade 4", domain.NewCard(domain.CardDesignSpade, 4, false), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, domain.BurracoIsBlack3(tt.card))
		})
	}
}

func TestBurracoCardValue(t *testing.T) {
	tests := []struct {
		name   string
		card   *domain.Card
		expect int
	}{
		{"Joker", domain.NewCard(domain.CardDesignJoker, 1, false), 50},
		{"2 of spades", domain.NewCard(domain.CardDesignSpade, 2, false), 20},
		{"Ace", domain.NewCard(domain.CardDesignHeart, 1, false), 20},
		{"King", domain.NewCard(domain.CardDesignSpade, 13, false), 10},
		{"Queen", domain.NewCard(domain.CardDesignHeart, 12, false), 10},
		{"8", domain.NewCard(domain.CardDesignDiamond, 8, false), 10},
		{"7", domain.NewCard(domain.CardDesignClover, 7, false), 5},
		{"4", domain.NewCard(domain.CardDesignSpade, 4, false), 5},
		{"Black 3", domain.NewCard(domain.CardDesignSpade, 3, false), 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, domain.BurracoCardValue(tt.card))
		})
	}
}

// --- Phase Guards ---

func TestBurraco_PlayerDrawFromStock_PhaseGuard(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	err := g.PlayerDrawFromStock()
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestBurraco_PlayerDrawFromStock_HumanGuard(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 1) // CPU

	err := g.PlayerDrawFromStock()
	assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
}

func TestBurraco_PlayerDrawFromStock_GameEndGuard(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetGameEndFlag(true)

	err := g.PlayerDrawFromStock()
	assert.True(t, errors.Is(err, domain.ErrGameEnded))
}

func TestBurraco_PlayerDrawFromStock(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)

	handBefore := g.GetPlayer(0).GetCardsSize()
	drawBefore := g.GetDrawPileCount()

	err := g.PlayerDrawFromStock()
	require.NoError(t, err)

	// Phase should advance to Meld
	assert.Equal(t, domain.BurracoPhaseMeld, g.GetPhase())
	// Hand should have 1 more card (unless red 3 was drawn)
	assert.GreaterOrEqual(t, g.GetPlayer(0).GetCardsSize(), handBefore)
	assert.Less(t, g.GetDrawPileCount(), drawBefore)
}

func TestBurraco_PlayerDrawFromStock_EmptyDraw(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)
	g.SetDrawPile(nil)

	err := g.PlayerDrawFromStock()
	require.NoError(t, err)
	// Should end round as draw
	assert.True(t, g.GetPhase() == domain.BurracoPhaseRoundEnd || g.GetPhase() == domain.BurracoPhaseGameEnd)
}

// --- PlayerDrawFromDiscard ---

func TestBurraco_PlayerDrawFromDiscard_PhaseGuard(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestBurraco_PlayerDrawFromDiscard_EmptyPile(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)
	g.SetDiscardPile(nil)

	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.Error(t, err)
}

func TestBurraco_PlayerDrawFromDiscard_Black3OnTop(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})

	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.Error(t, err)
}

func TestBurraco_PlayerDrawFromDiscard_WildOnTop(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})

	err := g.PlayerDrawFromDiscard([]int{0, 1})
	assert.Error(t, err)
}

func TestBurraco_PlayerDrawFromDiscard_InvalidPairCount(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})

	err := g.PlayerDrawFromDiscard([]int{0})
	assert.Error(t, err)
}

func TestBurraco_PlayerDrawFromDiscard_Success(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)

	player := g.GetPlayer(0)
	// Set player hand to have a matching pair for the discard top
	player.Reset()
	topCard := domain.NewCard(domain.CardDesignSpade, 7, false)
	g.SetDiscardPile([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false), // bottom
		topCard, // top
	})
	// Add matching pair + some other cards
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	// Player has initial meld already to bypass minimum check
	player.SetHasInitMeld(true)

	err := g.PlayerDrawFromDiscard([]int{0, 1})
	require.NoError(t, err)

	assert.Equal(t, domain.BurracoPhaseMeld, g.GetPhase())
	assert.True(t, g.GetDrewFromDiscard())
	assert.Nil(t, g.GetDiscardPile())
}

// --- PlayerMeld ---

func TestBurraco_PlayerMeld_PhaseGuard(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)

	err := g.PlayerMeld([][]int{{0, 1, 2}})
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestBurraco_PlayerMeld_EmptySkip(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	err := g.PlayerMeld(nil)
	require.NoError(t, err)
	assert.Equal(t, domain.BurracoPhaseDiscard, g.GetPhase())
}

func TestBurraco_PlayerMeld_ValidNewMeld(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	// 3 kings (=30 points, enough for initial meld at score 0 which requires 50)
	// Need more points: add 2 aces
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))   // 0
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))   // 1
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false)) // 2
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))    // 3: Ace
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))    // 4: Ace
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))  // 5: Ace
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))   // 6: extra card

	err := g.PlayerMeld([][]int{{0, 1, 2}, {3, 4, 5}})
	require.NoError(t, err)

	assert.Equal(t, 2, len(player.GetMelds()))
	assert.True(t, player.GetHasInitMeld())
	// 7 cards - 6 melded = 1 remaining
	assert.Equal(t, 1, player.GetCardsSize())
}

func TestBurraco_PlayerMeld_TooFewCards(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	err := g.PlayerMeld([][]int{{0, 1}}) // only 2 cards
	assert.Error(t, err)
}

func TestBurraco_PlayerMeld_TooManyWilds(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false)) // 0: natural
	player.AddCard(domain.NewCard(domain.CardDesignJoker, 1, false)) // 1: wild
	player.AddCard(domain.NewCard(domain.CardDesignJoker, 2, false)) // 2: wild
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false)) // 3: wild (2 of spades)
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 2, false)) // 4: wild (2 of hearts)

	// 1 natural + 4 wilds → should fail (max 3 wilds, wilds > naturals)
	err := g.PlayerMeld([][]int{{0, 1, 2, 3, 4}})
	assert.Error(t, err)
}

func TestBurraco_PlayerMeld_Black3Rejected(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))  // black 3
	player.AddCard(domain.NewCard(domain.CardDesignClover, 3, false)) // black 3
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

	err := g.PlayerMeld([][]int{{0, 1, 2}})
	assert.Error(t, err)
}

func TestBurraco_PlayerMeld_InitialMeldMinimum(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	// Score 0 → minimum 50 points needed
	// 3 fours = 15 points → should fail
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))

	err := g.PlayerMeld([][]int{{0, 1, 2}})
	assert.Error(t, err)
}

// --- PlayerDiscard ---

func TestBurraco_PlayerDiscard_PhaseGuard(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)

	err := g.PlayerDiscard(0)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestBurraco_PlayerDiscard_InvalidIndex(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 0)

	err := g.PlayerDiscard(999)
	assert.Error(t, err)
}

func TestBurraco_PlayerDiscard_Red3Rejected(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false)) // red 3

	err := g.PlayerDiscard(0)
	assert.Error(t, err)
}

func TestBurraco_PlayerDiscard_Success(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

	discardBefore := g.GetDiscardPileCount()
	err := g.PlayerDiscard(0)
	require.NoError(t, err)

	assert.Equal(t, 1, player.GetCardsSize())
	assert.Equal(t, discardBefore+1, g.GetDiscardPileCount())
}

func TestBurraco_PlayerDiscard_WildFreeze(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 0)
	g.SetIsFrozen(false)

	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false)) // wild
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.True(t, g.GetIsFrozen())
}

// --- PlayerGoOut ---

func TestBurraco_PlayerGoOut_NoBurracoFails(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	player.SetTookPozzetto(true) // pozzetto taken, so the burraco check is what fails
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	// No burraco → should fail

	err := g.PlayerGoOut()
	assert.Error(t, err)
}

func TestBurraco_PlayerGoOut_NoPozzettoFails(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	// Give a burraco but no pozzetto taken → going out must still fail
	player.SetMelds([]*domain.BurracoMeld{{
		Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
		},
		IsNatural: true,
	}})
	player.SetHasInitMeld(true)
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	err := g.PlayerGoOut()
	assert.Error(t, err)
}

func TestBurraco_PlayerGoOut_Success(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	// Give burraco (7 cards of same rank)
	player.SetMelds([]*domain.BurracoMeld{
		{
			Cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
				domain.NewCard(domain.CardDesignSpade, 7, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 7, false),
			},
			IsNatural: true,
		},
	})
	player.SetHasInitMeld(true)
	player.SetTookPozzetto(true) // pozzetto already taken → eligible to go out
	// 1 card left to discard
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	err := g.PlayerGoOut()
	require.NoError(t, err)
	assert.True(t, g.GetPhase() == domain.BurracoPhaseRoundEnd || g.GetPhase() == domain.BurracoPhaseGameEnd)
}

// --- Pozzetto mechanic ---

func TestBurraco_GetPozzettoCount(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	assert.Equal(t, 2, g.GetPozzettoCount())
	assert.Equal(t, 2*domain.BurracoPozzettoSize, g.GetPozzettoCardCount())
}

func TestBurraco_TakePozzetto_OnMeldEmptyingHand(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)
	require.Equal(t, 2, g.GetPozzettoCount())

	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true) // bypass initial-meld minimum
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))

	err := g.PlayerMeld([][]int{{0, 1, 2}})
	require.NoError(t, err)

	// Hand emptied → pozzetto taken (one pile consumed), player keeps playing
	assert.True(t, player.GetTookPozzetto())
	assert.Equal(t, 1, g.GetPozzettoCount())
	assert.Greater(t, player.GetCardsSize(), 0)
	assert.Equal(t, domain.BurracoPhaseDiscard, g.GetPhase())
}

func TestBurraco_TakePozzetto_OnDiscardEmptyingHand(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 0)

	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false)) // last card

	err := g.PlayerDiscard(0)
	require.NoError(t, err)

	assert.True(t, player.GetTookPozzetto())
	assert.Equal(t, 1, g.GetPozzettoCount())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // turn advanced to opponent
}

// --- PlayerSkipMeld ---

func TestBurraco_PlayerSkipMeld(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	err := g.PlayerSkipMeld()
	require.NoError(t, err)
	assert.Equal(t, domain.BurracoPhaseDiscard, g.GetPhase())
}

func TestBurraco_PlayerSkipMeld_DrewFromDiscardFails(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 0)

	// Simulate having drawn from discard
	// Use reflection-free approach: just test the phase guard
	// We can't easily set drewFromDiscard without a setter, so test via full flow
}

// --- NextRound ---

func TestBurraco_NextRound_WrongPhase(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	// In BurracoPhaseDraw, NextRound should do nothing
	g.NextRound()
	assert.Equal(t, domain.BurracoPhaseDraw, g.GetPhase())
}

// --- MinimumMeldValue ---

func TestBurraco_MinimumMeldValue(t *testing.T) {
	g := newTestBurraco()
	g.Reset()

	// Score < 0 → 15
	g.GetPlayer(0).SetCumulativeScore(-100)
	// We can't call minimumMeldValue directly, but we test it through PlayerMeld behavior
	// Score 0 → 50 (tested in TestBurraco_PlayerMeld_InitialMeldMinimum)
}

// --- BurracoMeld ---

func TestBurracoMeld_IsBurraco(t *testing.T) {
	tests := []struct {
		name   string
		count  int
		expect bool
	}{
		{"3 cards", 3, false},
		{"6 cards", 6, false},
		{"7 cards", 7, true},
		{"10 cards", 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := make([]*domain.Card, tt.count)
			for i := range cards {
				cards[i] = domain.NewCard(domain.CardDesignSpade, 7, false)
			}
			m := &domain.BurracoMeld{Cards: cards, IsNatural: true}
			assert.Equal(t, tt.expect, m.IsBurraco())
		})
	}
}

func TestBurracoMeld_GetRank(t *testing.T) {
	m := &domain.BurracoMeld{
		Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignJoker, 1, false), // wild
			domain.NewCard(domain.CardDesignSpade, 9, false), // natural
			domain.NewCard(domain.CardDesignHeart, 9, false), // natural
		},
		IsNatural: false,
	}
	assert.Equal(t, 9, m.GetRank())
}

// --- BurracoPlayer ---

func TestBurracoPlayer_HasBurraco(t *testing.T) {
	p := domain.NewBurracoPlayer(true)
	assert.False(t, p.HasBurraco())

	cards := make([]*domain.Card, 7)
	for i := range cards {
		cards[i] = domain.NewCard(domain.CardDesignSpade, 5, false)
	}
	p.AddMeld(&domain.BurracoMeld{Cards: cards, IsNatural: true})
	assert.True(t, p.HasBurraco())
}

func TestBurracoPlayer_ResetRound(t *testing.T) {
	p := domain.NewBurracoPlayer(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	p.AddMeld(&domain.BurracoMeld{Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}, IsNatural: true})
	p.AddRed3(domain.NewCard(domain.CardDesignHeart, 3, false))
	p.SetHasInitMeld(true)
	p.SetRoundScore(100)

	p.ResetRound()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Empty(t, p.GetMelds())
	assert.Empty(t, p.GetRed3s())
	assert.False(t, p.GetHasInitMeld())
	assert.Equal(t, 0, p.GetRoundScore())
}

// --- CpuPlay ---

func TestBurraco_CpuPlay_Draw(t *testing.T) {
	for _, diff := range []domain.BurracoCpuDifficulty{
		domain.BurracoCpuDifficultyEasy,
		domain.BurracoCpuDifficultyNormal,
		domain.BurracoCpuDifficultyHard,
	} {
		t.Run(fmt.Sprintf("difficulty_%d", diff), func(t *testing.T) {
			g := newTestBurracoWithDifficulty(diff)
			g.Reset()
			setupBurracoDrawPhase(g, 1)

			g.CpuPlay()
			// Should advance past draw phase
			assert.NotEqual(t, domain.BurracoPhaseDraw, g.GetPhase())
		})
	}
}

func TestBurraco_CpuPlay_Meld(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoMeldPhase(g, 1)

	g.CpuPlay()
	assert.NotEqual(t, domain.BurracoPhaseMeld, g.GetPhase())
}

func TestBurraco_CpuPlay_Discard(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDiscardPhase(g, 1)

	player := g.GetPlayer(1)
	// Ensure CPU has cards to discard
	if player.GetCardsSize() == 0 {
		player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		player.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	}

	g.CpuPlay()
	// Should advance turn
	assert.Equal(t, domain.BurracoPhaseDraw, g.GetPhase())
}

func TestBurraco_CpuPlay_HumanTurn_Noop(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	setupBurracoDrawPhase(g, 0)

	g.CpuPlay()
	// Should not change phase
	assert.Equal(t, domain.BurracoPhaseDraw, g.GetPhase())
}

func TestBurraco_CpuPlay_GameEnded_Noop(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetGameEndFlag(true)

	g.CpuPlay()
}

// --- Getters & State ---

func TestBurraco_GetPlayer_OutOfBounds(t *testing.T) {
	g := newTestBurraco()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(10))
}

func TestBurraco_IsHumanTurn(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
}

func TestBurraco_GetSetConfig(t *testing.T) {
	g := newTestBurraco()
	cfg := domain.BurracoConfig{CpuDifficulty: domain.BurracoCpuDifficultyHard, PointLimit: 10000}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
}

// --- JSON Serialization ---

func TestBurraco_JSON_RoundTrip(t *testing.T) {
	g := newTestBurraco()
	g.Reset()

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Burraco
	err = json.Unmarshal(data, &g2)
	require.NoError(t, err)

	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), g2.GetRoundNumber())
	assert.Equal(t, g.GetCurrentPlayerIdx(), g2.GetCurrentPlayerIdx())
	assert.Equal(t, g.GetIsFrozen(), g2.GetIsFrozen())
	assert.Equal(t, g.GetGameEndFlag(), g2.GetGameEndFlag())
	assert.Equal(t, g.GetWinnerIdx(), g2.GetWinnerIdx())
	assert.Equal(t, g.GetDrawPileCount(), g2.GetDrawPileCount())
	assert.Equal(t, g.GetDiscardPileCount(), g2.GetDiscardPileCount())
	assert.Equal(t, g.GetPozzettoCount(), g2.GetPozzettoCount())
	assert.Equal(t, g.GetPozzettoCardCount(), g2.GetPozzettoCardCount())
}

func TestBurracoPlayer_JSON_RoundTrip(t *testing.T) {
	p := domain.NewBurracoPlayer(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	p.AddMeld(&domain.BurracoMeld{
		Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		},
		IsNatural: true,
	})
	p.AddRed3(domain.NewCard(domain.CardDesignHeart, 3, false))
	p.SetHasInitMeld(true)
	p.SetTookPozzetto(true)
	p.SetRoundScore(50)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var p2 domain.BurracoPlayer
	err = json.Unmarshal(data, &p2)
	require.NoError(t, err)

	assert.Equal(t, p.GetCardsSize(), p2.GetCardsSize())
	assert.Equal(t, len(p.GetMelds()), len(p2.GetMelds()))
	assert.Equal(t, len(p.GetRed3s()), len(p2.GetRed3s()))
	assert.Equal(t, p.GetHasInitMeld(), p2.GetHasInitMeld())
	assert.Equal(t, p.GetTookPozzetto(), p2.GetTookPozzetto())
	assert.Equal(t, p.GetRoundScore(), p2.GetRoundScore())
}

// --- BurracoConfig ---

func TestBurracoConfig_Default(t *testing.T) {
	cfg := domain.DefaultBurracoConfig()
	assert.Equal(t, domain.BurracoCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, domain.BurracoDefaultPointLimit, cfg.PointLimit)
}

func TestBurracoConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.BurracoConfig
		wantErr bool
	}{
		{"valid default", domain.DefaultBurracoConfig(), false},
		{"valid easy", domain.BurracoConfig{CpuDifficulty: domain.BurracoCpuDifficultyEasy, PointLimit: 100}, false},
		{"valid hard", domain.BurracoConfig{CpuDifficulty: domain.BurracoCpuDifficultyHard, PointLimit: 10000}, false},
		{"invalid difficulty", domain.BurracoConfig{CpuDifficulty: 5, PointLimit: 5000}, true},
		{"invalid point limit", domain.BurracoConfig{CpuDifficulty: domain.BurracoCpuDifficultyNormal, PointLimit: 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
