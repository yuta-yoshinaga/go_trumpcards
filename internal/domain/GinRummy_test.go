//go:build test

package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestGinRummy() *domain.GinRummy {
	players := []*domain.GinRummyPlayer{
		domain.NewGinRummyPlayer(true),
		domain.NewGinRummyPlayer(false),
	}
	return domain.NewGinRummy(domain.NewTrumpCards(0), players, domain.DefaultGinRummyConfig())
}

func newTestGinRummyWithDifficulty(d domain.GinRummyCpuDifficulty) *domain.GinRummy {
	players := []*domain.GinRummyPlayer{
		domain.NewGinRummyPlayer(true),
		domain.NewGinRummyPlayer(false),
	}
	cfg := domain.DefaultGinRummyConfig()
	cfg.CpuDifficulty = d
	return domain.NewGinRummy(domain.NewTrumpCards(0), players, cfg)
}

func setupGinRummyDrawPhase(g *domain.GinRummy, currentIdx int) {
	g.SetPhase(domain.GinRummyPhaseDraw)
	g.SetCurrentPlayerIdx(currentIdx)
}

func setupGinRummyDiscardPhase(g *domain.GinRummy, currentIdx int) {
	g.SetPhase(domain.GinRummyPhaseDiscard)
	g.SetCurrentPlayerIdx(currentIdx)
}

func TestNewGinRummy(t *testing.T) {
	g := newTestGinRummy()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetKnockerIdx())
}

func TestGinRummy_Reset(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()

	assert.Equal(t, domain.GinRummyPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, -1, g.GetKnockerIdx())
	assert.False(t, g.GetIsGin())

	// Each player should have 10 cards
	for i := 0; i < 2; i++ {
		assert.Equal(t, 10, g.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, g.GetPlayer(i).GetRoundScore())
		assert.Equal(t, 0, g.GetPlayer(i).GetCumulativeScore())
	}

	// Discard pile should have 1 card
	assert.Len(t, g.GetDiscardPile(), 1)

	// Draw pile: 52 - 20 (2*10) - 1 = 31
	assert.Equal(t, 31, g.GetDrawPileCount())

	// Action log should be empty after reset
	assert.Nil(t, g.GetActionLog())
}

func TestGinRummy_Reset_ClearsAllState(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()

	g.GetPlayer(0).SetCumulativeScore(300)
	g.SetPhase(domain.GinRummyPhaseGameEnd)

	g.Reset()
	assert.Equal(t, domain.GinRummyPhaseDraw, g.GetPhase())
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore())
}

func TestGinRummy_Getters(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()

	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.NotNil(t, g.GetPlayer(0))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(2))

	cfg := g.GetConfig()
	assert.Equal(t, domain.GinRummyCpuDifficultyNormal, cfg.CpuDifficulty)

	g.SetConfig(domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyHard, PointLimit: 200})
	assert.Equal(t, domain.GinRummyCpuDifficultyHard, g.GetConfig().CpuDifficulty)
}

func TestGinRummy_IsHumanTurn(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()

	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())

	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())

	// Out of range
	g.SetCurrentPlayerIdx(-1)
	assert.False(t, g.IsHumanTurn())

	g.SetCurrentPlayerIdx(2)
	assert.False(t, g.IsHumanTurn())
}

func TestGinRummy_GetDiscardTop(t *testing.T) {
	g := newTestGinRummy()

	// Empty discard pile
	g.SetDiscardPile(nil)
	assert.Nil(t, g.GetDiscardTop())

	// With cards
	card := domain.NewCard(domain.CardDesignSpade, 5, false)
	g.SetDiscardPile([]*domain.Card{card})
	assert.Equal(t, card, g.GetDiscardTop())
}

// --- PlayerDrawFromStock ---

func TestGinRummy_PlayerDrawFromStock(t *testing.T) {
	t.Run("valid draw from stock", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDrawPhase(g, 0)

		drawCard := domain.NewCard(domain.CardDesignHeart, 2, false)
		g.SetDrawPile([]*domain.Card{drawCard})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

		initialSize := g.GetPlayer(0).GetCardsSize()
		err := g.PlayerDrawFromStock()
		assert.NoError(t, err)
		assert.Equal(t, initialSize+1, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.GinRummyPhaseDiscard, g.GetPhase())
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseGameEnd)
		// Simulate game end by setting knocker and scoring
		g.GetPlayer(0).SetCumulativeScore(200)
		g.SetKnockerIdx(0)
		g.SetKnockerDeadwood(nil)
		// Use PlayerDrawFromStock to trigger the gameEndFlag check -
		// but first we need to trigger game end through the normal flow
		// Instead, we'll set up state to force a game-end flow
		g2 := newTestGinRummy()
		g2.Reset()
		// Force game end
		g2.GetPlayer(0).SetCumulativeScore(200)
		g2.SetPhase(domain.GinRummyPhaseRoundEnd)
		// Make player 0 knock with gin to trigger scoreRound -> checkGameEnd
		// Simpler: just test the error path directly
		g3 := newTestGinRummy()
		g3.Reset()
		setupGinRummyDrawPhase(g3, 0)
		// Trigger game end manually: set gameEndFlag via scoring
		g3.GetPlayer(0).SetCumulativeScore(200)
		cfg := domain.DefaultGinRummyConfig()
		cfg.PointLimit = 100
		g3.SetConfig(cfg)
		// Force a knock-gin to end the game - just draw from stock when ended
		// Let's just use a fresh approach
		gEnd := newTestGinRummy()
		gEnd.Reset()
		// Build a gin hand for player 0
		gEnd.GetPlayer(0).Reset()
		// Three runs: Spade 1-2-3, Heart 1-2-3, Diamond 1-2-3, + Club 1
		for _, v := range []int{1, 2, 3} {
			gEnd.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			gEnd.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			gEnd.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, v, false))
		}
		gEnd.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		// Draw a card to make 11 cards
		gEnd.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		setupGinRummyDrawPhase(gEnd, 0)
		_ = gEnd.PlayerDrawFromStock()

		// Now in discard phase with 12 cards, gin possible
		// Set high cumulative score so game ends after knock
		cfgEnd := domain.DefaultGinRummyConfig()
		cfgEnd.PointLimit = 1
		gEnd.SetConfig(cfgEnd)

		// Knock discarding index 10 (Club 1) -> remaining is 3 runs of 3 + Club 2 -> not gin
		// Actually, let's simplify - just knock with gin hand
		// Build a simpler approach: set cumulative high, force knock, let scoreRound trigger game end
		gSimple := newTestGinRummy()
		gSimple.Reset()
		gSimple.GetPlayer(0).SetCumulativeScore(200)
		cfgSimple := domain.DefaultGinRummyConfig()
		cfgSimple.PointLimit = 50
		gSimple.SetConfig(cfgSimple)
		// Build gin hand: 10 cards that form perfect melds + 1 extra for discard
		gSimple.GetPlayer(0).Reset()
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		gSimple.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false)) // extra to discard
		setupGinRummyDiscardPhase(gSimple, 0)
		err := gSimple.PlayerKnock(10) // discard Club K
		assert.NoError(t, err)
		assert.True(t, gSimple.GetGameEndFlag())

		// Now try to draw from stock
		err = gSimple.PlayerDrawFromStock()
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseDiscard)
		g.SetCurrentPlayerIdx(0)

		err := g.PlayerDrawFromStock()
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDrawPhase(g, 1) // CPU turn

		err := g.PlayerDrawFromStock()
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("empty stock causes draw round", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDrawPhase(g, 0)
		g.SetDrawPile(nil) // empty stock

		err := g.PlayerDrawFromStock()
		assert.NoError(t, err)
		// Phase should be RoundEnd (draw round)
		assert.Equal(t, domain.GinRummyPhaseRoundEnd, g.GetPhase())
	})
}

// --- PlayerDrawFromDiscard ---

func TestGinRummy_PlayerDrawFromDiscard(t *testing.T) {
	t.Run("valid draw from discard", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDrawPhase(g, 0)
		topCard := domain.NewCard(domain.CardDesignHeart, 5, false)
		g.SetDiscardPile([]*domain.Card{topCard})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		err := g.PlayerDrawFromDiscard()
		assert.NoError(t, err)
		assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.GinRummyPhaseDiscard, g.GetPhase())
		assert.Empty(t, g.GetDiscardPile())
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		// Force game end
		g.GetPlayer(0).SetCumulativeScore(200)
		g.SetConfig(domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyNormal, PointLimit: 50})
		g.GetPlayer(0).Reset()
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, v, false))
		}
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		setupGinRummyDiscardPhase(g, 0)
		_ = g.PlayerKnock(10)
		assert.True(t, g.GetGameEndFlag())

		err := g.PlayerDrawFromDiscard()
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseDiscard)
		g.SetCurrentPlayerIdx(0)

		err := g.PlayerDrawFromDiscard()
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDrawPhase(g, 1)

		err := g.PlayerDrawFromDiscard()
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("empty discard pile error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDrawPhase(g, 0)
		g.SetDiscardPile(nil)

		err := g.PlayerDrawFromDiscard()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

// --- PlayerDiscard ---

func TestGinRummy_PlayerDiscard(t *testing.T) {
	t.Run("valid discard", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 0)
		g.SetDiscardPile(nil)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

		err := g.PlayerDiscard(0)
		assert.NoError(t, err)
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
		assert.Len(t, g.GetDiscardPile(), 1)
		// Turn should advance to CPU (player 1)
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
		assert.Equal(t, domain.GinRummyPhaseDraw, g.GetPhase())
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(200)
		g.SetConfig(domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyNormal, PointLimit: 50})
		g.GetPlayer(0).Reset()
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, v, false))
		}
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		setupGinRummyDiscardPhase(g, 0)
		_ = g.PlayerKnock(10)

		err := g.PlayerDiscard(0)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseDraw)
		g.SetCurrentPlayerIdx(0)

		err := g.PlayerDiscard(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 1)

		err := g.PlayerDiscard(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("invalid card index negative", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 0)

		err := g.PlayerDiscard(-1)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("invalid card index too large", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 0)

		err := g.PlayerDiscard(100)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})
}

// --- PlayerKnock ---

func TestGinRummy_PlayerKnock(t *testing.T) {
	t.Run("valid knock", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 0)
		g.GetPlayer(0).Reset()
		// Hand: Spade 1,2,3 (run) + Heart 1,2,3 (run) + Diamond 1,2,3 (run) + Club 5 (deadwood=5) + Club K (discard)
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, v, false))
		}
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 5, false)) // deadwood = 5
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

		err := g.PlayerKnock(10) // discard Club K
		assert.NoError(t, err)
		assert.Equal(t, 0, g.GetKnockerIdx())
		assert.False(t, g.GetIsGin())
		// Should be in layoff phase with opponent's turn
		assert.Equal(t, domain.GinRummyPhaseLayoff, g.GetPhase())
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	})

	t.Run("gin detection", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 0)
		g.GetPlayer(0).Reset()
		// Perfect gin hand: 3 runs of 3 + 1 set of 1 (remaining after discard)
		// Actually need 10 cards that form melds after discarding 1
		// Spade 1,2,3 + Heart 1,2,3 + Diamond 1,2,3 + Club 1 = 10 melds cards
		// + Club K to discard => after discard: 3 runs of 3 + Club 1 as deadwood? No.
		// Need: Spade 1,2,3 + Heart 1,2,3 + Diamond 1,2,3 + Spade 4 (extends run) = 10 cards
		// But with Spade 1,2,3,4 as run and Heart 1,2,3, Diamond 1,2,3 = all melds after discard
		// Actually we want 11 cards, discard 1, leaving 10 that form melds
		// Let's make: S1,S2,S3 + H1,H2,H3 + D1,D2,D3 + C1 (deadwood=1) + C13 (discard)
		// Wait, C1 alone can't make a meld by itself
		// For gin: all 10 remaining must be in melds
		// S1,S2,S3 (run) + H1,H2,H3 (run) + D1,D2,D3 (run) = 9 cards -> need 1 more in a meld
		// Add C1 -> set of Aces: S1,H1,D1,C1 is a set of 4
		// But S1 is already in the run. Melds can overlap? No.
		// Best melds algo: either use S1 in run or in set.
		// S1,S2,S3 run + H1,H2,H3 run + D1,D2,D3 run + C1 deadwood = deadwood 1. Not gin.
		// Alt: S1,H1,D1,C1 set + S2,S3,S4 run + H2,H3,H4 run = 4+3+3=10 cards, gin!
		for _, v := range []int{1} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		// 10 cards. Need 11 so we can discard 1
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false)) // discard this

		err := g.PlayerKnock(10) // discard Club K
		assert.NoError(t, err)
		assert.True(t, g.GetIsGin())
		assert.Equal(t, 0, g.GetKnockerIdx())
		// Gin -> no layoff, goes to round end or game end
		phase := g.GetPhase()
		assert.True(t, phase == domain.GinRummyPhaseRoundEnd || phase == domain.GinRummyPhaseGameEnd)
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		// Force game end
		g.GetPlayer(0).SetCumulativeScore(200)
		g.SetConfig(domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyNormal, PointLimit: 50})
		g.GetPlayer(0).Reset()
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, v, false))
		}
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		setupGinRummyDiscardPhase(g, 0)
		_ = g.PlayerKnock(10)

		err := g.PlayerKnock(0)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseDraw)
		g.SetCurrentPlayerIdx(0)

		err := g.PlayerKnock(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 1)

		err := g.PlayerKnock(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("invalid card index negative", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 0)

		err := g.PlayerKnock(-1)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("invalid card index too large", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 0)

		err := g.PlayerKnock(100)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("deadwood too high", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 0)
		g.GetPlayer(0).Reset()
		// All high-value cards with no melds -> deadwood >> 10
		for i := 0; i < 11; i++ {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade+(i%4)+1, 10+i%4, false))
		}
		// Reset and give specific bad hand
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 12, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))

		err := g.PlayerKnock(0) // discard Spade 10, deadwood still >> 10
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

// --- PlayerLayoff ---

func TestGinRummy_PlayerLayoff(t *testing.T) {
	t.Run("valid layoff", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseLayoff)
		g.SetCurrentPlayerIdx(0) // human is the non-knocker
		g.SetKnockerIdx(1)
		// Knocker has a set of Aces (Spade, Heart, Diamond)
		g.SetKnockerMelds([][]*domain.Card{
			{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 1, false),
				domain.NewCard(domain.CardDesignDiamond, 1, false),
			},
		})
		g.SetKnockerDeadwood([]*domain.Card{domain.NewCard(domain.CardDesignClover, 5, false)})

		// Human has Club Ace which can be laid off
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		err := g.PlayerLayoff([]int{0})
		assert.NoError(t, err)
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	})

	t.Run("empty indices ends layoff", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseLayoff)
		g.SetCurrentPlayerIdx(0)
		g.SetKnockerIdx(1)
		g.SetKnockerMelds([][]*domain.Card{
			{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 1, false),
				domain.NewCard(domain.CardDesignDiamond, 1, false),
			},
		})
		g.SetKnockerDeadwood([]*domain.Card{domain.NewCard(domain.CardDesignClover, 5, false)})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		err := g.PlayerLayoff([]int{})
		assert.NoError(t, err)
		// Should have scored the round
		phase := g.GetPhase()
		assert.True(t, phase == domain.GinRummyPhaseRoundEnd || phase == domain.GinRummyPhaseGameEnd)
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(200)
		g.SetConfig(domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyNormal, PointLimit: 50})
		g.GetPlayer(0).Reset()
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, v, false))
		}
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		setupGinRummyDiscardPhase(g, 0)
		_ = g.PlayerKnock(10)

		err := g.PlayerLayoff([]int{0})
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseDraw)
		g.SetCurrentPlayerIdx(0)

		err := g.PlayerLayoff([]int{0})
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseLayoff)
		g.SetCurrentPlayerIdx(1)

		err := g.PlayerLayoff([]int{0})
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("invalid index", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseLayoff)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))

		err := g.PlayerLayoff([]int{5})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("negative index", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseLayoff)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))

		err := g.PlayerLayoff([]int{-1})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("duplicate index", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseLayoff)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

		err := g.PlayerLayoff([]int{0, 0})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("invalid layoff card", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseLayoff)
		g.SetCurrentPlayerIdx(0)
		g.SetKnockerIdx(1)
		g.SetKnockerMelds([][]*domain.Card{
			{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 1, false),
				domain.NewCard(domain.CardDesignDiamond, 1, false),
			},
		})
		g.SetKnockerDeadwood(nil)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 7, false)) // Can't lay off 7

		err := g.PlayerLayoff([]int{0})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

// --- CpuPlay ---

func TestGinRummy_CpuPlay(t *testing.T) {
	t.Run("CPU draw phase", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDrawPhase(g, 1)
		g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		g.GetPlayer(1).Reset()
		for i := 1; i <= 10; i++ {
			g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, i, false))
		}

		g.CpuPlay()
		// CPU should have drawn and now be in discard phase or have discarded
		assert.True(t, g.GetPlayer(1).GetCardsSize() >= 10)
	})

	t.Run("CPU discard phase", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 1)
		g.GetPlayer(1).Reset()
		// Give CPU cards with high deadwood so it won't knock
		for i := 0; i < 11; i++ {
			suit := domain.CardDesignSpade + (i % 4)
			val := 10 + (i % 4)
			if val > 13 {
				val = val - 4
			}
			g.GetPlayer(1).AddCard(domain.NewCard(suit, val, false))
		}
		// Reset and give specific hand
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 12, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))

		g.CpuPlay()
		// CPU should have discarded one card
		assert.Equal(t, 10, g.GetPlayer(1).GetCardsSize())
	})

	t.Run("CPU layoff phase", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseLayoff)
		g.SetCurrentPlayerIdx(1)
		g.SetKnockerIdx(0)
		g.SetKnockerMelds([][]*domain.Card{
			{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 1, false),
				domain.NewCard(domain.CardDesignDiamond, 1, false),
			},
		})
		g.SetKnockerDeadwood([]*domain.Card{domain.NewCard(domain.CardDesignClover, 5, false)})

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 1, false)) // can lay off
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		g.CpuPlay()
		// CPU should have laid off the Ace
		assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
	})

	t.Run("no-op when game ended", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(200)
		g.SetConfig(domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyNormal, PointLimit: 50})
		g.GetPlayer(0).Reset()
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
		}
		for _, v := range []int{1, 2, 3} {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, v, false))
		}
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		setupGinRummyDiscardPhase(g, 0)
		_ = g.PlayerKnock(10)

		g.SetCurrentPlayerIdx(1)
		g.CpuPlay() // should not panic
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseDraw)
		g.SetCurrentPlayerIdx(0) // human
		g.CpuPlay()              // no-op
	})

	t.Run("CPU draw from discard when beneficial hard", func(t *testing.T) {
		g := newTestGinRummyWithDifficulty(domain.GinRummyCpuDifficultyHard)
		g.Reset()
		setupGinRummyDrawPhase(g, 1)
		// Discard top: Spade 1. CPU has Spade 2,3 -> drawing Spade 1 completes run
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignClover, 13, false)})
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		for i := 0; i < 8; i++ {
			g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, i+5, false))
		}

		g.CpuPlay()
		assert.Equal(t, domain.GinRummyPhaseDiscard, g.GetPhase())
	})

	t.Run("CPU draw from discard normal", func(t *testing.T) {
		g := newTestGinRummyWithDifficulty(domain.GinRummyCpuDifficultyNormal)
		g.Reset()
		setupGinRummyDrawPhase(g, 1)
		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignClover, 13, false)})
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		for i := 0; i < 8; i++ {
			g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, i+5, false))
		}

		g.CpuPlay()
		assert.Equal(t, domain.GinRummyPhaseDiscard, g.GetPhase())
	})

	t.Run("CPU knocks when deadwood low hard", func(t *testing.T) {
		g := newTestGinRummyWithDifficulty(domain.GinRummyCpuDifficultyHard)
		g.Reset()
		setupGinRummyDiscardPhase(g, 1)
		g.GetPlayer(1).Reset()
		// Give CPU a hand where deadwood after best discard is <= 5
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 2, false)) // deadwood=2
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

		g.CpuPlay()
		// CPU should knock
		assert.Equal(t, 1, g.GetKnockerIdx())
	})

	t.Run("CPU knocks when deadwood low normal", func(t *testing.T) {
		g := newTestGinRummyWithDifficulty(domain.GinRummyCpuDifficultyNormal)
		g.Reset()
		setupGinRummyDiscardPhase(g, 1)
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 5, false)) // deadwood=5
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

		g.CpuPlay()
		assert.Equal(t, 1, g.GetKnockerIdx())
	})

	t.Run("CPU knocks easy always when can", func(t *testing.T) {
		g := newTestGinRummyWithDifficulty(domain.GinRummyCpuDifficultyEasy)
		g.Reset()
		setupGinRummyDiscardPhase(g, 1)
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 10, false)) // deadwood=10
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

		g.CpuPlay()
		assert.Equal(t, 1, g.GetKnockerIdx())
	})

	t.Run("CPU gin", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDiscardPhase(g, 1)
		g.GetPlayer(1).Reset()
		// Perfect gin hand after discarding last card
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 13, false)) // discard

		g.CpuPlay()
		assert.True(t, g.GetIsGin())
		assert.Equal(t, 1, g.GetKnockerIdx())
	})

	t.Run("CPU empty draw pile causes draw round", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		setupGinRummyDrawPhase(g, 1)
		g.SetDrawPile(nil)
		g.SetDiscardPile(nil) // no discard either
		g.GetPlayer(1).Reset()
		for i := 0; i < 10; i++ {
			g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, i+1, false))
		}

		g.CpuPlay()
		assert.Equal(t, domain.GinRummyPhaseRoundEnd, g.GetPhase())
	})

	t.Run("CPU hard does not knock when deadwood > 5 and > 0", func(t *testing.T) {
		g := newTestGinRummyWithDifficulty(domain.GinRummyCpuDifficultyHard)
		g.Reset()
		setupGinRummyDiscardPhase(g, 1)
		g.GetPlayer(1).Reset()
		// Deadwood after best discard = 8 (> 5 for hard, won't knock)
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 8, false)) // deadwood=8
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

		g.CpuPlay()
		// Should discard, not knock (hard requires <= 5)
		assert.Equal(t, -1, g.GetKnockerIdx())
	})
}

// --- NextRound ---

func TestGinRummy_NextRound(t *testing.T) {
	t.Run("works in RoundEnd phase", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.SetPhase(domain.GinRummyPhaseRoundEnd)
		g.SetRoundNumber(1)

		g.NextRound()

		assert.Equal(t, domain.GinRummyPhaseDraw, g.GetPhase())
		assert.Equal(t, 2, g.GetRoundNumber())
		assert.Equal(t, 0, g.GetCurrentPlayerIdx())
		for i := 0; i < 2; i++ {
			assert.Equal(t, 10, g.GetPlayer(i).GetCardsSize())
		}
	})

	t.Run("no-op in Draw phase", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		round := g.GetRoundNumber()
		g.SetPhase(domain.GinRummyPhaseDraw)

		g.NextRound()

		assert.Equal(t, round, g.GetRoundNumber())
	})
}

// --- ScoreRound ---

func TestGinRummy_ScoreRound(t *testing.T) {
	t.Run("noop - scoring done in knock/layoff", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.ScoreRound() // should not panic
	})
}

// --- Meld detection ---

func TestGinRummy_FindBestMelds(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		melds, deadwood := domain.FindBestMelds(nil)
		assert.Nil(t, melds)
		assert.Nil(t, deadwood)
	})

	t.Run("no melds possible", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
		}
		melds, deadwood := domain.FindBestMelds(cards)
		assert.Empty(t, melds)
		assert.Len(t, deadwood, 3)
	})

	t.Run("set of 3", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
			domain.NewCard(domain.CardDesignClover, 9, false),
		}
		melds, deadwood := domain.FindBestMelds(cards)
		assert.Len(t, melds, 1)
		assert.Len(t, melds[0], 3)
		assert.Len(t, deadwood, 1)
	})

	t.Run("set of 4", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		}
		melds, deadwood := domain.FindBestMelds(cards)
		assert.Len(t, melds, 1)
		assert.Len(t, melds[0], 4)
		assert.Empty(t, deadwood)
	})

	t.Run("run of 3", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
		}
		melds, deadwood := domain.FindBestMelds(cards)
		assert.Len(t, melds, 1)
		assert.Len(t, deadwood, 1)
	})

	t.Run("run of 4", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
		}
		melds, deadwood := domain.FindBestMelds(cards)
		assert.NotEmpty(t, melds)
		assert.Empty(t, deadwood)
	})

	t.Run("combined melds", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		}
		melds, deadwood := domain.FindBestMelds(cards)
		assert.Len(t, melds, 2)
		assert.Empty(t, deadwood)
	})
}

func TestGinRummy_CalcDeadwoodValue(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, 0, domain.CalcDeadwoodValue(nil))
	})

	t.Run("mixed values", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),    // 1
			domain.NewCard(domain.CardDesignHeart, 5, false),    // 5
			domain.NewCard(domain.CardDesignDiamond, 10, false), // 10
			domain.NewCard(domain.CardDesignClover, 13, false),  // 10
		}
		assert.Equal(t, 26, domain.CalcDeadwoodValue(cards))
	})
}

func TestGinRummy_GinRummyCardValue(t *testing.T) {
	// Ace = 1
	assert.Equal(t, 1, domain.GinRummyCardValue(domain.NewCard(domain.CardDesignSpade, 1, false)))
	// 2-9 = face value
	assert.Equal(t, 5, domain.GinRummyCardValue(domain.NewCard(domain.CardDesignSpade, 5, false)))
	assert.Equal(t, 9, domain.GinRummyCardValue(domain.NewCard(domain.CardDesignSpade, 9, false)))
	// 10, J, Q, K = 10
	assert.Equal(t, 10, domain.GinRummyCardValue(domain.NewCard(domain.CardDesignSpade, 10, false)))
	assert.Equal(t, 10, domain.GinRummyCardValue(domain.NewCard(domain.CardDesignSpade, 11, false)))
	assert.Equal(t, 10, domain.GinRummyCardValue(domain.NewCard(domain.CardDesignSpade, 12, false)))
	assert.Equal(t, 10, domain.GinRummyCardValue(domain.NewCard(domain.CardDesignSpade, 13, false)))
}

// --- State setters ---

func TestGinRummy_StateSetters(t *testing.T) {
	g := newTestGinRummy()

	g.SetPhase(domain.GinRummyPhaseLayoff)
	assert.Equal(t, domain.GinRummyPhaseLayoff, g.GetPhase())

	g.SetRoundNumber(5)
	assert.Equal(t, 5, g.GetRoundNumber())

	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())

	pile := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	g.SetDiscardPile(pile)
	assert.Equal(t, pile, g.GetDiscardPile())

	draw := []*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)}
	g.SetDrawPile(draw)
	assert.Equal(t, 1, g.GetDrawPileCount())

	g.SetKnockerIdx(0)
	assert.Equal(t, 0, g.GetKnockerIdx())

	melds := [][]*domain.Card{{domain.NewCard(domain.CardDesignSpade, 1, false)}}
	g.SetKnockerMelds(melds)
	assert.Equal(t, melds, g.GetKnockerMelds())

	dw := []*domain.Card{domain.NewCard(domain.CardDesignClover, 5, false)}
	g.SetKnockerDeadwood(dw)
	assert.Equal(t, dw, g.GetKnockerDeadwood())

	g.SetIsGin(true)
	assert.True(t, g.GetIsGin())
	g.SetIsGin(false)
	assert.False(t, g.GetIsGin())
}

// --- playerName ---

func TestGinRummy_PlayerName(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()

	// Test via action log which contains player names
	setupGinRummyDrawPhase(g, 0)
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	err := g.PlayerDrawFromStock()
	require.NoError(t, err)

	log := g.GetActionLog()
	assert.NotEmpty(t, log)
	assert.Contains(t, log[0].Detail, "You")
}

// --- Action log ---

func TestGinRummy_ActionLog(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()
	assert.Nil(t, g.GetActionLog())

	setupGinRummyDrawPhase(g, 0)
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	err := g.PlayerDrawFromStock()
	require.NoError(t, err)

	log := g.GetActionLog()
	assert.NotEmpty(t, log)
	assert.Equal(t, "draw_stock", log[0].ActionType)
	assert.Equal(t, 0, log[0].PlayerIdx)
}

// --- scoreRound: gin, undercut, normal ---

func TestGinRummy_ScoreRound_Gin(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()
	setupGinRummyDiscardPhase(g, 0)
	g.GetPlayer(0).Reset()
	// Gin hand
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

	// Opponent has some deadwood
	g.GetPlayer(1).Reset()
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 10, false)) // 10
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 9, false))  // 9
	for i := 0; i < 8; i++ {
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, i+5, false))
	}

	err := g.PlayerKnock(10)
	assert.NoError(t, err)
	assert.True(t, g.GetIsGin())

	// Knocker should get opponent's deadwood + GinBonus(25)
	score := g.GetPlayer(0).GetRoundScore()
	assert.True(t, score > 0)
}

func TestGinRummy_ScoreRound_Undercut(t *testing.T) {
	// Human knocks with deadwood=10, opponent (CPU) has deadwood=0 -> undercut
	// After knock with non-gin, phase goes to Layoff with opponent's turn (CPU)
	// CPU does layoff automatically. So we test via human knock then CpuPlay for layoff.
	g := newTestGinRummy()
	g.Reset()
	setupGinRummyDiscardPhase(g, 0)
	g.GetPlayer(0).Reset()
	// Hand with deadwood=10 after discarding K
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 10, false)) // deadwood=10
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false)) // discard

	// Opponent has all melds -> deadwood = 0 -> undercut
	g.GetPlayer(1).Reset()
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

	err := g.PlayerKnock(10) // discard Club K
	assert.NoError(t, err)
	assert.Equal(t, 0, g.GetKnockerIdx())
	assert.False(t, g.GetIsGin())
	assert.Equal(t, domain.GinRummyPhaseLayoff, g.GetPhase())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // CPU's layoff turn

	// CPU does layoff (will call cpuLayoff which scores round)
	g.CpuPlay()

	// Opponent (CPU idx 1) should have gotten undercut bonus
	// knockerDeadwood=10, opponentDeadwood=0, score = 10-0+25 = 35
	assert.True(t, g.GetPlayer(1).GetRoundScore() > 0)
	assert.True(t, g.GetPlayer(1).GetCumulativeScore() > 0)
}

func TestGinRummy_ScoreRound_Normal(t *testing.T) {
	// Human knocks with low deadwood=2, opponent has higher deadwood -> normal scoring
	g := newTestGinRummy()
	g.Reset()
	setupGinRummyDiscardPhase(g, 0)
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 7, false))  // deadwood=7 (no set possible)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false)) // discard

	// CPU has high deadwood (no melds possible, scattered cards)
	g.GetPlayer(1).Reset()
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

	err := g.PlayerKnock(10) // discard Club K, deadwood=2
	assert.NoError(t, err)
	assert.Equal(t, 0, g.GetKnockerIdx())
	assert.False(t, g.GetIsGin())

	// CPU layoff phase
	assert.Equal(t, domain.GinRummyPhaseLayoff, g.GetPhase())
	g.CpuPlay() // CPU lays off what it can, then scores

	// Knocker (human) should score the deadwood difference
	assert.True(t, g.GetPlayer(0).GetRoundScore() > 0)
}

// --- endRoundDraw ---

func TestGinRummy_EndRoundDraw(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()
	setupGinRummyDrawPhase(g, 0)
	g.SetDrawPile(nil) // empty draw pile

	err := g.PlayerDrawFromStock()
	assert.NoError(t, err)
	assert.Equal(t, domain.GinRummyPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, -1, g.GetKnockerIdx())
}

// --- checkGameEnd ---

func TestGinRummy_CheckGameEnd(t *testing.T) {
	t.Run("game end detected", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(200)
		cfg := domain.DefaultGinRummyConfig()
		cfg.PointLimit = 50
		g.SetConfig(cfg)

		// Trigger gin to cause scoreRound -> checkGameEnd
		setupGinRummyDiscardPhase(g, 0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

		err := g.PlayerKnock(10)
		assert.NoError(t, err)
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, domain.GinRummyPhaseGameEnd, g.GetPhase())
		assert.Equal(t, 0, g.GetWinnerIdx())
	})

	t.Run("game not ended", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		// No one has enough points
		assert.False(t, g.GetGameEndFlag())
	})

	t.Run("tie-breaking highest score wins", func(t *testing.T) {
		g := newTestGinRummy()
		g.Reset()
		cfg := domain.DefaultGinRummyConfig()
		cfg.PointLimit = 50
		g.SetConfig(cfg)

		// Both players above limit, player 1 has higher
		g.GetPlayer(0).SetCumulativeScore(55)
		g.GetPlayer(1).SetCumulativeScore(70)

		// Trigger through draw round end (both above limit)
		setupGinRummyDrawPhase(g, 0)
		g.SetDrawPile(nil)
		err := g.PlayerDrawFromStock()
		assert.NoError(t, err)
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, 1, g.GetWinnerIdx())
	})
}

// --- Full round flow ---

func TestGinRummy_FullRoundFlow(t *testing.T) {
	g := newTestGinRummy()
	g.Reset()

	// Human draws, discards, next round
	setupGinRummyDrawPhase(g, 0)
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignClover, 13, false)})
	g.GetPlayer(0).Reset()
	for i := 0; i < 10; i++ {
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, i+1, false))
	}

	err := g.PlayerDrawFromStock()
	require.NoError(t, err)
	assert.Equal(t, domain.GinRummyPhaseDiscard, g.GetPhase())
	assert.Equal(t, 11, g.GetPlayer(0).GetCardsSize())

	err = g.PlayerDiscard(10) // discard the drawn card
	require.NoError(t, err)
	assert.Equal(t, 10, g.GetPlayer(0).GetCardsSize())
}

// **レイオフフェーズの主題そのもの (#4823)。**canLayoff と同じ判定を返すこと。
func TestGinRummy_LayoffTargets(t *testing.T) {
	g := domain.NewDefaultGinRummy()
	g.Reset()
	g.SetKnockerMelds([][]*domain.Card{
		{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		},
		{
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignSpade, 4, false),
		},
	})

	// 4 枚目の 7 は 1 つ目のセットへ。
	assert.Equal(t, []int{0}, g.LayoffTargets(domain.NewCard(domain.CardDesignDiamond, 7, false)))
	// ♠5 はランの上端へ。
	assert.Equal(t, []int{1}, g.LayoffTargets(domain.NewCard(domain.CardDesignSpade, 5, false)))
	// どこにも足せない札と nil。
	assert.Nil(t, g.LayoffTargets(domain.NewCard(domain.CardDesignHeart, 10, false)))
	assert.Nil(t, g.LayoffTargets(nil))
}

// #5500 の副産物: Easy の拾い判断は 1/GinRummyEasyPickOneIn の乱数分岐で、
// 片方の枝しか通らないと codecov に残る。**乱択なのでリトライで両方を通す**
// (internal/CLAUDE.md の「1000回までのリトライで両分岐を覆う」手法)。
//
// どちらを引いたかは行動ログ (draw_discard / draw_stock) で見る。盤面の枚数は
// 直後の捨てで戻るので、枚数からは区別できない。
func TestGinRummy_EasyCpuTakesTheDiscardSometimesAndSometimesNot(t *testing.T) {
	drewFromDiscard, drewFromStock := false, false

	for i := 0; i < 1000 && (!drewFromDiscard || !drewFromStock); i++ {
		g := domain.NewDefaultGinRummy()
		g.Reset()
		cfg := g.GetConfig()
		cfg.CpuDifficulty = domain.GinRummyCpuDifficultyEasy
		g.SetConfig(cfg)
		g.SetCurrentPlayerIdx(1)
		g.SetPhase(domain.GinRummyPhaseDraw)

		before := len(g.GetActionLog())
		g.CpuPlay()
		for _, e := range g.GetActionLog()[before:] {
			switch e.ActionType {
			case "draw_discard":
				drewFromDiscard = true
			case "draw_stock":
				drewFromStock = true
			}
		}
	}

	assert.True(t, drewFromDiscard, "Easy が捨て札を拾う枝を一度も通らなかった")
	assert.True(t, drewFromStock, "Easy が山札から引く枝を一度も通らなかった")
}
