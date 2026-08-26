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

func newTestSomerset() *domain.Somerset {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	return domain.NewSomerset(tc)
}

func setupPlayingSomerset() *domain.Somerset {
	bc := newTestSomerset()
	bc.Reset()
	return bc
}

func makeSomersetCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeSomersetTableauCard(design, value int) *domain.SomersetTableauCard {
	return &domain.SomersetTableauCard{Card: makeSomersetCard(design, value), FaceUp: true}
}

func clearSomersetTableau(bc *domain.Somerset) {
	var empty [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
	bc.SetTableau(empty)
}

func TestNewSomerset(t *testing.T) {
	bc := newTestSomerset()
	assert.NotNil(t, bc)
	assert.Equal(t, domain.SomersetPhase(0), bc.GetPhase())
}

func TestSomerset_Reset(t *testing.T) {
	f := setupPlayingSomerset()

	assert.Equal(t, domain.SomersetPhasePlaying, f.GetPhase())
	assert.Equal(t, 0, f.GetMoveCount())

	// Somerset deals the WHOLE deck to the tableau. Beleaguered Castle, which
	// this was cloned from, pulls the four Aces onto the foundations first --
	// so an empty foundation here is the thing that separates the two.
	for i := range domain.SomersetFoundationCnt {
		assert.Empty(t, f.GetFoundation()[i], "foundation %d must start empty", i)
	}

	// 52 cards over 10 columns, all face-up. 52 = 5*10 + 2, so the round-robin
	// deal gives the first two columns 6 and the remaining eight 5.
	tableau := f.GetTableau()
	total, sixes := 0, 0
	for i := range domain.SomersetTableauCnt {
		n := len(tableau[i])
		assert.Contains(t, []int{5, 6}, n, "column %d should hold 5 or 6 cards, got %d", i, n)
		if n == 6 {
			sixes++
		}
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "every Somerset card is dealt face up")
		}
		total += n
	}
	assert.Equal(t, 52, total, "the whole deck is dealt to the tableau")
	assert.Equal(t, 2, sixes, "exactly two columns get the two spare cards")
	assert.Equal(t, 10, domain.SomersetTableauCnt)
}

// seedSomersetAce puts the suit's Ace on foundation fIdx. Somerset deals every
// card to the tableau, so any test about building a foundation has to place the
// Ace itself -- Beleaguered Castle's Reset did that for free.
func seedSomersetAce(f *domain.Somerset, fIdx, design int) {
	foundation := f.GetFoundation()
	foundation[fIdx] = []*domain.Card{makeSomersetCard(design, 1)}
	f.SetFoundation(foundation)
}

// somersetDeadTableau returns a tableau with no legal move under Somerset rules.
// The Beleaguered Castle fixture this replaces filled 8 of its 8 columns, but
// Somerset has TEN, and an empty column accepts any card -- so the cloned
// version left two empty columns and a move always existed.
//
// Every column is occupied, every card is a 5 (same rank never connects, in
// either direction), and there is no Ace, so the empty foundations stay shut.
func somersetDeadTableau() [domain.SomersetTableauCnt][]*domain.SomersetTableauCard {
	var tab [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
	designs := []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	}
	for i := range domain.SomersetTableauCnt {
		tab[i] = []*domain.SomersetTableauCard{makeSomersetTableauCard(designs[i%len(designs)], 5)}
	}
	return tab
}

type somersetCardSpec struct {
	design int
	value  int
}

// twoColumnSomerset puts one card in column 0 and one in column 1 and empties
// the rest, so a move from 0 to 1 exercises exactly the tableau rule.
func twoColumnSomerset(t *testing.T, from, to somersetCardSpec) *domain.Somerset {
	t.Helper()
	f := newTestSomerset()
	f.Reset()
	clearSomersetTableau(f)
	var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
	tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(from.design, from.value)}
	tableau[1] = []*domain.SomersetTableauCard{makeSomersetTableauCard(to.design, to.value)}
	f.SetTableau(tableau)
	return f
}

func TestSomerset_MoveTableauToTableau(t *testing.T) {
	// Somerset builds by ALTERNATING COLOUR, descending only. Fortress -- the
	// domain this was cloned from -- builds by SUIT in either direction, so both
	// halves of that rule are a divergence and each gets a test plus a control.
	t.Run("valid single card move - alternating colour descending", func(t *testing.T) {
		// black 6 onto red 7
		f := twoColumnSomerset(t, somersetCardSpec{domain.CardDesignSpade, 6}, somersetCardSpec{domain.CardDesignHeart, 7})
		assert.NoError(t, f.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 0, len(f.GetTableau()[0]))
		assert.Equal(t, 2, len(f.GetTableau()[1]))
	})

	t.Run("reject SAME colour descending", func(t *testing.T) {
		// black 6 onto black 7: Fortress cared about suit, Somerset about colour.
		f := twoColumnSomerset(t, somersetCardSpec{domain.CardDesignSpade, 6}, somersetCardSpec{domain.CardDesignClover, 7})
		assert.Error(t, f.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("reject same suit descending", func(t *testing.T) {
		// Fortress's headline legal move must NOT be legal here.
		f := twoColumnSomerset(t, somersetCardSpec{domain.CardDesignSpade, 4}, somersetCardSpec{domain.CardDesignSpade, 5})
		assert.Error(t, f.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("reject ASCENDING even with alternating colour", func(t *testing.T) {
		// Somerset builds down only; Fortress allowed either direction.
		f := twoColumnSomerset(t, somersetCardSpec{domain.CardDesignSpade, 8}, somersetCardSpec{domain.CardDesignHeart, 7})
		assert.Error(t, f.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("reject alternating colour two ranks apart", func(t *testing.T) {
		f := twoColumnSomerset(t, somersetCardSpec{domain.CardDesignSpade, 5}, somersetCardSpec{domain.CardDesignHeart, 7})
		assert.Error(t, f.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("reject multi-card move", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		tableau[0] = []*domain.SomersetTableauCard{
			makeSomersetTableauCard(domain.CardDesignSpade, 6),
			makeSomersetTableauCard(domain.CardDesignSpade, 5),
		}
		tableau[1] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 7)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only the top card can be moved")
	})

	t.Run("reject same-rank move", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("allow move to empty column", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		// Empty columns accept any card in Beleaguered Castle.
		err := bc.MoveTableauToTableau(0, 0, 1)
		require.NoError(t, err)
		assert.Equal(t, 0, len(bc.GetTableau()[0]))
		assert.Equal(t, 1, len(bc.GetTableau()[1]))
	})

	t.Run("same column", func(t *testing.T) {
		bc := setupPlayingSomerset()
		err := bc.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		bc := setupPlayingSomerset()
		err := bc.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = bc.MoveTableauToTableau(0, 0, domain.SomersetTableauCnt)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		bc := setupPlayingSomerset()
		err := bc.MoveTableauToTableau(0, 99, 1)
		assert.Error(t, err)
	})

	t.Run("cardIndex -1 resolves to top card", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		// Alternating colour, descending: black 4 onto red 5. Fortress used a
		// same-suit pair here, which Somerset rejects.
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignHeart, 5)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(0, -1, 1))
		assert.Equal(t, 0, len(bc.GetTableau()[0]))
	})

	t.Run("cardIndex -1 errors on empty column", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		err := bc.MoveTableauToTableau(0, -1, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		bc := newTestSomerset()
		bc.SetPhase(domain.SomersetPhaseGameOver)
		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestSomerset_MoveTableauToFoundation(t *testing.T) {
	t.Run("place 2 on suit Ace", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 2)}
		bc.SetTableau(tableau)
		// Somerset starts with EMPTY foundations (Beleaguered Castle pre-seeds the
		// Aces), so the Ace has to be placed here for the 2 to have a home.
		seedSomersetAce(bc, 0, domain.CardDesignSpade)

		require.NoError(t, bc.MoveTableauToFoundation(0))
		assert.Equal(t, 2, len(bc.GetFoundation()[0]))
	})

	t.Run("cannot place non-Two on Ace-only foundation", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		bc := setupPlayingSomerset()
		err := bc.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = bc.MoveTableauToFoundation(domain.SomersetTableauCnt)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		bc := newTestSomerset()
		bc.SetPhase(domain.SomersetPhaseGameClear)
		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

func TestSomerset_GameClear(t *testing.T) {
	bc := newTestSomerset()
	bc.Reset()
	clearSomersetTableau(bc)

	// Pre-fill foundations with Ace..Queen.
	var foundation [domain.SomersetFoundationCnt][]*domain.Card
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	for i, s := range suits {
		pile := make([]*domain.Card, 0, domain.CardValueMax)
		for v := 1; v < domain.CardValueMax; v++ {
			pile = append(pile, makeSomersetCard(s, v))
		}
		foundation[i] = pile
	}
	bc.SetFoundation(foundation)

	// Place 4 kings on tableau, then drop them all to foundation.
	var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
	for i, s := range suits {
		tableau[i] = []*domain.SomersetTableauCard{makeSomersetTableauCard(s, domain.CardValueMax)}
	}
	bc.SetTableau(tableau)

	for i := range suits {
		require.NoError(t, bc.MoveTableauToFoundation(i))
	}
	assert.Equal(t, domain.SomersetPhaseGameClear, bc.GetPhase())
}

func TestSomerset_GiveUp(t *testing.T) {
	bc := setupPlayingSomerset()
	bc.GiveUp()
	assert.Equal(t, domain.SomersetPhaseGameOver, bc.GetPhase())
	assert.True(t, bc.GetGameEndFlag())

	// Calling GiveUp again is a no-op.
	bc.GiveUp()
	assert.Equal(t, domain.SomersetPhaseGameOver, bc.GetPhase())
}

func TestSomerset_Hint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		bc := newTestSomerset()
		bc.SetPhase(domain.SomersetPhaseGameOver)
		assert.Nil(t, bc.GetHint())
	})

	t.Run("priority foundation move", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 2)}
		bc.SetTableau(tableau)
		seedSomersetAce(bc, 0, domain.CardDesignSpade)

		hint := bc.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
	})

	t.Run("tableau-to-tableau when no foundation move", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		// Wipe foundations so no foundation move is possible.
		var foundation [domain.SomersetFoundationCnt][]*domain.Card
		bc.SetFoundation(foundation)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		// Alternating colour, descending: black 4 onto red 5. Fortress used a
		// same-suit pair here, which Somerset rejects.
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignHeart, 5)}
		bc.SetTableau(tableau)

		hint := bc.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("nil when stalemate", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		// All columns hold a 5 — no tableau-to-tableau move is legal (no
		// descending pair) and the Ace-only foundations cannot accept a 5.
		bc.SetTableau(somersetDeadTableau())
		assert.Nil(t, bc.GetHint())
	})

	// Negative control for the fixture above: it must be dead because of the
	// ranks, not because GetHint is broken. One adjacent same-suit card and the
	// hint has to come back.
	t.Run("hint reappears when the dead tableau gains one legal move", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		tab := somersetDeadTableau()
		tab[9] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 6)}
		bc.SetTableau(tab)
		assert.NotNil(t, bc.GetHint())
	})
}

func TestSomerset_AutoComplete(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		bc := newTestSomerset()
		bc.SetPhase(domain.SomersetPhaseGameOver)
		err := bc.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("re-evaluates stalemate after partial completion", func(t *testing.T) {
		// Build a board where AutoComplete drops a single Two onto its Ace but
		// leaves the remaining columns dead (only 5s). After the partial drop
		// the new state has no legal moves, so isStalemate must flip to true.
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		bc.SetIsStalemate(false)

		// Foundations: pre-seeded Aces.
		suits := []int{
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		var foundation [domain.SomersetFoundationCnt][]*domain.Card
		for i, s := range suits {
			foundation[i] = []*domain.Card{makeSomersetCard(s, 1)}
		}
		bc.SetFoundation(foundation)

		// Column 0 holds [5♣, 2♠] so AutoComplete drops the Two onto foundation
		// 0 but leaves a buried 5 behind. The other NINE columns each hold a
		// single 5 -- all ten must be occupied, because Somerset lets any card
		// move to an empty column, and the Beleaguered Castle version of this
		// fixture (8 columns) left two of Somerset's ten empty, which kept a
		// legal move alive and stopped the stalemate from ever being reached.
		// After AutoComplete every column still holds a 5 and no 6 or 3 exists,
		// so the position is genuinely stuck.
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		tableau[0] = []*domain.SomersetTableauCard{
			makeSomersetTableauCard(domain.CardDesignClover, 5),
			makeSomersetTableauCard(domain.CardDesignSpade, 2),
		}
		designs := []int{
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		for i := 1; i < domain.SomersetTableauCnt; i++ {
			tableau[i] = []*domain.SomersetTableauCard{
				makeSomersetTableauCard(designs[i%len(designs)], 5),
			}
		}
		bc.SetTableau(tableau)

		require.NoError(t, bc.AutoComplete())
		assert.Equal(t, domain.SomersetPhasePlaying, bc.GetPhase())
		assert.True(t, bc.IsStalemate(), "stalemate must be re-evaluated after partial AutoComplete")
	})

	t.Run("clears all to foundation when fully orderable", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var foundation [domain.SomersetFoundationCnt][]*domain.Card
		suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i, s := range suits {
			pile := make([]*domain.Card, 0)
			for v := 1; v < domain.CardValueMax; v++ {
				pile = append(pile, makeSomersetCard(s, v))
			}
			foundation[i] = pile
		}
		bc.SetFoundation(foundation)

		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		for i, s := range suits {
			tableau[i] = []*domain.SomersetTableauCard{makeSomersetTableauCard(s, domain.CardValueMax)}
		}
		bc.SetTableau(tableau)

		require.NoError(t, bc.AutoComplete())
		assert.Equal(t, domain.SomersetPhaseGameClear, bc.GetPhase())
	})
}

func TestSomerset_Undo(t *testing.T) {
	t.Run("undo restores previous state", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		// Alternating colour, descending: black 4 onto red 5. Fortress used a
		// same-suit pair here, which Somerset rejects.
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignHeart, 5)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(0, 0, 1))
		assert.True(t, bc.CanUndo())
		require.NoError(t, bc.Undo())
		assert.Equal(t, 1, len(bc.GetTableau()[0]))
		assert.Equal(t, 1, len(bc.GetTableau()[1]))
	})

	t.Run("undo with no history", func(t *testing.T) {
		bc := setupPlayingSomerset()
		// Reset wipes history; undo with no recorded moves must fail.
		err := bc.Undo()
		assert.Error(t, err)
	})

	t.Run("undo when not playing", func(t *testing.T) {
		bc := setupPlayingSomerset()
		bc.GiveUp()
		err := bc.Undo()
		assert.Error(t, err)
	})

	t.Run("undoN", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
		// Alternating colour, descending: black 4 onto red 5. Fortress used a
		// same-suit pair here, which Somerset rejects.
		tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignHeart, 5)}
		tableau[2] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 6)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(1, 0, 2))
		require.NoError(t, bc.MoveTableauToTableau(0, 0, 2))

		require.NoError(t, bc.UndoN(2))
		assert.Equal(t, 0, bc.GetMoveCount())
	})
}

func TestSomerset_UndoToEscape(t *testing.T) {
	t.Run("not stalemate", func(t *testing.T) {
		// Deterministic: UndoToEscape returns 0 whenever the game is not in a
		// stalemate. Setting the flag explicitly avoids the rare shuffled deal
		// that is an immediate stalemate (which made this flake on CI).
		bc := setupPlayingSomerset()
		bc.SetIsStalemate(false)
		assert.Equal(t, 0, bc.UndoToEscape())
	})

	t.Run("returns -1 when no escape", func(t *testing.T) {
		bc := newTestSomerset()
		bc.Reset()
		clearSomersetTableau(bc)
		bc.SetIsStalemate(true)
		assert.Equal(t, -1, bc.UndoToEscape())
	})
}

func TestSomerset_JSON(t *testing.T) {
	bc := setupPlayingSomerset()
	data, err := json.Marshal(bc)
	require.NoError(t, err)

	bc2 := newTestSomerset()
	err = json.Unmarshal(data, bc2)
	require.NoError(t, err)

	assert.Equal(t, bc.GetPhase(), bc2.GetPhase())
	assert.Equal(t, bc.GetMoveCount(), bc2.GetMoveCount())
}

func TestSomerset_NewDefault(t *testing.T) {
	bc := domain.NewDefaultSomerset()
	assert.NotNil(t, bc)
	bc.Reset()
	assert.Equal(t, domain.SomersetPhasePlaying, bc.GetPhase())
}

func TestSomerset_ActionLog(t *testing.T) {
	bc := newTestSomerset()
	bc.Reset()
	clearSomersetTableau(bc)
	var tableau [domain.SomersetTableauCnt][]*domain.SomersetTableauCard
	tableau[0] = []*domain.SomersetTableauCard{makeSomersetTableauCard(domain.CardDesignSpade, 2)}
	bc.SetTableau(tableau)
	// Somerset foundations start empty, so the Ace has to be placed first.
	seedSomersetAce(bc, 0, domain.CardDesignSpade)

	require.NoError(t, bc.MoveTableauToFoundation(0))
	log := bc.GetActionLog()
	assert.NotEmpty(t, log)
}
