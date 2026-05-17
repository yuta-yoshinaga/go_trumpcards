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

func TestNewBriscolaInteractor_NilGuards(t *testing.T) {
	bpMock := new(presenter.MockBriscolaPresenter)

	t.Run("panics when b is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BriscolaInteractor: b must not be nil", func() {
			usecase.NewBriscolaInteractor(nil, bpMock)
		})
	})

	t.Run("panics when bp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBriscolaGame)
		assert.PanicsWithValue(t, "BriscolaInteractor: bp must not be nil", func() {
			usecase.NewBriscolaInteractor(gameMock, nil)
		})
	})
}

func TestBriscolaInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBriscolaPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBriscolaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BriscolaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	got := bi.Reset()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "Reset")
}

func TestBriscolaInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBriscolaPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBriscolaGame)
	cfg := domain.DefaultBriscolaConfig()
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BriscolaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	got := bi.ResetWithConfig(cfg)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestBriscolaInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	bpMock := new(presenter.MockBriscolaPresenter)
	gameMock := new(interfaces.MockBriscolaGame)
	bpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	invalid := domain.BriscolaConfig{CpuDifficulty: 99}
	got := bi.ResetWithConfig(invalid)
	assert.Equal(t, "validation error", got)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestBriscolaInteractor_Play_Valid(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBriscolaPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBriscolaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.BriscolaPhasePlay)

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	got := bi.Play(1)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerPlay", 1)
}

func TestBriscolaInteractor_Play_TriggersResolveOnHumanFinalCard(t *testing.T) {
	mockOutput := `{"phase":1}`
	bpMock := new(presenter.MockBriscolaPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBriscolaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once() // for guardNotPlayable
	gameMock.On("PlayerPlay", 0).Return(nil)
	// Phase progresses to TrickEnd after PlayerPlay; ResolveTrick is called; then back to Play
	gameMock.On("GetPhase").Return(domain.BriscolaPhaseTrickEnd).Once()
	gameMock.On("ResolveTrick").Return()
	gameMock.On("GetPhase").Return(domain.BriscolaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	got := bi.Play(0)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestBriscolaInteractor_Play_Error(t *testing.T) {
	wantErr := errors.New("boom")
	bpMock := new(presenter.MockBriscolaPresenter)
	bpMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockBriscolaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 99).Return(wantErr)

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	got := bi.Play(99)
	assert.Equal(t, "error output", got)
}

func TestBriscolaInteractor_Play_GuardBlocksWhenGameEnded(t *testing.T) {
	bpMock := new(presenter.MockBriscolaPresenter)
	gameMock := new(interfaces.MockBriscolaGame)
	gameMock.On("GetGameEndFlag").Return(true)
	bpMock.On("Output", gameMock, nil).Return("game ended")

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	got := bi.Play(0)
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerPlay")
}

func TestBriscolaInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":0}`
	bpMock := new(presenter.MockBriscolaPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBriscolaGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BriscolaPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	got := bi.NextTrick()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "NextTrick")
}

func TestBriscolaInteractor_NextTrick_BlocksOnGameEnd(t *testing.T) {
	bpMock := new(presenter.MockBriscolaPresenter)
	gameMock := new(interfaces.MockBriscolaGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(true)
	bpMock.On("Output", gameMock, nil).Return("ended")

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	got := bi.NextTrick()
	assert.Equal(t, "ended", got)
}

func TestBriscolaInteractor_GetConfig(t *testing.T) {
	bpMock := new(presenter.MockBriscolaPresenter)
	gameMock := new(interfaces.MockBriscolaGame)
	cfg := domain.DefaultBriscolaConfig()
	gameMock.On("GetConfig").Return(cfg)

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	assert.Equal(t, cfg, bi.GetConfig())
}

func TestBriscolaInteractor_Hint(t *testing.T) {
	bpMock := new(presenter.MockBriscolaPresenter)
	gameMock := new(interfaces.MockBriscolaGame)
	bpMock.On("HintOutput", gameMock).Return("hint")

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	assert.Equal(t, "hint", bi.Hint())
}

func TestBriscolaInteractor_ActionLog(t *testing.T) {
	bpMock := new(presenter.MockBriscolaPresenter)
	gameMock := new(interfaces.MockBriscolaGame)
	bpMock.On("ActionLogOutput", gameMock).Return("log")

	bi := usecase.NewBriscolaInteractor(gameMock, bpMock)
	assert.Equal(t, "log", bi.ActionLog())
}

func TestBriscolaInteractor_Snapshot_RoundtripsViaRealGame(t *testing.T) {
	// Snapshot/Restore tests use the real domain object because the mock has no JSON.
	bpMock := new(presenter.MockBriscolaPresenter)
	game := domain.NewDefaultBriscola()
	game.Reset()

	bi := usecase.NewBriscolaInteractor(game, bpMock)
	data, err := bi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	bi2, err := usecase.RestoreBriscolaInteractor(data, bpMock)
	assert.NoError(t, err)
	assert.Equal(t, game.GetPhase(), bi2.Game.GetPhase())
}
