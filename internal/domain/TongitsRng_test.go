//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tongits uses its injected RNG for CPU choices, while Player.ShuffleCards uses
// the global RNG. This test fixes only the CPU choice and does not assert deal order.
func TestTongits_SetRandControlsEasyCpuDiscardChoice(t *testing.T) {
	tests := []struct {
		name            string
		randValue       int64
		wantDiscardSize int
		wantDrawSize    int
	}{
		{name: "choose discard", randValue: 0, wantDiscardSize: 0, wantDrawSize: 1},
		{name: "choose stock", randValue: 1 << 62, wantDiscardSize: 1, wantDrawSize: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			players := []*TongitsPlayer{NewTongitsPlayer(true), NewTongitsPlayer(false), NewTongitsPlayer(false)}
			cfg := DefaultTongitsConfig()
			cfg.CpuDifficulty = TongitsCpuDifficultyEasy
			g := NewTongits(NewTrumpCards(0), players, cfg)
			g.SetPhase(TongitsPhaseDraw)
			g.SetCurrentPlayerIdx(1)
			g.SetRand(rand.New(&tongitsFixedRandSource{value: tt.randValue}))
			players[1].AddCard(NewCard(CardDesignSpade, 9, false))
			g.SetDiscardPile([]*Card{NewCard(CardDesignHeart, 2, false)})
			g.SetDrawPile([]*Card{NewCard(CardDesignClover, 3, false)})

			g.CpuPlay()

			assert.Len(t, g.GetDiscardPile(), tt.wantDiscardSize)
			assert.Equal(t, tt.wantDrawSize, g.GetDrawPileCount())
			assert.Equal(t, TongitsPhaseDiscard, g.GetPhase())
		})
	}
}

// tongitsFixedRandSource makes rand.Intn(3) return a predictable value.
type tongitsFixedRandSource struct{ value int64 }

func (s *tongitsFixedRandSource) Int63() int64 { return s.value }
func (s *tongitsFixedRandSource) Seed(int64)   {}

// A restored game must recreate its RNG because the worker reconstructs games from JSON.
func TestTongits_RestoredGameSurvivesEasyCpu(t *testing.T) {
	src := NewDefaultTongits()
	src.Reset()
	cfg := src.GetConfig()
	cfg.CpuDifficulty = TongitsCpuDifficultyEasy
	src.SetConfig(cfg)

	data, err := src.MarshalJSON()
	require.NoError(t, err)

	var restored Tongits
	require.NoError(t, restored.UnmarshalJSON(data))
	require.Equal(t, TongitsCpuDifficultyEasy, restored.GetConfig().CpuDifficulty)

	for i := 0; i < restored.GetPlayerCnt(); i++ {
		if !restored.GetPlayer(i).GetIsHuman() {
			restored.SetCurrentPlayerIdx(i)
			break
		}
	}

	assert.NotPanics(t, func() { restored.CpuPlay() })
}
