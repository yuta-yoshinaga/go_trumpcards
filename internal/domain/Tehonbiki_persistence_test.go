//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTehonbikiPersistenceRoundTripAndParentRange(t *testing.T) {
	g := NewDefaultTehonbiki()
	require.NoError(t, g.SetParentCard(5))
	require.NoError(t, g.PlaceBet([]int{5}, TehonbikiBetSingle, 50))
	b, err := json.Marshal(g)
	require.NoError(t, err)
	restored := new(Tehonbiki)
	require.NoError(t, json.Unmarshal(b, restored))
	require.Equal(t, g.GetParentCard(), restored.GetParentCard())
	require.Equal(t, g.GetResult(), restored.GetResult())
	require.Equal(t, g.GetPayout(), restored.GetPayout())
	for _, n := range []int{0, 7} {
		tampered := map[string]any{}
		require.NoError(t, json.Unmarshal(b, &tampered))
		tampered["pc"] = n
		bad, marshalErr := json.Marshal(tampered)
		require.NoError(t, marshalErr)
		require.Error(t, json.Unmarshal(bad, new(Tehonbiki)))
	}
}
