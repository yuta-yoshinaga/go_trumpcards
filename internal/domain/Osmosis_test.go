//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestOsmosis() *domain.Osmosis {
	return domain.NewOsmosis(domain.NewTrumpCards(0))
}

func setupPlayingOsmosis() *domain.Osmosis {
	o := newTestOsmosis()
	o.Reset()
	return o
}

// osCard is a short alias for constructing a face-up card in tests.
func osCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func TestNewOsmosis(t *testing.T) {
	o := newTestOsmosis()
	assert.NotNil(t, o)
	assert.Equal(t, domain.OsmosisPhase(0), o.GetPhase())
}

func TestNewDefaultOsmosis(t *testing.T) {
	o := domain.NewDefaultOsmosis()
	assert.NotNil(t, o)
}

func TestOsmosis_Reset(t *testing.T) {
	o := setupPlayingOsmosis()
	assert.Equal(t, domain.OsmosisPhasePlaying, o.GetPhase())
	assert.Equal(t, 0, o.GetMoveCount())

	// 4 reserve piles of 4 = 16 cards
	reserve := o.GetReserve()
	total := 0
	for i := 0; i < domain.OsmosisReserveCnt; i++ {
		assert.Len(t, reserve[i], domain.OsmosisReservePileSize)
		total += len(reserve[i])
	}
	assert.Equal(t, 16, total)

	// Foundation row 0 has exactly the base card; rows 1-3 empty.
	foundation := o.GetFoundation()
	assert.Len(t, foundation[0], 1)
	assert.Equal(t, o.GetBaseRank(), foundation[0][0].GetValue())
	for i := 1; i < domain.OsmosisFoundationCnt; i++ {
		assert.Empty(t, foundation[i])
	}

	// Remaining 35 cards in stock, waste empty.
	assert.Equal(t, 35, o.GetStockCount())
	assert.Empty(t, o.GetWaste())
	assert.False(t, o.GetGameEndFlag())
	assert.False(t, o.CanUndo())
}

func TestOsmosis_Draw(t *testing.T) {
	o := setupPlayingOsmosis()
	before := o.GetStockCount()
	err := o.Draw()
	assert.NoError(t, err)
	assert.Equal(t, before-1, o.GetStockCount())
	assert.Len(t, o.GetWaste(), 1)
	assert.True(t, o.CanUndo())
}

func TestOsmosis_DrawRecycle(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetStock(nil)
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignSpade, 5), osCard(domain.CardDesignHeart, 6)})
	err := o.Draw()
	assert.NoError(t, err)
	assert.Equal(t, 2, o.GetStockCount())
	assert.Empty(t, o.GetWaste())
}

func TestOsmosis_DrawEmptyStockAndWaste(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetStock(nil)
	o.SetWaste(nil)
	err := o.Draw()
	assert.Error(t, err)
}

func TestOsmosis_DrawNotPlaying(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetPhase(domain.OsmosisPhaseGameOver)
	assert.Error(t, o.Draw())
}

// TestOsmosis_FoundationTopRowFreeOrder confirms the top row accepts any rank
// of its suit in any order.
func TestOsmosis_FoundationTopRowFreeOrder(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetBaseRank(8)
	var f [domain.OsmosisFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{osCard(domain.CardDesignSpade, 8)}
	o.SetFoundation(f)

	// Same suit, arbitrary rank (Jack) — allowed on top row.
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignSpade, 11)})
	assert.NoError(t, o.MoveWasteToFoundation(0))
	assert.Len(t, o.GetFoundation()[0], 2)

	// Different suit — rejected on the (now spade-locked) top row.
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 9)})
	assert.Error(t, o.MoveWasteToFoundation(0))
}

// TestOsmosis_LowerRowNeedsBaseRankToStart confirms a lower row can only be
// started with a base-rank card of an unused suit, and only after the row above
// has been started.
func TestOsmosis_LowerRowNeedsBaseRankToStart(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetBaseRank(8)
	var f [domain.OsmosisFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{osCard(domain.CardDesignSpade, 8)}
	o.SetFoundation(f)

	// Non-base rank cannot start row 1.
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 9)})
	assert.Error(t, o.MoveWasteToFoundation(1))

	// Base-rank heart starts row 1 (row 0 above is non-empty).
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 8)})
	assert.NoError(t, o.MoveWasteToFoundation(1))
	assert.Len(t, o.GetFoundation()[1], 1)

	// A suit already used by row 1 cannot start row 2.
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 8)})
	assert.Error(t, o.MoveWasteToFoundation(2))
}

// TestOsmosis_OsmosisSeepRule confirms a card joins a lower row only if its rank
// already appears in the row directly above.
func TestOsmosis_OsmosisSeepRule(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetBaseRank(8)
	var f [domain.OsmosisFoundationCnt][]*domain.Card
	// Row 0: 8, J (spades). Row 1 started with base heart 8.
	f[0] = []*domain.Card{osCard(domain.CardDesignSpade, 8), osCard(domain.CardDesignSpade, 11)}
	f[1] = []*domain.Card{osCard(domain.CardDesignHeart, 8)}
	o.SetFoundation(f)

	// Heart J can join row 1 because J exists in row 0 above.
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 11)})
	assert.NoError(t, o.MoveWasteToFoundation(1))

	// Heart 3 cannot join row 1 because 3 is not yet in row 0.
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 3)})
	assert.Error(t, o.MoveWasteToFoundation(1))
}

func TestOsmosis_MoveWasteInvalidIndex(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignSpade, 1)})
	assert.Error(t, o.MoveWasteToFoundation(-1))
	assert.Error(t, o.MoveWasteToFoundation(domain.OsmosisFoundationCnt))
}

func TestOsmosis_MoveWasteEmpty(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetWaste(nil)
	assert.Error(t, o.MoveWasteToFoundation(0))
}

func TestOsmosis_MoveReserveToFoundation(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetBaseRank(8)
	var f [domain.OsmosisFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{osCard(domain.CardDesignSpade, 8)}
	o.SetFoundation(f)

	var r [domain.OsmosisReserveCnt][]*domain.Card
	r[2] = []*domain.Card{osCard(domain.CardDesignSpade, 2), osCard(domain.CardDesignSpade, 9)}
	o.SetReserve(r)

	assert.NoError(t, o.MoveReserveToFoundation(2, 0))
	assert.Len(t, o.GetReserve()[2], 1)
	assert.Len(t, o.GetFoundation()[0], 2)
}

func TestOsmosis_MoveReserveInvalid(t *testing.T) {
	o := setupPlayingOsmosis()
	assert.Error(t, o.MoveReserveToFoundation(-1, 0))
	assert.Error(t, o.MoveReserveToFoundation(0, -1))
	var r [domain.OsmosisReserveCnt][]*domain.Card
	o.SetReserve(r) // all empty
	assert.Error(t, o.MoveReserveToFoundation(0, 0))
}

func TestOsmosis_MoveNotPlaying(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetPhase(domain.OsmosisPhaseGameOver)
	assert.Error(t, o.MoveWasteToFoundation(0))
	assert.Error(t, o.MoveReserveToFoundation(0, 0))
}

func TestOsmosis_GiveUp(t *testing.T) {
	o := setupPlayingOsmosis()
	o.GiveUp()
	assert.Equal(t, domain.OsmosisPhaseGameOver, o.GetPhase())
	assert.True(t, o.GetGameEndFlag())
	// Giving up again is a no-op.
	o.GiveUp()
	assert.Equal(t, domain.OsmosisPhaseGameOver, o.GetPhase())
}

func TestOsmosis_Hint(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetBaseRank(8)
	var f [domain.OsmosisFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{osCard(domain.CardDesignSpade, 8)}
	o.SetFoundation(f)

	// No playable card → no hint.
	var r [domain.OsmosisReserveCnt][]*domain.Card
	r[0] = []*domain.Card{osCard(domain.CardDesignHeart, 2)}
	o.SetReserve(r)
	o.SetWaste(nil)
	assert.Nil(t, o.GetHint())

	// Reserve playable → hint from reserve.
	r[0] = []*domain.Card{osCard(domain.CardDesignSpade, 9)}
	o.SetReserve(r)
	h := o.GetHint()
	assert.NotNil(t, h)
	assert.Equal(t, "reserve", h.FromZone)
	assert.Equal(t, 0, h.FromCol)
	assert.Equal(t, 0, h.ToCol)

	// Only waste playable → hint from waste.
	var empty [domain.OsmosisReserveCnt][]*domain.Card
	o.SetReserve(empty)
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignSpade, 5)})
	h = o.GetHint()
	assert.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
	assert.Equal(t, -1, h.FromCol)
}

func TestOsmosis_HintNotPlaying(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetPhase(domain.OsmosisPhaseGameClear)
	assert.Nil(t, o.GetHint())
}

func TestOsmosis_AutoComplete(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetBaseRank(8)
	var f [domain.OsmosisFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{osCard(domain.CardDesignSpade, 8)}
	o.SetFoundation(f)
	var r [domain.OsmosisReserveCnt][]*domain.Card
	r[0] = []*domain.Card{osCard(domain.CardDesignSpade, 9), osCard(domain.CardDesignSpade, 10)}
	o.SetReserve(r)
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignSpade, 5)})

	err := o.AutoComplete()
	assert.NoError(t, err)
	// All three spades land on the top row (8,10,9,5 order — all same suit).
	assert.Len(t, o.GetFoundation()[0], 4)
	assert.Empty(t, o.GetReserve()[0])
	assert.Empty(t, o.GetWaste())
}

func TestOsmosis_AutoCompleteNoMove(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetBaseRank(8)
	var f [domain.OsmosisFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{osCard(domain.CardDesignSpade, 8)}
	o.SetFoundation(f)
	var r [domain.OsmosisReserveCnt][]*domain.Card
	r[0] = []*domain.Card{osCard(domain.CardDesignHeart, 2)}
	o.SetReserve(r)
	o.SetWaste(nil)
	before := o.CanUndo()
	assert.Error(t, o.AutoComplete())
	// Snapshot must be rolled back when nothing moved.
	assert.Equal(t, before, o.CanUndo())
}

func TestOsmosis_AutoCompleteNotPlaying(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetPhase(domain.OsmosisPhaseGameOver)
	assert.Error(t, o.AutoComplete())
}

func TestOsmosis_Undo(t *testing.T) {
	o := setupPlayingOsmosis()
	stockBefore := o.GetStockCount()
	assert.NoError(t, o.Draw())
	assert.NoError(t, o.Undo())
	assert.Equal(t, stockBefore, o.GetStockCount())
	assert.Empty(t, o.GetWaste())
}

func TestOsmosis_UndoNoHistory(t *testing.T) {
	o := setupPlayingOsmosis()
	assert.Error(t, o.Undo())
}

func TestOsmosis_UndoNotPlaying(t *testing.T) {
	o := setupPlayingOsmosis()
	assert.NoError(t, o.Draw())
	o.SetPhase(domain.OsmosisPhaseGameOver)
	assert.Error(t, o.Undo())
}

func TestOsmosis_UndoN(t *testing.T) {
	o := setupPlayingOsmosis()
	assert.NoError(t, o.Draw())
	assert.NoError(t, o.Draw())
	assert.NoError(t, o.UndoN(2))
	assert.Empty(t, o.GetWaste())
	// UndoN past history returns an error.
	assert.Error(t, o.UndoN(1))
}

func TestOsmosis_WinCondition(t *testing.T) {
	o := setupPlayingOsmosis()
	o.SetBaseRank(1)
	// Fill all rows to 13 except leave one card for the final move.
	var f [domain.OsmosisFoundationCnt][]*domain.Card
	for i := 0; i < domain.OsmosisFoundationCnt; i++ {
		suit := i + 1
		for v := 1; v <= domain.CardValueMax; v++ {
			if i == 0 && v == 13 {
				continue // leave the spade King out
			}
			f[i] = append(f[i], osCard(suit, v))
		}
	}
	o.SetFoundation(f)
	o.SetWaste([]*domain.Card{osCard(domain.CardDesignSpade, 13)})

	assert.NoError(t, o.MoveWasteToFoundation(0))
	assert.Equal(t, domain.OsmosisPhaseGameClear, o.GetPhase())
	assert.True(t, o.GetGameEndFlag())
}

func TestOsmosis_GetActionLog(t *testing.T) {
	o := setupPlayingOsmosis()
	assert.Empty(t, o.GetActionLog())
	assert.NoError(t, o.Draw())
	assert.NotEmpty(t, o.GetActionLog())
}

// TestOsmosis_IsStalemate confirms the dead-end detector fires only when no
// card anywhere can still reach a foundation.
//
// **山札が0枚かどうかは条件にならない。**Draw はウェイストをストックに戻して
// 循環させるので、山札の中身は何周でも見に行ける。詰みは「リザーブのトップ・
// ストック・ウェイストのどのカードも置けない」ときだけ (#4808)。
func TestOsmosis_IsStalemate(t *testing.T) {
	// 8 をベースにし、♠ の段だけ開いた盤面を作る。
	setup := func() *domain.Osmosis {
		o := setupPlayingOsmosis()
		o.SetBaseRank(8)
		var f [domain.OsmosisFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{osCard(domain.CardDesignSpade, 8)}
		o.SetFoundation(f)
		o.SetStock(nil)
		o.SetWaste(nil)
		o.SetReserve([domain.OsmosisReserveCnt][]*domain.Card{})
		return o
	}

	t.Run("no card anywhere can be placed", func(t *testing.T) {
		o := setup()
		// ♥9 は ♠ 段には置けず、下の段はベースランクでないと始められない。
		o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 9)})
		assert.True(t, o.IsStalemate())
	})

	t.Run("a playable reserve top keeps the game alive", func(t *testing.T) {
		o := setup()
		o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 9)})
		var r [domain.OsmosisReserveCnt][]*domain.Card
		r[0] = []*domain.Card{osCard(domain.CardDesignSpade, 11)}
		o.SetReserve(r)
		assert.False(t, o.IsStalemate())
	})

	// **循環するので山札の奥も数える。**トップだけ見ると、めくれば置ける札が
	// 残っているのに詰みと言ってしまう。
	t.Run("a playable card buried in the stock keeps the game alive", func(t *testing.T) {
		o := setup()
		o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 9)})
		o.SetStock([]*domain.Card{
			osCard(domain.CardDesignSpade, 11),
			osCard(domain.CardDesignHeart, 10),
			osCard(domain.CardDesignClover, 3),
		})
		assert.False(t, o.IsStalemate())
	})

	t.Run("never reports a stalemate once the game has ended", func(t *testing.T) {
		o := setup()
		o.SetWaste([]*domain.Card{osCard(domain.CardDesignHeart, 9)})
		o.GiveUp()
		assert.False(t, o.IsStalemate())
	})
}
