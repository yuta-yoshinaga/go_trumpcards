//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnsunKaruta_MeriDeclarationActionLog(t *testing.T) {
	game := NewDefaultUnsunKaruta()
	game.Reset()
	game.currentPlayerIdx = 0
	game.phase = UnsunKarutaPhasePlay
	card := game.players[0].GetCard(0)
	game.trumpSuit = card.GetDesign()

	require.NoError(t, game.PlayerPlay(0, true))

	var declarationEntry *ActionLogEntry
	for _, entry := range game.GetActionLog() {
		if entry.ActionType == "declare" {
			declarationEntry = entry
			break
		}
	}
	require.NotNil(t, declarationEntry)
	assert.Equal(t, "unsunkaruta.log.declareMeri", declarationEntry.DetailCode)
	assert.Equal(t, map[string]string{"name": "You"}, declarationEntry.DetailParams)
	assert.NotContains(t, declarationEntry.DetailParams, "declaration")
	assert.Empty(t, declarationEntry.Detail)
}
