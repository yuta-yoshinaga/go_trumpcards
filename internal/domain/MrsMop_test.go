//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMrsMop(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	assert.NotNil(t, s)
	// **既定は4スート = 本来の Mrs. Mop。**クローン元の Spider は1スートを既定に
	// するが、それを引き継ぐと最初に開いた盤が Mrs. Mop でなくなる。
	assert.Equal(t, MrsMopDifficulty4Suit, s.GetDifficulty())
}

func TestNewDefaultMrsMopDealsTwoFullDecks(t *testing.T) {
	s := NewDefaultMrsMop()
	s.Reset()
	suits := map[int]int{}
	for _, col := range s.GetTableau() {
		for _, tc := range col {
			suits[tc.Card.GetDesign()]++
		}
	}
	assert.Len(t, suits, 4, "two full decks means all four suits are present")
	for d, n := range suits {
		assert.Equal(t, 26, n, "design %d appears twice per rank", d)
	}
}

func TestMrsMopReset(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	assert.Equal(t, MrsMopPhasePlaying, s.GetPhase())
	assert.Equal(t, 0, s.GetMoveCount())
	assert.Equal(t, 500, s.GetScore())
	assert.Equal(t, 0, s.GetCompletedSuits())
	// **山札は無い。**クローン元の Spider は 50 枚を残す。
	assert.Equal(t, 0, s.GetStockCount())
	assert.Nil(t, s.GetActionLog())
	assert.False(t, s.IsStalemate())

	// **13 列 x 8 枚を配り切り、全部表向き。**Spider は伏せ札を作る。
	assert.Equal(t, 13, MrsMopTableauCnt)
	tableau := s.GetTableau()
	total := 0
	for i := range MrsMopTableauCnt {
		assert.Len(t, tableau[i], MrsMopColSize, "col %d", i)
		total += len(tableau[i])
		for j, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "card %d in col %d must be face up", j, i)
		}
	}
	assert.Equal(t, MrsMopTotalCards, total, "the whole deck is dealt; nothing is held back")
	assert.True(t, s.AllFaceUp(), "no hidden information at any point")
}

func TestMrsMopResetWithConfig(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)

	t.Run("1 suit", func(t *testing.T) {
		s.ResetWithConfig(MrsMopConfig{Difficulty: MrsMopDifficulty1Suit})
		assert.Equal(t, MrsMopDifficulty1Suit, s.GetDifficulty())
		// All cards should be spades
		tableau := s.GetTableau()
		for i := range MrsMopTableauCnt {
			for _, tc := range tableau[i] {
				assert.Equal(t, CardDesignSpade, tc.Card.GetDesign())
			}
		}
	})

	t.Run("2 suits", func(t *testing.T) {
		s.ResetWithConfig(MrsMopConfig{Difficulty: MrsMopDifficulty2Suit})
		assert.Equal(t, MrsMopDifficulty2Suit, s.GetDifficulty())
	})

	t.Run("4 suits", func(t *testing.T) {
		s.ResetWithConfig(MrsMopConfig{Difficulty: MrsMopDifficulty4Suit})
		assert.Equal(t, MrsMopDifficulty4Suit, s.GetDifficulty())
	})

	// **未知の値は4スート = 本来の Mrs. Mop に落ちる。**クローン元は1スートに
	// 落とすが、それを引き継ぐと壊れた入力が「別のゲーム」に化ける。
	t.Run("invalid defaults to the proper 4-suit game", func(t *testing.T) {
		s.ResetWithConfig(MrsMopConfig{Difficulty: 99})
		assert.Equal(t, MrsMopDifficulty4Suit, s.GetDifficulty())
	})
}

func TestMrsMopMoveTableauToTableau(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// Set up a simple move: col 0 has [5♠], col 1 has [6♠]
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	for i := range MrsMopTableauCnt {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5+i, false), FaceUp: true}}
	}
	s.SetTableau(tableau)

	// Move 5♠ from col 0 onto 6♠ at col 1
	err := s.MoveTableauToTableau(0, 0, 1)
	require.NoError(t, err)
	result := s.GetTableau()
	assert.Len(t, result[0], 0)
	assert.Len(t, result[1], 2)
}

func TestMrsMopMoveTableauToTableau_Errors(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	t.Run("not playing", func(t *testing.T) {
		s.SetPhase(MrsMopPhaseGameOver)
		err := s.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		s.SetPhase(MrsMopPhasePlaying)
	})

	t.Run("invalid from col", func(t *testing.T) {
		assert.Error(t, s.MoveTableauToTableau(-1, 0, 1))
		assert.Error(t, s.MoveTableauToTableau(10, 0, 1))
	})

	t.Run("invalid to col", func(t *testing.T) {
		assert.Error(t, s.MoveTableauToTableau(0, 0, -1))
		assert.Error(t, s.MoveTableauToTableau(0, 0, 10))
	})

	t.Run("same col", func(t *testing.T) {
		assert.Error(t, s.MoveTableauToTableau(0, 0, 0))
	})

	t.Run("invalid card index", func(t *testing.T) {
		assert.Error(t, s.MoveTableauToTableau(0, -1, 1))
		assert.Error(t, s.MoveTableauToTableau(0, 100, 1))
	})

	t.Run("face down card", func(t *testing.T) {
		var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
		for i := range MrsMopTableauCnt {
			tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: false}}
		}
		s.SetTableau(tableau)
		assert.Error(t, s.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("not valid sequence", func(t *testing.T) {
		var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
		// Col 0: 5♠, 3♥ (not same suit descending)
		tableau[0] = []*MrsMopTableauCard{
			{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
			{Card: NewCard(CardDesignHeart, 3, false), FaceUp: true},
		}
		tableau[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 6, false), FaceUp: true}}
		for i := 2; i < MrsMopTableauCnt; i++ {
			tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
		}
		s.SetTableau(tableau)
		err := s.MoveTableauToTableau(0, 0, 1) // try to move 5♠,3♥ which is not valid sequence
		assert.Error(t, err)
	})

	t.Run("cannot place on tableau", func(t *testing.T) {
		var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
		tableau[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
		tableau[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 3, false), FaceUp: true}} // 5 cannot go on 3
		for i := 2; i < MrsMopTableauCnt; i++ {
			tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
		}
		s.SetTableau(tableau)
		err := s.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestMrsMopMoveToEmptyColumn(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	tableau[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
	tableau[1] = nil // empty
	for i := 2; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)

	err := s.MoveTableauToTableau(0, 0, 1)
	require.NoError(t, err)
}

func TestMrsMopAutoFlip(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	tableau[0] = []*MrsMopTableauCard{
		{Card: NewCard(CardDesignSpade, 7, false), FaceUp: false},
		{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
	}
	tableau[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 6, false), FaceUp: true}}
	for i := 2; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)

	err := s.MoveTableauToTableau(0, 1, 1)
	require.NoError(t, err)
	// The face-down 7♠ should now be flipped
	result := s.GetTableau()
	assert.True(t, result[0][0].FaceUp)
}

func TestMrsMopCompletedSuit(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// Set up a column with K-A same suit (13 cards)
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	col0 := make([]*MrsMopTableauCard, 0, CardValueMax+1)
	// Add an extra card at bottom so column isn't empty after removal
	col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, 3, false), FaceUp: true})
	for v := CardValueMax; v >= 1; v-- {
		col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, v, false), FaceUp: true})
	}
	tableau[0] = col0
	// Put a card that when moved to col 0 would complete the suit
	// Actually, the suit K-A is already there. Let's move a card to trigger check.
	for i := 1; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)
	s.SetScore(500)

	// The K-A sequence is already in col 0. We need a move to trigger the check.
	// Let's set up: col 0 has K through 2, col 1 has A.
	// Moving A from col 1 to col 0 completes the sequence.
	col0 = make([]*MrsMopTableauCard, 0, CardValueMax)
	for v := CardValueMax; v >= 2; v-- {
		col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, v, false), FaceUp: true})
	}
	tableau[0] = col0
	tableau[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}} // A♠

	s.SetTableau(tableau)
	s.SetCompletedSuits(0)

	err := s.MoveTableauToTableau(1, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, s.GetCompletedSuits())
	assert.Equal(t, 599, s.GetScore()) // 500 - 1 (move) + 100 (complete)
	// Col 0 should be empty after removal
	result := s.GetTableau()
	assert.Len(t, result[0], 0)
}

func TestMrsMopGameClear(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// Set completed suits to 7, then complete the 8th
	s.SetCompletedSuits(7)
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	col0 := make([]*MrsMopTableauCard, 0, CardValueMax)
	for v := CardValueMax; v >= 2; v-- {
		col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, v, false), FaceUp: true})
	}
	tableau[0] = col0
	tableau[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	for i := 2; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)

	err := s.MoveTableauToTableau(1, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, MrsMopPhaseGameClear, s.GetPhase())
	assert.Equal(t, 8, s.GetCompletedSuits())
}

func TestMrsMopGiveUp(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()
	s.GiveUp()
	assert.Equal(t, MrsMopPhaseGameOver, s.GetPhase())

	// GiveUp when not playing should be no-op
	s.GiveUp()
	assert.Equal(t, MrsMopPhaseGameOver, s.GetPhase())
}

func TestMrsMopGetHint(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	t.Run("not playing returns nil", func(t *testing.T) {
		s.SetPhase(MrsMopPhaseGameOver)
		assert.Nil(t, s.GetHint())
		s.SetPhase(MrsMopPhasePlaying)
	})

	t.Run("finds hint to expose face-down card", func(t *testing.T) {
		var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
		// Col 0: face-down card + 5♠ (can be moved to col 1's 6♠ to expose face-down)
		tableau[0] = []*MrsMopTableauCard{
			{Card: NewCard(CardDesignSpade, 9, false), FaceUp: false},
			{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
		}
		tableau[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 6, false), FaceUp: true}}
		for i := 2; i < MrsMopTableauCnt; i++ {
			tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
		}
		s.SetTableau(tableau)
		s.SetStock(nil)

		hint := s.GetHint()
		require.NotNil(t, hint)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, 1, hint.CardIndex)
		assert.Equal(t, 1, hint.ToCol)
	})

	t.Run("finds general hint", func(t *testing.T) {
		var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
		// All face-up: col 0 has 3♠, col 1 has 4♠
		tableau[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 3, false), FaceUp: true}}
		tableau[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 4, false), FaceUp: true}}
		for i := 2; i < MrsMopTableauCnt; i++ {
			tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
		}
		s.SetTableau(tableau)
		s.SetStock(nil)

		hint := s.GetHint()
		require.NotNil(t, hint)
	})

	t.Run("no hint available", func(t *testing.T) {
		var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
		// All columns have same value = no moves possible
		for i := range MrsMopTableauCnt {
			tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
		}
		s.SetTableau(tableau)
		s.SetStock(nil)

		hint := s.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("skip empty column to empty column move", func(t *testing.T) {
		var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
		// Col 0 has one card, col 1 is empty, rest have cards with no valid moves
		tableau[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
		tableau[1] = nil // empty
		for i := 2; i < MrsMopTableauCnt; i++ {
			tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
		}
		s.SetTableau(tableau)
		s.SetStock(nil)

		hint := s.GetHint()
		// Moving col 0's only card to empty col 1 is pointless
		assert.Nil(t, hint)
	})
}

func TestMrsMopAutoComplete(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	t.Run("not playing", func(t *testing.T) {
		s.SetPhase(MrsMopPhaseGameOver)
		assert.Error(t, s.AutoComplete())
		s.SetPhase(MrsMopPhasePlaying)
	})

	// **伏せ札のケースは存在しない。**Mrs. Mop は開始時点で全部表向きなので、
	// AutoComplete が「まだ見えていない札がある」で断ることは起きない。
	// クローン元の Spider にはこの状態がある。
	t.Run("everything is face up from the deal, so nothing blocks auto-complete", func(t *testing.T) {
		s.Reset()
		assert.True(t, s.AllFaceUp())
	})

	t.Run("auto complete removes completed suits", func(t *testing.T) {
		var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
		// Col 0: K-A same suit (complete sequence)
		col0 := make([]*MrsMopTableauCard, 0, CardValueMax)
		for v := CardValueMax; v >= 1; v-- {
			col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, v, false), FaceUp: true})
		}
		tableau[0] = col0
		for i := 1; i < MrsMopTableauCnt; i++ {
			tableau[i] = nil
		}
		s.SetTableau(tableau)
		s.SetStock(nil)
		s.SetCompletedSuits(0)
		s.SetScore(500)

		err := s.AutoComplete()
		require.NoError(t, err)
		assert.Equal(t, 1, s.GetCompletedSuits())
	})
}

func TestMrsMopUndo(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	t.Run("no history", func(t *testing.T) {
		assert.Error(t, s.Undo())
		assert.False(t, s.CanUndo())
	})

	t.Run("not playing", func(t *testing.T) {
		s.SetPhase(MrsMopPhaseGameOver)
		assert.Error(t, s.Undo())
		s.SetPhase(MrsMopPhasePlaying)
	})

	// **Deal は存在しない。**履歴を作れるのは移動だけ。配りに依存しない盤を組む。
	t.Run("undo after a move", func(t *testing.T) {
		s.Reset()
		var board [MrsMopTableauCnt][]*MrsMopTableauCard
		board[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax, true), FaceUp: true}}
		board[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax-1, true), FaceUp: true}}
		s.SetTableau(board)
		scoreBefore := s.GetScore()
		require.NoError(t, s.MoveTableauToTableau(1, 0, 0))
		assert.True(t, s.CanUndo())

		require.NoError(t, s.Undo())
		assert.Equal(t, scoreBefore, s.GetScore())
		assert.Len(t, s.GetTableau()[1], 1, "the Q returns to its own column")
		assert.False(t, s.CanUndo())
	})
}

func TestMrsMopAllFaceUp(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// **配った直後から全部表向き。**山札も伏せ札も無いので、開始時点で true。
	assert.True(t, s.AllFaceUp())

	// 壊れた state (伏せ札が混ざる) を復元したときだけ false になる。
	broken := s.GetTableau()
	broken[0][0].FaceUp = false
	s.SetTableau(broken)
	assert.False(t, s.AllFaceUp(), "a face-down card can only come from a corrupt restore")

	// All face up
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	for i := range MrsMopTableauCnt {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	assert.True(t, s.AllFaceUp())
}

func TestMrsMopStalemate(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// Set up stalemate: no moves, no stock
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	for i := range MrsMopTableauCnt {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)

	// Trigger stalemate check via a deal (will fail, but we need to check directly)
	// Actually, stalemate is checked internally. Let's move and see.
	// Can't move same-value cards. No hint. No stock. Should be stalemate.
	// Use MoveTableauToTableau to trigger checkMrsMopStalemate
	err := s.MoveTableauToTableau(0, 0, 1) // This will fail (5 can't go on 5)
	assert.Error(t, err)

	// Set isStalemate directly for test
	s.SetIsStalemate(true)
	assert.True(t, s.IsStalemate())
}

func TestMrsMopSetters(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	s.SetPhase(MrsMopPhaseGameClear)
	assert.Equal(t, MrsMopPhaseGameClear, s.GetPhase())

	s.SetCompletedSuits(5)
	assert.Equal(t, 5, s.GetCompletedSuits())

	s.SetScore(999)
	assert.Equal(t, 999, s.GetScore())

	s.SetIsStalemate(true)
	assert.True(t, s.IsStalemate())

	// **山札は無い。**共有の interface が要求するので残しているが、常に 0。
	s.SetStock([]*Card{NewCard(CardDesignSpade, 1, false)})
	assert.Equal(t, 0, s.GetStockCount(), "Mrs. Mop has no stock at all")
}

func TestMrsMopCheckAndRemoveCompletedSuit_NotEnoughCards(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// Col with less than 13 cards
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	tableau[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
	for i := 1; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)

	// No suit should be removed
	assert.Equal(t, 0, s.GetCompletedSuits())
}

func TestMrsMopCheckAndRemoveCompletedSuit_NotStartingWithK(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// Col with 13 cards but not starting with K
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	col0 := make([]*MrsMopTableauCard, 0, CardValueMax)
	for v := CardValueMax - 1; v >= 0; v-- {
		val := v
		if val == 0 {
			val = 1
		}
		col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, val, false), FaceUp: true})
	}
	tableau[0] = col0
	for i := 1; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)
	s.SetCompletedSuits(0)

	// Move something to trigger check - won't complete because not proper K-A
	assert.Equal(t, 0, s.GetCompletedSuits())
}

func TestMrsMopCheckAndRemoveCompletedSuit_FaceDownInSequence(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// K-A but one card is face down
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	col0 := make([]*MrsMopTableauCard, 0, CardValueMax)
	for v := CardValueMax; v >= 1; v-- {
		faceUp := true
		if v == 7 {
			faceUp = false
		}
		col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, v, false), FaceUp: faceUp})
	}
	tableau[0] = col0
	for i := 1; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)
	s.SetCompletedSuits(0)

	assert.Equal(t, 0, s.GetCompletedSuits())
}

func TestMrsMopCheckAndRemoveCompletedSuit_WrongSuit(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// K-A but mixed suits
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	col0 := make([]*MrsMopTableauCard, 0, CardValueMax)
	for v := CardValueMax; v >= 1; v-- {
		design := CardDesignSpade
		if v == 7 {
			design = CardDesignHeart
		}
		col0 = append(col0, &MrsMopTableauCard{Card: NewCard(design, v, false), FaceUp: true})
	}
	tableau[0] = col0
	for i := 1; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)
	s.SetCompletedSuits(0)

	assert.Equal(t, 0, s.GetCompletedSuits())
}

func TestMrsMopCheckAndRemoveCompletedSuit_WrongValues(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// 13 same-suit cards but not K-A descending (values not sequential)
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	col0 := make([]*MrsMopTableauCard, 0, CardValueMax)
	for v := CardValueMax; v >= 1; v-- {
		val := v
		if v == 5 {
			val = 6 // break sequence
		}
		col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, val, false), FaceUp: true})
	}
	tableau[0] = col0
	for i := 1; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)
	s.SetCompletedSuits(0)

	assert.Equal(t, 0, s.GetCompletedSuits())
}

func TestMrsMopCheckAndRemoveCompletedSuit_NotEndingWithA(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// 13 same-suit cards, starting with K but not ending with A
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	col0 := make([]*MrsMopTableauCard, 0, CardValueMax)
	for v := CardValueMax; v >= 2; v-- {
		col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, v, false), FaceUp: true})
	}
	// Last card is 2 instead of A - but we need 13 cards total
	col0 = append(col0, &MrsMopTableauCard{Card: NewCard(CardDesignSpade, 3, false), FaceUp: true})
	tableau[0] = col0
	for i := 1; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)
	s.SetCompletedSuits(0)

	assert.Equal(t, 0, s.GetCompletedSuits())
}

func TestMrsMopIsValidSequence(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)

	t.Run("single card is valid", func(t *testing.T) {
		cards := []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
		assert.True(t, s.isValidMrsMopSequence(cards))
	})

	t.Run("empty is valid", func(t *testing.T) {
		assert.True(t, s.isValidMrsMopSequence(nil))
	})

	t.Run("face down card in sequence is invalid", func(t *testing.T) {
		cards := []*MrsMopTableauCard{
			{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
			{Card: NewCard(CardDesignSpade, 4, false), FaceUp: false},
		}
		assert.False(t, s.isValidMrsMopSequence(cards))
	})
}

func TestDefaultMrsMopConfig(t *testing.T) {
	cfg := DefaultMrsMopConfig()
	// **既定は4スート = 本来の Mrs. Mop。**クローン元の Spider は1スート既定。
	assert.Equal(t, MrsMopDifficulty4Suit, cfg.Difficulty)
}

func TestMrsMopStalemateCheckNotPlayingPhase(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()
	s.SetPhase(MrsMopPhaseGameClear)
	s.checkMrsMopStalemate()
	// Should not change stalemate when not playing
	assert.False(t, s.IsStalemate())
}

func TestMrsMopHintSequenceStartBeyondFirstFaceUp(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// Set up: face-down, face-up card that breaks sequence, face-up card
	// The longest sequence from end doesn't start at firstFaceUp
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	tableau[0] = []*MrsMopTableauCard{
		{Card: NewCard(CardDesignSpade, 9, false), FaceUp: false},
		{Card: NewCard(CardDesignHeart, 8, false), FaceUp: true}, // different suit breaks sequence
		{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true},
	}
	tableau[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 6, false), FaceUp: true}}
	for i := 2; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 1, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)

	hint := s.GetHint()
	// seqStart=2 (only 5♠), firstFaceUp=1. seqStart > firstFaceUp → skip
	// But seqStart=2 and firstFaceUp=1, so seqStart > firstFaceUp is true → skip this col for "expose" hints
	// Then general hints should find 5♠ → 6♠
	require.NotNil(t, hint)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, 2, hint.CardIndex)
	assert.Equal(t, 1, hint.ToCol)
}

func TestMrsMopHintNoValidSequence(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()

	// Face-down card then face-up cards that don't form a valid sequence from seqStart
	var tableau [MrsMopTableauCnt][]*MrsMopTableauCard
	tableau[0] = []*MrsMopTableauCard{
		{Card: NewCard(CardDesignSpade, 9, false), FaceUp: false},
		{Card: NewCard(CardDesignSpade, 3, false), FaceUp: true},
		{Card: NewCard(CardDesignHeart, 5, false), FaceUp: true}, // breaks sequence
	}
	for i := 1; i < MrsMopTableauCnt; i++ {
		tableau[i] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, 5, false), FaceUp: true}}
	}
	s.SetTableau(tableau)
	s.SetStock(nil)

	// For expose-face-down hints: firstFaceUp=1, seqStart=2 (only 5♥)
	// seqStart > firstFaceUp → skip. But can seqStart=2, seq=[5♥] which is valid single card
	// 5♥ can go on 6♠ in some col? No, all other cols have 5♠.
	// So no expose hint. General hint: col 0 seqStart=2, card=5♥, can't go on any 5♠.
	// Other cols: all have 5♠. seqStart=0. Can they move? 5♠ → ??? None have 6 anywhere.
	// So no hint.
	hint := s.GetHint()
	assert.Nil(t, hint)
}

// --- UndoToEscape / UndoN tests ---

func TestMrsMopUndoToEscape_NotInStalemate(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()
	assert.Equal(t, 0, s.UndoToEscape())
}

func TestMrsMopUndoToEscape_StalemateNoHistory(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()
	s.SetIsStalemate(true)
	assert.Equal(t, -1, s.UndoToEscape())
}

func TestMrsMopUndoToEscape_StalemateWithEscape(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()
	var board [MrsMopTableauCnt][]*MrsMopTableauCard
	board[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax, true), FaceUp: true}}
	board[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax-1, true), FaceUp: true}}
	s.SetTableau(board)
	require.NoError(t, s.MoveTableauToTableau(1, 0, 0))
	s.SetIsStalemate(true)
	n := s.UndoToEscape()
	assert.Equal(t, 1, n)
}

func TestMrsMopUndoN_Zero(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()
	err := s.UndoN(0)
	assert.NoError(t, err)
}

func TestMrsMopUndoN_Valid(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()
	var board [MrsMopTableauCnt][]*MrsMopTableauCard
	// ♠K / ♠Q / ♠J: 2手続けて重ねられる決定的な盤。
	board[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax, true), FaceUp: true}}
	board[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax-1, true), FaceUp: true}}
	board[2] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax-2, true), FaceUp: true}}
	s.SetTableau(board)
	require.NoError(t, s.MoveTableauToTableau(1, 0, 0))
	require.NoError(t, s.MoveTableauToTableau(2, 0, 0))
	require.NoError(t, s.UndoN(2))
	assert.Len(t, s.GetTableau()[0], 1, "both moves are rewound")
}

func TestMrsMopUndoN_Excessive(t *testing.T) {
	tc := NewTrumpCardsWithSuits(MrsMopTotalCards, []int{CardDesignSpade})
	s := NewMrsMop(tc)
	s.Reset()
	var board [MrsMopTableauCnt][]*MrsMopTableauCard
	board[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax, true), FaceUp: true}}
	board[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax-1, true), FaceUp: true}}
	s.SetTableau(board)
	require.NoError(t, s.MoveTableauToTableau(1, 0, 0))
	err := s.UndoN(5)
	assert.Error(t, err, "only one move exists to rewind")
	assert.Contains(t, err.Error(), "undo step")
}
