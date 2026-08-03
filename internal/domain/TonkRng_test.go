//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **復元したゲームで Easy の CPU を動かす。**Cloudflare Worker は毎リクエスト
// KV から組み直すので `SetRand` は一度も呼ばれない。`UnmarshalJSON` が rng を
// 戻さないと、`cpuDraw` の Easy 分岐 (`g.rng.Intn(3)`) が nil の *rand.Rand を
// 触って落ちる —— サーバー実行ではインタラクタがメモリに残るため再現しない。
func TestTonk_RestoredGameSurvivesEasyCpu(t *testing.T) {
	src := NewDefaultTonk()
	src.Reset()
	cfg := src.GetConfig()
	cfg.CpuDifficulty = TonkCpuDifficultyEasy
	src.SetConfig(cfg)

	data, err := src.MarshalJSON()
	require.NoError(t, err)

	var restored Tonk
	require.NoError(t, restored.UnmarshalJSON(data))
	require.Equal(t, TonkCpuDifficultyEasy, restored.GetConfig().CpuDifficulty)

	for i := 0; i < restored.GetPlayerCnt(); i++ {
		if !restored.GetPlayer(i).GetIsHuman() {
			restored.SetCurrentPlayerIdx(i)
			break
		}
	}
	// SetRand は意図的に呼ばない —— 本番の復元経路には無い一行なので、
	// ここで呼ぶとこのテストが守るべき穴をそのまま隠してしまう。
	assert.NotPanics(t, func() { restored.CpuPlay() },
		"a restored game must not depend on SetRand having been called")
}
