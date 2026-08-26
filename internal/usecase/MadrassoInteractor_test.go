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

const madMockOutput = `{"phase":0}`

func TestNewMadrassoInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MadrassoInteractor: g must not be nil", func() {
			usecase.NewMadrassoInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMadrassoGame)
		assert.PanicsWithValue(t, "MadrassoInteractor: tp must not be nil", func() {
			usecase.NewMadrassoInteractor(gameMock, nil)
		})
	})
}

func TestMadrassoInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MadrassoPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMadrassoInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	cfg := domain.MadrassoConfig{CpuDifficulty: domain.MadrassoCpuDifficultyHard, TargetPoints: 31}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MadrassoPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestMadrassoInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	bad := domain.MadrassoConfig{CpuDifficulty: 99, TargetPoints: 21}
	assert.Equal(t, madMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestMadrassoInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 3).Return(nil)
	gameMock.On("GetPhase").Return(domain.MadrassoPhaseTrickEnd)
	gameMock.On("ResolveTrick").Return()

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.Play(3))
	gameMock.AssertCalled(t, "PlayerPlay", 3)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestMadrassoInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.MadrassoPhasePlay)

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestMadrassoInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("boom"))

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestMadrassoInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestMadrassoInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MadrassoPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestMadrassoInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.MadrassoPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestMadrassoInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(madMockOutput)
	gameMock := new(interfaces.MockMadrassoGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, madMockOutput, ti.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestMadrassoInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockMadrassoGame)
	cfg := domain.DefaultMadrassoConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewMadrassoInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreMadrassoInteractor(t *testing.T) {
	tpMock := new(presenter.MockMadrassoPresenter)
	src := domain.NewDefaultMadrasso()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreMadrassoInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreMadrassoInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
