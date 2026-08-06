//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestAgnes() *domain.Agnes {
	return domain.NewAgnes(domain.NewTrumpCards(0))
}

func setupPlayingAgnes() *domain.Agnes {
	a := newTestAgnes()
	a.Reset()
	return a
}

func makeAgnesTC(design, value int, faceUp bool) *domain.AgnesTableauCard {
	return &domain.AgnesTableauCard{Card: domain.NewCard(design, value, false), FaceUp: faceUp}
}

func TestNewAgnes(t *testing.T) {
	a := newTestAgnes()
	assert.NotNil(t, a)
	assert.Equal(t, domain.AgnesPhase(0), a.GetPhase())
}

func TestNewDefaultAgnes(t *testing.T) {
	a := domain.NewDefaultAgnes()
	assert.NotNil(t, a)
}

func TestAgnes_Reset(t *testing.T) {
	a := setupPlayingAgnes()

	assert.Equal(t, domain.AgnesPhasePlaying, a.GetPhase())
	assert.Equal(t, 0, a.GetMoveCount())

	// Tableau: 7 columns, staircase 1..7, total 28 cards, only bottom face-up.
	tab := a.GetTableau()
	total := 0
	for i := 0; i < domain.AgnesTableauCnt; i++ {
		assert.Equal(t, i+1, len(tab[i]))
		total += len(tab[i])
		for j, tc := range tab[i] {
			if j == len(tab[i])-1 {
				assert.True(t, tc.FaceUp)
			} else {
				assert.False(t, tc.FaceUp)
			}
		}
	}
	assert.Equal(t, 28, total)

	// BaseRank set 1..13.
	br := a.GetBaseRank()
	assert.GreaterOrEqual(t, br, 1)
	assert.LessOrEqual(t, br, 13)

	// One foundation holds the single base card.
	foundation := a.GetFoundation()
	baseCount := 0
	for i := 0; i < domain.AgnesFoundationCnt; i++ {
		baseCount += len(foundation[i])
		for _, c := range foundation[i] {
			assert.Equal(t, br, c.GetValue())
		}
	}
	assert.Equal(t, 1, baseCount)

	// Stock: 52 - 28 - 1 = 23.
	assert.Equal(t, 23, a.GetStockCount())
}

func TestAgnes_DealStock(t *testing.T) {
	t.Run("deals one card to all 7 columns", func(t *testing.T) {
		a := setupPlayingAgnes()
		stock := make([]*domain.Card, 0, 10)
		for i := 0; i < 10; i++ {
			stock = append(stock, domain.NewCard(domain.CardDesignSpade, (i%13)+1, false))
		}
		a.SetStock(stock)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		for i := 0; i < domain.AgnesTableauCnt; i++ {
			tab[i] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 1, true)}
		}
		a.SetTableau(tab)

		assert.NoError(t, a.DealStock())
		assert.Equal(t, 3, a.GetStockCount()) // 10 - 7 = 3
		newTab := a.GetTableau()
		for i := 0; i < domain.AgnesTableauCnt; i++ {
			assert.Equal(t, 2, len(newTab[i]))
			assert.True(t, newTab[i][len(newTab[i])-1].FaceUp)
		}
		assert.Equal(t, 1, a.GetMoveCount())
	})

	t.Run("partial last deal of 2 cards", func(t *testing.T) {
		a := setupPlayingAgnes()
		// Only 2 cards in stock -> only cols 0 and 1 get a card.
		a.SetStock([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignClover, 6, false),
		})
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		for i := 0; i < domain.AgnesTableauCnt; i++ {
			tab[i] = nil
		}
		a.SetTableau(tab)

		assert.NoError(t, a.DealStock())
		assert.Equal(t, 0, a.GetStockCount())
		newTab := a.GetTableau()
		assert.Equal(t, 1, len(newTab[0]))
		assert.Equal(t, 1, len(newTab[1]))
		assert.Equal(t, 0, len(newTab[2]))
	})

	t.Run("error: stock empty", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetStock(nil)
		assert.Error(t, a.DealStock())
	})

	t.Run("error: not playing", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetPhase(domain.AgnesPhaseGameClear)
		assert.Error(t, a.DealStock())
	})
}

func TestAgnes_MoveTableauToTableau(t *testing.T) {
	t.Run("success same color one lower", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 6, true)}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 7, true)}
		a.SetTableau(tab)
		assert.NoError(t, a.MoveTableauToTableau(0, -1, 1))
		assert.Equal(t, 0, len(a.GetTableau()[0]))
		assert.Equal(t, 2, len(a.GetTableau()[1]))
		assert.Equal(t, 1, a.GetMoveCount())
	})

	t.Run("error: alternating color", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignHeart, 6, true)}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 7, true)}
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToTableau(0, -1, 1))
	})

	t.Run("error: same color but not one lower", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 5, true)}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 7, true)}
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToTableau(0, -1, 1))
	})

	t.Run("error: wrap Ace on 2 rejected", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		// Ace (value 1) cannot stack on a 2 (no wrap: 1 != 2-1 is false... actually 1==2-1,
		// but nothing stacks on Ace going further; here test 2-on-Ace which must wrap-fail).
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 13, true)}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 1, true)}
		a.SetTableau(tab)
		// King (13) on Ace (1): 13 != 1-1 -> rejected.
		assert.Error(t, a.MoveTableauToTableau(0, -1, 1))
	})

	t.Run("error: empty destination column", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 6, true)}
		tab[1] = nil
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToTableau(0, -1, 1))
	})

	t.Run("error: same column", func(t *testing.T) {
		a := setupPlayingAgnes()
		assert.Error(t, a.MoveTableauToTableau(0, -1, 0))
	})

	t.Run("error: invalid columns", func(t *testing.T) {
		a := setupPlayingAgnes()
		assert.Error(t, a.MoveTableauToTableau(-1, -1, 1))
		assert.Error(t, a.MoveTableauToTableau(7, -1, 1))
		assert.Error(t, a.MoveTableauToTableau(0, -1, -1))
		assert.Error(t, a.MoveTableauToTableau(0, -1, 7))
	})

	t.Run("error: from column empty", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = nil
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 7, true)}
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToTableau(0, -1, 1))
	})

	t.Run("error: not end card", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{
			makeAgnesTC(domain.CardDesignSpade, 6, true),
			makeAgnesTC(domain.CardDesignSpade, 5, true),
		}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 7, true)}
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToTableau(0, 0, 1))
	})

	t.Run("error: card face down", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 6, false)}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 7, true)}
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToTableau(0, -1, 1))
	})

	t.Run("error: not playing", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetPhase(domain.AgnesPhaseGameOver)
		assert.Error(t, a.MoveTableauToTableau(0, -1, 1))
	})
}

func TestAgnes_MoveTableauToFoundation(t *testing.T) {
	t.Run("success base rank starts empty foundation", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 5, true)}
		a.SetTableau(tab)
		assert.NoError(t, a.MoveTableauToFoundation(0))
		assert.Equal(t, 0, len(a.GetTableau()[0]))
		assert.Equal(t, 1, len(a.GetFoundation()[0]))
	})

	t.Run("success ascending same suit", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 6, true)}
		a.SetTableau(tab)
		assert.NoError(t, a.MoveTableauToFoundation(0))
	})

	t.Run("success wrap King to Ace", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(2)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)}
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 1, true)}
		a.SetTableau(tab)
		assert.NoError(t, a.MoveTableauToFoundation(0))
	})

	t.Run("error: wrong rank on empty", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 8, true)}
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToFoundation(0))
	})

	t.Run("error: wrong suit", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[2] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignHeart, 6, true)}
		a.SetTableau(tab)
		// Heart goes to foundation idx 2 which is empty and 6 != baseRank.
		assert.Error(t, a.MoveTableauToFoundation(2))
	})

	t.Run("error: empty column", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = nil
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToFoundation(0))
	})

	t.Run("error: invalid column", func(t *testing.T) {
		a := setupPlayingAgnes()
		assert.Error(t, a.MoveTableauToFoundation(-1))
		assert.Error(t, a.MoveTableauToFoundation(7))
	})

	t.Run("error: face down", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 5, false)}
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToFoundation(0))
	})

	t.Run("error: joker invalid for foundation", func(t *testing.T) {
		a := setupPlayingAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignJoker, 0, true)}
		a.SetTableau(tab)
		assert.Error(t, a.MoveTableauToFoundation(0))
	})

	t.Run("error: not playing", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetPhase(domain.AgnesPhaseGameClear)
		assert.Error(t, a.MoveTableauToFoundation(0))
	})
}

func TestAgnes_AutoFlipTableau(t *testing.T) {
	a := setupPlayingAgnes()
	var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
	// Column 0: face-down on top of which sits a movable face-up card.
	tab[0] = []*domain.AgnesTableauCard{
		makeAgnesTC(domain.CardDesignSpade, 10, false),
		makeAgnesTC(domain.CardDesignSpade, 6, true),
	}
	tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 7, true)}
	a.SetTableau(tab)
	assert.NoError(t, a.MoveTableauToTableau(0, -1, 1))
	// The newly-exposed bottom card of col 0 must now be face-up.
	remaining := a.GetTableau()[0]
	assert.Equal(t, 1, len(remaining))
	assert.True(t, remaining[0].FaceUp)
}

func TestAgnes_Win(t *testing.T) {
	a := setupPlayingAgnes()
	a.SetBaseRank(1)
	designs := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	var f [domain.AgnesFoundationCnt][]*domain.Card
	for i, d := range designs {
		pile := make([]*domain.Card, 0, 13)
		// 12 cards present, last one will be moved from tableau.
		for v := 1; v <= 12; v++ {
			pile = append(pile, domain.NewCard(d, v, false))
		}
		f[i] = pile
	}
	a.SetFoundation(f)
	// Tableau col 0 holds the final card (Spade King -> next after Queen).
	var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
	tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 13, true)}
	tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 13, true)}
	tab[2] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignHeart, 13, true)}
	tab[3] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignDiamond, 13, true)}
	a.SetTableau(tab)

	assert.NoError(t, a.MoveTableauToFoundation(0))
	assert.NoError(t, a.MoveTableauToFoundation(1))
	assert.NoError(t, a.MoveTableauToFoundation(2))
	assert.NoError(t, a.MoveTableauToFoundation(3))
	assert.Equal(t, domain.AgnesPhaseGameClear, a.GetPhase())
	assert.True(t, a.GetGameEndFlag())
}

func TestAgnes_GiveUp(t *testing.T) {
	a := setupPlayingAgnes()
	a.GiveUp()
	assert.Equal(t, domain.AgnesPhaseGameOver, a.GetPhase())
	assert.True(t, a.GetGameEndFlag())
	// idempotent
	a.GiveUp()
	assert.Equal(t, domain.AgnesPhaseGameOver, a.GetPhase())
}

func TestAgnes_GetHint(t *testing.T) {
	t.Run("nil when not playing", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetPhase(domain.AgnesPhaseGameOver)
		assert.Nil(t, a.GetHint())
	})

	t.Run("tableau to foundation priority", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 5, true)}
		a.SetTableau(tab)
		h := a.GetHint()
		assert.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("tableau to tableau", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(13)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 6, true)}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 7, true)}
		a.SetTableau(tab)
		h := a.GetHint()
		assert.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, "tableau", h.ToZone)
	})

	t.Run("no hint", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(13)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 3, true)}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 3, true)}
		a.SetTableau(tab)
		assert.Nil(t, a.GetHint())
	})

	t.Run("skips face-down end card", func(t *testing.T) {
		a := setupPlayingAgnes()
		a.SetBaseRank(13)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignSpade, 6, false)}
		tab[1] = []*domain.AgnesTableauCard{makeAgnesTC(domain.CardDesignClover, 7, false)}
		a.SetTableau(tab)
		assert.Nil(t, a.GetHint())
	})
}

func TestAgnes_Undo(t *testing.T) {
	a := setupPlayingAgnes()
	a.SetStock([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	assert.False(t, a.CanUndo())
	assert.NoError(t, a.DealStock())
	assert.True(t, a.CanUndo())
	assert.NoError(t, a.Undo())
	assert.False(t, a.CanUndo())
}

func TestAgnes_UndoErrors(t *testing.T) {
	a := setupPlayingAgnes()
	assert.Error(t, a.Undo())
	a.SetPhase(domain.AgnesPhaseGameOver)
	assert.Error(t, a.Undo())
}

func TestAgnes_UndoN(t *testing.T) {
	a := setupPlayingAgnes()
	a.SetStock([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
	})
	assert.NoError(t, a.DealStock())
	a.SetStock([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	assert.NoError(t, a.DealStock())
	assert.NoError(t, a.UndoN(2))
	assert.False(t, a.CanUndo())
	assert.Error(t, a.UndoN(1))
}

func TestAgnes_JSON(t *testing.T) {
	a := setupPlayingAgnes()
	data, err := a.MarshalJSON()
	assert.NoError(t, err)
	a2 := newTestAgnes()
	assert.NoError(t, a2.UnmarshalJSON(data))
	assert.Equal(t, a.GetBaseRank(), a2.GetBaseRank())
	assert.Equal(t, a.GetStockCount(), a2.GetStockCount())
}

func TestAgnes_UnmarshalJSON_Errors(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		a := newTestAgnes()
		assert.Error(t, a.UnmarshalJSON([]byte("not json")))
	})

	t.Run("nil tableau card", func(t *testing.T) {
		a := newTestAgnes()
		raw := `{"tc":null,"tb":[[null],[],[],[],[],[],[]],"st":[],"fd":[[],[],[],[]],"br":1,"ps":0,"mc":0,"al":[]}`
		assert.Error(t, a.UnmarshalJSON([]byte(raw)))
	})

	t.Run("nil foundation card", func(t *testing.T) {
		a := newTestAgnes()
		raw := `{"tc":null,"tb":[[],[],[],[],[],[],[]],"st":[],"fd":[[null],[],[],[]],"br":1,"ps":0,"mc":0,"al":[]}`
		assert.Error(t, a.UnmarshalJSON([]byte(raw)))
	})

	t.Run("nil stock card", func(t *testing.T) {
		a := newTestAgnes()
		raw := `{"tc":null,"tb":[[],[],[],[],[],[],[]],"st":[null],"fd":[[],[],[],[]],"br":1,"ps":0,"mc":0,"al":[]}`
		assert.Error(t, a.UnmarshalJSON([]byte(raw)))
	})

	t.Run("empty defaults", func(t *testing.T) {
		a := newTestAgnes()
		raw := `{"tc":null,"tb":[[],[],[],[],[],[],[]],"st":null,"fd":[[],[],[],[]],"br":3,"ps":0,"mc":0,"al":null}`
		assert.NoError(t, a.UnmarshalJSON([]byte(raw)))
		assert.Equal(t, 3, a.GetBaseRank())
		assert.Equal(t, 0, a.GetStockCount())
	})
}

// **手詰まりは GetHint と同じ判定であること (#4830)。**別のスキャンを書くと
// 「手詰まり」と言いながらヒントが手を返す状態が作れる。
func TestAgnes_IsStalemate(t *testing.T) {
	newBoard := func() *domain.Agnes {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetBaseRank(7)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		for i := range domain.AgnesTableauCnt {
			tab[i] = []*domain.AgnesTableauCard{
				{Card: domain.NewCard(domain.CardDesignSpade, 13, false), FaceUp: true},
			}
		}
		a.SetTableau(tab)
		a.SetStock(nil)
		return a
	}

	a := newBoard()
	assert.True(t, a.IsStalemate(), "動かせる札が無い")
	assert.Nil(t, a.GetHint(), "ヒントも手を返さない")

	// ストックが残っていれば手詰まりではない。
	withStock := newBoard()
	withStock.SetStock([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	assert.False(t, withStock.IsStalemate())

	// プレイ中以外では false。
	ended := newBoard()
	ended.SetPhase(domain.AgnesPhaseGameOver)
	assert.False(t, ended.IsStalemate())

	// 1 手でもあれば false。ベースランクの札はファンデーションへ置ける。
	movable := newBoard()
	var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
	for i := range domain.AgnesTableauCnt {
		tab[i] = []*domain.AgnesTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, 13, false), FaceUp: true},
		}
	}
	tab[0] = []*domain.AgnesTableauCard{
		{Card: domain.NewCard(domain.CardDesignSpade, 7, false), FaceUp: true},
	}
	movable.SetTableau(tab)
	assert.False(t, movable.IsStalemate())
	assert.NotNil(t, movable.GetHint())
}
