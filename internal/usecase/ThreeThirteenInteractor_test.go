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

const ttMockOutput = `{"phase":0}`

func setupThreeThirteenMocks() (*presenter.MockThreeThirteenPresenter, *interfaces.MockThreeThirteenGame) {
	pMock := new(presenter.MockThreeThirteenPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(ttMockOutput)
	gameMock := new(interfaces.MockThreeThirteenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ThreeThirteenPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)
	return pMock, gameMock
}

func TestNewThreeThirteenInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockThreeThirteenPresenter)
	t.Run("g must not be nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ThreeThirteenInteractor: g must not be nil", func() {
			usecase.NewThreeThirteenInteractor(nil, pMock)
		})
	})
	t.Run("gp must not be nil", func(t *testing.T) {
		gameMock := new(interfaces.MockThreeThirteenGame)
		assert.PanicsWithValue(t, "ThreeThirteenInteractor: gp must not be nil", func() {
			usecase.NewThreeThirteenInteractor(gameMock, nil)
		})
	})
}

func TestThreeThirteenInteractor_Reset(t *testing.T) {
	pMock, gameMock := setupThreeThirteenMocks()
	gameMock.On("Reset").Return()
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, ttMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestThreeThirteenInteractor_ResetWithConfig_Valid(t *testing.T) {
	pMock, gameMock := setupThreeThirteenMocks()
	cfg := domain.ThreeThirteenConfig{CpuDifficulty: domain.ThreeThirteenCpuDifficultyHard, PlayerCount: 3}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, ttMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
}

func TestThreeThirteenInteractor_ResetWithConfig_Invalid(t *testing.T) {
	pMock := new(presenter.MockThreeThirteenPresenter)
	gameMock := new(interfaces.MockThreeThirteenGame)
	pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).
		Return("validation error")
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	bad := domain.ThreeThirteenConfig{CpuDifficulty: domain.ThreeThirteenCpuDifficulty(-1), PlayerCount: 4}
	assert.Equal(t, "validation error", ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestThreeThirteenInteractor_DrawFromStock(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockThreeThirteenPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(ttMockOutput)
		gameMock := new(interfaces.MockThreeThirteenGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
		assert.Equal(t, ttMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})
	t.Run("error from domain", func(t *testing.T) {
		pMock, gameMock := setupThreeThirteenMocks()
		gameMock.On("PlayerDrawFromStock").Return(errors.New("boom"))
		ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
		assert.Equal(t, ttMockOutput, ci.DrawFromStock())
	})
	t.Run("success", func(t *testing.T) {
		pMock, gameMock := setupThreeThirteenMocks()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
		assert.Equal(t, ttMockOutput, ci.DrawFromStock())
	})
}

func TestThreeThirteenInteractor_DrawFromDiscard(t *testing.T) {
	pMock, gameMock := setupThreeThirteenMocks()
	gameMock.On("PlayerDrawFromDiscard").Return(nil)
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, ttMockOutput, ci.DrawFromDiscard())
}

func TestThreeThirteenInteractor_Discard(t *testing.T) {
	pMock, gameMock := setupThreeThirteenMocks()
	gameMock.On("PlayerDiscard", 0).Return(nil)
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, ttMockOutput, ci.Discard(0))
	gameMock.AssertCalled(t, "PlayerDiscard", 0)
}

func TestThreeThirteenInteractor_Discard_Error(t *testing.T) {
	pMock, gameMock := setupThreeThirteenMocks()
	gameMock.On("PlayerDiscard", 1).Return(errors.New("bad"))
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, ttMockOutput, ci.Discard(1))
}

func TestThreeThirteenInteractor_Knock(t *testing.T) {
	pMock, gameMock := setupThreeThirteenMocks()
	gameMock.On("PlayerKnock", 2).Return(nil)
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, ttMockOutput, ci.Knock(2))
	gameMock.AssertCalled(t, "PlayerKnock", 2)
}

func TestThreeThirteenInteractor_Knock_Error(t *testing.T) {
	pMock, gameMock := setupThreeThirteenMocks()
	gameMock.On("PlayerKnock", 3).Return(errors.New("nope"))
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, ttMockOutput, ci.Knock(3))
}

func TestThreeThirteenInteractor_NextRound(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockThreeThirteenPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(ttMockOutput)
		gameMock := new(interfaces.MockThreeThirteenGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
		assert.Equal(t, ttMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
	t.Run("normal", func(t *testing.T) {
		pMock, gameMock := setupThreeThirteenMocks()
		gameMock.On("NextRound").Return()
		ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
		assert.Equal(t, ttMockOutput, ci.NextRound())
	})
}

func TestThreeThirteenInteractor_GetConfig(t *testing.T) {
	pMock, gameMock := setupThreeThirteenMocks()
	cfg := domain.DefaultThreeThirteenConfig()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestThreeThirteenInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockThreeThirteenPresenter)
	gameMock := new(interfaces.MockThreeThirteenGame)
	pMock.On("ActionLogOutput", gameMock).Return("log")
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestThreeThirteenInteractor_RunCpuTurns(t *testing.T) {
	pMock := new(presenter.MockThreeThirteenPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(ttMockOutput)
	gameMock := new(interfaces.MockThreeThirteenGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ThreeThirteenPhaseDraw)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	gameMock.On("Reset").Return()
	ci := usecase.NewThreeThirteenInteractor(gameMock, pMock)
	ci.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestRestoreThreeThirteenInteractor(t *testing.T) {
	g := domain.NewDefaultThreeThirteen()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)
	pMock := new(presenter.MockThreeThirteenPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(ttMockOutput)
	ci, err := usecase.RestoreThreeThirteenInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreThreeThirteenInteractor_BadJSON(t *testing.T) {
	pMock := new(presenter.MockThreeThirteenPresenter)
	_, err := usecase.RestoreThreeThirteenInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

func TestThreeThirteenInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultThreeThirteen()
	g.Reset()
	pMock := new(presenter.MockThreeThirteenPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(ttMockOutput)
	ci := usecase.NewThreeThirteenInteractor(g, pMock)
	data, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
