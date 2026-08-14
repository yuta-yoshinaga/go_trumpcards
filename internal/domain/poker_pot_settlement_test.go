//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **ハンドが終わったらポットは 0。**
//
// 配り終えたのに `pot` が残っていると、残高を読む側 (卓をまたいでチップを持ち回る
// ミックスゲーム、結果画面) が「まだ配っていない額がある」と読む。全員降りて終わる
// 経路 (`resolveLastPlayer`) は昔から 0 にしており、ショーダウンで終わる経路だけが
// 残していた —— 同じ「ハンドが終わった」状態なのに片方だけ違うのが誤りだった。
//
// 実測 (リバイ/アドオンを切った 200 ハンド): 176 ハンドで pot が残っていた。
// チップ自体は配られているので**総量は保存している**。
func TestHoldem_PotIsZeroOnceTheHandIsOver(t *testing.T) {
	zeroed, ended := 0, 0
	for range 200 {
		g := NewDefaultHoldem()
		cfg := g.GetConfig()
		// **外からチップが入らないようにする。** リバイは破産した席にチップを
		// 足すので、保存則の検査に混ぜると増減の理由が分からなくなる。
		cfg.RebuyEnabled = false
		cfg.AddonEnabled = false
		g.SetConfig(cfg)
		require.NoError(t, g.Reset())

		total := 0
		for i := 0; i < g.GetPlayerCnt(); i++ {
			total += g.GetPlayer(i).GetChips()
		}
		total += g.GetPot()

		for range 40 {
			ph := g.GetPhase()
			if ph == HoldemPhaseEnd || ph == HoldemPhaseShowdown || ph == HoldemPhaseRebuy {
				break
			}
			if err := g.PlayerAction(HoldemActionFold, 0, 0); err != nil {
				break
			}
		}
		if g.GetPhase() != HoldemPhaseEnd {
			continue // 人間のマック待ちなどはこの検査の対象外。
		}
		ended++
		assert.Equal(t, 0, g.GetPot(), "ハンドが終わったのにポットが残っている")
		if g.GetPot() == 0 {
			zeroed++
		}
		// **チップの総量は変わらない。** 配られただけで、消えても湧いてもいない。
		after := 0
		for i := 0; i < g.GetPlayerCnt(); i++ {
			after += g.GetPlayer(i).GetChips()
		}
		assert.Equal(t, total, after+g.GetPot(), "チップの総量が変わっている")
	}
	require.Positive(t, ended, "1 ハンドも終局まで進まなかった")
	assert.Equal(t, ended, zeroed)
}
