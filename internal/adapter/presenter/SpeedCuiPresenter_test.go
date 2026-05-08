//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSpeedCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.SpeedCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, nil)
		assert.Contains(t, result, "Speed (スピード)")
		assert.Contains(t, result, "CPU:")
		assert.Contains(t, result, "[台札]")
		assert.Contains(t, result, "あなた:")
	})

	t.Run("shows human hand cards", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, nil)
		assert.Contains(t, result, "手札")
		assert.Contains(t, result, "山札")
	})

	t.Run("shows error", func(t *testing.T) {
		s := setupSpeedWebTest()
		result := p.Output(s, errors.New("bad play"))
		assert.Contains(t, result, "bad play")
	})

	t.Run("shows win message", func(t *testing.T) {
		s := setupSpeedWebTest()
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		assert.Contains(t, result, "あなたの勝ちです")
	})

	t.Run("shows lose message", func(t *testing.T) {
		s := setupSpeedWebTest()
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		assert.Contains(t, result, "CPUの勝ちです")
	})

	t.Run("shows stuck message", func(t *testing.T) {
		s := setupSpeedWebTest()
		data, _ := json.Marshal(s)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.SpeedPhaseStuck)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, s)

		result := p.Output(s, nil)
		assert.Contains(t, result, "膠着状態")
	})
}

func TestSpeedCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpeedCuiPresenter)
	s := setupSpeedWebTest()
	result := p.ActionLogOutput(s)
	assert.NotEmpty(t, result)
}
