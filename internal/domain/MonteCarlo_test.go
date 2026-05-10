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
