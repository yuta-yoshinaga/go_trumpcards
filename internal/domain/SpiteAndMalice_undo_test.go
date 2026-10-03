//go:build test

package domain_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSpiteAndMaliceUndoDiscardAndJSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultSpiteAndMalice()
	g.Reset()
	before, err := json.Marshal(g)
	require.NoError(t, err)
	require.NoError(t, g.Discard(0, 0))
	assert.Equal(t, domain.SpiteAndMaliceCpuIdx, g.GetCurrent())
	assert.True(t, g.CanUndo())
	g.SetPlayerGoal(domain.SpiteAndMaliceCpuIdx, nil)
	for side := range domain.SpiteAndMaliceSideCnt {
		g.SetPlayerSide(domain.SpiteAndMaliceCpuIdx, side, nil)
	}
	g.SetPlayerHand(domain.SpiteAndMaliceCpuIdx, []*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, true)})
	require.NoError(t, g.CpuStep())
	assert.Equal(t, domain.SpiteAndMaliceHumanIdx, g.GetCurrent(), "the CPU's discard should finish its turn")
	assert.True(t, g.CanUndo(), "CPU actions must not replace the human snapshot")

	data, err := json.Marshal(g)
	require.NoError(t, err)
	restored := domain.NewDefaultSpiteAndMalice()
	require.NoError(t, json.Unmarshal(data, restored))
	assert.True(t, restored.CanUndo())
	require.NoError(t, restored.Undo())
	after, err := json.Marshal(restored)
	require.NoError(t, err)
	canonical := domain.NewDefaultSpiteAndMalice()
	require.NoError(t, json.Unmarshal(before, canonical))
	normalized, err := json.Marshal(canonical)
	require.NoError(t, err)
	var want, got map[string]any
	require.NoError(t, json.Unmarshal(normalized, &want))
	require.NoError(t, json.Unmarshal(after, &got))
	delete(want, "hi")
	delete(got, "hi")
	assert.Equal(t, want, got)
}

func TestSpiteAndMaliceUndoHistoryBoundAndGameOver(t *testing.T) {
	g := domain.NewDefaultSpiteAndMalice()
	g.Reset()
	err := g.Undo()
	require.ErrorIs(t, err, domain.ErrInvalidPlay)
	assert.Equal(t, "spiteandmalice.errNothingToUndo", err.(*domain.DomainError).MessageCode())
	for range domain.MaxUndoHistory + 2 {
		g.SetCurrent(domain.SpiteAndMaliceHumanIdx)
		g.SetPlayerHand(domain.SpiteAndMaliceHumanIdx, []*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, true)})
		require.NoError(t, g.Discard(0, 0))
	}
	b, err := json.Marshal(g)
	require.NoError(t, err)
	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(b, &payload))
	var history []json.RawMessage
	require.NoError(t, json.Unmarshal(payload["hi"], &history))
	assert.Len(t, history, domain.MaxUndoHistory)
	g.SetPhase(domain.SpiteAndMalicePhaseGameOver)
	assert.False(t, g.CanUndo())
	err = g.Undo()
	require.ErrorIs(t, err, domain.ErrInvalidPlay)
	assert.Equal(t, "spiteandmalice.errNothingToUndo", err.(*domain.DomainError).MessageCode())
}

func TestSpiteAndMaliceRejectsNilUndoHistoryEntry(t *testing.T) {
	g := domain.NewDefaultSpiteAndMalice()
	err := json.Unmarshal([]byte(`{"hi":[null]}`), g)
	require.Error(t, err)
}

func TestSpiteAndMaliceUndoMalformedSnapshotLeavesBoardUnchanged(t *testing.T) {
	g := domain.NewDefaultSpiteAndMalice()
	g.Reset()
	require.NoError(t, g.Discard(0, 0))
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var state map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &state))
	var history []map[string]string
	require.NoError(t, json.Unmarshal(state["hi"], &history))
	history[len(history)-1]["st"] = base64.StdEncoding.EncodeToString([]byte("{"))
	state["hi"], err = json.Marshal(history)
	require.NoError(t, err)
	data, err = json.Marshal(state)
	require.NoError(t, err)
	corrupt := domain.NewDefaultSpiteAndMalice()
	require.NoError(t, json.Unmarshal(data, corrupt))
	currentBefore := corrupt.GetCurrent()
	phaseBefore := corrupt.GetPhase()
	handBefore := corrupt.GetPlayer(domain.SpiteAndMaliceHumanIdx).HandSize()
	actionLogBefore := corrupt.GetActionLog()

	require.Error(t, corrupt.Undo())
	assert.Equal(t, currentBefore, corrupt.GetCurrent())
	assert.Equal(t, phaseBefore, corrupt.GetPhase())
	assert.Equal(t, handBefore, corrupt.GetPlayer(domain.SpiteAndMaliceHumanIdx).HandSize())
	assert.Equal(t, actionLogBefore, corrupt.GetActionLog())
}

func TestSpiteAndMaliceUndoRemovesExactlyOneHistoryEntry(t *testing.T) {
	g := domain.NewDefaultSpiteAndMalice()
	g.Reset()
	for i := 0; i < 2; i++ {
		g.SetCurrent(domain.SpiteAndMaliceHumanIdx)
		g.SetPlayerHand(domain.SpiteAndMaliceHumanIdx, []*domain.Card{domain.NewCard(domain.CardDesignHeart, 2+i, true)})
		require.NoError(t, g.Discard(0, 0))
	}
	before, err := json.Marshal(g)
	require.NoError(t, err)
	var beforeState struct {
		History []json.RawMessage `json:"hi"`
	}
	require.NoError(t, json.Unmarshal(before, &beforeState))
	require.NoError(t, g.Undo())
	after, err := json.Marshal(g)
	require.NoError(t, err)
	var afterState struct {
		History []json.RawMessage `json:"hi"`
	}
	require.NoError(t, json.Unmarshal(after, &afterState))
	assert.Len(t, afterState.History, len(beforeState.History)-1)
}

func TestSpiteAndMaliceAutoCompleteIsOneUndoableHumanAction(t *testing.T) {
	g := domain.NewDefaultSpiteAndMalice()
	g.Reset()
	g.SetFoundation(0, nil)
	g.SetStock(nil)
	g.SetPlayerHand(domain.SpiteAndMaliceHumanIdx, nil)
	for side := range domain.SpiteAndMaliceSideCnt {
		g.SetPlayerSide(domain.SpiteAndMaliceHumanIdx, side, nil)
	}
	// SetPlayerGoal uses the final card as the top. The ace plays onto an empty
	// foundation, leaving the five underneath so Undo remains available.
	g.SetPlayerGoal(domain.SpiteAndMaliceHumanIdx, []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, true),
		domain.NewCard(domain.CardDesignHeart, 1, true),
	})
	require.NoError(t, g.AutoComplete())
	assert.Equal(t, 1, g.GetFoundationTopValue(0))
	assert.Equal(t, domain.SpiteAndMalicePhasePlaying, g.GetPhase())
	assert.Equal(t, 1, g.GetPlayer(domain.SpiteAndMaliceHumanIdx).GoalSize())
	assert.Equal(t, 5, g.GetPlayer(domain.SpiteAndMaliceHumanIdx).GoalTop().GetValue())
	require.NoError(t, g.Undo())
	assert.Equal(t, 0, g.GetFoundationTopValue(0))
	assert.Equal(t, 2, g.GetPlayer(domain.SpiteAndMaliceHumanIdx).GoalSize())
	assert.Equal(t, 1, g.GetPlayer(domain.SpiteAndMaliceHumanIdx).GoalTop().GetValue())
}
