package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDramahaConfig(t *testing.T) {
	cfg := DefaultDramahaConfig()
	assert.Equal(t, 5, cfg.SmallBlind)
	assert.Equal(t, 10, cfg.BigBlind)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, HoldemTableSize4, cfg.TableSize)
}

func TestNewDramahaPlayersForTable(t *testing.T) {
	t.Run("4-max", func(t *testing.T) {
		players := NewDramahaPlayersForTable(HoldemTableSize4)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
		for i := 1; i < 4; i++ {
			assert.False(t, players[i].GetIsHuman())
		}
	})
	t.Run("6-max", func(t *testing.T) {
		players := NewDramahaPlayersForTable(HoldemTableSize6)
		assert.Equal(t, 6, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
	t.Run("9-max", func(t *testing.T) {
		players := NewDramahaPlayersForTable(HoldemTableSize9)
		assert.Equal(t, 9, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
	t.Run("invalid falls back to 4-max", func(t *testing.T) {
		players := NewDramahaPlayersForTable(5)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
}
