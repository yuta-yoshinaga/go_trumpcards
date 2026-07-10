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

const scartoMockOutput = `{"phase":1}`

func newScartoPlayMock() *interfaces.MockScartoGame {
	m := new(interfaces.MockScartoGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ScartoPhasePlay)
	m.On("IsHumanTurn").Return(true)
	m.On("IsHumanScartoTurn").Return(false)
	return m
}

func TestNewScartoInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ScartoInteractor: g must not be nil", func() {
			usecase.NewScartoInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockScartoGame)
		assert.PanicsWithValue(t, "ScartoInteractor: tp must not be nil", func() {
			usecase.NewScartoInteractor(gameMock, nil)
		})
	})
}

func TestScartoInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := newScartoPlayMock()
	gameMock.On("Reset").Return()

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, scartoMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestScartoInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := newScartoPlayMock()
	cfg := domain.ScartoConfig{CpuDifficulty: domain.ScartoCpuDifficultyHard, TargetDeals: 3}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, scartoMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestScartoInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := new(interfaces.MockScartoGame)

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	bad := domain.ScartoConfig{CpuDifficulty: domain.ScartoCpuDifficultyNormal, TargetDeals: 0}
	assert.Equal(t, scartoMockOutput, ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestScartoInteractor_Discard(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := newScartoPlayMock()
	gameMock.On("PlayerScarto", []int{0, 1, 2}).Return(nil)

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, scartoMockOutput, ci.Discard([]int{0, 1, 2}))
	gameMock.AssertCalled(t, "PlayerScarto", []int{0, 1, 2})
}

func TestScartoInteractor_DiscardError(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := new(interfaces.MockScartoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerScarto", []int{0, 1, 2}).Return(errors.New("bad"))

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, scartoMockOutput, ci.Discard([]int{0, 1, 2}))
}

func TestScartoInteractor_Play(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := newScartoPlayMock()
	gameMock.On("PlayerPlay", 2).Return(nil)

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, scartoMockOutput, ci.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
}

func TestScartoInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := newScartoPlayMock()
	gameMock.On("NextTrick").Return()

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, scartoMockOutput, ci.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestScartoInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := newScartoPlayMock()
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, scartoMockOutput, ci.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestScartoInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(scartoMockOutput)
	gameMock := new(interfaces.MockScartoGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, scartoMockOutput, ci.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestScartoInteractor_GetConfig(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	gameMock := new(interfaces.MockScartoGame)
	cfg := domain.DefaultScartoConfig()
	gameMock.On("GetConfig").Return(cfg)

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestScartoInteractor_HintAndLog(t *testing.T) {
	tpMock := new(presenter.MockScartoPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockScartoGame)

	ci := usecase.NewScartoInteractor(gameMock, tpMock)
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}
