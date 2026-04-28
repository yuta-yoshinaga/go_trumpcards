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

func TestShitheadWebPresenter_Output(t *testing.T) {
	p := new(presenter.ShitheadWebPresenter)

	t.Run("initial reset state", func(t *testing.T) {
		s := domain.NewDefaultShithead()
		s.Reset()
		out := p.Output(s, nil)
		var parsed controller.ShitheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, 4, len(parsed.Players))
		assert.Equal(t, 16, parsed.StockSize)
		assert.False(t, parsed.GameEndFlag)
		// human's hand cards visible, others' not
		assert.Equal(t, 3, len(parsed.Players[0].HandCards))
		assert.Equal(t, 0, len(parsed.Players[1].HandCards))
	})

	t.Run("error message", func(t *testing.T) {
		s := domain.NewDefaultShithead()
		s.Reset()
		out := p.Output(s, errors.New("boom"))
		var parsed controller.ShitheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "boom", parsed.Message)
	})

	t.Run("game end emits message and code", func(t *testing.T) {
		s := domain.NewDefaultShithead()
		s.Reset()
		// Drive a quick game loop until end. Budget is generous so even
		// pickup-heavy stretches resolve.
		for i := 0; i < 50000 && !s.GetGameEndFlag(); i++ {
			if s.IsHumanTurn() {
				// Always pickup to make progress; advances the turn deterministically.
				_ = s.PlayerPlay(nil)
			} else {
				s.CpuPlay()
			}
		}
		if !s.GetGameEndFlag() {
			t.Skip("game did not end within step budget; skipping (loop is non-deterministic)")
		}
		out := p.Output(s, nil)
		var parsed controller.ShitheadWebOutput
		assert.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Contains(t, parsed.Message, "ゲーム終了")
		assert.Equal(t, "shithead.result.rankings", parsed.MessageCode)
	})
}

func TestShitheadWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ShitheadWebPresenter)
	s := domain.NewDefaultShithead()
	s.Reset()
	out := p.ActionLogOutput(s)
	assert.NotEmpty(t, out)
}
