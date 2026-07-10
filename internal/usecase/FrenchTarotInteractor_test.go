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

const frenchTarotMockOutput = `{"phase":0}`

func newFrenchTarotPlayMock() *interfaces.MockFrenchTarotGame {
	m := new(interfaces.MockFrenchTarotGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.FrenchTarotPhasePlay)
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("IsHumanDiscardTurn").Return(false)
	return m
}

func TestNewFrenchTarotInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "FrenchTarotInteractor: g must not be nil", func() {
			usecase.NewFrenchTarotInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockFrenchTarotGame)
		assert.PanicsWithValue(t, "FrenchTarotInteractor: tp must not be nil", func() {
			usecase.NewFrenchTarotInteractor(gameMock, nil)
		})
	})
}

func TestFrenchTarotInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := newFrenchTarotPlayMock()
	gameMock.On("Reset").Return()

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestFrenchTarotInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := newFrenchTarotPlayMock()
	cfg := domain.FrenchTarotConfig{CpuDifficulty: domain.FrenchTarotCpuDifficultyHard, TargetDeals: 4}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestFrenchTarotInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := new(interfaces.MockFrenchTarotGame)

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	bad := domain.FrenchTarotConfig{CpuDifficulty: domain.FrenchTarotCpuDifficultyNormal, TargetDeals: 0}
	assert.Equal(t, frenchTarotMockOutput, ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestFrenchTarotInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := newFrenchTarotPlayMock()
	gameMock.On("PlayerBid", domain.FrenchTarotBidGarde).Return(nil)

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.Bid(domain.FrenchTarotBidGarde))
	gameMock.AssertCalled(t, "PlayerBid", domain.FrenchTarotBidGarde)
}

func TestFrenchTarotInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := new(interfaces.MockFrenchTarotGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.FrenchTarotBidPetite).Return(errors.New("bad"))

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.Bid(domain.FrenchTarotBidPetite))
}

func TestFrenchTarotInteractor_Pass(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := newFrenchTarotPlayMock()
	gameMock.On("PlayerPass").Return(nil)

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.Pass())
	gameMock.AssertCalled(t, "PlayerPass")
}

func TestFrenchTarotInteractor_Discard(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := newFrenchTarotPlayMock()
	gameMock.On("PlayerDiscard", []int{0, 1, 2, 3, 4, 5}).Return(nil)

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.Discard([]int{0, 1, 2, 3, 4, 5}))
	gameMock.AssertCalled(t, "PlayerDiscard", []int{0, 1, 2, 3, 4, 5})
}

func TestFrenchTarotInteractor_Play(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := newFrenchTarotPlayMock()
	gameMock.On("PlayerPlay", 2).Return(nil)

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
}

func TestFrenchTarotInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := newFrenchTarotPlayMock()
	gameMock.On("NextTrick").Return()

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestFrenchTarotInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := newFrenchTarotPlayMock()
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestFrenchTarotInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(frenchTarotMockOutput)
	gameMock := new(interfaces.MockFrenchTarotGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, frenchTarotMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestFrenchTarotInteractor_GetConfig(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	gameMock := new(interfaces.MockFrenchTarotGame)
	cfg := domain.DefaultFrenchTarotConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestFrenchTarotInteractor_HintAndLog(t *testing.T) {
	tpMock := new(presenter.MockFrenchTarotPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockFrenchTarotGame)

	ci := usecase.NewFrenchTarotInteractor(gameMock, tpMock)
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}
