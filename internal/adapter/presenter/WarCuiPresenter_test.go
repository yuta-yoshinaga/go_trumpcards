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

	// **打ち切りがいつ来るかを出す。**Web はアリーナに「ラウンド: n / max」を
	// 常時出しているのに、CUI は一度も出しておらず、あと何ラウンドで強制引き分けに
	// なるかが分からなかった (#4865)。
	t.Run("round progress", func(t *testing.T) {
		w := setupWarTest()
		cfg := w.GetConfig()
		cfg.MaxRounds = 120
		w.SetConfig(cfg)
		assert.Contains(t, p.Output(w, nil), "ラウンド: 0 / 120")
	})

	// 打ち切り目前は強調する。色を出す側と出さない側の両方を踏む。
	t.Run("round progress highlights the cap", func(t *testing.T) {
		orig := color.NoColor()
		color.SetNoColor(false)
		defer color.SetNoColor(orig)

		w := setupWarTest()
		cfg := w.GetConfig()
		cfg.MaxRounds = 100
		w.SetConfig(cfg)
		assert.NotContains(t, p.Output(w, nil), color.Yellow("ラウンド: 0 / 100（上限で保有枚数の多い方が勝ち）"))

		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["rp"], _ = json.Marshal(90)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)
		assert.Contains(t, p.Output(w, nil), color.Yellow("ラウンド: 90 / 100（上限で保有枚数の多い方が勝ち）"))
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

	t.Run("shows lastBurialCount when phase is resolved and > 0", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.WarPhaseResolved)
		raw["lb"], _ = json.Marshal(6) // 6 cards buried
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		assert.Contains(t, p.Output(w, nil), "（6枚伏せました）")
	})
}

func TestWarCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WarCuiPresenter)
	w := setupWarTest()
	assert.NotEmpty(t, p.ActionLogOutput(w))
}
