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

const sedmaMockOutput = `{"phase":0}`

func TestNewSedmaInteractor_NilGuards(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SedmaInteractor: g must not be nil", func() {
			usecase.NewSedmaInteractor(nil, mpMock)
		})
	})
	t.Run("panics when mp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSedmaGame)
		assert.PanicsWithValue(t, "SedmaInteractor: mp must not be nil", func() {
			usecase.NewSedmaInteractor(gameMock, nil)
		})
	})
}

func TestSedmaInteractor_Reset(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SedmaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSedmaInteractor_ResetWithConfig(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	cfg := domain.SedmaConfig{
		CpuDifficulty: domain.SedmaCpuDifficultyHard,
		TargetPoints:  101,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SedmaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestSedmaInteractor_ResetWithConfigInvalid(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	// TargetPoints must be >= 1; 0 is invalid
	bad := domain.SedmaConfig{
		CpuDifficulty: domain.SedmaCpuDifficultyNormal,
		TargetPoints:  0,
	}
	assert.Equal(t, sedmaMockOutput, mi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestSedmaInteractor_PlayResolvesTrick(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SedmaPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.SedmaPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestSedmaInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SedmaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSedmaInteractor_PlayError(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SedmaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSedmaInteractor_PlayNotHumanTurn(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SedmaPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestSedmaInteractor_NextTrick(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SedmaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestSedmaInteractor_NextRound(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.SedmaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestSedmaInteractor_NextRoundGameEnded(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(sedmaMockOutput)
	gameMock := new(interfaces.MockSedmaGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, sedmaMockOutput, mi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestSedmaInteractor_GetConfigHintActionLog(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	mpMock.On("HintOutput", mock.Anything).Return("hint")
	mpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockSedmaGame)
	cfg := domain.DefaultSedmaConfig()
	gameMock.On("GetConfig").Return(cfg)

	mi := usecase.NewSedmaInteractor(gameMock, mpMock)
	assert.Equal(t, cfg, mi.GetConfig())
	assert.Equal(t, "hint", mi.Hint())
	assert.Equal(t, "log", mi.ActionLog())
}

func TestRestoreSedmaInteractor(t *testing.T) {
	mpMock := new(presenter.MockSedmaPresenter)
	src := domain.NewDefaultSedma()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	mi, err := usecase.RestoreSedmaInteractor(data, mpMock)
	assert.NoError(t, err)
	assert.NotNil(t, mi)

	_, err = usecase.RestoreSedmaInteractor([]byte(`{`), mpMock)
	assert.Error(t, err)
}
