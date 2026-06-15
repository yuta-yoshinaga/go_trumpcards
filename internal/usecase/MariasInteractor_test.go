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

const mariasMockOutput = `{"phase":0}`

func TestNewMariasInteractor_NilGuards(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MariasInteractor: g must not be nil", func() {
			usecase.NewMariasInteractor(nil, mpMock)
		})
	})
	t.Run("panics when mp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMariasGame)
		assert.PanicsWithValue(t, "MariasInteractor: mp must not be nil", func() {
			usecase.NewMariasInteractor(gameMock, nil)
		})
	})
}

func TestMariasInteractor_Reset(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MariasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMariasInteractor_ResetWithConfig(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	cfg := domain.MariasConfig{
		CpuDifficulty: domain.MariasCpuDifficultyHard,
		TargetPoints:  15,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MariasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestMariasInteractor_ResetWithConfigInvalid(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	bad := domain.MariasConfig{
		CpuDifficulty: domain.MariasCpuDifficultyNormal,
		TargetPoints:  0,
	}
	assert.Equal(t, mariasMockOutput, mi.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestMariasInteractor_PlayResolvesTrick(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MariasPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.MariasPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestMariasInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MariasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestMariasInteractor_PlayError(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MariasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestMariasInteractor_PlayNotHumanTurn(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MariasPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestMariasInteractor_NextTrick(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MariasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestMariasInteractor_NextRound(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.MariasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestMariasInteractor_NextRoundGameEnded(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("Output", mock.Anything, mock.Anything).Return(mariasMockOutput)
	gameMock := new(interfaces.MockMariasGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, mariasMockOutput, mi.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestMariasInteractor_GetConfigHintActionLog(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	mpMock.On("HintOutput", mock.Anything).Return("hint")
	mpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockMariasGame)
	cfg := domain.DefaultMariasConfig()
	gameMock.On("GetConfig").Return(cfg)

	mi := usecase.NewMariasInteractor(gameMock, mpMock)
	assert.Equal(t, cfg, mi.GetConfig())
	assert.Equal(t, "hint", mi.Hint())
	assert.Equal(t, "log", mi.ActionLog())
}

func TestRestoreMariasInteractor(t *testing.T) {
	mpMock := new(presenter.MockMariasPresenter)
	src := domain.NewDefaultMarias()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	mi, err := usecase.RestoreMariasInteractor(data, mpMock)
	assert.NoError(t, err)
	assert.NotNil(t, mi)

	_, err = usecase.RestoreMariasInteractor([]byte(`{`), mpMock)
	assert.Error(t, err)
}
