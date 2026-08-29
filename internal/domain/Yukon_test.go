//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestYukon() *domain.Yukon {
	tc := domain.NewTrumpCards(0)
	return domain.NewYukon(tc)
}

func setupPlayingYukon() *domain.Yukon {
	y := newTestYukon()
	y.Reset()
	return y
}

func clearYukonTableau(y *domain.Yukon) {
	var empty [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
	y.SetTableau(empty)
}

// --- Tests ---

func TestNewYukon(t *testing.T) {
	y := newTestYukon()
	assert.NotNil(t, y)
	assert.Equal(t, domain.YukonPhase(0), y.GetPhase())
}

func TestYukon_Reset(t *testing.T) {
	y := setupPlayingYukon()

	assert.Equal(t, domain.YukonPhasePlaying, y.GetPhase())
	assert.Equal(t, 0, y.GetMoveCount())

	tableau := y.GetTableau()
	totalCards := 0
	// Column 0: 1 card (face-up)
	assert.Equal(t, 1, len(tableau[0]))
	assert.True(t, tableau[0][0].FaceUp)
	totalCards += len(tableau[0])

	// Columns 1-6: i+1 base cards + 4 extra = i+5 cards
	for i := 1; i < domain.YukonTableauCnt; i++ {
		expected := i + 1 + 4
		assert.Equal(t, expected, len(tableau[i]), "column %d should have %d cards", i, expected)
		// First i cards are face down, rest are face up
		for j, tc := range tableau[i] {
			if j < i {
				assert.False(t, tc.FaceUp, "col %d card %d should be face down", i, j)
			} else {
				assert.True(t, tc.FaceUp, "col %d card %d should be face up", i, j)
			}
		}
		totalCards += len(tableau[i])
	}
	assert.Equal(t, 52, totalCards)

	// Foundation: empty
	foundation := y.GetFoundation()
	for i := 0; i < domain.YukonFoundationCnt; i++ {
		assert.Nil(t, foundation[i])
	}
}

func TestYukon_MoveTableauToTableau(t *testing.T) {
	t.Run("valid move - alternating color descending", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		// Place a red 5 (Heart) on column 0
		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		// Place a black 6 (Spade) on column 1
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(y.GetTableau()[0]))
		assert.Equal(t, 2, len(y.GetTableau()[1]))
		assert.Equal(t, 1, y.GetMoveCount())
	})

	t.Run("valid move - unordered group (Yukon special)", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		// Column 0: red 5 (Heart) with unordered cards on top
		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
			makeTableauCard(domain.CardDesignHeart, 10, true), // not descending!
			makeTableauCard(domain.CardDesignSpade, 2, true),  // same color as 10!
		}
		// Column 1: black 6 (Spade) - can accept red 5
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		y.SetTableau(tab)

		// Move red 5 + unordered group to column 1
		err := y.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(y.GetTableau()[0]))
		assert.Equal(t, 4, len(y.GetTableau()[1]))
	})

	t.Run("cardIndex -1 shorthand moves top card", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignClover, 6, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToTableau(0, -1, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(y.GetTableau()[0]))
		assert.Equal(t, 2, len(y.GetTableau()[1]))
	})

	t.Run("King to empty column", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, domain.CardValueMax, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(y.GetTableau()[0]))
		assert.Equal(t, 1, len(y.GetTableau()[1]))
	})

	t.Run("non-King to empty column fails", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("face down card cannot be moved", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, false),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		// 英語の文面ではなく i18n キーを見る。文面はロケールファイルの担当で、
		// ここで固定すると翻訳を変えるたびにドメインのテストが落ちる (#6327)。
		code, _ := domain.ErrorMessageCode(err)
		assert.Equal(t, "yukon.errCardFaceDown", code)
	})

	t.Run("same color fails", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignClover, 6, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("auto-flip after move", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignClover, 8, false), // face down
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToTableau(0, 1, 1)
		assert.NoError(t, err)
		// The face-down card should now be face-up
		assert.True(t, y.GetTableau()[0][0].FaceUp)
	})

	t.Run("invalid from column", func(t *testing.T) {
		y := setupPlayingYukon()
		err := y.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = y.MoveTableauToTableau(7, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid to column", func(t *testing.T) {
		y := setupPlayingYukon()
		err := y.MoveTableauToTableau(0, 0, -1)
		assert.Error(t, err)
		err = y.MoveTableauToTableau(0, 0, 7)
		assert.Error(t, err)
	})

	t.Run("same column", func(t *testing.T) {
		y := setupPlayingYukon()
		err := y.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		y := setupPlayingYukon()
		// -1 is valid shorthand for "top card" (see "cardIndex -1 shorthand" sub-test above),
		// so test out-of-range values (<-1 and >=len).
		err := y.MoveTableauToTableau(0, -2, 1)
		assert.Error(t, err)
		err = y.MoveTableauToTableau(0, 100, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		y := setupPlayingYukon()
		y.SetPhase(domain.YukonPhaseGameOver)
		err := y.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestYukon_MoveTableauToFoundation(t *testing.T) {
	t.Run("move Ace to empty foundation", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 1, true), // Ace of Spades
		}
		y.SetTableau(tab)

		err := y.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(y.GetTableau()[0]))
		assert.Equal(t, 1, len(y.GetFoundation()[domain.CardDesignSpade-1]))
	})

	t.Run("move sequential card to foundation", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var fd [domain.YukonFoundationCnt][]*domain.Card
		fd[domain.CardDesignSpade-1] = []*domain.Card{makeCard(domain.CardDesignSpade, 1)}
		y.SetFoundation(fd)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 2, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(y.GetFoundation()[domain.CardDesignSpade-1]))
	})

	t.Run("wrong value fails", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("empty column fails", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)
		err := y.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		y := setupPlayingYukon()
		err := y.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = y.MoveTableauToFoundation(7)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		y := setupPlayingYukon()
		y.SetPhase(domain.YukonPhaseGameOver)
		err := y.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("auto-flip after move to foundation", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignClover, 8, false),
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.True(t, y.GetTableau()[0][0].FaceUp)
	})
}

func TestYukon_GiveUp(t *testing.T) {
	t.Run("give up during play", func(t *testing.T) {
		y := setupPlayingYukon()
		y.GiveUp()
		assert.Equal(t, domain.YukonPhaseGameOver, y.GetPhase())
	})

	t.Run("give up when already over", func(t *testing.T) {
		y := setupPlayingYukon()
		y.SetPhase(domain.YukonPhaseGameClear)
		y.GiveUp()
		assert.Equal(t, domain.YukonPhaseGameClear, y.GetPhase())
	})
}

func TestYukon_GetHint(t *testing.T) {
	t.Run("hint to foundation", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		y.SetTableau(tab)

		hint := y.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
	})

	t.Run("hint to tableau (reveal face-down)", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignClover, 8, false),
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		y.SetTableau(tab)

		hint := y.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, 1, hint.CardIndex)
	})

	t.Run("hint to tableau (all face-up)", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		y.SetTableau(tab)

		hint := y.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("no hint when no moves", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		// Only face-down cards -> no moves
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 5, false),
		}
		y.SetTableau(tab)

		hint := y.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("no hint when not playing", func(t *testing.T) {
		y := setupPlayingYukon()
		y.SetPhase(domain.YukonPhaseGameOver)
		hint := y.GetHint()
		assert.Nil(t, hint)
	})
}

func TestYukon_AutoComplete(t *testing.T) {
	t.Run("auto-complete with all face-up", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		// Set up a state where all cards are Ace through King in foundations + a few in tableau
		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		var fd [domain.YukonFoundationCnt][]*domain.Card
		// Fill foundations with 12 cards each
		for suit := 1; suit <= 4; suit++ {
			for v := 1; v <= 12; v++ {
				fd[suit-1] = append(fd[suit-1], makeCard(suit, v))
			}
		}
		y.SetFoundation(fd)
		// Put Kings in tableau (all face-up)
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(1, 13, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(2, 13, true)}
		tab[2] = []*domain.KlondikeTableauCard{makeTableauCard(3, 13, true)}
		tab[3] = []*domain.KlondikeTableauCard{makeTableauCard(4, 13, true)}
		y.SetTableau(tab)

		err := y.AutoComplete()
		assert.NoError(t, err)
		assert.Equal(t, domain.YukonPhaseGameClear, y.GetPhase())
	})

	t.Run("error when not all face-up", func(t *testing.T) {
		y := setupPlayingYukon()
		err := y.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("error when not playing", func(t *testing.T) {
		y := setupPlayingYukon()
		y.SetPhase(domain.YukonPhaseGameOver)
		err := y.AutoComplete()
		assert.Error(t, err)
	})
}

func TestYukon_AllFaceUp(t *testing.T) {
	t.Run("all face-up", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		y.SetTableau(tab)
		assert.True(t, y.AllFaceUp())
	})

	t.Run("not all face-up", func(t *testing.T) {
		y := setupPlayingYukon()
		assert.False(t, y.AllFaceUp())
	})
}

func TestYukon_Undo(t *testing.T) {
	t.Run("undo a move", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		y.SetTableau(tab)

		err := y.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(y.GetTableau()[1]))

		err = y.Undo()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(y.GetTableau()[0]))
		assert.Equal(t, 1, len(y.GetTableau()[1]))
	})

	t.Run("no history", func(t *testing.T) {
		y := setupPlayingYukon()
		err := y.Undo()
		assert.Error(t, err)
	})

	t.Run("not playing", func(t *testing.T) {
		y := setupPlayingYukon()
		y.SetPhase(domain.YukonPhaseGameOver)
		err := y.Undo()
		assert.Error(t, err)
	})
}

func TestYukon_CanUndo(t *testing.T) {
	y := setupPlayingYukon()
	assert.False(t, y.CanUndo())
}

func TestYukon_UndoToEscape(t *testing.T) {
	t.Run("not stalemate", func(t *testing.T) {
		y := setupPlayingYukon()
		assert.Equal(t, 0, y.UndoToEscape())
	})

	t.Run("stalemate with no escape", func(t *testing.T) {
		y := setupPlayingYukon()
		y.SetIsStalemate(true)
		assert.Equal(t, -1, y.UndoToEscape())
	})
}

func TestYukon_UndoN(t *testing.T) {
	t.Run("undo multiple", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		tab[2] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 7, true),
		}
		y.SetTableau(tab)

		// Move: heart 5 -> spade 6 (col 0 -> col 1)
		err := y.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		// Move: spade 6 + heart 5 -> heart 7 (col 1 -> col 2)
		err = y.MoveTableauToTableau(1, 0, 2)
		assert.NoError(t, err)

		err = y.UndoN(2)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(y.GetTableau()[0]))
		assert.Equal(t, 1, len(y.GetTableau()[1]))
		assert.Equal(t, 1, len(y.GetTableau()[2]))
	})

	t.Run("undo too many fails", func(t *testing.T) {
		y := setupPlayingYukon()
		err := y.UndoN(1)
		assert.Error(t, err)
	})
}

func TestYukon_Stalemate(t *testing.T) {
	t.Run("stalemate when no moves", func(t *testing.T) {
		y := newTestYukon()
		y.Reset()
		clearYukonTableau(y)

		// Set up a state with no possible moves
		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		// Two cards of same color and wrong value relation
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 3, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignClover, 5, true),
		}
		y.SetTableau(tab)

		// Trigger stalemate check by making a valid move first
		// Actually, let's just verify the hint is nil to confirm stalemate would be detected
		assert.Nil(t, y.GetHint())
		assert.False(t, y.IsStalemate()) // Not set yet - only set after moves
	})
}

func TestYukon_GameClear(t *testing.T) {
	y := newTestYukon()
	y.Reset()
	clearYukonTableau(y)

	// Set foundations to 12 cards each, then move last card
	var fd [domain.YukonFoundationCnt][]*domain.Card
	for suit := 1; suit <= 4; suit++ {
		for v := 1; v <= 12; v++ {
			fd[suit-1] = append(fd[suit-1], makeCard(suit, v))
		}
	}
	y.SetFoundation(fd)

	var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(1, 13, true)}
	y.SetTableau(tab)

	err := y.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	// Not yet clear - only 1 suit completed
	// Let's complete all 4

	// Add remaining kings
	var tab2 [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
	tab2[0] = []*domain.KlondikeTableauCard{makeTableauCard(2, 13, true)}
	tab2[1] = []*domain.KlondikeTableauCard{makeTableauCard(3, 13, true)}
	tab2[2] = []*domain.KlondikeTableauCard{makeTableauCard(4, 13, true)}
	y.SetTableau(tab2)

	_ = y.MoveTableauToFoundation(0)
	_ = y.MoveTableauToFoundation(1)
	err = y.MoveTableauToFoundation(2)
	assert.NoError(t, err)
	assert.Equal(t, domain.YukonPhaseGameClear, y.GetPhase())
}

func TestYukon_ActionLog(t *testing.T) {
	y := newTestYukon()
	y.Reset()
	clearYukonTableau(y)

	var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{
		makeTableauCard(domain.CardDesignHeart, 5, true),
	}
	tab[1] = []*domain.KlondikeTableauCard{
		makeTableauCard(domain.CardDesignSpade, 6, true),
	}
	y.SetTableau(tab)

	_ = y.MoveTableauToTableau(0, 0, 1)
	log := y.GetActionLog()
	assert.Equal(t, 1, len(log))
	assert.Equal(t, "move", log[0].ActionType)
}

func TestYukon_MarshalUnmarshalJSON(t *testing.T) {
	y := setupPlayingYukon()

	data, err := json.Marshal(y)
	assert.NoError(t, err)

	y2 := &domain.Yukon{}
	err = json.Unmarshal(data, y2)
	assert.NoError(t, err)

	assert.Equal(t, y.GetPhase(), y2.GetPhase())
	assert.Equal(t, y.GetMoveCount(), y2.GetMoveCount())

	tab1 := y.GetTableau()
	tab2 := y2.GetTableau()
	for i := 0; i < domain.YukonTableauCnt; i++ {
		assert.Equal(t, len(tab1[i]), len(tab2[i]), "column %d", i)
	}
}

func TestYukon_UnmarshalJSON_invalid(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		y := &domain.Yukon{}
		err := json.Unmarshal([]byte("invalid"), y)
		assert.Error(t, err)
	})

	t.Run("oversized action log", func(t *testing.T) {
		// Create JSON with too many action log entries
		bigLog := make([]*domain.ActionLogEntry, 1001)
		for i := range bigLog {
			bigLog[i] = &domain.ActionLogEntry{}
		}
		data, _ := json.Marshal(map[string]interface{}{
			"al": bigLog,
		})
		y := &domain.Yukon{}
		err := json.Unmarshal(data, y)
		assert.Error(t, err)
	})
}

// **CUI のエラー行は i18n コードが無いと英語のまま出る。** `cuiErrorBlock` は
// `ErrorMessageCode(lastErr)` が空のとき `lastErr.Error()` をそのまま印字するので、
// 素の `errors.New` は日本語ロケールのプレイヤーに英語で届く (#6327)。Web は
// クライアント側でボタンを無効化して翻訳済みツールチップを出すため、これは
// CUI 固有の漏れだった。
//
// 個々の文言ではなく**「コードを持っているか」**を見る ── 文言は翻訳ファイルの
// 担当で、ここで固定すると二重管理になる。
func TestYukon_ErrorsCarryAnI18nCode(t *testing.T) {
	codeOf := func(t *testing.T, err error) string {
		t.Helper()
		if err == nil {
			return ""
		}
		code, _ := domain.ErrorMessageCode(err)
		return code
	}

	t.Run("auto-complete refused because cards are still face down", func(t *testing.T) {
		y := setupPlayingYukon()
		clearYukonTableau(y)
		var tab [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 1, false)}
		y.SetTableau(tab)
		require.False(t, y.AllFaceUp(), "この盤で AllFaceUp が真だと、測りたい枝に入らない")

		err := y.AutoComplete()
		require.Error(t, err)
		assert.Equal(t, "yukon.errNotAllFaceUp", codeOf(t, err))
	})

	// 同じ漏れは AutoComplete だけではない。Yukon のドメインエラーは全部
	// この経路で画面に出るので、**1つだけ直すと残りが英語のまま残る**。
	t.Run("every refusal names a key instead of an English sentence", func(t *testing.T) {
		cases := []struct {
			name string
			run  func(y *domain.Yukon) error
		}{
			{"move from a column that does not exist", func(y *domain.Yukon) error {
				return y.MoveTableauToTableau(-1, 0, 0)
			}},
			{"move to a column that does not exist", func(y *domain.Yukon) error {
				return y.MoveTableauToTableau(0, 0, domain.YukonTableauCnt)
			}},
			{"move a column onto itself", func(y *domain.Yukon) error {
				return y.MoveTableauToTableau(0, 0, 0)
			}},
			{"send a card up from a column that does not exist", func(y *domain.Yukon) error {
				return y.MoveTableauToFoundation(-1)
			}},
			{"undo with nothing to undo", func(y *domain.Yukon) error {
				return y.Undo()
			}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				y := setupPlayingYukon()
				err := tc.run(y)
				require.Error(t, err, "この操作は拒否されるはずで、拒否されないと何も測れない")
				code := codeOf(t, err)
				assert.NotEmpty(t, code, "コードが無いと CUI は英語をそのまま出す")
				assert.Truef(t, strings.HasPrefix(code, "yukon."),
					"キーは yukon 名前空間に置く (got %q)", code)
			})
		}
	})
}
