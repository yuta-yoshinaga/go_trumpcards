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

func narCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

func newNarcotic() *domain.Narcotic {
	return domain.NewNarcotic(domain.NewTrumpCards(0))
}

func playingNarcotic(t *testing.T) *domain.Narcotic {
	t.Helper()
	g := newNarcotic()
	g.Reset()
	return g
}

// narBoard は4列の盤を作る。
func narBoard(cols ...[]*domain.Card) [domain.NarcoticColCnt][]*domain.Card {
	var out [domain.NarcoticColCnt][]*domain.Card
	for i := range out {
		if i < len(cols) {
			out[i] = cols[i]
		}
	}
	return out
}

// --- the deal ---

func TestNarcotic_ResetDealsOneCardToEachOfFourPiles(t *testing.T) {
	g := playingNarcotic(t)

	assert.Equal(t, 4, domain.NarcoticColCnt)
	cols := g.GetColumns()
	for c := range domain.NarcoticColCnt {
		assert.Len(t, cols[c], 1, "column %d", c)
	}
	assert.Equal(t, 48, g.GetStockCount(), "the remaining 48 wait in the stock")
	assert.Equal(t, 0, g.GetDiscardCount())
	assert.Equal(t, 0, g.GetRedealCount())
	assert.Equal(t, domain.NarcoticPhasePlaying, g.GetPhase())
}

func TestNarcotic_DrawDealsOnePerPile(t *testing.T) {
	g := playingNarcotic(t)
	require.NoError(t, g.Draw())

	for c := range domain.NarcoticColCnt {
		assert.Len(t, g.GetColumns()[c], 2, "column %d", c)
	}
	assert.Equal(t, 44, g.GetStockCount())
}

// --- the discard rule ---

// **除去は「露出4枚のランクが揃ったとき、4枚まとめて」だけ。**
// クローン元の Aces Up は「同スートで下位の1枚」を捨てるので、その述語のまま
// クローンすると全く違うゲームになる。
func TestNarcotic_RemovesAllFourWhenTheRanksMatch(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 7)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 7)},
	))

	require.True(t, g.CanRemoveSet())
	require.NoError(t, g.Remove())
	for c := range domain.NarcoticColCnt {
		assert.Empty(t, g.GetColumns()[c], "column %d", c)
	}
	assert.Equal(t, 4, g.GetDiscardCount(), "all four go, not one")
}

// 負のコントロール: 3枚だけ揃っていても捨てられない。
func TestNarcotic_RefusesWhenOnlyThreeMatch(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 7)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 9)},
	))

	assert.False(t, g.CanRemoveSet())
	assert.Error(t, g.Remove())
	assert.Equal(t, 0, g.GetDiscardCount(), "nothing is discarded on a refusal")
}

// **クローン元なら捨てられる盤で、こちらは捨てられない。**
// ♠9 と ♠5 は同スートなので Aces Up では ♠5 が落ちるが、Narcotic は無関係。
func TestNarcotic_IgnoresTheSameSuitLowerRule(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 9)},
		[]*domain.Card{narCard(domain.CardDesignSpade, 5)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 2)},
		[]*domain.Card{narCard(domain.CardDesignClover, 3)},
	))

	assert.False(t, g.CanRemoveSet(), "same suit means nothing here")
	assert.Error(t, g.Remove())
}

// 空の列があると揃えようが無い。
func TestNarcotic_RefusesWhenAPileIsEmpty(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 7)},
		nil,
	))
	assert.False(t, g.CanRemoveSet())
}

// --- the consolidation move ---

// **重複は「同ランクを露出している最も左の列」へ移す。**
// クローン元は「空き列へ移す」ので行き先の決め方がまったく違う。
func TestNarcotic_MovesDuplicateOntoTheLeftmostMatch(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 2)},
		[]*domain.Card{narCard(domain.CardDesignClover, 7)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 3)},
	))

	assert.Equal(t, 0, g.MoveTarget(2), "column 0 is the leftmost 7")
	require.True(t, g.CanMove(2))
	require.NoError(t, g.Move(2))

	assert.Len(t, g.GetColumns()[0], 2, "the ♣7 lands on the ♠7")
	assert.Empty(t, g.GetColumns()[2])
	assert.Equal(t, 0, g.GetDiscardCount(), "a move is not a discard")
}

// 同ランクが3つあるとき、行き先は常に一番左。
func TestNarcotic_AlwaysTargetsTheLeftmostOfSeveralMatches(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 7)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 3)},
	))
	assert.Equal(t, 0, g.MoveTarget(2))
	assert.Equal(t, 0, g.MoveTarget(1))
}

// **右へは動かせない。**動かせると同じ2枚を往復させて手数だけ増やせる。
func TestNarcotic_NeverMovesRightward(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		nil, nil,
	))
	assert.Equal(t, -1, g.MoveTarget(0), "column 0 has nothing to its left")
	assert.False(t, g.CanMove(0))
	assert.Error(t, g.Move(0))
	assert.Equal(t, 0, g.MoveTarget(1))
}

// 負のコントロール: 同ランクがどこにも無ければ動かせない。
func TestNarcotic_RefusesAMoveWithNoMatch(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 9)},
		nil, nil,
	))
	assert.Equal(t, -1, g.MoveTarget(1))
	assert.Error(t, g.Move(1))
}

// --- the redeal ---

// **右の列から左へ順に集め、シャッフルしない。**次の山札が現在の盤面から
// 一意に決まるので、運で変わる要素は無い。回数に上限も無い。
func TestNarcotic_RedealGathersRightToLeftWithoutShuffling(t *testing.T) {
	g := playingNarcotic(t)
	g.SetStock(nil)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 3)},
		[]*domain.Card{narCard(domain.CardDesignClover, 4)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 5)},
	))

	require.NoError(t, g.Redeal())
	assert.Equal(t, 1, g.GetRedealCount())
	assert.Equal(t, 4, g.GetStockCount())
	for c := range domain.NarcoticColCnt {
		assert.Empty(t, g.GetColumns()[c], "the table is cleared into the stock")
	}

	// 右から集めたので、山札の先頭は列3の札。
	require.NoError(t, g.Draw())
	assert.Equal(t, 5, g.GetColumns()[0][0].GetValue(), "column 3's card is dealt first")
	assert.Equal(t, 4, g.GetColumns()[1][0].GetValue())
	assert.Equal(t, 3, g.GetColumns()[2][0].GetValue())
	assert.Equal(t, 2, g.GetColumns()[3][0].GetValue())
}

func TestNarcotic_RedealRefusedWhileStockRemains(t *testing.T) {
	g := playingNarcotic(t)
	require.Positive(t, g.GetStockCount())
	assert.Error(t, g.Redeal(), "the stock must be spent first")
}

func TestNarcotic_RedealHasNoLimit(t *testing.T) {
	g := playingNarcotic(t)
	// 揃わない4枚。何度配り直しても揃わないが、拒まれはしない。
	board := narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 3)},
		[]*domain.Card{narCard(domain.CardDesignClover, 4)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 5)},
	)
	for i := 1; i <= 5; i++ {
		g.SetStock(nil)
		g.SetColumns(board)
		require.NoError(t, g.Redeal(), "redeal %d", i)
		assert.Equal(t, i, g.GetRedealCount())
	}
}

// --- loop detection ---

// **再配りに上限が無いので、これが唯一の終了保証。**同じ盤面が二度現れ、かつ
// 指せる手が無ければ、以後は同じ循環をたどるだけ。
func TestNarcotic_StalemateWhenABoardRepeatsWithNoMove(t *testing.T) {
	g := playingNarcotic(t)
	board := narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 3)},
		[]*domain.Card{narCard(domain.CardDesignClover, 4)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 5)},
	)
	g.SetStock(nil)
	g.SetColumns(board)
	g.CheckNarcoticStalemate()
	require.False(t, g.IsStalemate(), "first sighting is not yet a loop")

	// 同じ盤面をもう一度。
	g.CheckNarcoticStalemate()
	assert.True(t, g.IsStalemate(), "the same board with no move is a loop")
}

// 負のコントロール: 指せる手があれば、盤面が既出でも詰みではない。
func TestNarcotic_NotStalemateWhileAMoveExists(t *testing.T) {
	g := playingNarcotic(t)
	board := narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 4)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 5)},
	)
	g.SetStock(nil)
	g.SetColumns(board)
	for range 3 {
		g.CheckNarcoticStalemate()
		assert.False(t, g.IsStalemate(), "column 1 can still stack onto column 0")
	}
}

// --- end states ---

func TestNarcotic_GameClearWhenTheWholeDeckIsDiscarded(t *testing.T) {
	g := playingNarcotic(t)
	g.SetStock(nil)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 7)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 7)},
	))
	g.SetDiscardCount(48)

	require.NoError(t, g.Remove())
	assert.Equal(t, domain.NarcoticPhaseGameClear, g.GetPhase())
}

func TestNarcotic_GiveUp(t *testing.T) {
	g := playingNarcotic(t)
	g.GiveUp()
	assert.Equal(t, domain.NarcoticPhaseGameOver, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
}

// --- hint ---

func TestNarcotic_HintPrefersTheDiscard(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 7)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 7)},
	))
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "remove", hint.Type)
}

func TestNarcotic_HintSuggestsAMoveThenADraw(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 4)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 5)},
	))
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "move", hint.Type)
	assert.Equal(t, 1, hint.Col)

	// 重ねる手が無ければ配る。
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 3)},
		[]*domain.Card{narCard(domain.CardDesignClover, 4)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 5)},
	))
	assert.Equal(t, "draw", g.GetHint().Type)
}

// 山札が尽きても場に札があれば、勧めるのは再配り。
func TestNarcotic_HintSuggestsARedealWhenTheStockIsSpent(t *testing.T) {
	g := playingNarcotic(t)
	g.SetStock(nil)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 3)},
		[]*domain.Card{narCard(domain.CardDesignClover, 4)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 5)},
	))
	g.CheckNarcoticStalemate()
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "redeal", hint.Type)
}

// --- undo / persistence ---

func TestNarcotic_UndoRestoresAllFourDiscardedCards(t *testing.T) {
	g := playingNarcotic(t)
	g.SetColumns(narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 7)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{narCard(domain.CardDesignClover, 7)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 7)},
	))
	require.NoError(t, g.Remove())
	require.NoError(t, g.Undo())

	for c := range domain.NarcoticColCnt {
		assert.Len(t, g.GetColumns()[c], 1, "column %d gets its card back", c)
	}
	assert.Equal(t, 0, g.GetDiscardCount())
}

func TestNarcotic_UndoRewindsTheRedealCount(t *testing.T) {
	g := playingNarcotic(t)
	g.SetStock(nil)
	g.SetColumns(narBoard([]*domain.Card{narCard(domain.CardDesignSpade, 2)}))
	require.NoError(t, g.Redeal())
	require.Equal(t, 1, g.GetRedealCount())

	require.NoError(t, g.Undo())
	assert.Equal(t, 0, g.GetRedealCount(), "undoing a redeal takes the count back")
}

// **Worker はリクエストごとに KV から作り直す。**ループ検出の記憶が往復しないと、
// 毎回「初めて見る盤面」になり永久に詰まなくなる。
func TestNarcotic_LoopMemorySurvivesTheRoundTrip(t *testing.T) {
	g := playingNarcotic(t)
	board := narBoard(
		[]*domain.Card{narCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{narCard(domain.CardDesignHeart, 3)},
		[]*domain.Card{narCard(domain.CardDesignClover, 4)},
		[]*domain.Card{narCard(domain.CardDesignDiamond, 5)},
	)
	g.SetStock(nil)
	g.SetColumns(board)
	g.CheckNarcoticStalemate()
	require.False(t, g.IsStalemate())

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := domain.NewDefaultNarcotic()
	require.NoError(t, json.Unmarshal(data, restored))

	// 復元後に同じ盤面を見たら、既出として詰みになる。
	restored.CheckNarcoticStalemate()
	assert.True(t, restored.IsStalemate(), "the seen-board memory crossed the wire")
}

func TestNarcotic_JSONRoundTrip(t *testing.T) {
	g := playingNarcotic(t)
	require.NoError(t, g.Draw())

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := domain.NewDefaultNarcotic()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, g.GetRedealCount(), restored.GetRedealCount())
	require.True(t, restored.CanUndo(), "history survives")
	require.NoError(t, restored.Undo())

	assert.Error(t, json.Unmarshal([]byte(`{"rd":-1}`), domain.NewDefaultNarcotic()))
}
