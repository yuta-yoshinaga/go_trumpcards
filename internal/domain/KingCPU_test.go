//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKingGetHint_SelectContractCoversEveryContract(t *testing.T) {
	for contract := 0; contract < KingContractCnt; contract++ {
		g := NewDefaultKing()
		g.phase = KingPhaseSelectContract
		g.dealerIdx = 0
		for i := range g.usedContracts {
			g.usedContracts[i] = i != contract
		}
		player := g.players[0]
		player.Reset()
		for i := 0; i < KingHandSize; i++ {
			player.AddCard(NewCard(CardDesignSpade, 2, false))
		}

		hint := g.GetHint()
		require.NotNil(t, hint, "contract %d should be selectable", contract)
		require.Equal(t, contract, hint.Contract)
		require.Equal(t, "select_"+[]string{"no_tricks", "no_hearts", "no_queens", "king_heart", "no_last_two", "no_men", "king_trump"}[contract], hint.Reason)
	}
}
