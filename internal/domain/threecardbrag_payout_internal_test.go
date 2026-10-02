//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThreeCardBragRoundPayoutsIncludeRemainder(t *testing.T) {
	g := NewDefaultThreeCardBrag()
	g.SetPot(11)
	g.endDeal([]int{0, 1})
	assert.Equal(t, []int{6, 5, 0, 0}, g.GetRoundPayouts())
}

func TestThreeCardBragUnmarshalLegacySnapshotInitializesRoundPayouts(t *testing.T) {
	g := NewDefaultThreeCardBrag()
	g.Reset()
	snapshot, err := json.Marshal(g)
	assert.NoError(t, err)
	var fields map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(snapshot, &fields))
	delete(fields, "rp")
	legacySnapshot, err := json.Marshal(fields)
	assert.NoError(t, err)

	var restored ThreeCardBrag
	assert.NoError(t, json.Unmarshal(legacySnapshot, &restored))
	assert.Equal(t, make([]int, ThreeCardBragPlayerCnt), restored.GetRoundPayouts())
}

func TestThreeCardBragUnmarshalRejectsTooManyRoundPayouts(t *testing.T) {
	g := NewDefaultThreeCardBrag()
	g.Reset()
	snapshot, err := json.Marshal(g)
	assert.NoError(t, err)
	var fields map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(snapshot, &fields))
	fields["rp"] = json.RawMessage(`[0,0,0,0,0]`)
	invalidSnapshot, err := json.Marshal(fields)
	assert.NoError(t, err)

	var restored ThreeCardBrag
	assert.Error(t, json.Unmarshal(invalidSnapshot, &restored))
}
