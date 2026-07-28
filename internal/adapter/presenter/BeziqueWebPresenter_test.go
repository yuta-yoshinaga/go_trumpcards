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

func setupBeziqueWebMock(trumpCard *domain.Card) *interfaces.MockBeziqueGame {
	m := new(interfaces.MockBeziqueGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BeziquePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetDealerIdx").Return(0)
	m.On("GetLeadPlayerIdx").Return(1)
	m.On("GetStockRemaining").Return(40)
	m.On("IsEndgame").Return(false)
	m.On("GetValidPlayIndices", 0).Return([]int{0})
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultBeziqueConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupBeziqueWebMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockBeziqueGame, []*domain.BeziquePlayer) {
	m := setupBeziqueWebMock(trumpCard)
	players := []*domain.BeziquePlayer{
		domain.NewBeziquePlayer(true),
		domain.NewBeziquePlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetDealPoints", 0).Return(18)
	m.On("GetDealPoints", 1).Return(5)
	m.On("GetDealMeldPoints", 0).Return(8)
	m.On("GetDealMeldPoints", 1).Return(0)
	m.On("GetMatchScore", 0).Return(118)
	m.On("GetMatchScore", 1).Return(45)
	return m, players
}

func TestBeziqueWebPresenter_Output_InitialState(t *testing.T) {
	p := new(presenter.BeziqueWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBeziqueWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	players[0].SetRoundScore(18)
	players[0].SetCumulativeScore(118)

	got := p.Output(m, nil)
	assert.NotEmpty(t, got)

	var out controller.BeziqueWebOutput
	assert.NoError(t, json.Unmarshal([]byte(got), &out))
	assert.Equal(t, 2, len(out.Players))
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 1, out.RoundNumber)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
	assert.NotNil(t, out.TrumpCard)
	assert.Equal(t, 40, out.StockRemaining)
	assert.False(t, out.IsEndgame)
	assert.Equal(t, []int{18, 5}, out.DealPoints)
	assert.Equal(t, []int{118, 45}, out.MatchScore)
	assert.Equal(t, 18, out.Players[0].RoundScore)
	assert.Equal(t, 118, out.Players[0].CumulativeScore)
	assert.Empty(t, out.AvailableMelds)
}

func TestBeziqueWebPresenter_Output_HumanCardsShownCPUHidden(t *testing.T) {
	p := new(presenter.BeziqueWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBeziqueWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	got := p.Output(m, nil)
	var out controller.BeziqueWebOutput
	_ = json.Unmarshal([]byte(got), &out)

	assert.True(t, out.Players[0].IsHuman)
	assert.Len(t, out.Players[0].Cards, 1)
	for _, c := range out.Players[1].Cards {
		assert.Empty(t, c.Design, "CPU card design should be hidden")
	}
}

func TestBeziqueWebPresenter_Output_MeldPhaseListsMelds(t *testing.T) {
	p := new(presenter.BeziqueWebPresenter)
	m, _ := setupBeziqueWebMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.BeziquePhaseMeld)
	m.On("GetAvailableMelds", 0).Return([]domain.BeziqueMeld{
		{Type: domain.BeziqueMeldMarriage, Suit: domain.CardDesignSpade, Points: 40},
	})

	got := p.Output(m, nil)
	var out controller.BeziqueWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Len(t, out.AvailableMelds, 1)
	assert.Equal(t, 40, out.AvailableMelds[0].Points)
	assert.Equal(t, "bezique.meldPhase", out.MessageCode)
}

func TestBeziqueWebPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.BeziqueWebPresenter)
	m := setupBeziqueWebMock(nil)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(domain.NewBeziquePlayer(true))
	m.On("GetPlayer", 1).Return(domain.NewBeziquePlayer(false))
	m.On("GetDealPoints", 0).Return(0)
	m.On("GetDealPoints", 1).Return(0)
	m.On("GetDealMeldPoints", 0).Return(0)
	m.On("GetDealMeldPoints", 1).Return(0)
	m.On("GetMatchScore", 0).Return(1010)
	m.On("GetMatchScore", 1).Return(800)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)

	got := p.Output(m, nil)
	var out controller.BeziqueWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.True(t, out.GameEndFlag)
	assert.Equal(t, 0, out.WinnerIdx)
	assert.Equal(t, "bezique.result.p0Win", out.MessageCode)
}

func TestBeziqueWebPresenter_Output_Error(t *testing.T) {
	p := new(presenter.BeziqueWebPresenter)
	m, _ := setupBeziqueWebMockWithPlayers(nil)
	got := p.Output(m, errors.New("boom"))
	var out controller.BeziqueWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Equal(t, "boom", out.Message)
}

func TestBeziqueWebPresenter_HintOutput_Card(t *testing.T) {
	p := new(presenter.BeziqueWebPresenter)
	trump := domain.NewCard(domain.CardDesignSpade, 13, false)
	m, players := setupBeziqueWebMockWithPlayers(trump)
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	idx := 0
	m.On("GetHint").Return(&domain.BeziqueHint{CardIndex: &idx, Reason: "lead_low"})

	got := p.HintOutput(m)
	var out controller.BeziqueWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, 0, *out.Hint.CardIndex)
	assert.Equal(t, "lead_low", out.Hint.Reason)
}

func TestBeziqueWebPresenter_HintOutput_None(t *testing.T) {
	p := new(presenter.BeziqueWebPresenter)
	m, _ := setupBeziqueWebMockWithPlayers(nil)
	m.On("GetHint").Return((*domain.BeziqueHint)(nil))
	got := p.HintOutput(m)
	var out controller.BeziqueWebOutput
	_ = json.Unmarshal([]byte(got), &out)
	assert.Nil(t, out.Hint)
}

func TestBeziqueWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.BeziqueWebPresenter)
	m, _ := setupBeziqueWebMockWithPlayers(nil)
	got := p.ActionLogOutput(m)
	assert.NotEmpty(t, got)
}
