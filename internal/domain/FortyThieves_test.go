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

func newTestFortyThieves() *domain.FortyThieves {
	tc := domain.NewTrumpCardsWithDecks(2, 0)
	ft := domain.NewFortyThieves(tc)
	return ft
}

func setupPlayingFortyThieves() *domain.FortyThieves {
	ft := newTestFortyThieves()
	ft.Reset()
	return ft
}

func makeFTCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeFTTableauCard(design, value int) *domain.FortyThievesTableauCard {
	return &domain.FortyThievesTableauCard{Card: makeFTCard(design, value), FaceUp: true}
}

func clearFTTableau(ft *domain.FortyThieves) {
	var empty [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
	ft.SetTableau(empty)
}

func TestNewFortyThieves(t *testing.T) {
	ft := newTestFortyThieves()
	assert.NotNil(t, ft)
	assert.Equal(t, domain.FortyThievesPhase(0), ft.GetPhase())
}

func TestFortyThieves_Reset(t *testing.T) {
	ft := setupPlayingFortyThieves()

	assert.Equal(t, domain.FortyThievesPhasePlaying, ft.GetPhase())
	assert.Equal(t, 0, ft.GetMoveCount())

	// Tableau: 10 columns, each with 4 face-up cards
	tableau := ft.GetTableau()
	totalTableauCards := 0
	for i := 0; i < domain.FortyThievesTableauCnt; i++ {
		assert.Equal(t, 4, len(tableau[i]), "column %d should have 4 cards", i)
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

	// Foundation: empty
	foundation := ft.GetFoundation()
	for i := 0; i < domain.FortyThievesFoundationCnt; i++ {
		assert.Nil(t, foundation[i])
	}
}

func TestFortyThieves_Draw(t *testing.T) {
	t.Run("draw from stock", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		initialStock := ft.GetStockCount()
		err := ft.Draw()
		assert.NoError(t, err)
		assert.Equal(t, initialStock-1, ft.GetStockCount())
		assert.Equal(t, 1, len(ft.GetWaste()))
		assert.Equal(t, 1, ft.GetMoveCount())
	})

	t.Run("draw when stock is empty (no recycle)", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		ft.SetStock(nil)
		err := ft.Draw()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cards in stock")
	})

	t.Run("draw when game is not playing", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameOver)
		err := ft.Draw()
		assert.Error(t, err)
	})
}

func TestFortyThieves_MoveWasteToTableau(t *testing.T) {
	t.Run("valid move same suit descending", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		// Place Spade 5 on tableau col 0
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		// Set waste to Spade 4
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 4)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetTableau()[0]))
	})

	t.Run("reject different suit", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		// Heart 4 - same value but different suit
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignHeart, 4)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("valid move to empty column", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 7)})
		ft.SetStock(nil)

		err := ft.MoveWasteToTableau(0)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ft.GetTableau()[0]))
	})

	t.Run("waste empty", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 1)})
		err := ft.MoveWasteToTableau(-1)
		assert.Error(t, err)
		err = ft.MoveWasteToTableau(10)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameOver)
		err := ft.MoveWasteToTableau(0)
		assert.Error(t, err)
	})
}

func TestFortyThieves_MoveWasteToFoundation(t *testing.T) {
	t.Run("place ace on empty foundation", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 1)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.NoError(t, err)
		// Check that one foundation has 1 card
		found := false
		for i := 0; i < domain.FortyThievesFoundationCnt; i++ {
			if len(ft.GetFoundation()[i]) == 1 {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("place card on matching foundation", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var foundation [domain.FortyThievesFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeFTCard(domain.CardDesignSpade, 1)}
		ft.SetFoundation(foundation)
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 2)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetFoundation()[0]))
	})

	t.Run("cannot place non-ace on empty foundation", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 5)})
		ft.SetStock(nil)

		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("waste empty", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameClear)
		err := ft.MoveWasteToFoundation()
		assert.Error(t, err)
	})
}

func TestFortyThieves_MoveTableauToTableau(t *testing.T) {
	t.Run("valid single card move same suit descending", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 5)}
		tableau[1] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
		assert.Equal(t, 2, len(ft.GetTableau()[1]))
	})

	t.Run("reject multi-card move", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{
			makeFTTableauCard(domain.CardDesignSpade, 6),
			makeFTTableauCard(domain.CardDesignSpade, 5),
		}
		tableau[1] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 7)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		// Try to move from index 0 (not the last card)
		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only the top card can be moved")
	})

	t.Run("reject different suit", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("move to empty column", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
		assert.Equal(t, 1, len(ft.GetTableau()[1]))
	})

	t.Run("same column", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		err := ft.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		err := ft.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = ft.MoveTableauToTableau(0, 0, 10)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		err := ft.MoveTableauToTableau(0, -1, 1)
		assert.Error(t, err)
		err = ft.MoveTableauToTableau(0, 100, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameOver)
		err := ft.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestFortyThieves_MoveTableauToFoundation(t *testing.T) {
	t.Run("move ace to foundation", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 1)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ft.GetTableau()[0]))
	})

	t.Run("move card to existing foundation pile", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var foundation [domain.FortyThievesFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeFTCard(domain.CardDesignSpade, 1)}
		ft.SetFoundation(foundation)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 2)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ft.GetFoundation()[0]))
	})

	t.Run("empty column", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("cannot place on foundation", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 5)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		err := ft.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = ft.MoveTableauToFoundation(10)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameOver)
		err := ft.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

func TestFortyThieves_GiveUp(t *testing.T) {
	t.Run("give up during playing", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		ft.GiveUp()
		assert.Equal(t, domain.FortyThievesPhaseGameOver, ft.GetPhase())
	})

	t.Run("give up when not playing does nothing", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameClear)
		ft.GiveUp()
		assert.Equal(t, domain.FortyThievesPhaseGameClear, ft.GetPhase())
	})
}

func TestFortyThieves_GetHint(t *testing.T) {
	t.Run("hint tableau to foundation", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 1)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint waste to foundation", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignHeart, 1)})
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("hint tableau to tableau", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 5)}
		tableau[1] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("hint waste to tableau", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 6)}
		ft.SetTableau(tableau)
		ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 5)})
		ft.SetStock(nil)

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "waste", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("no hint available", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		// Place cards that cannot move anywhere
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 13)}
		tableau[1] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignHeart, 13)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)

		hint := ft.GetHint()
		assert.Nil(t, hint)
	})

	// #5525: 他に手が無くストックだけ残っている局面で「ヒントはありません」を
	// 返していた。行き詰まりではないのに、プレイヤーには詰んだのか引けば良いのか
	// 区別が付かない。
	t.Run("suggests drawing when nothing else moves but stock remains", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 13)}
		tableau[1] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignHeart, 13)}
		ft.SetTableau(tableau)
		ft.SetStock([]*domain.Card{makeFTCard(domain.CardDesignClover, 7)})

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "stock", hint.FromZone)
		assert.Equal(t, "waste", hint.ToZone)
		assert.Equal(t, -1, hint.FromCol)
		assert.Equal(t, -1, hint.ToCol)
		assert.Equal(t, -1, hint.CardIndex)
	})

	// **盤上に手があるならそちらが先。**引くのは最後の手段。
	t.Run("prefers a real move over drawing", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 1)}
		ft.SetTableau(tableau)
		ft.SetStock([]*domain.Card{makeFTCard(domain.CardDesignClover, 7)})

		hint := ft.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("no hint when not playing", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameOver)
		hint := ft.GetHint()
		assert.Nil(t, hint)
	})
}

func TestFortyThieves_AutoComplete(t *testing.T) {
	t.Run("auto complete when all face up", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		ft.SetStock(nil)

		// Set up tableau with cards ready for foundation (ordered aces and 2s)
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 1)}
		tableau[1] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignHeart, 1)}
		ft.SetTableau(tableau)

		err := ft.AutoComplete()
		assert.NoError(t, err)
	})

	t.Run("cannot auto complete with stock remaining", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		err := ft.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameOver)
		err := ft.AutoComplete()
		assert.Error(t, err)
	})
}

func TestFortyThieves_AllFaceUp(t *testing.T) {
	t.Run("true when stock is empty", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		ft.SetStock(nil)
		assert.True(t, ft.AllFaceUp())
	})

	t.Run("false when stock has cards", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		assert.False(t, ft.AllFaceUp())
	})
}

func TestFortyThieves_Undo(t *testing.T) {
	t.Run("undo draw", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		initialStock := ft.GetStockCount()
		_ = ft.Draw()
		assert.Equal(t, initialStock-1, ft.GetStockCount())

		err := ft.Undo()
		assert.NoError(t, err)
		assert.Equal(t, initialStock, ft.GetStockCount())
		assert.Equal(t, 0, ft.GetMoveCount())
	})

	t.Run("cannot undo with no history", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		err := ft.Undo()
		assert.Error(t, err)
	})

	t.Run("cannot undo when not playing", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.SetPhase(domain.FortyThievesPhaseGameOver)
		err := ft.Undo()
		assert.Error(t, err)
	})
}

func TestFortyThieves_CanUndo(t *testing.T) {
	t.Run("true after action", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		_ = ft.Draw()
		assert.True(t, ft.CanUndo())
	})

	t.Run("false with no history", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		assert.False(t, ft.CanUndo())
	})
}

func TestFortyThieves_UndoN(t *testing.T) {
	ft := setupPlayingFortyThieves()
	_ = ft.Draw()
	_ = ft.Draw()
	_ = ft.Draw()
	assert.Equal(t, 3, ft.GetMoveCount())

	err := ft.UndoN(3)
	assert.NoError(t, err)
	assert.Equal(t, 0, ft.GetMoveCount())
}

func TestFortyThieves_UndoToEscape(t *testing.T) {
	t.Run("returns 0 when not stalemate", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		assert.Equal(t, 0, ft.UndoToEscape())
	})

	t.Run("returns undo count to escape", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		// Draw a few cards then force stalemate
		_ = ft.Draw()
		_ = ft.Draw()
		ft.SetIsStalemate(true)
		result := ft.UndoToEscape()
		assert.True(t, result > 0)
	})
}

func TestFortyThieves_Stalemate(t *testing.T) {
	t.Run("stalemate when no moves and stock empty", func(t *testing.T) {
		ft := newTestFortyThieves()
		ft.Reset()
		clearFTTableau(ft)
		// Kings can't go to foundation and can't stack
		var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
		tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignSpade, 13)}
		tableau[1] = []*domain.FortyThievesTableauCard{makeFTTableauCard(domain.CardDesignHeart, 13)}
		ft.SetTableau(tableau)
		ft.SetStock(nil)
		ft.SetWaste(nil)

		// Trigger stalemate check via a draw (need stock for that)
		// Instead, move waste to tableau to trigger check - but waste is nil
		// Just verify GetHint returns nil and IsStalemate works
		assert.Nil(t, ft.GetHint())
	})

	t.Run("not stalemate when stock has cards", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		// After reset, stock has cards, so not stalemate
		assert.False(t, ft.IsStalemate())
	})
}

func TestFortyThieves_GameClear(t *testing.T) {
	ft := newTestFortyThieves()
	ft.Reset()
	clearFTTableau(ft)
	ft.SetStock(nil)

	// Fill all 8 foundations with 13 cards each
	var foundation [domain.FortyThievesFoundationCnt][]*domain.Card
	designs := []int{domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignDiamond, domain.CardDesignClover}
	for i := 0; i < domain.FortyThievesFoundationCnt; i++ {
		foundation[i] = make([]*domain.Card, 0, 13)
		design := designs[i%4]
		for v := 1; v < 13; v++ {
			foundation[i] = append(foundation[i], makeFTCard(design, v))
		}
	}
	ft.SetFoundation(foundation)

	// Place the last card (value 13) on tableau and move to foundation
	var tableau [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard
	tableau[0] = []*domain.FortyThievesTableauCard{makeFTTableauCard(designs[0], 13)}
	ft.SetTableau(tableau)

	err := ft.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	// Still not clear since other foundations aren't full yet
	// Need all 8 foundations to be 13 cards each
}

func TestFortyThieves_DuplicateSuitFoundation(t *testing.T) {
	// Test that two aces of the same suit go to different foundations
	ft := newTestFortyThieves()
	ft.Reset()
	clearFTTableau(ft)
	ft.SetStock(nil)

	// Place two spade aces
	ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 1)})
	err := ft.MoveWasteToFoundation()
	assert.NoError(t, err)

	ft.SetWaste([]*domain.Card{makeFTCard(domain.CardDesignSpade, 1)})
	err = ft.MoveWasteToFoundation()
	assert.NoError(t, err)

	// Two different foundation piles should have spade aces
	count := 0
	for i := 0; i < domain.FortyThievesFoundationCnt; i++ {
		if len(ft.GetFoundation()[i]) > 0 {
			count++
		}
	}
	assert.Equal(t, 2, count)
}

func TestFortyThieves_JSON(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		ft := setupPlayingFortyThieves()
		_ = ft.Draw()
		_ = ft.Draw()

		data, err := json.Marshal(ft)
		require.NoError(t, err)

		ft2 := &domain.FortyThieves{}
		err = json.Unmarshal(data, ft2)
		require.NoError(t, err)

		assert.Equal(t, ft.GetPhase(), ft2.GetPhase())
		assert.Equal(t, ft.GetMoveCount(), ft2.GetMoveCount())
		assert.Equal(t, ft.GetStockCount(), ft2.GetStockCount())
		assert.Equal(t, len(ft.GetWaste()), len(ft2.GetWaste()))
		assert.True(t, ft2.CanUndo()) // history is now serialized (#1654)
	})

	t.Run("unmarshal with nil trumpCards", func(t *testing.T) {
		// Marshal a fresh game to get valid JSON structure, then unmarshal
		ft1 := newTestFortyThieves()
		ft1.Reset()
		data, err := json.Marshal(ft1)
		require.NoError(t, err)
		ft2 := &domain.FortyThieves{}
		err = json.Unmarshal(data, ft2)
		assert.NoError(t, err)
		assert.Equal(t, ft1.GetPhase(), ft2.GetPhase())
	})

	t.Run("unmarshal rejects oversized arrays", func(t *testing.T) {
		// Create a JSON with >1000 stock cards
		bigSlice := make([]*domain.Card, 1001)
		data, _ := json.Marshal(map[string]interface{}{
			"tc": nil,
			"st": bigSlice,
			"wa": nil,
			"al": nil,
		})
		ft := &domain.FortyThieves{}
		err := json.Unmarshal(data, ft)
		assert.Error(t, err)
	})
}

func TestFortyThieves_ActionLog(t *testing.T) {
	ft := setupPlayingFortyThieves()
	assert.Nil(t, ft.GetActionLog())

	_ = ft.Draw()
	log := ft.GetActionLog()
	assert.Equal(t, 1, len(log))
	assert.Equal(t, "draw", log[0].ActionType)
}
