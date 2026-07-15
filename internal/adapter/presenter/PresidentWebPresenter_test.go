package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPresidentWebPresenter_Output(t *testing.T) {
	p := new(presenter.PresidentWebPresenter)

	t.Run("initial state serialises", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		raw := p.Output(pg, nil)
		var out controller.PresidentWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Len(t, out.Players, 4)
		assert.Equal(t, 0, out.CurrentTurn)
		assert.Equal(t, -1, out.LastPlayPlayerIdx)
		assert.True(t, out.Config.RevolutionEnabled)
	})

	t.Run("error message is propagated", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		raw := p.Output(pg, errors.New("test-err"))
		var out controller.PresidentWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Equal(t, "test-err", out.Message)
	})
}

func TestPresidentWebPresenter_ActionLog(t *testing.T) {
	p := new(presenter.PresidentWebPresenter)
	players := makePresidentPlayersForPresenter()
	pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
	out := p.ActionLogOutput(pg)
	assert.NotEmpty(t, out)
}

func TestPresidentWebPresenter_HintOutput(t *testing.T) {
	// Web hints are client-side, so HintOutput mirrors Output.
	p := new(presenter.PresidentWebPresenter)
	players := makePresidentPlayersForPresenter()
	pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
	assert.Equal(t, p.Output(pg, nil), p.HintOutput(pg))
}
