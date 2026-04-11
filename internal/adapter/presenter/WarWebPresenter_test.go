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

func setupWarTest() *domain.War {
	tc := domain.NewTrumpCards(0)
	players := []*domain.WarPlayer{
		domain.NewWarPlayer(true),
		domain.NewWarPlayer(false),
	}
	w := domain.NewWar(tc, players, domain.DefaultWarConfig())
	w.Reset()
	return w
}

func TestWarWebPresenter_Output(t *testing.T) {
	p := new(presenter.WarWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		w := setupWarTest()
		result := p.Output(w, nil)
		assert.NotEmpty(t, result)

		var resObj controller.WarWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 2)
		assert.Equal(t, 26, resObj.Players[0].DrawPileSize)
		assert.Equal(t, 26, resObj.Players[1].DrawPileSize)
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, "reveal", resObj.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		w := setupWarTest()
		result := p.Output(w, errors.New("bad"))
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "error", resObj.MessageCode)
		assert.Equal(t, "bad", resObj.Message)
	})

	t.Run("reveal populated after step", func(t *testing.T) {
		w := setupWarTest()
		_ = w.Step()
		result := p.Output(w, nil)
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.PlayerRevealed)
		assert.NotNil(t, resObj.CpuRevealed)
		assert.Equal(t, 2, resObj.WarPotSize)
	})

	t.Run("game end message via JSON override", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.WarPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		result := p.Output(w, nil)
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "gameEnd", resObj.MessageCode)
		assert.True(t, resObj.GameEndFlag)
	})

	t.Run("war phase message", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.WarPhaseWarBury)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		result := p.Output(w, nil)
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "war", resObj.MessageCode)
	})

	t.Run("resolved phase message", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ph"], _ = json.Marshal(domain.WarPhaseResolved)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		result := p.Output(w, nil)
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "resolved", resObj.MessageCode)
	})
}

func TestWarWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WarWebPresenter)
	w := setupWarTest()
	assert.NotEmpty(t, p.ActionLogOutput(w))
}
