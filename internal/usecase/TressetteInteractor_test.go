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

const trMockOutput = `{"phase":0}`

func TestNewTressetteInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TressetteInteractor: g must not be nil", func() {
			usecase.NewTressetteInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTressetteGame)
		assert.PanicsWithValue(t, "TressetteInteractor: tp must not be nil", func() {
			usecase.NewTressetteInteractor(gameMock, nil)
		})
	})
}

func TestTressetteInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TressettePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTressetteInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	cfg := domain.TressetteConfig{CpuDifficulty: domain.TressetteCpuDifficultyHard, TargetPoints: 31}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TressettePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestTressetteInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	bad := domain.TressetteConfig{CpuDifficulty: 99, TargetPoints: 21}
	assert.Equal(t, trMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestTressetteInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 3).Return(nil)
	gameMock.On("GetPhase").Return(domain.TressettePhaseTrickEnd)
	gameMock.On("ResolveTrick").Return()

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.Play(3))
	gameMock.AssertCalled(t, "PlayerPlay", 3)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestTressetteInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.TressettePhasePlay)

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTressetteInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("boom"))

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTressetteInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestTressetteInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TressettePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestTressetteInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.TressettePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestTressetteInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trMockOutput)
	gameMock := new(interfaces.MockTressetteGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, trMockOutput, ti.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestTressetteInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockTressetteGame)
	cfg := domain.DefaultTressetteConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewTressetteInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreTressetteInteractor(t *testing.T) {
	tpMock := new(presenter.MockTressettePresenter)
	src := domain.NewDefaultTressette()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreTressetteInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreTressetteInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
