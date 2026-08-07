//go:build test

package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestCrazyEights() *domain.CrazyEights {
	players := []*domain.CrazyEightsPlayer{
		domain.NewCrazyEightsPlayer(true),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
	}
	return domain.NewCrazyEights(domain.NewTrumpCards(0), players, domain.DefaultCrazyEightsConfig())
}

func newTestCrazyEightsWithDifficulty(d domain.CrazyEightsCpuDifficulty) *domain.CrazyEights {
	players := []*domain.CrazyEightsPlayer{
		domain.NewCrazyEightsPlayer(true),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
	}
	cfg := domain.DefaultCrazyEightsConfig()
	cfg.CpuDifficulty = d
	return domain.NewCrazyEights(domain.NewTrumpCards(0), players, cfg)
}

// setupCrazyEightsPlayPhase sets up a deterministic play state.
func setupCrazyEightsPlayPhase(g *domain.CrazyEights, currentIdx int, topCard *domain.Card) {
	g.SetPhase(domain.CrazyEightsPhasePlay)
	g.SetCurrentPlayerIdx(currentIdx)
	g.SetDiscardPile([]*domain.Card{topCard})
	g.SetChosenSuit(-1)
}

func TestNewCrazyEights(t *testing.T) {
	g := newTestCrazyEights()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.Equal(t, -1, g.GetChosenSuit())
	assert.False(t, g.GetGameEndFlag())
}

func TestNewDefaultCrazyEights(t *testing.T) {
	g := domain.NewDefaultCrazyEights()
	assert.NotNil(t, g)
	assert.Equal(t, domain.CrazyEightsPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	for i := 1; i < g.GetPlayerCnt(); i++ {
		assert.False(t, g.GetPlayer(i).GetIsHuman(), "player %d should be CPU", i)
	}
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.False(t, g.GetGameEndFlag())
}

func TestCrazyEights_Reset(t *testing.T) {
	g := newTestCrazyEights()
	g.Reset()

	assert.Equal(t, domain.CrazyEightsPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, -1, g.GetChosenSuit())

	// Each player should have 5 cards
	for i := 0; i < 4; i++ {
		assert.Equal(t, 5, g.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, g.GetPlayer(i).GetRoundScore())
		assert.Equal(t, 0, g.GetPlayer(i).GetCumulativeScore())
	}

	// Discard pile should have 1 card
	assert.Len(t, g.GetDiscardPile(), 1)

	// Draw pile: 52 - 20 (4*5) - 1 = 31
	assert.Equal(t, 31, g.GetDrawPileCount())

	// Action log should be empty after reset
	assert.Nil(t, g.GetActionLog())
}

func TestCrazyEights_Reset_ClearsAllState(t *testing.T) {
	g := newTestCrazyEights()
	g.Reset()

	// Modify state then re-reset
	g.GetPlayer(0).SetCumulativeScore(300)
	g.SetPhase(domain.CrazyEightsPhaseGameEnd)

	g.Reset()
	assert.Equal(t, domain.CrazyEightsPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore())
}

func TestCrazyEights_Getters(t *testing.T) {
	g := newTestCrazyEights()
	g.Reset()

	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.NotNil(t, g.GetPlayer(0))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(4))

	cfg := g.GetConfig()
	assert.Equal(t, domain.CrazyEightsCpuDifficultyNormal, cfg.CpuDifficulty)

	g.SetConfig(domain.CrazyEightsConfig{CpuDifficulty: domain.CrazyEightsCpuDifficultyHard, PointLimit: 100})
	assert.Equal(t, domain.CrazyEightsCpuDifficultyHard, g.GetConfig().CpuDifficulty)
}

func TestCrazyEights_IsHumanTurn(t *testing.T) {
	g := newTestCrazyEights()
	g.Reset()

	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())

	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())

	// Out of range
	g.SetCurrentPlayerIdx(-1)
	assert.False(t, g.IsHumanTurn())

	g.SetCurrentPlayerIdx(4)
	assert.False(t, g.IsHumanTurn())
}

func TestCrazyEights_GetDiscardTop(t *testing.T) {
	g := newTestCrazyEights()

	// Empty discard pile
	g.SetDiscardPile(nil)
	assert.Nil(t, g.GetDiscardTop())

	// With cards
	card := domain.NewCard(domain.CardDesignSpade, 5, false)
	g.SetDiscardPile([]*domain.Card{card})
	assert.Equal(t, card, g.GetDiscardTop())
}

func TestCrazyEights_GetValidPlayIndices(t *testing.T) {
	g := newTestCrazyEights()
	g.Reset()

	// Setup: top card is Spade 5, player has Spade 3 and Heart 7
	topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
	g.SetDiscardPile([]*domain.Card{topCard})
	g.SetChosenSuit(-1)

	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // suit match
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // no match
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false)) // rank match

	indices := g.GetValidPlayIndices(0)
	assert.Len(t, indices, 2)
	assert.Contains(t, indices, 0)
	assert.Contains(t, indices, 2)
}

// --- NextRound ---

func TestCrazyEights_NextRound(t *testing.T) {
	t.Run("works in RoundEnd phase", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)
		g.SetRoundNumber(1)

		g.NextRound()

		assert.Equal(t, domain.CrazyEightsPhasePlay, g.GetPhase())
		assert.Equal(t, 2, g.GetRoundNumber())
		assert.Equal(t, 0, g.GetCurrentPlayerIdx())
		for i := 0; i < 4; i++ {
			assert.Equal(t, 5, g.GetPlayer(i).GetCardsSize())
		}
	})

	t.Run("no-op in Play phase", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		round := g.GetRoundNumber()
		g.SetPhase(domain.CrazyEightsPhasePlay)

		g.NextRound()

		assert.Equal(t, round, g.GetRoundNumber())
	})
}

// --- PlayerPlay ---

func TestCrazyEights_PlayerPlay(t *testing.T) {
	t.Run("valid play with suit match", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		// Card was played to discard pile
		assert.Equal(t, 2, len(g.GetDiscardPile()))
	})

	t.Run("valid play with rank match", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // keep hand non-empty

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
	})

	t.Run("play wild 8 enters choose suit phase", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false)) // keep non-empty

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, domain.CrazyEightsPhaseChooseSuit, g.GetPhase())
	})

	t.Run("play last card ends round", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // only card

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, domain.CrazyEightsPhaseRoundEnd, g.GetPhase())
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		// Force game end by scoring
		g.GetPlayer(0).SetCumulativeScore(300)
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)
		g.GetPlayer(0).Reset() // empty hand = winner
		g.ScoreRound()

		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)

		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard) // CPU turn

		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("card index out of range negative", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		err := g.PlayerPlay(-1)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("card index out of range too large", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		err := g.PlayerPlay(100)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidCard))
	})

	t.Run("invalid play card does not match", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // no match

		err := g.PlayerPlay(0)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

// --- PlayerChooseSuit ---

func TestCrazyEights_PlayerChooseSuit(t *testing.T) {
	t.Run("valid suit selection", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		g.SetCurrentPlayerIdx(0)

		err := g.PlayerChooseSuit(domain.CardDesignHeart)
		assert.NoError(t, err)
		assert.Equal(t, domain.CardDesignHeart, g.GetChosenSuit())
		assert.Equal(t, domain.CrazyEightsPhasePlay, g.GetPhase())
		// Turn advanced
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	})

	t.Run("all valid suits", func(t *testing.T) {
		for _, suit := range []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond} {
			g := newTestCrazyEights()
			g.Reset()
			g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
			g.SetCurrentPlayerIdx(0)
			err := g.PlayerChooseSuit(suit)
			assert.NoError(t, err)
			assert.Equal(t, suit, g.GetChosenSuit())
		}
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(300)
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)
		g.GetPlayer(0).Reset()
		g.ScoreRound()

		err := g.PlayerChooseSuit(domain.CardDesignHeart)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhasePlay)
		err := g.PlayerChooseSuit(domain.CardDesignHeart)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		g.SetCurrentPlayerIdx(1) // CPU
		err := g.PlayerChooseSuit(domain.CardDesignHeart)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("suit below range", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		g.SetCurrentPlayerIdx(0)
		err := g.PlayerChooseSuit(0)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("suit above range", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		g.SetCurrentPlayerIdx(0)
		err := g.PlayerChooseSuit(5)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

// --- PlayerDraw ---

func TestCrazyEights_PlayerDraw(t *testing.T) {
	t.Run("draw a card from draw pile", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		drawCard := domain.NewCard(domain.CardDesignHeart, 2, false)
		g.SetDrawPile([]*domain.Card{drawCard})

		g.GetPlayer(0).Reset()
		// Give player only non-playable cards so drawn card is also non-playable => turn advances
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

		initialSize := g.GetPlayer(0).GetCardsSize()
		err := g.PlayerDraw()
		assert.NoError(t, err)
		assert.Equal(t, initialSize+1, g.GetPlayer(0).GetCardsSize())
	})

	t.Run("draw when draw pile empty triggers recycle", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhasePlay)
		g.SetCurrentPlayerIdx(0)

		// Empty draw pile, discard pile with multiple cards
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		recycled := domain.NewCard(domain.CardDesignHeart, 3, false)
		g.SetDiscardPile([]*domain.Card{recycled, topCard})
		g.SetDrawPile(nil)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))

		err := g.PlayerDraw()
		assert.NoError(t, err)
	})

	t.Run("draw when both piles empty passes", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhasePlay)
		g.SetCurrentPlayerIdx(0)

		g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		g.SetDrawPile(nil)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))

		err := g.PlayerDraw()
		assert.NoError(t, err)
		// Turn should advance (pass)
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(300)
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)
		g.GetPlayer(0).Reset()
		g.ScoreRound()

		err := g.PlayerDraw()
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		err := g.PlayerDraw()
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhasePlay)
		g.SetCurrentPlayerIdx(1)
		err := g.PlayerDraw()
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("draw playable card keeps turn", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)

		// Draw pile has a playable card (same suit)
		playableCard := domain.NewCard(domain.CardDesignSpade, 9, false)
		g.SetDrawPile([]*domain.Card{playableCard})

		g.GetPlayer(0).Reset()
		// Give player only non-playable cards
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

		err := g.PlayerDraw()
		assert.NoError(t, err)
		// Turn should NOT advance because drawn card is playable
		assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	})
}

// --- CpuPlay ---

func TestCrazyEights_CpuPlay(t *testing.T) {
	t.Run("CPU plays a valid card", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard) // CPU 1

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		g.CpuPlay()
		// Card was played
		assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
	})

	t.Run("CPU draws when no valid play", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // no match

		drawCard := domain.NewCard(domain.CardDesignDiamond, 3, false)
		g.SetDrawPile([]*domain.Card{drawCard})

		g.CpuPlay()
		// Drew a card
		assert.Equal(t, 2, g.GetPlayer(1).GetCardsSize())
	})

	t.Run("no-op when game ended", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(300)
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)
		g.GetPlayer(0).Reset()
		g.ScoreRound()

		g.SetCurrentPlayerIdx(1)
		g.CpuPlay() // should not panic
	})

	t.Run("no-op in wrong phase", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		g.SetCurrentPlayerIdx(1)
		g.CpuPlay() // no-op
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhasePlay)
		g.SetCurrentPlayerIdx(0) // human
		g.CpuPlay()              // no-op
	})

	t.Run("CPU plays last card ends round", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // only card, suit match

		g.CpuPlay()
		assert.Equal(t, domain.CrazyEightsPhaseRoundEnd, g.GetPhase())
	})

	t.Run("CPU plays 8 enters choose suit phase", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

		g.CpuPlay()
		assert.Equal(t, domain.CrazyEightsPhaseChooseSuit, g.GetPhase())
	})
}

// --- CpuChooseSuit ---

func TestCrazyEights_CpuChooseSuit(t *testing.T) {
	t.Run("CPU chooses suit normal", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyNormal)
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		g.SetCurrentPlayerIdx(1)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

		g.CpuChooseSuit()
		assert.Equal(t, domain.CardDesignHeart, g.GetChosenSuit())
		assert.Equal(t, domain.CrazyEightsPhasePlay, g.GetPhase())
	})

	t.Run("no-op when game ended", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(300)
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)
		g.GetPlayer(0).Reset()
		g.ScoreRound()

		g.CpuChooseSuit()
	})

	t.Run("no-op in wrong phase", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhasePlay)
		g.SetCurrentPlayerIdx(1)
		g.CpuChooseSuit() // no-op
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		g.SetCurrentPlayerIdx(0) // human
		g.CpuChooseSuit()
	})
}

// --- ScoreRound ---

func TestCrazyEights_ScoreRound(t *testing.T) {
	t.Run("scores remaining cards to winner", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)

		// Player 0 has empty hand (winner)
		g.GetPlayer(0).Reset()
		// Player 1 has an 8 (50 pts)
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		// Player 2 has a K (10 pts)
		g.GetPlayer(2).Reset()
		g.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
		// Player 3 has an A (1 pt) and a 5 (5 pts)
		g.GetPlayer(3).Reset()
		g.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		g.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))

		g.ScoreRound()

		// Winner gets total: 50 + 10 + 1 + 5 = 66
		assert.Equal(t, 66, g.GetPlayer(0).GetRoundScore())
		assert.Equal(t, 66, g.GetPlayer(0).GetCumulativeScore())
	})

	t.Run("no-op in wrong phase", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhasePlay)
		g.ScoreRound() // no-op
	})

	t.Run("no winner found", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)

		// All players have cards
		for i := 0; i < 4; i++ {
			g.GetPlayer(i).Reset()
			g.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		}

		g.ScoreRound()
		// No score changes
		for i := 0; i < 4; i++ {
			assert.Equal(t, 0, g.GetPlayer(i).GetCumulativeScore())
		}
	})

	t.Run("game ends when score reaches point limit", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.PointLimit = 50
		g.SetConfig(cfg)

		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)
		g.GetPlayer(0).Reset() // empty hand = winner

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 8, false)) // 50 pts

		g.GetPlayer(2).Reset()
		g.GetPlayer(3).Reset()

		g.ScoreRound()

		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, domain.CrazyEightsPhaseGameEnd, g.GetPhase())
		assert.Equal(t, 0, g.GetWinnerIdx())
	})

	t.Run("card scoring: J Q K = 10 pts each", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)

		g.GetPlayer(0).Reset() // winner
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 11, false)) // J = 10
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // Q = 10
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // K = 10
		g.GetPlayer(2).Reset()
		g.GetPlayer(3).Reset()

		g.ScoreRound()
		assert.Equal(t, 30, g.GetPlayer(0).GetRoundScore())
	})

	t.Run("card scoring: numeric cards = face value", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)

		g.GetPlayer(0).Reset() // winner
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		g.GetPlayer(2).Reset()
		g.GetPlayer(3).Reset()

		g.ScoreRound()
		assert.Equal(t, 9, g.GetPlayer(0).GetRoundScore()) // 2+7
	})

	t.Run("winner with highest cumulative score", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.PointLimit = 10
		g.SetConfig(cfg)

		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)

		// Player 2 already has high score
		g.GetPlayer(2).SetCumulativeScore(50)

		g.GetPlayer(0).Reset() // winner of round
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // 3 pts
		g.GetPlayer(2).Reset()
		g.GetPlayer(3).Reset()

		g.ScoreRound()

		// Game should end; winner should be player 2 (highest cumulative)
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, 2, g.GetWinnerIdx())
	})
}

// --- CPU AI difficulty levels ---

func TestCrazyEights_CpuPlayEasy(t *testing.T) {
	t.Run("easy CPU plays a valid card", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyEasy)
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))

		g.CpuPlay()
		assert.Equal(t, 2, g.GetPlayer(1).GetCardsSize())
	})

}

func TestCrazyEights_CpuPlayNormal(t *testing.T) {
	t.Run("normal CPU saves 8 when non-wild available", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyNormal)
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))  // suit match, playable
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))  // wild
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // no match

		g.CpuPlay()
		// Should play Spade 3 (non-wild), not the 8
		assert.Equal(t, 2, g.GetPlayer(1).GetCardsSize())
		// Check 8 is still in hand
		has8 := false
		for i := 0; i < g.GetPlayer(1).GetCardsSize(); i++ {
			if g.GetPlayer(1).GetCard(i).GetValue() == 8 {
				has8 = true
			}
		}
		assert.True(t, has8)
	})

	t.Run("normal CPU plays 8 when only wild available", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyNormal)
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))  // wild
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // no match

		g.CpuPlay()
		assert.Equal(t, domain.CrazyEightsPhaseChooseSuit, g.GetPhase())
	})

	t.Run("normal CPU prefers most common suit", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyNormal)
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))  // spade match
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))  // rank match
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))  // extra heart
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // extra heart

		g.CpuPlay()
		// Should play Heart 5 (rank match, heart is most common suit)
		// Check that Spade 3 is still in hand
		hasSpade3 := false
		for i := 0; i < g.GetPlayer(1).GetCardsSize(); i++ {
			c := g.GetPlayer(1).GetCard(i)
			if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 3 {
				hasSpade3 = true
			}
		}
		assert.True(t, hasSpade3)
	})
}

func TestCrazyEights_CpuPlayHard(t *testing.T) {
	t.Run("hard CPU saves 8 when hand > 2", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyHard)
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

		g.CpuPlay()
		// Should not play the 8
		has8 := false
		for i := 0; i < g.GetPlayer(1).GetCardsSize(); i++ {
			if g.GetPlayer(1).GetCard(i).GetValue() == 8 {
				has8 = true
			}
		}
		assert.True(t, has8)
	})

	t.Run("hard CPU uses 8 when hand <= 2", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyHard)
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // not playable

		g.CpuPlay()
		// With 2 cards, hard CPU should use 8
		assert.Equal(t, domain.CrazyEightsPhaseChooseSuit, g.GetPhase())
	})

	t.Run("hard CPU prefers high-score cards", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyHard)
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
		setupCrazyEightsPlayPhase(g, 1, topCard)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))  // low score (2)
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // high score (10, K)
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // not playable

		g.CpuPlay()
		// Should play the K (higher score) first
		has2 := false
		for i := 0; i < g.GetPlayer(1).GetCardsSize(); i++ {
			c := g.GetPlayer(1).GetCard(i)
			if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 2 {
				has2 = true
			}
		}
		assert.True(t, has2, "should keep the low-value card")
	})
}

// --- CpuChooseSuit with difficulty levels ---

func TestCrazyEights_CpuChooseSuitHard(t *testing.T) {
	t.Run("hard CPU chooses most common suit", func(t *testing.T) {
		g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyHard)
		g.Reset()
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
		g.SetCurrentPlayerIdx(1)

		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))

		g.CpuChooseSuit()
		assert.Equal(t, domain.CardDesignDiamond, g.GetChosenSuit())
	})
}

// --- chosenSuit interaction ---

func TestCrazyEights_ChosenSuitPlay(t *testing.T) {
	t.Run("must match chosen suit after 8 played", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 8, false) // top is an 8
		setupCrazyEightsPlayPhase(g, 0, topCard)
		g.SetChosenSuit(domain.CardDesignHeart) // previous player chose heart

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // spade, not heart
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false)) // heart, matches

		// Spade 3 should be invalid
		err := g.PlayerPlay(0)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))

		// Heart 7 should work (it's now at index 0 after failed play still has 2 cards)
		// Actually the hand hasn't changed. Heart 7 is at index 1
		err = g.PlayerPlay(1)
		assert.NoError(t, err)
	})

	t.Run("chosenSuit resets after next play", func(t *testing.T) {
		g := newTestCrazyEights()
		g.Reset()
		topCard := domain.NewCard(domain.CardDesignSpade, 8, false)
		setupCrazyEightsPlayPhase(g, 0, topCard)
		g.SetChosenSuit(domain.CardDesignHeart)

		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, -1, g.GetChosenSuit()) // resets after play
	})
}

// --- Full round flow ---

func TestCrazyEights_FullRoundFlow(t *testing.T) {
	g := newTestCrazyEights()
	g.Reset()

	// Reset all players with known cards
	topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupCrazyEightsPlayPhase(g, 0, topCard)

	// Player 0 (human): one card matching
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

	// Player 1: one card
	g.GetPlayer(1).Reset()
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))

	// Player 2: one card
	g.GetPlayer(2).Reset()
	g.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

	// Player 3: one card
	g.GetPlayer(3).Reset()
	g.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

	// Human plays last card
	err := g.PlayerPlay(0)
	require.NoError(t, err)
	assert.Equal(t, domain.CrazyEightsPhaseRoundEnd, g.GetPhase())
	assert.True(t, g.GetPlayer(0).GetIsFinished())

	// Score the round
	g.ScoreRound()
	// Winner (player 0) gets sum of remaining: 7+9+10 = 26
	assert.Equal(t, 26, g.GetPlayer(0).GetCumulativeScore())

	// Next round
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.CrazyEightsPhasePlay, g.GetPhase())
}

// --- Action log ---

func TestCrazyEights_ActionLog(t *testing.T) {
	g := newTestCrazyEights()
	g.Reset()
	topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupCrazyEightsPlayPhase(g, 0, topCard)

	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))

	err := g.PlayerPlay(0)
	require.NoError(t, err)

	log := g.GetActionLog()
	assert.NotEmpty(t, log)
	assert.Equal(t, "play", log[0].ActionType)
	assert.Equal(t, 0, log[0].PlayerIdx)
}

// --- CPU single valid card (len(validIndices) == 1 branch) ---

func TestCrazyEights_CpuSingleValidCard(t *testing.T) {
	g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyNormal)
	g.Reset()
	topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupCrazyEightsPlayPhase(g, 1, topCard)

	g.GetPlayer(1).Reset()
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))  // only valid card
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // no match

	g.CpuPlay()
	// Should play the only valid card (Spade 3)
	assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, 10, g.GetPlayer(1).GetCard(0).GetValue())
}

// --- cpuSelectSuitSmart with no non-8 cards ---

func TestCrazyEights_CpuChooseSuitSmartAllEights(t *testing.T) {
	g := newTestCrazyEightsWithDifficulty(domain.CrazyEightsCpuDifficultyNormal)
	g.Reset()
	g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
	g.SetCurrentPlayerIdx(1)

	// All remaining cards are 8s (excluded from suit count)
	g.GetPlayer(1).Reset()
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

	g.CpuChooseSuit()
	// With no non-8 cards, all suit counts are 0, bestSuit defaults to Spade (1)
	assert.Equal(t, domain.CardDesignSpade, g.GetChosenSuit())
}

// **Hearts / Spades はサーバー計算の理由付きヒントを返すのに、CrazyEights には
// ドメインの GetHint すら無かった (#4737)。**推奨手は CPU の最善手選択をそのまま
// 使う。別ロジックを書くと「CPU は選ばない手を人間に勧める」ことになる。
func TestCrazyEights_GetHint(t *testing.T) {
	setup := func(t *testing.T) *domain.CrazyEights {
		t.Helper()
		g := domain.NewDefaultCrazyEights()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		return g
	}

	t.Run("recommends a card the rules actually allow", func(t *testing.T) {
		g := setup(t)
		hint := g.GetHint()
		if hint == nil {
			t.Fatal("配り直後の人間の手番ではヒントが出る")
		}
		if hint.CardIndex == nil {
			t.Fatal("プレイフェーズでは CardIndex が入る")
		}
		// **勧めた札が本当に出せること。**別ロジックで選ぶと、出せない札を
		// 勧めてしまう。ドメインの合法手判定で裏を取る。
		// 本番の入口で裏を取る。合法手判定をテスト側に写すと、それがもう1つの
		// 実装になってしまう。
		if err := g.PlayerPlay(*hint.CardIndex); err != nil {
			t.Errorf("推奨札 (index %d) が実際には出せなかった: %v", *hint.CardIndex, err)
		}
		if hint.Reason == "" {
			t.Error("理由キーが空")
		}
	})

	t.Run("recommends a suit during the choose-suit phase", func(t *testing.T) {
		g := setup(t)
		g.SetPhase(domain.CrazyEightsPhaseChooseSuit)

		hint := g.GetHint()
		if hint == nil || hint.Suit == nil {
			t.Fatal("スート選択フェーズでは Suit が入る")
		}
		if hint.CardIndex != nil {
			t.Error("スート選択フェーズで CardIndex は入らない")
		}
		if *hint.Suit < domain.CardDesignSpade || *hint.Suit > domain.CardDesignDiamond {
			t.Errorf("Suit = %d はスートの範囲外", *hint.Suit)
		}
	})

	// **CPU の手番では出さない。**相手の手札を見て助言することになる。
	t.Run("no hint on a CPU turn", func(t *testing.T) {
		g := setup(t)
		g.SetCurrentPlayerIdx(1)
		if g.GetHint() != nil {
			t.Error("CPU の手番ではヒントを出さない")
		}
	})

	// プレイでもスート選択でもないフェーズでは出さない。
	t.Run("no hint outside the playable phases", func(t *testing.T) {
		g := setup(t)
		g.SetPhase(domain.CrazyEightsPhaseRoundEnd)
		if g.GetHint() != nil {
			t.Error("ラウンド終了フェーズではヒントを出さない")
		}
	})
}
