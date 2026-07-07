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

const koenigrufenMockOutput = `{"phase":0}`

func newKoenigrufenPlayMock() *interfaces.MockKoenigrufenGame {
	m := new(interfaces.MockKoenigrufenGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.KoenigrufenPhasePlay)
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanBidTurn").Return(false)
	m.On("IsHumanCallTurn").Return(false)
	m.On("IsHumanDiscardTurn").Return(false)
	return m
}

func TestNewKoenigrufenInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KoenigrufenInteractor: g must not be nil", func() {
			usecase.NewKoenigrufenInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKoenigrufenGame)
		assert.PanicsWithValue(t, "KoenigrufenInteractor: tp must not be nil", func() {
			usecase.NewKoenigrufenInteractor(gameMock, nil)
		})
	})
}

func TestKoenigrufenInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	gameMock.On("Reset").Return()

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestKoenigrufenInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	cfg := domain.KoenigrufenConfig{CpuDifficulty: domain.KoenigrufenCpuDifficultyHard, TargetDeals: 4}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestKoenigrufenInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := new(interfaces.MockKoenigrufenGame)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	bad := domain.KoenigrufenConfig{CpuDifficulty: domain.KoenigrufenCpuDifficultyNormal, TargetDeals: 0}
	assert.Equal(t, koenigrufenMockOutput, ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestKoenigrufenInteractor_Bid(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	gameMock.On("PlayerBid", domain.KoenigrufenBidRufer).Return(nil)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.Bid(domain.KoenigrufenBidRufer))
	gameMock.AssertCalled(t, "PlayerBid", domain.KoenigrufenBidRufer)
}

func TestKoenigrufenInteractor_BidError(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := new(interfaces.MockKoenigrufenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerBid", domain.KoenigrufenBidRufer).Return(errors.New("bad"))

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.Bid(domain.KoenigrufenBidRufer))
}

func TestKoenigrufenInteractor_Pass(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	gameMock.On("PlayerPass").Return(nil)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.Pass())
	gameMock.AssertCalled(t, "PlayerPass")
}

func TestKoenigrufenInteractor_CallKing(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	gameMock.On("PlayerCallKing", 3).Return(nil)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.CallKing(3))
	gameMock.AssertCalled(t, "PlayerCallKing", 3)
}

func TestKoenigrufenInteractor_CallKingError(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := new(interfaces.MockKoenigrufenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerCallKing", 3).Return(errors.New("bad"))

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.CallKing(3))
}

func TestKoenigrufenInteractor_Discard(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	gameMock.On("PlayerDiscard", []int{0, 1, 2, 3, 4, 5}).Return(nil)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.Discard([]int{0, 1, 2, 3, 4, 5}))
	gameMock.AssertCalled(t, "PlayerDiscard", []int{0, 1, 2, 3, 4, 5})
}

func TestKoenigrufenInteractor_Play(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	gameMock.On("PlayerPlay", 2).Return(nil)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
}

func TestKoenigrufenInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	gameMock.On("NextTrick").Return()

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestKoenigrufenInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := newKoenigrufenPlayMock()
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestKoenigrufenInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(koenigrufenMockOutput)
	gameMock := new(interfaces.MockKoenigrufenGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, koenigrufenMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestKoenigrufenInteractor_GetConfig(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	gameMock := new(interfaces.MockKoenigrufenGame)
	cfg := domain.DefaultKoenigrufenConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestKoenigrufenInteractor_HintAndLog(t *testing.T) {
	tpMock := new(presenter.MockKoenigrufenPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockKoenigrufenGame)

	ci := usecase.NewKoenigrufenInteractor(gameMock, tpMock)
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}
