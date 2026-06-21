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

const manilleMockOutput = `{"phase":0}`

func TestNewManilleInteractor_NilGuards(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ManilleInteractor: g must not be nil", func() {
			usecase.NewManilleInteractor(nil, mpMock)
		})
	})
	t.Run("panics when mp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockManilleGame)
		assert.PanicsWithValue(t, "ManilleInteractor: mp must not be nil", func() {
			usecase.NewManilleInteractor(gameMock, nil)
		})
	})
}

func TestManilleInteractor_Reset(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ManillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestManilleInteractor_ResetWithConfig(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	cfg := domain.ManilleConfig{
		CpuDifficulty: domain.ManilleCpuDifficultyHard,
		TargetPoints:  101,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ManillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestManilleInteractor_ResetWithConfigInvalid(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	// TargetPoints must be >= 1; 0 is invalid
	bad := domain.ManilleConfig{
		CpuDifficulty: domain.ManilleCpuDifficultyNormal,
		TargetPoints:  0,
	}
	assert.Equal(t, manilleMockOutput, mi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestManilleInteractor_PlayResolvesTrick(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase sequence:
	//   guardNotPlayable: GetGameEndFlag (false) + IsHumanTurn (true) — no GetPhase call
	//   post-PlayerPlay check: 1st GetPhase → TrickEnd → triggers ResolveTrick
	//   runCpuTurnsLoop: 2nd GetPhase → RoundEnd → exits immediately
	gameMock.On("GetPhase").Return(domain.ManillePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.ManillePhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestManilleInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ManillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestManilleInteractor_PlayError(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ManillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestManilleInteractor_PlayNotHumanTurn(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ManillePhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestManilleInteractor_NextTrick(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ManillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestManilleInteractor_NextRound(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.ManillePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestManilleInteractor_NextRoundGameEnded(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(manilleMockOutput)
	gameMock := new(interfaces.MockManilleGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, manilleMockOutput, mi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestManilleInteractor_GetConfigHintActionLog(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	mpMock.On("HintOutput", mock.Anything).Return("hint")
	mpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockManilleGame)
	cfg := domain.DefaultManilleConfig()
	gameMock.On("GetConfig").Return(cfg)

	mi := usecase.NewManilleInteractor(gameMock, mpMock)
	assert.Equal(t, cfg, mi.GetConfig())
	assert.Equal(t, "hint", mi.Hint())
	assert.Equal(t, "log", mi.ActionLog())
}

func TestRestoreManilleInteractor(t *testing.T) {
	mpMock := new(presenter.MockManillePresenter)
	src := domain.NewDefaultManille()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	mi, err := usecase.RestoreManilleInteractor(data, mpMock)
	assert.NoError(t, err)
	assert.NotNil(t, mi)

	_, err = usecase.RestoreManilleInteractor([]byte(`{`), mpMock)
	assert.Error(t, err)
}
