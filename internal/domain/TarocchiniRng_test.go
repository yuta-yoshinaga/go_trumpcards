//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **復元したゲームで Easy の CPU を動かす。**Worker は毎リクエスト KV から
// 組み直すので `SetRand` は一度も呼ばれない。`UnmarshalJSON` が rng を戻さない
// と、Easy の分岐が nil の *rand.Rand を触って落ちる。
//
// KV 往復テスト (TarocchiniKV_test.go) はこれを隠す —— 往復のたびに手で
// SetRand を呼んでいるため。本番の復元経路には無い一行なので、ここでは呼ばない。
func TestTarocchini_RestoredGameSurvivesEasyCpu(t *testing.T) {
	src := NewDefaultTarocchini()
	src.Reset()
	cfg := src.GetConfig()
	cfg.CpuDifficulty = TarocchiniCpuDifficultyEasy
	src.SetConfig(cfg)
	src.CpuScarto()

	data, err := src.MarshalJSON()
	require.NoError(t, err)

	var restored Tarocchini
	require.NoError(t, restored.UnmarshalJSON(data))
	require.Equal(t, TarocchiniCpuDifficultyEasy, restored.GetConfig().CpuDifficulty)

	// 復元直後に CPU の手番を回す。SetRand は呼ばない。
	restored.SetPhase(TarocchiniPhasePlay)
	restored.SetCurrentPlayerIdx(1)
	assert.NotPanics(t, func() { restored.CpuPlay() },
		"a restored game must not depend on SetRand having been called")
	assert.Len(t, restored.GetCurrentTrick(), 1, "the CPU should have played a card")
}
