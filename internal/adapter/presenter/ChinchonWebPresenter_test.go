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

func setupChinchonWebMock(phase domain.ChinchonPhase, ended bool, winner, knocker int) (*interfaces.MockChinchonGame, []*domain.ChinchonPlayer) {
	m := new(interfaces.MockChinchonGame)
	players := makeChinchonPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(20)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(ended)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetKnockerIdx").Return(knocker)
	m.On("GetKnockerMelds").Return(([][]*domain.Card)(nil))
	m.On("GetConfig").Return(domain.DefaultChinchonConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestChinchonWebPresenter_Output(t *testing.T) {
	p := new(presenter.ChinchonWebPresenter)

	t.Run("initial state serialises", func(t *testing.T) {
		m, players := setupChinchonWebMock(domain.ChinchonPhaseDraw, false, -1, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		out := p.Output(m, nil)
		var parsed controller.ChinchonWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, 0, parsed.Phase)
		assert.Len(t, parsed.Players, 2)
		// CPU hand hidden during play.
		assert.Empty(t, parsed.Players[1].Cards)
	})

	t.Run("layoff reveals knocker melds and CPU hand", func(t *testing.T) {
		m := new(interfaces.MockChinchonGame)
		players := makeChinchonPlayers()
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		m.On("GetRoundNumber").Return(1)
		m.On("GetDrawPileCount").Return(20)
		m.On("GetDiscardTop").Return((*domain.Card)(nil))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.ChinchonPhaseLayoff)
		m.On("GetCurrentPlayerIdx").Return(1)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetKnockerIdx").Return(0)
		m.On("GetKnockerMelds").Return([][]*domain.Card{{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignSpade, 3, false),
		}})
		m.On("GetConfig").Return(domain.DefaultChinchonConfig())
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		out := p.Output(m, nil)
		var parsed controller.ChinchonWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Len(t, parsed.KnockerMelds, 1)
		assert.Len(t, parsed.Players[1].Cards, 1)
	})

	t.Run("game end with human winner", func(t *testing.T) {
		m, _ := setupChinchonWebMock(domain.ChinchonPhaseGameEnd, true, 0, -1)
		out := p.Output(m, nil)
		var parsed controller.ChinchonWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.True(t, parsed.GameEndFlag)
		assert.Equal(t, "chinchon.result.humanWin", parsed.MessageCode)
	})

	t.Run("game end no winner", func(t *testing.T) {
		m, _ := setupChinchonWebMock(domain.ChinchonPhaseGameEnd, true, -1, -1)
		out := p.Output(m, nil)
		var parsed controller.ChinchonWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "chinchon.draw", parsed.MessageCode)
	})

	t.Run("discard phase message", func(t *testing.T) {
		m, _ := setupChinchonWebMock(domain.ChinchonPhaseDiscard, false, -1, -1)
		out := p.Output(m, nil)
		var parsed controller.ChinchonWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "chinchon.discardPhase", parsed.MessageCode)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupChinchonWebMock(domain.ChinchonPhaseRoundEnd, false, -1, -1)
		out := p.Output(m, nil)
		var parsed controller.ChinchonWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "chinchon.roundEnd", parsed.MessageCode)
	})

	t.Run("error is surfaced", func(t *testing.T) {
		m, _ := setupChinchonWebMock(domain.ChinchonPhaseDraw, false, -1, -1)
		out := p.Output(m, errors.New("boom"))
		var parsed controller.ChinchonWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "boom", parsed.Message)
	})
}

func TestChinchonWebPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupChinchonWebMock(domain.ChinchonPhaseDraw, false, -1, -1)
	p := new(presenter.ChinchonWebPresenter)
	assert.NotPanics(t, func() { p.ActionLogOutput(m) })
}
