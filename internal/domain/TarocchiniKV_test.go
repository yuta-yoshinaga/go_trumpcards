//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// Worker のセッションは毎リクエスト JSON を往復する。1 局を通して往復し続けても
// 状態が壊れないことを確かめる —— コーデックの取りこぼしは 1 フィールドでも
// 積み上がって局面を狂わせる。
func TestTarocchini_SurvivesRoundTrippingEveryRequest(t *testing.T) {
	g := NewDefaultTarocchini()
	g.SetRand(rand.New(rand.NewSource(3)))
	g.Reset()

	roundTrip := func(src *Tarocchini) *Tarocchini {
		t.Helper()
		data, err := src.MarshalJSON()
		require.NoError(t, err)
		var out Tarocchini
		require.NoError(t, out.UnmarshalJSON(data))
		out.SetRand(rand.New(rand.NewSource(3)))
		return &out
	}

	g.CpuScarto()
	if g.GetPhase() == TarocchiniPhaseScarto {
		d := g.GetPlayer(g.GetDealerIdx())
		idx := make([]int, 0, TarocchiniSurplus)
		for i := 0; i < d.GetCardsSize() && len(idx) < TarocchiniSurplus; i++ {
			if tarocchiniCanDiscard(d.GetCard(i)) {
				idx = append(idx, i)
			}
		}
		require.NoError(t, g.PlayerScarto(idx))
	}
	g = roundTrip(g)

	for trick := 0; trick < TarocchiniTrickCount; trick++ {
		for range TarocchiniPlayerCnt {
			if g.IsHumanTurn() {
				valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid, "trick %d", trick)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
			g = roundTrip(g) // 1 手ごとに KV を往復する
		}
		require.Equal(t, TarocchiniPhaseTrickEnd, g.GetPhase(), "trick %d", trick)
		g.ResolveTrick()
		g = roundTrip(g)
		if g.GetPhase() == TarocchiniPhaseTrickEnd {
			g.NextTrick()
			g = roundTrip(g)
		}
	}
	require.Equal(t, TarocchiniPhaseRoundEnd, g.GetPhase())
	total := 0
	for _, n := range g.GetRoundTricks() {
		total += n
	}
	require.Equal(t, TarocchiniTrickCount, total, "tricks were lost across the round trips")
}
