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

func TestNewTwoTenJackInteractor_NilGuards(t *testing.T) {
	tpMock := new(presenter.MockTwoTenJackPresenter)

	t.Run("panics when t is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "TwoTenJackInteractor: t must not be nil", func() {
			usecase.NewTwoTenJackInteractor(nil, tpMock)
		})
	})

	t.Run("panics when tp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockTwoTenJackGame)
		assert.PanicsWithValue(t, "TwoTenJackInteractor: tp must not be nil", func() {
			usecase.NewTwoTenJackInteractor(gameMock, nil)
		})
	})
}

func TestTwoTenJackInteractor_Reset_StaysInDeclare(t *testing.T) {
	mockOutput := `{"phase":0}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwoTenJackPhaseDeclare)
	gameMock.On("IsHumanDeclareTurn").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.Reset()
	assert.Equal(t, mockOutput, result)
}

func TestTwoTenJackInteractor_Reset_CpuDeclaresThenPlay(t *testing.T) {
	mockOutput := `{"phase":1}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	// Declare phase, CPU declares once, then moves to play phase
	gameMock.On("GetPhase").Return(domain.TwoTenJackPhaseDeclare).Once()
	gameMock.On("IsHumanDeclareTurn").Return(false).Once()
	gameMock.On("CpuDeclareTrump").Return()
	// after declare: phase=play
	gameMock.On("GetPhase").Return(domain.TwoTenJackPhasePlay)
	// runCpuTurns: human turn
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.Reset()
	assert.Equal(t, mockOutput, result)
}

func TestTwoTenJackInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	cfg := domain.TwoTenJackConfig{CpuDifficulty: domain.TwoTenJackCpuDifficultyHard, PointLimit: 80}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwoTenJackPhaseDeclare)
	gameMock.On("IsHumanDeclareTurn").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.ResetWithConfig(cfg)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestTwoTenJackInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	tpMock := new(presenter.MockTwoTenJackPresenter)
	gameMock := new(interfaces.MockTwoTenJackGame)
	tpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	cfg := domain.TwoTenJackConfig{CpuDifficulty: -1, PointLimit: 50}
	result := ti.ResetWithConfig(cfg)
	assert.Equal(t, "validation error", result)
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestTwoTenJackInteractor_DeclareTrump_GameEnded(t *testing.T) {
	mockOutput := `{"ge":true}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.DeclareTrump(domain.CardDesignSpade)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "PlayerDeclareTrump")
}

func TestTwoTenJackInteractor_DeclareTrump_Error(t *testing.T) {
	mockOutput := `{"err":true}`
	declErr := errors.New("bad")
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, declErr).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerDeclareTrump", domain.CardDesignSpade).Return(declErr)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.DeclareTrump(domain.CardDesignSpade)
	assert.Equal(t, mockOutput, result)
}

func TestTwoTenJackInteractor_DeclareTrump_Success(t *testing.T) {
	mockOutput := `{"ok":true}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("PlayerDeclareTrump", domain.CardDesignSpade).Return(nil)
	gameMock.On("GetPhase").Return(domain.TwoTenJackPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.DeclareTrump(domain.CardDesignSpade)
	assert.Equal(t, mockOutput, result)
}

func TestTwoTenJackInteractor_Play_Guard(t *testing.T) {
	mockOutput := `{}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.Play(0)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "PlayerPlay")
}

func TestTwoTenJackInteractor_Play_Success(t *testing.T) {
	mockOutput := `{"ok":1}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 0).Return(nil)
	gameMock.On("GetPhase").Return(domain.TwoTenJackPhasePlay)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.Play(0)
	assert.Equal(t, mockOutput, result)
}

func TestTwoTenJackInteractor_Play_Error(t *testing.T) {
	mockOutput := `{"err":1}`
	playErr := errors.New("bad")
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, playErr).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("PlayerPlay", 0).Return(playErr)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.Play(0)
	assert.Equal(t, mockOutput, result)
}

func TestTwoTenJackInteractor_NextTrick(t *testing.T) {
	mockOutput := `{"n":1}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("NextTrick").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.TwoTenJackPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.NextTrick()
	assert.Equal(t, mockOutput, result)
}

func TestTwoTenJackInteractor_NextRound(t *testing.T) {
	mockOutput := `{"nr":1}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("NextRound").Return()
	gameMock.On("GetPhase").Return(domain.TwoTenJackPhaseDeclare)
	gameMock.On("IsHumanDeclareTurn").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.NextRound()
	assert.Equal(t, mockOutput, result)
}

func TestTwoTenJackInteractor_NextRound_GameEnded(t *testing.T) {
	mockOutput := `{"end":1}`
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockTwoTenJackGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	result := ti.NextRound()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestTwoTenJackInteractor_GetConfig(t *testing.T) {
	tpMock := new(presenter.MockTwoTenJackPresenter)
	gameMock := new(interfaces.MockTwoTenJackGame)
	cfg := domain.DefaultTwoTenJackConfig()
	gameMock.On("GetConfig").Return(cfg)
	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	assert.Equal(t, cfg, ti.GetConfig())
}

func TestTwoTenJackInteractor_Hint(t *testing.T) {
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("HintOutput", mock.Anything).Return("hint-out")
	gameMock := new(interfaces.MockTwoTenJackGame)
	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	assert.Equal(t, "hint-out", ti.Hint())
}

func TestTwoTenJackInteractor_ActionLog(t *testing.T) {
	tpMock := new(presenter.MockTwoTenJackPresenter)
	tpMock.On("ActionLogOutput", mock.Anything).Return("log-out")
	gameMock := new(interfaces.MockTwoTenJackGame)
	ti := usecase.NewTwoTenJackInteractor(gameMock, tpMock)
	assert.Equal(t, "log-out", ti.ActionLog())
}

func TestTwoTenJackInteractor_Snapshot(t *testing.T) {
	tpMock := new(presenter.MockTwoTenJackPresenter)
	players := []*domain.TwoTenJackPlayer{
		domain.NewTwoTenJackPlayer(true),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
	}
	game := domain.NewTwoTenJack(domain.NewTrumpCards(0), players, domain.DefaultTwoTenJackConfig())
	ti := usecase.NewTwoTenJackInteractor(game, tpMock)
	data, err := ti.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreTwoTenJackInteractor(t *testing.T) {
	tpMock := new(presenter.MockTwoTenJackPresenter)
	players := []*domain.TwoTenJackPlayer{
		domain.NewTwoTenJackPlayer(true),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
	}
	game := domain.NewTwoTenJack(domain.NewTrumpCards(0), players, domain.DefaultTwoTenJackConfig())
	ti := usecase.NewTwoTenJackInteractor(game, tpMock)
	data, err := ti.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreTwoTenJackInteractor(data, tpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}
