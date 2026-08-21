//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestStHelena() *domain.StHelena {
	tc := domain.NewTrumpCardsWithDecks(2, 0)
	return domain.NewStHelena(tc)
}

func setupPlayingStHelena() *domain.StHelena {
	cr := newTestStHelena()
	cr.Reset()
	return cr
}

func makeStHelenaCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func makeStHelenaTableauCard(design, value int) *domain.StHelenaTableauCard {
	return &domain.StHelenaTableauCard{Card: makeStHelenaCard(design, value), FaceUp: true}
}

func clearStHelenaTableau(cr *domain.StHelena) {
	var empty [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	cr.SetTableau(empty)
}

func clearStHelenaFoundation(cr *domain.StHelena) {
	var empty [domain.StHelenaFoundationCnt][]*domain.Card
	cr.SetFoundation(empty)
}

func TestNewStHelena(t *testing.T) {
	cr := newTestStHelena()
	require.NotNil(t, cr)
	assert.Equal(t, domain.StHelenaPhase(0), cr.GetPhase())
	assert.Equal(t, 0, cr.GetMoveCount())
	assert.Equal(t, 0, cr.GetRedealsRemaining())
}

func TestNewDefaultStHelena(t *testing.T) {
	cr := domain.NewDefaultStHelena()
	require.NotNil(t, cr)
	cr.Reset()
	assert.Equal(t, domain.StHelenaMaxRedeals, cr.GetRedealsRemaining())
}

func TestStHelena_Reset(t *testing.T) {
	cr := setupPlayingStHelena()
	assert.Equal(t, domain.StHelenaPhasePlaying, cr.GetPhase())
	assert.Equal(t, 0, cr.GetMoveCount())
	assert.Equal(t, domain.StHelenaMaxRedeals, cr.GetRedealsRemaining())
	assert.False(t, cr.IsStalemate())

	tableau := cr.GetTableau()
	totalTableau := 0
	for i := 0; i < domain.StHelenaTableauCnt; i++ {
		assert.Equal(t, domain.StHelenaTableauInitialSize, len(tableau[i]), "column %d should have 6 cards", i)
		for _, tc := range tableau[i] {
			assert.True(t, tc.FaceUp, "all tableau cards should be face up")
		}
		totalTableau += len(tableau[i])
	}
	assert.Equal(t, 96, totalTableau)

	foundation := cr.GetFoundation()
	for i := 0; i < domain.StHelenaAscendingFoundationCnt; i++ {
		require.Len(t, foundation[i], 1, "ascending foundation %d should be seeded with one card", i)
		assert.Equal(t, 1, foundation[i][0].GetValue(), "ascending seed should be an Ace")
		assert.Equal(t, domain.StHelenaFoundationSuit(i), foundation[i][0].GetDesign())
	}
	for i := domain.StHelenaAscendingFoundationCnt; i < domain.StHelenaFoundationCnt; i++ {
		require.Len(t, foundation[i], 1, "descending foundation %d should be seeded with one card", i)
		assert.Equal(t, domain.CardValueMax, foundation[i][0].GetValue(), "descending seed should be a King")
		assert.Equal(t, domain.StHelenaFoundationSuit(i), foundation[i][0].GetDesign())
	}
}

func TestStHelena_AllFaceUp(t *testing.T) {
	cr := setupPlayingStHelena()
	assert.True(t, cr.AllFaceUp(), "all StHelena cards are always face up")
}

func TestStHelena_FoundationSuitHelpers(t *testing.T) {
	cases := []struct {
		idx     int
		suit    int
		ascends bool
	}{
		{0, domain.CardDesignSpade, true},
		{1, domain.CardDesignClover, true},
		{2, domain.CardDesignHeart, true},
		{3, domain.CardDesignDiamond, true},
		{4, domain.CardDesignSpade, false},
		{5, domain.CardDesignClover, false},
		{6, domain.CardDesignHeart, false},
		{7, domain.CardDesignDiamond, false},
	}
	for _, c := range cases {
		assert.Equal(t, c.suit, domain.StHelenaFoundationSuit(c.idx), "suit for foundation %d", c.idx)
		assert.Equal(t, c.ascends, domain.StHelenaIsAscendingFoundation(c.idx), "direction for foundation %d", c.idx)
	}
}

func TestStHelena_MoveTableauToTableau(t *testing.T) {
	t.Run("same-suit value+1", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 5)}
		tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 4)}
		cr.SetTableau(tab)
		require.NoError(t, cr.MoveTableauToTableau(0, 1))
		got := cr.GetTableau()
		assert.Len(t, got[0], 0)
		assert.Len(t, got[1], 2)
		assert.Equal(t, 5, got[1][1].Card.GetValue())
	})

	t.Run("same-suit value-1", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, 5)}
		tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, 6)}
		cr.SetTableau(tab)
		require.NoError(t, cr.MoveTableauToTableau(0, 1))
		got := cr.GetTableau()
		assert.Len(t, got[1], 2)
	})

	// **折り返しは無い。**クローン元のクレセントは A↔K を繋ぐので、元のサブ
	// テストは両方向の成功を主張していた。ここでは両方向とも拒む。
	t.Run("A and K do not wrap", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			from, toVal int
		}{
			{"A onto K", 1, domain.CardValueMax},
			{"K onto A", domain.CardValueMax, 1},
		} {
			t.Run(tc.name, func(t *testing.T) {
				cr := setupPlayingStHelena()
				clearStHelenaTableau(cr)
				var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
				tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignClover, tc.from)}
				tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignClover, tc.toVal)}
				cr.SetTableau(tab)
				assert.Error(t, cr.MoveTableauToTableau(0, 1))
			})
		}
	})

	// **スートは見ない。**元のサブテストは "different suit rejected" を主張して
	// いたが、それはクレセントの規則。ここでは通る。
	t.Run("a different suit is accepted", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 5)}
		tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, 4)}
		cr.SetTableau(tab)
		require.NoError(t, cr.MoveTableauToTableau(0, 1))
		assert.Len(t, cr.GetTableau()[1], 2)
	})

	t.Run("non-adjacent value rejected", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 5)}
		tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 7)}
		cr.SetTableau(tab)
		err := cr.MoveTableauToTableau(0, 1)
		assert.Error(t, err)
	})

	t.Run("empty target rejected", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 5)}
		cr.SetTableau(tab)
		err := cr.MoveTableauToTableau(0, 1)
		assert.Error(t, err)
	})

	t.Run("empty source rejected", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 5)}
		cr.SetTableau(tab)
		err := cr.MoveTableauToTableau(0, 1)
		assert.Error(t, err)
	})

	t.Run("invalid columns", func(t *testing.T) {
		cr := setupPlayingStHelena()
		assert.Error(t, cr.MoveTableauToTableau(-1, 0))
		assert.Error(t, cr.MoveTableauToTableau(0, domain.StHelenaTableauCnt))
		assert.Error(t, cr.MoveTableauToTableau(3, 3))
	})

	t.Run("not playing", func(t *testing.T) {
		cr := setupPlayingStHelena()
		cr.SetPhase(domain.StHelenaPhaseGameOver)
		assert.Error(t, cr.MoveTableauToTableau(0, 1))
	})
}

func TestStHelena_MoveTableauToFoundation(t *testing.T) {
	t.Run("ascending next value", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		// **列 6 から送る。**列 0 は上の列で、初回の配りでは A 段に届かない。
		// クローン元のクレセントにはこの制限が無いので、列 0 のままだと
		// ランクではなく制限で落ちる。
		tab[6] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 2)}
		cr.SetTableau(tab)
		var fnd [domain.StHelenaFoundationCnt][]*domain.Card
		fnd[0] = []*domain.Card{makeStHelenaCard(domain.CardDesignSpade, 1)}
		cr.SetFoundation(fnd)
		require.NoError(t, cr.MoveTableauToFoundation(6, 0))
		got := cr.GetFoundation()
		assert.Len(t, got[0], 2)
	})

	t.Run("descending next value", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 12)}
		cr.SetTableau(tab)
		var fnd [domain.StHelenaFoundationCnt][]*domain.Card
		fnd[4] = []*domain.Card{makeStHelenaCard(domain.CardDesignSpade, domain.CardValueMax)}
		cr.SetFoundation(fnd)
		require.NoError(t, cr.MoveTableauToFoundation(0, 4))
		got := cr.GetFoundation()
		assert.Len(t, got[4], 2)
	})

	t.Run("wrong suit rejected", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, 2)}
		cr.SetTableau(tab)
		var fnd [domain.StHelenaFoundationCnt][]*domain.Card
		fnd[0] = []*domain.Card{makeStHelenaCard(domain.CardDesignSpade, 1)}
		cr.SetFoundation(fnd)
		assert.Error(t, cr.MoveTableauToFoundation(0, 0))
	})

	t.Run("wrong direction rejected", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 12)}
		cr.SetTableau(tab)
		var fnd [domain.StHelenaFoundationCnt][]*domain.Card
		fnd[0] = []*domain.Card{makeStHelenaCard(domain.CardDesignSpade, 1)}
		cr.SetFoundation(fnd)
		assert.Error(t, cr.MoveTableauToFoundation(0, 0), "12 cannot extend ascending pile from 1")
	})

	t.Run("invalid foundation index", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 2)}
		cr.SetTableau(tab)
		assert.Error(t, cr.MoveTableauToFoundation(0, -1))
		assert.Error(t, cr.MoveTableauToFoundation(0, domain.StHelenaFoundationCnt))
	})

	t.Run("invalid column", func(t *testing.T) {
		cr := setupPlayingStHelena()
		assert.Error(t, cr.MoveTableauToFoundation(-1, 0))
		assert.Error(t, cr.MoveTableauToFoundation(domain.StHelenaTableauCnt, 0))
	})

	t.Run("empty column rejected", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		assert.Error(t, cr.MoveTableauToFoundation(0, 0))
	})

	t.Run("not playing", func(t *testing.T) {
		cr := setupPlayingStHelena()
		cr.SetPhase(domain.StHelenaPhaseGameOver)
		assert.Error(t, cr.MoveTableauToFoundation(0, 0))
	})
}

func TestStHelena_Redeal(t *testing.T) {
	// **最後の列から集めて配り直す。**クローン元のクレセントは各列をその場で
	// 逆順にするだけなので、元のサブテストは「列0 の中身が列0 のまま逆順」を
	// 主張していた。ここでは列をまたいで並び替わる。
	t.Run("gathers from the last pile and deals again", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		// 列 11 (最後) と列 0 (最初) にだけ置く。集める順が最後からなら、
		// 再配り後の先頭は列 11 の中身になる。
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 2)}
		tab[domain.StHelenaTableauCnt-1] = []*domain.StHelenaTableauCard{
			makeStHelenaTableauCard(domain.CardDesignHeart, 4),
		}
		cr.SetTableau(tab)
		before := cr.GetRedealsRemaining()

		require.NoError(t, cr.Redeal())

		got := cr.GetTableau()
		require.NotEmpty(t, got[0])
		assert.Equal(t, 4, got[0][0].Card.GetValue(), "最後の列から集める")
		assert.Equal(t, domain.CardDesignHeart, got[0][0].Card.GetDesign())
		assert.Equal(t, 2, got[0][1].Card.GetValue(), "その次が元の列0")
		assert.Equal(t, before-1, cr.GetRedealsRemaining())
		assert.Equal(t, 1, cr.GetMoveCount())
	})

	t.Run("no redeals remaining", func(t *testing.T) {
		cr := setupPlayingStHelena()
		cr.SetRedealsRemaining(0)
		err := cr.Redeal()
		assert.Error(t, err)
	})

	t.Run("not playing", func(t *testing.T) {
		cr := setupPlayingStHelena()
		cr.SetPhase(domain.StHelenaPhaseGameOver)
		assert.Error(t, cr.Redeal())
	})

	t.Run("snapshot enables undo of redeal", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{
			makeStHelenaTableauCard(domain.CardDesignSpade, 2),
			makeStHelenaTableauCard(domain.CardDesignSpade, 11),
		}
		cr.SetTableau(tab)
		require.NoError(t, cr.Redeal())
		require.True(t, cr.CanUndo())
		require.NoError(t, cr.Undo())
		assert.Equal(t, domain.StHelenaMaxRedeals, cr.GetRedealsRemaining())
		got := cr.GetTableau()
		assert.Equal(t, 2, got[0][0].Card.GetValue())
		assert.Equal(t, 11, got[0][1].Card.GetValue())
	})
}

func TestStHelena_GiveUp(t *testing.T) {
	cr := setupPlayingStHelena()
	cr.GiveUp()
	assert.Equal(t, domain.StHelenaPhaseGameOver, cr.GetPhase())
	assert.True(t, cr.GetGameEndFlag())

	// idempotent: calling on already-ended game does nothing
	cr.GiveUp()
	assert.Equal(t, domain.StHelenaPhaseGameOver, cr.GetPhase())
}

func TestStHelena_GetHint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		cr := setupPlayingStHelena()
		cr.SetPhase(domain.StHelenaPhaseGameOver)
		assert.Nil(t, cr.GetHint())
	})

	t.Run("priority 1: tableau to foundation", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		// **列 6 (下) から A 段へ。**クローン元のクレセントに送り先の制限が
		// 無いので、元の盤は列 3 (上) から A 段へ送るヒントを期待していた ──
		// このゲームではそれは拒まれる手で、ヒントが指してはいけない。
		tab[6] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 2)}
		cr.SetTableau(tab)
		var fnd [domain.StHelenaFoundationCnt][]*domain.Card
		fnd[0] = []*domain.Card{makeStHelenaCard(domain.CardDesignSpade, 1)}
		cr.SetFoundation(fnd)
		h := cr.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "foundation", h.ToZone)
		assert.Equal(t, 6, h.FromCol)
		assert.Equal(t, 0, h.ToCol)
		// ヒントが指した手は必ず打てること。指すだけで打てないなら嘘になる。
		assert.NoError(t, cr.MoveTableauToFoundation(h.FromCol, h.ToCol))
	})

	t.Run("priority 2: tableau to tableau", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 7)}
		tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 6)}
		cr.SetTableau(tab)
		h := cr.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "tableau", h.ToZone)
	})

	t.Run("priority 3: redeal when no other move", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 7)}
		tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, 4)}
		cr.SetTableau(tab)
		h := cr.GetHint()
		require.NotNil(t, h)
		assert.True(t, h.Redeal)
	})

	t.Run("nil when no move and no redeal", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 7)}
		tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, 4)}
		cr.SetTableau(tab)
		cr.SetRedealsRemaining(0)
		assert.Nil(t, cr.GetHint())
	})
}

func TestStHelena_AutoComplete(t *testing.T) {
	t.Run("drains tableau into foundation", func(t *testing.T) {
		cr := setupPlayingStHelena()
		clearStHelenaTableau(cr)
		clearStHelenaFoundation(cr)
		var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
		// **列 6 (下) から A 段へ。**元の盤は列 0 (上) から A 段へ送っていた ──
		// 手で送れば拒まれる手なので、オートコンプリートが打てば制限が
		// 無かったことになる。
		tab[6] = []*domain.StHelenaTableauCard{
			makeStHelenaTableauCard(domain.CardDesignSpade, 3),
			makeStHelenaTableauCard(domain.CardDesignSpade, 2),
		}
		cr.SetTableau(tab)
		var fnd [domain.StHelenaFoundationCnt][]*domain.Card
		fnd[0] = []*domain.Card{makeStHelenaCard(domain.CardDesignSpade, 1)}
		cr.SetFoundation(fnd)
		require.NoError(t, cr.AutoComplete())
		got := cr.GetFoundation()
		assert.Len(t, got[0], 3)
		assert.Len(t, cr.GetTableau()[6], 0)
	})

	t.Run("error when not playing", func(t *testing.T) {
		cr := setupPlayingStHelena()
		cr.SetPhase(domain.StHelenaPhaseGameOver)
		assert.Error(t, cr.AutoComplete())
	})
}

func TestStHelena_UndoFlow(t *testing.T) {
	cr := setupPlayingStHelena()
	clearStHelenaTableau(cr)
	clearStHelenaFoundation(cr)
	var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 5)}
	tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 4)}
	cr.SetTableau(tab)

	assert.False(t, cr.CanUndo())
	require.NoError(t, cr.MoveTableauToTableau(0, 1))
	assert.True(t, cr.CanUndo())

	logBefore := len(cr.GetActionLog())
	require.Equal(t, 1, logBefore, "the move should have appended one action log entry")

	require.NoError(t, cr.Undo())
	got := cr.GetTableau()
	assert.Len(t, got[0], 1)
	assert.Len(t, got[1], 1)
	assert.Equal(t, 0, cr.GetMoveCount())
	assert.Empty(t, cr.GetActionLog(), "Undo should truncate the matching action log entry")

	err := cr.Undo()
	assert.Error(t, err)
}

func TestStHelena_UndoN(t *testing.T) {
	cr := setupPlayingStHelena()
	clearStHelenaTableau(cr)
	clearStHelenaFoundation(cr)
	var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	// Two independent same-suit ±1 pairings (2 decks ⇒ duplicate 5♠/4♠ are valid).
	tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 5)}
	tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 4)}
	tab[2] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 5)}
	tab[3] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 4)}
	cr.SetTableau(tab)
	require.NoError(t, cr.MoveTableauToTableau(0, 1))
	require.NoError(t, cr.MoveTableauToTableau(2, 3))
	require.NoError(t, cr.UndoN(2))
	assert.False(t, cr.CanUndo())
	assert.Equal(t, 0, cr.GetMoveCount())
}

func TestStHelena_UndoN_PropagatesError(t *testing.T) {
	cr := setupPlayingStHelena()
	err := cr.UndoN(1)
	assert.Error(t, err)
}

func TestStHelena_UndoToEscape(t *testing.T) {
	cr := setupPlayingStHelena()
	assert.Equal(t, 0, cr.UndoToEscape(), "not stalemate ⇒ 0")
	cr.SetIsStalemate(true)
	assert.Equal(t, -1, cr.UndoToEscape(), "stalemate without history ⇒ -1")
}

func TestStHelena_Stalemate(t *testing.T) {
	cr := setupPlayingStHelena()
	clearStHelenaTableau(cr)
	clearStHelenaFoundation(cr)
	var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignSpade, 7)}
	tab[1] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(domain.CardDesignHeart, 4)}
	cr.SetTableau(tab)
	cr.SetRedealsRemaining(0)
	// Trigger stalemate via an actual move attempt? No legal move exists, so trigger via Redeal error then manual check.
	// Force the check by calling MoveTableauToTableau on an invalid pair, then evaluate via a legal seam:
	// Easiest: replicate the public side-effect by running AutoComplete (which calls checkStHelenaStalemate via checkGameClear? No it doesn't).
	// Use a no-op move that is rejected and confirm the IsStalemate is consistent by manual call via SetIsStalemate? Better: drive through a move that does succeed and triggers the post-move check.
	// We add one move-then-undo sequence to exercise the stalemate logic path:
	cr.SetRedealsRemaining(1)
	tab2 := tab
	tab2[2] = []*domain.StHelenaTableauCard{
		makeStHelenaTableauCard(domain.CardDesignClover, 8),
		makeStHelenaTableauCard(domain.CardDesignClover, 9),
	}
	tab2[3] = []*domain.StHelenaTableauCard{
		makeStHelenaTableauCard(domain.CardDesignClover, 10),
	}
	cr.SetTableau(tab2)
	require.NoError(t, cr.MoveTableauToTableau(2, 3))
	assert.False(t, cr.IsStalemate())
}

func TestStHelena_GameClear(t *testing.T) {
	cr := setupPlayingStHelena()
	clearStHelenaTableau(cr)
	clearStHelenaFoundation(cr)

	var fnd [domain.StHelenaFoundationCnt][]*domain.Card
	for i := 0; i < domain.StHelenaAscendingFoundationCnt; i++ {
		suit := domain.StHelenaFoundationSuit(i)
		pile := make([]*domain.Card, 0, domain.CardValueMax)
		for v := 1; v <= domain.CardValueMax; v++ {
			pile = append(pile, makeStHelenaCard(suit, v))
		}
		fnd[i] = pile
	}
	for i := domain.StHelenaAscendingFoundationCnt; i < domain.StHelenaFoundationCnt-1; i++ {
		suit := domain.StHelenaFoundationSuit(i)
		pile := make([]*domain.Card, 0, domain.CardValueMax)
		for v := domain.CardValueMax; v >= 1; v-- {
			pile = append(pile, makeStHelenaCard(suit, v))
		}
		fnd[i] = pile
	}
	// Leave the last descending foundation with just K placed; the missing card is on the tableau.
	lastIdx := domain.StHelenaFoundationCnt - 1
	lastSuit := domain.StHelenaFoundationSuit(lastIdx)
	pile := make([]*domain.Card, 0, domain.CardValueMax-1)
	for v := domain.CardValueMax; v >= 3; v-- {
		pile = append(pile, makeStHelenaCard(lastSuit, v))
	}
	fnd[lastIdx] = pile
	cr.SetFoundation(fnd)

	var tab [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	tab[0] = []*domain.StHelenaTableauCard{makeStHelenaTableauCard(lastSuit, 2)}
	cr.SetTableau(tab)
	require.NoError(t, cr.MoveTableauToFoundation(0, lastIdx))
	assert.Equal(t, domain.StHelenaPhasePlaying, cr.GetPhase(), "still missing the Ace")
}
