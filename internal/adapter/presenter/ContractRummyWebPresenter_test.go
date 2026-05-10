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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupContractRummyWebMock() (*interfaces.MockContractRummyGame, []*domain.ContractRummyPlayer) {
	m := new(interfaces.MockContractRummyGame)
	players := []*domain.ContractRummyPlayer{
		domain.NewContractRummyPlayer(true),
		domain.NewContractRummyPlayer(false),
		domain.NewContractRummyPlayer(false),
	}
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(60)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ContractRummyPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultContractRummyConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetCurrentContract").Return(domain.ContractForRound(1))
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func unmarshalContractRummy(t *testing.T, s string) controller.ContractRummyWebOutput {
	t.Helper()
	var out controller.ContractRummyWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestContractRummyWebPresenter_Output(t *testing.T) {
	p := new(presenter.ContractRummyWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupContractRummyWebMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		out := unmarshalContractRummy(t, p.Output(m, nil))
		assert.Len(t, out.Players, 3)
		assert.Equal(t, 1, out.RoundNumber)
		assert.Equal(t, domain.ContractRummyTotalRounds, out.TotalRounds)
		assert.Len(t, out.ContractSlots, 2) // R1 has 2 sets
		assert.Equal(t, "contractrummy.drawPhase", out.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupContractRummyWebMock()
		out := unmarshalContractRummy(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("game ended", func(t *testing.T) {
		m := new(interfaces.MockContractRummyGame)
		players := []*domain.ContractRummyPlayer{
			domain.NewContractRummyPlayer(true),
			domain.NewContractRummyPlayer(false),
			domain.NewContractRummyPlayer(false),
		}
		m.On("GetRoundNumber").Return(7)
		m.On("GetDrawPileCount").Return(0)
		m.On("GetDiscardTop").Return((*domain.Card)(nil))
		m.On("GetGameEndFlag").Return(true)
		m.On("GetPhase").Return(domain.ContractRummyPhaseGameEnd)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(0)
		m.On("GetConfig").Return(domain.DefaultContractRummyConfig())
		m.On("GetRoundWinnerIdx").Return(0)
		m.On("GetCurrentContract").Return(domain.ContractForRound(7))
		m.On("GetPlayerCnt").Return(3)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		m.On("GetPlayer", 2).Return(players[2])
		out := unmarshalContractRummy(t, p.Output(m, nil))
		assert.True(t, out.GameEndFlag)
	})
}

func TestContractRummyWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ContractRummyWebPresenter)
	m, _ := setupContractRummyWebMock()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
