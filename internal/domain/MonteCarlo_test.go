package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMonteCarloForTest() *MonteCarlo {
	return NewMonteCarlo(NewTrumpCards(0))
}

func drainStock(g *MonteCarlo) {
	for g.GetStockCount() > 0 {
		_ = g.trumpCards.DrawCard()
	}
}

func TestMonteCarlo_Reset_FillsBoardAndStockHas27(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	assert.Equal(t, MonteCarloPhasePlaying, g.GetPhase())
	assert.Equal(t, 0, g.GetRemovedCount())
	assert.Equal(t, 0, g.GetDealCount())
	assert.Equal(t, 27, g.GetStockCount())
	assert.False(t, g.GetGameEndFlag())
	assert.False(t, g.IsComplete())
	for r := range MonteCarloGridSize {
		for c := range MonteCarloGridSize {
			assert.NotNilf(t, g.GetBoard()[r][c], "cell (%d,%d) should be filled", r, c)
		}
	}
}

func TestMonteCarlo_Reset_ClearsHistoryAndLog(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	board := g.GetBoard()
	board[0][0] = NewCard(0, 7, true)
	board[0][1] = NewCard(1, 7, true)
	g.SetBoard(board)
	require.NoError(t, g.Remove(0, 0, 0, 1))
	assert.True(t, g.CanUndo())
	assert.NotEmpty(t, g.GetActionLog())

	g.Reset()
	assert.False(t, g.CanUndo())
	assert.Empty(t, g.GetActionLog())
}

func TestMonteCarlo_NewDefault_Smoke(t *testing.T) {
	g := NewDefaultMonteCarlo()
	g.Reset()
	assert.Equal(t, 27, g.GetStockCount())
}

func TestMonteCarlo_RemoveSuccess_AdjacentSameRank(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	board := g.GetBoard()
	a := NewCard(0, 7, true)
	b := NewCard(1, 7, true)
	board[0][0] = a
	board[0][1] = b
	g.SetBoard(board)

	require.NoError(t, g.Remove(0, 0, 0, 1))
	assert.Nil(t, g.GetBoard()[0][0])
	assert.Nil(t, g.GetBoard()[0][1])
	assert.Equal(t, 2, g.GetRemovedCount())
	assert.Equal(t, 1, len(g.GetActionLog()))
}

func TestMonteCarlo_RemoveAllAdjacentDirections(t *testing.T) {
	dirs := []struct{ dr, dc int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	for _, d := range dirs {
		g := newMonteCarloForTest()
		g.Reset()
		board := g.GetBoard()
		board[1][1] = NewCard(0, 9, true)
		board[1+d.dr][1+d.dc] = NewCard(1, 9, true)
		g.SetBoard(board)
		assert.NoErrorf(t, g.Remove(1, 1, 1+d.dr, 1+d.dc), "direction (%d,%d)", d.dr, d.dc)
	}
}

func TestMonteCarlo_RemoveErrors(t *testing.T) {
	cases := []struct {
		name           string
		setup          func(g *MonteCarlo)
		r1, c1, r2, c2 int
	}{
		{
			"wrong phase",
			func(g *MonteCarlo) { g.SetPhase(MonteCarloPhaseGameOver) },
			0, 0, 0, 1,
		},
		{"out of bounds (r1 negative)", nil, -1, 0, 0, 1},
		{"out of bounds (c1 too big)", nil, 0, 5, 0, 4},
		{"out of bounds (r2 too big)", nil, 0, 0, 5, 0},
		{"out of bounds (c2 negative)", nil, 0, 0, 0, -1},
		{"same cell", nil, 1, 1, 1, 1},
		{
			"non-adjacent (far row, same rank)",
			func(g *MonteCarlo) {
				board := g.GetBoard()
				board[0][0] = NewCard(0, 5, true)
				board[2][0] = NewCard(1, 5, true)
				g.SetBoard(board)
			}, 0, 0, 2, 0,
		},
		{
			"non-adjacent (far col, same rank)",
			func(g *MonteCarlo) {
				board := g.GetBoard()
				board[0][0] = NewCard(0, 5, true)
				board[0][2] = NewCard(1, 5, true)
				g.SetBoard(board)
			}, 0, 0, 0, 2,
		},
		{
			"empty cell (first)",
			func(g *MonteCarlo) {
				board := g.GetBoard()
				board[0][0] = nil
				g.SetBoard(board)
			}, 0, 0, 0, 1,
		},
		{
			"empty cell (second)",
			func(g *MonteCarlo) {
				board := g.GetBoard()
				board[0][1] = nil
				g.SetBoard(board)
			}, 0, 0, 0, 1,
		},
		{
			"rank mismatch",
			func(g *MonteCarlo) {
				board := g.GetBoard()
				board[0][0] = NewCard(0, 5, true)
				board[0][1] = NewCard(1, 6, true)
				g.SetBoard(board)
			}, 0, 0, 0, 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newMonteCarloForTest()
			g.Reset()
			if tc.setup != nil {
				tc.setup(g)
			}
			assert.Error(t, g.Remove(tc.r1, tc.c1, tc.r2, tc.c2))
		})
	}
}

func TestMonteCarlo_Deal_CompressesAndRefills(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	stockBefore := g.GetStockCount()
	board := g.GetBoard()
	board[0][0] = nil
	board[0][1] = nil
	board[2][3] = nil
	g.SetBoard(board)

	require.NoError(t, g.Deal())
	for r := range MonteCarloGridSize {
		for c := range MonteCarloGridSize {
			assert.NotNilf(t, g.GetBoard()[r][c], "cell (%d,%d)", r, c)
		}
	}
	assert.Equal(t, stockBefore-3, g.GetStockCount())
	assert.Equal(t, 1, g.GetDealCount())
}

func TestMonteCarlo_Deal_WhenStockEmpty_CompressesOnly(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	drainStock(g)
	board := g.GetBoard()
	board[0][0] = nil
	board[1][2] = nil
	g.SetBoard(board)

	require.NoError(t, g.Deal())
	nonNil := 0
	for r := range MonteCarloGridSize {
		for c := range MonteCarloGridSize {
			if g.GetBoard()[r][c] != nil {
				nonNil++
			}
		}
	}
	assert.Equal(t, 23, nonNil)
	assert.Nil(t, g.GetBoard()[4][3])
	assert.Nil(t, g.GetBoard()[4][4])
}

func TestMonteCarlo_Deal_WrongPhase(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	g.SetPhase(MonteCarloPhaseGameClear)
	assert.Error(t, g.Deal())
}

func TestMonteCarlo_GameClear_OnAllCardsRemoved(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	drainStock(g)
	var board [MonteCarloGridSize][MonteCarloGridSize]*Card
	board[0][0] = NewCard(0, 5, true)
	board[0][1] = NewCard(1, 5, true)
	g.SetBoard(board)
	g.SetRemovedCount(50)
	require.NoError(t, g.Remove(0, 0, 0, 1))
	assert.Equal(t, MonteCarloPhaseGameClear, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	assert.True(t, g.IsComplete())
}

func TestMonteCarlo_GiveUp_FromPlaying(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	g.GiveUp()
	assert.Equal(t, MonteCarloPhaseGameOver, g.GetPhase())
}

func TestMonteCarlo_GiveUp_NoOpFromTerminalPhase(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	g.SetPhase(MonteCarloPhaseGameClear)
	g.GiveUp()
	assert.Equal(t, MonteCarloPhaseGameClear, g.GetPhase())
}

func TestMonteCarlo_UndoRemove_RestoresBoard(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	board := g.GetBoard()
	a := NewCard(0, 7, true)
	b := NewCard(1, 7, true)
	board[0][0] = a
	board[0][1] = b
	g.SetBoard(board)
	require.NoError(t, g.Remove(0, 0, 0, 1))
	assert.True(t, g.CanUndo())

	require.NoError(t, g.Undo())
	assert.Equal(t, a, g.GetBoard()[0][0])
	assert.Equal(t, b, g.GetBoard()[0][1])
	assert.Equal(t, 0, g.GetRemovedCount())
	assert.False(t, g.CanUndo())
}

func TestMonteCarlo_UndoDeal_RestoresStockAndBoard(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	stockBefore := g.GetStockCount()
	board := g.GetBoard()
	board[2][2] = nil
	g.SetBoard(board)

	require.NoError(t, g.Deal())
	require.NoError(t, g.Undo())
	assert.Equal(t, stockBefore, g.GetStockCount())
	assert.Nil(t, g.GetBoard()[2][2])
	assert.Equal(t, 0, g.GetDealCount())
}

func TestMonteCarlo_Undo_FailsOnEmptyHistory(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	assert.False(t, g.CanUndo())
	assert.Error(t, g.Undo())
}

func TestMonteCarlo_CanUndo_FalseAfterGameEnd(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	board := g.GetBoard()
	board[0][0] = NewCard(0, 7, true)
	board[0][1] = NewCard(1, 7, true)
	g.SetBoard(board)
	require.NoError(t, g.Remove(0, 0, 0, 1))
	g.SetPhase(MonteCarloPhaseGameOver)
	assert.False(t, g.CanUndo())
}

func TestMonteCarlo_Hint_FindsAdjacentPair(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	var board [MonteCarloGridSize][MonteCarloGridSize]*Card
	board[2][3] = NewCard(0, 9, true)
	board[3][4] = NewCard(1, 9, true)
	g.SetBoard(board)

	h := g.Hint()
	require.NotNil(t, h)
	assert.Equal(t, MonteCarloHintActionRemove, h.Action)
	pair1 := h.FromR == 2 && h.FromC == 3 && h.ToR == 3 && h.ToC == 4
	pair2 := h.FromR == 3 && h.FromC == 4 && h.ToR == 2 && h.ToC == 3
	assert.True(t, pair1 || pair2, "hint pair should reference (2,3) and (3,4)")
}

func TestMonteCarlo_Hint_FallsBackToDeal(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	var board [MonteCarloGridSize][MonteCarloGridSize]*Card
	board[0][0] = NewCard(0, 1, true)
	board[4][4] = NewCard(1, 13, true)
	g.SetBoard(board)
	h := g.Hint()
	require.NotNil(t, h)
	assert.Equal(t, MonteCarloHintActionDeal, h.Action)
}

func TestMonteCarlo_Hint_NilWhenStalemateAndStockEmpty(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	drainStock(g)
	var board [MonteCarloGridSize][MonteCarloGridSize]*Card
	board[0][0] = NewCard(0, 1, true)
	board[0][1] = NewCard(1, 13, true)
	g.SetBoard(board)
	assert.Nil(t, g.Hint())
}

func TestMonteCarlo_Hint_NilWhenGameEnded(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	g.SetPhase(MonteCarloPhaseGameOver)
	assert.Nil(t, g.Hint())
}

func TestMonteCarlo_Stalemate_DetectedWhenNoPairAndStockEmpty(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	drainStock(g)
	var board [MonteCarloGridSize][MonteCarloGridSize]*Card
	// Two non-pairable cards in the leading prefix of row-major order
	// (no compression gap).
	board[0][0] = NewCard(0, 1, true)
	board[0][1] = NewCard(1, 13, true)
	g.SetBoard(board)
	g.CheckMonteCarloStalemate()
	assert.True(t, g.IsStalemate())
}

func TestMonteCarlo_Stalemate_NotDetectedWhenStockHasCards(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	var board [MonteCarloGridSize][MonteCarloGridSize]*Card
	board[0][0] = NewCard(0, 1, true)
	g.SetBoard(board)
	g.CheckMonteCarloStalemate()
	assert.False(t, g.IsStalemate())
}

func TestMonteCarlo_Stalemate_NotDetectedWhenCompressionGap(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	drainStock(g)
	var board [MonteCarloGridSize][MonteCarloGridSize]*Card
	// Gap: (0,0) is nil but (0,1) and (4,4) are filled — Deal compression rearranges them.
	board[0][1] = NewCard(0, 1, true)
	board[4][4] = NewCard(1, 13, true)
	g.SetBoard(board)
	g.CheckMonteCarloStalemate()
	assert.False(t, g.IsStalemate())
}

func TestMonteCarlo_JSON_RoundTrip(t *testing.T) {
	g := newMonteCarloForTest()
	g.Reset()
	board := g.GetBoard()
	board[1][1] = NewCard(0, 7, true)
	board[1][2] = NewCard(1, 7, true)
	g.SetBoard(board)
	require.NoError(t, g.Remove(1, 1, 1, 2))

	data, err := json.Marshal(g)
	require.NoError(t, err)

	g2 := newMonteCarloForTest()
	require.NoError(t, json.Unmarshal(data, g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetRemovedCount(), g2.GetRemovedCount())
	assert.Equal(t, g.GetDealCount(), g2.GetDealCount())
	assert.Equal(t, g.GetStockCount(), g2.GetStockCount())
	assert.Equal(t, g.IsStalemate(), g2.IsStalemate())
}

func TestMonteCarlo_JSON_UnmarshalRejectsHugeActionLog(t *testing.T) {
	huge := make([]map[string]any, monteCarloMaxSliceLen+1)
	for i := range huge {
		huge[i] = map[string]any{}
	}
	wire := map[string]any{
		"tc": nil,
		"al": huge,
	}
	data, err := json.Marshal(wire)
	require.NoError(t, err)
	g := newMonteCarloForTest()
	assert.Error(t, json.Unmarshal(data, g))
}

// #5587: 取り除ける組の数はこのゲームの判断材料そのもの。ヒント・手詰まり判定と
// **同じ走査**から出ること — 規則を 3 箇所に置くと、片方だけ直したときに
// 「取れる組があるのに手詰まり」になる。
func TestMonteCarlo_CountRemovablePairs(t *testing.T) {
	m := NewDefaultMonteCarlo()
	m.Reset()

	board := m.GetBoard()
	// 盤面を作り直す: 同ランクを隣接させた 2 組と、離れた同ランク 1 組。
	for r := range board {
		for c := range board[r] {
			board[r][c] = nil
		}
	}
	board[0][0] = NewCard(CardDesignSpade, 5, true)
	board[0][1] = NewCard(CardDesignHeart, 5, true)   // 右隣: 1 組
	board[1][0] = NewCard(CardDesignClover, 5, true)  // 下と右下: さらに 2 組
	board[4][4] = NewCard(CardDesignDiamond, 9, true) // 離れた同ランク
	board[0][3] = NewCard(CardDesignSpade, 9, true)
	m.SetBoard(board)

	// (0,0)-(0,1), (0,0)-(1,0), (0,1)-(1,0) の 3 組。9 は隣接していないので数えない。
	assert.Equal(t, 3, m.CountRemovablePairs())

	// **ヒントと矛盾しないこと。**組があるならヒントは取り除きを勧める。
	hint := m.Hint()
	if assert.NotNil(t, hint) {
		assert.Equal(t, MonteCarloHintActionRemove, hint.Action)
	}
}

// 0 組の盤面ではヒントが取り除きを勧めないこと。数え方と判定が食い違わない。
func TestMonteCarlo_CountRemovablePairsAgreesWithTheHint(t *testing.T) {
	m := NewDefaultMonteCarlo()
	m.Reset()

	board := m.GetBoard()
	for r := range board {
		for c := range board[r] {
			board[r][c] = nil
		}
	}
	board[0][0] = NewCard(CardDesignSpade, 5, true)
	board[2][2] = NewCard(CardDesignHeart, 5, true) // 隣接していない
	m.SetBoard(board)

	assert.Zero(t, m.CountRemovablePairs())
	if hint := m.Hint(); hint != nil {
		assert.NotEqual(t, MonteCarloHintActionRemove, hint.Action,
			"nothing is removable, so the hint must not suggest a removal")
	}
}

// ヒントは走査順で**最初の**組を返す。共有の走査に早期終了が無いと最後の組に
// 変わり、盤面が同じでも案内が動く。
func TestMonteCarlo_HintReturnsTheFirstPairInScanOrder(t *testing.T) {
	m := NewDefaultMonteCarlo()
	m.Reset()

	board := m.GetBoard()
	for r := range board {
		for c := range board[r] {
			board[r][c] = nil
		}
	}
	// 走査は行優先。(0,0)-(0,1) が最初、(3,3)-(3,4) が後。
	board[0][0] = NewCard(CardDesignSpade, 4, true)
	board[0][1] = NewCard(CardDesignHeart, 4, true)
	board[3][3] = NewCard(CardDesignClover, 9, true)
	board[3][4] = NewCard(CardDesignDiamond, 9, true)
	m.SetBoard(board)

	hint := m.Hint()
	require.NotNil(t, hint)
	assert.Equal(t, MonteCarloHintActionRemove, hint.Action)
	assert.Equal(t, 0, hint.FromR)
	assert.Equal(t, 0, hint.FromC)
}
