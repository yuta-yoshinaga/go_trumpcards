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

const spoilFiveMockOutput = `{"phase":0}`

func TestNewSpoilFiveInteractor_NilGuards(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SpoilFiveInteractor: g must not be nil", func() {
			usecase.NewSpoilFiveInteractor(nil, mpMock)
		})
	})
	t.Run("panics when mp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSpoilFiveGame)
		assert.PanicsWithValue(t, "SpoilFiveInteractor: mp must not be nil", func() {
			usecase.NewSpoilFiveInteractor(gameMock, nil)
		})
	})
}

func TestSpoilFiveInteractor_Reset(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSpoilFiveInteractor_ResetWithConfig(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	cfg := domain.SpoilFiveConfig{CpuDifficulty: domain.SpoilFiveCpuDifficultyHard, TargetPoints: 30}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestSpoilFiveInteractor_ResetWithConfigInvalid(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	// CpuDifficulty must be 0-2; 9 is invalid
	bad := domain.SpoilFiveConfig{CpuDifficulty: domain.SpoilFiveCpuDifficulty(9), TargetPoints: 30}
	assert.Equal(t, spoilFiveMockOutput, mi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestSpoilFiveInteractor_PlayResolvesTrick(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SpoilFivePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.SpoilFivePhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestSpoilFiveInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSpoilFiveInteractor_PlayError(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSpoilFiveInteractor_PlayNotHumanTurn(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestSpoilFiveInteractor_NextTrick(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestSpoilFiveInteractor_NextRound(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.SpoilFivePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestSpoilFiveInteractor_NextRoundGameEnded(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(spoilFiveMockOutput)
	gameMock := new(interfaces.MockSpoilFiveGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, spoilFiveMockOutput, mi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestSpoilFiveInteractor_GetConfigHintActionLog(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	mpMock.On("HintOutput", mock.Anything).Return("hint")
	mpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockSpoilFiveGame)
	cfg := domain.DefaultSpoilFiveConfig()
	gameMock.On("GetConfig").Return(cfg)

	mi := usecase.NewSpoilFiveInteractor(gameMock, mpMock)
	assert.Equal(t, cfg, mi.GetConfig())
	assert.Equal(t, "hint", mi.Hint())
	assert.Equal(t, "log", mi.ActionLog())
}

func TestRestoreSpoilFiveInteractor(t *testing.T) {
	mpMock := new(presenter.MockSpoilFivePresenter)
	src := domain.NewDefaultSpoilFive()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	mi, err := usecase.RestoreSpoilFiveInteractor(data, mpMock)
	assert.NoError(t, err)
	assert.NotNil(t, mi)

	_, err = usecase.RestoreSpoilFiveInteractor([]byte(`{`), mpMock)
	assert.Error(t, err)
}
