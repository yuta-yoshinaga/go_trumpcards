//go:build test

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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupConquianWebMock(phase domain.ConquianPhase, ended bool, winner, roundWinner int) (*interfaces.MockConquianGame, []*domain.ConquianPlayer) {
	m := new(interfaces.MockConquianGame)
	players := makeConquianPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(20)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(ended)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetRoundWinnerIdx").Return(roundWinner)
	m.On("GetTookDiscard").Return(false)
	m.On("GetConfig").Return(domain.DefaultConquianConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestConquianWebPresenter_Output(t *testing.T) {
	p := new(presenter.ConquianWebPresenter)

	t.Run("initial state serialises", func(t *testing.T) {
		m, players := setupConquianWebMock(domain.ConquianPhaseDraw, false, -1, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignDiamond, 3, false),
		})
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		out := p.Output(m, nil)
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, 0, parsed.Phase)
		assert.Len(t, parsed.Players, 2)
		assert.Len(t, parsed.Players[0].Melds, 1)
		// CPU hand hidden during play
		assert.Empty(t, parsed.Players[1].Cards)
	})

	t.Run("reveals CPU hand at round end", func(t *testing.T) {
		m, players := setupConquianWebMock(domain.ConquianPhaseRoundEnd, false, -1, 0)
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		out := p.Output(m, nil)
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Len(t, parsed.Players[1].Cards, 1)
	})

	t.Run("game end with human winner", func(t *testing.T) {
		m, _ := setupConquianWebMock(domain.ConquianPhaseGameEnd, true, 0, 0)
		out := p.Output(m, nil)
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.True(t, parsed.GameEndFlag)
		assert.Equal(t, "conquian.result.humanWin", parsed.MessageCode)
	})

	t.Run("game end draw", func(t *testing.T) {
		m, _ := setupConquianWebMock(domain.ConquianPhaseGameEnd, true, -1, -1)
		out := p.Output(m, nil)
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "conquian.draw", parsed.MessageCode)
	})

	t.Run("error is surfaced", func(t *testing.T) {
		m, _ := setupConquianWebMock(domain.ConquianPhaseDraw, false, -1, -1)
		out := p.Output(m, errors.New("boom"))
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "boom", parsed.Message)
	})
}

func TestConquianWebPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupConquianWebMock(domain.ConquianPhaseDraw, false, -1, -1)
	p := new(presenter.ConquianWebPresenter)
	assert.NotPanics(t, func() { p.ActionLogOutput(m) })
}
