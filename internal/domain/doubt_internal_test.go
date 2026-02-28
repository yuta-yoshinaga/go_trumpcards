package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDecideCpuDoubters_NilLastAction lastAction が nil のとき早期リターンする
func TestDecideCpuDoubters_NilLastAction(t *testing.T) {
	players := []*DoubtPlayer{
		NewDoubtPlayer(true),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
	}
	d := NewDoubt(NewTrumpCards(0), players)
	// lastAction is nil (initial state)
	d.decideCpuDoubters()
	// Should be nil (early return)
	assert.Nil(t, d.cpuDoubters)
}

// TestMemoryRetentionChance_UnknownLevel 未知のレベルは Normal と同じ保持率を返す
func TestMemoryRetentionChance_UnknownLevel(t *testing.T) {
	unknown := DoubtMemoryLevel(99)
	assert.Equal(t, retentionChanceNormal, memoryRetentionChance(unknown))
}

// TestCheckLying_NilLastAction lastAction が nil のとき false を返す
func TestCheckLying_NilLastAction(t *testing.T) {
	players := []*DoubtPlayer{
		NewDoubtPlayer(true),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
	}
	d := NewDoubt(NewTrumpCards(0), players)
	// lastAction is nil (initial state)
	result := d.checkLying()
	assert.False(t, result)
}
