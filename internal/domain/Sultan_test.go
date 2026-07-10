//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestSultan() *domain.Sultan {
	tc := domain.NewTrumpCardsWithDecks(2, 0)
	return domain.NewSultan(tc)
}

func setupPlayingSultan() *domain.Sultan {
	su := newTestSultan()
	su.Reset()
	return su
}

func makeSuCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

// kingFoundations builds 8 foundations each seeded with a King, two per suit.
func kingFoundations() [domain.SultanFoundationCnt][]*domain.Card {
	var fd [domain.SultanFoundationCnt][]*domain.Card
	idx := 0
	for _, suit := range []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond} {
		for range 2 {
			fd[idx] = []*domain.Card{makeSuCard(suit, domain.CardValueMax)}
			idx++
		}
	}
	return fd
}

func TestNewSultan(t *testing.T) {
	su := newTestSultan()
	assert.NotNil(t, su)
	assert.Equal(t, domain.SultanPhase(0), su.GetPhase())
}

func TestNewDefaultSultan(t *testing.T) {
	su := domain.NewDefaultSultan()
	su.Reset()
	assert.Equal(t, domain.SultanPhasePlaying, su.GetPhase())
}

func TestSultan_Reset(t *testing.T) {
	su := setupPlayingSultan()

	assert.Equal(t, domain.SultanPhasePlaying, su.GetPhase())
	assert.Equal(t, 0, su.GetMoveCount())
	assert.Equal(t, 0, su.GetRedealCount())
	assert.False(t, su.IsStalemate())

	// 8 foundations, each seeded with exactly one King (value 13).
	foundation := su.GetFoundation()
	for i := range domain.SultanFoundationCnt {
		require.Len(t, foundation[i], 1, "foundation %d should start with one King", i)
		assert.Equal(t, domain.CardValueMax, foundation[i][0].GetValue(), "base must be a King")
	}

	// Divan: exactly 8 face-up cards.
	divan := su.GetDivan()
	require.Len(t, divan, domain.SultanDivanCnt)
	for i, c := range divan {
		assert.NotNil(t, c, "divan slot %d should be filled at start", i)
	}

	// Stock: 88 cards (104 - 8 Kings - 8 divan).
	assert.Equal(t, 88, su.GetStockCount())
	assert.Empty(t, su.GetWaste())

	// Total accounting: 8 + 8 + 88 = 104.
	assert.Equal(t, 8+domain.SultanDivanCnt+88, 104)
}

func TestSultan_Draw(t *testing.T) {
	su := setupPlayingSultan()
	before := su.GetStockCount()
	require.NoError(t, su.Draw())
	assert.Equal(t, before-1, su.GetStockCount())
	require.Len(t, su.GetWaste(), 1)
	assert.Equal(t, 1, su.GetMoveCount())
}

func TestSultan_DrawEmptyStock(t *testing.T) {
	su := setupPlayingSultan()
	su.SetStock(nil)
	err := su.Draw()
	require.Error(t, err)
}

func TestSultan_DrawNotPlaying(t *testing.T) {
	su := setupPlayingSultan()
	su.SetPhase(domain.SultanPhaseGameOver)
	require.Error(t, su.Draw())
}

func TestSultan_canPlaceFoundation_AceOnKing(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	su.SetDivan([]*domain.Card{makeSuCard(domain.CardDesignSpade, 1)}) // Ace of Spades
	su.SetStock(nil)
	require.NoError(t, su.MoveDivanToFoundation(0))
	// Ace landed on a spade King foundation.
	foundation := su.GetFoundation()
	placed := false
	for i := range domain.SultanFoundationCnt {
		if len(foundation[i]) == 2 && foundation[i][1].GetValue() == 1 {
			placed = true
		}
	}
	assert.True(t, placed, "Ace should be placed on a King foundation")
}

func TestSultan_canPlaceFoundation_WrongSuitRejected(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	// Heart Ace cannot go on any non-heart foundation; but there IS a heart King
	// foundation, so it should be accepted. Use a card whose suit has no matching
	// next-rank slot to force rejection: Spade 2 (needs Ace on top, but tops are Kings).
	su.SetDivan([]*domain.Card{makeSuCard(domain.CardDesignSpade, 2)})
	su.SetStock(nil)
	err := su.MoveDivanToFoundation(0)
	require.Error(t, err, "2 cannot be placed directly on a King")
}

func TestSultan_canPlaceFoundation_Sequence(t *testing.T) {
	su := setupPlayingSultan()
	fd := kingFoundations()
	// foundation[0] is a spade King; build King->Ace.
	fd[0] = append(fd[0], makeSuCard(domain.CardDesignSpade, 1))
	su.SetFoundation(fd)
	su.SetDivan([]*domain.Card{makeSuCard(domain.CardDesignSpade, 2)}) // needs 2 next
	su.SetStock(nil)
	require.NoError(t, su.MoveDivanToFoundation(0))
	assert.Equal(t, 2, su.GetFoundation()[0][2].GetValue())
}

func TestSultan_canPlaceFoundation_QueenCompletesAndBlocks(t *testing.T) {
	su := setupPlayingSultan()
	fd := kingFoundations()
	// Build foundation[0] fully up to Jack (King,A..J = 12 cards), needs Queen.
	for v := 1; v <= 11; v++ {
		fd[0] = append(fd[0], makeSuCard(domain.CardDesignSpade, v))
	}
	su.SetFoundation(fd)
	su.SetDivan([]*domain.Card{makeSuCard(domain.CardDesignSpade, 12)}) // Queen
	su.SetStock(nil)
	require.NoError(t, su.MoveDivanToFoundation(0))
	assert.Len(t, su.GetFoundation()[0], domain.SultanFoundationFull)

	// Now a card cannot be placed on a completed (Queen-topped) pile.
	fd2 := su.GetFoundation()
	su.SetFoundation(fd2)
	su.SetDivan([]*domain.Card{makeSuCard(domain.CardDesignSpade, 12)})
	err := su.MoveDivanToFoundation(0)
	require.Error(t, err, "completed Queen-topped pile rejects further cards")
}

func TestSultan_MoveDivanToFoundation_Refill(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	divan := make([]*domain.Card, domain.SultanDivanCnt)
	divan[0] = makeSuCard(domain.CardDesignHeart, 1) // Heart Ace
	su.SetDivan(divan)
	su.SetStock([]*domain.Card{makeSuCard(domain.CardDesignClover, 5)}) // refill source
	require.NoError(t, su.MoveDivanToFoundation(0))
	// Slot 0 refilled from stock top.
	assert.NotNil(t, su.GetDivan()[0])
	assert.Equal(t, domain.CardDesignClover, su.GetDivan()[0].GetDesign())
	assert.Equal(t, 0, su.GetStockCount())
}

func TestSultan_MoveDivanToFoundation_NoRefillWhenStockEmpty(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	divan := make([]*domain.Card, domain.SultanDivanCnt)
	divan[0] = makeSuCard(domain.CardDesignHeart, 1)
	su.SetDivan(divan)
	su.SetStock(nil)
	require.NoError(t, su.MoveDivanToFoundation(0))
	assert.Nil(t, su.GetDivan()[0], "slot stays nil when stock empty")
}

func TestSultan_MoveDivanToFoundation_Errors(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())

	// invalid index
	require.Error(t, su.MoveDivanToFoundation(-1))
	require.Error(t, su.MoveDivanToFoundation(999))

	// nil slot
	divan := make([]*domain.Card, domain.SultanDivanCnt)
	su.SetDivan(divan)
	require.Error(t, su.MoveDivanToFoundation(0))

	// card with no valid foundation (Spade 7 on Kings)
	divan[0] = makeSuCard(domain.CardDesignSpade, 7)
	su.SetDivan(divan)
	require.Error(t, su.MoveDivanToFoundation(0))

	// not playing
	su.SetPhase(domain.SultanPhaseGameOver)
	require.Error(t, su.MoveDivanToFoundation(0))
}

func TestSultan_MoveWasteToFoundation(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	su.SetDivan(make([]*domain.Card, domain.SultanDivanCnt))
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignDiamond, 1)})
	require.NoError(t, su.MoveWasteToFoundation())
	assert.Empty(t, su.GetWaste())
}

func TestSultan_MoveWasteToFoundation_Errors(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	su.SetWaste(nil)
	require.Error(t, su.MoveWasteToFoundation(), "empty waste")

	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 5)})
	require.Error(t, su.MoveWasteToFoundation(), "no valid foundation")

	su.SetPhase(domain.SultanPhaseGameOver)
	require.Error(t, su.MoveWasteToFoundation())
}

func TestSultan_Redeal(t *testing.T) {
	su := setupPlayingSultan()
	su.SetStock(nil)
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 3), makeSuCard(domain.CardDesignHeart, 4)})
	require.NoError(t, su.Redeal())
	assert.Equal(t, 2, su.GetStockCount())
	assert.Empty(t, su.GetWaste())
	assert.Equal(t, 1, su.GetRedealCount())
}

func TestSultan_Redeal_CapTwo(t *testing.T) {
	su := setupPlayingSultan()
	su.SetStock(nil)
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 3)})
	require.NoError(t, su.Redeal()) // 1
	su.SetStock(nil)
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 3)})
	require.NoError(t, su.Redeal()) // 2
	su.SetStock(nil)
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 3)})
	require.Error(t, su.Redeal(), "third redeal rejected")
	assert.Equal(t, domain.SultanMaxRedeal, su.GetRedealCount())
}

func TestSultan_Redeal_StockNotEmpty(t *testing.T) {
	su := setupPlayingSultan()
	su.SetStock([]*domain.Card{makeSuCard(domain.CardDesignSpade, 3)})
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 4)})
	require.Error(t, su.Redeal())
}

func TestSultan_Redeal_EmptyWaste(t *testing.T) {
	su := setupPlayingSultan()
	su.SetStock(nil)
	su.SetWaste(nil)
	require.Error(t, su.Redeal())
}

func TestSultan_Redeal_NotPlaying(t *testing.T) {
	su := setupPlayingSultan()
	su.SetPhase(domain.SultanPhaseGameClear)
	require.Error(t, su.Redeal())
}

func TestSultan_CanRedeal(t *testing.T) {
	su := setupPlayingSultan()
	su.SetStock(nil)
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 3)})
	assert.True(t, su.CanRedeal())
	su.SetRedealCount(domain.SultanMaxRedeal)
	assert.False(t, su.CanRedeal())
}

func TestSultan_GiveUp(t *testing.T) {
	su := setupPlayingSultan()
	su.GiveUp()
	assert.Equal(t, domain.SultanPhaseGameOver, su.GetPhase())
	assert.True(t, su.GetGameEndFlag())
	// no-op when not playing
	su.GiveUp()
	assert.Equal(t, domain.SultanPhaseGameOver, su.GetPhase())
}

func TestSultan_GetHint_Divan(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	divan := make([]*domain.Card, domain.SultanDivanCnt)
	divan[2] = makeSuCard(domain.CardDesignHeart, 1)
	su.SetDivan(divan)
	su.SetWaste(nil)
	hint := su.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "divan", hint.FromZone)
	assert.Equal(t, 2, hint.FromIdx)
	assert.GreaterOrEqual(t, hint.ToFoundation, 0)
}

func TestSultan_GetHint_Waste(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	su.SetDivan(make([]*domain.Card, domain.SultanDivanCnt))
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignClover, 1)})
	hint := su.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "waste", hint.FromZone)
	assert.Equal(t, -1, hint.FromIdx)
}

func TestSultan_GetHint_None(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	su.SetDivan(make([]*domain.Card, domain.SultanDivanCnt))
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 7)})
	assert.Nil(t, su.GetHint())

	su.SetPhase(domain.SultanPhaseGameOver)
	assert.Nil(t, su.GetHint())
}

func TestSultan_Stalemate(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	su.SetDivan(make([]*domain.Card, domain.SultanDivanCnt))
	su.SetStock(nil)
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignSpade, 7)})
	su.SetRedealCount(domain.SultanMaxRedeal)
	// trigger stalemate recompute via a draw error path? Use MoveWaste error doesn't recompute.
	// Force via a successful no-op: draw fails (empty stock). Instead use Redeal which recomputes,
	// but redeal is exhausted. Use a winning-free move: place an Ace then evaluate.
	divan := make([]*domain.Card, domain.SultanDivanCnt)
	divan[0] = makeSuCard(domain.CardDesignSpade, 1)
	su.SetDivan(divan)
	require.NoError(t, su.MoveDivanToFoundation(0)) // recomputes stalemate
	assert.True(t, su.IsStalemate())
}

func TestSultan_NotStalemateWithStock(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	divan := make([]*domain.Card, domain.SultanDivanCnt)
	divan[0] = makeSuCard(domain.CardDesignSpade, 1)
	su.SetDivan(divan)
	// Two stock cards: one refills the played slot, one remains so the player
	// can still draw (not a stalemate).
	su.SetStock([]*domain.Card{makeSuCard(domain.CardDesignClover, 7), makeSuCard(domain.CardDesignClover, 8)})
	su.SetWaste(nil)
	require.NoError(t, su.MoveDivanToFoundation(0))
	assert.False(t, su.IsStalemate(), "stock remaining means not stalemate")
}

func TestSultan_Win(t *testing.T) {
	su := setupPlayingSultan()
	fd := kingFoundations()
	// Fill all 8 foundations up to Jack (12 cards each), one card short of Queen.
	suits := []int{domain.CardDesignSpade, domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignHeart, domain.CardDesignDiamond, domain.CardDesignDiamond}
	for i := range domain.SultanFoundationCnt {
		for v := 1; v <= 11; v++ {
			fd[i] = append(fd[i], makeSuCard(suits[i], v))
		}
	}
	su.SetFoundation(fd)
	su.SetStock(nil)
	// Place the final Queen on each foundation via divan slot 0; the domain
	// auto-routes each Queen to the matching same-suit foundation.
	for i := range domain.SultanFoundationCnt {
		divan := make([]*domain.Card, domain.SultanDivanCnt)
		divan[0] = makeSuCard(suits[i], 12)
		su.SetDivan(divan)
		require.NoError(t, su.MoveDivanToFoundation(0))
	}
	assert.Equal(t, domain.SultanPhaseGameClear, su.GetPhase())
}

func TestSultan_Undo(t *testing.T) {
	su := setupPlayingSultan()
	require.False(t, su.CanUndo())
	before := su.GetStockCount()
	require.NoError(t, su.Draw())
	require.True(t, su.CanUndo())
	require.NoError(t, su.Undo())
	assert.Equal(t, before, su.GetStockCount())
	assert.Empty(t, su.GetWaste())

	require.Error(t, su.Undo(), "no history")
	su.SetPhase(domain.SultanPhaseGameOver)
	require.Error(t, su.Undo())
}

func TestSultan_UndoN(t *testing.T) {
	su := setupPlayingSultan()
	require.NoError(t, su.Draw())
	require.NoError(t, su.Draw())
	require.NoError(t, su.UndoN(2))
	assert.Empty(t, su.GetWaste())
	require.Error(t, su.UndoN(1))
}

func TestSultan_UndoToEscape(t *testing.T) {
	su := setupPlayingSultan()
	assert.Equal(t, 0, su.UndoToEscape(), "not stalemate -> 0")
	su.SetIsStalemate(true)
	assert.Equal(t, -1, su.UndoToEscape(), "no history -> -1")
}

func TestSultan_AutoComplete(t *testing.T) {
	su := setupPlayingSultan()
	su.SetFoundation(kingFoundations())
	divan := make([]*domain.Card, domain.SultanDivanCnt)
	divan[0] = makeSuCard(domain.CardDesignSpade, 1)
	su.SetDivan(divan)
	su.SetWaste([]*domain.Card{makeSuCard(domain.CardDesignHeart, 1)})
	su.SetStock(nil)
	require.NoError(t, su.AutoComplete())
	assert.Nil(t, su.GetDivan()[0])
	assert.Empty(t, su.GetWaste())

	// not all face up
	su.SetStock([]*domain.Card{makeSuCard(domain.CardDesignSpade, 2)})
	require.Error(t, su.AutoComplete())

	su.SetStock(nil)
	su.SetPhase(domain.SultanPhaseGameOver)
	require.Error(t, su.AutoComplete())
}

func TestSultan_AllFaceUp(t *testing.T) {
	su := setupPlayingSultan()
	su.SetStock(nil)
	assert.True(t, su.AllFaceUp())
	su.SetStock([]*domain.Card{makeSuCard(domain.CardDesignSpade, 1)})
	assert.False(t, su.AllFaceUp())
}

func TestSultan_GetActionLog(t *testing.T) {
	su := setupPlayingSultan()
	require.NoError(t, su.Draw())
	assert.NotEmpty(t, su.GetActionLog())
}
