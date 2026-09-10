//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTehonbikiPayoutTableAndLossWithFixedParent(t *testing.T) {
	tests := []struct {
		name    string
		kind    TehonbikiBetType
		numbers []int
		want    int
	}{
		{"single", TehonbikiBetSingle, []int{4}, 225},
		{"double", TehonbikiBetDouble, []int{2, 4}, 90},
		{"triple", TehonbikiBetTriple, []int{1, 4, 6}, 45},
		{"half", TehonbikiBetHalf, []int{3, 4, 5}, 45},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDefaultTehonbiki()
			require.NoError(t, g.SetParentCard(4))
			require.NoError(t, g.PlaceBet(tt.numbers, tt.kind, 50))
			require.Equal(t, TehonbikiResultWin, g.GetResult())
			require.Equal(t, tt.want, g.GetPayout())
		})
	}
	g := NewDefaultTehonbiki()
	require.NoError(t, g.SetParentCard(6))
	require.NoError(t, g.PlaceBet([]int{1, 2}, TehonbikiBetDouble, 50))
	require.Equal(t, TehonbikiResultLose, g.GetResult())
	require.Equal(t, 0, g.GetPayout())
	require.Equal(t, 950, g.GetChips(), "外れた賭け金は没収される")
}

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
