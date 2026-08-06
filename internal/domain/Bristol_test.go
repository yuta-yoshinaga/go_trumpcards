//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestBristol() *domain.Bristol {
	return domain.NewBristol(domain.NewTrumpCards(0))
}

func setupPlayingBristol() *domain.Bristol {
	b := newTestBristol()
	b.Reset()
	return b
}

// brCard is a short alias for constructing a face-up card in tests.
func brCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func TestNewBristol(t *testing.T) {
	b := newTestBristol()
	assert.NotNil(t, b)
	assert.Equal(t, domain.BristolPhase(0), b.GetPhase())
}

func TestNewDefaultBristol(t *testing.T) {
	b := domain.NewDefaultBristol()
	assert.NotNil(t, b)
}

func TestBristol_Reset(t *testing.T) {
	b := setupPlayingBristol()
	assert.Equal(t, domain.BristolPhasePlaying, b.GetPhase())
	assert.Equal(t, 0, b.GetMoveCount())

	// 8 tableau columns of 3 = 24 cards.
	tableau := b.GetTableau()
	total := 0
	for i := 0; i < domain.BristolTableauCnt; i++ {
		assert.Len(t, tableau[i], domain.BristolTableauInitial)
		total += len(tableau[i])
	}
	assert.Equal(t, 24, total)

	// Fans start empty.
	fan := b.GetFan()
	for i := 0; i < domain.BristolFanCnt; i++ {
		assert.Empty(t, fan[i])
	}

	// Foundations start empty.
	foundation := b.GetFoundation()
	for i := 0; i < domain.BristolFoundationCnt; i++ {
		assert.Empty(t, foundation[i])
	}

	// Remaining 28 cards in stock.
	assert.Equal(t, 28, b.GetStockCount())
	assert.False(t, b.GetGameEndFlag())
	assert.False(t, b.CanUndo())
}

func TestBristol_Draw(t *testing.T) {
	b := setupPlayingBristol()
	before := b.GetStockCount()
	err := b.Draw()
	assert.NoError(t, err)
	// 3 cards dealt, one to each fan.
	assert.Equal(t, before-3, b.GetStockCount())
	for i := 0; i < domain.BristolFanCnt; i++ {
		assert.Len(t, b.GetFan()[i], 1)
	}
	assert.True(t, b.CanUndo())
}

func TestBristol_DrawPartialStock(t *testing.T) {
	b := setupPlayingBristol()
	b.SetStock([]*domain.Card{brCard(domain.CardDesignSpade, 5), brCard(domain.CardDesignHeart, 6)})
	err := b.Draw()
	assert.NoError(t, err)
	assert.Equal(t, 0, b.GetStockCount())
	assert.Len(t, b.GetFan()[0], 1)
	assert.Len(t, b.GetFan()[1], 1)
	assert.Empty(t, b.GetFan()[2])
}

func TestBristol_DrawEmptyStock(t *testing.T) {
	b := setupPlayingBristol()
	b.SetStock(nil)
	assert.Error(t, b.Draw())
}

func TestBristol_DrawNotPlaying(t *testing.T) {
	b := setupPlayingBristol()
	b.SetPhase(domain.BristolPhaseGameOver)
	assert.Error(t, b.Draw())
}

func TestBristol_MoveTableauToTableau(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 8)}
	tb[1] = []*domain.Card{brCard(domain.CardDesignHeart, 9)}
	b.SetTableau(tb)

	// Spade 8 onto Heart 9 (descending, any suit) is allowed.
	assert.NoError(t, b.MoveTableauToTableau(0, 1))
	assert.Empty(t, b.GetTableau()[0])
	assert.Len(t, b.GetTableau()[1], 2)
}

func TestBristol_MoveTableauToTableauInvalid(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 8)}
	tb[1] = []*domain.Card{brCard(domain.CardDesignHeart, 5)}
	b.SetTableau(tb)

	// 8 cannot go on 5 (not descending by 1).
	assert.Error(t, b.MoveTableauToTableau(0, 1))
	// Same column.
	assert.Error(t, b.MoveTableauToTableau(0, 0))
	// Invalid indices.
	assert.Error(t, b.MoveTableauToTableau(-1, 1))
	assert.Error(t, b.MoveTableauToTableau(0, domain.BristolTableauCnt))
}

// TestBristol_EmptyColumnIsDead confirms a card cannot be placed on an empty
// tableau column (Bristol rule: once a column empties it can no longer be used).
func TestBristol_EmptyColumnIsDead(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 8)}
	// column 1 is empty.
	b.SetTableau(tb)
	assert.Error(t, b.MoveTableauToTableau(0, 1))
}

func TestBristol_MoveTableauToTableauEmptySource(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[1] = []*domain.Card{brCard(domain.CardDesignHeart, 9)}
	b.SetTableau(tb)
	assert.Error(t, b.MoveTableauToTableau(0, 1))
}

func TestBristol_MoveTableauToFoundation(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 1)}
	tb[1] = []*domain.Card{brCard(domain.CardDesignHeart, 2)}
	b.SetTableau(tb)

	// Ace starts a foundation.
	assert.NoError(t, b.MoveTableauToFoundation(0))
	assert.Len(t, b.GetFoundation()[0], 1)
	// 2 (any suit) builds up on it.
	assert.NoError(t, b.MoveTableauToFoundation(1))
	assert.Len(t, b.GetFoundation()[0], 2)
}

func TestBristol_MoveTableauToFoundationInvalid(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 5)}
	b.SetTableau(tb)
	// 5 cannot start an empty foundation.
	assert.Error(t, b.MoveTableauToFoundation(0))
	// Empty column.
	assert.Error(t, b.MoveTableauToFoundation(1))
	// Invalid column.
	assert.Error(t, b.MoveTableauToFoundation(-1))
}

func TestBristol_MoveFanToTableau(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 9)}
	b.SetTableau(tb)
	var fn [domain.BristolFanCnt][]*domain.Card
	fn[1] = []*domain.Card{brCard(domain.CardDesignClover, 8)}
	b.SetFan(fn)

	assert.NoError(t, b.MoveFanToTableau(1, 0))
	assert.Empty(t, b.GetFan()[1])
	assert.Len(t, b.GetTableau()[0], 2)
}

func TestBristol_MoveFanToTableauInvalid(t *testing.T) {
	b := setupPlayingBristol()
	assert.Error(t, b.MoveFanToTableau(-1, 0))
	assert.Error(t, b.MoveFanToTableau(0, -1))
	var fn [domain.BristolFanCnt][]*domain.Card
	b.SetFan(fn) // all empty
	assert.Error(t, b.MoveFanToTableau(0, 0))
}

func TestBristol_MoveFanToFoundation(t *testing.T) {
	b := setupPlayingBristol()
	var fn [domain.BristolFanCnt][]*domain.Card
	fn[2] = []*domain.Card{brCard(domain.CardDesignDiamond, 1)}
	b.SetFan(fn)
	assert.NoError(t, b.MoveFanToFoundation(2))
	assert.Empty(t, b.GetFan()[2])
	assert.Len(t, b.GetFoundation()[0], 1)
}

func TestBristol_MoveFanToFoundationInvalid(t *testing.T) {
	b := setupPlayingBristol()
	assert.Error(t, b.MoveFanToFoundation(-1))
	var fn [domain.BristolFanCnt][]*domain.Card
	fn[0] = []*domain.Card{brCard(domain.CardDesignSpade, 7)}
	b.SetFan(fn)
	// 7 cannot start a foundation.
	assert.Error(t, b.MoveFanToFoundation(0))
	// Empty fan.
	assert.Error(t, b.MoveFanToFoundation(1))
}

func TestBristol_MoveNotPlaying(t *testing.T) {
	b := setupPlayingBristol()
	b.SetPhase(domain.BristolPhaseGameOver)
	assert.Error(t, b.MoveTableauToTableau(0, 1))
	assert.Error(t, b.MoveTableauToFoundation(0))
	assert.Error(t, b.MoveFanToTableau(0, 0))
	assert.Error(t, b.MoveFanToFoundation(0))
}

func TestBristol_GiveUp(t *testing.T) {
	b := setupPlayingBristol()
	b.GiveUp()
	assert.Equal(t, domain.BristolPhaseGameOver, b.GetPhase())
	assert.True(t, b.GetGameEndFlag())
	// Giving up again is a no-op.
	b.GiveUp()
	assert.Equal(t, domain.BristolPhaseGameOver, b.GetPhase())
}

func TestBristol_Hint(t *testing.T) {
	b := setupPlayingBristol()

	// No playable card → no hint.
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignHeart, 7)}
	b.SetTableau(tb)
	assert.Nil(t, b.GetHint())

	// Tableau Ace → foundation hint (priority 1).
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 1)}
	b.SetTableau(tb)
	h := b.GetHint()
	assert.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, "foundation", h.ToZone)

	// Fan Ace → foundation hint (priority 2) when no tableau move exists.
	var empty [domain.BristolTableauCnt][]*domain.Card
	empty[0] = []*domain.Card{brCard(domain.CardDesignHeart, 7)}
	b.SetTableau(empty)
	var fn [domain.BristolFanCnt][]*domain.Card
	fn[0] = []*domain.Card{brCard(domain.CardDesignClover, 1)}
	b.SetFan(fn)
	h = b.GetHint()
	assert.NotNil(t, h)
	assert.Equal(t, "fan", h.FromZone)
	assert.Equal(t, "foundation", h.ToZone)
}

func TestBristol_HintTableauToTableau(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 8)}
	tb[1] = []*domain.Card{brCard(domain.CardDesignHeart, 9)}
	b.SetTableau(tb)
	h := b.GetHint()
	assert.NotNil(t, h)
	assert.Equal(t, "tableau", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
}

func TestBristol_HintFanToTableau(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignHeart, 9)}
	b.SetTableau(tb)
	var fn [domain.BristolFanCnt][]*domain.Card
	fn[0] = []*domain.Card{brCard(domain.CardDesignSpade, 8)}
	b.SetFan(fn)
	h := b.GetHint()
	assert.NotNil(t, h)
	assert.Equal(t, "fan", h.FromZone)
	assert.Equal(t, "tableau", h.ToZone)
}

func TestBristol_HintNotPlaying(t *testing.T) {
	b := setupPlayingBristol()
	b.SetPhase(domain.BristolPhaseGameClear)
	assert.Nil(t, b.GetHint())
}

func TestBristol_AutoComplete(t *testing.T) {
	b := setupPlayingBristol()
	// AutoComplete only moves the top card of each pile, so the Ace and 2 must be
	// reachable (separate column tops) rather than stacked under one another.
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 1)}
	tb[1] = []*domain.Card{brCard(domain.CardDesignHeart, 2)}
	b.SetTableau(tb)
	var fn [domain.BristolFanCnt][]*domain.Card
	fn[0] = []*domain.Card{brCard(domain.CardDesignClover, 3)}
	b.SetFan(fn)

	err := b.AutoComplete()
	assert.NoError(t, err)
	// A,2,3 all land on foundation 0.
	assert.Len(t, b.GetFoundation()[0], 3)
	assert.Empty(t, b.GetTableau()[0])
	assert.Empty(t, b.GetTableau()[1])
	assert.Empty(t, b.GetFan()[0])
}

func TestBristol_AutoCompleteNoMove(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 7)}
	b.SetTableau(tb)
	before := b.CanUndo()
	assert.Error(t, b.AutoComplete())
	// Snapshot must be rolled back when nothing moved.
	assert.Equal(t, before, b.CanUndo())
}

func TestBristol_AutoCompleteNotPlaying(t *testing.T) {
	b := setupPlayingBristol()
	b.SetPhase(domain.BristolPhaseGameOver)
	assert.Error(t, b.AutoComplete())
}

func TestBristol_Undo(t *testing.T) {
	b := setupPlayingBristol()
	stockBefore := b.GetStockCount()
	assert.NoError(t, b.Draw())
	assert.NoError(t, b.Undo())
	assert.Equal(t, stockBefore, b.GetStockCount())
	for i := 0; i < domain.BristolFanCnt; i++ {
		assert.Empty(t, b.GetFan()[i])
	}
}

func TestBristol_UndoNoHistory(t *testing.T) {
	b := setupPlayingBristol()
	assert.Error(t, b.Undo())
}

func TestBristol_UndoNotPlaying(t *testing.T) {
	b := setupPlayingBristol()
	assert.NoError(t, b.Draw())
	b.SetPhase(domain.BristolPhaseGameOver)
	assert.Error(t, b.Undo())
}

func TestBristol_UndoN(t *testing.T) {
	b := setupPlayingBristol()
	assert.NoError(t, b.Draw())
	assert.NoError(t, b.Draw())
	assert.NoError(t, b.UndoN(2))
	for i := 0; i < domain.BristolFanCnt; i++ {
		assert.Empty(t, b.GetFan()[i])
	}
	// UndoN past history returns an error.
	assert.Error(t, b.UndoN(1))
}

func TestBristol_WinCondition(t *testing.T) {
	b := setupPlayingBristol()
	// Fill all four foundations to K (suit-agnostic), leaving the last King
	// of foundation 0 in a fan for the final move.
	var f [domain.BristolFoundationCnt][]*domain.Card
	for i := 0; i < domain.BristolFoundationCnt; i++ {
		suit := i + 1
		for v := 1; v <= domain.CardValueMax; v++ {
			if i == 0 && v == domain.CardValueMax {
				continue // leave foundation 0's King out
			}
			f[i] = append(f[i], brCard(suit, v))
		}
	}
	b.SetFoundation(f)
	var fn [domain.BristolFanCnt][]*domain.Card
	fn[0] = []*domain.Card{brCard(domain.CardDesignSpade, domain.CardValueMax)}
	b.SetFan(fn)

	assert.NoError(t, b.MoveFanToFoundation(0))
	assert.Equal(t, domain.BristolPhaseGameClear, b.GetPhase())
	assert.True(t, b.GetGameEndFlag())
}

func TestBristol_GetActionLog(t *testing.T) {
	b := setupPlayingBristol()
	assert.Empty(t, b.GetActionLog())
	assert.NoError(t, b.Draw())
	assert.NotEmpty(t, b.GetActionLog())
}

// **押すまで合法か分からなかった (#4813)。**移動元ごとに、置ける先だけを返すこと。
func TestBristol_LegalTargets(t *testing.T) {
	b := setupPlayingBristol()
	var tb [domain.BristolTableauCnt][]*domain.Card
	tb[0] = []*domain.Card{brCard(domain.CardDesignSpade, 8)}
	tb[1] = []*domain.Card{brCard(domain.CardDesignHeart, 9)} // 8 を置ける
	tb[2] = []*domain.Card{brCard(domain.CardDesignClover, 4)}
	b.SetTableau(tb)
	var fn [domain.BristolFanCnt][]*domain.Card
	fn[0] = []*domain.Card{brCard(domain.CardDesignDiamond, 5)} // 4 の上には置けない
	b.SetFan(fn)
	var fd [domain.BristolFoundationCnt][]*domain.Card
	fd[0] = []*domain.Card{brCard(domain.CardDesignSpade, 7)} // 8 を置ける
	b.SetFoundation(fd)

	tab, found := b.LegalTargets("tableau", 0)
	assert.Equal(t, []int{1}, tab, "♠8 は ♥9 の上だけ")
	assert.Equal(t, []int{0}, found, "♠7 の上に ♠8")

	// 置ける先が無い移動元。
	tab2, found2 := b.LegalTargets("fan", 0)
	assert.Nil(t, tab2)
	assert.Nil(t, found2)

	// 空の移動元・未知のゾーン・範囲外。
	for _, tc := range []struct {
		zone string
		col  int
	}{{"fan", 1}, {"tableau", 99}, {"fan", -1}, {"stock", 0}} {
		tab3, found3 := b.LegalTargets(tc.zone, tc.col)
		assert.Nil(t, tab3, tc.zone)
		assert.Nil(t, found3, tc.zone)
	}
}
