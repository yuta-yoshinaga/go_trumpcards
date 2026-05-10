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

func TestWarCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.WarCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		w := setupWarTest()
		result := p.Output(w, nil)
		assert.Contains(t, result, "War (戦争)")
		assert.Contains(t, result, "CPU:")
		assert.Contains(t, result, "あなた:")
		assert.Contains(t, result, "[場札]")
	})

	t.Run("error", func(t *testing.T) {
		w := setupWarTest()
		result := p.Output(w, errors.New("oops"))
		assert.Contains(t, result, "oops")
	})

	t.Run("win message", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.WarPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		assert.Contains(t, p.Output(w, nil), "あなたの勝ちです")
	})

	t.Run("lose message", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.WarPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		assert.Contains(t, p.Output(w, nil), "CPUの勝ちです")
	})

	t.Run("war phase", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.WarPhaseWarBury)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		assert.Contains(t, p.Output(w, nil), "戦争発生")
	})

	t.Run("resolved phase - player", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.WarPhaseResolved)
		raw["lw"], _ = json.Marshal(0)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		assert.Contains(t, p.Output(w, nil), "あなたがラウンドに勝利")
	})

	t.Run("resolved phase - cpu", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.WarPhaseResolved)
		raw["lw"], _ = json.Marshal(1)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		assert.Contains(t, p.Output(w, nil), "CPUがラウンドに勝利")
	})
}

func TestWarCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WarCuiPresenter)
	w := setupWarTest()
	assert.NotEmpty(t, p.ActionLogOutput(w))
}
