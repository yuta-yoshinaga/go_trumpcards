//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestWasp() *domain.Wasp {
	tc := domain.NewTrumpCards(0)
	return domain.NewWasp(tc)
}

func setupPlayingWasp() *domain.Wasp {
	s := newTestWasp()
	s.Reset()
	return s
}

func clearWaspTableau(s *domain.Wasp) {
	var empty [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
	s.SetTableau(empty)
	s.SetStock(nil)
}

// --- Tests ---

func TestNewWasp(t *testing.T) {
	s := newTestWasp()
	assert.NotNil(t, s)
	assert.Equal(t, domain.WaspPhase(0), s.GetPhase())
}

func TestWasp_Reset(t *testing.T) {
	s := setupPlayingWasp()

	assert.Equal(t, domain.WaspPhasePlaying, s.GetPhase())
	assert.Equal(t, 0, s.GetMoveCount())
	assert.Equal(t, 0, s.GetCompletedSuits())
	assert.Equal(t, domain.WaspStockSize, s.GetStockCount())

	tableau := s.GetTableau()
	totalCards := 0
	for i := range domain.WaspTableauCnt {
		assert.Equal(t, domain.WaspColSize, len(tableau[i]),
			"column %d should have %d cards", i, domain.WaspColSize)
		for j, tc := range tableau[i] {
			if i < domain.WaspFaceDownCols && j < domain.WaspFaceDownPerCol {
				assert.False(t, tc.FaceUp, "col %d card %d should be face down", i, j)
			} else {
				assert.True(t, tc.FaceUp, "col %d card %d should be face up", i, j)
			}
		}
		totalCards += len(tableau[i])
	}
	assert.Equal(t, 49, totalCards)
}

func TestWasp_MoveTableauToTableau(t *testing.T) {
	t.Run("valid move - same suit descending", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(s.GetTableau()[0]))
		assert.Equal(t, 2, len(s.GetTableau()[1]))
		assert.Equal(t, 1, s.GetMoveCount())
	})

	t.Run("valid move - unordered group (Wasp special)", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
			makeTableauCard(domain.CardDesignClover, 10, true), // unordered
			makeTableauCard(domain.CardDesignDiamond, 2, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(s.GetTableau()[0]))
		assert.Equal(t, 4, len(s.GetTableau()[1]))
	})

	t.Run("cardIndex -1 shorthand moves top card", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 6, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, -1, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(s.GetTableau()[0]))
		assert.Equal(t, 2, len(s.GetTableau()[1]))
	})

	t.Run("King to empty column", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, domain.CardValueMax, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(s.GetTableau()[0]))
		assert.Equal(t, 1, len(s.GetTableau()[1]))
	})

	t.Run("non-King to empty column succeeds (Wasp rule)", func(t *testing.T) {
		// Wasp's only difference from Scorpion: an empty column accepts any
		// card, not just a King. A lone 5 must therefore move successfully.
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 7, true),
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(s.GetTableau()[0]))
		assert.Equal(t, 1, len(s.GetTableau()[1]))
	})

	t.Run("face down card cannot be moved", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, false),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
		assert.Equal(t, "card is face down", err.Error())
	})

	t.Run("different suit fails", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("wrong value fails", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 10, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})

	t.Run("auto-flip after move", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignClover, 8, false),
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 1, 1)
		assert.NoError(t, err)
		assert.True(t, s.GetTableau()[0][0].FaceUp)
	})

	t.Run("suit completion removes K-A sequence", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		// Column 0 has a K+Q...+2 of spades, column 1 has the Ace of spades
		// Moving the Ace to column 0 completes the K-to-A run → removal.
		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		col := make([]*domain.KlondikeTableauCard, 0, 12)
		for v := domain.CardValueMax; v >= 2; v-- {
			col = append(col, makeTableauCard(domain.CardDesignSpade, v, true))
		}
		tab[0] = col
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(1, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(s.GetTableau()[0]))
		assert.Equal(t, 1, s.GetCompletedSuits())
	})

	t.Run("game clear after 4 suits", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		col := make([]*domain.KlondikeTableauCard, 0, 12)
		for v := domain.CardValueMax; v >= 2; v-- {
			col = append(col, makeTableauCard(domain.CardDesignSpade, v, true))
		}
		tab[0] = col
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		s.SetTableau(tab)
		s.SetCompletedSuits(3)

		err := s.MoveTableauToTableau(1, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, domain.WaspPhaseGameClear, s.GetPhase())
	})

	t.Run("invalid from column", func(t *testing.T) {
		s := setupPlayingWasp()
		err := s.MoveTableauToTableau(-1, 0, 1)
		assert.Error(t, err)
		err = s.MoveTableauToTableau(domain.WaspTableauCnt, 0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid to column", func(t *testing.T) {
		s := setupPlayingWasp()
		err := s.MoveTableauToTableau(0, 0, -1)
		assert.Error(t, err)
		err = s.MoveTableauToTableau(0, 0, domain.WaspTableauCnt)
		assert.Error(t, err)
	})

	t.Run("same column", func(t *testing.T) {
		s := setupPlayingWasp()
		err := s.MoveTableauToTableau(0, 0, 0)
		assert.Error(t, err)
	})

	t.Run("invalid card index", func(t *testing.T) {
		s := setupPlayingWasp()
		err := s.MoveTableauToTableau(0, -2, 1)
		assert.Error(t, err)
		err = s.MoveTableauToTableau(0, 100, 1)
		assert.Error(t, err)
	})

	t.Run("not playing phase", func(t *testing.T) {
		s := setupPlayingWasp()
		s.SetPhase(domain.WaspPhaseGameOver)
		err := s.MoveTableauToTableau(0, 0, 1)
		assert.Error(t, err)
	})
}

func TestWasp_Deal(t *testing.T) {
	t.Run("deal all 3 stock cards to columns 0-2", func(t *testing.T) {
		s := setupPlayingWasp()
		before0 := len(s.GetTableau()[0])
		before1 := len(s.GetTableau()[1])
		before2 := len(s.GetTableau()[2])
		stockBefore := s.GetStockCount()
		assert.Equal(t, domain.WaspStockSize, stockBefore)

		err := s.Deal()
		assert.NoError(t, err)
		assert.Equal(t, 0, s.GetStockCount())
		assert.Equal(t, before0+1, len(s.GetTableau()[0]))
		assert.Equal(t, before1+1, len(s.GetTableau()[1]))
		assert.Equal(t, before2+1, len(s.GetTableau()[2]))
	})

	t.Run("deal fails when stock empty", func(t *testing.T) {
		s := setupPlayingWasp()
		s.SetStock(nil)
		err := s.Deal()
		assert.Error(t, err)
	})

	t.Run("deal fails with empty column", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 5, true),
		}
		s.SetTableau(tab)
		err := s.Deal()
		assert.Error(t, err)
	})

	t.Run("deal fails when not playing", func(t *testing.T) {
		s := setupPlayingWasp()
		s.SetPhase(domain.WaspPhaseGameOver)
		err := s.Deal()
		assert.Error(t, err)
	})
}

func TestWasp_GiveUp(t *testing.T) {
	t.Run("give up during play", func(t *testing.T) {
		s := setupPlayingWasp()
		s.GiveUp()
		assert.Equal(t, domain.WaspPhaseGameOver, s.GetPhase())
	})

	t.Run("give up when already over", func(t *testing.T) {
		s := setupPlayingWasp()
		s.SetPhase(domain.WaspPhaseGameClear)
		s.GiveUp()
		assert.Equal(t, domain.WaspPhaseGameClear, s.GetPhase())
	})
}

func TestWasp_GetHint(t *testing.T) {
	t.Run("hint reveals face-down card", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignClover, 8, false),
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		hint := s.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, 1, hint.CardIndex)
		assert.Equal(t, 1, hint.ToCol)
		// #6340: **なぜこの手が勧められたのか**をヒント自身が持つこと。
		// ワスプの肝は12枚の裏カードをどれだけ早く開けるかで、
		// 移動先だけ見せても学べない。
		assert.True(t, hint.ExposesFaceDown, "裏カードが開く手")
	})

	// **裏カードが開かない手では立てない。**常に true なら理由にならない。
	t.Run("hint does not claim an exposure when card below is face-up", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 9, true),
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		hint := s.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, 1, hint.CardIndex)
		assert.Equal(t, 1, hint.ToCol)
		assert.False(t, hint.ExposesFaceDown)
	})

	t.Run("hint does not claim an exposure when moving top of column", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		hint := s.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, 0, hint.FromCol)
		assert.Equal(t, 0, hint.CardIndex)
		assert.Equal(t, 1, hint.ToCol)
		assert.False(t, hint.ExposesFaceDown)
	})

	t.Run("hint all face-up", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		hint := s.GetHint()
		assert.NotNil(t, hint)
		assert.False(t, hint.ExposesFaceDown)
	})

	t.Run("hint deal when no moves", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		// Columns cover all 7 so no empty column; no same-suit decreasing pairs.
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 3, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 3, true)}
		tab[2] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignClover, 3, true)}
		tab[3] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignDiamond, 3, true)}
		tab[4] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 7, true)}
		tab[5] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 9, true)}
		tab[6] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignClover, 11, true)}
		s.SetTableau(tab)
		s.SetStock([]*domain.Card{makeCard(domain.CardDesignSpade, 2)})

		hint := s.GetHint()
		assert.NotNil(t, hint)
		assert.Equal(t, -1, hint.FromCol)
		assert.True(t, hint.IsDeal())
		assert.False(t, hint.ExposesFaceDown, "配るヒントでは裏カードは開かない")
	})

	t.Run("no hint when no moves and no stock", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, false)}
		s.SetTableau(tab)

		hint := s.GetHint()
		assert.Nil(t, hint)
	})

	t.Run("no hint when not playing", func(t *testing.T) {
		s := setupPlayingWasp()
		s.SetPhase(domain.WaspPhaseGameOver)
		hint := s.GetHint()
		assert.Nil(t, hint)
	})
}

func TestWasp_AutoComplete(t *testing.T) {
	t.Run("auto-complete removes completed suits", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		col := make([]*domain.KlondikeTableauCard, 0, domain.CardValueMax)
		for v := domain.CardValueMax; v >= 1; v-- {
			col = append(col, makeTableauCard(domain.CardDesignSpade, v, true))
		}
		tab[0] = col
		s.SetTableau(tab)

		err := s.AutoComplete()
		assert.NoError(t, err)
		assert.Equal(t, 1, s.GetCompletedSuits())
		assert.Equal(t, 0, len(s.GetTableau()[0]))
	})

	t.Run("error when not all face-up", func(t *testing.T) {
		s := setupPlayingWasp()
		err := s.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("error when not playing", func(t *testing.T) {
		s := setupPlayingWasp()
		s.SetPhase(domain.WaspPhaseGameOver)
		err := s.AutoComplete()
		assert.Error(t, err)
	})

	t.Run("re-evaluates stalemate after removing a completed suit", func(t *testing.T) {
		// Before: completedSuits=3, stalemate=true, one column ready to complete.
		// AutoComplete removes the K-A ♠ run → completedSuits=4 → game clear.
		// checkWaspStalemate fires but short-circuits because phase != Playing.
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)
		s.SetIsStalemate(true)
		s.SetCompletedSuits(3)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		col := make([]*domain.KlondikeTableauCard, 0, domain.CardValueMax)
		for v := domain.CardValueMax; v >= 1; v-- {
			col = append(col, makeTableauCard(domain.CardDesignSpade, v, true))
		}
		tab[0] = col
		s.SetTableau(tab)

		err := s.AutoComplete()
		assert.NoError(t, err)
		assert.Equal(t, domain.WaspPhaseGameClear, s.GetPhase())
	})
}

func TestWasp_ResetClearsStalemate(t *testing.T) {
	s := setupPlayingWasp()
	s.SetIsStalemate(true)
	s.Reset()
	// Initial deal has moves available so isStalemate should be re-computed to false.
	assert.False(t, s.IsStalemate())
}

func TestWasp_DealRejectsWhenNotPlaying(t *testing.T) {
	s := setupPlayingWasp()
	s.SetPhase(domain.WaspPhaseGameOver)
	err := s.Deal()
	assert.Error(t, err)
}

func TestWaspHint_IsDeal(t *testing.T) {
	t.Run("nil hint", func(t *testing.T) {
		var h *domain.WaspHint
		assert.False(t, h.IsDeal())
	})
	t.Run("normal move", func(t *testing.T) {
		h := &domain.WaspHint{FromCol: 0, CardIndex: 1, ToCol: 2}
		assert.False(t, h.IsDeal())
	})
	t.Run("deal sentinel", func(t *testing.T) {
		h := &domain.WaspHint{
			FromCol:   domain.WaspHintDeal,
			CardIndex: domain.WaspHintDeal,
			ToCol:     domain.WaspHintDeal,
		}
		assert.True(t, h.IsDeal())
	})
}

func TestWasp_AllFaceUp(t *testing.T) {
	t.Run("all face-up with empty stock", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignSpade, 1, true),
		}
		s.SetTableau(tab)
		assert.True(t, s.AllFaceUp())
	})

	t.Run("not all face-up", func(t *testing.T) {
		s := setupPlayingWasp()
		assert.False(t, s.AllFaceUp())
	})

	t.Run("stock not empty means false", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)
		s.SetStock([]*domain.Card{makeCard(domain.CardDesignSpade, 1)})
		assert.False(t, s.AllFaceUp())
	})
}

func TestWasp_Undo(t *testing.T) {
	t.Run("undo a move", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 5, true),
		}
		tab[1] = []*domain.KlondikeTableauCard{
			makeTableauCard(domain.CardDesignHeart, 6, true),
		}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(s.GetTableau()[1]))

		err = s.Undo()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(s.GetTableau()[0]))
		assert.Equal(t, 1, len(s.GetTableau()[1]))
	})

	t.Run("no history", func(t *testing.T) {
		s := setupPlayingWasp()
		err := s.Undo()
		assert.Error(t, err)
	})

	t.Run("not playing", func(t *testing.T) {
		s := setupPlayingWasp()
		s.SetPhase(domain.WaspPhaseGameOver)
		err := s.Undo()
		assert.Error(t, err)
	})
}

func TestWasp_CanUndo(t *testing.T) {
	s := setupPlayingWasp()
	assert.False(t, s.CanUndo())
}

func TestWasp_UndoToEscape(t *testing.T) {
	t.Run("not stalemate", func(t *testing.T) {
		s := setupPlayingWasp()
		assert.Equal(t, 0, s.UndoToEscape())
	})

	t.Run("stalemate with no escape", func(t *testing.T) {
		s := setupPlayingWasp()
		s.SetIsStalemate(true)
		assert.Equal(t, -1, s.UndoToEscape())
	})
}

func TestWasp_UndoN(t *testing.T) {
	t.Run("undo multiple", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 5, true)}
		tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
		tab[2] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 7, true)}
		s.SetTableau(tab)

		err := s.MoveTableauToTableau(0, 0, 1)
		assert.NoError(t, err)
		err = s.MoveTableauToTableau(1, 0, 2)
		assert.NoError(t, err)

		err = s.UndoN(2)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(s.GetTableau()[0]))
		assert.Equal(t, 1, len(s.GetTableau()[1]))
		assert.Equal(t, 1, len(s.GetTableau()[2]))
	})

	t.Run("undo too many fails", func(t *testing.T) {
		s := setupPlayingWasp()
		err := s.UndoN(1)
		assert.Error(t, err)
	})
}

func TestWasp_ActionLog(t *testing.T) {
	s := newTestWasp()
	s.Reset()
	clearWaspTableau(s)

	var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 5, true)}
	tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
	s.SetTableau(tab)

	_ = s.MoveTableauToTableau(0, 0, 1)
	log := s.GetActionLog()
	assert.Equal(t, 1, len(log))
	assert.Equal(t, "move", log[0].ActionType)
}

func TestWasp_MarshalUnmarshalJSON(t *testing.T) {
	s := setupPlayingWasp()

	data, err := json.Marshal(s)
	assert.NoError(t, err)

	s2 := &domain.Wasp{}
	err = json.Unmarshal(data, s2)
	assert.NoError(t, err)

	assert.Equal(t, s.GetPhase(), s2.GetPhase())
	assert.Equal(t, s.GetMoveCount(), s2.GetMoveCount())
	assert.Equal(t, s.GetStockCount(), s2.GetStockCount())

	tab1 := s.GetTableau()
	tab2 := s2.GetTableau()
	for i := range domain.WaspTableauCnt {
		assert.Equal(t, len(tab1[i]), len(tab2[i]), "column %d", i)
	}
}

func TestWasp_UnmarshalJSON_invalid(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		s := &domain.Wasp{}
		err := json.Unmarshal([]byte("invalid"), s)
		assert.Error(t, err)
	})

	t.Run("oversized action log", func(t *testing.T) {
		bigLog := make([]*domain.ActionLogEntry, 1001)
		for i := range bigLog {
			bigLog[i] = &domain.ActionLogEntry{}
		}
		data, _ := json.Marshal(map[string]any{"al": bigLog})
		s := &domain.Wasp{}
		err := json.Unmarshal(data, s)
		assert.Error(t, err)
	})

	t.Run("oversized tableau column", func(t *testing.T) {
		bigCol := make([]*domain.KlondikeTableauCard, 1001)
		for i := range bigCol {
			bigCol[i] = &domain.KlondikeTableauCard{}
		}
		data, _ := json.Marshal(map[string]any{
			"tb": [domain.WaspTableauCnt][]*domain.KlondikeTableauCard{bigCol},
		})
		s := &domain.Wasp{}
		err := json.Unmarshal(data, s)
		assert.Error(t, err)
	})
}

func TestWasp_Stalemate(t *testing.T) {
	t.Run("stalemate when no moves and no stock", func(t *testing.T) {
		s := newTestWasp()
		s.Reset()
		clearWaspTableau(s)

		var tab [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, false)}
		s.SetTableau(tab)

		assert.Nil(t, s.GetHint())
	})
}
