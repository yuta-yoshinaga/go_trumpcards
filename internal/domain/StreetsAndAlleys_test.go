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

func newTestStreetsAndAlleys() *domain.StreetsAndAlleys {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	return domain.NewStreetsAndAlleys(tc)
}

func setupPlayingStreetsAndAlleys() *domain.StreetsAndAlleys {
	sa := newTestStreetsAndAlleys()
	sa.Reset()
	return sa
}

func makeSACard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeSATableauCard(design, value int) *domain.StreetsAndAlleysTableauCard {
	return &domain.StreetsAndAlleysTableauCard{Card: makeSACard(design, value), FaceUp: true}
}

func clearSATableau(sa *domain.StreetsAndAlleys) {
	var empty [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
	sa.SetTableau(empty)
}

func TestNewStreetsAndAlleys(t *testing.T) {
	sa := newTestStreetsAndAlleys()
	assert.NotNil(t, sa)
	assert.Equal(t, domain.StreetsAndAlleysPhase(0), sa.GetPhase())
}

func TestStreetsAndAlleys_Reset(t *testing.T) {
	sa := setupPlayingStreetsAndAlleys()

	assert.Equal(t, domain.StreetsAndAlleysPhasePlaying, sa.GetPhase())
	assert.Equal(t, 0, sa.GetMoveCount())

	// Foundations: empty at the start (player must move Aces out themselves).
	foundation := sa.GetFoundation()
	for i := range domain.StreetsAndAlleysFoundationCnt {
		assert.Equal(t, 0, len(foundation[i]), "foundation %d must start empty", i)
	}

	// Tableau: columns 0-3 hold 7 cards, columns 4-7 hold 6, all face-up,
	// summing to all 52 cards (Aces included).
	tableau := sa.GetTableau()
	totalTableauCards := 0
	for i := range domain.StreetsAndAlleysTableauCnt {
		want := domain.StreetsAndAlleysShortColumnLen
		if i < domain.StreetsAndAlleysLongColumnCnt {
			want = domain.StreetsAndAlleysLongColumnLen
		}
		assert.Equal(t, want, len(tableau[i]),
			"column %d should have %d cards", i, want)
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "all cards should be face up")
		}
		totalTableauCards += len(tableau[i])
	}
	assert.Equal(t, domain.CardCnt, totalTableauCards)
}

func TestStreetsAndAlleys_MoveTableauToTableau(t *testing.T) {
	t.Run("valid single card move descending any suit", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 5)}
		sa.SetTableau(tableau)

		err := sa.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(sa.GetTableau()[0]))
		assert.Equal(t, 2, len(sa.GetTableau()[1]))
	})

	t.Run("reject multi-card move", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{
			makeSATableauCard(domain.CardDesignSpade, 6),
			makeSATableauCard(domain.CardDesignSpade, 5),
		}
		tableau[1] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 7)}
		sa.SetTableau(tableau)

		err := sa.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only the top card can be moved")
	})

	t.Run("reject same-rank move", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 5)}
		sa.SetTableau(tableau)

		err := sa.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("allow move to empty column", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 5)}
		sa.SetTableau(tableau)

		// Empty columns accept any card in Streets and Alleys.
		err := sa.MoveTableauToTableau(0, 0, 1)
		require.NoError(t, err)
		assert.Equal(t, 0, len(sa.GetTableau()[0]))
		assert.Equal(t, 1, len(sa.GetTableau()[1]))
	})

	t.Run("same column", func(t *testing.T) {
		sa := setupPlayingStreetsAndAlleys()
		err := sa.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		sa := setupPlayingStreetsAndAlleys()
		err := sa.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = sa.MoveTableauToTableau(0, 0, domain.StreetsAndAlleysTableauCnt)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		sa := setupPlayingStreetsAndAlleys()
		err := sa.MoveTableauToTableau(0, 99, 1)
		assert.Error(t, err)
	})

	t.Run("cardIndex -1 resolves to top card", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 5)}
		sa.SetTableau(tableau)

		require.NoError(t, sa.MoveTableauToTableau(0, -1, 1))
		assert.Equal(t, 0, len(sa.GetTableau()[0]))
	})

	t.Run("cardIndex -1 errors on empty column", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		err := sa.MoveTableauToTableau(0, -1, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.SetPhase(domain.StreetsAndAlleysPhaseGameOver)
		err := sa.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestStreetsAndAlleys_MoveTableauToFoundation(t *testing.T) {
	t.Run("place Ace on empty foundation", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 1)}
		sa.SetTableau(tableau)

		// Foundations start empty in Streets and Alleys, so an Ace is the only
		// card a player may move onto a fresh foundation pile.
		require.NoError(t, sa.MoveTableauToFoundation(0))
		assert.Equal(t, 1, len(sa.GetFoundation()[0]))
	})

	t.Run("place 2 on suit Ace", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		// Seed the Spade Ace onto foundation 0 by hand, then drop the Two.
		var foundation [domain.StreetsAndAlleysFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeSACard(domain.CardDesignSpade, 1)}
		sa.SetFoundation(foundation)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 2)}
		sa.SetTableau(tableau)

		require.NoError(t, sa.MoveTableauToFoundation(0))
		assert.Equal(t, 2, len(sa.GetFoundation()[0]))
	})

	t.Run("cannot place non-Ace on empty foundation", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 5)}
		sa.SetTableau(tableau)

		err := sa.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		sa := setupPlayingStreetsAndAlleys()
		err := sa.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = sa.MoveTableauToFoundation(domain.StreetsAndAlleysTableauCnt)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		err := sa.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.SetPhase(domain.StreetsAndAlleysPhaseGameClear)
		err := sa.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

func TestStreetsAndAlleys_GameClear(t *testing.T) {
	sa := newTestStreetsAndAlleys()
	sa.Reset()
	clearSATableau(sa)

	// Pre-fill foundations with Ace..Queen.
	var foundation [domain.StreetsAndAlleysFoundationCnt][]*domain.Card
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	for i, s := range suits {
		pile := make([]*domain.Card, 0, domain.CardValueMax)
		for v := 1; v < domain.CardValueMax; v++ {
			pile = append(pile, makeSACard(s, v))
		}
		foundation[i] = pile
	}
	sa.SetFoundation(foundation)

	// Place 4 kings on tableau, then drop them all to foundation.
	var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
	for i, s := range suits {
		tableau[i] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(s, domain.CardValueMax)}
	}
	sa.SetTableau(tableau)

	for i := range suits {
		require.NoError(t, sa.MoveTableauToFoundation(i))
	}
	assert.Equal(t, domain.StreetsAndAlleysPhaseGameClear, sa.GetPhase())
}

func TestStreetsAndAlleys_GiveUp(t *testing.T) {
	sa := setupPlayingStreetsAndAlleys()
	sa.GiveUp()
	assert.Equal(t, domain.StreetsAndAlleysPhaseGameOver, sa.GetPhase())
	assert.True(t, sa.GetGameEndFlag())

	// Calling GiveUp again is a no-op.
	sa.GiveUp()
	assert.Equal(t, domain.StreetsAndAlleysPhaseGameOver, sa.GetPhase())
}

func TestStreetsAndAlleys_Hint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.SetPhase(domain.StreetsAndAlleysPhaseGameOver)
		assert.Nil(t, sa.GetHint())
	})

	t.Run("priority foundation move", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		// Foundations start empty, so a lone Ace can always drop to a foundation
		// and is the highest-priority hint.
		var foundation [domain.StreetsAndAlleysFoundationCnt][]*domain.Card
		sa.SetFoundation(foundation)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 1)}
		sa.SetTableau(tableau)

		hint := sa.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
	})

	t.Run("tableau-to-tableau when no foundation move", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		// Wipe foundations so no foundation move is possible.
		var foundation [domain.StreetsAndAlleysFoundationCnt][]*domain.Card
		sa.SetFoundation(foundation)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 5)}
		sa.SetTableau(tableau)

		hint := sa.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("nil when stalemate", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		// All columns hold a 5 — no tableau-to-tableau move is legal (no
		// descending pair) and the empty foundations cannot accept a 5.
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		suits := []int{
			domain.CardDesignSpade, domain.CardDesignSpade,
			domain.CardDesignClover, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignHeart,
			domain.CardDesignDiamond, domain.CardDesignDiamond,
		}
		for i, s := range suits {
			tableau[i] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(s, 5)}
		}
		sa.SetTableau(tableau)
		assert.Nil(t, sa.GetHint())
	})
}

func TestStreetsAndAlleys_AutoComplete(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.SetPhase(domain.StreetsAndAlleysPhaseGameOver)
		err := sa.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("re-evaluates stalemate after partial completion", func(t *testing.T) {
		// Build a board where AutoComplete drops a single Two onto its Ace but
		// leaves the remaining columns dead (only 5s). After the partial drop
		// the new state has no legal moves, so isStalemate must flip to true.
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		sa.SetIsStalemate(false)

		// Foundations: seed Aces by hand (Streets and Alleys deals them onto the tableau).
		suits := []int{
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		var foundation [domain.StreetsAndAlleysFoundationCnt][]*domain.Card
		for i, s := range suits {
			foundation[i] = []*domain.Card{makeSACard(s, 1)}
		}
		sa.SetFoundation(foundation)

		// Column 0 holds [5♣, 2♠] so AutoComplete drops the Two onto foundation
		// 0 but leaves a buried 5 behind. The other 7 columns each hold a
		// single 5. After AutoComplete every column still holds a 5 (no empty
		// column to dump into) and there are no 6s or 3s available, so the
		// position is genuinely stuck.
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{
			makeSATableauCard(domain.CardDesignClover, 5),
			makeSATableauCard(domain.CardDesignSpade, 2),
		}
		fiveSuits := []int{
			domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignDiamond,
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		for i, s := range fiveSuits {
			tableau[i+1] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(s, 5)}
		}
		sa.SetTableau(tableau)

		require.NoError(t, sa.AutoComplete())
		assert.Equal(t, domain.StreetsAndAlleysPhasePlaying, sa.GetPhase())
		assert.True(t, sa.IsStalemate(), "stalemate must be re-evaluated after partial AutoComplete")
	})

	t.Run("clears all to foundation when fully orderable", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var foundation [domain.StreetsAndAlleysFoundationCnt][]*domain.Card
		suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i, s := range suits {
			pile := make([]*domain.Card, 0)
			for v := 1; v < domain.CardValueMax; v++ {
				pile = append(pile, makeSACard(s, v))
			}
			foundation[i] = pile
		}
		sa.SetFoundation(foundation)

		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		for i, s := range suits {
			tableau[i] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(s, domain.CardValueMax)}
		}
		sa.SetTableau(tableau)

		require.NoError(t, sa.AutoComplete())
		assert.Equal(t, domain.StreetsAndAlleysPhaseGameClear, sa.GetPhase())
	})
}

func TestStreetsAndAlleys_Undo(t *testing.T) {
	t.Run("undo restores previous state", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 5)}
		sa.SetTableau(tableau)

		require.NoError(t, sa.MoveTableauToTableau(0, 0, 1))
		assert.True(t, sa.CanUndo())
		require.NoError(t, sa.Undo())
		assert.Equal(t, 1, len(sa.GetTableau()[0]))
		assert.Equal(t, 1, len(sa.GetTableau()[1]))
	})

	t.Run("undo with no history", func(t *testing.T) {
		sa := setupPlayingStreetsAndAlleys()
		// Reset wipes history; undo with no recorded moves must fail.
		err := sa.Undo()
		assert.Error(t, err)
	})

	t.Run("undo when not playing", func(t *testing.T) {
		sa := setupPlayingStreetsAndAlleys()
		sa.GiveUp()
		err := sa.Undo()
		assert.Error(t, err)
	})

	t.Run("undoN", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
		tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 5)}
		tableau[2] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 6)}
		sa.SetTableau(tableau)

		require.NoError(t, sa.MoveTableauToTableau(1, 0, 2))
		require.NoError(t, sa.MoveTableauToTableau(0, 0, 2))

		require.NoError(t, sa.UndoN(2))
		assert.Equal(t, 0, sa.GetMoveCount())
	})
}

func TestStreetsAndAlleys_UndoToEscape(t *testing.T) {
	t.Run("not stalemate", func(t *testing.T) {
		// Deterministic: UndoToEscape returns 0 whenever the game is not in a
		// stalemate. Setting the flag explicitly avoids the rare shuffled deal
		// that is an immediate stalemate (which made this flake on CI).
		sa := setupPlayingStreetsAndAlleys()
		sa.SetIsStalemate(false)
		assert.Equal(t, 0, sa.UndoToEscape())
	})

	t.Run("returns -1 when no escape", func(t *testing.T) {
		sa := newTestStreetsAndAlleys()
		sa.Reset()
		clearSATableau(sa)
		sa.SetIsStalemate(true)
		assert.Equal(t, -1, sa.UndoToEscape())
	})
}

func TestStreetsAndAlleys_JSON(t *testing.T) {
	sa := setupPlayingStreetsAndAlleys()
	data, err := json.Marshal(sa)
	require.NoError(t, err)

	sa2 := newTestStreetsAndAlleys()
	err = json.Unmarshal(data, sa2)
	require.NoError(t, err)

	assert.Equal(t, sa.GetPhase(), sa2.GetPhase())
	assert.Equal(t, sa.GetMoveCount(), sa2.GetMoveCount())
}

func TestStreetsAndAlleys_NewDefault(t *testing.T) {
	sa := domain.NewDefaultStreetsAndAlleys()
	assert.NotNil(t, sa)
	sa.Reset()
	assert.Equal(t, domain.StreetsAndAlleysPhasePlaying, sa.GetPhase())
}

func TestStreetsAndAlleys_ActionLog(t *testing.T) {
	sa := newTestStreetsAndAlleys()
	sa.Reset()
	clearSATableau(sa)
	var tableau [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard
	tableau[0] = []*domain.StreetsAndAlleysTableauCard{makeSATableauCard(domain.CardDesignSpade, 2)}
	sa.SetTableau(tableau)

	require.NoError(t, sa.MoveTableauToFoundation(0))
	log := sa.GetActionLog()
	assert.NotEmpty(t, log)
}
