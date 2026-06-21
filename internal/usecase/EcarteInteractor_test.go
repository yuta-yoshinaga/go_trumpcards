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

func TestNewEcarteInteractor_NilGuards(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)

	t.Run("panics when b is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "EcarteInteractor: b must not be nil", func() {
			usecase.NewEcarteInteractor(nil, epMock)
		})
	})

	t.Run("panics when ep is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockEcarteGame)
		assert.PanicsWithValue(t, "EcarteInteractor: ep must not be nil", func() {
			usecase.NewEcarteInteractor(gameMock, nil)
		})
	})
}

func TestEcarteInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.EcartePhaseExchange)
	gameMock.On("IsHumanTurn").Return(true)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Reset()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "Reset")
}

func TestEcarteInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	cfg := domain.DefaultEcarteConfig()
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.EcartePhaseExchange)
	gameMock.On("IsHumanTurn").Return(true)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.ResetWithConfig(cfg)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestEcarteInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	gameMock := new(interfaces.MockEcarteGame)
	epMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	invalid := domain.EcarteConfig{CpuDifficulty: 99, TargetScore: 5}
	got := ei.ResetWithConfig(invalid)
	assert.Equal(t, "validation error", got)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestEcarteInteractor_Propose_Valid(t *testing.T) {
	mockOutput := `{"phase":0}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPropose").Return(nil)
	gameMock.On("GetPhase").Return(domain.EcartePhaseExchange)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Propose()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerPropose")
}

func TestEcarteInteractor_Propose_Error(t *testing.T) {
	wantErr := errors.New("boom")
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPropose").Return(wantErr)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Propose()
	assert.Equal(t, "error output", got)
}

func TestEcarteInteractor_Propose_GuardBlocksWhenGameEnded(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(true)
	epMock.On("Output", gameMock, nil).Return("game ended")

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Propose()
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerPropose")
}

func TestEcarteInteractor_Stand_Valid(t *testing.T) {
	mockOutput := `{"phase":1}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerStand").Return(nil)
	gameMock.On("GetPhase").Return(domain.EcartePhasePlay)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Stand()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerStand")
}

func TestEcarteInteractor_Stand_Error(t *testing.T) {
	wantErr := errors.New("wrong phase")
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerStand").Return(wantErr)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Stand()
	assert.Equal(t, "error output", got)
}

func TestEcarteInteractor_Respond_Valid(t *testing.T) {
	mockOutput := `{"phase":0}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerRespond", true).Return(nil)
	gameMock.On("GetPhase").Return(domain.EcartePhaseExchange)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Respond(true)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerRespond", true)
}

func TestEcarteInteractor_Respond_Error(t *testing.T) {
	wantErr := errors.New("not your turn")
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerRespond", false).Return(wantErr)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Respond(false)
	assert.Equal(t, "error output", got)
}

func TestEcarteInteractor_Respond_GuardBlocksWhenGameEnded(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(true)
	epMock.On("Output", gameMock, nil).Return("game ended")

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Respond(true)
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerRespond")
}

func TestEcarteInteractor_Discard_Valid(t *testing.T) {
	mockOutput := `{"phase":0}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDiscard", []int{0, 2}).Return(nil)
	gameMock.On("GetPhase").Return(domain.EcartePhaseExchange)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Discard([]int{0, 2})
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerDiscard", []int{0, 2})
}

func TestEcarteInteractor_Discard_Error(t *testing.T) {
	wantErr := errors.New("invalid discard")
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDiscard", []int{9}).Return(wantErr)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Discard([]int{9})
	assert.Equal(t, "error output", got)
}

func TestEcarteInteractor_Play_Valid(t *testing.T) {
	mockOutput := `{"phase":1}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.EcartePhasePlay)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Play(1)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerPlay", 1)
}

func TestEcarteInteractor_Play_Error(t *testing.T) {
	wantErr := errors.New("boom")
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 99).Return(wantErr)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Play(99)
	assert.Equal(t, "error output", got)
}

func TestEcarteInteractor_Play_GuardBlocksWhenGameEnded(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(true)
	epMock.On("Output", gameMock, nil).Return("game ended")

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Play(0)
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerPlay")
}

// TestEcarteInteractor_Stand_DrivesCpuExchangeThenStopsAtHumanPlay exercises the
// custom CPU loop: after the human stands and play starts, the CPU (now on lead)
// must be driven via CpuPlay until it becomes the human's turn in the Play phase.
func TestEcarteInteractor_Stand_DrivesCpuExchangeThenStopsAtHumanPlay(t *testing.T) {
	mockOutput := `{"phase":1}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once() // guardNotPlayable
	gameMock.On("PlayerStand").Return(nil)
	// loop iter 1: Play phase, CPU turn -> CpuPlay
	gameMock.On("GetPhase").Return(domain.EcartePhasePlay).Once()
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("CpuPlay").Return()
	// loop iter 2: Play phase, human turn -> stop
	gameMock.On("GetPhase").Return(domain.EcartePhasePlay).Once()
	gameMock.On("IsHumanTurn").Return(true).Once()

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Stand()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "CpuPlay")
}

// TestEcarteInteractor_Propose_DrivesCpuExchange verifies that in the exchange
// phase the loop calls CpuExchange while the current player is the CPU and then
// stops once it becomes the human's turn again (still in the exchange phase).
func TestEcarteInteractor_Propose_DrivesCpuExchange(t *testing.T) {
	mockOutput := `{"phase":0}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once() // guardNotPlayable
	gameMock.On("PlayerPropose").Return(nil)
	// loop iter 1: Exchange phase, CPU turn -> CpuExchange (dealer responds)
	gameMock.On("GetPhase").Return(domain.EcartePhaseExchange).Once()
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("CpuExchange").Return()
	// loop iter 2: Exchange phase, human turn -> stop (elder discards)
	gameMock.On("GetPhase").Return(domain.EcartePhaseExchange).Once()
	gameMock.On("IsHumanTurn").Return(true).Once()

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Propose()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "CpuExchange")
}

// TestEcarteInteractor_Play_StopsAtRoundEnd verifies the loop does NOT
// auto-advance the RoundEnd phase (waits for the explicit NextRound command).
func TestEcarteInteractor_Play_StopsAtRoundEnd(t *testing.T) {
	mockOutput := `{"phase":2}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once() // guardNotPlayable
	gameMock.On("PlayerPlay", 0).Return(nil)
	gameMock.On("GetPhase").Return(domain.EcartePhaseRoundEnd).Once()

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.Play(0)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertNotCalled(t, "CpuPlay")
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestEcarteInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`
	epMock := new(presenter.MockEcartePresenter)
	epMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.EcartePhaseExchange)
	gameMock.On("IsHumanTurn").Return(true)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.NextRound()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "NextRound")
}

func TestEcarteInteractor_NextRound_BlocksOnGameEnd(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	gameMock := new(interfaces.MockEcarteGame)
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)
	epMock.On("Output", gameMock, nil).Return("ended")

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	got := ei.NextRound()
	assert.Equal(t, "ended", got)
}

func TestEcarteInteractor_GetConfig(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	gameMock := new(interfaces.MockEcarteGame)
	cfg := domain.DefaultEcarteConfig()
	gameMock.On("GetConfig").Return(cfg)

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	assert.Equal(t, cfg, ei.GetConfig())
}

func TestEcarteInteractor_Hint(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	gameMock := new(interfaces.MockEcarteGame)
	epMock.On("HintOutput", gameMock).Return("hint")

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	assert.Equal(t, "hint", ei.Hint())
}

func TestEcarteInteractor_ActionLog(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	gameMock := new(interfaces.MockEcarteGame)
	epMock.On("ActionLogOutput", gameMock).Return("log")

	ei := usecase.NewEcarteInteractor(gameMock, epMock)
	assert.Equal(t, "log", ei.ActionLog())
}

func TestEcarteInteractor_Snapshot_RoundtripsViaRealGame(t *testing.T) {
	epMock := new(presenter.MockEcartePresenter)
	game := domain.NewDefaultEcarte()
	game.Reset()

	ei := usecase.NewEcarteInteractor(game, epMock)
	data, err := ei.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	ei2, err := usecase.RestoreEcarteInteractor(data, epMock)
	assert.NoError(t, err)
	assert.Equal(t, game.GetPhase(), ei2.Game.GetPhase())
}
