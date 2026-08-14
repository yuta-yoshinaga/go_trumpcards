//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **賭ける前にフロップの 1 枚目が見えている。** ここがこのゲームの全部。
func TestCourchevel_ExposesOneCardBeforeTheFirstBet(t *testing.T) {
	for range 20 {
		o := NewDefaultCourchevel()
		require.NoError(t, o.Reset())
		require.Equal(t, OmahaPhasePreFlop, o.GetPhase())
		require.Len(t, o.GetCommunityCards(), 1,
			"プリフロップで見えている場が 1 枚でない")
		assert.NotNil(t, o.GetCommunityCards()[0])
	}
}

// **通常のオマチと Big O は変わらない。** 片側だけの検査は、共有エンジンを
// 壊したときに気づけない ── 先に見せる枚数を足した変更なので、0 枚のままで
// あることを負のコントロールとして踏む。
func TestOmahaAndBigO_StillHideTheFlopBeforeTheFirstBet(t *testing.T) {
	for _, tc := range []struct {
		name string
		make func() *Omaha
	}{
		{"omaha", NewDefaultOmaha},
		{"omaha hi-lo", NewDefaultOmahaHiLo},
		{"big o", NewDefaultBigO},
		{"big o hi-lo", NewDefaultBigOHiLo},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.make()
			require.NoError(t, o.Reset())
			assert.Empty(t, o.GetCommunityCards(),
				"%s のプリフロップで場が見えている", tc.name)
		})
	}
}

// **手札は Big O と同じ 5 枚。** 変えたのは公開の時刻だけ。
func TestCourchevel_DealsFiveHoleCards(t *testing.T) {
	o := NewDefaultCourchevel()
	require.NoError(t, o.Reset())
	for i, p := range o.GetPlayers() {
		assert.Equal(t, bigOHoleCards, p.GetCardsSize(), "席 %d の手札", i)
	}
}

// **場は最後まで 5 枚。** 先に見せた 1 枚を引かずにフロップで 3 枚配ると
// 6 枚になり、役の作り方が変わってしまう。
func TestCourchevel_BoardEndsWithFiveCards(t *testing.T) {
	sawFlop := false
	for range 20 {
		o := NewDefaultCourchevel()
		require.NoError(t, o.Reset())

		for steps := 0; o.GetPhase() >= OmahaPhasePreFlop && o.GetPhase() <= OmahaPhaseRiver; steps++ {
			require.Less(t, steps, 200, "ハンドが終わらない")
			before := o.GetPhase()
			if err := o.PlayerAction(OmahaActionCheck, 0, 0); err != nil {
				require.NoError(t, o.PlayerAction(OmahaActionCall, 0, 0))
			}
			if before == OmahaPhasePreFlop && o.GetPhase() == OmahaPhaseFlop {
				sawFlop = true
				assert.Len(t, o.GetCommunityCards(), 3,
					"フロップ後の場が 3 枚でない (先に見せた 1 枚を二重に数えている)")
			}
		}
		assert.LessOrEqual(t, len(o.GetCommunityCards()), 5, "場が 5 枚を超えた")
	}
	require.True(t, sawFlop, "20 回回してもフロップに到達しなかった")
}

// **公開枚数は保存にも乗る。** 落とすと復元した卓が通常のオマハに戻り、
// フロップで 1 枚多く配られる。
func TestCourchevel_PreflopExposureSurvivesARoundTrip(t *testing.T) {
	o := NewDefaultCourchevel()
	require.NoError(t, o.Reset())
	require.Len(t, o.GetCommunityCards(), 1)

	data, err := json.Marshal(o)
	require.NoError(t, err)
	restored := new(Omaha)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, len(o.GetCommunityCards()), len(restored.GetCommunityCards()))

	// 復元した卓をフロップまで進めても場は 3 枚のまま。
	for steps := 0; restored.GetPhase() == OmahaPhasePreFlop; steps++ {
		require.Less(t, steps, 200)
		if err := restored.PlayerAction(OmahaActionCheck, 0, 0); err != nil {
			require.NoError(t, restored.PlayerAction(OmahaActionCall, 0, 0))
		}
	}
	if restored.GetPhase() == OmahaPhaseFlop {
		assert.Len(t, restored.GetCommunityCards(), 3,
			"復元した卓のフロップが 3 枚でない")
	}

	// **効くのは次のハンドから。** 公開枚数を保存に載せ忘れても、この局は
	// 場の札そのものが保存されているので何も起こらない ── 静かに Big O に
	// 戻るのは復元した卓が次を配ったときで、そこまで踏まないと検出できない。
	require.NoError(t, restored.Reset())
	assert.Len(t, restored.GetCommunityCards(), courchevelPreflopCommunity,
		"復元した卓が次のハンドで 1 枚見せていない (公開枚数が保存されていない)")
}

// **Hi-Lo も同じ公開順で始まる。**
func TestCourchevelHiLo_ExposesOneCardAndSplitsPots(t *testing.T) {
	o := NewDefaultCourchevelHiLo()
	require.NoError(t, o.Reset())
	assert.Len(t, o.GetCommunityCards(), 1)
	assert.True(t, o.GetIsHiLo(), "Hi-Lo になっていない")
}
