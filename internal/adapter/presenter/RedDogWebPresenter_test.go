package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupRedDogWebMockDefaults(m *interfaces.MockRedDogGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.RedDogPhaseBet).Maybe()
	m.On("GetInitialCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetThirdCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnte").Return(0).Maybe()
	m.On("GetRaise").Return(0).Maybe()
	m.On("GetSpread").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseRedDogOutput(t *testing.T, jsonStr string) *controller.RedDogWebOutput {
	t.Helper()
	var out controller.RedDogWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestRedDogWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(RedDogWebPresenter)
	m := new(interfaces.MockRedDogGame)
	setupRedDogWebMockDefaults(m)

	r := parseRedDogOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.RedDogPhaseBet, r.Phase)
	assert.Equal(t, 1000, r.Chips)
	assert.Empty(t, r.InitialCards)
	assert.Nil(t, r.ThirdCard)
	assert.Empty(t, r.Message)
}

func TestRedDogWebPresenter_Output_Error(t *testing.T) {
	p := new(RedDogWebPresenter)
	m := new(interfaces.MockRedDogGame)
	setupRedDogWebMockDefaults(m)
	r := parseRedDogOutput(t, p.Output(m, errors.New("oops")))
	assert.Equal(t, "oops", r.Message)
}

func TestRedDogWebPresenter_Output_EndWin(t *testing.T) {
	p := new(RedDogWebPresenter)
	m := new(interfaces.MockRedDogGame)
	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	}
	third := domain.NewCard(domain.CardDesignClover, 7, false)
	m.On("GetChips").Return(1100)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(cards)
	m.On("GetThirdCard").Return(third)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(100)
	m.On("GetSpread").Return(4)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetTotalPayout").Return(400)

	r := parseRedDogOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player wins!", r.Message)
	assert.Equal(t, "reddog.result.playerWins", r.MessageCode)
	assert.Len(t, r.InitialCards, 2)
	assert.NotNil(t, r.ThirdCard)
	assert.Equal(t, 400, r.TotalPayout)
}

func TestRedDogWebPresenter_Output_EndLose(t *testing.T) {
	p := new(RedDogWebPresenter)
	m := new(interfaces.MockRedDogGame)
	setupRedDogWebMockDefaults(m)
	m.ExpectedCalls = nil
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(([]*domain.Card)(nil))
	m.On("GetThirdCard").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(3)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetTotalPayout").Return(0)
	r := parseRedDogOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player loses.", r.Message)
	assert.Equal(t, "reddog.result.playerLoses", r.MessageCode)
}

func TestRedDogWebPresenter_Output_EndPush(t *testing.T) {
	p := new(RedDogWebPresenter)
	m := new(interfaces.MockRedDogGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.RedDogPhaseEnd)
	m.On("GetInitialCards").Return(([]*domain.Card)(nil))
	m.On("GetThirdCard").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetRaise").Return(0)
	m.On("GetSpread").Return(0)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetTotalPayout").Return(100)
	r := parseRedDogOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push.", r.Message)
	assert.Equal(t, "reddog.result.push", r.MessageCode)
}

func TestRedDogWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(RedDogWebPresenter)
	m := new(interfaces.MockRedDogGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
