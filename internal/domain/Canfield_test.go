//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestCanfield() *domain.Canfield {
	return domain.NewCanfield(domain.NewTrumpCards(0))
}

func setupPlayingCanfield() *domain.Canfield {
	c := newTestCanfield()
	c.Reset()
	return c
}

func makeCanfieldTC(design, value int) *domain.CanfieldTableauCard {
	return &domain.CanfieldTableauCard{Card: domain.NewCard(design, value, false)}
}

func clearCanfieldTableau(c *domain.Canfield) {
	var empty [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
	c.SetTableau(empty)
}

func TestNewCanfield(t *testing.T) {
	c := newTestCanfield()
	assert.NotNil(t, c)
	assert.Equal(t, domain.CanfieldPhase(0), c.GetPhase())
}

func TestCanfield_Reset(t *testing.T) {
	c := setupPlayingCanfield()

	assert.Equal(t, domain.CanfieldPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())

	// Tableau: 4 columns, 1 card each
	tab := c.GetTableau()
	for i := 0; i < domain.CanfieldTableauCnt; i++ {
		assert.Equal(t, 1, len(tab[i]))
	}

	// Reserve: 13 cards
	assert.Equal(t, 13, len(c.GetReserve()))

	// One foundation has the base card
	foundation := c.GetFoundation()
	baseCount := 0
	for i := 0; i < domain.CanfieldFoundationCnt; i++ {
		baseCount += len(foundation[i])
	}
	assert.Equal(t, 1, baseCount)

	// BaseRank set 1..13
	br := c.GetBaseRank()
	assert.GreaterOrEqual(t, br, 1)
	assert.LessOrEqual(t, br, 13)

	// Stock: 52 - 13 - 4 - 1 = 34
	assert.Equal(t, 34, c.GetStockCount())
	assert.Nil(t, c.GetWaste())
}

func TestCanfield_Draw(t *testing.T) {
	t.Run("draw 3 from stock", func(t *testing.T) {
		c := setupPlayingCanfield()
		initial := c.GetStockCount()
		err := c.Draw()
		assert.NoError(t, err)
		assert.Equal(t, initial-3, c.GetStockCount())
		assert.Equal(t, 3, len(c.GetWaste()))
		assert.Equal(t, 1, c.GetMoveCount())
	})

	t.Run("recycle waste to stock", func(t *testing.T) {
		c := setupPlayingCanfield()
		for c.GetStockCount() > 0 {
			_ = c.Draw()
		}
		wasteLen := len(c.GetWaste())
		assert.Greater(t, wasteLen, 0)
		err := c.Draw()
		assert.NoError(t, err)
		assert.Equal(t, wasteLen, c.GetStockCount())
		assert.Nil(t, c.GetWaste())
	})

	t.Run("error: stock and waste empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetStock(nil)
		c.SetWaste(nil)
		err := c.Draw()
		assert.Error(t, err)
	})

	t.Run("error: not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameClear)
		err := c.Draw()
		assert.Error(t, err)
	})

	t.Run("draw fewer than 3 when stock small", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetStock([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		c.SetWaste(nil)
		err := c.Draw()
		assert.NoError(t, err)
		assert.Equal(t, 0, c.GetStockCount())
		assert.Equal(t, 1, len(c.GetWaste()))
	})
}

func TestCanfield_MoveWasteToTableau(t *testing.T) {
	t.Run("success alternate color descending", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)})
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 7)}
		c.SetTableau(tab)
		assert.NoError(t, c.MoveWasteToTableau(0))
		assert.Equal(t, 0, len(c.GetWaste()))
		assert.Equal(t, 2, len(c.GetTableau()[0]))
	})

	t.Run("success A on 2 wrap", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)})
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignHeart, 1)}
		c.SetTableau(tab)
		assert.NoError(t, c.MoveWasteToTableau(0))
	})

	t.Run("error: same color", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 6, false)})
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignClover, 7)}
		c.SetTableau(tab)
		assert.Error(t, c.MoveWasteToTableau(0))
	})

	t.Run("error: wrong rank", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 7)}
		c.SetTableau(tab)
		assert.Error(t, c.MoveWasteToTableau(0))
	})

	t.Run("error: waste empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetWaste(nil)
		assert.Error(t, c.MoveWasteToTableau(0))
	})

	t.Run("error: invalid col", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		assert.Error(t, c.MoveWasteToTableau(-1))
		assert.Error(t, c.MoveWasteToTableau(4))
	})

	t.Run("error: empty tableau while reserve non-empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})
		clearCanfieldTableau(c)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 8, false)})
		assert.Error(t, c.MoveWasteToTableau(0))
	})

	t.Run("success: any card on empty when reserve empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		clearCanfieldTableau(c)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		assert.NoError(t, c.MoveWasteToTableau(0))
	})

	t.Run("error: not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameClear)
		assert.Error(t, c.MoveWasteToTableau(0))
	})
}

func TestCanfield_MoveWasteToFoundation(t *testing.T) {
	t.Run("success: base rank", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(7)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		assert.NoError(t, c.MoveWasteToFoundation())
		assert.Equal(t, 1, len(c.GetFoundation()[0]))
	})

	t.Run("success: next rank same suit", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(7)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		c.SetFoundation(f)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 8, false)})
		assert.NoError(t, c.MoveWasteToFoundation())
	})

	t.Run("success: wrap K to A", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(2)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)}
		c.SetFoundation(f)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		assert.NoError(t, c.MoveWasteToFoundation())
	})

	t.Run("error: not base rank on empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(7)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 8, false)})
		assert.Error(t, c.MoveWasteToFoundation())
	})

	t.Run("error: wrong suit", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(7)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
		c.SetFoundation(f)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 8, false)})
		assert.Error(t, c.MoveWasteToFoundation())
	})

	t.Run("error: waste empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetWaste(nil)
		assert.Error(t, c.MoveWasteToFoundation())
	})

	t.Run("error: joker", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignJoker, 0, false)})
		assert.Error(t, c.MoveWasteToFoundation())
	})

	t.Run("error: not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameOver)
		assert.Error(t, c.MoveWasteToFoundation())
	})
}

func TestCanfield_MoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignHeart, 6)}
		tab[1] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 7)}
		c.SetTableau(tab)
		assert.NoError(t, c.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 0, len(c.GetTableau()[0]))
		assert.Equal(t, 2, len(c.GetTableau()[1]))
	})

	t.Run("success multi-card", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{
			makeCanfieldTC(domain.CardDesignSpade, 7),
			makeCanfieldTC(domain.CardDesignHeart, 6),
		}
		tab[1] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignHeart, 8)}
		c.SetTableau(tab)
		assert.NoError(t, c.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 3, len(c.GetTableau()[1]))
	})

	t.Run("error: same column", func(t *testing.T) {
		c := setupPlayingCanfield()
		assert.Error(t, c.MoveTableauToTableau(0, 0, 0))
	})

	t.Run("error: invalid columns", func(t *testing.T) {
		c := setupPlayingCanfield()
		assert.Error(t, c.MoveTableauToTableau(-1, 0, 1))
		assert.Error(t, c.MoveTableauToTableau(4, 0, 1))
		assert.Error(t, c.MoveTableauToTableau(0, 0, -1))
		assert.Error(t, c.MoveTableauToTableau(0, 0, 4))
	})

	t.Run("error: invalid card index", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignHeart, 6)}
		c.SetTableau(tab)
		assert.Error(t, c.MoveTableauToTableau(0, -1, 1))
		assert.Error(t, c.MoveTableauToTableau(0, 1, 1))
	})

	t.Run("error: cannot place", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignHeart, 6)}
		tab[1] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 5)}
		c.SetTableau(tab)
		assert.Error(t, c.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("error: not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameOver)
		assert.Error(t, c.MoveTableauToTableau(0, 0, 1))
	})
}

func TestCanfield_MoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(5)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 5)}
		c.SetTableau(tab)
		c.SetReserve(nil)
		assert.NoError(t, c.MoveTableauToFoundation(0))
		assert.Equal(t, 0, len(c.GetTableau()[0]))
	})

	t.Run("error: empty col", func(t *testing.T) {
		c := setupPlayingCanfield()
		clearCanfieldTableau(c)
		c.SetReserve(nil)
		assert.Error(t, c.MoveTableauToFoundation(0))
	})

	t.Run("error: invalid col", func(t *testing.T) {
		c := setupPlayingCanfield()
		assert.Error(t, c.MoveTableauToFoundation(-1))
		assert.Error(t, c.MoveTableauToFoundation(4))
	})

	t.Run("error: cannot place", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(5)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 3)}
		c.SetTableau(tab)
		assert.Error(t, c.MoveTableauToFoundation(0))
	})

	t.Run("error: joker", func(t *testing.T) {
		c := setupPlayingCanfield()
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignJoker, 0)}
		c.SetTableau(tab)
		assert.Error(t, c.MoveTableauToFoundation(0))
	})

	t.Run("error: not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameClear)
		assert.Error(t, c.MoveTableauToFoundation(0))
	})
}

func TestCanfield_MoveReserveToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)})
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 7)}
		c.SetTableau(tab)
		assert.NoError(t, c.MoveReserveToTableau(0))
		assert.Equal(t, 0, len(c.GetReserve()))
	})

	t.Run("error: reserve empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		assert.Error(t, c.MoveReserveToTableau(0))
	})

	t.Run("error: invalid col", func(t *testing.T) {
		c := setupPlayingCanfield()
		assert.Error(t, c.MoveReserveToTableau(-1))
		assert.Error(t, c.MoveReserveToTableau(4))
	})

	t.Run("error: cannot place", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 6, false)})
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignClover, 7)}
		c.SetTableau(tab)
		assert.Error(t, c.MoveReserveToTableau(0))
	})

	t.Run("error: not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameOver)
		assert.Error(t, c.MoveReserveToTableau(0))
	})
}

func TestCanfield_MoveReserveToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(5)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		assert.NoError(t, c.MoveReserveToFoundation())
	})

	t.Run("error: reserve empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve(nil)
		assert.Error(t, c.MoveReserveToFoundation())
	})

	t.Run("error: cannot place", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(5)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		assert.Error(t, c.MoveReserveToFoundation())
	})

	t.Run("error: joker", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignJoker, 0, false)})
		assert.Error(t, c.MoveReserveToFoundation())
	})

	t.Run("error: not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameOver)
		assert.Error(t, c.MoveReserveToFoundation())
	})
}

func TestCanfield_GiveUp(t *testing.T) {
	c := setupPlayingCanfield()
	c.GiveUp()
	assert.Equal(t, domain.CanfieldPhaseGameOver, c.GetPhase())
	// Already game over: noop
	c.GiveUp()
	assert.Equal(t, domain.CanfieldPhaseGameOver, c.GetPhase())
}

func TestCanfield_GetHint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameOver)
		assert.Nil(t, c.GetHint())
	})

	t.Run("reserve to foundation", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(5)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		c.SetWaste(nil)
		h := c.GetHint()
		assert.NotNil(t, h)
		assert.Equal(t, "reserve", h.FromZone)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(5)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetReserve(nil)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 5)}
		c.SetTableau(tab)
		c.SetWaste(nil)
		h := c.GetHint()
		assert.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
	})

	t.Run("waste to foundation", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(5)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetReserve(nil)
		clearCanfieldTableau(c)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		h := c.GetHint()
		assert.NotNil(t, h)
		assert.Equal(t, "waste", h.FromZone)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("reserve to tableau", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(1)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)})
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 7)}
		c.SetTableau(tab)
		c.SetWaste(nil)
		h := c.GetHint()
		assert.NotNil(t, h)
		assert.Equal(t, "reserve", h.FromZone)
		assert.Equal(t, "tableau", h.ToZone)
	})

	t.Run("waste to tableau", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(1)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetReserve(nil)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 7)}
		c.SetTableau(tab)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)})
		h := c.GetHint()
		assert.NotNil(t, h)
		assert.Equal(t, "waste", h.FromZone)
		assert.Equal(t, "tableau", h.ToZone)
	})

	t.Run("no hint", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(13)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		c.SetReserve(nil)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		tab[0] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 3)}
		tab[1] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignClover, 3)}
		tab[2] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 5)}
		tab[3] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignClover, 5)}
		c.SetTableau(tab)
		c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		assert.Nil(t, c.GetHint())
	})
}

func TestCanfield_AutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetBaseRank(1)
		c.SetReserve(nil)
		c.SetStock(nil)
		c.SetWaste(nil)
		var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
		designs := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i, d := range designs {
			col := make([]*domain.CanfieldTableauCard, 0)
			for v := 13; v >= 1; v-- {
				col = append(col, makeCanfieldTC(d, v))
			}
			tab[i] = col
		}
		c.SetTableau(tab)
		var f [domain.CanfieldFoundationCnt][]*domain.Card
		c.SetFoundation(f)
		assert.NoError(t, c.AutoComplete())
		assert.Equal(t, domain.CanfieldPhaseGameClear, c.GetPhase())
	})

	t.Run("error: reserve not empty", func(t *testing.T) {
		c := setupPlayingCanfield()
		assert.Error(t, c.AutoComplete())
	})

	t.Run("error: not playing", func(t *testing.T) {
		c := setupPlayingCanfield()
		c.SetPhase(domain.CanfieldPhaseGameOver)
		assert.Error(t, c.AutoComplete())
	})
}

func TestCanfield_Undo(t *testing.T) {
	c := setupPlayingCanfield()
	initialStock := c.GetStockCount()
	assert.False(t, c.CanUndo())
	assert.NoError(t, c.Draw())
	assert.True(t, c.CanUndo())
	assert.NoError(t, c.Undo())
	assert.Equal(t, initialStock, c.GetStockCount())
	assert.False(t, c.CanUndo())
}

func TestCanfield_UndoErrors(t *testing.T) {
	c := setupPlayingCanfield()
	assert.Error(t, c.Undo())
	c.SetPhase(domain.CanfieldPhaseGameOver)
	assert.Error(t, c.Undo())
}

func TestCanfield_UndoN(t *testing.T) {
	c := setupPlayingCanfield()
	_ = c.Draw()
	_ = c.Draw()
	assert.NoError(t, c.UndoN(2))
	assert.False(t, c.CanUndo())
	assert.Error(t, c.UndoN(1))
}

func TestCanfield_AutoFillFromReserve(t *testing.T) {
	c := setupPlayingCanfield()
	// Make col 0 empty, reserve has 1 card
	var tab [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
	tab[0] = nil
	tab[1] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 7)}
	tab[2] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 8)}
	tab[3] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignSpade, 9)}
	c.SetTableau(tab)
	c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)})
	c.SetWaste([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)})
	// Move waste to col 1 triggers autoFill? No — col 0 is already empty with reserve; auto fills on move
	assert.NoError(t, c.MoveWasteToTableau(1))
	assert.Equal(t, 0, len(c.GetReserve()))
	assert.Equal(t, 1, len(c.GetTableau()[0]))
}

func TestCanfield_JSON(t *testing.T) {
	c := setupPlayingCanfield()
	data, err := c.MarshalJSON()
	assert.NoError(t, err)
	c2 := newTestCanfield()
	assert.NoError(t, c2.UnmarshalJSON(data))
	assert.Equal(t, c.GetBaseRank(), c2.GetBaseRank())
	assert.Equal(t, c.GetStockCount(), c2.GetStockCount())
}

func TestCanfield_MoveToTableau_Errors(t *testing.T) {
	c := newTestCanfield()
	c.SetPhase(domain.CanfieldPhasePlaying)
	c.SetBaseRank(13)

	var emptyTableau [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard
	// Column 0 is empty. Column 1 has Heart 5.
	emptyTableau[1] = []*domain.CanfieldTableauCard{makeCanfieldTC(domain.CardDesignHeart, 5)}
	c.SetTableau(emptyTableau)

	// Set reserve with 1 card
	c.SetReserve([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 10, false)})
	// Set waste with 1 card (Club 5)
	wasteCard1 := domain.NewCard(domain.CardDesignClover, 5, false)
	c.SetWaste([]*domain.Card{wasteCard1})

	t.Run("empty column with reserve", func(t *testing.T) {
		err := c.MoveWasteToTableau(0) // Empty column
		assert.Error(t, err)
		code, _ := domain.ErrorMessageCode(err)
		assert.Equal(t, "canfield.errEmptyColumnAutoFillOnly", code)
	})

	t.Run("empty column without reserve (negative control)", func(t *testing.T) {
		c.SetReserve([]*domain.Card{}) // Empty reserve
		err := c.MoveWasteToTableau(0) // Empty column
		assert.NoError(t, err, "Should be able to move to empty column if reserve is empty")
	})

	t.Run("not alternate descending", func(t *testing.T) {
		// ランクは合っていて色だけ違う札を選ぶ。ランクも色も外すと、
		// どちらの規則で弾かれたのか分からないまま緑になる。
		wasteCard2 := domain.NewCard(domain.CardDesignHeart, 4, false)
		c.SetWaste([]*domain.Card{wasteCard2})
		err := c.MoveWasteToTableau(1) // Column 1 has Heart 5
		assert.Error(t, err)
		code, _ := domain.ErrorMessageCode(err)
		assert.Equal(t, "canfield.errNotAlternateDescending", code)
	})
}
