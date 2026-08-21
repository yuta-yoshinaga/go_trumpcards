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

func newTestFortress() *domain.Fortress {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	return domain.NewFortress(tc)
}

func setupPlayingFortress() *domain.Fortress {
	bc := newTestFortress()
	bc.Reset()
	return bc
}

func makeFortressCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeFortressTableauCard(design, value int) *domain.FortressTableauCard {
	return &domain.FortressTableauCard{Card: makeFortressCard(design, value), FaceUp: true}
}

func clearFortressTableau(bc *domain.Fortress) {
	var empty [domain.FortressTableauCnt][]*domain.FortressTableauCard
	bc.SetTableau(empty)
}

func TestNewFortress(t *testing.T) {
	bc := newTestFortress()
	assert.NotNil(t, bc)
	assert.Equal(t, domain.FortressPhase(0), bc.GetPhase())
}

func TestFortress_Reset(t *testing.T) {
	f := setupPlayingFortress()

	assert.Equal(t, domain.FortressPhasePlaying, f.GetPhase())
	assert.Equal(t, 0, f.GetMoveCount())

	// Fortress deals the WHOLE deck to the tableau. Beleaguered Castle, which
	// this was cloned from, pulls the four Aces onto the foundations first --
	// so an empty foundation here is the thing that separates the two.
	for i := range domain.FortressFoundationCnt {
		assert.Empty(t, f.GetFoundation()[i], "foundation %d must start empty", i)
	}

	// 52 cards over 10 columns, all face-up. 52 = 5*10 + 2, so the round-robin
	// deal gives the first two columns 6 and the remaining eight 5.
	tableau := f.GetTableau()
	total, sixes := 0, 0
	for i := range domain.FortressTableauCnt {
		n := len(tableau[i])
		assert.Contains(t, []int{5, 6}, n, "column %d should hold 5 or 6 cards, got %d", i, n)
		if n == 6 {
			sixes++
		}
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "every Fortress card is dealt face up")
		}
		total += n
	}
	assert.Equal(t, 52, total, "the whole deck is dealt to the tableau")
	assert.Equal(t, 2, sixes, "exactly two columns get the two spare cards")
	assert.Equal(t, 10, domain.FortressTableauCnt)
}

// seedFortressAce puts the suit's Ace on foundation fIdx. Fortress deals every
// card to the tableau, so any test about building a foundation has to place the
// Ace itself -- Beleaguered Castle's Reset did that for free.
func seedFortressAce(f *domain.Fortress, fIdx, design int) {
	foundation := f.GetFoundation()
	foundation[fIdx] = []*domain.Card{makeFortressCard(design, 1)}
	f.SetFoundation(foundation)
}

// fortressDeadTableau returns a tableau with no legal move under Fortress rules.
// The Beleaguered Castle fixture this replaces filled 8 of its 8 columns, but
// Fortress has TEN, and an empty column accepts any card -- so the cloned
// version left two empty columns and a move always existed.
//
// Every column is occupied, every card is a 5 (same rank never connects, in
// either direction), and there is no Ace, so the empty foundations stay shut.
func fortressDeadTableau() [domain.FortressTableauCnt][]*domain.FortressTableauCard {
	var tab [domain.FortressTableauCnt][]*domain.FortressTableauCard
	designs := []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	}
	for i := range domain.FortressTableauCnt {
		tab[i] = []*domain.FortressTableauCard{makeFortressTableauCard(designs[i%len(designs)], 5)}
	}
	return tab
}

type cardSpec struct {
	design int
	value  int
}

// twoColumnFortress puts one card in column 0 and one in column 1 and empties
// the rest, so a move from 0 to 1 exercises exactly the tableau rule.
func twoColumnFortress(t *testing.T, from, to cardSpec) *domain.Fortress {
	t.Helper()
	f := newTestFortress()
	f.Reset()
	clearFortressTableau(f)
	var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
	tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(from.design, from.value)}
	tableau[1] = []*domain.FortressTableauCard{makeFortressTableauCard(to.design, to.value)}
	f.SetTableau(tableau)
	return f
}

func TestFortress_MoveTableauToTableau(t *testing.T) {
	// Fortress builds by SUIT, in either direction. Beleaguered Castle -- the
	// domain this was cloned from -- ignores suit entirely and only builds down,
	// so both halves of that rule are a real divergence and each gets a test
	// plus a negative control.
	t.Run("valid single card move - same suit descending", func(t *testing.T) {
		f := twoColumnFortress(t, cardSpec{domain.CardDesignSpade, 4}, cardSpec{domain.CardDesignSpade, 5})
		assert.NoError(t, f.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 0, len(f.GetTableau()[0]))
		assert.Equal(t, 2, len(f.GetTableau()[1]))
	})

	t.Run("valid single card move - same suit ASCENDING", func(t *testing.T) {
		f := twoColumnFortress(t, cardSpec{domain.CardDesignSpade, 6}, cardSpec{domain.CardDesignSpade, 5})
		assert.NoError(t, f.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 2, len(f.GetTableau()[1]))
	})

	t.Run("reject adjacent rank of a DIFFERENT suit", func(t *testing.T) {
		// Beleaguered Castle would allow this; Fortress must not.
		f := twoColumnFortress(t, cardSpec{domain.CardDesignHeart, 4}, cardSpec{domain.CardDesignSpade, 5})
		assert.Error(t, f.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("reject same suit two ranks apart", func(t *testing.T) {
		// The bidirectional rule must not degrade into "same suit, anything".
		f := twoColumnFortress(t, cardSpec{domain.CardDesignSpade, 3}, cardSpec{domain.CardDesignSpade, 5})
		assert.Error(t, f.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("reject multi-card move", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		tableau[0] = []*domain.FortressTableauCard{
			makeFortressTableauCard(domain.CardDesignSpade, 6),
			makeFortressTableauCard(domain.CardDesignSpade, 5),
		}
		tableau[1] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 7)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only the top card can be moved")
	})

	t.Run("reject same-rank move", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("allow move to empty column", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		// Empty columns accept any card in Beleaguered Castle.
		err := bc.MoveTableauToTableau(0, 0, 1)
		require.NoError(t, err)
		assert.Equal(t, 0, len(bc.GetTableau()[0]))
		assert.Equal(t, 1, len(bc.GetTableau()[1]))
	})

	t.Run("same column", func(t *testing.T) {
		bc := setupPlayingFortress()
		err := bc.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		bc := setupPlayingFortress()
		err := bc.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = bc.MoveTableauToTableau(0, 0, domain.FortressTableauCnt)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		bc := setupPlayingFortress()
		err := bc.MoveTableauToTableau(0, 99, 1)
		assert.Error(t, err)
	})

	t.Run("cardIndex -1 resolves to top card", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		// Same suit: Fortress builds by suit, unlike Beleaguered Castle.
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(0, -1, 1))
		assert.Equal(t, 0, len(bc.GetTableau()[0]))
	})

	t.Run("cardIndex -1 errors on empty column", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		err := bc.MoveTableauToTableau(0, -1, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		bc := newTestFortress()
		bc.SetPhase(domain.FortressPhaseGameOver)
		err := bc.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestFortress_MoveTableauToFoundation(t *testing.T) {
	t.Run("place 2 on suit Ace", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 2)}
		bc.SetTableau(tableau)
		// Fortress starts with EMPTY foundations (Beleaguered Castle pre-seeds the
		// Aces), so the Ace has to be placed here for the 2 to have a home.
		seedFortressAce(bc, 0, domain.CardDesignSpade)

		require.NoError(t, bc.MoveTableauToFoundation(0))
		assert.Equal(t, 2, len(bc.GetFoundation()[0]))
	})

	t.Run("cannot place non-Two on Ace-only foundation", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		bc := setupPlayingFortress()
		err := bc.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = bc.MoveTableauToFoundation(domain.FortressTableauCnt)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		bc := newTestFortress()
		bc.SetPhase(domain.FortressPhaseGameClear)
		err := bc.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

func TestFortress_GameClear(t *testing.T) {
	bc := newTestFortress()
	bc.Reset()
	clearFortressTableau(bc)

	// Pre-fill foundations with Ace..Queen.
	var foundation [domain.FortressFoundationCnt][]*domain.Card
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	for i, s := range suits {
		pile := make([]*domain.Card, 0, domain.CardValueMax)
		for v := 1; v < domain.CardValueMax; v++ {
			pile = append(pile, makeFortressCard(s, v))
		}
		foundation[i] = pile
	}
	bc.SetFoundation(foundation)

	// Place 4 kings on tableau, then drop them all to foundation.
	var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
	for i, s := range suits {
		tableau[i] = []*domain.FortressTableauCard{makeFortressTableauCard(s, domain.CardValueMax)}
	}
	bc.SetTableau(tableau)

	for i := range suits {
		require.NoError(t, bc.MoveTableauToFoundation(i))
	}
	assert.Equal(t, domain.FortressPhaseGameClear, bc.GetPhase())
}

func TestFortress_GiveUp(t *testing.T) {
	bc := setupPlayingFortress()
	bc.GiveUp()
	assert.Equal(t, domain.FortressPhaseGameOver, bc.GetPhase())
	assert.True(t, bc.GetGameEndFlag())

	// Calling GiveUp again is a no-op.
	bc.GiveUp()
	assert.Equal(t, domain.FortressPhaseGameOver, bc.GetPhase())
}

func TestFortress_Hint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		bc := newTestFortress()
		bc.SetPhase(domain.FortressPhaseGameOver)
		assert.Nil(t, bc.GetHint())
	})

	t.Run("priority foundation move", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 2)}
		bc.SetTableau(tableau)
		seedFortressAce(bc, 0, domain.CardDesignSpade)

		hint := bc.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
	})

	t.Run("tableau-to-tableau when no foundation move", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		// Wipe foundations so no foundation move is possible.
		var foundation [domain.FortressFoundationCnt][]*domain.Card
		bc.SetFoundation(foundation)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		// Same suit: Fortress builds by suit, unlike Beleaguered Castle.
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		hint := bc.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("nil when stalemate", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		// All columns hold a 5 — no tableau-to-tableau move is legal (no
		// descending pair) and the Ace-only foundations cannot accept a 5.
		bc.SetTableau(fortressDeadTableau())
		assert.Nil(t, bc.GetHint())
	})

	// Negative control for the fixture above: it must be dead because of the
	// ranks, not because GetHint is broken. One adjacent same-suit card and the
	// hint has to come back.
	t.Run("hint reappears when the dead tableau gains one legal move", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		tab := fortressDeadTableau()
		tab[9] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 6)}
		bc.SetTableau(tab)
		assert.NotNil(t, bc.GetHint())
	})
}

func TestFortress_AutoComplete(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		bc := newTestFortress()
		bc.SetPhase(domain.FortressPhaseGameOver)
		err := bc.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("re-evaluates stalemate after partial completion", func(t *testing.T) {
		// Build a board where AutoComplete drops a single Two onto its Ace but
		// leaves the remaining columns dead (only 5s). After the partial drop
		// the new state has no legal moves, so isStalemate must flip to true.
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		bc.SetIsStalemate(false)

		// Foundations: pre-seeded Aces.
		suits := []int{
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		var foundation [domain.FortressFoundationCnt][]*domain.Card
		for i, s := range suits {
			foundation[i] = []*domain.Card{makeFortressCard(s, 1)}
		}
		bc.SetFoundation(foundation)

		// Column 0 holds [5♣, 2♠] so AutoComplete drops the Two onto foundation
		// 0 but leaves a buried 5 behind. The other NINE columns each hold a
		// single 5 -- all ten must be occupied, because Fortress lets any card
		// move to an empty column, and the Beleaguered Castle version of this
		// fixture (8 columns) left two of Fortress's ten empty, which kept a
		// legal move alive and stopped the stalemate from ever being reached.
		// After AutoComplete every column still holds a 5 and no 6 or 3 exists,
		// so the position is genuinely stuck.
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		tableau[0] = []*domain.FortressTableauCard{
			makeFortressTableauCard(domain.CardDesignClover, 5),
			makeFortressTableauCard(domain.CardDesignSpade, 2),
		}
		designs := []int{
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		for i := 1; i < domain.FortressTableauCnt; i++ {
			tableau[i] = []*domain.FortressTableauCard{
				makeFortressTableauCard(designs[i%len(designs)], 5),
			}
		}
		bc.SetTableau(tableau)

		require.NoError(t, bc.AutoComplete())
		assert.Equal(t, domain.FortressPhasePlaying, bc.GetPhase())
		assert.True(t, bc.IsStalemate(), "stalemate must be re-evaluated after partial AutoComplete")
	})

	t.Run("clears all to foundation when fully orderable", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var foundation [domain.FortressFoundationCnt][]*domain.Card
		suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i, s := range suits {
			pile := make([]*domain.Card, 0)
			for v := 1; v < domain.CardValueMax; v++ {
				pile = append(pile, makeFortressCard(s, v))
			}
			foundation[i] = pile
		}
		bc.SetFoundation(foundation)

		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		for i, s := range suits {
			tableau[i] = []*domain.FortressTableauCard{makeFortressTableauCard(s, domain.CardValueMax)}
		}
		bc.SetTableau(tableau)

		require.NoError(t, bc.AutoComplete())
		assert.Equal(t, domain.FortressPhaseGameClear, bc.GetPhase())
	})
}

func TestFortress_Undo(t *testing.T) {
	t.Run("undo restores previous state", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		// Same suit: Fortress builds by suit, unlike Beleaguered Castle.
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 5)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(0, 0, 1))
		assert.True(t, bc.CanUndo())
		require.NoError(t, bc.Undo())
		assert.Equal(t, 1, len(bc.GetTableau()[0]))
		assert.Equal(t, 1, len(bc.GetTableau()[1]))
	})

	t.Run("undo with no history", func(t *testing.T) {
		bc := setupPlayingFortress()
		// Reset wipes history; undo with no recorded moves must fail.
		err := bc.Undo()
		assert.Error(t, err)
	})

	t.Run("undo when not playing", func(t *testing.T) {
		bc := setupPlayingFortress()
		bc.GiveUp()
		err := bc.Undo()
		assert.Error(t, err)
	})

	t.Run("undoN", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		// Same suit: Fortress builds by suit, unlike Beleaguered Castle.
		tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 5)}
		tableau[2] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 6)}
		bc.SetTableau(tableau)

		require.NoError(t, bc.MoveTableauToTableau(1, 0, 2))
		require.NoError(t, bc.MoveTableauToTableau(0, 0, 2))

		require.NoError(t, bc.UndoN(2))
		assert.Equal(t, 0, bc.GetMoveCount())
	})
}

func TestFortress_UndoToEscape(t *testing.T) {
	t.Run("not stalemate", func(t *testing.T) {
		// Deterministic: UndoToEscape returns 0 whenever the game is not in a
		// stalemate. Setting the flag explicitly avoids the rare shuffled deal
		// that is an immediate stalemate (which made this flake on CI).
		bc := setupPlayingFortress()
		bc.SetIsStalemate(false)
		assert.Equal(t, 0, bc.UndoToEscape())
	})

	t.Run("returns -1 when no escape", func(t *testing.T) {
		bc := newTestFortress()
		bc.Reset()
		clearFortressTableau(bc)
		bc.SetIsStalemate(true)
		assert.Equal(t, -1, bc.UndoToEscape())
	})
}

func TestFortress_JSON(t *testing.T) {
	bc := setupPlayingFortress()
	data, err := json.Marshal(bc)
	require.NoError(t, err)

	bc2 := newTestFortress()
	err = json.Unmarshal(data, bc2)
	require.NoError(t, err)

	assert.Equal(t, bc.GetPhase(), bc2.GetPhase())
	assert.Equal(t, bc.GetMoveCount(), bc2.GetMoveCount())
}

func TestFortress_NewDefault(t *testing.T) {
	bc := domain.NewDefaultFortress()
	assert.NotNil(t, bc)
	bc.Reset()
	assert.Equal(t, domain.FortressPhasePlaying, bc.GetPhase())
}

func TestFortress_ActionLog(t *testing.T) {
	bc := newTestFortress()
	bc.Reset()
	clearFortressTableau(bc)
	var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
	tableau[0] = []*domain.FortressTableauCard{makeFortressTableauCard(domain.CardDesignSpade, 2)}
	bc.SetTableau(tableau)
	// Fortress foundations start empty, so the Ace has to be placed first.
	seedFortressAce(bc, 0, domain.CardDesignSpade)

	require.NoError(t, bc.MoveTableauToFoundation(0))
	log := bc.GetActionLog()
	assert.NotEmpty(t, log)
}
