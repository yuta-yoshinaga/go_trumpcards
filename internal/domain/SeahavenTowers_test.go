//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestSeahavenTowers() *SeahavenTowers {
	return NewSeahavenTowers(NewTrumpCards(0))
}

func setupPlayingSeahavenTowers() *SeahavenTowers {
	s := newTestSeahavenTowers()
	s.Reset()
	return s
}

func clearTableauST(s *SeahavenTowers) {
	for i := 0; i < SeahavenTowersTableauCnt; i++ {
		s.tableau[i] = nil
	}
}

func clearReservedST(s *SeahavenTowers) {
	for i := 0; i < SeahavenTowersCellCnt; i++ {
		s.freeCells[i] = nil
	}
}

// --- Construction ---

func TestNewDefaultSeahavenTowers(t *testing.T) {
	s := NewDefaultSeahavenTowers()
	assert.NotNil(t, s)
	s.Reset()
	assert.Equal(t, SeahavenTowersPhasePlaying, s.GetPhase())
}

// --- Reset tests ---

func TestSeahavenTowersReset(t *testing.T) {
	s := newTestSeahavenTowers()
	s.Reset()

	assert.Equal(t, SeahavenTowersPhasePlaying, s.GetPhase())
	assert.Equal(t, 0, s.GetMoveCount())

	tableau := s.GetTableau()
	totalTableau := 0
	for i := 0; i < SeahavenTowersTableauCnt; i++ {
		assert.Equal(t, SeahavenTowersTableauPerCol, len(tableau[i]),
			"column %d should hold %d cards", i, SeahavenTowersTableauPerCol)
		totalTableau += len(tableau[i])
	}
	assert.Equal(t, 50, totalTableau)

	cells := s.GetFreeCells()
	occupied := 0
	for i := 0; i < SeahavenTowersCellCnt; i++ {
		if cells[i] != nil {
			occupied++
		}
	}
	assert.Equal(t, SeahavenTowersCellCnt, occupied, "both reserved cells should hold a card")

	foundation := s.GetFoundation()
	for i := 0; i < SeahavenTowersFoundationCnt; i++ {
		assert.Empty(t, foundation[i])
	}
}

func TestSeahavenTowersResetClearsState(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = s.MoveTableauToFoundation(0)
	assert.NotEmpty(t, s.GetActionLog())

	s.Reset()
	assert.Equal(t, 0, s.GetMoveCount())
	assert.Nil(t, s.GetActionLog())
	assert.False(t, s.IsStalemate())
}

// --- MoveTableauToTableau tests ---

func TestSeahavenTowersMoveTableauToTableauSameSuit(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)

	// Same suit descending succeeds.
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	s.tableau[1] = []*Card{makeCard(CardDesignSpade, 12)}

	err := s.MoveTableauToTableau(1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(s.tableau[0]))
	assert.Equal(t, 0, len(s.tableau[1]))
	assert.Equal(t, 1, s.GetMoveCount())
}

func TestSeahavenTowersMoveTableauToTableauDifferentSuitRejected(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)

	// FreeCell allowed Q♥ on K♠ (alternate colors); Seahaven requires same suit.
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	s.tableau[1] = []*Card{makeCard(CardDesignHeart, 12)}

	err := s.MoveTableauToTableau(1, 0, 0)
	assert.Error(t, err)
}

func TestSeahavenTowersMoveTableauToTableauKingToEmpty(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)

	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}

	err := s.MoveTableauToTableau(0, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(s.tableau[0]))
	assert.Equal(t, 1, len(s.tableau[1]))
}

func TestSeahavenTowersMoveTableauToTableauNonKingToEmptyRejected(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)

	// Seahaven Towers empty-column rule: only Kings allowed.
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}

	err := s.MoveTableauToTableau(0, 0, 1)
	assert.Error(t, err)
}

func TestSeahavenTowersMoveTableauToTableauSupermove(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)

	// Two reserved cells empty → maxMovable = 3. Same-suit 3-card sequence onto same-suit-1.
	s.tableau[0] = []*Card{
		makeCard(CardDesignSpade, 10),
		makeCard(CardDesignSpade, 9),
		makeCard(CardDesignSpade, 8),
	}
	s.tableau[1] = []*Card{makeCard(CardDesignSpade, 11)}

	err := s.MoveTableauToTableau(0, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(s.tableau[1]))
	assert.Equal(t, 0, len(s.tableau[0]))
}

func TestSeahavenTowersMoveTableauToTableauTopCardShortcut(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)

	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	s.tableau[1] = []*Card{makeCard(CardDesignSpade, 12)}

	err := s.MoveTableauToTableau(1, -1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(s.tableau[0]))
}

func TestSeahavenTowersMoveTableauToTableauTopShortcutEmpty(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)

	err := s.MoveTableauToTableau(0, -1, 1)
	assert.Error(t, err)
}

func TestSeahavenTowersMoveTableauToTableauErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		s.SetPhase(SeahavenTowersPhaseGameOver)
		assert.Error(t, s.MoveTableauToTableau(0, 0, 1))
	})
	t.Run("invalid from col negative", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToTableau(-1, 0, 1))
	})
	t.Run("invalid from col too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToTableau(SeahavenTowersTableauCnt, 0, 1))
	})
	t.Run("invalid to col negative", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToTableau(0, 0, -1))
	})
	t.Run("invalid to col too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToTableau(0, 0, SeahavenTowersTableauCnt))
	})
	t.Run("same column", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToTableau(0, 0, 0))
	})
	t.Run("card index too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearTableauST(s)
		s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
		assert.Error(t, s.MoveTableauToTableau(0, 5, 1))
	})
	t.Run("card index < -1", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToTableau(0, -2, 1))
	})
	t.Run("invalid sequence (different suits)", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearTableauST(s)
		s.tableau[0] = []*Card{
			makeCard(CardDesignSpade, 10),
			makeCard(CardDesignHeart, 9), // suit mismatch
		}
		s.tableau[1] = []*Card{makeCard(CardDesignSpade, 11)}
		assert.Error(t, s.MoveTableauToTableau(0, 0, 1))
	})
	t.Run("too many cards (over reserve capacity)", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearTableauST(s)
		// Fill both reserved cells → maxMovable = 1.
		s.freeCells[0] = makeCard(CardDesignSpade, 1)
		s.freeCells[1] = makeCard(CardDesignSpade, 2)
		s.tableau[0] = []*Card{
			makeCard(CardDesignSpade, 10),
			makeCard(CardDesignSpade, 9),
		}
		s.tableau[1] = []*Card{makeCard(CardDesignSpade, 11)}
		assert.Error(t, s.MoveTableauToTableau(0, 0, 1))
	})
	t.Run("cannot place on tableau (rank mismatch)", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearTableauST(s)
		s.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
		s.tableau[1] = []*Card{makeCard(CardDesignSpade, 6)} // same rank
		assert.Error(t, s.MoveTableauToTableau(0, 0, 1))
	})
}

// --- MoveTableauToFoundation tests ---

func TestSeahavenTowersMoveTableauToFoundationAce(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	err := s.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(s.foundation[0]))
}

func TestSeahavenTowersMoveTableauToFoundationSequence(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	s.foundation[0] = []*Card{makeCard(CardDesignSpade, 1)}
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 2)}
	err := s.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(s.foundation[0]))
}

func TestSeahavenTowersMoveTableauToFoundationErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		s.SetPhase(SeahavenTowersPhaseGameOver)
		assert.Error(t, s.MoveTableauToFoundation(0))
	})
	t.Run("invalid column negative", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToFoundation(-1))
	})
	t.Run("invalid column too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToFoundation(SeahavenTowersTableauCnt))
	})
	t.Run("empty column", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearTableauST(s)
		assert.Error(t, s.MoveTableauToFoundation(0))
	})
	t.Run("invalid card for foundation (joker)", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearTableauST(s)
		s.tableau[0] = []*Card{makeCard(CardDesignJoker, 0)}
		assert.Error(t, s.MoveTableauToFoundation(0))
	})
	t.Run("cannot place on foundation (rank)", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearTableauST(s)
		s.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
		assert.Error(t, s.MoveTableauToFoundation(0))
	})
}

// --- MoveTableauToFreeCell tests ---

func TestSeahavenTowersMoveTableauToFreeCell(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 5)}
	err := s.MoveTableauToFreeCell(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(s.tableau[0]))
	assert.NotNil(t, s.freeCells[0])
}

func TestSeahavenTowersMoveTableauToFreeCellErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		s.SetPhase(SeahavenTowersPhaseGameOver)
		assert.Error(t, s.MoveTableauToFreeCell(0, 0))
	})
	t.Run("invalid column negative", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToFreeCell(-1, 0))
	})
	t.Run("invalid column too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToFreeCell(SeahavenTowersTableauCnt, 0))
	})
	t.Run("invalid cell negative", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToFreeCell(0, -1))
	})
	t.Run("invalid cell too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveTableauToFreeCell(0, SeahavenTowersCellCnt))
	})
	t.Run("empty column", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearTableauST(s)
		clearReservedST(s)
		assert.Error(t, s.MoveTableauToFreeCell(0, 0))
	})
	t.Run("cell occupied", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		s.freeCells[0] = makeCard(CardDesignSpade, 1)
		assert.Error(t, s.MoveTableauToFreeCell(0, 0))
	})
}

// --- MoveFreeCellToTableau tests ---

func TestSeahavenTowersMoveFreeCellToTableauKingToEmpty(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.freeCells[0] = makeCard(CardDesignSpade, 13)
	err := s.MoveFreeCellToTableau(0, 0)
	assert.NoError(t, err)
	assert.Nil(t, s.freeCells[0])
	assert.Equal(t, 1, len(s.tableau[0]))
}

func TestSeahavenTowersMoveFreeCellToTableauOnSameSuit(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	s.freeCells[0] = makeCard(CardDesignSpade, 5)
	err := s.MoveFreeCellToTableau(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(s.tableau[0]))
}

func TestSeahavenTowersMoveFreeCellToTableauWrongSuitRejected(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	s.freeCells[0] = makeCard(CardDesignHeart, 5) // different suit — rejected
	err := s.MoveFreeCellToTableau(0, 0)
	assert.Error(t, err)
}

func TestSeahavenTowersMoveFreeCellToTableauNonKingToEmptyRejected(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.freeCells[0] = makeCard(CardDesignSpade, 5)
	err := s.MoveFreeCellToTableau(0, 0)
	assert.Error(t, err)
}

func TestSeahavenTowersMoveFreeCellToTableauErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		s.SetPhase(SeahavenTowersPhaseGameOver)
		assert.Error(t, s.MoveFreeCellToTableau(0, 0))
	})
	t.Run("invalid cell negative", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveFreeCellToTableau(-1, 0))
	})
	t.Run("invalid cell too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveFreeCellToTableau(SeahavenTowersCellCnt, 0))
	})
	t.Run("invalid col negative", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveFreeCellToTableau(0, -1))
	})
	t.Run("invalid col too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveFreeCellToTableau(0, SeahavenTowersTableauCnt))
	})
	t.Run("empty cell", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearReservedST(s)
		assert.Error(t, s.MoveFreeCellToTableau(0, 0))
	})
}

// --- MoveFreeCellToFoundation tests ---

func TestSeahavenTowersMoveFreeCellToFoundation(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearReservedST(s)
	s.freeCells[0] = makeCard(CardDesignSpade, 1)
	err := s.MoveFreeCellToFoundation(0)
	assert.NoError(t, err)
	assert.Nil(t, s.freeCells[0])
	assert.Equal(t, 1, len(s.foundation[0]))
}

func TestSeahavenTowersMoveFreeCellToFoundationErrors(t *testing.T) {
	t.Run("not playing", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		s.SetPhase(SeahavenTowersPhaseGameOver)
		assert.Error(t, s.MoveFreeCellToFoundation(0))
	})
	t.Run("invalid cell negative", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveFreeCellToFoundation(-1))
	})
	t.Run("invalid cell too large", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		assert.Error(t, s.MoveFreeCellToFoundation(SeahavenTowersCellCnt))
	})
	t.Run("empty cell", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		clearReservedST(s)
		assert.Error(t, s.MoveFreeCellToFoundation(0))
	})
	t.Run("invalid card for foundation (joker)", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		s.freeCells[0] = makeCard(CardDesignJoker, 0)
		assert.Error(t, s.MoveFreeCellToFoundation(0))
	})
	t.Run("cannot place on foundation", func(t *testing.T) {
		s := setupPlayingSeahavenTowers()
		s.freeCells[0] = makeCard(CardDesignSpade, 5)
		assert.Error(t, s.MoveFreeCellToFoundation(0))
	})
}

// --- GiveUp tests ---

func TestSeahavenTowersGiveUp(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	s.GiveUp()
	assert.Equal(t, SeahavenTowersPhaseGameOver, s.GetPhase())
	assert.Equal(t, 1, len(s.GetActionLog()))
}

func TestSeahavenTowersGiveUpNotPlaying(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	s.SetPhase(SeahavenTowersPhaseGameClear)
	s.GiveUp()
	assert.Equal(t, SeahavenTowersPhaseGameClear, s.GetPhase())
}

func TestSeahavenTowersGetGameEndFlag(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	assert.False(t, s.GetGameEndFlag())
	s.SetPhase(SeahavenTowersPhaseGameOver)
	assert.True(t, s.GetGameEndFlag())
}

// --- GetHint tests ---

func TestSeahavenTowersGetHintNotPlaying(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	s.SetPhase(SeahavenTowersPhaseGameOver)
	assert.Nil(t, s.GetHint())
}

func TestSeahavenTowersGetHintTableauToFoundation(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[2] = []*Card{makeCard(CardDesignHeart, 1)}
	hint := s.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, 2, hint.FromCol)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestSeahavenTowersGetHintFreeCellToFoundation(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.freeCells[1] = makeCard(CardDesignClover, 1)
	hint := s.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "reserved", hint.FromZone)
	assert.Equal(t, 1, hint.FromCol)
	assert.Equal(t, "foundation", hint.ToZone)
}

func TestSeahavenTowersGetHintTableauToTableau(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	s.tableau[1] = []*Card{makeCard(CardDesignSpade, 5)}
	hint := s.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestSeahavenTowersGetHintFreeCellToTableau(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 6)}
	s.freeCells[0] = makeCard(CardDesignSpade, 5)
	hint := s.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "reserved", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestSeahavenTowersGetHintTableauToReserved(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	// One reserved cell occupied, the other empty.
	s.freeCells[0] = makeCard(CardDesignSpade, 5)
	// Single, non-King, no foundation match, no tableau pairing.
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 9)}
	// Fill remaining columns with non-matching cards so the only available move is tableau→reserved.
	for i := 1; i < SeahavenTowersTableauCnt; i++ {
		s.tableau[i] = []*Card{makeCard(CardDesignSpade, 2)}
	}
	hint := s.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.FromZone)
	assert.Equal(t, "reserved", hint.ToZone)
	assert.Equal(t, 1, hint.ToCol)
}

func TestSeahavenTowersGetHintKingToEmpty(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.freeCells[0] = makeCard(CardDesignSpade, 13)
	hint := s.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "reserved", hint.FromZone)
	assert.Equal(t, "tableau", hint.ToZone)
}

func TestSeahavenTowersGetHintNoHint(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	// Reserved cells full with cards that have no legal destination.
	s.freeCells[0] = makeCard(CardDesignSpade, 9)
	s.freeCells[1] = makeCard(CardDesignClover, 9)
	// Fill every column with single non-King cards whose ranks differ by more than 1
	// from any same-suit card on the board — guaranteeing no same-suit-descending stack
	// and no foundation move (no Aces).
	cards := []*Card{
		makeCard(CardDesignSpade, 5), makeCard(CardDesignClover, 5),
		makeCard(CardDesignHeart, 5), makeCard(CardDesignDiamond, 5),
		makeCard(CardDesignSpade, 8), makeCard(CardDesignClover, 8),
		makeCard(CardDesignHeart, 8), makeCard(CardDesignDiamond, 8),
		makeCard(CardDesignHeart, 11), makeCard(CardDesignDiamond, 11),
	}
	for i := 0; i < SeahavenTowersTableauCnt; i++ {
		s.tableau[i] = []*Card{cards[i]}
	}
	hint := s.GetHint()
	assert.Nil(t, hint)
}

// --- AutoComplete tests ---

func TestSeahavenTowersAutoCompleteFinishesGame(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	// 12 cards on each foundation; 4 kings left on tableau / reserved.
	for suit := 0; suit < 4; suit++ {
		for val := 1; val <= 12; val++ {
			s.foundation[suit] = append(s.foundation[suit], makeCard(suit+1, val))
		}
	}
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	s.tableau[1] = []*Card{makeCard(CardDesignClover, 13)}
	s.freeCells[0] = makeCard(CardDesignHeart, 13)
	s.freeCells[1] = makeCard(CardDesignDiamond, 13)

	err := s.AutoComplete()
	assert.NoError(t, err)
	assert.Equal(t, SeahavenTowersPhaseGameClear, s.GetPhase())
}

func TestSeahavenTowersAutoCompletePartial(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	s.tableau[1] = []*Card{makeCard(CardDesignHeart, 5)}

	err := s.AutoComplete()
	assert.NoError(t, err)
	assert.Equal(t, SeahavenTowersPhasePlaying, s.GetPhase())
	assert.Equal(t, 1, len(s.foundation[0]))
}

func TestSeahavenTowersAutoCompleteNotPlaying(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	s.SetPhase(SeahavenTowersPhaseGameOver)
	assert.Error(t, s.AutoComplete())
}

func TestSeahavenTowersAutoCompleteSkipsInvalidCards(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.freeCells[0] = makeCard(CardDesignJoker, 0) // skipped: invalid foundation index
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	s.tableau[1] = []*Card{makeCard(CardDesignJoker, 0)} // skipped

	err := s.AutoComplete()
	assert.NoError(t, err)
	assert.NotNil(t, s.freeCells[0])
	assert.Equal(t, 1, len(s.foundation[0]))
	assert.Equal(t, 1, len(s.tableau[1]))
}

// --- Undo tests ---

func TestSeahavenTowersUndo(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	assert.NoError(t, s.MoveTableauToFoundation(0))
	assert.NoError(t, s.Undo())
	assert.Equal(t, 1, len(s.tableau[0]))
	assert.Equal(t, 0, len(s.foundation[0]))
}

func TestSeahavenTowersUndoNotPlaying(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	s.SetPhase(SeahavenTowersPhaseGameOver)
	assert.Error(t, s.Undo())
}

func TestSeahavenTowersUndoNoHistory(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	assert.Error(t, s.Undo())
}

func TestSeahavenTowersCanUndo(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	assert.False(t, s.CanUndo())

	clearTableauST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = s.MoveTableauToFoundation(0)
	assert.True(t, s.CanUndo())

	s.SetPhase(SeahavenTowersPhaseGameOver)
	assert.False(t, s.CanUndo())
}

func TestSeahavenTowersUndoN(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	s.foundation[0] = []*Card{}
	s.tableau[1] = []*Card{makeCard(CardDesignSpade, 2)}
	_ = s.MoveTableauToFoundation(0)
	_ = s.MoveTableauToFoundation(1)
	assert.NoError(t, s.UndoN(2))
	assert.Equal(t, 0, len(s.foundation[0]))
}

func TestSeahavenTowersUndoNStopsOnError(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	err := s.UndoN(2)
	assert.Error(t, err)
}

func TestSeahavenTowersUndoToEscape(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	assert.Equal(t, 0, s.UndoToEscape())
	s.SetIsStalemate(true)
	assert.Equal(t, -1, s.UndoToEscape())
}

func TestSeahavenTowersUndoToEscapeWithHistory(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 1)}
	_ = s.MoveTableauToFoundation(0)
	s.SetIsStalemate(true) // mark current as stalemate; the snapshot before was non-stalemate.
	assert.Equal(t, 1, s.UndoToEscape())
}

// --- Game clear ---

func TestSeahavenTowersGameClearFromTableau(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	for suit := 0; suit < 4; suit++ {
		for val := 1; val <= 12; val++ {
			s.foundation[suit] = append(s.foundation[suit], makeCard(suit+1, val))
		}
	}
	s.tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	s.tableau[1] = []*Card{makeCard(CardDesignClover, 13)}
	s.tableau[2] = []*Card{makeCard(CardDesignHeart, 13)}
	s.tableau[3] = []*Card{makeCard(CardDesignDiamond, 13)}
	_ = s.MoveTableauToFoundation(0)
	_ = s.MoveTableauToFoundation(1)
	_ = s.MoveTableauToFoundation(2)
	_ = s.MoveTableauToFoundation(3)
	assert.Equal(t, SeahavenTowersPhaseGameClear, s.GetPhase())
}

func TestSeahavenTowersGameClearFromReserved(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	clearTableauST(s)
	clearReservedST(s)
	for suit := 0; suit < 4; suit++ {
		for val := 1; val <= 12; val++ {
			s.foundation[suit] = append(s.foundation[suit], makeCard(suit+1, val))
		}
	}
	s.freeCells[0] = makeCard(CardDesignSpade, 13)
	s.freeCells[1] = makeCard(CardDesignClover, 13)
	// Only 2 reserved cells; remaining 2 kings on tableau.
	s.tableau[0] = []*Card{makeCard(CardDesignHeart, 13)}
	s.tableau[1] = []*Card{makeCard(CardDesignDiamond, 13)}
	_ = s.MoveFreeCellToFoundation(0)
	_ = s.MoveFreeCellToFoundation(1)
	_ = s.MoveTableauToFoundation(0)
	_ = s.MoveTableauToFoundation(1)
	assert.Equal(t, SeahavenTowersPhaseGameClear, s.GetPhase())
}

// --- Setters (test-only) ---

func TestSeahavenTowersSetters(t *testing.T) {
	s := setupPlayingSeahavenTowers()
	var tableau [SeahavenTowersTableauCnt][]*Card
	tableau[0] = []*Card{makeCard(CardDesignSpade, 13)}
	s.SetTableau(tableau)
	assert.Equal(t, 1, len(s.GetTableau()[0]))

	var cells [SeahavenTowersCellCnt]*Card
	cells[0] = makeCard(CardDesignHeart, 12)
	s.SetFreeCells(cells)
	assert.NotNil(t, s.GetFreeCells()[0])

	var found [SeahavenTowersFoundationCnt][]*Card
	found[0] = []*Card{makeCard(CardDesignSpade, 1)}
	s.SetFoundation(found)
	assert.Equal(t, 1, len(s.GetFoundation()[0]))
}

// CanAutoComplete lives in domain, so its own coverage has to come from a
// domain test -- exercising it only through the CUI presenter leaves the rule
// counted as untested. Fixtures are built by hand rather than dealt.
func TestSeahavenTowersCanAutoComplete(t *testing.T) {
	descending := func() [SeahavenTowersTableauCnt][]*Card {
		var tb [SeahavenTowersTableauCnt][]*Card
		for col := range tb {
			tb[col] = []*Card{
				NewCard(CardDesignSpade, 9, false),
				NewCard(CardDesignHeart, 5, false),
			}
		}
		return tb
	}

	t.Run("true once every column descends", func(t *testing.T) {
		s := newTestSeahavenTowers()
		s.Reset()
		s.SetTableau(descending())
		s.SetPhase(SeahavenTowersPhasePlaying)
		assert.True(t, s.CanAutoComplete())
	})

	t.Run("false while a single column ascends", func(t *testing.T) {
		s := newTestSeahavenTowers()
		s.Reset()
		tb := descending()
		tb[3] = []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 9, false),
		}
		s.SetTableau(tb)
		s.SetPhase(SeahavenTowersPhasePlaying)
		assert.False(t, s.CanAutoComplete())
	})

	t.Run("false outside the playing phase even with a finished board", func(t *testing.T) {
		s := newTestSeahavenTowers()
		s.Reset()
		s.SetTableau(descending())
		s.SetPhase(SeahavenTowersPhaseGameClear)
		assert.False(t, s.CanAutoComplete())
	})

	t.Run("a single-card column cannot be out of order", func(t *testing.T) {
		s := newTestSeahavenTowers()
		s.Reset()
		var tb [SeahavenTowersTableauCnt][]*Card
		for col := range tb {
			tb[col] = []*Card{NewCard(CardDesignSpade, 7, false)}
		}
		s.SetTableau(tb)
		s.SetPhase(SeahavenTowersPhasePlaying)
		assert.True(t, s.CanAutoComplete())
	})

	t.Run("a nil slot is skipped rather than treated as out of order", func(t *testing.T) {
		s := newTestSeahavenTowers()
		s.Reset()
		var tb [SeahavenTowersTableauCnt][]*Card
		for col := range tb {
			tb[col] = []*Card{NewCard(CardDesignSpade, 9, false)}
		}
		// A hole in the middle must not read as an ascending pair.
		tb[0] = []*Card{NewCard(CardDesignSpade, 9, false), nil, NewCard(CardDesignHeart, 5, false)}
		s.SetTableau(tb)
		s.SetPhase(SeahavenTowersPhasePlaying)
		assert.True(t, s.CanAutoComplete())
	})
}
