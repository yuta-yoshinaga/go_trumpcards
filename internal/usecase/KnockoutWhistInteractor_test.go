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

const knockoutWhistMockOutput = `{"phase":0}`

func TestNewKnockoutWhistInteractor_NilGuards(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KnockoutWhistInteractor: g must not be nil", func() {
			usecase.NewKnockoutWhistInteractor(nil, mpMock)
		})
	})
	t.Run("panics when mp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKnockoutWhistGame)
		assert.PanicsWithValue(t, "KnockoutWhistInteractor: mp must not be nil", func() {
			usecase.NewKnockoutWhistInteractor(gameMock, nil)
		})
	})
}

func TestKnockoutWhistInteractor_Reset(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestKnockoutWhistInteractor_ResetWithConfig(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	cfg := domain.KnockoutWhistConfig{CpuDifficulty: domain.KnockoutWhistCpuDifficultyHard}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestKnockoutWhistInteractor_ResetWithConfigInvalid(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	// CpuDifficulty must be 0-2; 9 is invalid
	bad := domain.KnockoutWhistConfig{CpuDifficulty: domain.KnockoutWhistCpuDifficulty(9)}
	assert.Equal(t, knockoutWhistMockOutput, mi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestKnockoutWhistInteractor_PlayResolvesTrick(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestKnockoutWhistInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestKnockoutWhistInteractor_PlayError(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestKnockoutWhistInteractor_PlayNotHumanTurn(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestKnockoutWhistInteractor_NextTrick(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestKnockoutWhistInteractor_NextRound(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.KnockoutWhistPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestKnockoutWhistInteractor_NextRoundGameEnded(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(knockoutWhistMockOutput)
	gameMock := new(interfaces.MockKnockoutWhistGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, knockoutWhistMockOutput, mi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestKnockoutWhistInteractor_GetConfigHintActionLog(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	mpMock.On("HintOutput", mock.Anything).Return("hint")
	mpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockKnockoutWhistGame)
	cfg := domain.DefaultKnockoutWhistConfig()
	gameMock.On("GetConfig").Return(cfg)

	mi := usecase.NewKnockoutWhistInteractor(gameMock, mpMock)
	assert.Equal(t, cfg, mi.GetConfig())
	assert.Equal(t, "hint", mi.Hint())
	assert.Equal(t, "log", mi.ActionLog())
}

func TestRestoreKnockoutWhistInteractor(t *testing.T) {
	mpMock := new(presenter.MockKnockoutWhistPresenter)
	src := domain.NewDefaultKnockoutWhist()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	mi, err := usecase.RestoreKnockoutWhistInteractor(data, mpMock)
	assert.NoError(t, err)
	assert.NotNil(t, mi)

	_, err = usecase.RestoreKnockoutWhistInteractor([]byte(`{`), mpMock)
	assert.Error(t, err)
}
