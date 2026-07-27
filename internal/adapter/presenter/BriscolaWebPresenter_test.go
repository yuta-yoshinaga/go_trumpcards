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

func setupBriscolaWebMock(trumpCard *domain.Card) *interfaces.MockBriscolaGame {
	m := new(interfaces.MockBriscolaGame)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BriscolaPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetDealerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(1)
	m.On("GetStockRemaining").Return(33)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultBriscolaConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupBriscolaWebMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockBriscolaGame, []*domain.BriscolaPlayer) {
	m := setupBriscolaWebMock(trumpCard)
	players := []*domain.BriscolaPlayer{
		domain.NewBriscolaPlayer(true),
		domain.NewBriscolaPlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayerPoints", 0).Return(15)
	m.On("GetPlayerPoints", 1).Return(5)
	return m, players
}

func TestBriscolaWebPresenter_Output_InitialState(t *testing.T) {
	p := new(presenter.BriscolaWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBriscolaWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	got := p.Output(m, nil)
	assert.NotEmpty(t, got)

	var out controller.BriscolaWebOutput
	assert.NoError(t, json.Unmarshal([]byte(got), &out))
	assert.Equal(t, 2, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 1, out.TrickNumber)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
	assert.NotNil(t, out.TrumpCard)
	assert.Equal(t, 33, out.StockRemaining)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, 15, out.Players[0].Points)
	assert.Equal(t, 5, out.Players[1].Points)
}

func TestBriscolaWebPresenter_Output_HumanCardsShownCPUHidden(t *testing.T) {
	p := new(presenter.BriscolaWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBriscolaWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	got := p.Output(m, nil)
	var out controller.BriscolaWebOutput
	_ = json.Unmarshal([]byte(got), &out)

	human := out.Players[0]
	assert.True(t, human.IsHuman)
	assert.Equal(t, 1, human.CardCount)
	assert.Len(t, human.Cards, 1)

	cpu := out.Players[1]
	assert.False(t, cpu.IsHuman)
	assert.Equal(t, 1, cpu.CardCount)
	for _, c := range cpu.Cards {
		assert.Empty(t, c.Design, "CPU card design should be hidden")
	}
}

func TestBriscolaWebPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.BriscolaWebPresenter)
	m := setupBriscolaWebMock(nil)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewBriscolaPlayer(true))
	m.On("GetPlayer", 1).Return(domain.NewBriscolaPlayer(false))
	m.On("GetPlayerPoints", 0).Return(70)
	m.On("GetPlayerPoints", 1).Return(50)
	// Override: game ended, p0 wins
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)

	got := p.Output(m, nil)
	var out controller.BriscolaWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.True(t, out.GameEndFlag)
	assert.Equal(t, 0, out.WinnerIdx)
	assert.Equal(t, "briscola.result.p0Win", out.MessageCode)
}

func TestBriscolaWebPresenter_Output_GameEndTie(t *testing.T) {
	p := new(presenter.BriscolaWebPresenter)
	m := setupBriscolaWebMock(nil)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewBriscolaPlayer(true))
	m.On("GetPlayer", 1).Return(domain.NewBriscolaPlayer(false))
	m.On("GetPlayerPoints", 0).Return(60)
	m.On("GetPlayerPoints", 1).Return(60)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(-1)

	got := p.Output(m, nil)
	var out controller.BriscolaWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Equal(t, "briscola.result.tie", out.MessageCode)
}

func TestBriscolaWebPresenter_Output_Error(t *testing.T) {
	p := new(presenter.BriscolaWebPresenter)
	m, _ := setupBriscolaWebMockWithPlayers(nil)
	got := p.Output(m, errors.New("boom"))
	var out controller.BriscolaWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Equal(t, "boom", out.Message)
}

func TestBriscolaWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BriscolaWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBriscolaWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	idx := 0
	m.On("GetHint").Return(&domain.BriscolaHint{CardIndex: &idx, Reason: "lead_low"})

	got := p.HintOutput(m)
	var out controller.BriscolaWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, 0, *out.Hint.CardIndex)
	assert.Equal(t, "lead_low", out.Hint.Reason)
}

func TestBriscolaWebPresenter_HintOutput_None(t *testing.T) {
	p := new(presenter.BriscolaWebPresenter)
	m, _ := setupBriscolaWebMockWithPlayers(nil)
	m.On("GetHint").Return((*domain.BriscolaHint)(nil))
	got := p.HintOutput(m)
	var out controller.BriscolaWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Nil(t, out.Hint)
}

func TestBriscolaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BriscolaWebPresenter)
	m, _ := setupBriscolaWebMockWithPlayers(nil)
	got := p.ActionLogOutput(m)
	assert.NotEmpty(t, got)
}
