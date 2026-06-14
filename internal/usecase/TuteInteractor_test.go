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

const tuteMockOutput = `{"phase":0}`

func TestNewTuteInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TuteInteractor: g must not be nil", func() {
			usecase.NewTuteInteractor(nil, tpMock)
		})
	})
	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTuteGame)
		assert.PanicsWithValue(t, "TuteInteractor: tp must not be nil", func() {
			usecase.NewTuteInteractor(gameMock, nil)
		})
	})
}

func TestTuteInteractor_Reset(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TutePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestTuteInteractor_ResetWithConfig(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	cfg := domain.TuteConfig{
		CpuDifficulty: domain.TuteCpuDifficultyHard,
		TargetPoints:  200,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TutePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestTuteInteractor_ResetWithConfigInvalid(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	// TargetPoints must be >= 1; 0 is invalid
	bad := domain.TuteConfig{
		CpuDifficulty: domain.TuteCpuDifficultyNormal,
		TargetPoints:  0,
	}
	assert.Equal(t, tuteMockOutput, ti.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestTuteInteractor_PlayResolvesTrick(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase sequence:
	//   guardNotPlayable: GetGameEndFlag (false) + IsHumanTurn (true) — no GetPhase call
	//   post-PlayerPlay check: 1st GetPhase → TrickEnd → triggers ResolveTrick
	//   runCpuTurnsLoop: 2nd GetPhase → RoundEnd → exits immediately
	gameMock.On("GetPhase").Return(domain.TutePhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.TutePhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestTuteInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TutePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTuteInteractor_PlayError(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TutePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestTuteInteractor_PlayNotHumanTurn(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TutePhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestTuteInteractor_DeclareMarriage(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)

	t.Run("marriage success", func(t *testing.T) {
		gameMock := new(interfaces.MockTuteGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclareMarriage", 1).Return(nil)

		ti := usecase.NewTuteInteractor(gameMock, tpMock)
		assert.Equal(t, tuteMockOutput, ti.DeclareMarriage(1))
		gameMock.AssertCalled(t, "PlayerDeclareMarriage", 1)
	})

	t.Run("marriage error", func(t *testing.T) {
		gameMock := new(interfaces.MockTuteGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclareMarriage", 3).Return(errors.New("cannot declare"))

		ti := usecase.NewTuteInteractor(gameMock, tpMock)
		assert.Equal(t, tuteMockOutput, ti.DeclareMarriage(3))
	})

	t.Run("game ended blocks marriage", func(t *testing.T) {
		gameMock := new(interfaces.MockTuteGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ti := usecase.NewTuteInteractor(gameMock, tpMock)
		assert.Equal(t, tuteMockOutput, ti.DeclareMarriage(1))
		gameMock.AssertNotCalled(t, "PlayerDeclareMarriage")
	})
}

func TestTuteInteractor_DeclareTute(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)

	t.Run("tute success", func(t *testing.T) {
		gameMock := new(interfaces.MockTuteGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclareTute").Return(nil)

		ti := usecase.NewTuteInteractor(gameMock, tpMock)
		assert.Equal(t, tuteMockOutput, ti.DeclareTute())
		gameMock.AssertCalled(t, "PlayerDeclareTute")
	})

	t.Run("tute error", func(t *testing.T) {
		gameMock := new(interfaces.MockTuteGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerDeclareTute").Return(errors.New("cannot tute"))

		ti := usecase.NewTuteInteractor(gameMock, tpMock)
		assert.Equal(t, tuteMockOutput, ti.DeclareTute())
	})

	t.Run("game ended blocks tute", func(t *testing.T) {
		gameMock := new(interfaces.MockTuteGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ti := usecase.NewTuteInteractor(gameMock, tpMock)
		assert.Equal(t, tuteMockOutput, ti.DeclareTute())
		gameMock.AssertNotCalled(t, "PlayerDeclareTute")
	})
}

func TestTuteInteractor_NextTrick(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TutePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestTuteInteractor_NextRound(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.TutePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestTuteInteractor_NextRoundGameEnded(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(tuteMockOutput)
	gameMock := new(interfaces.MockTuteGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, tuteMockOutput, ti.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestTuteInteractor_GetConfigHintActionLog(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint")
	tpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockTuteGame)
	cfg := domain.DefaultTuteConfig()
	gameMock.On("GetConfig").Return(cfg)

	ti := usecase.NewTuteInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

func TestRestoreTuteInteractor(t *testing.T) {
	tpMock := new(presenter.MockTutePresenter)
	src := domain.NewDefaultTute()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	ti, err := usecase.RestoreTuteInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, ti)

	_, err = usecase.RestoreTuteInteractor([]byte(`{`), tpMock)
	assert.Error(t, err)
}
