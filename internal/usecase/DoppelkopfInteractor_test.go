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

const dkMockOutput = `{"phase":0}`

func TestNewDoppelkopfInteractor_NilGuards(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DoppelkopfInteractor: g must not be nil", func() {
			usecase.NewDoppelkopfInteractor(nil, dpMock)
		})
	})
	t.Run("panics when dp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockDoppelkopfGame)
		assert.PanicsWithValue(t, "DoppelkopfInteractor: dp must not be nil", func() {
			usecase.NewDoppelkopfInteractor(gameMock, nil)
		})
	})
}

func TestDoppelkopfInteractor_Reset(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestDoppelkopfInteractor_ResetWithConfig(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	cfg := domain.DoppelkopfConfig{
		CpuDifficulty: domain.DoppelkopfCpuDifficultyHard,
		BaseChips:     2,
		StartChips:    20,
		TargetChips:   60,
	}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestDoppelkopfInteractor_ResetWithConfigInvalid(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	// TargetChips must be > StartChips; StartChips=10, TargetChips=5 is invalid
	bad := domain.DoppelkopfConfig{
		CpuDifficulty: domain.DoppelkopfCpuDifficultyNormal,
		BaseChips:     2,
		StartChips:    10,
		TargetChips:   5,
	}
	assert.Equal(t, dkMockOutput, di.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "Reset")
}

func TestDoppelkopfInteractor_PlayResolvesTrick(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	gameMock.On("GetGameEndFlag").Return(false)
	// Phase sequence:
	//   guardNotPlayable: GetGameEndFlag (false) + IsHumanTurn (true) — no GetPhase call
	//   post-PlayerPlay check: 1st GetPhase → TrickEnd → triggers ResolveTrick
	//   runCpuTurnsLoop: 2nd GetPhase → RoundEnd → exits immediately
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhaseTrickEnd).Once()
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhaseRoundEnd)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 2).Return(nil)
	gameMock.On("ResolveTrick").Return()

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.Play(2))
	gameMock.AssertCalled(t, "PlayerPlay", 2)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestDoppelkopfInteractor_PlayNoResolveWhenNotTrickEnd(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.Play(1))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestDoppelkopfInteractor_PlayError(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 9).Return(errors.New("invalid card"))

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.Play(9))
	gameMock.AssertNotCalled(t, "ResolveTrick")
}

func TestDoppelkopfInteractor_PlayNotHumanTurn(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(false)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.Play(0))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestDoppelkopfInteractor_Announce(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)

	t.Run("announce success", func(t *testing.T) {
		gameMock := new(interfaces.MockDoppelkopfGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerAnnounce").Return(nil)

		di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
		assert.Equal(t, dkMockOutput, di.Announce())
		gameMock.AssertCalled(t, "PlayerAnnounce")
	})

	t.Run("announce error", func(t *testing.T) {
		gameMock := new(interfaces.MockDoppelkopfGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("PlayerAnnounce").Return(errors.New("cannot announce"))

		di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
		assert.Equal(t, dkMockOutput, di.Announce())
	})

	t.Run("game ended blocks announce", func(t *testing.T) {
		gameMock := new(interfaces.MockDoppelkopfGame)
		gameMock.On("GetGameEndFlag").Return(true)

		di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
		assert.Equal(t, dkMockOutput, di.Announce())
		gameMock.AssertNotCalled(t, "PlayerAnnounce")
	})
}

func TestDoppelkopfInteractor_NextTrick(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.NextTrick())
	gameMock.AssertCalled(t, "NextTrick")
}

func TestDoppelkopfInteractor_NextRound(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.DoppelkopfPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.NextRound())
	gameMock.AssertCalled(t, "NextRound")
}

func TestDoppelkopfInteractor_NextRoundGameEnded(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(dkMockOutput)
	gameMock := new(interfaces.MockDoppelkopfGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, dkMockOutput, di.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestDoppelkopfInteractor_GetConfigHintActionLog(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	dpMock.On("HintOutput", mock.Anything).Return("hint")
	dpMock.On("ActionLogOutput", mock.Anything).Return("log")
	gameMock := new(interfaces.MockDoppelkopfGame)
	cfg := domain.DefaultDoppelkopfConfig()
	gameMock.On("GetConfig").Return(cfg)

	di := usecase.NewDoppelkopfInteractor(gameMock, dpMock)
	assert.Equal(t, cfg, di.GetConfig())
	assert.Equal(t, "hint", di.Hint())
	assert.Equal(t, "log", di.ActionLog())
}

func TestRestoreDoppelkopfInteractor(t *testing.T) {
	dpMock := new(presenter.MockDoppelkopfPresenter)
	src := domain.NewDefaultDoppelkopf()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	di, err := usecase.RestoreDoppelkopfInteractor(data, dpMock)
	assert.NoError(t, err)
	assert.NotNil(t, di)

	_, err = usecase.RestoreDoppelkopfInteractor([]byte(`{`), dpMock)
	assert.Error(t, err)
}
