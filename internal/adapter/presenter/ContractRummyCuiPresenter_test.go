//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupContractRummyCuiMock(phase domain.ContractRummyPhase, gameEnd bool) (*interfaces.MockContractRummyGame, []*domain.ContractRummyPlayer) {
	m := new(interfaces.MockContractRummyGame)
	players := []*domain.ContractRummyPlayer{
		domain.NewContractRummyPlayer(true),
		domain.NewContractRummyPlayer(false),
		domain.NewContractRummyPlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(60)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultContractRummyConfig())
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetCurrentContract").Return(domain.ContractForRound(1))
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestContractRummyCuiPresenter_Output(t *testing.T) {
	p := new(presenter.ContractRummyCuiPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseDraw, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("play phase", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhasePlay, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseRoundEnd, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseGameEnd, true)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseDraw, false)
		out := p.Output(m, errors.New("err"))
		assert.NotEmpty(t, out)
	})

	t.Run("contract met player", func(t *testing.T) {
		m, players := setupContractRummyCuiMock(domain.ContractRummyPhasePlay, false)
		players[0].SetContractMet(true)
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		})
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})
}

func TestContractRummyCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ContractRummyCuiPresenter)
	m, _ := setupContractRummyCuiMock(domain.ContractRummyPhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
