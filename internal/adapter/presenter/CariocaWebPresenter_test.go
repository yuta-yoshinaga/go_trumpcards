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

func setupCariocaWebMock() (*interfaces.MockCariocaGame, []*domain.CariocaPlayer) {
	m := new(interfaces.MockCariocaGame)
	players := []*domain.CariocaPlayer{
		domain.NewCariocaPlayer(true),
		domain.NewCariocaPlayer(false),
		domain.NewCariocaPlayer(false),
		domain.NewCariocaPlayer(false),
	}
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(60)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CariocaPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultCariocaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetCurrentContract").Return(domain.CariocaContractForRound(1))
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func unmarshalCarioca(t *testing.T, s string) controller.CariocaWebOutput {
	t.Helper()
	var out controller.CariocaWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestCariocaWebPresenter_Output(t *testing.T) {
	p := new(presenter.CariocaWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupCariocaWebMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		out := unmarshalCarioca(t, p.Output(m, nil))
		assert.Len(t, out.Players, 4)
		assert.Equal(t, 1, out.RoundNumber)
		assert.Equal(t, domain.CariocaTotalRounds, out.TotalRounds)
		assert.Len(t, out.ContractSlots, 2) // R1 has 2 sets
		assert.Equal(t, "carioca.drawPhase", out.MessageCode)
		assert.Equal(t, domain.CariocaDefaultPlayerCount, out.Config.PlayerCount)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupCariocaWebMock()
		out := unmarshalCarioca(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("game ended", func(t *testing.T) {
		m := new(interfaces.MockCariocaGame)
		players := []*domain.CariocaPlayer{
			domain.NewCariocaPlayer(true),
			domain.NewCariocaPlayer(false),
			domain.NewCariocaPlayer(false),
			domain.NewCariocaPlayer(false),
		}
		m.On("GetRoundNumber").Return(7)
		m.On("GetDrawPileCount").Return(0)
		m.On("GetDiscardTop").Return((*domain.Card)(nil))
		m.On("GetGameEndFlag").Return(true)
		m.On("GetPhase").Return(domain.CariocaPhaseGameEnd)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(0)
		m.On("GetConfig").Return(domain.DefaultCariocaConfig())
		m.On("GetRoundWinnerIdx").Return(0)
		m.On("GetCurrentContract").Return(domain.CariocaContractForRound(7))
		m.On("GetPlayerCnt").Return(4)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		m.On("GetPlayer", 2).Return(players[2])
		m.On("GetPlayer", 3).Return(players[3])
		out := unmarshalCarioca(t, p.Output(m, nil))
		assert.True(t, out.GameEndFlag)
	})
}

func TestCariocaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CariocaWebPresenter)
	m, _ := setupCariocaWebMock()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
