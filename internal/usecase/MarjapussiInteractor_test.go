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

const marjapussiMockOutput = `{"phase":0}`

func TestNewMarjapussiInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MarjapussiInteractor: g must not be nil", func() {
			usecase.NewMarjapussiInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMarjapussiGame)
		assert.PanicsWithValue(t, "MarjapussiInteractor: tp must not be nil", func() {
			usecase.NewMarjapussiInteractor(gameMock, nil)
		})
	})
}

func TestMarjapussiInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMarjapussiInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	cfg := domain.MarjapussiConfig{
		CpuDifficulty: domain.MarjapussiCpuDifficultyHard,
		TargetPoints:  500,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestMarjapussiInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	bad := domain.MarjapussiConfig{
		CpuDifficulty: domain.MarjapussiCpuDifficultyNormal,
		TargetPoints:  0,
	}
	assert.Equal(t, marjapussiMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestMarjapussiInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MarjapussiPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.MarjapussiPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestMarjapussiInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestMarjapussiInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestMarjapussiInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestMarjapussiInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestMarjapussiInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestMarjapussiInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(marjapussiMockOutput)
	gameMock := new(interfaces.MockMarjapussiGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, marjapussiMockOutput, ti.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestMarjapussiInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockMarjapussiGame)
	cfg := domain.DefaultMarjapussiConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewMarjapussiInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreMarjapussiInteractor(t *testing.T) {
	tpMock := new(presenter.MockMarjapussiPresenter)
	src := domain.NewDefaultMarjapussi()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreMarjapussiInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreMarjapussiInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
