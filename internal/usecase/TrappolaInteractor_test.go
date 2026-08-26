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

const trapMockOutput = `{"phase":0}`

func TestNewTrappolaInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TrappolaInteractor: g must not be nil", func() {
			usecase.NewTrappolaInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTrappolaGame)
		assert.PanicsWithValue(t, "TrappolaInteractor: tp must not be nil", func() {
			usecase.NewTrappolaInteractor(gameMock, nil)
		})
	})
}

func TestTrappolaInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TrappolaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTrappolaInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	cfg := domain.TrappolaConfig{CpuDifficulty: domain.TrappolaCpuDifficultyHard, TargetPoints: 31}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TrappolaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestTrappolaInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	bad := domain.TrappolaConfig{CpuDifficulty: 99, TargetPoints: 21}
	assert.Equal(t, trapMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestTrappolaInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 3).Return(nil)
	gameMock.On("GetPhase").Return(domain.TrappolaPhaseTrickEnd)
	gameMock.On("ResolveTrick").Return()

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.Play(3))
	gameMock.AssertCalled(t, "PlayerPlay", 3)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestTrappolaInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.TrappolaPhasePlay)

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTrappolaInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("boom"))

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTrappolaInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestTrappolaInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TrappolaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestTrappolaInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.TrappolaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestTrappolaInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(trapMockOutput)
	gameMock := new(interfaces.MockTrappolaGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, trapMockOutput, ti.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestTrappolaInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockTrappolaGame)
	cfg := domain.DefaultTrappolaConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewTrappolaInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreTrappolaInteractor(t *testing.T) {
	tpMock := new(presenter.MockTrappolaPresenter)
	src := domain.NewDefaultTrappola()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreTrappolaInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreTrappolaInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
