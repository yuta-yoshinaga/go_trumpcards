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

func TestNewSchnapsenInteractor_NilGuards(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)

	t.Run("panics when s is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "SchnapsenInteractor: s must not be nil", func() {
			usecase.NewSchnapsenInteractor(nil, spMock)
		})
	})

	t.Run("panics when sp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockSchnapsenGame)
		assert.PanicsWithValue(t, "SchnapsenInteractor: sp must not be nil", func() {
			usecase.NewSchnapsenInteractor(gameMock, nil)
		})
	})
}

func TestSchnapsenInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	spMock := new(presenter.MockSchnapsenPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchnapsenPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.Reset()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "Reset")
}

func TestSchnapsenInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`
	spMock := new(presenter.MockSchnapsenPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSchnapsenGame)
	cfg := domain.DefaultSchnapsenConfig()
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchnapsenPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.ResetWithConfig(cfg)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestSchnapsenInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)
	gameMock := new(interfaces.MockSchnapsenGame)
	spMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	invalid := domain.SchnapsenConfig{CpuDifficulty: 99}
	got := si.ResetWithConfig(invalid)
	assert.Equal(t, "validation error", got)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestSchnapsenInteractor_Play_Valid(t *testing.T) {
	mockOutput := `{"phase":0}`
	spMock := new(presenter.MockSchnapsenPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.SchnapsenPhasePlay)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.Play(1)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerPlay", 1)
}

func TestSchnapsenInteractor_Play_TriggersResolveOnHumanFinalCard(t *testing.T) {
	mockOutput := `{"phase":1}`
	spMock := new(presenter.MockSchnapsenPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once() // for guardNotPlayable
	gameMock.On("PlayerPlay", 0).Return(nil)
	gameMock.On("GetPhase").Return(domain.SchnapsenPhaseTrickEnd).Once()
	gameMock.On("ResolveTrick").Return()
	gameMock.On("GetPhase").Return(domain.SchnapsenPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.Play(0)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "ResolveTrick")
}

func TestSchnapsenInteractor_Play_Error(t *testing.T) {
	wantErr := errors.New("boom")
	spMock := new(presenter.MockSchnapsenPresenter)
	spMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 99).Return(wantErr)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.Play(99)
	assert.Equal(t, "error output", got)
}

func TestSchnapsenInteractor_Play_GuardBlocksWhenGameEnded(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("GetGameEndFlag").Return(true)
	spMock.On("Output", gameMock, nil).Return("game ended")

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.Play(0)
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerPlay")
}

func TestSchnapsenInteractor_DeclareMarriage_Valid(t *testing.T) {
	mockOutput := `{"phase":0}`
	spMock := new(presenter.MockSchnapsenPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDeclareMarriage", 2).Return(nil)
	gameMock.On("GetPhase").Return(domain.SchnapsenPhasePlay)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.DeclareMarriage(2)
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "PlayerDeclareMarriage", 2)
}

func TestSchnapsenInteractor_DeclareMarriage_Error(t *testing.T) {
	wantErr := errors.New("not a marriage")
	spMock := new(presenter.MockSchnapsenPresenter)
	spMock.On("Output", mock.Anything, wantErr).Return("error output")
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerDeclareMarriage", 0).Return(wantErr)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.DeclareMarriage(0)
	assert.Equal(t, "error output", got)
}

func TestSchnapsenInteractor_DeclareMarriage_GuardBlocksWhenGameEnded(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("GetGameEndFlag").Return(true)
	spMock.On("Output", gameMock, nil).Return("game ended")

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.DeclareMarriage(0)
	assert.Equal(t, "game ended", got)
	gameMock.AssertNotCalled(t, "PlayerDeclareMarriage")
}

func TestSchnapsenInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"phase":0}`
	spMock := new(presenter.MockSchnapsenPresenter)
	spMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.SchnapsenPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.NextTrick()
	assert.Equal(t, mockOutput, got)
	gameMock.AssertCalled(t, "NextTrick")
}

func TestSchnapsenInteractor_NextTrick_BlocksOnGameEnd(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)
	gameMock := new(interfaces.MockSchnapsenGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(true)
	spMock.On("Output", gameMock, nil).Return("ended")

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	got := si.NextTrick()
	assert.Equal(t, "ended", got)
}

func TestSchnapsenInteractor_GetConfig(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)
	gameMock := new(interfaces.MockSchnapsenGame)
	cfg := domain.DefaultSchnapsenConfig()
	gameMock.On("GetConfig").Return(cfg)

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	assert.Equal(t, cfg, si.GetConfig())
}

func TestSchnapsenInteractor_Hint(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)
	gameMock := new(interfaces.MockSchnapsenGame)
	spMock.On("HintOutput", gameMock).Return("hint")

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	assert.Equal(t, "hint", si.Hint())
}

func TestSchnapsenInteractor_ActionLog(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)
	gameMock := new(interfaces.MockSchnapsenGame)
	spMock.On("ActionLogOutput", gameMock).Return("log")

	si := usecase.NewSchnapsenInteractor(gameMock, spMock)
	assert.Equal(t, "log", si.ActionLog())
}

func TestSchnapsenInteractor_Snapshot_RoundtripsViaRealGame(t *testing.T) {
	spMock := new(presenter.MockSchnapsenPresenter)
	game := domain.NewDefaultSchnapsen()
	game.Reset()

	si := usecase.NewSchnapsenInteractor(game, spMock)
	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	si2, err := usecase.RestoreSchnapsenInteractor(data, spMock)
	assert.NoError(t, err)
	assert.Equal(t, game.GetPhase(), si2.Game.GetPhase())
}
