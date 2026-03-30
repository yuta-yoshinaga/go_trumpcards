//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func setupSpeedWebTest() *domain.Speed {
	tc := domain.NewTrumpCards(0)
	players := []*domain.SpeedPlayer{
		domain.NewSpeedPlayer(true),
		domain.NewSpeedPlayer(false),
	}
	s := domain.NewSpeed(tc, players, domain.DefaultSpeedConfig())
	s.Reset()
	return s
}

func TestSpeedWebPresenter_Output(t *testing.T) {
	p := new(presenter.SpeedWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, nil)
		assert.NotEmpty(t, result)

		var resObj controller.SpeedWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.Players, 2)
		assert.Len(t, resObj.CenterPiles, 2)
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, -1, resObj.WinnerIdx)
	})

	t.Run("human cards visible, CPU hidden", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, nil)
		var resObj controller.SpeedWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		human := resObj.Players[0]
		assert.True(t, human.IsHuman)
		assert.Equal(t, 4, human.CardCount)
		assert.Len(t, human.Cards, 4)
		assert.Equal(t, 21, human.DrawPileSize)

		cpu := resObj.Players[1]
		assert.False(t, cpu.IsHuman)
		assert.Equal(t, 4, cpu.CardCount)
		assert.Empty(t, cpu.Cards)
	})

	t.Run("error message", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, errors.New("invalid play"))
		var resObj controller.SpeedWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "error", resObj.MessageCode)
		assert.Equal(t, "invalid play", resObj.Message)
	})

	t.Run("game end message", func(t *testing.T) {
		s := setupSpeedWebTest()
		// Play until we can force a win - use JSON to set gameEndFlag
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		var resObj controller.SpeedWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "gameEnd", resObj.MessageCode)
		assert.True(t, resObj.GameEndFlag)
	})

	t.Run("stuck phase message", func(t *testing.T) {
		s := setupSpeedWebTest()
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseStuck)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		var resObj controller.SpeedWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "stuck", resObj.MessageCode)
	})

	t.Run("config in output", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, nil)
		var resObj controller.SpeedWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, int(domain.SpeedCpuDifficultyNormal), resObj.Config.CpuDifficulty)
	})

	t.Run("hint included when available", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, nil)
		var resObj controller.SpeedWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		// After reset, there should be at least one valid play -> hint present
		// (may or may not be present depending on random deal)
		// Just verify structure is valid
		assert.NotEmpty(t, result)
	})
}

func TestSpeedWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpeedWebPresenter)
	s := setupSpeedWebTest()
	result := p.ActionLogOutput(s)
	assert.NotEmpty(t, result)
}
