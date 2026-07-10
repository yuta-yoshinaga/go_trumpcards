//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const cegoMockOutput = `{"phase":0}`

func newCegoPlayMock() *interfaces.MockCegoGame {
	m := new(interfaces.MockCegoGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CegoPhasePlay)
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("IsHumanContractTurn").Return(false)
	m.On("IsHumanExchangeTurn").Return(false)
	return m
}

func TestNewCegoInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CegoInteractor: g must not be nil", func() {
			usecase.NewCegoInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCegoGame)
		assert.PanicsWithValue(t, "CegoInteractor: tp must not be nil", func() {
			usecase.NewCegoInteractor(gameMock, nil)
		})
	})
}

func TestCegoInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	gameMock.On("Reset").Return()

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestCegoInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	cfg := domain.CegoConfig{CpuDifficulty: domain.CegoCpuDifficultyHard, TargetDeals: 5}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestCegoInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	gameMock.On("PlayerBid", domain.CegoBidPlay).Return(nil)

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.Bid(domain.CegoBidPlay))
	gameMock.AssertCalled(t, "PlayerBid", domain.CegoBidPlay)
}

func TestCegoInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := new(interfaces.MockCegoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.CegoBidPlay).Return(errors.New("bad"))

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.Bid(domain.CegoBidPlay))
}

func TestCegoInteractor_Pass(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	gameMock.On("PlayerPass").Return(nil)

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.Pass())
	gameMock.AssertCalled(t, "PlayerPass")
}

func TestCegoInteractor_ChooseContract(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	gameMock.On("PlayerChooseContract", domain.CegoContractCego).Return(nil)

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.ChooseContract(domain.CegoContractCego))
	gameMock.AssertCalled(t, "PlayerChooseContract", domain.CegoContractCego)
}

func TestCegoInteractor_ChooseContractError(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := new(interfaces.MockCegoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerChooseContract", domain.CegoContractCego).Return(errors.New("bad"))

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.ChooseContract(domain.CegoContractCego))
}

func TestCegoInteractor_Discard(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	gameMock.On("PlayerDiscard", []int{0}).Return(nil)

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.Discard([]int{0}))
	gameMock.AssertCalled(t, "PlayerDiscard", []int{0})
}

func TestCegoInteractor_Play(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	gameMock.On("PlayerPlay", 2).Return(nil)

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
}

func TestCegoInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	gameMock.On("NextTrick").Return()

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestCegoInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := newCegoPlayMock()
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestCegoInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(cegoMockOutput)
	gameMock := new(interfaces.MockCegoGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cegoMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestCegoInteractor_GetConfig(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	gameMock := new(interfaces.MockCegoGame)
	cfg := domain.DefaultCegoConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestCegoInteractor_HintAndLog(t *testing.T) {
	tpMock := new(presenter.MockCegoPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockCegoGame)

	ci := usecase.NewCegoInteractor(gameMock, tpMock)
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}
