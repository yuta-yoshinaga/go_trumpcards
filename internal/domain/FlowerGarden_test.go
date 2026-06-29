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

func newTestFlowerGarden() *domain.FlowerGarden {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	return domain.NewFlowerGarden(tc)
}

func setupPlayingFlowerGarden() *domain.FlowerGarden {
	fg := newTestFlowerGarden()
	fg.Reset()
	return fg
}

func makeFGCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeFGTableauCard(design, value int) *domain.FlowerGardenTableauCard {
	return &domain.FlowerGardenTableauCard{Card: makeFGCard(design, value), FaceUp: true}
}

func clearFGTableau(fg *domain.FlowerGarden) {
	var empty [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	fg.SetTableau(empty)
}

func TestNewFlowerGarden(t *testing.T) {
	fg := newTestFlowerGarden()
	assert.NotNil(t, fg)
	assert.Equal(t, domain.FlowerGardenPhase(0), fg.GetPhase())
}

func TestFlowerGarden_Reset(t *testing.T) {
	fg := setupPlayingFlowerGarden()

	assert.Equal(t, domain.FlowerGardenPhasePlaying, fg.GetPhase())
	assert.Equal(t, 0, fg.GetMoveCount())

	// Foundations: empty at the start (player must move Aces out themselves).
	foundation := fg.GetFoundation()
	for i := range domain.FlowerGardenFoundationCnt {
		assert.Equal(t, 0, len(foundation[i]), "foundation %d must start empty", i)
	}

	// Tableau: 6 flower-bed fans of 6 cards each -> 36 cards, all face-up.
	tableau := fg.GetTableau()
	totalTableauCards := 0
	for i := range domain.FlowerGardenTableauCnt {
		assert.Equal(t, domain.FlowerGardenColumnLen, len(tableau[i]),
			"fan %d should have %d cards", i, domain.FlowerGardenColumnLen)
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "all cards should be face up")
		}
		totalTableauCards += len(tableau[i])
	}
	assert.Equal(t, 36, totalTableauCards)

	// Reserve: 16 single face-up cards. Tableau (36) + reserve (16) = 52.
	reserve := fg.GetReserve()
	assert.Equal(t, domain.FlowerGardenReserveCnt, len(reserve))
	for i := range reserve {
		assert.NotNil(t, reserve[i], "reserve %d should be dealt a card", i)
	}
	assert.Equal(t, domain.CardCnt, totalTableauCards+len(reserve))
}

func TestFlowerGarden_MoveTableauToTableau(t *testing.T) {
	t.Run("valid descending move (suit ignored, different suit)", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		// heart 4 onto spade 5 — valid (one lower, suit ignored).
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
		fg.SetTableau(tableau)

		err := fg.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(fg.GetTableau()[0]))
		assert.Equal(t, 2, len(fg.GetTableau()[1]))
	})

	t.Run("valid descending move (suit ignored, same suit)", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		// spade 4 onto spade 5 — valid because suit is IGNORED in Flower Garden.
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 4)}
		tableau[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
		fg.SetTableau(tableau)

		err := fg.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(fg.GetTableau()[1]))
	})

	t.Run("valid descending move (same color clubs onto spades)", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		// clover 4 onto spade 5 — valid (one lower, suit/color ignored).
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignClover, 4)}
		tableau[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
		fg.SetTableau(tableau)

		err := fg.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(fg.GetTableau()[1]))
	})

	t.Run("reject wrong rank", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignHeart, 3)}
		tableau[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
		fg.SetTableau(tableau)

		err := fg.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("empty flower-bed accepts any card", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 7)}
		// fan 1 empty
		fg.SetTableau(tableau)

		err := fg.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(fg.GetTableau()[1]))
	})

	t.Run("only bottom card movable", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{
			makeFGTableauCard(domain.CardDesignSpade, 5),
			makeFGTableauCard(domain.CardDesignHeart, 4),
		}
		fg.SetTableau(tableau)

		err := fg.MoveTableauToTableau(0, 0, 1) // index 0 is not the bottom card
		assert.Error(t, err)
	})

	t.Run("invalid columns and indices", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		assert.Error(t, fg.MoveTableauToTableau(-1, -1, 1))
		assert.Error(t, fg.MoveTableauToTableau(0, -1, 99))
		assert.Error(t, fg.MoveTableauToTableau(0, -1, 0))
	})

	t.Run("not playing phase", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		fg.SetPhase(domain.FlowerGardenPhaseGameOver)
		assert.Error(t, fg.MoveTableauToTableau(0, -1, 1))
	})
}

func TestFlowerGarden_MoveTableauToFoundation(t *testing.T) {
	t.Run("ace then two", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 1)}
		tableau[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 2)}
		fg.SetTableau(tableau)

		require.NoError(t, fg.MoveTableauToFoundation(0))
		require.NoError(t, fg.MoveTableauToFoundation(1))
	})

	t.Run("reject non-ace on empty foundation", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
		fg.SetTableau(tableau)

		assert.Error(t, fg.MoveTableauToFoundation(0))
	})

	t.Run("empty column", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		assert.Error(t, fg.MoveTableauToFoundation(0))
	})

	t.Run("invalid column", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		assert.Error(t, fg.MoveTableauToFoundation(99))
	})
}

func TestFlowerGarden_MoveReserveToTableau(t *testing.T) {
	t.Run("valid move depletes reserve cell one-way", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
		fg.SetTableau(tableau)
		fg.SetReserve([]*domain.Card{makeFGCard(domain.CardDesignHeart, 4)})

		require.NoError(t, fg.MoveReserveToTableau(0, 0))
		assert.Equal(t, 2, len(fg.GetTableau()[0]))
		// The reserve cell is now empty (one-way depletion) — nothing fills it.
		assert.Nil(t, fg.GetReserve()[0])
	})

	t.Run("reject invalid placement", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
		fg.SetTableau(tableau)
		fg.SetReserve([]*domain.Card{makeFGCard(domain.CardDesignClover, 6)}) // wrong rank

		assert.Error(t, fg.MoveReserveToTableau(0, 0))
	})

	t.Run("empty reserve cell rejected", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		fg.SetReserve([]*domain.Card{nil})
		assert.Error(t, fg.MoveReserveToTableau(0, 0))
	})

	t.Run("invalid indices", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		fg.SetReserve([]*domain.Card{makeFGCard(domain.CardDesignHeart, 4)})
		assert.Error(t, fg.MoveReserveToTableau(-1, 0))
		assert.Error(t, fg.MoveReserveToTableau(0, 99))
	})
}

func TestFlowerGarden_MoveReserveToFoundation(t *testing.T) {
	t.Run("ace from reserve", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		fg.SetReserve([]*domain.Card{makeFGCard(domain.CardDesignDiamond, 1)})

		require.NoError(t, fg.MoveReserveToFoundation(0))
		assert.Nil(t, fg.GetReserve()[0])
	})

	t.Run("reject non-ace", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		fg.SetReserve([]*domain.Card{makeFGCard(domain.CardDesignDiamond, 5)})
		assert.Error(t, fg.MoveReserveToFoundation(0))
	})

	t.Run("empty reserve cell", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		fg.SetReserve([]*domain.Card{nil})
		assert.Error(t, fg.MoveReserveToFoundation(0))
	})

	t.Run("invalid index", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		assert.Error(t, fg.MoveReserveToFoundation(99))
	})
}

func TestFlowerGarden_GiveUp(t *testing.T) {
	fg := setupPlayingFlowerGarden()
	fg.GiveUp()
	assert.Equal(t, domain.FlowerGardenPhaseGameOver, fg.GetPhase())
	assert.True(t, fg.GetGameEndFlag())
}

func TestFlowerGarden_Hint(t *testing.T) {
	t.Run("tableau to foundation has priority", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 1)}
		fg.SetTableau(tableau)
		fg.SetReserve(nil)

		hint := fg.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("reserve to foundation", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		fg.SetReserve([]*domain.Card{makeFGCard(domain.CardDesignSpade, 1)})

		hint := fg.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "reserve", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("tableau to tableau", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		clearFGTableau(fg)
		var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
		tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
		fg.SetTableau(tableau)
		fg.SetReserve(nil)

		hint := fg.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("nil when not playing", func(t *testing.T) {
		fg := setupPlayingFlowerGarden()
		fg.SetPhase(domain.FlowerGardenPhaseGameOver)
		assert.Nil(t, fg.GetHint())
	})
}

func TestFlowerGarden_AutoComplete(t *testing.T) {
	fg := setupPlayingFlowerGarden()
	clearFGTableau(fg)
	var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 2)}
	fg.SetTableau(tableau)
	fg.SetReserve([]*domain.Card{makeFGCard(domain.CardDesignSpade, 1)})

	require.NoError(t, fg.AutoComplete())
	// Ace from reserve then 2 from tableau should both land on the spade pile.
	foundation := fg.GetFoundation()
	total := 0
	for i := range domain.FlowerGardenFoundationCnt {
		total += len(foundation[i])
	}
	assert.Equal(t, 2, total)
}

func TestFlowerGarden_AutoCompleteNotPlaying(t *testing.T) {
	fg := setupPlayingFlowerGarden()
	fg.SetPhase(domain.FlowerGardenPhaseGameOver)
	assert.Error(t, fg.AutoComplete())
}

func TestFlowerGarden_Win(t *testing.T) {
	fg := setupPlayingFlowerGarden()
	clearFGTableau(fg)
	fg.SetReserve(nil)
	// Seed all four foundations up to King, leaving the final King of spades
	// on a tableau column to play.
	var foundation [domain.FlowerGardenFoundationCnt][]*domain.Card
	designs := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	for fi, d := range designs {
		maxVal := domain.CardValueMax
		if fi == 0 {
			maxVal = domain.CardValueMax - 1 // leave spade King to play
		}
		for v := 1; v <= maxVal; v++ {
			foundation[fi] = append(foundation[fi], makeFGCard(d, v))
		}
	}
	fg.SetFoundation(foundation)
	var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, domain.CardValueMax)}
	fg.SetTableau(tableau)

	require.NoError(t, fg.MoveTableauToFoundation(0))
	assert.Equal(t, domain.FlowerGardenPhaseGameClear, fg.GetPhase())
}

func TestFlowerGarden_UndoAndReset(t *testing.T) {
	fg := setupPlayingFlowerGarden()
	clearFGTableau(fg)
	var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignHeart, 4)}
	tableau[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
	fg.SetTableau(tableau)

	assert.False(t, fg.CanUndo())
	require.NoError(t, fg.MoveTableauToTableau(0, 0, 1))
	assert.True(t, fg.CanUndo())
	require.NoError(t, fg.Undo())
	assert.Equal(t, 1, len(fg.GetTableau()[0]))

	// Undo with no history errors.
	assert.Error(t, fg.Undo())
}

func TestFlowerGarden_UndoN(t *testing.T) {
	fg := setupPlayingFlowerGarden()
	clearFGTableau(fg)
	var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	tableau[0] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignHeart, 4)}
	tableau[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
	tableau[2] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 7)}
	fg.SetTableau(tableau)

	require.NoError(t, fg.MoveTableauToTableau(0, 0, 1)) // heart4 onto spade5
	require.NoError(t, fg.MoveTableauToTableau(2, 0, 0)) // spade7 onto empty col0
	require.NoError(t, fg.UndoN(2))
	assert.Equal(t, 1, len(fg.GetTableau()[0]))
	assert.Equal(t, 1, len(fg.GetTableau()[1]))
	assert.Equal(t, 1, len(fg.GetTableau()[2]))
}

func TestFlowerGarden_Stalemate(t *testing.T) {
	fg := setupPlayingFlowerGarden()
	clearFGTableau(fg)
	var tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	// Every fan holds a single King (value 13); no card is one lower than
	// another King, no Ace to play, reserve empty -> stalemate, no hint.
	for c := range domain.FlowerGardenTableauCnt {
		tableau[c] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, domain.CardValueMax)}
	}
	fg.SetTableau(tableau)
	fg.SetReserve(nil)

	assert.Nil(t, fg.GetHint())

	// Set up a board with exactly one legal move that leads to a dead board,
	// then perform it and confirm checkStalemate flagged the result. No fan
	// ends up empty (which would keep the empty-column fallback alive) and
	// there is no Ace to advance.
	clearFGTableau(fg)
	var live [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard
	// col0 bottom is the only movable card (heart4 onto spade5 in col1).
	live[0] = []*domain.FlowerGardenTableauCard{
		makeFGTableauCard(domain.CardDesignSpade, domain.CardValueMax),
		makeFGTableauCard(domain.CardDesignHeart, 4),
	}
	live[1] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, 5)}
	for c := 2; c < domain.FlowerGardenTableauCnt; c++ {
		live[c] = []*domain.FlowerGardenTableauCard{makeFGTableauCard(domain.CardDesignSpade, domain.CardValueMax)}
	}
	fg.SetTableau(live)
	fg.SetReserve(nil)
	require.NoError(t, fg.MoveTableauToTableau(0, 1, 1)) // heart4 (bottom of col0) onto spade5
	assert.True(t, fg.IsStalemate())
}

func TestFlowerGarden_MarshalRoundTrip(t *testing.T) {
	fg := setupPlayingFlowerGarden()
	data, err := json.Marshal(fg)
	require.NoError(t, err)

	var restored domain.FlowerGarden
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, fg.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, len(fg.GetReserve()), len(restored.GetReserve()))
}

func TestFlowerGarden_UnmarshalErrors(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var fg domain.FlowerGarden
		assert.Error(t, json.Unmarshal([]byte("not json"), &fg))
	})

	t.Run("nil-fills empty payload", func(t *testing.T) {
		var fg domain.FlowerGarden
		require.NoError(t, json.Unmarshal([]byte(`{}`), &fg))
		tableau := fg.GetTableau()
		for i := range domain.FlowerGardenTableauCnt {
			assert.NotNil(t, tableau[i])
		}
		assert.NotNil(t, fg.GetReserve())
		foundation := fg.GetFoundation()
		for i := range domain.FlowerGardenFoundationCnt {
			assert.NotNil(t, foundation[i])
		}
	})

	t.Run("rejects a nil tableau card", func(t *testing.T) {
		var fg domain.FlowerGarden
		err := json.Unmarshal([]byte(`{"tb":[[null]]}`), &fg)
		assert.Error(t, err)
	})

	t.Run("rejects a nil foundation card", func(t *testing.T) {
		var fg domain.FlowerGarden
		err := json.Unmarshal([]byte(`{"fd":[[null]]}`), &fg)
		assert.Error(t, err)
	})

	t.Run("allows a nil reserve cell (depleted slot)", func(t *testing.T) {
		var fg domain.FlowerGarden
		// Reserve uses nil to mark a played-out cell, so it stays valid.
		require.NoError(t, json.Unmarshal([]byte(`{"rs":[null]}`), &fg))
	})
}
