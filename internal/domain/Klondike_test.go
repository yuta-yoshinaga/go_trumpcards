//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestKlondike() *domain.Klondike {
	tc := domain.NewTrumpCards(0)
	k := domain.NewKlondike(tc)
	return k
}

func setupPlayingKlondike() *domain.Klondike {
	k := newTestKlondike()
	k.Reset()
	return k
}

func makeCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeTableauCard(design, value int, faceUp bool) *domain.KlondikeTableauCard {
	return &domain.KlondikeTableauCard{Card: makeCard(design, value), FaceUp: faceUp}
}

// clearTableau clears all tableau columns
func clearTableau(k *domain.Klondike) {
	var empty [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
	k.SetTableau(empty)
}

func TestNewKlondike(t *testing.T) {
	k := newTestKlondike()
	assert.NotNil(t, k)
	assert.Equal(t, domain.KlondikePhase(0), k.GetPhase())
}

func TestKlondike_Reset(t *testing.T) {
	k := setupPlayingKlondike()

	assert.Equal(t, domain.KlondikePhasePlaying, k.GetPhase())
	assert.Equal(t, 0, k.GetMoveCount())

	// Tableau: 7 columns, col i has i+1 cards
	tableau := k.GetTableau()
	totalTableauCards := 0
	for i := 0; i < domain.KlondikeTableauCnt; i++ {
		assert.Equal(t, i+1, len(tableau[i]), "column %d should have %d cards", i, i+1)
		// Only last card is face up
		for j, tc := range tableau[i] {
			if j == len(tableau[i])-1 {
				assert.True(t, tc.FaceUp, "top card of column %d should be face up", i)
			} else {
				assert.False(t, tc.FaceUp, "non-top card of column %d should be face down", i)
			}
		}
		totalTableauCards += len(tableau[i])
	}
	assert.Equal(t, 28, totalTableauCards)

	// Stock: 24 cards
	assert.Equal(t, 24, k.GetStockCount())

	// Waste: empty
	assert.Nil(t, k.GetWaste())

	// Foundation: empty
	foundation := k.GetFoundation()
	for i := 0; i < domain.KlondikeFoundationCnt; i++ {
		assert.Nil(t, foundation[i])
	}
}

func TestKlondike_Draw(t *testing.T) {
	t.Run("draw from stock", func(t *testing.T) {
		k := setupPlayingKlondike()
		initialStock := k.GetStockCount()
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, initialStock-1, k.GetStockCount())
		assert.Equal(t, 1, len(k.GetWaste()))
		assert.Equal(t, 1, k.GetMoveCount())
	})

	t.Run("recycle waste to stock", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		// Draw all stock cards
		for k.GetStockCount() > 0 {
			_ = k.Draw()
		}
		wasteLen := len(k.GetWaste())
		assert.Greater(t, wasteLen, 0)

		// Now drawing recycles waste to stock
		err := k.Draw()
		assert.NoError(t, err)
		assert.Equal(t, wasteLen, k.GetStockCount())
		assert.Nil(t, k.GetWaste())
	})

	t.Run("error when both stock and waste empty", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetStock(nil)
		k.SetWaste(nil)
		err := k.Draw()
		assert.Error(t, err)
		assert.Equal(t, "no cards in stock or waste", err.Error())
	})

	t.Run("error when not playing", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameClear)
		err := k.Draw()
		assert.Error(t, err)
		assert.Equal(t, "game is not in playing phase", err.Error())
	})
}

func TestKlondike_MoveWasteToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		// Set up: waste has a red 6, tableau col 0 has a black 7
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignHeart, 6)})
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetWaste()))
		assert.Equal(t, 2, len(k.GetTableau()[0]))
	})

	t.Run("K on empty column", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 13)})
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(k.GetTableau()[0]))
	})

	t.Run("error: not K on empty column", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 5)})
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
		assert.Equal(t, "cannot place card on tableau", err.Error())
	})

	t.Run("error: same color", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 6)})
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignClover, 7, true)}
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
	})

	t.Run("error: wrong value", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignHeart, 5)})
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
	})

	t.Run("error: invalid column", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 1)})
		assert.Error(t, k.MoveWasteToTableau(-1))
		assert.Error(t, k.MoveWasteToTableau(7))
	})

	t.Run("error: waste empty", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetWaste(nil)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
		assert.Equal(t, "waste is empty", err.Error())
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameOver)
		err := k.MoveWasteToTableau(0)
		assert.Error(t, err)
	})
}

func TestKlondike_MoveWasteToFoundation(t *testing.T) {
	t.Run("success ace", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 1)})
		err := k.MoveWasteToFoundation()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetWaste()))
		assert.Equal(t, 1, len(k.GetFoundation()[0])) // Spade = design 1, index 0
	})

	t.Run("success sequential", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var f [domain.KlondikeFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{makeCard(domain.CardDesignSpade, 1)}
		k.SetFoundation(f)
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 2)})
		err := k.MoveWasteToFoundation()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(k.GetFoundation()[0]))
	})

	t.Run("error: wrong sequence", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var f [domain.KlondikeFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{makeCard(domain.CardDesignSpade, 1)}
		k.SetFoundation(f)
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 3)})
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("error: not ace on empty foundation", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 2)})
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("error: waste empty", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetWaste(nil)
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameOver)
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("error: joker card", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignJoker, 0)})
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
		assert.Equal(t, "invalid card for foundation", err.Error())
	})

	t.Run("error: design out of range", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignDiamond+1, 1)})
		err := k.MoveWasteToFoundation()
		assert.Error(t, err)
		assert.Equal(t, "invalid card for foundation", err.Error())
	})
}

func TestKlondike_MoveTableauToTableau(t *testing.T) {
	t.Run("success single card", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetTableau()[0]))
		assert.Equal(t, 2, len(k.GetTableau()[1]))
	})

	t.Run("success multiple cards", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 7, true),
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 8, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetTableau()[0]))
		assert.Equal(t, 3, len(k.GetTableau()[1]))
	})

	t.Run("success K to empty column", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 13, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(k.GetTableau()[1]))
	})

	t.Run("auto-flip after move", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 10, false),
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 1, 1)
		assert.NoError(t, err)
		assert.True(t, k.GetTableau()[0][0].FaceUp, "exposed card should auto-flip")
	})

	t.Run("error: face down card", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, false)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Equal(t, "card is face down", err.Error())
	})

	t.Run("error: same column", func(t *testing.T) {
		k := setupPlayingKlondike()
		err := k.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("error: invalid from column", func(t *testing.T) {
		k := setupPlayingKlondike()
		assert.Error(t, k.MoveTableauToTableau(-1, 0, 1))
		assert.Error(t, k.MoveTableauToTableau(7, 0, 1))
	})

	t.Run("error: invalid to column", func(t *testing.T) {
		k := setupPlayingKlondike()
		assert.Error(t, k.MoveTableauToTableau(0, 0, -1))
		assert.Error(t, k.MoveTableauToTableau(0, 0, 7))
	})

	t.Run("error: invalid card index", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
		k.SetTableau(tab)
		assert.Error(t, k.MoveTableauToTableau(0, -1, 1))
		assert.Error(t, k.MoveTableauToTableau(0, 1, 1))
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameOver)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("error: cannot place", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestKlondike_MoveTableauToFoundation(t *testing.T) {
	t.Run("success ace", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 1, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(k.GetTableau()[0]))
		assert.Equal(t, 1, len(k.GetFoundation()[0]))
	})

	t.Run("auto-flip after move", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, false),
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.True(t, k.GetTableau()[0][0].FaceUp)
	})

	t.Run("error: empty column", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		clearTableau(k)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
		assert.Equal(t, "tableau column is empty", err.Error())
	})

	t.Run("error: invalid column", func(t *testing.T) {
		k := setupPlayingKlondike()
		assert.Error(t, k.MoveTableauToFoundation(-1))
		assert.Error(t, k.MoveTableauToFoundation(7))
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameClear)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("error: cannot place on foundation", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("error: joker card", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignJoker, 0, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
		assert.Equal(t, "invalid card for foundation", err.Error())
	})

	t.Run("error: design out of range", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignDiamond+1, 1, true)}
		k.SetTableau(tab)
		err := k.MoveTableauToFoundation(0)
		assert.Error(t, err)
		assert.Equal(t, "invalid card for foundation", err.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		// Set up foundation with 12 cards each
		var f [domain.KlondikeFoundationCnt][]*domain.Card
		for i := 0; i < domain.KlondikeFoundationCnt; i++ {
			f[i] = make([]*domain.Card, 0)
			for v := 1; v <= 12; v++ {
				f[i] = append(f[i], makeCard(i+1, v))
			}
		}
		k.SetFoundation(f)
		// Set last cards on tableau
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		for i := 0; i < domain.KlondikeFoundationCnt; i++ {
			tab[i] = []*domain.KlondikeTableauCard{makeTableauCard(i+1, 13, true)}
		}
		k.SetTableau(tab)
		k.SetStock(nil)
		k.SetWaste(nil)

		for i := 0; i < domain.KlondikeFoundationCnt; i++ {
			err := k.MoveTableauToFoundation(i)
			assert.NoError(t, err)
		}
		assert.Equal(t, domain.KlondikePhaseGameClear, k.GetPhase())
	})
}

func TestKlondike_GiveUp(t *testing.T) {
	t.Run("give up during playing", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.GiveUp()
		assert.Equal(t, domain.KlondikePhaseGameOver, k.GetPhase())
	})

	t.Run("give up when already game over", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameOver)
		k.GiveUp()
		assert.Equal(t, domain.KlondikePhaseGameOver, k.GetPhase())
	})

	t.Run("give up when game clear", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameClear)
		k.GiveUp()
		assert.Equal(t, domain.KlondikePhaseGameClear, k.GetPhase())
	})
}

func TestKlondike_GetHint(t *testing.T) {
	t.Run("no hint when not playing", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameOver)
		assert.Nil(t, k.GetHint())
	})

	t.Run("hint: tableau to foundation", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 1, true)}
		k.SetTableau(tab)
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint: waste to foundation", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		clearTableau(k)
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignHeart, 1)})
		k.SetStock(nil)
		hint := k.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint: tableau to tableau (expose face down)", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 10, false),
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, "tableau", hint.ToZone)
		assert.Equal(t, 1, hint.ToCol)
	})

	t.Run("hint: waste to tableau", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		k.SetTableau(tab)
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignHeart, 6)})
		k.SetStock(nil)
		hint := k.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
		assert.Equal(t, 0, hint.ToCol)
	})

	t.Run("no hint available", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		// Set up a state with no valid moves
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignClover, 5, true)}
		k.SetTableau(tab)
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 3)})
		k.SetStock(nil)
		hint := k.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("skip empty tableau columns in hint", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		clearTableau(k)
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("skip all face-up tableau in hint", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		// All face up in col 0 - no face-down to expose, should skip
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 7, true),
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 8, true)}
		k.SetTableau(tab)
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		// No hint for foundation, and tab[0] has no face-down cards to expose
		assert.Nil(t, hint)
	})

	t.Run("tableau to tableau: no valid target", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 10, false),
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		// No column has a card that 6♥ can go under
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
		k.SetTableau(tab)
		k.SetWaste(nil)
		k.SetStock(nil)
		hint := k.GetHint()
		assert.Nil(t, hint)
	})
}

func TestKlondike_AutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		// Set up: all cards face up in tableau, each column has cards from K down to A
		// so the top card (last in slice) is A and can be moved to foundation first
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		designs := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for suit := 0; suit < 4; suit++ {
			tab[suit] = make([]*domain.KlondikeTableauCard, 0)
			for v := 13; v >= 1; v-- {
				tab[suit] = append(tab[suit], makeTableauCard(designs[suit], v, true))
			}
		}
		k.SetTableau(tab)
		k.SetStock(nil)
		k.SetWaste(nil)
		err := k.AutoComplete()
		assert.NoError(t, err)
		assert.Equal(t, domain.KlondikePhaseGameClear, k.GetPhase())
		for i := 0; i < domain.KlondikeFoundationCnt; i++ {
			assert.Equal(t, 13, len(k.GetFoundation()[i]))
		}
	})

	t.Run("success with waste", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		// Set up: all cards face up, some in waste
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 1)})
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 2, true)}
		k.SetTableau(tab)
		k.SetStock(nil)

		// Build rest of foundations
		var f [domain.KlondikeFoundationCnt][]*domain.Card
		for i := 1; i < domain.KlondikeFoundationCnt; i++ {
			f[i] = make([]*domain.Card, 0)
			for v := 1; v <= 13; v++ {
				f[i] = append(f[i], makeCard(i+1, v))
			}
		}
		// Spade foundation has 0 cards, waste+tab has A,2
		f[0] = make([]*domain.Card, 0)
		for v := 3; v <= 13; v++ {
			f[0] = append(f[0], makeCard(domain.CardDesignSpade, v))
		}
		k.SetFoundation(f)
		err := k.AutoComplete()
		assert.NoError(t, err)
	})

	t.Run("error: not playing", func(t *testing.T) {
		k := setupPlayingKlondike()
		k.SetPhase(domain.KlondikePhaseGameOver)
		err := k.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("error: not all face up", func(t *testing.T) {
		k := setupPlayingKlondike()
		err := k.AutoComplete()
		assert.Error(t, err)
		assert.Equal(t, "not all cards are face up", err.Error())
	})

	t.Run("waste card cannot go to foundation stops loop", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 5)})
		clearTableau(k)
		k.SetStock(nil)
		err := k.AutoComplete()
		assert.NoError(t, err)
		// Card 5 can't go on empty foundation, so it stays in waste
		assert.Equal(t, 1, len(k.GetWaste()))
	})
}

func TestKlondike_AllFaceUp(t *testing.T) {
	t.Run("all face up", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 1, true)}
		k.SetTableau(tab)
		k.SetStock(nil)
		assert.True(t, k.AllFaceUp())
	})

	t.Run("stock not empty", func(t *testing.T) {
		k := setupPlayingKlondike()
		assert.False(t, k.AllFaceUp())
	})

	t.Run("face down card in tableau", func(t *testing.T) {
		k := newTestKlondike()
		k.Reset()
		var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 1, false)}
		k.SetTableau(tab)
		k.SetStock(nil)
		assert.False(t, k.AllFaceUp())
	})
}

func TestKlondike_ActionLog(t *testing.T) {
	k := setupPlayingKlondike()
	assert.Nil(t, k.GetActionLog())
	_ = k.Draw()
	assert.NotNil(t, k.GetActionLog())
	assert.Equal(t, 1, len(k.GetActionLog()))
}

func TestKlondike_MoveWasteToFoundation_DifferentSuit(t *testing.T) {
	k := newTestKlondike()
	k.Reset()
	var f [domain.KlondikeFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{makeCard(domain.CardDesignSpade, 1)}
	k.SetFoundation(f)
	// Try to put heart 2 on spade foundation - different design
	k.SetWaste([]*domain.Card{makeCard(domain.CardDesignHeart, 2)})
	err := k.MoveWasteToFoundation()
	// This should fail because heart goes to foundation[2] (CardDesignHeart-1=2), not [0]
	// But foundation[2] is empty and card is 2, not ace
	assert.Error(t, err)
}

func TestKlondike_IsBlackColor(t *testing.T) {
	k := newTestKlondike()
	k.Reset()
	// Spade (black) on Spade (black) should fail
	var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
	k.SetTableau(tab)
	k.SetWaste([]*domain.Card{makeCard(domain.CardDesignClover, 6)})
	err := k.MoveWasteToTableau(0)
	assert.Error(t, err) // Same color (both black)

	// Diamond (red) on Spade (black) should succeed
	k.SetWaste([]*domain.Card{makeCard(domain.CardDesignDiamond, 6)})
	err = k.MoveWasteToTableau(0)
	assert.NoError(t, err)
}

func TestKlondike_ResetClearsState(t *testing.T) {
	k := setupPlayingKlondike()
	_ = k.Draw()
	k.GiveUp()
	assert.Equal(t, domain.KlondikePhaseGameOver, k.GetPhase())

	k.Reset()
	assert.Equal(t, domain.KlondikePhasePlaying, k.GetPhase())
	assert.Equal(t, 0, k.GetMoveCount())
	assert.Nil(t, k.GetActionLog())
}

func TestKlondike_AutoFlipEmptyColumn(t *testing.T) {
	// Auto-flip should not panic on empty column
	k := newTestKlondike()
	k.Reset()
	var tab [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 1, true)}
	k.SetTableau(tab)
	err := k.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(k.GetTableau()[0]))
}

func TestKlondike_CanPlaceOnFoundation_WrongSequence(t *testing.T) {
	k := newTestKlondike()
	k.Reset()
	var f [domain.KlondikeFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{makeCard(domain.CardDesignSpade, 1), makeCard(domain.CardDesignSpade, 2)}
	k.SetFoundation(f)
	// Try to place Spade 4 (skipping 3)
	k.SetWaste([]*domain.Card{makeCard(domain.CardDesignSpade, 4)})
	err := k.MoveWasteToFoundation()
	assert.Error(t, err)
}
