package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHoldemConfig(t *testing.T) {
	cfg := DefaultHoldemConfig()
	assert.Equal(t, 5, cfg.SmallBlind)
	assert.Equal(t, 10, cfg.BigBlind)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, HoldemTableSize4, cfg.TableSize)
}

func TestHoldemTableSizeConstants(t *testing.T) {
	assert.Equal(t, 4, HoldemTableSize4)
	assert.Equal(t, 6, HoldemTableSize6)
	assert.Equal(t, 9, HoldemTableSize9)
}

func TestValidHoldemTableSizes(t *testing.T) {
	assert.True(t, ValidHoldemTableSizes[4])
	assert.True(t, ValidHoldemTableSizes[6])
	assert.True(t, ValidHoldemTableSizes[9])
	assert.False(t, ValidHoldemTableSizes[2])
	assert.False(t, ValidHoldemTableSizes[5])
	assert.False(t, ValidHoldemTableSizes[10])
}

func TestDefaultCpuStyles(t *testing.T) {
	t.Run("4-max", func(t *testing.T) {
		styles := DefaultCpuStyles(HoldemTableSize4)
		assert.Equal(t, 3, len(styles))
		assert.Equal(t, HoldemStyleLAP, styles[0])
		assert.Equal(t, HoldemStyleTAP, styles[1])
		assert.Equal(t, HoldemStyleGTO, styles[2])
	})
	t.Run("6-max", func(t *testing.T) {
		styles := DefaultCpuStyles(HoldemTableSize6)
		assert.Equal(t, 5, len(styles))
		assert.Equal(t, HoldemStyleTAG, styles[0])
		assert.Equal(t, HoldemStyleLAP, styles[1])
		assert.Equal(t, HoldemStyleTAP, styles[2])
		assert.Equal(t, HoldemStyleLAG, styles[3])
		assert.Equal(t, HoldemStyleGTO, styles[4])
	})
	t.Run("9-max", func(t *testing.T) {
		styles := DefaultCpuStyles(HoldemTableSize9)
		assert.Equal(t, 8, len(styles))
	})
	t.Run("unknown defaults to 4-max", func(t *testing.T) {
		styles := DefaultCpuStyles(3)
		assert.Equal(t, 3, len(styles))
	})
}

func TestHoldemPlayStyle_Constants(t *testing.T) {
	assert.Equal(t, HoldemPlayStyle(0), HoldemStyleTAG)
	assert.Equal(t, HoldemPlayStyle(1), HoldemStyleLAP)
	assert.Equal(t, HoldemPlayStyle(2), HoldemStyleTAP)
	assert.Equal(t, HoldemPlayStyle(3), HoldemStyleLAG)
	assert.Equal(t, HoldemPlayStyle(4), HoldemStyleGTO)
}

func TestHoldemPlayStyleNames(t *testing.T) {
	assert.Equal(t, "TAG", HoldemPlayStyleNames[HoldemStyleTAG])
	assert.Equal(t, "LAP", HoldemPlayStyleNames[HoldemStyleLAP])
	assert.Equal(t, "TAP", HoldemPlayStyleNames[HoldemStyleTAP])
	assert.Equal(t, "LAG", HoldemPlayStyleNames[HoldemStyleLAG])
	assert.Equal(t, "GTO", HoldemPlayStyleNames[HoldemStyleGTO])
}
