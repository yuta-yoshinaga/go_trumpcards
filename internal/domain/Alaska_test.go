//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

// makeAlaskaTableauCard mirrors makeTableauCard, which returns Klondike's card
// type. Alaska carries its own AlaskaTableauCard because it lives in a different
// worker bucket (extra4), so the shared helper does not typecheck here.
func makeAlaskaTableauCard(design, value int, faceUp bool) *domain.AlaskaTableauCard {
	return &domain.AlaskaTableauCard{Card: makeCard(design, value), FaceUp: faceUp}
}

func newTestAlaska() *domain.Alaska {
	tc := domain.NewTrumpCards(0)
	return domain.NewAlaska(tc)
}

func setupPlayingAlaska() *domain.Alaska {
	r := newTestAlaska()
	r.Reset()
	return r
}

// alaskaDeadTableau returns a tableau with no legal move at all under Alaska's
// rules. Getting this right needs more care than the Russian Solitaire fixture
// it replaces, which only filled 2 of 7 columns:
//
//   - EVERY column must be occupied. Alaska lets any card move to an empty
//     column, so a single empty column means a move always exists.
//   - No two columns may hold the same suit at adjacent ranks, in either
//     direction (Alaska builds up as well as down).
//   - No Aces, or the empty foundations would accept one.
//
// 3/5/7 in two suits plus one spare suit satisfies all three.
func alaskaDeadTableau() [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard {
	spec := []struct {
		design int
		value  int
	}{
		{domain.CardDesignSpade, 3}, {domain.CardDesignSpade, 5}, {domain.CardDesignSpade, 7},
		{domain.CardDesignHeart, 3}, {domain.CardDesignHeart, 5}, {domain.CardDesignHeart, 7},
		{domain.CardDesignClover, 3},
	}
	var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
	for i, c := range spec {
		tab[i] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(c.design, c.value, true)}
	}
	return tab
}

func clearAlaskaTableau(r *domain.Alaska) {
	var empty [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
	r.SetTableau(empty)
}

// --- Tests ---

func TestNewAlaska(t *testing.T) {
	r := newTestAlaska()
	assert.NotNil(t, r)
	assert.Equal(t, domain.AlaskaPhase(0), r.GetPhase())
}

func TestAlaska_Reset(t *testing.T) {
	r := setupPlayingAlaska()

	assert.Equal(t, domain.AlaskaPhasePlaying, r.GetPhase())
	assert.Equal(t, 0, r.GetMoveCount())

	tableau := r.GetTableau()
	totalCards := 0
	// Column 0: 1 card (face-up)
	assert.Equal(t, 1, len(tableau[0]))
	assert.True(t, tableau[0][0].FaceUp)
	totalCards += len(tableau[0])

	// Columns 1-6: i+1 base cards + 4 extra = i+5 cards
	for i := 1; i < domain.AlaskaTableauCnt; i++ {
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
	foundation := r.GetFoundation()
	for i := 0; i < domain.AlaskaFoundationCnt; i++ {
		assert.Nil(t, foundation[i])
	}
}

func TestAlaska_MoveTableauToTableau(t *testing.T) {
	t.Run("valid move - same suit descending", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		// Place a Spade 5 on column 0
		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		// Place a Spade 6 on column 1 (same suit, one rank higher)
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(r.GetTableau()[0]))
		assert.Equal(t, 2, len(r.GetTableau()[1]))
		assert.Equal(t, 1, r.GetMoveCount())
	})

	t.Run("valid move - unordered group (Russian Solitaire special)", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		// Column 0: Spade 5 with unordered cards on top
		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
			makeAlaskaTableauCard(domain.CardDesignHeart, 10, true), // not descending!
			makeAlaskaTableauCard(domain.CardDesignClover, 2, true), // wrong suit!
		}
		// Column 1: Spade 6 - can accept Spade 5 (same-suit rule)
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		// Move Spade 5 + unordered group to column 1
		err := r.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(r.GetTableau()[0]))
		assert.Equal(t, 4, len(r.GetTableau()[1]))
	})

	t.Run("cardIndex -1 shorthand moves top card", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignHeart, 6, true),
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, -1, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(r.GetTableau()[0]))
		assert.Equal(t, 2, len(r.GetTableau()[1]))
	})

	t.Run("King to empty column", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, domain.CardValueMax, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(r.GetTableau()[0]))
		assert.Equal(t, 1, len(r.GetTableau()[1]))
	})

	// Alaska rule 5: any card may go to an empty column. Yukon and Russian
	// Solitaire both restrict this to a King, so this is one of the two places
	// Alaska diverges from the domain it was cloned from.
	t.Run("non-King to empty column succeeds (Alaska rule)", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(r.GetTableau()[0]))
		assert.Equal(t, 1, len(r.GetTableau()[1]))
	})

	// Alaska rule 3: same suit, ascending OR descending. Russian Solitaire
	// allows descending only.
	t.Run("valid move - same suit ASCENDING (Alaska rule)", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 7, true),
		}
		// Spade 6 -- one rank LOWER, so accepting Spade 7 is the ascending case.
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(r.GetTableau()[1]))
	})

	// Negative control: the bidirectional rule must not degrade into "anything
	// goes". A same-suit card two ranks away, and an adjacent card of another
	// suit, both stay rejected.
	t.Run("same suit but not adjacent is rejected", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 8, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		assert.Error(t, r.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("adjacent rank but different suit is rejected", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignHeart, 7, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		assert.Error(t, r.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("face down card cannot be moved", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, false),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Equal(t, "card is face down", err.Error())
	})

	t.Run("different suit fails (Yukon would allow)", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		// Yukon would accept this (alternating colours: Heart 5 -> Spade 6).
		// Russian Solitaire requires the SAME suit, so this must fail.
		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("auto-flip after move", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignClover, 8, false), // face down
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, 1, 1)
		assert.NoError(t, err)
		// The face-down card should now be face-up
		assert.True(t, r.GetTableau()[0][0].FaceUp)
	})

	t.Run("invalid from column", func(t *testing.T) {
		r := setupPlayingAlaska()
		err := r.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = r.MoveTableauToTableau(7, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid to column", func(t *testing.T) {
		r := setupPlayingAlaska()
		err := r.MoveTableauToTableau(0, 0, -1)
		assert.Error(t, err)
		err = r.MoveTableauToTableau(0, 0, 7)
		assert.Error(t, err)
	})

	t.Run("same column", func(t *testing.T) {
		r := setupPlayingAlaska()
		err := r.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		r := setupPlayingAlaska()
		// -1 is valid shorthand for "top card", so test out-of-range.
		err := r.MoveTableauToTableau(0, -2, 1)
		assert.Error(t, err)
		err = r.MoveTableauToTableau(0, 100, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		r := setupPlayingAlaska()
		r.SetPhase(domain.AlaskaPhaseGameOver)
		err := r.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestAlaska_MoveTableauToFoundation(t *testing.T) {
	t.Run("move Ace to empty foundation", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 1, true), // Ace of Spades
		}
		r.SetTableau(tab)

		err := r.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(r.GetTableau()[0]))
		assert.Equal(t, 1, len(r.GetFoundation()[domain.CardDesignSpade-1]))
	})

	t.Run("move sequential card to foundation", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var fd [domain.AlaskaFoundationCnt][]*domain.Card
		fd[domain.CardDesignSpade-1] = []*domain.Card{makeCard(domain.CardDesignSpade, 1)}
		r.SetFoundation(fd)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 2, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(r.GetFoundation()[domain.CardDesignSpade-1]))
	})

	t.Run("wrong value fails", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("empty column fails", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)
		err := r.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		r := setupPlayingAlaska()
		err := r.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = r.MoveTableauToFoundation(7)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		r := setupPlayingAlaska()
		r.SetPhase(domain.AlaskaPhaseGameOver)
		err := r.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("auto-flip after move to foundation", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignClover, 8, false),
			makeAlaskaTableauCard(domain.CardDesignSpade, 1, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToFoundation(0)
		assert.NoError(t, err)
		assert.True(t, r.GetTableau()[0][0].FaceUp)
	})
}

func TestAlaska_GiveUp(t *testing.T) {
	t.Run("give up during play", func(t *testing.T) {
		r := setupPlayingAlaska()
		r.GiveUp()
		assert.Equal(t, domain.AlaskaPhaseGameOver, r.GetPhase())
	})

	t.Run("give up when already over", func(t *testing.T) {
		r := setupPlayingAlaska()
		r.SetPhase(domain.AlaskaPhaseGameClear)
		r.GiveUp()
		assert.Equal(t, domain.AlaskaPhaseGameClear, r.GetPhase())
	})
}

func TestAlaska_GetHint(t *testing.T) {
	t.Run("hint to foundation", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 1, true),
		}
		r.SetTableau(tab)

		hint := r.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
	})

	t.Run("hint to tableau (reveal face-down) - same suit", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignClover, 8, false),
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		hint := r.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, 1, hint.CardIndex)
	})

	t.Run("hint to tableau (all face-up)", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		hint := r.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("no hint when no moves", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		r.SetTableau(alaskaDeadTableau())

		hint := r.GetHint()
		assert.Nil(t, hint)
	})

	// Negative control for the fixture above: it must be dead because of the
	// ranks, not because GetHint is broken. Swapping one card in gives a legal
	// same-suit ascending move, and the hint must reappear.
	t.Run("hint reappears when the dead tableau gains one legal move", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		tab := alaskaDeadTableau()
		// Spade 4 sits adjacent to the Spade 3 in column 0.
		tab[6] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 4, true),
		}
		r.SetTableau(tab)

		assert.NotNil(t, r.GetHint())
	})

	t.Run("no hint when not playing", func(t *testing.T) {
		r := setupPlayingAlaska()
		r.SetPhase(domain.AlaskaPhaseGameOver)
		hint := r.GetHint()
		assert.Nil(t, hint)
	})
}

func TestAlaska_AutoComplete(t *testing.T) {
	t.Run("auto-complete with all face-up", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		// Set up a state where all cards are Ace through King in foundations + a few in tableau
		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		var fd [domain.AlaskaFoundationCnt][]*domain.Card
		// Fill foundations with 12 cards each
		for suit := 1; suit <= 4; suit++ {
			for v := 1; v <= 12; v++ {
				fd[suit-1] = append(fd[suit-1], makeCard(suit, v))
			}
		}
		r.SetFoundation(fd)
		// Put Kings in tableau (all face-up)
		tab[0] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(1, 13, true)}
		tab[1] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(2, 13, true)}
		tab[2] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(3, 13, true)}
		tab[3] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(4, 13, true)}
		r.SetTableau(tab)

		err := r.AutoComplete()
		assert.NoError(t, err)
		assert.Equal(t, domain.AlaskaPhaseGameClear, r.GetPhase())
	})

	t.Run("error when not all face-up", func(t *testing.T) {
		r := setupPlayingAlaska()
		err := r.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("error when not playing", func(t *testing.T) {
		r := setupPlayingAlaska()
		r.SetPhase(domain.AlaskaPhaseGameOver)
		err := r.AutoComplete()
		assert.Error(t, err)
	})
}

func TestAlaska_AllFaceUp(t *testing.T) {
	t.Run("all face-up", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 1, true),
		}
		r.SetTableau(tab)
		assert.True(t, r.AllFaceUp())
	})

	t.Run("not all face-up", func(t *testing.T) {
		r := setupPlayingAlaska()
		assert.False(t, r.AllFaceUp())
	})
}

func TestAlaska_Undo(t *testing.T) {
	t.Run("undo a move", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		r.SetTableau(tab)

		err := r.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(r.GetTableau()[1]))

		err = r.Undo()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(r.GetTableau()[0]))
		assert.Equal(t, 1, len(r.GetTableau()[1]))
	})

	t.Run("no history", func(t *testing.T) {
		r := setupPlayingAlaska()
		err := r.Undo()
		assert.Error(t, err)
	})

	t.Run("not playing", func(t *testing.T) {
		r := setupPlayingAlaska()
		r.SetPhase(domain.AlaskaPhaseGameOver)
		err := r.Undo()
		assert.Error(t, err)
	})
}

func TestAlaska_CanUndo(t *testing.T) {
	r := setupPlayingAlaska()
	assert.False(t, r.CanUndo())
}

func TestAlaska_UndoToEscape(t *testing.T) {
	t.Run("not stalemate", func(t *testing.T) {
		r := setupPlayingAlaska()
		assert.Equal(t, 0, r.UndoToEscape())
	})

	t.Run("stalemate with no escape", func(t *testing.T) {
		r := setupPlayingAlaska()
		r.SetIsStalemate(true)
		assert.Equal(t, -1, r.UndoToEscape())
	})
}

func TestAlaska_UndoN(t *testing.T) {
	t.Run("undo multiple", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		tab[0] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
		}
		tab[2] = []*domain.AlaskaTableauCard{
			makeAlaskaTableauCard(domain.CardDesignSpade, 7, true),
		}
		r.SetTableau(tab)

		// Move: Spade 5 -> Spade 6 (col 0 -> col 1)
		err := r.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		// Move: Spade 6 + Spade 5 -> Spade 7 (col 1 -> col 2)
		err = r.MoveTableauToTableau(1, 0, 2)
		assert.NoError(t, err)

		err = r.UndoN(2)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(r.GetTableau()[0]))
		assert.Equal(t, 1, len(r.GetTableau()[1]))
		assert.Equal(t, 1, len(r.GetTableau()[2]))
	})

	t.Run("undo too many fails", func(t *testing.T) {
		r := setupPlayingAlaska()
		err := r.UndoN(1)
		assert.Error(t, err)
	})
}

func TestAlaska_Stalemate(t *testing.T) {
	t.Run("stalemate when no moves", func(t *testing.T) {
		r := newTestAlaska()
		r.Reset()
		clearAlaskaTableau(r)

		r.SetTableau(alaskaDeadTableau())

		assert.Nil(t, r.GetHint())
		assert.False(t, r.IsStalemate()) // Not set yet - only set after moves
	})
}

func TestAlaska_GameClear(t *testing.T) {
	r := newTestAlaska()
	r.Reset()
	clearAlaskaTableau(r)

	// Set foundations to 12 cards each, then move last card
	var fd [domain.AlaskaFoundationCnt][]*domain.Card
	for suit := 1; suit <= 4; suit++ {
		for v := 1; v <= 12; v++ {
			fd[suit-1] = append(fd[suit-1], makeCard(suit, v))
		}
	}
	r.SetFoundation(fd)

	var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
	tab[0] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(1, 13, true)}
	r.SetTableau(tab)

	err := r.MoveTableauToFoundation(0)
	assert.NoError(t, err)

	// Add remaining kings
	var tab2 [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
	tab2[0] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(2, 13, true)}
	tab2[1] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(3, 13, true)}
	tab2[2] = []*domain.AlaskaTableauCard{makeAlaskaTableauCard(4, 13, true)}
	r.SetTableau(tab2)

	_ = r.MoveTableauToFoundation(0)
	_ = r.MoveTableauToFoundation(1)
	err = r.MoveTableauToFoundation(2)
	assert.NoError(t, err)
	assert.Equal(t, domain.AlaskaPhaseGameClear, r.GetPhase())
}

func TestAlaska_ActionLog(t *testing.T) {
	r := newTestAlaska()
	r.Reset()
	clearAlaskaTableau(r)

	var tab [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
	tab[0] = []*domain.AlaskaTableauCard{
		makeAlaskaTableauCard(domain.CardDesignSpade, 5, true),
	}
	tab[1] = []*domain.AlaskaTableauCard{
		makeAlaskaTableauCard(domain.CardDesignSpade, 6, true),
	}
	r.SetTableau(tab)

	_ = r.MoveTableauToTableau(0, 0, 1)
	log := r.GetActionLog()
	assert.Equal(t, 1, len(log))
	assert.Equal(t, "move", log[0].ActionType)
}

func TestAlaska_MarshalUnmarshalJSON(t *testing.T) {
	r := setupPlayingAlaska()

	data, err := json.Marshal(r)
	assert.NoError(t, err)

	r2 := &domain.Alaska{}
	err = json.Unmarshal(data, r2)
	assert.NoError(t, err)

	assert.Equal(t, r.GetPhase(), r2.GetPhase())
	assert.Equal(t, r.GetMoveCount(), r2.GetMoveCount())

	tab1 := r.GetTableau()
	tab2 := r2.GetTableau()
	for i := 0; i < domain.AlaskaTableauCnt; i++ {
		assert.Equal(t, len(tab1[i]), len(tab2[i]), "column %d", i)
	}
}

func TestAlaska_UnmarshalJSON_invalid(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		r := &domain.Alaska{}
		err := json.Unmarshal([]byte("invalid"), r)
		assert.Error(t, err)
	})

	t.Run("oversized action log", func(t *testing.T) {
		bigLog := make([]*domain.ActionLogEntry, 1001)
		for i := range bigLog {
			bigLog[i] = &domain.ActionLogEntry{}
		}
		data, _ := json.Marshal(map[string]interface{}{
			"al": bigLog,
		})
		r := &domain.Alaska{}
		err := json.Unmarshal(data, r)
		assert.Error(t, err)
	})
}
