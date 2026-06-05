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

const gzMockOutput = `{"phase":0}`

func TestNewGongZhuInteractor_NilGuards(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "GongZhuInteractor: g must not be nil", func() {
			usecase.NewGongZhuInteractor(nil, gpMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockGongZhuGame)
		assert.PanicsWithValue(t, "GongZhuInteractor: gp must not be nil", func() {
			usecase.NewGongZhuInteractor(gameMock, nil)
		})
	})
}

func TestGongZhuInteractor_Reset(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("Reset").Return()

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestGongZhuInteractor_ResetWithConfig(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	cfg := domain.GongZhuConfig{CpuDifficulty: domain.GongZhuCpuDifficultyHard, PointLimit: 500}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestGongZhuInteractor_ResetWithConfigInvalid(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	// invalid config (bad difficulty) → validation fails, no SetConfig/Reset
	bad := domain.GongZhuConfig{CpuDifficulty: 99, PointLimit: 1000}
	assert.Equal(t, gzMockOutput, gi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestGongZhuInteractor_Expose(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerExpose", []int{0}).Return(nil)
	gameMock.On("CpuExpose").Return()
	gameMock.On("ExecuteExpose").Return()
	gameMock.On("GetPhase").Return(domain.GongZhuPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.Expose([]int{0}))
	gameMock.AssertCalled(t, "ExecuteExpose")
}

func TestGongZhuInteractor_ExposeError(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerExpose", mock.Anything).Return(errors.New("boom"))

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.Expose([]int{9}))
	gameMock.AssertNotCalled(t, "CpuExpose")
}

func TestGongZhuInteractor_ExposeGameEnded(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("GetGameEndFlag").Return(true)

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.Expose([]int{0}))
	gameMock.AssertNotCalled(t, "PlayerExpose")
}

func TestGongZhuInteractor_Play(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 3).Return(nil)
	gameMock.On("GetPhase").Return(domain.GongZhuPhaseTrickEnd)

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.Play(3))
	gameMock.AssertCalled(t, "PlayerPlay", 3)
}

func TestGongZhuInteractor_PlayNotHumanTurn(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestGongZhuInteractor_NextTrick(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GongZhuPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestGongZhuInteractor_NextRound(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestGongZhuInteractor_NextRoundGameEnded(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("Output", mock.Anything, mock.Anything).Return(gzMockOutput)
	gameMock := new(interfaces.MockGongZhuGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, gzMockOutput, gi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestGongZhuInteractor_GetConfigHintActionLog(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	gpMock.On("HintOutput", mock.Anything).Return("hint")
	gpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockGongZhuGame)
	cfg := domain.DefaultGongZhuConfig()
	gameMock.On("GetConfig").Return(cfg)

	gi := usecase.NewGongZhuInteractor(gameMock, gpMock)
	assert.Equal(t, cfg, gi.GetConfig())
	assert.Equal(t, "hint", gi.Hint())
	assert.Equal(t, "log", gi.ActionLog())
}

func TestRestoreGongZhuInteractor(t *testing.T) {
	gpMock := new(presenter.MockGongZhuPresenter)
	src := domain.NewDefaultGongZhu()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	gi, err := usecase.RestoreGongZhuInteractor(data, gpMock)
	assert.NoError(t, err)
	assert.NotNil(t, gi)

	_, err = usecase.RestoreGongZhuInteractor([]byte(`{`), gpMock)
	assert.Error(t, err)
}
