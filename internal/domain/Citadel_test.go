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

func newTestCitadel() *domain.Citadel {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	return domain.NewCitadel(tc)
}

func setupPlayingCitadel() *domain.Citadel {
	c := newTestCitadel()
	c.Reset()
	return c
}

func makeCitadelCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeCitadelTableauCard(design, value int) *domain.CitadelTableauCard {
	return &domain.CitadelTableauCard{Card: makeCitadelCard(design, value), FaceUp: true}
}

func clearCitadelTableau(c *domain.Citadel) {
	var empty [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
	c.SetTableau(empty)
}

func TestNewCitadel(t *testing.T) {
	c := newTestCitadel()
	assert.NotNil(t, c)
	assert.Equal(t, domain.CitadelPhase(0), c.GetPhase())
}

func TestCitadel_NewDefault(t *testing.T) {
	c := domain.NewDefaultCitadel()
	assert.NotNil(t, c)
	c.Reset()
	assert.Equal(t, domain.CitadelPhasePlaying, c.GetPhase())
}

func TestCitadel_Reset_InitialLayout(t *testing.T) {
	c := setupPlayingCitadel()

	// Phase and counters
	assert.True(t, c.GetPhase() == domain.CitadelPhasePlaying || c.GetPhase() == domain.CitadelPhaseGameClear)
	assert.Equal(t, 0, c.GetMoveCount())

	// Foundations: each suit foundation must contain at least its Ace at index 0.
	foundation := c.GetFoundation()
	totalFoundationCards := 0
	for i := range domain.CitadelFoundationCnt {
		require.GreaterOrEqual(t, len(foundation[i]), 1, "foundation %d must hold at least its Ace", i)
		assert.Equal(t, 1, foundation[i][0].GetValue(), "foundation %d base must be an Ace", i)
		// Verify foundation cards are strictly ascending and same suit
		for j := 1; j < len(foundation[i]); j++ {
			assert.Equal(t, foundation[i][0].GetDesign(), foundation[i][j].GetDesign())
			assert.Equal(t, foundation[i][j-1].GetValue()+1, foundation[i][j].GetValue())
		}
		totalFoundationCards += len(foundation[i])
	}

	// Tableau: all cards face-up, no Aces, total cards + foundation cards == 52.
	tableau := c.GetTableau()
	totalTableauCards := 0
	for i := range domain.CitadelTableauCnt {
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "all cards should be face up")
			assert.NotEqual(t, 1, tc.Card.GetValue(), "Aces must be pulled to foundation, not in tableau")
		}
		totalTableauCards += len(tableau[i])
	}

	// Total cards must always equal 52.
	assert.Equal(t, 52, totalFoundationCards+totalTableauCards)
}

func TestCitadel_Reset_AutoMoveToFoundation_NoTwosInTableau(t *testing.T) {
	c := setupPlayingCitadel()

	twoCount := 0
	for _, col := range c.GetTableau() {
		for _, tc := range col {
			if tc.Card.GetValue() == 2 {
				twoCount++
			}
		}
	}
	assert.Equal(t, 0, twoCount, "all Twos should be auto-moved to foundation during deal")
}

func TestCitadel_Reset_AutoMoveToFoundation_MoreThanAces(t *testing.T) {
	c := setupPlayingCitadel()

	totalFoundationCards := 0
	for _, pile := range c.GetFoundation() {
		totalFoundationCards += len(pile)
	}
	assert.Greater(t, totalFoundationCards, domain.CitadelFoundationCnt, "foundation must contain more than just Aces after auto-move")
}

func TestCitadel_MoveTableauToTableau(t *testing.T) {
	t.Run("valid single card move descending any suit", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)

		err := c.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(c.GetTableau()[0]))
		assert.Equal(t, 2, len(c.GetTableau()[1]))
	})

	t.Run("reject multi-card move", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{
			makeCitadelTableauCard(domain.CardDesignSpade, 6),
			makeCitadelTableauCard(domain.CardDesignSpade, 5),
		}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 7)}
		c.SetTableau(tableau)

		err := c.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only the top card can be moved")
	})

	t.Run("reject same-rank move", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 5)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)

		err := c.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("reject ascending rank move", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 6)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)

		err := c.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on tableau")
	})

	t.Run("same column", func(t *testing.T) {
		c := setupPlayingCitadel()
		err := c.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		c := setupPlayingCitadel()
		err := c.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = c.MoveTableauToTableau(0, 0, domain.CitadelTableauCnt)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		c := setupPlayingCitadel()
		err := c.MoveTableauToTableau(0, 99, 1)
		assert.Error(t, err)
	})

	t.Run("cardIndex -1 resolves to top card", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)

		require.NoError(t, c.MoveTableauToTableau(0, -1, 1))
		assert.Equal(t, 0, len(c.GetTableau()[0]))
	})

	t.Run("cardIndex -1 errors on empty column", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		err := c.MoveTableauToTableau(0, -1, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		c := newTestCitadel()
		c.SetPhase(domain.CitadelPhaseGameOver)
		err := c.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestCitadel_MoveTableauToTableau_EmptyColumn(t *testing.T) {
	c := newTestCitadel()
	c.Reset()
	clearCitadelTableau(c)
	var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
	tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
	c.SetTableau(tableau)

	// Empty column 1 accepts any card in Citadel.
	err := c.MoveTableauToTableau(0, 0, 1)
	require.NoError(t, err)
	assert.Equal(t, 0, len(c.GetTableau()[0]))
	assert.Equal(t, 1, len(c.GetTableau()[1]))
}

func TestCitadel_MoveTableauToFoundation(t *testing.T) {
	t.Run("place 2 on suit Ace", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var foundation [domain.CitadelFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeCitadelCard(domain.CardDesignSpade, 1)}
		c.SetFoundation(foundation)

		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 2)}
		c.SetTableau(tableau)

		require.NoError(t, c.MoveTableauToFoundation(0))
		assert.Equal(t, 2, len(c.GetFoundation()[0]))
	})

	t.Run("cannot place non-Two on Ace-only foundation", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var foundation [domain.CitadelFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeCitadelCard(domain.CardDesignSpade, 1)}
		c.SetFoundation(foundation)

		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)

		err := c.MoveTableauToFoundation(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot place card on foundation")
	})

	t.Run("cannot place wrong suit on foundation", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var foundation [domain.CitadelFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeCitadelCard(domain.CardDesignSpade, 1)}
		c.SetFoundation(foundation)

		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 2)}
		c.SetTableau(tableau)

		err := c.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("invalid column", func(t *testing.T) {
		c := setupPlayingCitadel()
		err := c.MoveTableauToFoundation(-1)
		assert.Error(t, err)
		err = c.MoveTableauToFoundation(domain.CitadelTableauCnt)
		assert.Error(t, err)
	})

	t.Run("empty column", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		err := c.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		c := newTestCitadel()
		c.SetPhase(domain.CitadelPhaseGameClear)
		err := c.MoveTableauToFoundation(0)
		assert.Error(t, err)
	})
}

func TestCitadel_GameClear(t *testing.T) {
	c := newTestCitadel()
	c.Reset()
	clearCitadelTableau(c)

	// Pre-fill foundations with Ace..Queen (12 cards each).
	var foundation [domain.CitadelFoundationCnt][]*domain.Card
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	for i, s := range suits {
		pile := make([]*domain.Card, 0, domain.CardValueMax)
		for v := 1; v < domain.CardValueMax; v++ {
			pile = append(pile, makeCitadelCard(s, v))
		}
		foundation[i] = pile
	}
	c.SetFoundation(foundation)

	// Place 4 Kings on tableau, then move them all to foundation.
	var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
	for i, s := range suits {
		tableau[i] = []*domain.CitadelTableauCard{makeCitadelTableauCard(s, domain.CardValueMax)}
	}
	c.SetTableau(tableau)

	for i := range suits {
		require.NoError(t, c.MoveTableauToFoundation(i))
	}
	assert.Equal(t, domain.CitadelPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

func TestCitadel_GiveUp(t *testing.T) {
	c := setupPlayingCitadel()
	c.GiveUp()
	assert.Equal(t, domain.CitadelPhaseGameOver, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())

	// Calling GiveUp again is a no-op.
	c.GiveUp()
	assert.Equal(t, domain.CitadelPhaseGameOver, c.GetPhase())
}

func TestCitadel_Hint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		c := newTestCitadel()
		c.SetPhase(domain.CitadelPhaseGameOver)
		assert.Nil(t, c.GetHint())
	})

	t.Run("priority foundation move", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var foundation [domain.CitadelFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{makeCitadelCard(domain.CardDesignSpade, 1)}
		c.SetFoundation(foundation)

		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 2)}
		c.SetTableau(tableau)

		hint := c.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, 0, hint.ToCol)
	})

	t.Run("tableau-to-tableau when no foundation move", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var foundation [domain.CitadelFoundationCnt][]*domain.Card
		c.SetFoundation(foundation)

		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)

		hint := c.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("nil when stalemate", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var foundation [domain.CitadelFoundationCnt][]*domain.Card
		baseSuits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i, s := range baseSuits {
			foundation[i] = []*domain.Card{makeCitadelCard(s, 1)}
		}
		c.SetFoundation(foundation)

		// All columns hold a 5 — no tableau-to-tableau move is legal (no
		// descending pair) and the Ace-only foundations cannot accept a 5.
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		suits := []int{
			domain.CardDesignSpade, domain.CardDesignSpade,
			domain.CardDesignClover, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignHeart,
			domain.CardDesignDiamond, domain.CardDesignDiamond,
		}
		for i, s := range suits {
			tableau[i] = []*domain.CitadelTableauCard{makeCitadelTableauCard(s, 5)}
		}
		c.SetTableau(tableau)
		assert.Nil(t, c.GetHint())
	})
}

func TestCitadel_AutoComplete(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		c := newTestCitadel()
		c.SetPhase(domain.CitadelPhaseGameOver)
		err := c.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("re-evaluates stalemate after partial completion", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		c.SetIsStalemate(false)

		suits := []int{
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		var foundation [domain.CitadelFoundationCnt][]*domain.Card
		for i, s := range suits {
			foundation[i] = []*domain.Card{makeCitadelCard(s, 1)}
		}
		c.SetFoundation(foundation)

		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{
			makeCitadelTableauCard(domain.CardDesignClover, 5),
			makeCitadelTableauCard(domain.CardDesignSpade, 2),
		}
		fiveSuits := []int{
			domain.CardDesignSpade, domain.CardDesignHeart, domain.CardDesignDiamond,
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		}
		for i, s := range fiveSuits {
			tableau[i+1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(s, 5)}
		}
		c.SetTableau(tableau)

		require.NoError(t, c.AutoComplete())
		assert.Equal(t, domain.CitadelPhasePlaying, c.GetPhase())
		assert.True(t, c.IsStalemate(), "stalemate must be re-evaluated after partial AutoComplete")
	})

	t.Run("clears all to foundation when fully orderable", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var foundation [domain.CitadelFoundationCnt][]*domain.Card
		suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		for i, s := range suits {
			pile := make([]*domain.Card, 0)
			for v := 1; v < domain.CardValueMax; v++ {
				pile = append(pile, makeCitadelCard(s, v))
			}
			foundation[i] = pile
		}
		c.SetFoundation(foundation)

		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		for i, s := range suits {
			tableau[i] = []*domain.CitadelTableauCard{makeCitadelTableauCard(s, domain.CardValueMax)}
		}
		c.SetTableau(tableau)

		require.NoError(t, c.AutoComplete())
		assert.Equal(t, domain.CitadelPhaseGameClear, c.GetPhase())
	})
}

func TestCitadel_Undo(t *testing.T) {
	t.Run("undo restores previous state", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)

		assert.False(t, c.CanUndo())
		require.NoError(t, c.MoveTableauToTableau(0, 0, 1))
		assert.True(t, c.CanUndo())
		require.NoError(t, c.Undo())
		assert.False(t, c.CanUndo())
		assert.Equal(t, 1, len(c.GetTableau()[0]))
		assert.Equal(t, 1, len(c.GetTableau()[1]))
	})

	t.Run("undo with no history", func(t *testing.T) {
		c := setupPlayingCitadel()
		err := c.Undo()
		assert.Error(t, err)
	})

	t.Run("undo when not playing", func(t *testing.T) {
		c := setupPlayingCitadel()
		c.GiveUp()
		err := c.Undo()
		assert.Error(t, err)
	})

	t.Run("undoN", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		tableau[2] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 6)}
		c.SetTableau(tableau)

		require.NoError(t, c.MoveTableauToTableau(1, 0, 2))
		require.NoError(t, c.MoveTableauToTableau(0, 0, 2))

		require.NoError(t, c.UndoN(2))
		assert.Equal(t, 0, c.GetMoveCount())
	})
}

func TestCitadel_UndoToEscape(t *testing.T) {
	t.Run("not stalemate", func(t *testing.T) {
		c := setupPlayingCitadel()
		c.SetIsStalemate(false)
		assert.Equal(t, 0, c.UndoToEscape())
	})

	t.Run("returns -1 when no escape", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		c.SetIsStalemate(true)
		assert.Equal(t, -1, c.UndoToEscape())
	})

	t.Run("returns steps to escape when history has non-stalemate state", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)
		c.SetIsStalemate(false)

		require.NoError(t, c.MoveTableauToTableau(0, 0, 1))
		c.SetIsStalemate(true)
		assert.Equal(t, 1, c.UndoToEscape())
	})
}

func TestCitadel_JSON(t *testing.T) {
	t.Run("roundtrip with history", func(t *testing.T) {
		c := newTestCitadel()
		c.Reset()
		clearCitadelTableau(c)
		var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
		tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 5)}
		c.SetTableau(tableau)

		require.NoError(t, c.MoveTableauToTableau(0, 0, 1))
		data, err := json.Marshal(c)
		require.NoError(t, err)

		c2 := newTestCitadel()
		err = json.Unmarshal(data, c2)
		require.NoError(t, err)

		assert.Equal(t, c.GetPhase(), c2.GetPhase())
		assert.Equal(t, c.GetMoveCount(), c2.GetMoveCount())
		assert.True(t, c2.CanUndo())
		require.NoError(t, c2.Undo())
	})

	t.Run("unmarshal invalid JSON errors", func(t *testing.T) {
		c := newTestCitadel()
		err := json.Unmarshal([]byte("invalid json"), c)
		assert.Error(t, err)
	})
}

func TestCitadel_Reset_UnknownSuitOrJoker(t *testing.T) {
	tc := domain.NewTrumpCards(1)
	c := domain.NewCitadel(tc)
	c.Reset()
	assert.Equal(t, domain.CitadelPhasePlaying, c.GetPhase())
}

func TestCitadel_ActionLog(t *testing.T) {
	c := newTestCitadel()
	c.Reset()
	clearCitadelTableau(c)
	var foundation [domain.CitadelFoundationCnt][]*domain.Card
	foundation[0] = []*domain.Card{makeCitadelCard(domain.CardDesignSpade, 1)}
	c.SetFoundation(foundation)

	var tableau [domain.CitadelTableauCnt][]*domain.CitadelTableauCard
	tableau[0] = []*domain.CitadelTableauCard{makeCitadelTableauCard(domain.CardDesignSpade, 2)}
	c.SetTableau(tableau)

	require.NoError(t, c.MoveTableauToFoundation(0))
	log := c.GetActionLog()
	assert.NotEmpty(t, log)
}

func TestCitadel_AllFaceUp(t *testing.T) {
	c := newTestCitadel()
	assert.True(t, c.AllFaceUp())
}
