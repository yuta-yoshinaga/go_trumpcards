package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPigsTailWebPresenter_Output(t *testing.T) {
	p := &presenter.PigsTailWebPresenter{}

	t.Run("initial state", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		output := p.Output(pt, nil)

		var result controller.PigsTailWebOutput
		require.NoError(t, json.Unmarshal([]byte(output), &result))

		assert.Equal(t, 52, result.CircleCount)
		assert.Equal(t, 0, result.CenterCount)
		assert.False(t, result.GameEndFlag)
		assert.Equal(t, -1, result.LoserIdx)
		assert.Equal(t, 4, len(result.Players))
	})
	t.Run("with error", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		output := p.Output(pt, errors.New("test error"))

		var result controller.PigsTailWebOutput
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.Equal(t, "test error", result.Message)
	})
	t.Run("game ended - human loses", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		for !pt.GetGameEndFlag() {
			if pt.IsHumanTurn() {
				_ = pt.PlayerAction(0)
			} else {
				_ = pt.CpuAction()
			}
		}
		output := p.Output(pt, nil)

		var result controller.PigsTailWebOutput
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.True(t, result.GameEndFlag)
		assert.GreaterOrEqual(t, result.LoserIdx, 0)
		assert.NotEmpty(t, result.Message)
		assert.NotEmpty(t, result.MessageCode)
	})
	t.Run("players contain human cards", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		// Play a few turns to accumulate some penalty cards
		for i := 0; i < 10 && !pt.GetGameEndFlag(); i++ {
			if pt.IsHumanTurn() {
				_ = pt.PlayerAction(0)
			} else {
				_ = pt.CpuAction()
			}
		}
		output := p.Output(pt, nil)

		var result controller.PigsTailWebOutput
		require.NoError(t, json.Unmarshal([]byte(output), &result))

		for _, player := range result.Players {
			if player.IsHuman && player.CardCount > 0 {
				assert.NotEmpty(t, player.Cards)
			}
			if !player.IsHuman {
				assert.Empty(t, player.Cards)
			}
		}
	})
	t.Run("with cpu actions", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		card := domain.NewCard(domain.CardDesignSpade, 5, false)
		pt.SetCpuActions([]*domain.PigsTailCpuAction{
			{DrawPlayerIdx: 1, DrawnCard: card, PenaltyFlag: true, PenaltyCount: 3},
		})
		output := p.Output(pt, nil)

		var result controller.PigsTailWebOutput
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.Len(t, result.CpuActions, 1)
		assert.True(t, result.CpuActions[0].PenaltyFlag)
		assert.Equal(t, 3, result.CpuActions[0].PenaltyCount)
	})
	t.Run("with human action", func(t *testing.T) {
		pt := newTestPigsTailForPresenter()
		card := domain.NewCard(domain.CardDesignHeart, 10, false)
		pt.SetHumanAction(&domain.PigsTailCpuAction{
			DrawPlayerIdx: 0,
			DrawnCard:     card,
			PenaltyFlag:   false,
		})
		output := p.Output(pt, nil)

		var result controller.PigsTailWebOutput
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.NotNil(t, result.HumanAction)
		assert.False(t, result.HumanAction.PenaltyFlag)
	})
}

func TestPigsTailWebPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.PigsTailWebPresenter{}
	pt := newTestPigsTailForPresenter()
	output := p.ActionLogOutput(pt)
	assert.NotEmpty(t, output)
}
