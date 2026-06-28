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

func TestBeggarMyNeighbourWebPresenter_Output(t *testing.T) {
	p := new(presenter.BeggarMyNeighbourWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		result := p.Output(g, nil)
		assert.NotEmpty(t, result)

		var resObj controller.BeggarMyNeighbourWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 2)
		assert.Equal(t, 26, resObj.Players[0].DrawPileSize)
		assert.Equal(t, 26, resObj.Players[1].DrawPileSize)
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Empty(t, resObj.MessageCode)
		assert.Equal(t, -1, resObj.PenaltyOwnerIdx)
	})

	t.Run("error message", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		result := p.Output(g, errors.New("bad"))
		var resObj controller.BeggarMyNeighbourWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "error", resObj.MessageCode)
		assert.Equal(t, "bad", resObj.Message)
	})

	t.Run("game end message human win", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["gf"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.BeggarMyNeighbourPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		result := p.Output(g, nil)
		var resObj controller.BeggarMyNeighbourWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "beggarmyneighbour.result.humanWin", resObj.MessageCode)
		assert.True(t, resObj.GameEndFlag)
	})

	t.Run("game end CPU win message", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["gf"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.BeggarMyNeighbourPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		result := p.Output(g, nil)
		var resObj controller.BeggarMyNeighbourWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "beggarmyneighbour.result.cpuWin", resObj.MessageCode)
	})

	t.Run("game end draw message", func(t *testing.T) {
		g := setupBeggarMyNeighbourTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["gf"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(-1)
		raw["ph"], _ = json.Marshal(domain.BeggarMyNeighbourPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		result := p.Output(g, nil)
		var resObj controller.BeggarMyNeighbourWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "beggarmyneighbour.result.draw", resObj.MessageCode)
	})
}

func TestBeggarMyNeighbourWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BeggarMyNeighbourWebPresenter)
	g := setupBeggarMyNeighbourTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
