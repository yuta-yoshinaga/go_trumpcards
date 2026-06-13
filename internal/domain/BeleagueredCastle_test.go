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

func newTestBeleagueredCastle() *domain.BeleagueredCastle {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	return domain.NewBeleagueredCastle(tc)
}

func setupPlayingBeleagueredCastle() *domain.BeleagueredCastle {
	bc := newTestBeleagueredCastle()
	bc.Reset()
	return bc
}

func makeBCCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeBCTableauCard(design, value int) *domain.BeleagueredCastleTableauCard {
	return &domain.BeleagueredCastleTableauCard{Card: makeBCCard(design, value), FaceUp: true}
}

func clearBCTableau(bc *domain.BeleagueredCastle) {
	var empty [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
	bc.SetTableau(empty)
}

func TestNewBeleagueredCastle(t *testing.T) {
	bc := newTestBeleagueredCastle()
	assert.NotNil(t, bc)
	assert.Equal(t, domain.BeleagueredCastlePhase(0), bc.GetPhase())
}

func TestBeleagueredCastle_Reset(t *testing.T) {
	bc := setupPlayingBeleagueredCastle()

	assert.Equal(t, domain.BeleagueredCastlePhasePlaying, bc.GetPhase())
	assert.Equal(t, 0, bc.GetMoveCount())

	// Foundations: each suit foundation pre-seeded with its Ace.
	foundation := bc.GetFoundation()
	totalFoundationCards := 0
	for i := range domain.BeleagueredCastleFoundationCnt {
		require.Equal(t, 1, len(foundation[i]), "foundation %d must hold the suit's Ace", i)
		assert.Equal(t, 1, foundation[i][0].GetValue(), "foundation %d must be an Ace", i)
		totalFoundationCards++
	}
	assert.Equal(t, domain.BeleagueredCastleFoundationCnt, totalFoundationCards)

	// Tableau: 8 columns × 6 cards, all face-up, sums to 48.
	tableau := bc.GetTableau()
	totalTableauCards := 0
	for i := range domain.BeleagueredCastleTableauCnt {
		assert.Equal(t, domain.BeleagueredCastleColumnLen, len(tableau[i]),
			"column %d should have 6 cards", i)
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "all cards should be face up")
			assert.NotEqual(t, 1, tc.Card.GetValue(),
				"Aces must be pulled to the foundation, not in tableau")
		}
		totalTableauCards += len(tableau[i])
	}
	assert.Equal(t, 48, totalTableauCards)
}

func TestBeleagueredCastle_MoveTableauToTableau(t *testing.T) {
	t.Run("valid single card move descending any suit", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(bc.GetTableau()[0]))
		assert.Equal(t, 2, len(bc.GetTableau()[1]))
	})

	t.Run("reject multi-card move", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{
			makeBCTableauCard(domain.CardDesignSpade, 6),
			makeBCTableauCard(domain.CardDesignSpade, 5),
		}
		tableau[1] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 7)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only the top card can be moved")
	})

	t.Run("reject same-rank move", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("allow move to empty column", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		// Empty columns accept any card in Beleaguered Castle.
		err := bc.MoveTableauToTableau(0, 0, 1)
		require.NoError(t, err)
		assert.Equal(t, 0, len(bc.GetTableau()[0]))
		assert.Equal(t, 1, len(bc.GetTableau()[1]))
	})

	t.Run("same column", func(t *testing.T) {
		bc := setupPlayingBeleagueredCastle()
		err := bc.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		bc := setupPlayingBeleagueredCastle()
		err := bc.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = bc.MoveTableauToTableau(0, 0, domain.BeleagueredCastleTableauCnt)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		bc := setupPlayingBeleagueredCastle()
		err := bc.MoveTableauToTableau(0, 99, 1)
		assert.Error(t, err)
	})

	t.Run("cardIndex -1 resolves to top card", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(0, -1, 1))
		assert.Equal(t, 0, len(bc.GetTableau()[0]))
	})

	t.Run("cardIndex -1 errors on empty column", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		err := bc.MoveTableauToTableau(0, -1, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.SetPhase(domain.BeleagueredCastlePhaseGameOver)
		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestBeleagueredCastle_MoveTableauToFoundation(t *testing.T) {
	t.Run("place 2 on suit Ace", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 2)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToFoundation(0))
		// Spade Ace lives on foundation 0 by the suit ordering in Reset.
		assert.Equal(t, 2, len(bc.GetFoundation()[0]))
	})

	t.Run("cannot place non-Two on Ace-only foundation", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		bc := setupPlayingBeleagueredCastle()
		err := bc.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = bc.MoveTableauToFoundation(domain.BeleagueredCastleTableauCnt)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.SetPhase(domain.BeleagueredCastlePhaseGameClear)
		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

func TestBeleagueredCastle_GameClear(t *testing.T) {
	bc := newTestBeleagueredCastle()
	bc.Reset()
	clearBCTableau(bc)

	// Pre-fill foundations with Ace..Queen.
	var foundation [domain.BeleagueredCastleFoundationCnt][]*domain.Card
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	for i, s := range suits {
		pile := make([]*domain.Card, 0, domain.CardValueMax)
		for v := 1; v < domain.CardValueMax; v++ {
			pile = append(pile, makeBCCard(s, v))
		}
		foundation[i] = pile
	}
	bc.SetFoundation(foundation)

	// Place 4 kings on tableau, then drop them all to foundation.
	var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
	for i, s := range suits {
		tableau[i] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(s, domain.CardValueMax)}
	}
	bc.SetTableau(tableau)

	for i := range suits {
		require.NoError(t, bc.MoveTableauToFoundation(i))
	}
	assert.Equal(t, domain.BeleagueredCastlePhaseGameClear, bc.GetPhase())
}

func TestBeleagueredCastle_GiveUp(t *testing.T) {
	bc := setupPlayingBeleagueredCastle()
	bc.GiveUp()
	assert.Equal(t, domain.BeleagueredCastlePhaseGameOver, bc.GetPhase())
	assert.True(t, bc.GetGameEndFlag())

	// Calling GiveUp again is a no-op.
	bc.GiveUp()
	assert.Equal(t, domain.BeleagueredCastlePhaseGameOver, bc.GetPhase())
}

func TestBeleagueredCastle_Hint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.SetPhase(domain.BeleagueredCastlePhaseGameOver)
		assert.Nil(t, bc.GetHint())
	})

	t.Run("priority foundation move", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		// Foundations already hold Spade Ace on pile 0; a Spade 2 can drop.
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 2)}
		bc.SetTableau(tableau)

		hint := bc.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
	})

	t.Run("tableau-to-tableau when no foundation move", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		// Wipe foundations so no foundation move is possible.
		var foundation [domain.BeleagueredCastleFoundationCnt][]*domain.Card
		bc.SetFoundation(foundation)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		hint := bc.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("nil when stalemate", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		// All columns hold a 5 — no tableau-to-tableau move is legal (no
		// descending pair) and the Ace-only foundations cannot accept a 5.
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		suits := []int{
			domain.CardDesignSpade, domain.CardDesignSpade,
			domain.CardDesignClover, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignHeart,
			domain.CardDesignDiamond, domain.CardDesignDiamond,
		}
		for i, s := range suits {
			tableau[i] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(s, 5)}
		}
		bc.SetTableau(tableau)
		assert.Nil(t, bc.GetHint())
	})
}

func TestBeleagueredCastle_AutoComplete(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.SetPhase(domain.BeleagueredCastlePhaseGameOver)
		err := bc.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("re-evaluates stalemate after partial completion", func(t *testing.T) {
		// Build a board where AutoComplete drops a single Two onto its Ace but
		// leaves the remaining columns dead (only 5s). After the partial drop
		// the new state has no legal moves, so isStalemate must flip to true.
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		bc.SetIsStalemate(false)

		// Foundations: pre-seeded Aces.
		suits := []int{
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		var foundation [domain.BeleagueredCastleFoundationCnt][]*domain.Card
		for i, s := range suits {
			foundation[i] = []*domain.Card{makeBCCard(s, 1)}
		}
		bc.SetFoundation(foundation)

		// Column 0 holds [5♣, 2♠] so AutoComplete drops the Two onto foundation
		// 0 but leaves a buried 5 behind. The other 7 columns each hold a
		// single 5. After AutoComplete every column still holds a 5 (no empty
		// column to dump into) and there are no 6s or 3s available, so the
		// position is genuinely stuck.
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{
			makeBCTableauCard(domain.CardDesignClover, 5),
			makeBCTableauCard(domain.CardDesignSpade, 2),
		}
		fiveSuits := []int{
			domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignDiamond,
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		for i, s := range fiveSuits {
			tableau[i+1] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(s, 5)}
		}
		bc.SetTableau(tableau)

		require.NoError(t, bc.AutoComplete())
		assert.Equal(t, domain.BeleagueredCastlePhasePlaying, bc.GetPhase())
		assert.True(t, bc.IsStalemate(), "stalemate must be re-evaluated after partial AutoComplete")
	})

	t.Run("clears all to foundation when fully orderable", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var foundation [domain.BeleagueredCastleFoundationCnt][]*domain.Card
		suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i, s := range suits {
			pile := make([]*domain.Card, 0)
			for v := 1; v < domain.CardValueMax; v++ {
				pile = append(pile, makeBCCard(s, v))
			}
			foundation[i] = pile
		}
		bc.SetFoundation(foundation)

		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		for i, s := range suits {
			tableau[i] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(s, domain.CardValueMax)}
		}
		bc.SetTableau(tableau)

		require.NoError(t, bc.AutoComplete())
		assert.Equal(t, domain.BeleagueredCastlePhaseGameClear, bc.GetPhase())
	})
}

func TestBeleagueredCastle_Undo(t *testing.T) {
	t.Run("undo restores previous state", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(0, 0, 1))
		assert.True(t, bc.CanUndo())
		require.NoError(t, bc.Undo())
		assert.Equal(t, 1, len(bc.GetTableau()[0]))
		assert.Equal(t, 1, len(bc.GetTableau()[1]))
	})

	t.Run("undo with no history", func(t *testing.T) {
		bc := setupPlayingBeleagueredCastle()
		// Reset wipes history; undo with no recorded moves must fail.
		err := bc.Undo()
		assert.Error(t, err)
	})

	t.Run("undo when not playing", func(t *testing.T) {
		bc := setupPlayingBeleagueredCastle()
		bc.GiveUp()
		err := bc.Undo()
		assert.Error(t, err)
	})

	t.Run("undoN", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 5)}
		tableau[2] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 6)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(1, 0, 2))
		require.NoError(t, bc.MoveTableauToTableau(0, 0, 2))

		require.NoError(t, bc.UndoN(2))
		assert.Equal(t, 0, bc.GetMoveCount())
	})
}

func TestBeleagueredCastle_UndoToEscape(t *testing.T) {
	t.Run("not stalemate", func(t *testing.T) {
		// Deterministic: UndoToEscape returns 0 whenever the game is not in a
		// stalemate. Setting the flag explicitly avoids the rare shuffled deal
		// that is an immediate stalemate (which made this flake on CI).
		bc := setupPlayingBeleagueredCastle()
		bc.SetIsStalemate(false)
		assert.Equal(t, 0, bc.UndoToEscape())
	})

	t.Run("returns -1 when no escape", func(t *testing.T) {
		bc := newTestBeleagueredCastle()
		bc.Reset()
		clearBCTableau(bc)
		bc.SetIsStalemate(true)
		assert.Equal(t, -1, bc.UndoToEscape())
	})
}

func TestBeleagueredCastle_JSON(t *testing.T) {
	bc := setupPlayingBeleagueredCastle()
	data, err := json.Marshal(bc)
	require.NoError(t, err)

	bc2 := newTestBeleagueredCastle()
	err = json.Unmarshal(data, bc2)
	require.NoError(t, err)

	assert.Equal(t, bc.GetPhase(), bc2.GetPhase())
	assert.Equal(t, bc.GetMoveCount(), bc2.GetMoveCount())
}

func TestBeleagueredCastle_NewDefault(t *testing.T) {
	bc := domain.NewDefaultBeleagueredCastle()
	assert.NotNil(t, bc)
	bc.Reset()
	assert.Equal(t, domain.BeleagueredCastlePhasePlaying, bc.GetPhase())
}

func TestBeleagueredCastle_ActionLog(t *testing.T) {
	bc := newTestBeleagueredCastle()
	bc.Reset()
	clearBCTableau(bc)
	var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
	tableau[0] = []*domain.BeleagueredCastleTableauCard{makeBCTableauCard(domain.CardDesignSpade, 2)}
	bc.SetTableau(tableau)

	require.NoError(t, bc.MoveTableauToFoundation(0))
	log := bc.GetActionLog()
	assert.NotEmpty(t, log)
}
