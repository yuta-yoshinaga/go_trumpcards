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

func setupSlapjackTest() *domain.Slapjack {
	g := domain.NewDefaultSlapjack()
	g.Reset()
	return g
}

func TestSlapjackWebPresenter_Output(t *testing.T) {
	p := new(presenter.SlapjackWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		g := setupSlapjackTest()
		result := p.Output(g, nil)
		assert.NotEmpty(t, result)

		var out controller.SlapjackWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Len(t, out.Players, 2)
		assert.Equal(t, 26, out.Players[0].StockSize)
		assert.Equal(t, 26, out.Players[1].StockSize)
		assert.False(t, out.GameEndFlag)
		assert.Equal(t, -1, out.WinnerIdx)
		assert.True(t, out.IsHumanTurn)
		assert.Empty(t, out.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := setupSlapjackTest()
		result := p.Output(g, errors.New("bad"))
		var out controller.SlapjackWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "error", out.MessageCode)
		assert.Equal(t, "bad", out.Message)
	})

	t.Run("game end human win", func(t *testing.T) {
		g := setupSlapjackTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.SlapjackPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		result := p.Output(g, nil)
		var out controller.SlapjackWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "slapjack.result.humanWin", out.MessageCode)
		assert.True(t, out.GameEndFlag)
	})

	t.Run("game end CPU win", func(t *testing.T) {
		g := setupSlapjackTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.SlapjackPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		result := p.Output(g, nil)
		var out controller.SlapjackWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "slapjack.result.cpuWin", out.MessageCode)
	})
}

func TestSlapjackWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SlapjackWebPresenter)
	g := setupSlapjackTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
