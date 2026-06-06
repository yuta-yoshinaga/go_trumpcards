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

func newTestEasthaven() *domain.Easthaven {
	return domain.NewEasthaven(domain.NewTrumpCards(0))
}

func setupPlayingEasthaven() *domain.Easthaven {
	e := newTestEasthaven()
	e.Reset()
	return e
}

func clearEasthavenTableau(e *domain.Easthaven) {
	var empty [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
	e.SetTableau(empty)
}

// fullFoundationPile returns an A..maxValue same-suit pile.
func fullFoundationPile(design, maxValue int) []*domain.Card {
	pile := make([]*domain.Card, 0, maxValue)
	for v := 1; v <= maxValue; v++ {
		pile = append(pile, domain.NewCard(design, v, true))
	}
	return pile
}

// --- Tests ---

func TestNewEasthaven(t *testing.T) {
	e := newTestEasthaven()
	assert.NotNil(t, e)
	assert.Equal(t, domain.EasthavenPhase(0), e.GetPhase())
	assert.NotNil(t, domain.NewDefaultEasthaven())
}

func TestEasthaven_Reset(t *testing.T) {
	e := setupPlayingEasthaven()

	assert.Equal(t, domain.EasthavenPhasePlaying, e.GetPhase())
	assert.Equal(t, 0, e.GetMoveCount())
	assert.False(t, e.IsStalemate())

	tableau := e.GetTableau()
	total := 0
	for i := range domain.EasthavenTableauCnt {
		assert.Equal(t, domain.EasthavenInitialColCards, len(tableau[i]))
		for j, tc := range tableau[i] {
			if j == domain.EasthavenInitialColCards-1 {
				assert.True(t, tc.FaceUp, "last card face up")
			} else {
				assert.False(t, tc.FaceUp, "earlier cards face down")
			}
		}
		total += len(tableau[i])
	}
	assert.Equal(t, domain.EasthavenTableauCnt*domain.EasthavenInitialColCards, total)

	// Stock holds the remaining 31 cards.
	assert.Equal(t, 52-total, e.GetStockCount())

	foundation := e.GetFoundation()
	for i := range domain.EasthavenFoundationCnt {
		assert.Nil(t, foundation[i])
	}
}

func TestEasthaven_Deal(t *testing.T) {
	t.Run("deals one card to each column", func(t *testing.T) {
		e := setupPlayingEasthaven()
		before := e.GetStockCount()
		err := e.Deal()
		require.NoError(t, err)
		assert.Equal(t, before-domain.EasthavenTableauCnt, e.GetStockCount())
		assert.Equal(t, 1, e.GetMoveCount())
		for i := range domain.EasthavenTableauCnt {
			col := e.GetTableau()[i]
			assert.True(t, col[len(col)-1].FaceUp, "dealt card is face up")
		}
	})

	t.Run("partial final deal", func(t *testing.T) {
		e := setupPlayingEasthaven()
		// Leave only 3 stock cards.
		e.SetStock([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, true),
			domain.NewCard(domain.CardDesignHeart, 3, true),
			domain.NewCard(domain.CardDesignClover, 4, true),
		})
		require.NoError(t, e.Deal())
		assert.Equal(t, 0, e.GetStockCount())
		// Only the first 3 columns grew.
		tab := e.GetTableau()
		for i := range domain.EasthavenTableauCnt {
			expected := domain.EasthavenInitialColCards
			if i < 3 {
				expected++
			}
			assert.Equal(t, expected, len(tab[i]), "col %d", i)
		}
	})

	t.Run("cannot deal with empty column", func(t *testing.T) {
		e := setupPlayingEasthaven()
		tab := e.GetTableau()
		tab[0] = nil
		e.SetTableau(tab)
		assert.Error(t, e.Deal())
	})

	t.Run("cannot deal with empty stock", func(t *testing.T) {
		e := setupPlayingEasthaven()
		e.SetStock(nil)
		assert.Error(t, e.Deal())
	})

	t.Run("cannot deal when not playing", func(t *testing.T) {
		e := setupPlayingEasthaven()
		e.SetPhase(domain.EasthavenPhaseGameOver)
		assert.Error(t, e.Deal())
	})
}

func TestEasthaven_MoveTableauToTableau(t *testing.T) {
	t.Run("valid alternating-color descending single card", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
		e.SetTableau(tab)

		require.NoError(t, e.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 0, len(e.GetTableau()[0]))
		assert.Equal(t, 2, len(e.GetTableau()[1]))
	})

	t.Run("valid multi-card sequence move", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		// Heart6(red), Spade5(black) is a valid alt-color descending run.
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		// Spade7(black) accepts Heart6(red).
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		e.SetTableau(tab)

		require.NoError(t, e.MoveTableauToTableau(0, 0, 1))
		assert.Equal(t, 0, len(e.GetTableau()[0]))
		assert.Equal(t, 3, len(e.GetTableau()[1]))
	})

	t.Run("invalid non-sequence group is rejected", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		// Same color = not a valid sequence.
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
			makeTableauCard(domain.CardDesignClover, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 7, true)}
		e.SetTableau(tab)
		assert.Error(t, e.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("empty column accepts only a King", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 13, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 5, true)}
		// tab[2] empty
		e.SetTableau(tab)

		require.NoError(t, e.MoveTableauToTableau(0, 0, 2), "King moves to empty column")
		// Non-King to empty column fails.
		assert.Error(t, e.MoveTableauToTableau(1, 0, 0))
	})

	t.Run("error cases", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, false)}
		e.SetTableau(tab)
		assert.Error(t, e.MoveTableauToTableau(-1, 0, 1), "invalid from col")
		assert.Error(t, e.MoveTableauToTableau(0, 0, 9), "invalid to col")
		assert.Error(t, e.MoveTableauToTableau(0, 0, 0), "same col")
		assert.Error(t, e.MoveTableauToTableau(0, 5, 1), "bad index")
		assert.Error(t, e.MoveTableauToTableau(0, 0, 1), "face-down card")

		e.SetPhase(domain.EasthavenPhaseGameOver)
		assert.Error(t, e.MoveTableauToTableau(0, 0, 1), "not playing")
	})
}

func TestEasthaven_MoveTableauToFoundation(t *testing.T) {
	t.Run("ace then build up same suit", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 2, true),
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		e.SetTableau(tab)

		require.NoError(t, e.MoveTableauToFoundation(0)) // Ace
		require.NoError(t, e.MoveTableauToFoundation(0)) // 2
		assert.Equal(t, 2, len(e.GetFoundation()[0]))
	})

	t.Run("error cases", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
		e.SetTableau(tab)
		assert.Error(t, e.MoveTableauToFoundation(-1), "invalid col")
		assert.Error(t, e.MoveTableauToFoundation(1), "empty col")
		assert.Error(t, e.MoveTableauToFoundation(0), "5 cannot start foundation")

		e.SetPhase(domain.EasthavenPhaseGameOver)
		assert.Error(t, e.MoveTableauToFoundation(0))
	})

	t.Run("completing the last King clears the game", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		e.SetStock(nil)
		// Three full foundations, one needing only its King.
		var fnd [domain.EasthavenFoundationCnt][]*domain.Card
		fnd[0] = fullFoundationPile(domain.CardDesignSpade, 12) // needs K
		fnd[1] = fullFoundationPile(domain.CardDesignClover, 13)
		fnd[2] = fullFoundationPile(domain.CardDesignHeart, 13)
		fnd[3] = fullFoundationPile(domain.CardDesignDiamond, 13)
		e.SetFoundation(fnd)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 13, true)}
		e.SetTableau(tab)

		require.NoError(t, e.MoveTableauToFoundation(0))
		assert.Equal(t, domain.EasthavenPhaseGameClear, e.GetPhase())
		assert.True(t, e.GetGameEndFlag())
	})
}

func TestEasthaven_GetHint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		e := setupPlayingEasthaven()
		e.SetPhase(domain.EasthavenPhaseGameOver)
		assert.Nil(t, e.GetHint())
	})

	t.Run("tableau to foundation has priority", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		e.SetStock(nil)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 1, true)}
		e.SetTableau(tab)
		hint := e.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
	})

	t.Run("tableau to tableau move", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		e.SetStock(nil)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
		e.SetTableau(tab)
		hint := e.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
	})
}

func TestEasthaven_AutoComplete(t *testing.T) {
	t.Run("moves all tableau cards to foundation", func(t *testing.T) {
		e := setupPlayingEasthaven()
		clearEasthavenTableau(e)
		e.SetStock(nil)
		var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 2, true),
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		e.SetTableau(tab)
		require.NoError(t, e.AutoComplete())
		assert.Equal(t, 2, len(e.GetFoundation()[0]))
	})

	t.Run("error when stock not empty (not all face up)", func(t *testing.T) {
		e := setupPlayingEasthaven()
		assert.Error(t, e.AutoComplete())
	})

	t.Run("error when not playing", func(t *testing.T) {
		e := setupPlayingEasthaven()
		e.SetPhase(domain.EasthavenPhaseGameClear)
		assert.Error(t, e.AutoComplete())
	})
}

func TestEasthaven_AllFaceUp(t *testing.T) {
	e := setupPlayingEasthaven()
	assert.False(t, e.AllFaceUp(), "stock present")
	e.SetStock(nil)
	clearEasthavenTableau(e)
	var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 1, false)}
	e.SetTableau(tab)
	assert.False(t, e.AllFaceUp(), "face-down card present")
	tab[0][0].FaceUp = true
	e.SetTableau(tab)
	assert.True(t, e.AllFaceUp())
}

func TestEasthaven_GiveUpAndUndo(t *testing.T) {
	e := setupPlayingEasthaven()
	assert.False(t, e.CanUndo())

	require.NoError(t, e.Deal())
	assert.True(t, e.CanUndo())
	require.NoError(t, e.Undo())
	assert.Equal(t, 0, e.GetMoveCount())
	assert.Error(t, e.Undo(), "no history left")

	require.NoError(t, e.Deal())
	require.NoError(t, e.Deal())
	require.NoError(t, e.UndoN(2))
	assert.Equal(t, 0, e.GetMoveCount())

	e.GiveUp()
	assert.Equal(t, domain.EasthavenPhaseGameOver, e.GetPhase())
	assert.False(t, e.CanUndo(), "cannot undo once ended")
	assert.Error(t, e.Undo())
}

func TestEasthaven_UndoToEscape(t *testing.T) {
	e := setupPlayingEasthaven()
	assert.Equal(t, 0, e.UndoToEscape(), "not stalemate")
	e.SetIsStalemate(true)
	assert.Equal(t, -1, e.UndoToEscape(), "no history to escape to")
}

func TestEasthaven_Stalemate(t *testing.T) {
	e := setupPlayingEasthaven()
	clearEasthavenTableau(e)
	e.SetStock(nil)
	// After playing the Ace to its foundation, the lone Diamond 9 has no legal
	// destination (no King-empty move, no foundation slot, no stock) → stalemate.
	var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{
		makeTableauCard(domain.CardDesignDiamond, 9, true),
		makeTableauCard(domain.CardDesignSpade, 1, true),
	}
	e.SetTableau(tab)

	require.NoError(t, e.MoveTableauToFoundation(0))
	assert.True(t, e.IsStalemate())
	assert.Nil(t, e.GetHint())
}

func TestEasthaven_NilCardSafety(t *testing.T) {
	// Game state can be restored from untrusted KV JSON, so tableau slices may
	// contain nil entries or cards with a nil Card. No public method must panic.
	e := setupPlayingEasthaven()
	clearEasthavenTableau(e)
	e.SetStock(nil)
	var tab [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{nil}
	tab[1] = []*domain.KlondikeTableauCard{{Card: nil, FaceUp: true}}
	tab[2] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
	e.SetFoundation([domain.EasthavenFoundationCnt][]*domain.Card{{nil}, nil, nil, nil})
	e.SetTableau(tab)

	assert.Error(t, e.MoveTableauToTableau(0, 0, 2), "nil entry rejected")
	assert.Error(t, e.MoveTableauToTableau(1, 0, 2), "nil Card rejected")
	assert.Error(t, e.MoveTableauToFoundation(0), "nil entry rejected")
	assert.Error(t, e.MoveTableauToFoundation(1), "nil Card rejected")
	assert.False(t, e.AllFaceUp(), "nil entry is not face up")
	assert.NotPanics(t, func() { _ = e.GetHint() })
	assert.NotPanics(t, func() { _ = e.AutoComplete() })
}

func TestEasthaven_JSONRoundTrip(t *testing.T) {
	e := setupPlayingEasthaven()
	require.NoError(t, e.Deal())

	data, err := json.Marshal(e)
	require.NoError(t, err)

	var restored domain.Easthaven
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, e.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, e.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, e.GetPhase(), restored.GetPhase())
}
