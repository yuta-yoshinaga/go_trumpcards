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

const indianRummyMockOutput = `{"phase":0}`

func setupIndianRummyMocks() (*presenter.MockIndianRummyPresenter, *interfaces.MockIndianRummyGame) {
	pMock := new(presenter.MockIndianRummyPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(indianRummyMockOutput)
	gameMock := new(interfaces.MockIndianRummyGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.IndianRummyPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)
	return pMock, gameMock
}

func TestNewIndianRummyInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockIndianRummyPresenter)
	t.Run("g must not be nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "IndianRummyInteractor: g must not be nil", func() {
			usecase.NewIndianRummyInteractor(nil, pMock)
		})
	})
	t.Run("gp must not be nil", func(t *testing.T) {
		gameMock := new(interfaces.MockIndianRummyGame)
		assert.PanicsWithValue(t, "IndianRummyInteractor: gp must not be nil", func() {
			usecase.NewIndianRummyInteractor(gameMock, nil)
		})
	})
}

func TestIndianRummyInteractor_Reset(t *testing.T) {
	pMock, gameMock := setupIndianRummyMocks()
	gameMock.On("Reset").Return()
	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	assert.Equal(t, indianRummyMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestIndianRummyInteractor_ResetWithConfig_Valid(t *testing.T) {
	pMock, gameMock := setupIndianRummyMocks()
	cfg := domain.IndianRummyConfig{PlayerCount: 3, CpuDifficulty: domain.IndianRummyCpuDifficultyHard, TargetRounds: 5}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	assert.Equal(t, indianRummyMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
	gameMock.AssertCalled(t, "Reset")
}

func TestIndianRummyInteractor_ResetWithConfig_Invalid(t *testing.T) {
	pMock := new(presenter.MockIndianRummyPresenter)
	gameMock := new(interfaces.MockIndianRummyGame)
	pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).
		Return("validation error")

	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	bad := domain.IndianRummyConfig{PlayerCount: 1}
	assert.Equal(t, "validation error", ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestIndianRummyInteractor_DrawFromStock(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockIndianRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(indianRummyMockOutput)
		gameMock := new(interfaces.MockIndianRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
		assert.Equal(t, indianRummyMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})
	t.Run("error from domain", func(t *testing.T) {
		pMock, gameMock := setupIndianRummyMocks()
		gameMock.On("PlayerDrawFromStock").Return(errors.New("boom"))
		ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
		assert.Equal(t, indianRummyMockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
	t.Run("success", func(t *testing.T) {
		pMock, gameMock := setupIndianRummyMocks()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
		assert.Equal(t, indianRummyMockOutput, ci.DrawFromStock())
	})
}

func TestIndianRummyInteractor_DrawFromDiscard(t *testing.T) {
	pMock, gameMock := setupIndianRummyMocks()
	gameMock.On("PlayerDrawFromDiscard").Return(nil)
	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	assert.Equal(t, indianRummyMockOutput, ci.DrawFromDiscard())
}

func TestIndianRummyInteractor_Discard(t *testing.T) {
	pMock, gameMock := setupIndianRummyMocks()
	gameMock.On("PlayerDiscard", 0).Return(nil)
	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	assert.Equal(t, indianRummyMockOutput, ci.Discard(0))
	gameMock.AssertCalled(t, "PlayerDiscard", 0)
}

func TestIndianRummyInteractor_Discard_Error(t *testing.T) {
	pMock, gameMock := setupIndianRummyMocks()
	gameMock.On("PlayerDiscard", 5).Return(errors.New("bad"))
	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	assert.Equal(t, indianRummyMockOutput, ci.Discard(5))
}

func TestIndianRummyInteractor_Declare(t *testing.T) {
	pMock, gameMock := setupIndianRummyMocks()
	gameMock.On("PlayerDeclare", 3).Return(nil)
	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	assert.Equal(t, indianRummyMockOutput, ci.Declare(3))
	gameMock.AssertCalled(t, "PlayerDeclare", 3)
}

func TestIndianRummyInteractor_NextRound(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockIndianRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(indianRummyMockOutput)
		gameMock := new(interfaces.MockIndianRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
		assert.Equal(t, indianRummyMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
	t.Run("normal", func(t *testing.T) {
		pMock, gameMock := setupIndianRummyMocks()
		gameMock.On("NextRound").Return()
		ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
		assert.Equal(t, indianRummyMockOutput, ci.NextRound())
	})
}

func TestIndianRummyInteractor_GetConfig(t *testing.T) {
	pMock, gameMock := setupIndianRummyMocks()
	cfg := domain.DefaultIndianRummyConfig()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestIndianRummyInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockIndianRummyPresenter)
	gameMock := new(interfaces.MockIndianRummyGame)
	pMock.On("ActionLogOutput", gameMock).Return("log")
	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestIndianRummyInteractor_CpuTurns(t *testing.T) {
	// IsHumanTurn=false を 1 度返して CpuPlay を呼ばせ、次で終了フラグを立てる。
	pMock := new(presenter.MockIndianRummyPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(indianRummyMockOutput)
	gameMock := new(interfaces.MockIndianRummyGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetPhase").Return(domain.IndianRummyPhaseDraw)
	gameMock.On("GetGameEndFlag").Return(false).Once()
	gameMock.On("GetGameEndFlag").Return(true)
	gameMock.On("IsHumanTurn").Return(false)
	gameMock.On("CpuPlay").Return()

	ci := usecase.NewIndianRummyInteractor(gameMock, pMock)
	_ = ci.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestRestoreIndianRummyInteractor(t *testing.T) {
	g := domain.NewDefaultIndianRummy()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	pMock := new(presenter.MockIndianRummyPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(indianRummyMockOutput)

	ci, err := usecase.RestoreIndianRummyInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreIndianRummyInteractor_BadJSON(t *testing.T) {
	pMock := new(presenter.MockIndianRummyPresenter)
	_, err := usecase.RestoreIndianRummyInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

func TestIndianRummyInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultIndianRummy()
	g.Reset()
	pMock := new(presenter.MockIndianRummyPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(indianRummyMockOutput)
	ci := usecase.NewIndianRummyInteractor(g, pMock)
	data, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
