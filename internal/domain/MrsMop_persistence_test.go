//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMrsMop_PersistsUndoHistory verifies issue #1654: when a MrsMop
// game is round-tripped through JSON (e.g. a Cloudflare KV restore), the
// undo history must survive so the player can still step backward.
func TestMrsMop_PersistsUndoHistory(t *testing.T) {
	t.Parallel()

	original := NewDefaultMrsMop()
	original.Reset()
	require.False(t, original.CanUndo(), "fresh game has no history")

	// **Deal は存在しない。**山札が無いので、履歴を作るのは移動だけ。
	// 配りに依存しないよう、盤を明示的に組んでから動かす (配り依存はフレークになる)。
	original.SetTableau(mrsMopMoveableBoard())
	require.NoError(t, original.MoveTableauToTableau(1, 0, 0))
	require.True(t, original.CanUndo(), "a move should leave undoable history")
	originalHistoryLen := len(original.history)
	originalMoveCount := original.GetMoveCount()

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MrsMop
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, originalMoveCount, restored.GetMoveCount(), "moveCount must round-trip")
	assert.Equal(t, originalHistoryLen, len(restored.history), "history length must round-trip")
	assert.True(t, restored.CanUndo(), "restored game must allow Undo")

	require.NoError(t, restored.Undo())
	assert.Equal(t, originalMoveCount-1, restored.GetMoveCount(), "Undo should rewind one step")
}

// mrsMopMoveableBoard は「列1 の ♠Q を列0 の ♠K に重ねられる」だけの決定的な盤。
// 配りに依存しないので、パッケージ全体を回してもフレークにならない。
func mrsMopMoveableBoard() [MrsMopTableauCnt][]*MrsMopTableauCard {
	var board [MrsMopTableauCnt][]*MrsMopTableauCard
	board[0] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax, true), FaceUp: true}}
	board[1] = []*MrsMopTableauCard{{Card: NewCard(CardDesignSpade, CardValueMax-1, true), FaceUp: true}}
	return board
}

// TestMrsMop_PersistsHistoryRestoresExactSnapshot ensures the snapshot
// fields (tableau, completedSuits, score) are preserved exactly.
func TestMrsMop_PersistsHistoryRestoresExactSnapshot(t *testing.T) {
	t.Parallel()

	original := NewDefaultMrsMop()
	original.Reset()
	original.SetTableau(mrsMopMoveableBoard())

	preMoveScore := original.score
	require.NoError(t, original.MoveTableauToTableau(1, 0, 0))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MrsMop
	require.NoError(t, json.Unmarshal(data, &restored))

	require.NoError(t, restored.Undo())
	assert.Equal(t, preMoveScore, restored.score, "Undo restores score")
	assert.Len(t, restored.GetTableau()[0], 1, "the Q goes back off the K")
	assert.Len(t, restored.GetTableau()[1], 1, "and returns to its own column")
}

// TestMrsMop_HistoryRespectsMaxSliceLen rejects oversized history payloads.
func TestMrsMop_HistoryRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHistory := make([]map[string]any, mrsMopMaxSliceLen+1)
	for i := range bigHistory {
		bigHistory[i] = map[string]any{}
	}
	payload := map[string]any{
		"tc": nil,
		"hi": bigHistory,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MrsMop
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized history must be rejected")
}

// TestMrsMop_SnapshotStockRespectsMaxSliceLen rejects payloads that
// smuggle an oversized inner Stock slice inside a history snapshot.
func TestMrsMop_SnapshotTableauColumnRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigCol := make([]map[string]any, mrsMopMaxSliceLen+1)
	for i := range bigCol {
		bigCol[i] = map[string]any{}
	}
	snapshot := map[string]any{
		"tb": []any{bigCol, nil, nil, nil, nil, nil, nil, nil, nil, nil},
	}
	payload := map[string]any{
		"tc": nil,
		"hi": []any{snapshot},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MrsMop
	err = json.Unmarshal(data, &restored)
	require.Error(t, err, "oversized snapshot tableau column must be rejected")
}
