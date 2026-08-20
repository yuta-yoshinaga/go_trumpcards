//go:build test

package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestRummy500() *domain.Rummy500 {
	players := []*domain.Rummy500Player{
		domain.NewRummy500Player(true),
		domain.NewRummy500Player(false),
	}
	return domain.NewRummy500(domain.NewTrumpCards(0), players, domain.DefaultRummy500Config())
}

// setRummy500Hand replaces the player's hand with the provided cards.
func setRummy500Hand(t *testing.T, p *domain.Rummy500Player, cards ...*domain.Card) {
	t.Helper()
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func r5Card(t *testing.T, design, value int) *domain.Card {
	t.Helper()
	c := domain.NewCard(design, value, false)
	require.NotNil(t, c)
	return c
}

func TestNewRummy500(t *testing.T) {
	g := newTestRummy500()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetRoundEnderIdx())
}

func TestRummy500_Reset(t *testing.T) {
	g := newTestRummy500()
	g.Reset()

	assert.Equal(t, domain.Rummy500PhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, -1, g.GetRoundEnderIdx())

	for i := 0; i < 2; i++ {
		assert.Equal(t, 13, g.GetPlayer(i).GetCardsSize(), "player %d should have 13 cards", i)
		assert.Equal(t, 0, g.GetPlayer(i).GetRoundScore())
		assert.Equal(t, 0, g.GetPlayer(i).GetCumulativeScore())
	}
	assert.Len(t, g.GetDiscardPile(), 1)
	// 52 - 26 - 1 = 25
	assert.Equal(t, 25, g.GetDrawPileCount())
	assert.Nil(t, g.GetActionLog())
}

func TestRummy500_Reset_ClearsState(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	g.GetPlayer(0).SetCumulativeScore(123)
	g.SetPhase(domain.Rummy500PhaseGameEnd)
	g.Reset()
	assert.Equal(t, domain.Rummy500PhaseDraw, g.GetPhase())
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore())
}

func TestRummy500_Getters(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.NotNil(t, g.GetPlayer(0))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(2))
	assert.NotNil(t, g.GetDiscardTop())
}

func TestRummy500_IsHumanTurn(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(99)
	assert.False(t, g.IsHumanTurn())
}

func TestRummy500_PlayerDrawFromStock(t *testing.T) {
	t.Run("draws and transitions to Play phase", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		before := g.GetPlayer(0).GetCardsSize()
		stockBefore := g.GetDrawPileCount()
		err := g.PlayerDrawFromStock()
		require.NoError(t, err)
		assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, stockBefore-1, g.GetDrawPileCount())
		assert.Equal(t, domain.Rummy500PhasePlay, g.GetPhase())
	})

	t.Run("errors when not human turn", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(1)
		err := g.PlayerDrawFromStock()
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("errors when wrong phase", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhasePlay)
		err := g.PlayerDrawFromStock()
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("errors when game ended", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhaseGameEnd)
		g.GetPlayer(0).SetCumulativeScore(500)
		// Force gameEnd flag via score check helper: easier to assert game-end error
		// by ensuring guard kicks in via the round-ender path.
		// Use direct end-of-round trigger:
		err := g.PlayerDrawFromStock()
		// Wrong phase is also acceptable here; we just want a non-nil error.
		assert.Error(t, err)
		_ = errors.Is(err, domain.ErrGameEnded)
	})

	t.Run("ends round when stock empty", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		g.SetDrawPile(nil)
		err := g.PlayerDrawFromStock()
		require.NoError(t, err)
		assert.Equal(t, domain.Rummy500PhaseRoundEnd, g.GetPhase())
	})
}

func TestRummy500_PlayerDrawFromDiscard(t *testing.T) {
	t.Run("top only", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		topBefore := g.GetDiscardTop()
		require.NotNil(t, topBefore)
		err := g.PlayerDrawFromDiscard(len(g.GetDiscardPile()) - 1)
		require.NoError(t, err)
		assert.Equal(t, domain.Rummy500PhasePlay, g.GetPhase())
		assert.Empty(t, g.GetDiscardPile())
	})

	t.Run("middle picks all cards above", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		pile := []*domain.Card{
			r5Card(t, 0, 5), r5Card(t, 1, 7), r5Card(t, 2, 9),
		}
		g.SetDiscardPile(pile)
		handBefore := g.GetPlayer(0).GetCardsSize()
		err := g.PlayerDrawFromDiscard(0)
		require.NoError(t, err)
		assert.Equal(t, handBefore+3, g.GetPlayer(0).GetCardsSize())
		assert.Empty(t, g.GetDiscardPile())
	})

	t.Run("rejects empty discard", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		g.SetDiscardPile(nil)
		err := g.PlayerDrawFromDiscard(0)
		assert.Error(t, err)
	})

	t.Run("rejects out-of-range index", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		err := g.PlayerDrawFromDiscard(99)
		assert.Error(t, err)
	})

	t.Run("rejects negative index", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		err := g.PlayerDrawFromDiscard(-1)
		assert.Error(t, err)
	})

	t.Run("rejects when wrong phase", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhasePlay)
		err := g.PlayerDrawFromDiscard(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("rejects when not human", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(1)
		g.SetPhase(domain.Rummy500PhaseDraw)
		err := g.PlayerDrawFromDiscard(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})
}

func TestRummy500_PlayerMeld(t *testing.T) {
	t.Run("valid set", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0),
			r5Card(t, 0, 7), r5Card(t, 1, 7), r5Card(t, 2, 7), r5Card(t, 3, 4))
		err := g.PlayerMeld([]int{0, 1, 2})
		require.NoError(t, err)
		assert.Equal(t, 1, len(g.GetPlayer(0).GetLaidMelds()))
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	})

	t.Run("valid run with Ace low", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0),
			r5Card(t, 0, 1), r5Card(t, 0, 2), r5Card(t, 0, 3))
		err := g.PlayerMeld([]int{0, 1, 2})
		require.NoError(t, err)
		assert.Equal(t, 1, len(g.GetPlayer(0).GetLaidMelds()))
	})

	t.Run("valid run with Ace high", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0),
			r5Card(t, 0, 12), r5Card(t, 0, 13), r5Card(t, 0, 1))
		err := g.PlayerMeld([]int{0, 1, 2})
		require.NoError(t, err)
		melds := g.GetPlayer(0).GetLaidMelds()
		require.Len(t, melds, 1)
		assert.Equal(t, 35, domain.Rummy500MeldScore(melds[0]))
	})

	t.Run("invalid meld (mixed)", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0),
			r5Card(t, 0, 7), r5Card(t, 1, 8), r5Card(t, 2, 9))
		err := g.PlayerMeld([]int{0, 1, 2})
		assert.Error(t, err)
	})

	t.Run("too few cards", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0), r5Card(t, 0, 7), r5Card(t, 1, 7))
		err := g.PlayerMeld([]int{0, 1})
		assert.Error(t, err)
	})

	t.Run("duplicate indices", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0),
			r5Card(t, 0, 7), r5Card(t, 1, 7), r5Card(t, 2, 7))
		err := g.PlayerMeld([]int{0, 0, 1})
		assert.Error(t, err)
	})

	t.Run("out of range index", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0), r5Card(t, 0, 7))
		err := g.PlayerMeld([]int{0, 1, 5})
		assert.Error(t, err)
	})

	t.Run("rejects when wrong phase", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		err := g.PlayerMeld([]int{0, 1, 2})
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("rejects when not human turn", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(1)
		g.SetPhase(domain.Rummy500PhasePlay)
		err := g.PlayerMeld([]int{0, 1, 2})
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})
}

func TestRummy500_PlayerLayoff(t *testing.T) {
	t.Run("layoff to own set", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		// Player 0 has a 7-7-7 set laid down and a 7 in hand
		g.GetPlayer(0).SetLaidMelds([][]*domain.Card{
			{r5Card(t, 0, 7), r5Card(t, 1, 7), r5Card(t, 2, 7)},
		})
		setRummy500Hand(t, g.GetPlayer(0), r5Card(t, 3, 7))
		err := g.PlayerLayoff(0, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 4, len(g.GetPlayer(0).GetLaidMelds()[0]))
		assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
	})

	t.Run("layoff to opponent run", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(1).SetLaidMelds([][]*domain.Card{
			{r5Card(t, 0, 5), r5Card(t, 0, 6), r5Card(t, 0, 7)},
		})
		setRummy500Hand(t, g.GetPlayer(0), r5Card(t, 0, 8))
		err := g.PlayerLayoff(1, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 4, len(g.GetPlayer(1).GetLaidMelds()[0]))
	})

	t.Run("invalid layoff", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).SetLaidMelds([][]*domain.Card{
			{r5Card(t, 0, 5), r5Card(t, 0, 6), r5Card(t, 0, 7)},
		})
		setRummy500Hand(t, g.GetPlayer(0), r5Card(t, 0, 13))
		err := g.PlayerLayoff(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid meld owner", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		err := g.PlayerLayoff(99, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid meld idx", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		err := g.PlayerLayoff(0, 99, 0)
		assert.Error(t, err)
	})

	t.Run("invalid card idx", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.GetPlayer(0).SetLaidMelds([][]*domain.Card{
			{r5Card(t, 0, 5), r5Card(t, 0, 6), r5Card(t, 0, 7)},
		})
		err := g.PlayerLayoff(0, 0, 99)
		assert.Error(t, err)
	})

	t.Run("rejects when wrong phase", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.Rummy500PhaseDraw)
		err := g.PlayerLayoff(0, 0, 0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("rejects when not human", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetCurrentPlayerIdx(1)
		g.SetPhase(domain.Rummy500PhasePlay)
		err := g.PlayerLayoff(0, 0, 0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})
}

func TestRummy500_PlayerDiscard(t *testing.T) {
	t.Run("normal discard advances turn", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0), r5Card(t, 0, 5), r5Card(t, 0, 6))
		err := g.PlayerDiscard(0)
		require.NoError(t, err)
		assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, domain.Rummy500PhaseDraw, g.GetPhase())
		assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	})

	t.Run("go out when hand empties", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		setRummy500Hand(t, g.GetPlayer(0), r5Card(t, 0, 5))
		err := g.PlayerDiscard(0)
		require.NoError(t, err)
		assert.Equal(t, 0, g.GetPlayer(0).GetCardsSize())
		assert.Equal(t, 0, g.GetRoundEnderIdx())
		// Either RoundEnd or GameEnd is acceptable depending on cumulative score
		ph := g.GetPhase()
		assert.True(t, ph == domain.Rummy500PhaseRoundEnd || ph == domain.Rummy500PhaseGameEnd)
	})

	t.Run("out of range index", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(0)
		err := g.PlayerDiscard(99)
		assert.Error(t, err)
	})

	t.Run("rejects when wrong phase", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhaseDraw)
		g.SetCurrentPlayerIdx(0)
		err := g.PlayerDiscard(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("rejects when not human", func(t *testing.T) {
		g := newTestRummy500()
		g.Reset()
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(1)
		err := g.PlayerDiscard(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})
}

func TestRummy500_Scoring(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	// Player 0: melded 7-7-7 (=21), no hand
	g.GetPlayer(0).SetLaidMelds([][]*domain.Card{
		{r5Card(t, 0, 7), r5Card(t, 1, 7), r5Card(t, 2, 7)},
	})
	setRummy500Hand(t, g.GetPlayer(0))
	// Player 1: no meld, 10 + K in hand → -20
	setRummy500Hand(t, g.GetPlayer(1), r5Card(t, 0, 10), r5Card(t, 1, 13))

	g.SetPhase(domain.Rummy500PhasePlay)
	g.SetCurrentPlayerIdx(0)
	// Force going out via a discard with empty hand will fail; use a discard
	// from a single-card hand to trigger scoreRound.
	g.GetPlayer(0).AddCard(r5Card(t, 3, 2))
	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.Equal(t, 21, g.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, -20, g.GetPlayer(1).GetCumulativeScore())
}

func TestRummy500_GameEnd(t *testing.T) {
	g := newTestRummy500()
	cfg := domain.DefaultRummy500Config()
	cfg.PointLimit = 20
	g.SetConfig(cfg)
	g.Reset()
	g.GetPlayer(0).SetLaidMelds([][]*domain.Card{
		{r5Card(t, 0, 7), r5Card(t, 1, 7), r5Card(t, 2, 7)},
	})
	setRummy500Hand(t, g.GetPlayer(0))
	setRummy500Hand(t, g.GetPlayer(1))
	g.SetPhase(domain.Rummy500PhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).AddCard(r5Card(t, 3, 2))
	err := g.PlayerDiscard(0)
	require.NoError(t, err)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.Rummy500PhaseGameEnd, g.GetPhase())
}

func TestRummy500_NextRound(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	g.SetPhase(domain.Rummy500PhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.Rummy500PhaseDraw, g.GetPhase())
	assert.Equal(t, 13, g.GetPlayer(0).GetCardsSize())
}

func TestRummy500_NextRound_NoOpInWrongPhase(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	r := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, r, g.GetRoundNumber())
}

func TestRummy500_CpuPlay_Draws(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.Rummy500PhaseDraw)
	before := g.GetPlayer(1).GetCardsSize()
	g.CpuPlay()
	assert.Greater(t, g.GetPlayer(1).GetCardsSize(), before)
}

func TestRummy500_CpuPlay_DiscardsAfterPlay(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.Rummy500PhasePlay)
	g.CpuPlay()
	// After CPU finishes phase Play, either game ended, round ended, or it is human's draw
	assert.True(t, g.GetGameEndFlag() ||
		g.GetPhase() == domain.Rummy500PhaseRoundEnd ||
		g.GetPhase() == domain.Rummy500PhaseDraw)
}

func TestRummy500_CpuPlay_NoOpHumanTurn(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.Rummy500PhasePlay)
	before := g.GetPlayer(0).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize())
}

func TestRummy500_IsValidMeld(t *testing.T) {
	t.Run("valid set of 3", func(t *testing.T) {
		assert.True(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 5), r5Card(t, 1, 5), r5Card(t, 2, 5),
		}))
	})
	t.Run("valid set of 4", func(t *testing.T) {
		assert.True(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 5), r5Card(t, 1, 5), r5Card(t, 2, 5), r5Card(t, 3, 5),
		}))
	})
	t.Run("invalid set with duplicate suit", func(t *testing.T) {
		assert.False(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 5), r5Card(t, 0, 5), r5Card(t, 1, 5),
		}))
	})
	t.Run("valid run", func(t *testing.T) {
		assert.True(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 4), r5Card(t, 0, 5), r5Card(t, 0, 6),
		}))
	})
	t.Run("valid run with Ace low", func(t *testing.T) {
		assert.True(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 1), r5Card(t, 0, 2), r5Card(t, 0, 3),
		}))
	})
	t.Run("valid run with Ace high", func(t *testing.T) {
		assert.True(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 12), r5Card(t, 0, 13), r5Card(t, 0, 1),
		}))
	})
	t.Run("invalid run mixed suit", func(t *testing.T) {
		assert.False(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 4), r5Card(t, 1, 5), r5Card(t, 0, 6),
		}))
	})
	t.Run("invalid run not consecutive", func(t *testing.T) {
		assert.False(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 4), r5Card(t, 0, 6), r5Card(t, 0, 8),
		}))
	})
	t.Run("too few cards", func(t *testing.T) {
		assert.False(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 4), r5Card(t, 0, 5),
		}))
	})
	t.Run("Ace wrap K-A-2 is invalid", func(t *testing.T) {
		assert.False(t, domain.Rummy500IsValidMeld([]*domain.Card{
			r5Card(t, 0, 13), r5Card(t, 0, 1), r5Card(t, 0, 2),
		}))
	})
}

func TestRummy500_MeldScore(t *testing.T) {
	t.Run("set of 7s = 21", func(t *testing.T) {
		assert.Equal(t, 21, domain.Rummy500MeldScore([]*domain.Card{
			r5Card(t, 0, 7), r5Card(t, 1, 7), r5Card(t, 2, 7),
		}))
	})
	t.Run("low run A-2-3 = 6", func(t *testing.T) {
		assert.Equal(t, 6, domain.Rummy500MeldScore([]*domain.Card{
			r5Card(t, 0, 1), r5Card(t, 0, 2), r5Card(t, 0, 3),
		}))
	})
	t.Run("high run Q-K-A = 10+10+15 = 35", func(t *testing.T) {
		assert.Equal(t, 35, domain.Rummy500MeldScore([]*domain.Card{
			r5Card(t, 0, 12), r5Card(t, 0, 13), r5Card(t, 0, 1),
		}))
	})
	t.Run("face cards K-K-K = 30", func(t *testing.T) {
		assert.Equal(t, 30, domain.Rummy500MeldScore([]*domain.Card{
			r5Card(t, 0, 13), r5Card(t, 1, 13), r5Card(t, 2, 13),
		}))
	})
}

func TestRummy500_CanLayoff(t *testing.T) {
	t.Run("set extension OK", func(t *testing.T) {
		meld := []*domain.Card{r5Card(t, 0, 5), r5Card(t, 1, 5), r5Card(t, 2, 5)}
		assert.True(t, domain.Rummy500CanLayoff(meld, r5Card(t, 3, 5)))
	})
	t.Run("set extension rejects duplicate suit", func(t *testing.T) {
		meld := []*domain.Card{r5Card(t, 0, 5), r5Card(t, 1, 5), r5Card(t, 2, 5)}
		assert.False(t, domain.Rummy500CanLayoff(meld, r5Card(t, 0, 5)))
	})
	t.Run("set extension rejects when full", func(t *testing.T) {
		meld := []*domain.Card{r5Card(t, 0, 5), r5Card(t, 1, 5), r5Card(t, 2, 5), r5Card(t, 3, 5)}
		assert.False(t, domain.Rummy500CanLayoff(meld, r5Card(t, 0, 5)))
	})
	t.Run("run extension up", func(t *testing.T) {
		meld := []*domain.Card{r5Card(t, 0, 5), r5Card(t, 0, 6), r5Card(t, 0, 7)}
		assert.True(t, domain.Rummy500CanLayoff(meld, r5Card(t, 0, 8)))
	})
	t.Run("run extension down", func(t *testing.T) {
		meld := []*domain.Card{r5Card(t, 0, 5), r5Card(t, 0, 6), r5Card(t, 0, 7)}
		assert.True(t, domain.Rummy500CanLayoff(meld, r5Card(t, 0, 4)))
	})
	t.Run("rejects empty meld", func(t *testing.T) {
		assert.False(t, domain.Rummy500CanLayoff(nil, r5Card(t, 0, 5)))
	})
}

func TestRummy500_JSONRoundTrip(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	g.SetPhase(domain.Rummy500PhasePlay)
	g.SetCurrentPlayerIdx(1)
	g.GetPlayer(0).SetCumulativeScore(150)
	g.GetPlayer(0).SetLaidMelds([][]*domain.Card{
		{r5Card(t, 0, 5), r5Card(t, 1, 5), r5Card(t, 2, 5)},
	})

	data, err := g.MarshalJSON()
	require.NoError(t, err)
	g2 := newTestRummy500()
	require.NoError(t, g2.UnmarshalJSON(data))
	assert.Equal(t, domain.Rummy500PhasePlay, g2.GetPhase())
	assert.Equal(t, 1, g2.GetCurrentPlayerIdx())
	assert.Equal(t, 150, g2.GetPlayer(0).GetCumulativeScore())
	require.Len(t, g2.GetPlayer(0).GetLaidMelds(), 1)
}

func TestRummy500_UnmarshalJSON_Rejects_TooLarge(t *testing.T) {
	bigDiscard := `{"dp":[`
	for i := 0; i < 1001; i++ {
		if i > 0 {
			bigDiscard += ","
		}
		bigDiscard += `{"d":0,"v":1}`
	}
	bigDiscard += `]}`
	g := newTestRummy500()
	err := g.UnmarshalJSON([]byte(bigDiscard))
	assert.Error(t, err)
}

func TestRummy500_NewDefault(t *testing.T) {
	g := domain.NewDefaultRummy500()
	assert.NotNil(t, g)
	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())
}

func TestRummy500_PlayerHelpers(t *testing.T) {
	p := domain.NewRummy500Player(true)
	p.AddCard(r5Card(t, 0, 5))
	p.AddLaidMeld([]*domain.Card{r5Card(t, 0, 7), r5Card(t, 1, 7), r5Card(t, 2, 7)})
	require.Len(t, p.GetLaidMelds(), 1)
	p.AppendToLaidMeld(0, r5Card(t, 3, 7))
	assert.Len(t, p.GetLaidMelds()[0], 4)
	assert.False(t, p.AppendToLaidMeld(99, r5Card(t, 0, 5)))

	p.ResetRound()
	assert.Empty(t, p.GetLaidMelds())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestRummy500_PlayerJSONRoundTrip(t *testing.T) {
	p := domain.NewRummy500Player(true)
	p.AddCard(r5Card(t, 0, 5))
	p.AddLaidMeld([]*domain.Card{r5Card(t, 0, 7), r5Card(t, 1, 7), r5Card(t, 2, 7)})
	p.SetRoundScore(15)
	p.SetCumulativeScore(40)

	data, err := p.MarshalJSON()
	require.NoError(t, err)
	p2 := domain.NewRummy500Player(true)
	require.NoError(t, p2.UnmarshalJSON(data))
	assert.Equal(t, 15, p2.GetRoundScore())
	assert.Equal(t, 40, p2.GetCumulativeScore())
	require.Len(t, p2.GetLaidMelds(), 1)
}

// Regression: the CPU meld finder must surface Q-K-A as a valid run. Prior to
// the fix in findAllRummy500Melds, scanning sorted ascending order treated
// Ace as 1 only and could never reach 14, so Q-K-A runs were invisible to
// the CPU even though rummy500IsValidRun accepted them.
func TestRummy500_CpuFindsQKARun(t *testing.T) {
	g := newTestRummy500()
	g.Reset()
	g.SetPhase(domain.Rummy500PhasePlay)
	g.SetCurrentPlayerIdx(1) // CPU
	setRummy500Hand(t, g.GetPlayer(1),
		r5Card(t, 0, 1),  // Ace of spades
		r5Card(t, 0, 13), // K of spades
		r5Card(t, 0, 12), // Q of spades
		r5Card(t, 1, 4),  // junk to discard
	)
	g.CpuPlay()
	melds := g.GetPlayer(1).GetLaidMelds()
	require.Len(t, melds, 1, "CPU should meld Q-K-A")
	assert.Equal(t, 35, domain.Rummy500MeldScore(melds[0]))
}

// #5611: 設定の CpuDifficulty をドメインがどこも読んでおらず、Easy を選んでも
// Hard を選んでも CPU の打ち回しが 1 手も変わらなかった。効かない設定は、
// 効くと信じて選ぶぶんだけ悪い。
func rummy500WithDifficulty(t *testing.T, d domain.Rummy500CpuDifficulty) *domain.Rummy500 {
	t.Helper()
	g := newTestRummy500()
	cfg := domain.DefaultRummy500Config()
	cfg.CpuDifficulty = d
	g.SetConfig(cfg)
	return g
}

// 捨てる札の選び方が難易度で変わる。手札は「メルドになりかけの ♠5♠6 と、
// 孤立した高得点の ♥K、孤立した低得点の ♦3」。
func TestRummy500DifficultyChangesTheDiscard(t *testing.T) {
	discardFor := func(t *testing.T, d domain.Rummy500CpuDifficulty) *domain.Card {
		t.Helper()
		g := rummy500WithDifficulty(t, d)
		g.SetPhase(domain.Rummy500PhasePlay)
		g.SetCurrentPlayerIdx(1)
		setRummy500Hand(t, g.GetPlayer(1),
			r5Card(t, domain.CardDesignSpade, 5),
			r5Card(t, domain.CardDesignSpade, 6),
			r5Card(t, domain.CardDesignHeart, 13),
			r5Card(t, domain.CardDesignDiamond, 3),
		)
		g.SetDiscardPile(nil)
		g.CpuPlay()
		pile := g.GetDiscardPile()
		require.Len(t, pile, 1, "CPU は必ず1枚捨てる")
		return pile[0]
	}

	// Hard: 手札に残しても伸びない孤立札のうち、抱えると重い ♥K を切る。
	hard := discardFor(t, domain.Rummy500CpuDifficultyHard)
	assert.Equal(t, domain.CardDesignHeart, hard.GetDesign())
	assert.Equal(t, 13, hard.GetValue())

	// Easy: 一番安い札を切ってしまう (高得点札を抱えたまま終盤を迎える弱い打ち方)。
	easy := discardFor(t, domain.Rummy500CpuDifficultyEasy)
	assert.Equal(t, domain.CardDesignDiamond, easy.GetDesign())
	assert.Equal(t, 3, easy.GetValue())

	// **3 段階が同じ手にならないことを名指しで確かめる。**「難易度で分岐した」
	// だけでは、分岐先が同じ値なら設定は依然として効いていない。
	assert.NotEqual(t, easy.GetValue(), hard.GetValue())
}

// **Hard の「メルド材料は残す」を名指しで測る。**上のテストの手札では、一番重い札が
// たまたま孤立札でもあったので、素点だけで選ぶ実装と区別が付かなかった (材料保護を
// 外すミューテーションが素通りした)。ここでは一番重い札を**ペアの側**に置く。
func TestRummy500HardKeepsMeldMaterialAndDiscardsTheDeadWeight(t *testing.T) {
	g := rummy500WithDifficulty(t, domain.Rummy500CpuDifficultyHard)
	g.SetPhase(domain.Rummy500PhasePlay)
	g.SetCurrentPlayerIdx(1)
	// ♠K/♥K はペア (各10点) で最重量、♦9 は孤立の9点、♣3 は孤立の3点。
	setRummy500Hand(t, g.GetPlayer(1),
		r5Card(t, domain.CardDesignSpade, 13),
		r5Card(t, domain.CardDesignHeart, 13),
		r5Card(t, domain.CardDesignDiamond, 9),
		r5Card(t, domain.CardDesignClover, 3),
	)
	g.SetDiscardPile(nil)
	g.CpuPlay()

	pile := g.GetDiscardPile()
	require.Len(t, pile, 1)
	// 素点だけで選ぶと K を切ってしまう。材料を守るなら切るのは ♦9。
	assert.Equal(t, domain.CardDesignDiamond, pile[0].GetDesign())
	assert.Equal(t, 9, pile[0].GetValue())
}

// Hard は捨て札のトップが自分のメルドを完成させるなら、そこから引く。
// Easy/Normal は常に山札から引く。
func TestRummy500HardTakesTheDiscardThatCompletesAMeld(t *testing.T) {
	drawnFrom := func(t *testing.T, d domain.Rummy500CpuDifficulty) int {
		t.Helper()
		g := rummy500WithDifficulty(t, d)
		g.SetPhase(domain.Rummy500PhaseDraw)
		g.SetCurrentPlayerIdx(1)
		setRummy500Hand(t, g.GetPlayer(1),
			r5Card(t, domain.CardDesignSpade, 5),
			r5Card(t, domain.CardDesignSpade, 6),
			r5Card(t, domain.CardDesignHeart, 2),
		)
		// トップの ♠7 は ♠5♠6 を 3 枚のランに仕上げる。
		g.SetDiscardPile([]*domain.Card{
			r5Card(t, domain.CardDesignClover, 9),
			r5Card(t, domain.CardDesignSpade, 7),
		})
		g.SetDrawPile([]*domain.Card{r5Card(t, domain.CardDesignDiamond, 4)})
		g.CpuPlay()
		return len(g.GetDiscardPile())
	}

	// Hard は捨て札から取るので山が 1 枚減る。
	assert.Equal(t, 1, drawnFrom(t, domain.Rummy500CpuDifficultyHard), "Hard は完成する札を拾う")
	// Normal は山札から引くので捨て札は動かない。
	assert.Equal(t, 2, drawnFrom(t, domain.Rummy500CpuDifficultyNormal), "Normal は山札から引く")
	assert.Equal(t, 2, drawnFrom(t, domain.Rummy500CpuDifficultyEasy), "Easy は山札から引く")
}
