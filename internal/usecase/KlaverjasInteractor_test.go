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

const klaverjasMockOutput = `{"phase":0}`

func TestNewKlaverjasInteractor_NilGuards(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KlaverjasInteractor: g must not be nil", func() {
			usecase.NewKlaverjasInteractor(nil, kpMock)
		})
	})
	t.Run("panics when kp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKlaverjasGame)
		assert.PanicsWithValue(t, "KlaverjasInteractor: kp must not be nil", func() {
			usecase.NewKlaverjasInteractor(gameMock, nil)
		})
	})
}

func TestKlaverjasInteractor_Reset(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestKlaverjasInteractor_ResetWithConfig(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	cfg := domain.KlaverjasConfig{
		CpuDifficulty: domain.KlaverjasCpuDifficultyHard,
		TargetPoints:  1501,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestKlaverjasInteractor_ResetWithConfigInvalid(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	// TargetPoints must be >= 1; 0 is invalid
	bad := domain.KlaverjasConfig{
		CpuDifficulty: domain.KlaverjasCpuDifficultyNormal,
		TargetPoints:  0,
	}
	assert.Equal(t, klaverjasMockOutput, ki.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestKlaverjasInteractor_PlayResolvesTrick(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase sequence:
	//   guardNotPlayable: GetGameEndFlag (false) + IsHumanTurn (true) — no GetPhase call
	//   post-PlayerPlay check: 1st GetPhase → TrickEnd → triggers ResolveTrick
	//   runCpuTurnsLoop: 2nd GetPhase → RoundEnd → exits immediately
	gameMock.On("GetPhase").Return(domain.KlaverjasPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.KlaverjasPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestKlaverjasInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestKlaverjasInteractor_PlayError(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestKlaverjasInteractor_PlayNotHumanTurn(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestKlaverjasInteractor_NextTrick(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestKlaverjasInteractor_NextRound(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.KlaverjasPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestKlaverjasInteractor_NextRoundGameEnded(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("Output", mock.Anything, mock.Anything).Return(klaverjasMockOutput)
	gameMock := new(interfaces.MockKlaverjasGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, klaverjasMockOutput, ki.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestKlaverjasInteractor_GetConfigHintActionLog(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	kpMock.On("HintOutput", mock.Anything).Return("hint")
	kpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockKlaverjasGame)
	cfg := domain.DefaultKlaverjasConfig()
	gameMock.On("GetConfig").Return(cfg)

	ki := usecase.NewKlaverjasInteractor(gameMock, kpMock)
	assert.Equal(t, cfg, ki.GetConfig())
	assert.Equal(t, "hint", ki.Hint())
	assert.Equal(t, "log", ki.ActionLog())
}

func TestRestoreKlaverjasInteractor(t *testing.T) {
	kpMock := new(presenter.MockKlaverjasPresenter)
	src := domain.NewDefaultKlaverjas()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ki, err := usecase.RestoreKlaverjasInteractor(data, kpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ki)

	_, err = usecase.RestoreKlaverjasInteractor([]byte(`{`), kpMock)
	assert.Error(t, err)
}
