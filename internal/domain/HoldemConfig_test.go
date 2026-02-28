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
}

func TestHoldemPlayStyle_Constants(t *testing.T) {
	assert.Equal(t, HoldemPlayStyle(0), HoldemStyleTAG)
	assert.Equal(t, HoldemPlayStyle(1), HoldemStyleLAP)
	assert.Equal(t, HoldemPlayStyle(2), HoldemStyleTAP)
	assert.Equal(t, HoldemPlayStyle(3), HoldemStyleLAG)
}

func TestHoldemPlayStyleNames(t *testing.T) {
	assert.Equal(t, "TAG", HoldemPlayStyleNames[HoldemStyleTAG])
	assert.Equal(t, "LAP", HoldemPlayStyleNames[HoldemStyleLAP])
	assert.Equal(t, "TAP", HoldemPlayStyleNames[HoldemStyleTAP])
	assert.Equal(t, "LAG", HoldemPlayStyleNames[HoldemStyleLAG])
}
