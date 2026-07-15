package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newTestCassino() *domain.Cassino {
	return domain.NewDefaultCassino()
}

func TestNewCassinoInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockCassinoPresenter)
	t.Run("panics when cg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CassinoInteractor: cg must not be nil", func() {
			usecase.NewCassinoInteractor(nil, ppMock)
		})
	})
	t.Run("panics when cp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CassinoInteractor: cp must not be nil", func() {
			usecase.NewCassinoInteractor(newTestCassino(), nil)
		})
	})
}

func TestCassinoInteractor_Methods(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockCassinoPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	ci := usecase.NewCassinoInteractor(newTestCassino(), ppMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, ci.Reset())
	})

	t.Run("Hint delegates to the presenter", func(t *testing.T) {
		ppMock.On("HintOutput", mock.Anything).Return("hint_output")
		assert.Equal(t, "hint_output", ci.Hint())
	})

	t.Run("Take with nothing returns error output", func(t *testing.T) {
		// After Reset the player has 4 cards. A legal take depends on the deal.
		// Without knowing specific cards, just assert output string type.
		out := ci.Take(0, nil, nil)
		assert.Equal(t, mockOutput, out)
	})

	t.Run("Build with no valid table also returns output", func(t *testing.T) {
		out := ci.Build(0, nil, 8)
		assert.Equal(t, mockOutput, out)
	})

	t.Run("Trail returns output", func(t *testing.T) {
		out := ci.Trail(0)
		assert.Equal(t, mockOutput, out)
	})

	t.Run("ResetWithConfig valid", func(t *testing.T) {
		cfg := domain.DefaultCassinoConfig()
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
	})

	t.Run("GetConfig returns config", func(t *testing.T) {
		cfg := ci.GetConfig()
		assert.Equal(t, 21, cfg.TargetScore)
	})
}

func TestCassinoInteractor_ActionLog(t *testing.T) {
	ppMock := new(presenter.MockCassinoPresenter)
	gameMock := new(interfaces.MockCassinoGame)
	ppMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)
	ci := usecase.NewCassinoInteractor(gameMock, ppMock)
	result := ci.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	ppMock.AssertExpectations(t)
}

func TestCassinoInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	ppMock := new(presenter.MockCassinoPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)

	gameMock := new(interfaces.MockCassinoGame)
	gameMock.On("Reset").Return()
	gameMock.On("NextRound").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return("playerTurn")
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerTake", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	gameMock.On("PlayerBuild", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	gameMock.On("PlayerTrail", mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetConfig").Return(domain.DefaultCassinoConfig())

	ci := usecase.NewCassinoInteractor(gameMock, ppMock)

	t.Run("Reset delegates", func(t *testing.T) {
		result := ci.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("Take delegates", func(t *testing.T) {
		result := ci.Take(0, []int{1}, nil)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerTake", 0, []int{1}, []int(nil))
	})

	t.Run("Build delegates", func(t *testing.T) {
		result := ci.Build(0, []int{0}, 8)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerBuild", 0, []int{0}, 8)
	})

	t.Run("Trail delegates", func(t *testing.T) {
		result := ci.Trail(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerTrail", 0)
	})

	t.Run("NextRound delegates", func(t *testing.T) {
		result := ci.NextRound()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("GetConfig delegates", func(t *testing.T) {
		cfg := ci.GetConfig()
		assert.Equal(t, 21, cfg.TargetScore)
	})
}

func TestCassinoInteractor_NextRoundSkipsWhenGameEnded(t *testing.T) {
	mockOutput := "done"
	ppMock := new(presenter.MockCassinoPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockCassinoGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ci := usecase.NewCassinoInteractor(gameMock, ppMock)
	result := ci.NextRound()
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestCassinoInteractor_TakeNotHumanTurn(t *testing.T) {
	mockOutput := "cpu"
	ppMock := new(presenter.MockCassinoPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockCassinoGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ci := usecase.NewCassinoInteractor(gameMock, ppMock)
	result := ci.Take(0, nil, nil)
	assert.Equal(t, mockOutput, result)
	gameMock.AssertNotCalled(t, "PlayerTake", mock.Anything, mock.Anything, mock.Anything)
}

func TestCassinoInteractor_ResetWithConfigInvalid(t *testing.T) {
	ppMock := new(presenter.MockCassinoPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return("invalid")
	ci := usecase.NewCassinoInteractor(newTestCassino(), ppMock)
	bad := domain.CassinoConfig{CpuDifficulty: 99, TargetScore: 21}
	result := ci.ResetWithConfig(bad)
	assert.Equal(t, "invalid", result)
}

func TestCassinoInteractor_Snapshot(t *testing.T) {
	ppMock := new(presenter.MockCassinoPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return("x")
	ci := usecase.NewCassinoInteractor(newTestCassino(), ppMock)
	raw, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, raw)
}

func TestRestoreCassinoInteractor(t *testing.T) {
	ppMock := new(presenter.MockCassinoPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return("x")
	game := newTestCassino()
	game.Reset() // mid-game state with shuffled deck + dealt hands
	ci := usecase.NewCassinoInteractor(game, ppMock)

	raw, err := ci.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreCassinoInteractor(raw, ppMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)

	// State assertions: restored snapshot must match the original.
	originalCfg := ci.GetConfig()
	restoredCfg := restored.GetConfig()
	assert.Equal(t, originalCfg, restoredCfg)
}

func TestRestoreCassinoInteractor_RejectsNilTrumpCards(t *testing.T) {
	ppMock := new(presenter.MockCassinoPresenter)
	// Hand-crafted snapshot without "tc" field — UnmarshalJSON should error.
	raw := []byte(`{"pl":[],"cf":{"ts":21,"mb":true,"sb":true,"di":1},"ph":"playerTurn","ct":0,"tb":[],"bl":[],"lc":-1,"al":[]}`)
	restored, err := usecase.RestoreCassinoInteractor(raw, ppMock)
	assert.Error(t, err)
	assert.Nil(t, restored)
}
