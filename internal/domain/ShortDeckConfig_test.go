package domain_test

import (
	"testing"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestDefaultShortDeckConfig(t *testing.T) {
	cfg := domain.DefaultShortDeckConfig()
	assert.Equal(t, 5, cfg.SmallBlind)
	assert.Equal(t, 10, cfg.BigBlind)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, domain.HoldemTableSize4, cfg.TableSize)
}

func TestNewShortDeckPlayersForTable(t *testing.T) {
	t.Run("4-max", func(t *testing.T) {
		players := domain.NewShortDeckPlayersForTable(domain.HoldemTableSize4)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
		for i := 1; i < 4; i++ {
			assert.False(t, players[i].GetIsHuman())
		}
	})
	t.Run("6-max", func(t *testing.T) {
		players := domain.NewShortDeckPlayersForTable(domain.HoldemTableSize6)
		assert.Equal(t, 6, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
	t.Run("9-max", func(t *testing.T) {
		players := domain.NewShortDeckPlayersForTable(domain.HoldemTableSize9)
		assert.Equal(t, 9, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
	t.Run("invalid falls back to 4-max", func(t *testing.T) {
		players := domain.NewShortDeckPlayersForTable(5)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
}
