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

func setupDragonTigerWebMockDefaults(m *interfaces.MockDragonTigerGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.DragonTigerPhaseBet).Maybe()
	m.On("GetDragonCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetTigerCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetBetType").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseDragonTigerOutput(t *testing.T, jsonStr string) *controller.DragonTigerWebOutput {
	t.Helper()
	var out controller.DragonTigerWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestDragonTigerWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(DragonTigerWebPresenter)
	m := new(interfaces.MockDragonTigerGame)
	setupDragonTigerWebMockDefaults(m)

	r := parseDragonTigerOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.DragonTigerPhaseBet, r.Phase)
	assert.Equal(t, 1000, r.Chips)
	assert.Nil(t, r.DragonCard)
	assert.Nil(t, r.TigerCard)
	assert.NotNil(t, r.History)
	assert.Empty(t, r.History)
	assert.Empty(t, r.Message)
}

func TestDragonTigerWebPresenter_Output_Error(t *testing.T) {
	p := new(DragonTigerWebPresenter)
	m := new(interfaces.MockDragonTigerGame)
	setupDragonTigerWebMockDefaults(m)
	r := parseDragonTigerOutput(t, p.Output(m, errors.New("oops")))
	assert.Equal(t, "oops", r.Message)
}

func TestDragonTigerWebPresenter_Output_DragonWins(t *testing.T) {
	p := new(DragonTigerWebPresenter)
	m := new(interfaces.MockDragonTigerGame)
	dc := domain.NewCard(domain.CardDesignSpade, 13, false)
	tc := domain.NewCard(domain.CardDesignHeart, 5, false)
	m.On("GetChips").Return(1100)
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd)
	m.On("GetDragonCard").Return(dc)
	m.On("GetTigerCard").Return(tc)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBetAmount").Return(100)
	m.On("GetBetType").Return(domain.DragonTigerBetDragon)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetPayout").Return(200)
	m.On("GetHistory").Return([]int{domain.DragonTigerResultDragon})

	r := parseDragonTigerOutput(t, p.Output(m, nil))
	assert.Empty(t, r.Message)
	assert.Equal(t, "dragontiger.result.dragonWins", r.MessageCode)
	assert.NotNil(t, r.DragonCard)
	assert.NotNil(t, r.TigerCard)
	assert.Equal(t, []int{domain.DragonTigerResultDragon}, r.History)
}

func TestDragonTigerWebPresenter_Output_TigerWins(t *testing.T) {
	p := new(DragonTigerWebPresenter)
	m := new(interfaces.MockDragonTigerGame)
	dc := domain.NewCard(domain.CardDesignSpade, 3, false)
	tc := domain.NewCard(domain.CardDesignHeart, 13, false)
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd)
	m.On("GetDragonCard").Return(dc)
	m.On("GetTigerCard").Return(tc)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBetAmount").Return(100)
	m.On("GetBetType").Return(domain.DragonTigerBetDragon)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetPayout").Return(0)
	m.On("GetHistory").Return([]int{domain.DragonTigerResultTiger})

	r := parseDragonTigerOutput(t, p.Output(m, nil))
	assert.Empty(t, r.Message)
	assert.Equal(t, "dragontiger.result.tigerWins", r.MessageCode)
}

func TestDragonTigerWebPresenter_Output_Tie_OnDragonBet(t *testing.T) {
	p := new(DragonTigerWebPresenter)
	m := new(interfaces.MockDragonTigerGame)
	dc := domain.NewCard(domain.CardDesignSpade, 7, false)
	tc := domain.NewCard(domain.CardDesignHeart, 7, false)
	m.On("GetChips").Return(950)
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd)
	m.On("GetDragonCard").Return(dc)
	m.On("GetTigerCard").Return(tc)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBetAmount").Return(100)
	m.On("GetBetType").Return(domain.DragonTigerBetDragon)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetPayout").Return(50)
	m.On("GetHistory").Return([]int{domain.DragonTigerResultTie})

	r := parseDragonTigerOutput(t, p.Output(m, nil))
	assert.Empty(t, r.Message)
	assert.Equal(t, "dragontiger.result.tieRefund", r.MessageCode)
}

func TestDragonTigerWebPresenter_Output_Tie_OnTieBet(t *testing.T) {
	p := new(DragonTigerWebPresenter)
	m := new(interfaces.MockDragonTigerGame)
	dc := domain.NewCard(domain.CardDesignSpade, 7, false)
	tc := domain.NewCard(domain.CardDesignHeart, 7, false)
	m.On("GetChips").Return(1800)
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd)
	m.On("GetDragonCard").Return(dc)
	m.On("GetTigerCard").Return(tc)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBetAmount").Return(100)
	m.On("GetBetType").Return(domain.DragonTigerBetTie)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetPayout").Return(900)
	m.On("GetHistory").Return([]int{domain.DragonTigerResultTie})

	r := parseDragonTigerOutput(t, p.Output(m, nil))
	assert.Empty(t, r.Message)
	assert.Equal(t, "dragontiger.result.tieWin", r.MessageCode)
}

func TestDragonTigerWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(DragonTigerWebPresenter)
	m := new(interfaces.MockDragonTigerGame)
	t.Run("game not ended", func(t *testing.T) {
		m.On("GetGameEndFlag").Return(false).Once()
		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
	})
}
