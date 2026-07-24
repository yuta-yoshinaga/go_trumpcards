//go:build test

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

func setupClockSolitaireWebMockDefaults(gg *interfaces.MockClockSolitaireGame) {
	gg.On("GetPhase").Return(domain.ClockSolitairePhasePlaying).Maybe()
	gg.On("GetStepCount").Return(0).Maybe()
	gg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 5, false)).Maybe()

	var piles [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
	var fuc [domain.ClockSolitairePileCount]int
	for i := range domain.ClockSolitairePileCount {
		piles[i] = make([]*domain.ClockSolitaireCard, domain.ClockSolitaireCardsPerPile)
		for j := range domain.ClockSolitaireCardsPerPile {
			piles[i][j] = &domain.ClockSolitaireCard{
				Card:   domain.NewCard(domain.CardDesignSpade, i+1, false),
				FaceUp: false,
			}
		}
		fuc[i] = 0
	}
	gg.On("GetPiles").Return(piles).Maybe()
	gg.On("GetFaceUpCount").Return(fuc).Maybe()
	gg.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	gg.On("CanUndo").Return(false).Maybe()
}

func parseClockSolitaireOutput(t *testing.T, jsonStr string) *controller.ClockSolitaireWebOutput {
	t.Helper()
	var out controller.ClockSolitaireWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestClockSolitaireWebPresenterOutput_Playing(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	p := &ClockSolitaireWebPresenter{}

	result := p.Output(gg, nil)
	out := parseClockSolitaireOutput(t, result)

	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, "clocksolitaire.playing", out.MessageCode)
	assert.Len(t, out.Piles, domain.ClockSolitairePileCount)
	assert.Len(t, out.FaceUpCount, domain.ClockSolitairePileCount)
	assert.NotNil(t, out.CurrentCard)
}

func TestClockSolitaireWebPresenterOutput_Error(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	p := &ClockSolitaireWebPresenter{}

	result := p.Output(gg, errors.New("test error"))
	out := parseClockSolitaireOutput(t, result)

	assert.Equal(t, "test error", out.Message)
}

func TestClockSolitaireWebPresenterOutput_GameClear(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	gg.ExpectedCalls = nil // clear defaults
	gg.On("GetPhase").Return(domain.ClockSolitairePhaseGameClear)
	gg.On("GetStepCount").Return(42)
	gg.On("GetCurrentCard").Return((*domain.Card)(nil))

	var piles [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
	var fuc [domain.ClockSolitairePileCount]int
	for i := range domain.ClockSolitairePileCount {
		piles[i] = make([]*domain.ClockSolitaireCard, 0)
		fuc[i] = 4
	}
	gg.On("GetPiles").Return(piles)
	gg.On("GetFaceUpCount").Return(fuc)
	gg.On("CanUndo").Return(false)

	p := &ClockSolitaireWebPresenter{}

	result := p.Output(gg, nil)
	out := parseClockSolitaireOutput(t, result)

	assert.Equal(t, 1, out.Phase)
	assert.Equal(t, "clocksolitaire.gameClear", out.MessageCode)
	assert.Equal(t, 42, out.StepCount)
	assert.Nil(t, out.CurrentCard)
}

func TestClockSolitaireWebPresenterOutput_GameOver(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.ClockSolitairePhaseGameOver)
	gg.On("GetStepCount").Return(30)
	gg.On("GetCurrentCard").Return((*domain.Card)(nil))

	var piles [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
	var fuc [domain.ClockSolitairePileCount]int
	for i := range domain.ClockSolitairePileCount {
		piles[i] = make([]*domain.ClockSolitaireCard, 0)
	}
	gg.On("GetPiles").Return(piles)
	gg.On("GetFaceUpCount").Return(fuc)
	gg.On("CanUndo").Return(false)

	p := &ClockSolitaireWebPresenter{}

	result := p.Output(gg, nil)
	out := parseClockSolitaireOutput(t, result)

	assert.Equal(t, 2, out.Phase)
	assert.Equal(t, "clocksolitaire.gameOver", out.MessageCode)
}

func TestClockSolitaireWebPresenterOutput_FaceUpCards(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	setupClockSolitaireWebMockDefaults(gg)

	// 1枚表向きのパイル
	var piles [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard
	var fuc [domain.ClockSolitairePileCount]int
	for i := range domain.ClockSolitairePileCount {
		piles[i] = []*domain.ClockSolitaireCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
			{Card: domain.NewCard(domain.CardDesignClover, i+1, false), FaceUp: false},
			{Card: domain.NewCard(domain.CardDesignHeart, i+1, false), FaceUp: false},
			{Card: domain.NewCard(domain.CardDesignDiamond, i+1, false), FaceUp: false},
		}
		fuc[i] = 1
	}
	gg.ExpectedCalls = nil
	gg.On("GetPhase").Return(domain.ClockSolitairePhasePlaying)
	gg.On("GetStepCount").Return(5)
	gg.On("GetCurrentCard").Return(domain.NewCard(domain.CardDesignSpade, 3, false))
	gg.On("GetPiles").Return(piles)
	gg.On("GetFaceUpCount").Return(fuc)
	gg.On("CanUndo").Return(true)

	p := &ClockSolitaireWebPresenter{}

	result := p.Output(gg, nil)
	out := parseClockSolitaireOutput(t, result)

	assert.True(t, out.CanUndo)
	// 表向きカードにはcard情報がある
	assert.NotNil(t, out.Piles[0][0].Card)
	assert.True(t, out.Piles[0][0].FaceUp)
	// 裏向きカードにはcard情報がない
	assert.Nil(t, out.Piles[0][1].Card)
	assert.False(t, out.Piles[0][1].FaceUp)
}

func TestClockSolitaireWebPresenterActionLog_Playing(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	gg.On("GetPhase").Return(domain.ClockSolitairePhasePlaying)

	gg.On("GetGameEndFlag").Return(false)
	p := &ClockSolitaireWebPresenter{}
	result := p.ActionLogOutput(gg)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Empty(t, out.Entries)
}

func TestClockSolitaireWebPresenterActionLog_GameOver(t *testing.T) {
	gg := new(interfaces.MockClockSolitaireGame)
	gg.On("GetPhase").Return(domain.ClockSolitairePhaseGameOver)
	gg.On("GetGameEndFlag").Return(true)
	gg.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, ActionType: "step", Detail: "test"},
	})

	p := &ClockSolitaireWebPresenter{}
	result := p.ActionLogOutput(gg)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Len(t, out.Entries, 1)
}
