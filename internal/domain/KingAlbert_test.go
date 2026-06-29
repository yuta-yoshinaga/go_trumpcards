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

func newTestKingAlbert() *domain.KingAlbert {
	tc := domain.NewTrumpCardsWithDecks(1, 0)
	return domain.NewKingAlbert(tc)
}

func setupPlayingKingAlbert() *domain.KingAlbert {
	ka := newTestKingAlbert()
	ka.Reset()
	return ka
}

func makeKACard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeKATableauCard(design, value int) *domain.KingAlbertTableauCard {
	return &domain.KingAlbertTableauCard{Card: makeKACard(design, value), FaceUp: true}
}

func clearKATableau(ka *domain.KingAlbert) {
	var empty [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	ka.SetTableau(empty)
}

func TestNewKingAlbert(t *testing.T) {
	ka := newTestKingAlbert()
	assert.NotNil(t, ka)
	assert.Equal(t, domain.KingAlbertPhase(0), ka.GetPhase())
}

func TestKingAlbert_Reset(t *testing.T) {
	ka := setupPlayingKingAlbert()

	assert.Equal(t, domain.KingAlbertPhasePlaying, ka.GetPhase())
	assert.Equal(t, 0, ka.GetMoveCount())

	// Foundations: empty at the start (player must move Aces out themselves).
	foundation := ka.GetFoundation()
	for i := range domain.KingAlbertFoundationCnt {
		assert.Equal(t, 0, len(foundation[i]), "foundation %d must start empty", i)
	}

	// Tableau: column N (0-indexed) holds N+1 cards -> 1,2,...,9 = 45 cards,
	// all face-up.
	tableau := ka.GetTableau()
	totalTableauCards := 0
	for i := range domain.KingAlbertTableauCnt {
		want := i + 1
		assert.Equal(t, want, len(tableau[i]),
			"column %d should have %d cards", i, want)
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "all cards should be face up")
		}
		totalTableauCards += len(tableau[i])
	}
	assert.Equal(t, 45, totalTableauCards)

	// Reserve: 7 single face-up cards. Tableau (45) + reserve (7) = 52.
	reserve := ka.GetReserve()
	assert.Equal(t, domain.KingAlbertReserveCnt, len(reserve))
	for i := range reserve {
		assert.NotNil(t, reserve[i], "reserve %d should be dealt a card", i)
	}
	assert.Equal(t, domain.CardCnt, totalTableauCards+len(reserve))
}

func TestKingAlbert_MoveTableauToTableau(t *testing.T) {
	t.Run("valid alternating-color descending move", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		// red 4 (heart) onto black 5 (spade) — valid.
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
		ka.SetTableau(tableau)

		err := ka.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(ka.GetTableau()[0]))
		assert.Equal(t, 2, len(ka.GetTableau()[1]))
	})

	t.Run("reject same color", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		// black 4 (clover) onto black 5 (spade) — invalid (same color).
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignClover, 4)}
		tableau[1] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
		ka.SetTableau(tableau)

		err := ka.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("reject wrong rank", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignHeart, 3)}
		tableau[1] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
		ka.SetTableau(tableau)

		err := ka.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("empty column accepts any card", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 7)}
		// column 1 empty
		ka.SetTableau(tableau)

		err := ka.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ka.GetTableau()[1]))
	})

	t.Run("only bottom card movable", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{
			makeKATableauCard(domain.CardDesignSpade, 5),
			makeKATableauCard(domain.CardDesignHeart, 4),
		}
		ka.SetTableau(tableau)

		err := ka.MoveTableauToTableau(0, 0, 1) // index 0 is not the bottom card
		assert.Error(t, err)
	})

	t.Run("invalid columns and indices", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		assert.Error(t, ka.MoveTableauToTableau(-1, -1, 1))
		assert.Error(t, ka.MoveTableauToTableau(0, -1, 99))
		assert.Error(t, ka.MoveTableauToTableau(0, -1, 0))
	})

	t.Run("not playing phase", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		ka.SetPhase(domain.KingAlbertPhaseGameOver)
		assert.Error(t, ka.MoveTableauToTableau(0, -1, 1))
	})
}

func TestKingAlbert_MoveTableauToFoundation(t *testing.T) {
	t.Run("ace then two", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 1)}
		tableau[1] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 2)}
		ka.SetTableau(tableau)

		require.NoError(t, ka.MoveTableauToFoundation(0))
		require.NoError(t, ka.MoveTableauToFoundation(1))
	})

	t.Run("reject non-ace on empty foundation", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
		ka.SetTableau(tableau)

		assert.Error(t, ka.MoveTableauToFoundation(0))
	})

	t.Run("empty column", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		assert.Error(t, ka.MoveTableauToFoundation(0))
	})

	t.Run("invalid column", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		assert.Error(t, ka.MoveTableauToFoundation(99))
	})
}

func TestKingAlbert_MoveReserveToTableau(t *testing.T) {
	t.Run("valid move depletes reserve cell one-way", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
		ka.SetTableau(tableau)
		ka.SetReserve([]*domain.Card{makeKACard(domain.CardDesignHeart, 4)})

		require.NoError(t, ka.MoveReserveToTableau(0, 0))
		assert.Equal(t, 2, len(ka.GetTableau()[0]))
		// The reserve cell is now empty (one-way depletion) — nothing fills it.
		assert.Nil(t, ka.GetReserve()[0])
	})

	t.Run("reject invalid placement", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
		ka.SetTableau(tableau)
		ka.SetReserve([]*domain.Card{makeKACard(domain.CardDesignClover, 4)}) // same color

		assert.Error(t, ka.MoveReserveToTableau(0, 0))
	})

	t.Run("empty reserve cell rejected", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		ka.SetReserve([]*domain.Card{nil})
		assert.Error(t, ka.MoveReserveToTableau(0, 0))
	})

	t.Run("invalid indices", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		ka.SetReserve([]*domain.Card{makeKACard(domain.CardDesignHeart, 4)})
		assert.Error(t, ka.MoveReserveToTableau(-1, 0))
		assert.Error(t, ka.MoveReserveToTableau(0, 99))
	})
}

func TestKingAlbert_MoveReserveToFoundation(t *testing.T) {
	t.Run("ace from reserve", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		ka.SetReserve([]*domain.Card{makeKACard(domain.CardDesignDiamond, 1)})

		require.NoError(t, ka.MoveReserveToFoundation(0))
		assert.Nil(t, ka.GetReserve()[0])
	})

	t.Run("reject non-ace", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		ka.SetReserve([]*domain.Card{makeKACard(domain.CardDesignDiamond, 5)})
		assert.Error(t, ka.MoveReserveToFoundation(0))
	})

	t.Run("empty reserve cell", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		ka.SetReserve([]*domain.Card{nil})
		assert.Error(t, ka.MoveReserveToFoundation(0))
	})

	t.Run("invalid index", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		assert.Error(t, ka.MoveReserveToFoundation(99))
	})
}

func TestKingAlbert_GiveUp(t *testing.T) {
	ka := setupPlayingKingAlbert()
	ka.GiveUp()
	assert.Equal(t, domain.KingAlbertPhaseGameOver, ka.GetPhase())
	assert.True(t, ka.GetGameEndFlag())
}

func TestKingAlbert_Hint(t *testing.T) {
	t.Run("tableau to foundation has priority", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 1)}
		ka.SetTableau(tableau)
		ka.SetReserve(nil)

		hint := ka.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("reserve to foundation", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		ka.SetReserve([]*domain.Card{makeKACard(domain.CardDesignSpade, 1)})

		hint := ka.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "reserve", hint.FromZone)
		assert.Equal(t, "foundation", hint.ToZone)
	})

	t.Run("tableau to tableau", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		clearKATableau(ka)
		var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
		tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignHeart, 4)}
		tableau[1] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
		ka.SetTableau(tableau)
		ka.SetReserve(nil)

		hint := ka.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, "tableau", hint.FromZone)
		assert.Equal(t, "tableau", hint.ToZone)
	})

	t.Run("nil when not playing", func(t *testing.T) {
		ka := setupPlayingKingAlbert()
		ka.SetPhase(domain.KingAlbertPhaseGameOver)
		assert.Nil(t, ka.GetHint())
	})
}

func TestKingAlbert_AutoComplete(t *testing.T) {
	ka := setupPlayingKingAlbert()
	clearKATableau(ka)
	var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 2)}
	ka.SetTableau(tableau)
	ka.SetReserve([]*domain.Card{makeKACard(domain.CardDesignSpade, 1)})

	require.NoError(t, ka.AutoComplete())
	// Ace from reserve then 2 from tableau should both land on the spade pile.
	foundation := ka.GetFoundation()
	total := 0
	for i := range domain.KingAlbertFoundationCnt {
		total += len(foundation[i])
	}
	assert.Equal(t, 2, total)
}

func TestKingAlbert_AutoCompleteNotPlaying(t *testing.T) {
	ka := setupPlayingKingAlbert()
	ka.SetPhase(domain.KingAlbertPhaseGameOver)
	assert.Error(t, ka.AutoComplete())
}

func TestKingAlbert_Win(t *testing.T) {
	ka := setupPlayingKingAlbert()
	clearKATableau(ka)
	ka.SetReserve(nil)
	// Seed all four foundations up to King, leaving the final King of spades
	// on a tableau column to play.
	var foundation [domain.KingAlbertFoundationCnt][]*domain.Card
	designs := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	for fi, d := range designs {
		maxVal := domain.CardValueMax
		if fi == 0 {
			maxVal = domain.CardValueMax - 1 // leave spade King to play
		}
		for v := 1; v <= maxVal; v++ {
			foundation[fi] = append(foundation[fi], makeKACard(d, v))
		}
	}
	ka.SetFoundation(foundation)
	var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, domain.CardValueMax)}
	ka.SetTableau(tableau)

	require.NoError(t, ka.MoveTableauToFoundation(0))
	assert.Equal(t, domain.KingAlbertPhaseGameClear, ka.GetPhase())
}

func TestKingAlbert_UndoAndReset(t *testing.T) {
	ka := setupPlayingKingAlbert()
	clearKATableau(ka)
	var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignHeart, 4)}
	tableau[1] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
	ka.SetTableau(tableau)

	assert.False(t, ka.CanUndo())
	require.NoError(t, ka.MoveTableauToTableau(0, 0, 1))
	assert.True(t, ka.CanUndo())
	require.NoError(t, ka.Undo())
	assert.Equal(t, 1, len(ka.GetTableau()[0]))

	// Undo with no history errors.
	assert.Error(t, ka.Undo())
}

func TestKingAlbert_UndoN(t *testing.T) {
	ka := setupPlayingKingAlbert()
	clearKATableau(ka)
	var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	tableau[0] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignHeart, 4)}
	tableau[1] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
	tableau[2] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 7)}
	ka.SetTableau(tableau)

	// Two moves: heart4 -> col1 (onto spade5), then spade7 stays; instead move
	// the now-bottom heart4 back to an empty col is not needed — do two valid moves.
	require.NoError(t, ka.MoveTableauToTableau(0, 0, 1)) // heart4 onto spade5
	require.NoError(t, ka.MoveTableauToTableau(2, 0, 0)) // spade7 onto empty col0
	require.NoError(t, ka.UndoN(2))
	assert.Equal(t, 1, len(ka.GetTableau()[0]))
	assert.Equal(t, 1, len(ka.GetTableau()[1]))
	assert.Equal(t, 1, len(ka.GetTableau()[2]))
}

func TestKingAlbert_Stalemate(t *testing.T) {
	ka := setupPlayingKingAlbert()
	clearKATableau(ka)
	var tableau [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	// Every column holds a single black 9; no two cards stack (same color),
	// no Ace to play, and the reserve is empty -> stalemate, no hint.
	for c := range domain.KingAlbertTableauCnt {
		tableau[c] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 9)}
	}
	ka.SetTableau(tableau)
	ka.SetReserve(nil)

	assert.Nil(t, ka.GetHint())

	// Set up a board with exactly one legal move that leads to a dead board,
	// then perform it and confirm checkStalemate flagged the result. No column
	// ends up empty (which would otherwise keep the empty-column fallback alive)
	// and there is no Ace to advance.
	clearKATableau(ka)
	var live [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard
	// col0 bottom is the only movable card (heart4 onto spade5 in col1).
	live[0] = []*domain.KingAlbertTableauCard{
		makeKATableauCard(domain.CardDesignSpade, 9),
		makeKATableauCard(domain.CardDesignHeart, 4),
	}
	live[1] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 5)}
	for c := 2; c < domain.KingAlbertTableauCnt; c++ {
		live[c] = []*domain.KingAlbertTableauCard{makeKATableauCard(domain.CardDesignSpade, 9)}
	}
	ka.SetTableau(live)
	ka.SetReserve(nil)
	require.NoError(t, ka.MoveTableauToTableau(0, 1, 1)) // heart4 (bottom of col0) onto spade5
	assert.True(t, ka.IsStalemate())
}

func TestKingAlbert_MarshalRoundTrip(t *testing.T) {
	ka := setupPlayingKingAlbert()
	data, err := json.Marshal(ka)
	require.NoError(t, err)

	var restored domain.KingAlbert
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, ka.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, len(ka.GetReserve()), len(restored.GetReserve()))
}

func TestKingAlbert_UnmarshalErrors(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var ka domain.KingAlbert
		assert.Error(t, json.Unmarshal([]byte("not json"), &ka))
	})

	t.Run("nil-fills empty payload", func(t *testing.T) {
		var ka domain.KingAlbert
		require.NoError(t, json.Unmarshal([]byte(`{}`), &ka))
		tableau := ka.GetTableau()
		for i := range domain.KingAlbertTableauCnt {
			assert.NotNil(t, tableau[i])
		}
		assert.NotNil(t, ka.GetReserve())
		foundation := ka.GetFoundation()
		for i := range domain.KingAlbertFoundationCnt {
			assert.NotNil(t, foundation[i])
		}
	})
}
