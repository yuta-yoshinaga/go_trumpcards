//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrenteEtQuaranteSessionStatsAreRecordedBySettlement(t *testing.T) {
	g := NewTrenteEtQuarante(nil, NewTrenteEtQuarantePlayer(1000), DefaultTrenteEtQuaranteConfig())
	settle := func(noir, rouge int, bet TrenteEtQuaranteBet) {
		g.state.phase = TrenteEtQuarantePhaseBet
		g.state.currentBet, g.state.stake = bet, 100
		g.state.noirTotal, g.state.rougeTotal = noir, rouge
		g.state.winningRow, g.state.refait = TrenteEtQuaranteRowNone, false
		require.True(t, g.player.SubtractChips(100))
		g.resolve()
	}
	settle(31, 39, TrenteEtQuaranteBetNoir) // win
	settle(39, 31, TrenteEtQuaranteBetNoir) // loss
	settle(32, 32, TrenteEtQuaranteBetNoir) // push
	settle(31, 31, TrenteEtQuaranteBetNoir) // Refait
	require.Equal(t, 1, g.GetWins())
	require.Equal(t, 1, g.GetLosses())
	require.Equal(t, 1, g.GetDraws())
	require.Equal(t, 1, g.GetRefaits())
	require.Equal(t, 1000, g.GetStartingChips())
	require.Equal(t, g.GetChips()-g.GetStartingChips(), g.GetNet())
	require.Equal(t, -50, g.GetNet())
	require.Equal(t, 4, g.GetRoundNumber())

	g.NextRound()
	require.Equal(t, 1, g.GetWins())
	require.Equal(t, 1, g.GetLosses())
	require.Equal(t, 1, g.GetDraws())
	require.Equal(t, 1, g.GetRefaits())
	require.Equal(t, -50, g.GetNet())
	require.Equal(t, 4, g.GetRoundNumber())

	settle(31, 39, TrenteEtQuaranteBetNoir) // next round win
	require.Equal(t, 2, g.GetWins())
	require.Equal(t, 1, g.GetLosses())
	require.Equal(t, 1, g.GetDraws())
	require.Equal(t, 1, g.GetRefaits())
	require.Equal(t, 50, g.GetNet())
	require.Equal(t, 5, g.GetRoundNumber())
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var restored TrenteEtQuarante
	require.NoError(t, json.Unmarshal(data, &restored))
	require.Equal(t, 1, restored.GetRefaits())
	require.Equal(t, 50, restored.GetNet())

	g.Reset()
	require.Equal(t, 0, g.GetWins())
	require.Equal(t, 0, g.GetLosses())
	require.Equal(t, 0, g.GetDraws())
	require.Equal(t, 0, g.GetRefaits())
	require.Equal(t, g.GetChips(), g.GetStartingChips())
}
