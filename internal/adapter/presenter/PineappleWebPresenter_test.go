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

func TestPineappleWebPresenter_Output(t *testing.T) {
	p := new(presenter.PineappleWebPresenter)

	setup := func() (*domain.Pineapple, []*domain.PineapplePlayer) {
		tc := domain.NewTrumpCards(0)
		cfg := domain.DefaultPineappleConfig()
		players := domain.NewPineapplePlayersForTable(cfg.TableSize)
		game := domain.NewPineapple(tc, players, cfg)
		return game, players
	}

	t.Run("initial state", func(t *testing.T) {
		game, _ := setup()
		game.SetPhase(domain.PineapplePhasePreFlop)

		result := p.Output(game, nil)
		var out controller.PineappleWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, domain.PineapplePhasePreFlop, out.Phase)
		assert.False(t, out.GameEndFlag)
		assert.Equal(t, "", out.Message)
		assert.Equal(t, "", out.MessageCode)
		assert.False(t, out.IsDiscardPhase)
		assert.False(t, out.MuckAvailable)
	})

	t.Run("error message", func(t *testing.T) {
		game, _ := setup()
		game.SetPhase(domain.PineapplePhasePreFlop)

		result := p.Output(game, errors.New("invalid action"))
		var out controller.PineappleWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "invalid action", out.Message)
		assert.Equal(t, "", out.MessageCode)
	})

	t.Run("discard phase message", func(t *testing.T) {
		game, _ := setup()
		game.SetPhase(domain.PineapplePhaseDiscard)

		result := p.Output(game, nil)
		var out controller.PineappleWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Select a card to discard.", out.Message)
		assert.Equal(t, "pineapple.discard.prompt", out.MessageCode)
		assert.True(t, out.IsDiscardPhase)
	})

	t.Run("muck available message", func(t *testing.T) {
		game, players := setup()
		game.SetPhase(domain.PineapplePhaseShowdown)
		// IsMuckAvailable returns true when phase==Showdown and human has WonAmount==0
		game.SetRoundResults([]domain.PineappleResult{
			{PlayerIdx: 0, WonAmount: 0},
		})
		_ = players // players[0] is human (IsHuman=true)

		result := p.Output(game, nil)
		var out controller.PineappleWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Muck or show your hand.", out.Message)
		assert.Equal(t, "pineapple.muck.prompt", out.MessageCode)
		assert.True(t, out.MuckAvailable)
	})

	t.Run("game end message - human wins", func(t *testing.T) {
		game, _ := setup()
		game.SetPhase(domain.PineapplePhaseEnd)
		game.SetGameEndFlag(true)
		game.SetRoundResults([]domain.PineappleResult{
			{PlayerIdx: 0, WonAmount: 100},
		})

		result := p.Output(game, nil)
		var out controller.PineappleWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Contains(t, out.Message, "You are the winner.")
		assert.Equal(t, "pineapple.result.win", out.MessageCode)
	})
}
