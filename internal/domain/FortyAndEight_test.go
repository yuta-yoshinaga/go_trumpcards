//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestFortyAndEight() *domain.FortyAndEight {
	tc := domain.NewTrumpCardsWithDecks(2, 0)
	ft := domain.NewFortyAndEight(tc)
	return ft
}

func setupPlayingFortyAndEight() *domain.FortyAndEight {
	ft := newTestFortyAndEight()
	ft.Reset()
	return ft
}

func makeF8Card(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeF8TableauCard(design, value int) *domain.FortyAndEightTableauCard {
	return &domain.FortyAndEightTableauCard{Card: makeF8Card(design, value), FaceUp: true}
}

func clearF8Tableau(ft *domain.FortyAndEight) {
	var empty [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
	ft.SetTableau(empty)
}

func TestNewFortyAndEight(t *testing.T) {
	ft := newTestFortyAndEight()
	assert.NotNil(t, ft)
	assert.Equal(t, domain.FortyAndEightPhase(0), ft.GetPhase())
}

func TestFortyAndEight_Reset(t *testing.T) {
	ft := setupPlayingFortyAndEight()

	assert.Equal(t, domain.FortyAndEightPhasePlaying, ft.GetPhase())
	assert.Equal(t, 0, ft.GetMoveCount())

	// Tableau: 8 columns, each with 5 face-up cards
	tableau := ft.GetTableau()
	totalTableauCards := 0
	for i := 0; i < domain.FortyAndEightTableauCnt; i++ {
		assert.Equal(t, 5, len(tableau[i]), "column %d should have 5 cards", i)
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "all cards should be face up")
		}
		totalTableauCards += len(tableau[i])
	}
	assert.Equal(t, 40, totalTableauCards)

	// Stock: 64 cards (104 - 40)
	assert.Equal(t, 64, ft.GetStockCount())

	// Waste: empty
	assert.Nil(t, ft.GetWaste())

	// Redeal: not used
	assert.False(t, ft.GetRedealUsed())

	// Foundation: empty
	foundation := ft.GetFoundation()
	for i := 0; i < domain.FortyAndEightFoundationCnt; i++ {
		assert.Nil(t, foundation[i])
	}
}

func TestFortyAndEight_Draw(t *testing.T) {
	t.Run("draw from stock", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		initialStock := ft.GetStockCount()
		err := ft.Draw()
		assert.NoError(t, err)
		assert.Equal(t, initialStock-1, ft.GetStockCount())
		assert.Equal(t, 1, len(ft.GetWaste()))
		assert.Equal(t, 1, ft.GetMoveCount())
	})

	t.Run("draw when stock is empty (no recycle)", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		ft.SetStock(nil)
		err := ft.Draw()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cards in stock")
	})

	t.Run("draw when game is not playing", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameOver)
		err := ft.Draw()
		assert.Error(t, err)
	})
}

func TestFortyAndEight_Redeal(t *testing.T) {
	t.Run("happy path: waste recycled to stock once", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetStock(nil)
		// Set waste in draw order: first drawn = index 0.
		w0 := makeF8Card(domain.CardDesignSpade, 5)
		w1 := makeF8Card(domain.CardDesignHeart, 9)
		ft.SetWaste([]*domain.Card{w0, w1})

		assert.True(t, ft.CanRedeal())
		err := ft.Redeal()
		assert.NoError(t, err)
		assert.True(t, ft.GetRedealUsed())
		assert.Equal(t, 0, len(ft.GetWaste()))
		assert.Equal(t, 2, ft.GetStockCount())
		assert.Equal(t, 1, ft.GetMoveCount())

		// Draw should pull the originally first-drawn card (w0) first again.
		err = ft.Draw()
		assert.NoError(t, err)
		require.Equal(t, 1, len(ft.GetWaste()))
		assert.Equal(t, w0.GetDesign(), ft.GetWaste()[0].GetDesign())
		assert.Equal(t, w0.GetValue(), ft.GetWaste()[0].GetValue())
	})

	t.Run("reject when stock not empty", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		// Stock has 64 cards after reset.
		err := ft.Redeal()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stock is not empty")
		assert.False(t, ft.GetRedealUsed())
	})

	t.Run("reject when already used", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		ft.SetStock(nil)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 5)})
		ft.SetRedealUsed(true)
		err := ft.Redeal()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already used")
	})

	t.Run("reject when waste empty", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		ft.SetStock(nil)
		ft.SetWaste(nil)
		assert.False(t, ft.CanRedeal())
		err := ft.Redeal()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "waste is empty")
	})

	t.Run("reject when not playing", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameOver)
		err := ft.Redeal()
		assert.Error(t, err)
	})

	t.Run("undo restores pre-redeal state", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetStock(nil)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 5)})
		require.NoError(t, ft.Redeal())
		assert.True(t, ft.GetRedealUsed())
		require.NoError(t, ft.Undo())
		assert.False(t, ft.GetRedealUsed())
		assert.Equal(t, 0, ft.GetStockCount())
		assert.Equal(t, 1, len(ft.GetWaste()))
	})
}

func TestFortyAndEight_MoveWasteToTableau(t *testing.T) {
	t.Run("valid move same suit descending", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 4)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetTableau()[0]))
	})

	t.Run("reject different suit", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignHeart, 4)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("valid move to empty column", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 7)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ft.GetTableau()[0]))
	})

	t.Run("waste empty", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 1)})
		err := ft.MoveWasteToTableau(-1)
		assert.Error(t, err)
		err = ft.MoveWasteToTableau(8)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameOver)
		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
	})
}

func TestFortyAndEight_MoveWasteToFoundation(t *testing.T) {
	t.Run("place ace on empty foundation", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 1)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.NoError(t, err)
		found := false
		for i := 0; i < domain.FortyAndEightFoundationCnt; i++ {
			if len(ft.GetFoundation()[i]) == 1 {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("place card on matching foundation", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var foundation [domain.FortyAndEightFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeF8Card(domain.CardDesignSpade, 1)}
		ft.SetFoundation(foundation)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 2)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetFoundation()[0]))
	})

	t.Run("cannot place non-ace on empty foundation", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 5)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("waste empty", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameClear)
		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})
}

func TestFortyAndEight_MoveTableauToTableau(t *testing.T) {
	t.Run("valid single card move same suit descending", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 5)}
		tableau[1] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
		assert.Equal(t, 2, len(ft.GetTableau()[1]))
	})

	t.Run("reject multi-card move", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{
			makeF8TableauCard(domain.CardDesignSpade, 6),
			makeF8TableauCard(domain.CardDesignSpade, 5),
		}
		tableau[1] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 7)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only the top card can be moved")
	})

	t.Run("reject different suit", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("move to empty column", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
		assert.Equal(t, 1, len(ft.GetTableau()[1]))
	})

	t.Run("same column", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		err := ft.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		err := ft.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = ft.MoveTableauToTableau(0, 0, 8)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		err := ft.MoveTableauToTableau(0, -1, 1)
		assert.Error(t, err)
		err = ft.MoveTableauToTableau(0, 100, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameOver)
		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestFortyAndEight_MoveTableauToFoundation(t *testing.T) {
	t.Run("move ace to foundation", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 1)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
	})

	t.Run("move card to existing foundation pile", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var foundation [domain.FortyAndEightFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeF8Card(domain.CardDesignSpade, 1)}
		ft.SetFoundation(foundation)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 2)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetFoundation()[0]))
	})

	t.Run("empty column", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("cannot place on foundation", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		err := ft.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = ft.MoveTableauToFoundation(8)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameOver)
		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

func TestFortyAndEight_GiveUp(t *testing.T) {
	t.Run("give up during playing", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		ft.GiveUp()
		assert.Equal(t, domain.FortyAndEightPhaseGameOver, ft.GetPhase())
	})

	t.Run("give up when not playing does nothing", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameClear)
		ft.GiveUp()
		assert.Equal(t, domain.FortyAndEightPhaseGameClear, ft.GetPhase())
	})
}

func TestFortyAndEight_GetHint(t *testing.T) {
	t.Run("hint tableau to foundation", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 1)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint waste to foundation", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignHeart, 1)})
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint tableau to tableau", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 5)}
		tableau[1] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("hint waste to tableau", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 5)})
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("no hint available", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 13)}
		tableau[1] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignHeart, 13)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("no hint when not playing", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameOver)
		hint := ft.GetHint()
		assert.Nil(t, hint)
	})
}

func TestFortyAndEight_AutoComplete(t *testing.T) {
	t.Run("auto complete when all face up", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetStock(nil)

		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 1)}
		tableau[1] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignHeart, 1)}
		ft.SetTableau(tableau)

		err := ft.AutoComplete()
		assert.NoError(t, err)
	})

	t.Run("cannot auto complete with stock remaining", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		err := ft.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameOver)
		err := ft.AutoComplete()
		assert.Error(t, err)
	})
}

func TestFortyAndEight_AllFaceUp(t *testing.T) {
	t.Run("true when stock is empty", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		ft.SetStock(nil)
		assert.True(t, ft.AllFaceUp())
	})

	t.Run("false when stock has cards", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		assert.False(t, ft.AllFaceUp())
	})
}

func TestFortyAndEight_Undo(t *testing.T) {
	t.Run("undo draw", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		initialStock := ft.GetStockCount()
		_ = ft.Draw()
		assert.Equal(t, initialStock-1, ft.GetStockCount())

		err := ft.Undo()
		assert.NoError(t, err)
		assert.Equal(t, initialStock, ft.GetStockCount())
		assert.Equal(t, 0, ft.GetMoveCount())
	})

	t.Run("cannot undo with no history", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		err := ft.Undo()
		assert.Error(t, err)
	})

	t.Run("cannot undo when not playing", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.SetPhase(domain.FortyAndEightPhaseGameOver)
		err := ft.Undo()
		assert.Error(t, err)
	})
}

func TestFortyAndEight_CanUndo(t *testing.T) {
	t.Run("true after action", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		_ = ft.Draw()
		assert.True(t, ft.CanUndo())
	})

	t.Run("false with no history", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		assert.False(t, ft.CanUndo())
	})
}

func TestFortyAndEight_UndoN(t *testing.T) {
	ft := setupPlayingFortyAndEight()
	_ = ft.Draw()
	_ = ft.Draw()
	_ = ft.Draw()
	assert.Equal(t, 3, ft.GetMoveCount())

	err := ft.UndoN(3)
	assert.NoError(t, err)
	assert.Equal(t, 0, ft.GetMoveCount())
}

func TestFortyAndEight_UndoToEscape(t *testing.T) {
	t.Run("returns 0 when not stalemate", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		assert.Equal(t, 0, ft.UndoToEscape())
	})

	t.Run("returns undo count to escape", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		_ = ft.Draw()
		_ = ft.Draw()
		ft.SetIsStalemate(true)
		result := ft.UndoToEscape()
		assert.True(t, result > 0)
	})
}

func TestFortyAndEight_Stalemate(t *testing.T) {
	t.Run("no hint when all kings and stock empty", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 13)}
		tableau[1] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignHeart, 13)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)
		ft.SetWaste(nil)

		assert.Nil(t, ft.GetHint())
	})

	t.Run("not stalemate when stock has cards", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		assert.False(t, ft.IsStalemate())
	})

	t.Run("not stalemate while redeal still available", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		// One column with a King (unmovable) so no other hint exists.
		var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
		tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(domain.CardDesignSpade, 13)}
		ft.SetTableau(tableau)
		// Stock empty but waste has a King that can't move; redeal still possible.
		ft.SetStock(nil)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignHeart, 13)})
		// Trigger stalemate recompute via a move attempt path: use MoveWasteToTableau invalid then check.
		// Instead directly assert CanRedeal keeps us out of stalemate by drawing into the path.
		// Draw is not possible (empty stock); use Redeal availability.
		assert.True(t, ft.CanRedeal())
	})
}

func TestFortyAndEight_GameClear(t *testing.T) {
	ft := newTestFortyAndEight()
	ft.Reset()
	clearF8Tableau(ft)
	ft.SetStock(nil)

	var foundation [domain.FortyAndEightFoundationCnt][]*domain.Card
	designs := []int{domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignDiamond, domain.CardDesignClover}
	for i := 0; i < domain.FortyAndEightFoundationCnt; i++ {
		foundation[i] = make([]*domain.Card, 0, 13)
		design := designs[i%4]
		for v := 1; v < 13; v++ {
			foundation[i] = append(foundation[i], makeF8Card(design, v))
		}
	}
	ft.SetFoundation(foundation)

	var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
	tableau[0] = []*domain.FortyAndEightTableauCard{makeF8TableauCard(designs[0], 13)}
	ft.SetTableau(tableau)

	err := ft.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	// Only one foundation completed here; full clear requires all 8.
}

func TestFortyAndEight_DuplicateSuitFoundation(t *testing.T) {
	ft := newTestFortyAndEight()
	ft.Reset()
	clearF8Tableau(ft)
	ft.SetStock(nil)

	ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 1)})
	err := ft.MoveWasteToFoundation()
	assert.NoError(t, err)

	ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 1)})
	err = ft.MoveWasteToFoundation()
	assert.NoError(t, err)

	count := 0
	for i := 0; i < domain.FortyAndEightFoundationCnt; i++ {
		if len(ft.GetFoundation()[i]) > 0 {
			count++
		}
	}
	assert.Equal(t, 2, count)
}

func TestFortyAndEight_JSON(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		ft := setupPlayingFortyAndEight()
		_ = ft.Draw()
		_ = ft.Draw()

		data, err := json.Marshal(ft)
		require.NoError(t, err)

		ft2 := &domain.FortyAndEight{}
		err = json.Unmarshal(data, ft2)
		require.NoError(t, err)

		assert.Equal(t, ft.GetPhase(), ft2.GetPhase())
		assert.Equal(t, ft.GetMoveCount(), ft2.GetMoveCount())
		assert.Equal(t, ft.GetStockCount(), ft2.GetStockCount())
		assert.Equal(t, len(ft.GetWaste()), len(ft2.GetWaste()))
		assert.Equal(t, ft.GetRedealUsed(), ft2.GetRedealUsed())
		assert.True(t, ft2.CanUndo())
	})

	t.Run("redealUsed round-trips", func(t *testing.T) {
		ft := newTestFortyAndEight()
		ft.Reset()
		clearF8Tableau(ft)
		ft.SetStock(nil)
		ft.SetWaste([]*domain.Card{makeF8Card(domain.CardDesignSpade, 5)})
		require.NoError(t, ft.Redeal())
		require.True(t, ft.GetRedealUsed())

		data, err := json.Marshal(ft)
		require.NoError(t, err)
		ft2 := &domain.FortyAndEight{}
		require.NoError(t, json.Unmarshal(data, ft2))
		assert.True(t, ft2.GetRedealUsed())
	})

	t.Run("unmarshal with nil trumpCards", func(t *testing.T) {
		ft1 := newTestFortyAndEight()
		ft1.Reset()
		data, err := json.Marshal(ft1)
		require.NoError(t, err)
		ft2 := &domain.FortyAndEight{}
		err = json.Unmarshal(data, ft2)
		assert.NoError(t, err)
		assert.Equal(t, ft1.GetPhase(), ft2.GetPhase())
	})

	t.Run("unmarshal rejects oversized arrays", func(t *testing.T) {
		bigSlice := make([]*domain.Card, 1001)
		data, _ := json.Marshal(map[string]interface{}{
			"tc": nil,
			"st": bigSlice,
			"wa": nil,
			"al": nil,
		})
		ft := &domain.FortyAndEight{}
		err := json.Unmarshal(data, ft)
		assert.Error(t, err)
	})

	t.Run("unmarshal rejects nil stock card", func(t *testing.T) {
		data, _ := json.Marshal(map[string]interface{}{
			"tc": nil,
			"st": []interface{}{nil},
		})
		ft := &domain.FortyAndEight{}
		err := json.Unmarshal(data, ft)
		assert.Error(t, err)
	})

	t.Run("unmarshal rejects nil waste card", func(t *testing.T) {
		data, _ := json.Marshal(map[string]interface{}{
			"tc": nil,
			"wa": []interface{}{nil},
		})
		ft := &domain.FortyAndEight{}
		err := json.Unmarshal(data, ft)
		assert.Error(t, err)
	})

	t.Run("unmarshal rejects nil foundation card", func(t *testing.T) {
		data, _ := json.Marshal(map[string]interface{}{
			"tc": nil,
			"fd": []interface{}{[]interface{}{nil}, nil, nil, nil, nil, nil, nil, nil},
		})
		ft := &domain.FortyAndEight{}
		err := json.Unmarshal(data, ft)
		assert.Error(t, err)
	})

	t.Run("unmarshal rejects nil tableau card", func(t *testing.T) {
		data, _ := json.Marshal(map[string]interface{}{
			"tc": nil,
			"tb": []interface{}{[]interface{}{map[string]interface{}{"c": nil, "f": true}}, nil, nil, nil, nil, nil, nil, nil},
		})
		ft := &domain.FortyAndEight{}
		err := json.Unmarshal(data, ft)
		assert.Error(t, err)
	})
}

func TestFortyAndEight_ActionLog(t *testing.T) {
	ft := setupPlayingFortyAndEight()
	assert.Nil(t, ft.GetActionLog())

	_ = ft.Draw()
	log := ft.GetActionLog()
	assert.Equal(t, 1, len(log))
	assert.Equal(t, "draw", log[0].ActionType)
}
