//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newS27Interactor() *usecase.SevenTwentySevenInteractor {
	return usecase.NewSevenTwentySevenInteractor(
		domain.NewDefaultSevenTwentySeven(), new(presenter.SevenTwentySevenWebPresenter))
}

func TestSevenTwentySevenInteractor_ResetOpensTheDrawPhase(t *testing.T) {
	ti := newS27Interactor()
	assert.NotEmpty(t, ti.Reset())
	assert.Equal(t, domain.SevenTwentySevenPhaseDraw, ti.Game.GetPhase())
	assert.Equal(t, 1, ti.Game.GetDrawRound())
}

// **止まると、そのラウンドは決着まで自分で回る。** 人間に打つ手が無いのに
// 待ってしまうと盤が固まる。
func TestSevenTwentySevenInteractor_StandingFinishesTheRound(t *testing.T) {
	ti := newS27Interactor()
	ti.Reset()
	assert.NotEmpty(t, ti.TakeCard(false))
	assert.Equal(t, domain.SevenTwentySevenPhaseResult, ti.Game.GetPhase())
}

// 引き続ければ手札が増える。
func TestSevenTwentySevenInteractor_DrawingAddsACard(t *testing.T) {
	ti := newS27Interactor()
	ti.Reset()
	before := ti.Game.GetPlayer(0).GetCardsSize()
	ti.TakeCard(true)
	assert.Greater(t, ti.Game.GetPlayer(0).GetCardsSize(), before)
}

// ラウンドを最後まで回して次へ進める。
func TestSevenTwentySevenInteractor_PlaysThroughToTheNextRound(t *testing.T) {
	ti := newS27Interactor()
	ti.Reset()
	for guard := 0; guard < 30 && ti.Game.GetPhase() == domain.SevenTwentySevenPhaseDraw; guard++ {
		ti.TakeCard(guard < 1)
	}
	require.Equal(t, domain.SevenTwentySevenPhaseResult, ti.Game.GetPhase())

	round := ti.Game.GetRoundNumber()
	assert.NotEmpty(t, ti.NextRound())
	if !ti.Game.GetGameEndFlag() {
		assert.Equal(t, round+1, ti.Game.GetRoundNumber())
		assert.Equal(t, domain.SevenTwentySevenPhaseDraw, ti.Game.GetPhase())
	}
}

func TestSevenTwentySevenInteractor_HintAndLog(t *testing.T) {
	ti := newS27Interactor()
	ti.Reset()
	assert.NotEmpty(t, ti.Hint())
	assert.NotEmpty(t, ti.ActionLog())
}

func TestSevenTwentySevenInteractor_ResetWithConfig(t *testing.T) {
	ti := newS27Interactor()
	cfg := domain.DefaultSevenTwentySevenConfig()
	cfg.PlayerCount = 3
	cfg.Ante = 25
	ti.ResetWithConfig(cfg)
	assert.Equal(t, 3, ti.GetConfig().PlayerCount)
	assert.Equal(t, 25, ti.GetConfig().Ante)
	assert.Equal(t, 3, ti.Game.GetPlayerCnt())
}

// **KV 往復。** 復元した盤で指し続けられることを見る。
func TestSevenTwentySevenInteractor_SurvivesAKVRoundTrip(t *testing.T) {
	ti := newS27Interactor()
	ti.Reset()
	require.NotEmpty(t, ti.TakeCard(true))
	require.Equal(t, domain.SevenTwentySevenPhaseDraw, ti.Game.GetPhase(), "1 手で決着してしまった")

	data, err := ti.Snapshot()
	require.NoError(t, err)
	restored, err := usecase.RestoreSevenTwentySevenInteractor(data, new(presenter.SevenTwentySevenWebPresenter))
	require.NoError(t, err)

	assert.Equal(t, ti.Game.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, ti.Game.GetDrawRound(), restored.Game.GetDrawRound())
	assert.Equal(t, ti.Game.GetPot(), restored.Game.GetPot())
	for i := 0; i < ti.Game.GetPlayerCnt(); i++ {
		assert.Equal(t, ti.Game.GetPlayer(i).GetCardsSize(), restored.Game.GetPlayer(i).GetCardsSize(), "p%d の手札", i)
		assert.Equal(t, ti.Game.GetPlayer(i).GetChips(), restored.Game.GetPlayer(i).GetChips(), "p%d のチップ", i)
		assert.Equal(t, ti.Game.GetPlayer(i).GetStanding(), restored.Game.GetPlayer(i).GetStanding(), "p%d の止まり", i)
		// 得点が往復すること（手札が落ちていれば必ずずれる）。
		for _, side := range []int{domain.SevenTwentySevenSideLow, domain.SevenTwentySevenSideHigh} {
			wantV, wantOK := ti.Game.GetScore(i, side)
			gotV, gotOK := restored.Game.GetScore(i, side)
			assert.Equal(t, wantOK, gotOK, "p%d side %d の生存", i, side)
			assert.Equal(t, wantV, gotV, "p%d side %d の得点", i, side)
		}
	}

	before := restored.Game.GetPlayer(0).GetCardsSize()
	restored.TakeCard(true)
	assert.Greater(t, restored.Game.GetPlayer(0).GetCardsSize(), before, "復元後に引けていない")
}
