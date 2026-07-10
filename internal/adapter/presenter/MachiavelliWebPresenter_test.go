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

func machiavelliTable() [][]*domain.Card {
	return [][]*domain.Card{
		{ // set
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
			domain.NewCard(domain.CardDesignClover, 9, false),
		},
		{ // run
			domain.NewCard(domain.CardDesignSpade, 4, false),
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignSpade, 6, false),
		},
	}
}

func setupMachiavelliWebMock(phase domain.MachiavelliPhase, gameEnd bool) (*interfaces.MockMachiavelliGame, []*domain.MachiavelliPlayer) {
	m := new(interfaces.MockMachiavelliGame)
	players := []*domain.MachiavelliPlayer{
		domain.NewMachiavelliPlayer(true),
		domain.NewMachiavelliPlayer(false),
	}
	m.On("GetRoundNumber").Return(1)
	m.On("GetTargetRounds").Return(3)
	m.On("GetDrawPileCount").Return(52)
	m.On("GetTable").Return(machiavelliTable())
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetRoundWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultMachiavelliConfig())
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("PlayerDeadwoodValue", 0).Return(0)
	m.On("PlayerDeadwoodValue", 1).Return(15)
	return m, players
}

func unmarshalMachiavelli(t *testing.T, s string) controller.MachiavelliWebOutput {
	t.Helper()
	var out controller.MachiavelliWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestMachiavelliWebPresenter_Output(t *testing.T) {
	p := new(presenter.MachiavelliWebPresenter)

	t.Run("turn phase", func(t *testing.T) {
		m, players := setupMachiavelliWebMock(domain.MachiavelliPhaseTurn, false)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		out := unmarshalMachiavelli(t, p.Output(m, nil))
		assert.Len(t, out.Players, 2)
		assert.Len(t, out.Table, 2)
		assert.Equal(t, 0, out.Table[0].Kind) // set
		assert.Equal(t, 1, out.Table[1].Kind) // run
		assert.Equal(t, 1, out.RoundNumber)
		assert.Equal(t, "machiavelli.turnPhase", out.MessageCode)
		assert.Equal(t, domain.MachiavelliDefaultPlayerCount, out.Config.PlayerCount)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupMachiavelliWebMock(domain.MachiavelliPhaseTurn, false)
		out := unmarshalMachiavelli(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("round end reveals deadwood", func(t *testing.T) {
		m, _ := setupMachiavelliWebMock(domain.MachiavelliPhaseRoundEnd, false)
		out := unmarshalMachiavelli(t, p.Output(m, nil))
		assert.Equal(t, 15, out.Players[1].Deadwood)
		assert.Equal(t, "machiavelli.roundEnd", out.MessageCode)
	})

	t.Run("game ended", func(t *testing.T) {
		m, _ := setupMachiavelliWebMock(domain.MachiavelliPhaseGameEnd, true)
		out := unmarshalMachiavelli(t, p.Output(m, nil))
		assert.True(t, out.GameEndFlag)
	})
}

func TestMachiavelliWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MachiavelliWebPresenter)
	m, _ := setupMachiavelliWebMock(domain.MachiavelliPhaseTurn, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
