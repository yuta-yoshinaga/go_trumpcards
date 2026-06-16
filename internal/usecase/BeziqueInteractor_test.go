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

func TestNewBeziqueInteractor_NilGuards(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)

	t.Run("panics when b is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BeziqueInteractor: b must not be nil", func() {
			usecase.NewBeziqueInteractor(nil, bpMock)
		})
	})

	t.Run("panics when bp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBeziqueGame)
		assert.PanicsWithValue(t, "BeziqueInteractor: bp must not be nil", func() {
			usecase.NewBeziqueInteractor(gameMock, nil)
		})
	})
}

func TestBeziqueInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BeziquePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.Reset()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "Reset")
}

func TestBeziqueInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBeziqueGame)
	cfg := domain.DefaultBeziqueConfig()
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BeziquePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.ResetWithConfig(cfg)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestBeziqueInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	gameMock := new(interfaces.MockBeziqueGame)
	bpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	invalid := domain.BeziqueConfig{CpuDifficulty: 99, TargetScore: 1000}
	got := bi.ResetWithConfig(invalid)
	assert.Equal(t, "validation error", got)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestBeziqueInteractor_Play_Valid(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.BeziquePhasePlay)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.Play(1)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerPlay", 1)
}

func TestBeziqueInteractor_Play_Error(t *testing.T) {
	wantErr := errors.New("boom")
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 99).Return(wantErr)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.Play(99)
	assert.Equal(t, "error output", got)
}

func TestBeziqueInteractor_Play_GuardBlocksWhenGameEnded(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(true)
	bpMock.On("Output", gameMock, nil).Return("game ended")

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.Play(0)
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerPlay")
}

// TestBeziqueInteractor_Play_DrivesCpuMeldThenStopsAtHuman exercises the custom
// CPU loop: after the human plays and the CPU wins the trick (Meld phase, CPU
// turn), the loop must call CpuMeld and then stop once it becomes the human's
// turn again in the Play phase.
func TestBeziqueInteractor_Play_DrivesCpuMeldThenStopsAtHuman(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once() // guardNotPlayable
	gameMock.On("PlayerPlay", 0).Return(nil)
	// loop iter 1: Meld phase, CPU turn -> CpuMeld
	gameMock.On("GetPhase").Return(domain.BeziquePhaseMeld).Once()
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("CpuMeld").Return()
	// loop iter 2: Play phase, human turn -> stop
	gameMock.On("GetPhase").Return(domain.BeziquePhasePlay).Once()
	gameMock.On("IsHumanTurn").Return(true).Once()

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.Play(0)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "CpuMeld")
}

// TestBeziqueInteractor_Play_StopsAtRoundEnd verifies the loop does NOT
// auto-advance the RoundEnd phase (waits for the explicit NextRound command).
func TestBeziqueInteractor_Play_StopsAtRoundEnd(t *testing.T) {
	mockOutput := `{"phase":2}`
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once() // guardNotPlayable
	gameMock.On("PlayerPlay", 0).Return(nil)
	gameMock.On("GetPhase").Return(domain.BeziquePhaseRoundEnd).Once()

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.Play(0)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertNotCalled(t, "CpuPlay")
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestBeziqueInteractor_DeclareMeld_Valid(t *testing.T) {
	mockOutput := `{"phase":1}`
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDeclareMeld", 0).Return(nil)
	gameMock.On("GetPhase").Return(domain.BeziquePhasePlay)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.DeclareMeld(0)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerDeclareMeld", 0)
}

func TestBeziqueInteractor_DeclareMeld_Error(t *testing.T) {
	wantErr := errors.New("no meld")
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDeclareMeld", 5).Return(wantErr)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.DeclareMeld(5)
	assert.Equal(t, "error output", got)
}

func TestBeziqueInteractor_DeclareMeld_GuardBlocksWhenGameEnded(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(true)
	bpMock.On("Output", gameMock, nil).Return("game ended")

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.DeclareMeld(0)
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerDeclareMeld")
}

func TestBeziqueInteractor_SkipMeld_Valid(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerSkipMeld").Return(nil)
	gameMock.On("GetPhase").Return(domain.BeziquePhasePlay)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.SkipMeld()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerSkipMeld")
}

func TestBeziqueInteractor_SkipMeld_Error(t *testing.T) {
	wantErr := errors.New("wrong phase")
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerSkipMeld").Return(wantErr)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.SkipMeld()
	assert.Equal(t, "error output", got)
}

func TestBeziqueInteractor_SkipMeld_GuardBlocksWhenGameEnded(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("GetGameEndFlag").Return(true)
	bpMock.On("Output", gameMock, nil).Return("game ended")

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.SkipMeld()
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerSkipMeld")
}

func TestBeziqueInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBeziquePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BeziquePhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.NextRound()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "NextRound")
}

func TestBeziqueInteractor_NextRound_BlocksOnGameEnd(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	gameMock := new(interfaces.MockBeziqueGame)
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)
	bpMock.On("Output", gameMock, nil).Return("ended")

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	got := bi.NextRound()
	assert.Equal(t, "ended", got)
}

func TestBeziqueInteractor_GetConfig(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	gameMock := new(interfaces.MockBeziqueGame)
	cfg := domain.DefaultBeziqueConfig()
	gameMock.On("GetConfig").Return(cfg)

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	assert.Equal(t, cfg, bi.GetConfig())
}

func TestBeziqueInteractor_Hint(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	gameMock := new(interfaces.MockBeziqueGame)
	bpMock.On("HintOutput", gameMock).Return("hint")

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	assert.Equal(t, "hint", bi.Hint())
}

func TestBeziqueInteractor_ActionLog(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	gameMock := new(interfaces.MockBeziqueGame)
	bpMock.On("ActionLogOutput", gameMock).Return("log")

	bi := usecase.NewBeziqueInteractor(gameMock, bpMock)
	assert.Equal(t, "log", bi.ActionLog())
}

func TestBeziqueInteractor_Snapshot_RoundtripsViaRealGame(t *testing.T) {
	bpMock := new(presenter.MockBeziquePresenter)
	game := domain.NewDefaultBezique()
	game.Reset()

	bi := usecase.NewBeziqueInteractor(game, bpMock)
	data, err := bi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	bi2, err := usecase.RestoreBeziqueInteractor(data, bpMock)
	assert.NoError(t, err)
	assert.Equal(t, game.GetPhase(), bi2.Game.GetPhase())
}
