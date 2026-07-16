//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestMacau() *domain.Macau {
	players := []*domain.MacauPlayer{
		domain.NewMacauPlayer(true),
		domain.NewMacauPlayer(false),
		domain.NewMacauPlayer(false),
		domain.NewMacauPlayer(false),
	}
	return domain.NewMacau(domain.NewTrumpCards(0), players, domain.DefaultMacauConfig())
}

func newTestMacauWithDifficulty(d domain.MacauCpuDifficulty) *domain.Macau {
	players := []*domain.MacauPlayer{
		domain.NewMacauPlayer(true),
		domain.NewMacauPlayer(false),
		domain.NewMacauPlayer(false),
		domain.NewMacauPlayer(false),
	}
	cfg := domain.DefaultMacauConfig()
	cfg.CpuDifficulty = d
	return domain.NewMacau(domain.NewTrumpCards(0), players, cfg)
}

// setupMacauPlayPhase sets up a deterministic play state.
func setupMacauPlayPhase(g *domain.Macau, currentIdx int, topCard *domain.Card) {
	g.SetPhase(domain.MacauPhasePlay)
	g.SetCurrentPlayerIdx(currentIdx)
	g.SetDiscardPile([]*domain.Card{topCard})
	g.SetChosenSuit(-1)
	g.SetPenaltyDrawCount(0)
	g.SetDirection(1)
}

func TestNewMacau(t *testing.T) {
	g := newTestMacau()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.Equal(t, -1, g.GetChosenSuit())
	assert.Equal(t, 1, g.GetDirection())
	assert.False(t, g.GetGameEndFlag())
}

func TestNewDefaultMacau(t *testing.T) {
	g := domain.NewDefaultMacau()
	assert.NotNil(t, g)
	assert.Equal(t, domain.MacauPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	for i := 1; i < g.GetPlayerCnt(); i++ {
		assert.False(t, g.GetPlayer(i).GetIsHuman(), "player %d should be CPU", i)
	}
}

func TestMacau_Reset(t *testing.T) {
	g := newTestMacau()
	g.Reset()

	assert.Equal(t, domain.MacauPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, -1, g.GetChosenSuit())
	assert.Equal(t, 0, g.GetPenaltyDrawCount())
	assert.Equal(t, 1, g.GetDirection())

	for i := 0; i < 4; i++ {
		assert.Equal(t, 5, g.GetPlayer(i).GetCardsSize())
		assert.False(t, g.GetPlayer(i).GetHasDeclared())
	}
	assert.Len(t, g.GetDiscardPile(), 1)
	assert.Equal(t, 31, g.GetDrawPileCount()) // 52 - 20 - 1
	assert.Nil(t, g.GetActionLog())
}

func TestMacau_Getters(t *testing.T) {
	g := newTestMacau()
	g.Reset()

	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.NotNil(t, g.GetPlayer(0))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(4))

	assert.Equal(t, domain.MacauCpuDifficultyNormal, g.GetConfig().CpuDifficulty)
	g.SetConfig(domain.MacauConfig{CpuDifficulty: domain.MacauCpuDifficultyHard, PointLimit: 100})
	assert.Equal(t, domain.MacauCpuDifficultyHard, g.GetConfig().CpuDifficulty)

	g.SetPenaltyDrawCount(6)
	assert.Equal(t, 6, g.GetPenaltyDrawCount())
	g.SetDirection(-1)
	assert.Equal(t, -1, g.GetDirection())
	g.SetRoundNumber(3)
	assert.Equal(t, 3, g.GetRoundNumber())
}

func TestMacau_IsHumanTurn(t *testing.T) {
	g := newTestMacau()
	g.Reset()

	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(-1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(4)
	assert.False(t, g.IsHumanTurn())
}

func TestMacau_GetDiscardTop(t *testing.T) {
	g := newTestMacau()
	g.SetDiscardPile(nil)
	assert.Nil(t, g.GetDiscardTop())
	card := domain.NewCard(domain.CardDesignSpade, 5, false)
	g.SetDiscardPile([]*domain.Card{card})
	assert.Equal(t, card, g.GetDiscardTop())
}

func TestMacau_IsValidPlay(t *testing.T) {
	g := newTestMacau()
	g.Reset()
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	g.SetChosenSuit(-1)
	g.SetPenaltyDrawCount(0)

	assert.True(t, g.IsValidPlay(domain.NewCard(domain.CardDesignSpade, 3, false)), "same suit is playable")
	assert.True(t, g.IsValidPlay(domain.NewCard(domain.CardDesignHeart, 5, false)), "same rank is playable")
	assert.False(t, g.IsValidPlay(domain.NewCard(domain.CardDesignHeart, 7, false)), "different suit and rank is not playable")
}

func TestMacau_GetValidPlayIndices(t *testing.T) {
	g := newTestMacau()
	g.Reset()
	topCard := domain.NewCard(domain.CardDesignSpade, 5, false)
	g.SetDiscardPile([]*domain.Card{topCard})
	g.SetChosenSuit(-1)
	g.SetPenaltyDrawCount(0)

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

func TestMacau_NextRound(t *testing.T) {
	t.Run("works in RoundEnd phase", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseRoundEnd)
		g.SetRoundNumber(1)
		g.SetDirection(-1)
		g.SetPenaltyDrawCount(4)

		g.NextRound()

		assert.Equal(t, domain.MacauPhasePlay, g.GetPhase())
		assert.Equal(t, 2, g.GetRoundNumber())
		assert.Equal(t, 0, g.GetCurrentPlayerIdx())
		assert.Equal(t, 1, g.GetDirection())
		assert.Equal(t, 0, g.GetPenaltyDrawCount())
		for i := 0; i < 4; i++ {
			assert.Equal(t, 5, g.GetPlayer(i).GetCardsSize())
		}
	})

	t.Run("no-op in Play phase", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		round := g.GetRoundNumber()
		g.SetPhase(domain.MacauPhasePlay)
		g.NextRound()
		assert.Equal(t, round, g.GetRoundNumber())
	})
}

// --- PlayerPlay ---

func TestMacau_PlayerPlay(t *testing.T) {
	t.Run("valid suit match", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(g.GetDiscardPile()))
	})

	t.Run("play wild 8 enters choose suit", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, domain.MacauPhaseChooseSuit, g.GetPhase())
	})

	t.Run("play last card ends round", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, domain.MacauPhaseRoundEnd, g.GetPhase())
	})

	t.Run("reaching one card enters must declare", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))

		err := g.PlayerPlay(0) // plays spade 3, one card remains
		assert.NoError(t, err)
		assert.Equal(t, domain.MacauPhaseMustDeclare, g.GetPhase())
	})

	t.Run("playing 2 sets penalty", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

		err := g.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, domain.MacauDrawTwoAmount, g.GetPenaltyDrawCount())
	})

	t.Run("during penalty cannot play non-2", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 2, false))
		g.SetPenaltyDrawCount(2)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // suit match but not a 2

		err := g.PlayerPlay(0)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})

	t.Run("game ended error", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.GetPlayer(0).SetCumulativeScore(300)
		g.SetPhase(domain.MacauPhaseRoundEnd)
		g.GetPlayer(0).Reset()
		g.ScoreRound()
		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase error", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseChooseSuit)
		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn error", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 1, domain.NewCard(domain.CardDesignSpade, 5, false))
		err := g.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("card index out of range", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		assert.True(t, errors.Is(g.PlayerPlay(-1), domain.ErrInvalidCard))
		assert.True(t, errors.Is(g.PlayerPlay(100), domain.ErrInvalidCard))
	})

	t.Run("invalid play no match", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		err := g.PlayerPlay(0)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

// --- PlayerChooseSuit ---

func TestMacau_PlayerChooseSuit(t *testing.T) {
	t.Run("valid suit advances turn", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseChooseSuit)
		g.SetCurrentPlayerIdx(0)
		g.SetDirection(1)
		// player has 2 cards so no declaration triggered
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))

		err := g.PlayerChooseSuit(domain.CardDesignHeart)
		assert.NoError(t, err)
		assert.Equal(t, domain.CardDesignHeart, g.GetChosenSuit())
		assert.Equal(t, domain.MacauPhasePlay, g.GetPhase())
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	})

	t.Run("8 as last card requires declaration after choosing", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseChooseSuit)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false)) // one card left

		err := g.PlayerChooseSuit(domain.CardDesignHeart)
		assert.NoError(t, err)
		assert.Equal(t, domain.MacauPhaseMustDeclare, g.GetPhase())
	})

	t.Run("errors", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhasePlay)
		assert.ErrorIs(t, g.PlayerChooseSuit(domain.CardDesignHeart), domain.ErrWrongPhase)

		g.SetPhase(domain.MacauPhaseChooseSuit)
		g.SetCurrentPlayerIdx(1)
		assert.ErrorIs(t, g.PlayerChooseSuit(domain.CardDesignHeart), domain.ErrNotHumanTurn)

		g.SetCurrentPlayerIdx(0)
		assert.True(t, errors.Is(g.PlayerChooseSuit(0), domain.ErrInvalidPlay))
		assert.True(t, errors.Is(g.PlayerChooseSuit(5), domain.ErrInvalidPlay))
	})
}

// --- PlayerDraw ---

func TestMacau_PlayerDraw(t *testing.T) {
	t.Run("draws from draw pile", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

		err := g.PlayerDraw()
		assert.NoError(t, err)
		assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())
	})

	t.Run("penalty draw takes the whole stack", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 2, false))
		g.SetPenaltyDrawCount(4)
		g.SetDrawPile([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
		})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))

		err := g.PlayerDraw()
		assert.NoError(t, err)
		assert.Equal(t, 5, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, 0, g.GetPenaltyDrawCount())
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	})

	t.Run("errors", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseChooseSuit)
		assert.ErrorIs(t, g.PlayerDraw(), domain.ErrWrongPhase)

		g.SetPhase(domain.MacauPhasePlay)
		g.SetCurrentPlayerIdx(1)
		assert.ErrorIs(t, g.PlayerDraw(), domain.ErrNotHumanTurn)
	})
}

// --- Declaration ---

func TestMacau_Declaration(t *testing.T) {
	t.Run("PlayerDeclare advances turn", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseMustDeclare)
		g.SetCurrentPlayerIdx(0)
		g.SetDirection(1)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		err := g.PlayerDeclare()
		assert.NoError(t, err)
		assert.True(t, g.GetPlayer(0).GetHasDeclared())
		assert.Equal(t, domain.MacauPhasePlay, g.GetPhase())
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	})

	t.Run("PlayerSkipDeclare draws penalty", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseMustDeclare)
		g.SetCurrentPlayerIdx(0)
		g.SetDirection(1)
		g.SetDrawPile([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignHeart, 4, false),
		})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

		err := g.PlayerSkipDeclare()
		assert.NoError(t, err)
		assert.Equal(t, 3, g.GetPlayer(0).GetCardsSize()) // 1 + 2 penalty
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	})

	t.Run("declaration errors", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhasePlay)
		assert.ErrorIs(t, g.PlayerDeclare(), domain.ErrWrongPhase)
		assert.ErrorIs(t, g.PlayerSkipDeclare(), domain.ErrWrongPhase)

		g.SetPhase(domain.MacauPhaseMustDeclare)
		g.SetCurrentPlayerIdx(1)
		assert.ErrorIs(t, g.PlayerDeclare(), domain.ErrNotHumanTurn)
		assert.ErrorIs(t, g.PlayerSkipDeclare(), domain.ErrNotHumanTurn)
	})

	t.Run("CpuDeclare normal declares", func(t *testing.T) {
		g := newTestMacauWithDifficulty(domain.MacauCpuDifficultyNormal)
		g.Reset()
		g.SetPhase(domain.MacauPhaseMustDeclare)
		g.SetCurrentPlayerIdx(1)
		g.SetDirection(1)
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		g.CpuDeclare()
		assert.True(t, g.GetPlayer(1).GetHasDeclared())
		assert.Equal(t, domain.MacauPhasePlay, g.GetPhase())
	})

	t.Run("CpuDeclare no-op when human turn", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseMustDeclare)
		g.SetCurrentPlayerIdx(0)
		g.CpuDeclare() // no-op
		assert.Equal(t, domain.MacauPhaseMustDeclare, g.GetPhase())
	})
}

// --- CpuPlay magic cards & penalty ---

func TestMacau_CpuPlay(t *testing.T) {
	t.Run("CPU plays a valid card", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 1, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

		g.CpuPlay()
		assert.Equal(t, 2, g.GetPlayer(1).GetCardsSize())
	})

	t.Run("CPU draws when no valid play", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 1, domain.NewCard(domain.CardDesignSpade, 5, false))
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 3, false)})

		g.CpuPlay()
		assert.Equal(t, 2, g.GetPlayer(1).GetCardsSize())
	})

	t.Run("CPU takes penalty when no 2 to stack", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		setupMacauPlayPhase(g, 1, domain.NewCard(domain.CardDesignSpade, 2, false))
		g.SetPenaltyDrawCount(2)
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false)) // no 2
		g.SetDrawPile([]*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignDiamond, 4, false),
		})

		g.CpuPlay()
		assert.Equal(t, 3, g.GetPlayer(1).GetCardsSize()) // 1 + 2 penalty
		assert.Equal(t, 0, g.GetPenaltyDrawCount())
	})

	t.Run("no-op when wrong phase or human", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseChooseSuit)
		g.SetCurrentPlayerIdx(1)
		g.CpuPlay()
		g.SetPhase(domain.MacauPhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.CpuPlay()
	})
}

// --- CpuChooseSuit ---

func TestMacau_CpuChooseSuit(t *testing.T) {
	t.Run("CPU chooses most common suit", func(t *testing.T) {
		g := newTestMacauWithDifficulty(domain.MacauCpuDifficultyNormal)
		g.Reset()
		g.SetPhase(domain.MacauPhaseChooseSuit)
		g.SetCurrentPlayerIdx(1)
		g.SetDirection(1)
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

		g.CpuChooseSuit()
		assert.Equal(t, domain.CardDesignHeart, g.GetChosenSuit())
		assert.Equal(t, domain.MacauPhasePlay, g.GetPhase())
	})

	t.Run("no-op when human turn", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseChooseSuit)
		g.SetCurrentPlayerIdx(0)
		g.CpuChooseSuit()
	})
}

// --- ScoreRound ---

func TestMacau_ScoreRound(t *testing.T) {
	t.Run("scores remaining cards to winner", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseRoundEnd)
		g.GetPlayer(0).Reset() // winner
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 8, false)) // 50
		g.GetPlayer(2).Reset()
		g.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignHeart, 13, false)) // 10
		g.GetPlayer(3).Reset()
		g.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignClover, 1, false)) // 1

		g.ScoreRound()
		assert.Equal(t, 61, g.GetPlayer(0).GetCumulativeScore())
	})

	t.Run("no-op wrong phase", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhasePlay)
		g.ScoreRound()
	})

	t.Run("no winner found", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		g.SetPhase(domain.MacauPhaseRoundEnd)
		for i := 0; i < 4; i++ {
			g.GetPlayer(i).Reset()
			g.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		}
		g.ScoreRound()
		for i := 0; i < 4; i++ {
			assert.Equal(t, 0, g.GetPlayer(i).GetCumulativeScore())
		}
	})

	t.Run("game ends at point limit", func(t *testing.T) {
		g := newTestMacau()
		g.Reset()
		cfg := domain.DefaultMacauConfig()
		cfg.PointLimit = 50
		g.SetConfig(cfg)
		g.SetPhase(domain.MacauPhaseRoundEnd)
		g.GetPlayer(0).Reset()
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 8, false)) // 50
		g.GetPlayer(2).Reset()
		g.GetPlayer(3).Reset()

		g.ScoreRound()
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, domain.MacauPhaseGameEnd, g.GetPhase())
		assert.Equal(t, 0, g.GetWinnerIdx())
	})
}

// --- Full flows ---

func TestMacau_SkipFlow(t *testing.T) {
	g := newTestMacau()
	g.Reset()
	setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false)) // skip
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	err := g.PlayerPlay(0)
	require.NoError(t, err)
	assert.Equal(t, 2, g.GetCurrentPlayerIdx()) // player 1 skipped
}

func TestMacau_ReverseFlow(t *testing.T) {
	g := newTestMacau()
	g.Reset()
	setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 11, false)) // J reverse
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	err := g.PlayerPlay(0)
	require.NoError(t, err)
	assert.Equal(t, -1, g.GetDirection())
	assert.Equal(t, 3, g.GetCurrentPlayerIdx()) // reverse to player 3
}

func TestMacau_FullRoundFlow(t *testing.T) {
	g := newTestMacau()
	g.Reset()
	setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))

	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // last card
	for i := 1; i < 4; i++ {
		g.GetPlayer(i).Reset()
		g.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	}

	err := g.PlayerPlay(0)
	require.NoError(t, err)
	assert.Equal(t, domain.MacauPhaseRoundEnd, g.GetPhase())
	assert.True(t, g.GetPlayer(0).GetIsFinished())

	g.ScoreRound()
	assert.Equal(t, 27, g.GetPlayer(0).GetCumulativeScore()) // 9+9+9

	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.MacauPhasePlay, g.GetPhase())
}

// --- Action log ---

func TestMacau_ActionLog(t *testing.T) {
	g := newTestMacau()
	g.Reset()
	setupMacauPlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))

	err := g.PlayerPlay(0)
	require.NoError(t, err)
	log := g.GetActionLog()
	assert.NotEmpty(t, log)
	assert.Equal(t, "play", log[0].ActionType)
}

// --- JSON round-trip ---

func TestMacau_JSONRoundTrip(t *testing.T) {
	g := newTestMacau()
	g.Reset()
	g.SetPenaltyDrawCount(4)
	g.SetDirection(-1)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	restored := newTestMacau()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPenaltyDrawCount(), restored.GetPenaltyDrawCount())
	assert.Equal(t, g.GetDirection(), restored.GetDirection())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
}

// macauTamperedJSON marshals a fresh, valid 4-player game and overwrites one
// wire field, so UnmarshalJSON validation branches can be exercised against an
// otherwise-valid payload.
func macauTamperedJSON(t *testing.T, key, raw string) []byte {
	t.Helper()
	g := newTestMacau()
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &m))
	m[key] = json.RawMessage(raw)
	out, err := json.Marshal(m)
	require.NoError(t, err)
	return out
}

func TestMacau_UnmarshalValidation(t *testing.T) {
	t.Run("normalises direction 0 to 1", func(t *testing.T) {
		var g domain.Macau
		require.NoError(t, json.Unmarshal(macauTamperedJSON(t, "dr", "0"), &g))
		assert.Equal(t, 1, g.GetDirection())
	})

	t.Run("normalises out-of-range direction to 1", func(t *testing.T) {
		var g domain.Macau
		require.NoError(t, json.Unmarshal(macauTamperedJSON(t, "dr", "99"), &g))
		assert.Equal(t, 1, g.GetDirection())
	})

	t.Run("rejects wrong player count", func(t *testing.T) {
		var g domain.Macau
		assert.Error(t, json.Unmarshal([]byte(`{"pl":[]}`), &g))
	})

	t.Run("rejects out-of-range currentPlayerIdx", func(t *testing.T) {
		var g domain.Macau
		assert.Error(t, json.Unmarshal(macauTamperedJSON(t, "ci", "9"), &g))
	})

	t.Run("rejects out-of-range phase", func(t *testing.T) {
		var g domain.Macau
		assert.Error(t, json.Unmarshal(macauTamperedJSON(t, "ps", "99"), &g))
	})

	t.Run("caps oversized penaltyDrawCount", func(t *testing.T) {
		var g domain.Macau
		require.NoError(t, json.Unmarshal(macauTamperedJSON(t, "pd", "5000"), &g))
		assert.LessOrEqual(t, g.GetPenaltyDrawCount(), 1000)
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		var g domain.Macau
		assert.Error(t, json.Unmarshal([]byte(`{`), &g))
	})
}
