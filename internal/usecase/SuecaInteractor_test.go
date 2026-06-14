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

const suecaMockOutput = `{"phase":0}`

func TestNewSuecaInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SuecaInteractor: g must not be nil", func() {
			usecase.NewSuecaInteractor(nil, spMock)
		})
	})
	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSuecaGame)
		assert.PanicsWithValue(t, "SuecaInteractor: sp must not be nil", func() {
			usecase.NewSuecaInteractor(gameMock, nil)
		})
	})
}

func TestSuecaInteractor_Reset(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SuecaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestSuecaInteractor_ResetWithConfig(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	cfg := domain.SuecaConfig{
		CpuDifficulty:    domain.SuecaCpuDifficultyHard,
		TargetGamePoints: 8,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SuecaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestSuecaInteractor_ResetWithConfigInvalid(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	// TargetGamePoints must be >= 1; 0 is invalid
	bad := domain.SuecaConfig{
		CpuDifficulty:    domain.SuecaCpuDifficultyNormal,
		TargetGamePoints: 0,
	}
	assert.Equal(t, suecaMockOutput, si.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestSuecaInteractor_PlayResolvesTrick(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase sequence:
	//   guardNotPlayable: GetGameEndFlag (false) + IsHumanTurn (true) — no GetPhase call
	//   post-PlayerPlay check: 1st GetPhase → TrickEnd → triggers ResolveTrick
	//   runCpuTurnsLoop: 2nd GetPhase → RoundEnd → exits immediately
	gameMock.On("GetPhase").Return(domain.SuecaPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.SuecaPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestSuecaInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SuecaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSuecaInteractor_PlayError(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SuecaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestSuecaInteractor_PlayNotHumanTurn(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SuecaPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestSuecaInteractor_NextTrick(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SuecaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestSuecaInteractor_NextRound(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.SuecaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestSuecaInteractor_NextRoundGameEnded(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(suecaMockOutput)
	gameMock := new(interfaces.MockSuecaGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, suecaMockOutput, si.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestSuecaInteractor_GetConfigHintActionLog(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	spMock.On("HintOutput", mock.Anything).Return("hint")
	spMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockSuecaGame)
	cfg := domain.DefaultSuecaConfig()
	gameMock.On("GetConfig").Return(cfg)

	si := usecase.NewSuecaInteractor(gameMock, spMock)
	assert.Equal(t, cfg, si.GetConfig())
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

func TestRestoreSuecaInteractor(t *testing.T) {
	spMock := new(presenter.MockSuecaPresenter)
	src := domain.NewDefaultSueca()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	si, err := usecase.RestoreSuecaInteractor(data, spMock)
	assert.NoError(t, err)
	assert.NotNil(t, si)

	_, err = usecase.RestoreSuecaInteractor([]byte(`{`), spMock)
	assert.Error(t, err)
}
