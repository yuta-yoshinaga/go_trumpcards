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

func setupSpoonsTest() *domain.Spoons {
	g := domain.NewDefaultSpoons()
	g.Reset()
	return g
}

// spoonsSetField mutates a single JSON-encoded field on a Spoons game.
func spoonsSetField(g *domain.Spoons, fields map[string]any) {
	data, _ := json.Marshal(g)
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	for k, v := range fields {
		raw[k], _ = json.Marshal(v)
	}
	newData, _ := json.Marshal(raw)
	_ = json.Unmarshal(newData, g)
}

func TestSpoonsWebPresenter_Output(t *testing.T) {
	p := new(presenter.SpoonsWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		g := setupSpoonsTest()
		result := p.Output(g, nil)
		assert.NotEmpty(t, result)

		var out controller.SpoonsWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Len(t, out.Players, domain.SpoonsPlayerCnt)
		assert.Equal(t, domain.SpoonsPlayerCnt-1, out.SpoonsRemaining)
		assert.Equal(t, -1, out.WinnerIdx)
		assert.False(t, out.GameEndFlag)
		// 人間 (idx 0) のみ手札公開。
		assert.NotEmpty(t, out.Players[0].Hand)
		assert.Empty(t, out.Players[1].Hand)
		assert.True(t, out.Players[0].IsHuman)
	})

	t.Run("error message", func(t *testing.T) {
		g := setupSpoonsTest()
		var out controller.SpoonsWebOutput
		_ = json.Unmarshal([]byte(p.Output(g, errors.New("bad"))), &out)
		assert.Equal(t, "error", out.MessageCode)
		assert.Equal(t, "bad", out.Message)
	})

	t.Run("game end human win", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ge": true, "wi": 0, "ph": domain.SpoonsPhaseGameEnd})
		var out controller.SpoonsWebOutput
		_ = json.Unmarshal([]byte(p.Output(g, nil)), &out)
		assert.Equal(t, "spoons.result.humanWin", out.MessageCode)
		assert.True(t, out.GameEndFlag)
	})

	t.Run("game end cpu win", func(t *testing.T) {
		g := setupSpoonsTest()
		spoonsSetField(g, map[string]any{"ge": true, "wi": 2, "ph": domain.SpoonsPhaseGameEnd})
		var out controller.SpoonsWebOutput
		_ = json.Unmarshal([]byte(p.Output(g, nil)), &out)
		assert.Equal(t, "spoons.result.cpuWin", out.MessageCode)
	})
}

func TestSpoonsWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SpoonsWebPresenter)
	g := setupSpoonsTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
