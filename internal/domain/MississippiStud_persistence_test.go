//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMississippiStud_PersistsMidRound(t *testing.T) {
	t.Parallel()

	original := NewDefaultMississippiStud()
	require.NoError(t, original.Bet(100))
	require.NoError(t, original.Play(3))
	require.NoError(t, original.Play(2))

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MississippiStud
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original.GetPhase(), restored.GetPhase())
	assert.Equal(t, original.GetAnteAmount(), restored.GetAnteAmount())
	assert.Equal(t, original.GetStreetMultipliers(), restored.GetStreetMultipliers())
	assert.Equal(t, original.GetCommunityRevealed(), restored.GetCommunityRevealed())
	assert.Equal(t, original.GetChips(), restored.GetChips())
	assert.Equal(t, original.GetGameEndFlag(), restored.GetGameEndFlag())
	require.Len(t, restored.GetPlayerHand(), MississippiStudHoleCardCnt)
	require.Len(t, restored.GetCommunityCards(), MississippiStudCommunityCnt)
}

func TestMississippiStud_PersistsCompletedRound(t *testing.T) {
	t.Parallel()

	original := NewDefaultMississippiStud()
	require.NoError(t, original.Bet(100))
	require.NoError(t, original.Play(1))
	require.NoError(t, original.Play(1))
	require.NoError(t, original.Play(1))
	require.True(t, original.GetGameEndFlag())

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored MississippiStud
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original.GetResult(), restored.GetResult())
	assert.Equal(t, original.GetHandRank(), restored.GetHandRank())
	assert.Equal(t, original.GetPayoutMultiplier(), restored.GetPayoutMultiplier())
	assert.Equal(t, original.GetAntePayout(), restored.GetAntePayout())
	assert.Equal(t, original.GetStreetPayouts(), restored.GetStreetPayouts())
	assert.Equal(t, original.GetTotalPayout(), restored.GetTotalPayout())
	assert.Equal(t, original.GetChips(), restored.GetChips())
}

func TestMississippiStud_PlayerHandRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigHand := make([]map[string]any, mississippiStudMaxSliceLen+1)
	for i := range bigHand {
		bigHand[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "ph": bigHand}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MississippiStud
	require.Error(t, json.Unmarshal(data, &restored))
}

func TestMississippiStud_CommunityCardsRespectMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigComm := make([]map[string]any, mississippiStudMaxSliceLen+1)
	for i := range bigComm {
		bigComm[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "cc": bigComm}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MississippiStud
	require.Error(t, json.Unmarshal(data, &restored))
}

func TestMississippiStud_ActionLogRespectsMaxSliceLen(t *testing.T) {
	t.Parallel()

	bigLog := make([]map[string]any, mississippiStudMaxSliceLen+1)
	for i := range bigLog {
		bigLog[i] = map[string]any{}
	}
	payload := map[string]any{"tc": nil, "al": bigLog}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MississippiStud
	require.Error(t, json.Unmarshal(data, &restored))
}

func TestMississippiStud_UnmarshalNilTrumpCardsDefaults(t *testing.T) {
	t.Parallel()

	payload := map[string]any{"tc": nil}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored MississippiStud
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.NotNil(t, restored.trumpCards)
}
