package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChipHolder_GetChips(t *testing.T) {
	t.Run("initial value is 0", func(t *testing.T) {
		ch := ChipHolder{}
		assert.Equal(t, 0, ch.GetChips())
	})
}

func TestChipHolder_SetChips(t *testing.T) {
	t.Run("sets chips to given value", func(t *testing.T) {
		ch := ChipHolder{}
		ch.SetChips(1000)
		assert.Equal(t, 1000, ch.GetChips())
	})
}

func TestChipHolder_AddChips(t *testing.T) {
	t.Run("adds amount to current chips", func(t *testing.T) {
		ch := ChipHolder{}
		ch.SetChips(500)
		ch.AddChips(300)
		assert.Equal(t, 800, ch.GetChips())
	})
}

func TestChipHolder_SubtractChips(t *testing.T) {
	t.Run("sufficient chips returns true and subtracts", func(t *testing.T) {
		ch := ChipHolder{}
		ch.SetChips(1000)
		ok := ch.SubtractChips(300)
		assert.True(t, ok)
		assert.Equal(t, 700, ch.GetChips())
	})

	t.Run("exact chips returns true and zeroes out", func(t *testing.T) {
		ch := ChipHolder{}
		ch.SetChips(500)
		ok := ch.SubtractChips(500)
		assert.True(t, ok)
		assert.Equal(t, 0, ch.GetChips())
	})

	t.Run("insufficient chips returns false and unchanged", func(t *testing.T) {
		ch := ChipHolder{}
		ch.SetChips(100)
		ok := ch.SubtractChips(200)
		assert.False(t, ok)
		assert.Equal(t, 100, ch.GetChips())
	})
}
