//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubCaribbeanDrawPresenter is a minimal presenter for snapshot tests.
type stubCaribbeanDrawPresenter struct{}

func (s *stubCaribbeanDrawPresenter) Output(_ interfaces.CaribbeanDrawGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubCaribbeanDrawPresenter) ActionLogOutput(_ interfaces.CaribbeanDrawGame) string {
	return `{"log":[]}`
}

func (s *stubCaribbeanDrawPresenter) HintOutput(_ interfaces.CaribbeanDrawGame) string {
	return `{}`
}

func TestCaribbeanDrawInteractor_SnapshotRestore(t *testing.T) {
	// **`NotEmpty(result)` は復元を何も検査していない。** スタブのプレゼンタは
	// 盤面を無視して固定文字列を返すので、`RestoreCaribbeanDrawInteractor` が
	// `data` を丸ごと捨てて新品の卓を作っても通る —— 実測済み。
	// **復元しなければ再現しない値**を見る。
	cs := domain.NewDefaultCaribbeanDraw()
	ci := NewCaribbeanDrawInteractor(cs, new(stubCaribbeanDrawPresenter))

	ci.Bet(100, 10)
	ci.Draw([]int{0, 2})

	wantChips := cs.GetChips()
	wantPhase := cs.GetPhase()
	wantHand := append([]*domain.Card(nil), cs.GetPlayerHand()...)
	require.Equal(t, domain.CaribbeanDrawPhaseAction, wantPhase, "ドローまで進んでいること")
	require.Equal(t, 100, cs.GetDrawCost(), "交換手数料が乗っていること")

	data, err := ci.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 100, "非公開フィールドだけの構造体は黙って `{}` になる")

	restored, err := RestoreCaribbeanDrawInteractor(data, new(stubCaribbeanDrawPresenter))
	require.NoError(t, err)

	got := restored.Game
	assert.Equal(t, wantChips, got.GetChips())
	assert.Equal(t, wantPhase, got.GetPhase())
	assert.Equal(t, 100, got.GetAnteBet())
	assert.Equal(t, 10, got.GetJackpotBet())
	assert.Equal(t, 100, got.GetDrawCost(), "交換手数料が往復で消えないこと")
	require.Len(t, got.GetPlayerHand(), 5)
	for i, c := range wantHand {
		assert.Equal(t, c.GetDesign(), got.GetPlayerHand()[i].GetDesign(), "札 %d のスート", i)
		assert.Equal(t, c.GetValue(), got.GetPlayerHand()[i].GetValue(), "札 %d の数字", i)
	}

	// 復元した卓でそのまま勝負を続けられること (退化していない)。
	assert.NotEmpty(t, restored.Play())
	assert.Equal(t, domain.CaribbeanDrawPhaseEnd, got.GetPhase())
}

func TestCaribbeanDrawInteractor_SnapshotRestore_BetPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanDraw()
	cs.SetChips(4321)
	ci := NewCaribbeanDrawInteractor(cs, new(stubCaribbeanDrawPresenter))

	data, err := ci.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreCaribbeanDrawInteractor(data, new(stubCaribbeanDrawPresenter))
	require.NoError(t, err)

	// 4321 は既定値ではないので、復元を飛ばして新品を作ると一致しない。
	assert.Equal(t, 4321, restored.Game.GetChips())
	assert.Equal(t, domain.CaribbeanDrawPhaseBet, restored.Game.GetPhase())

	assert.NotEmpty(t, restored.Bet(100, 0))
	assert.Equal(t, domain.CaribbeanDrawPhaseDraw, restored.Game.GetPhase(),
		"復元した卓でもベットの次はドロー")
}
