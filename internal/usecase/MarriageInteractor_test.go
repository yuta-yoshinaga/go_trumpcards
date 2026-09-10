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

const marriageMockOutput = `{"phase":0}`

func setupMarriageMocks() (*presenter.MockMarriagePresenter, *interfaces.MockMarriageGame) {
	pMock := new(presenter.MockMarriagePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(marriageMockOutput)
	gameMock := new(interfaces.MockMarriageGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MarriagePhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)
	return pMock, gameMock
}

func TestNewMarriageInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockMarriagePresenter)
	t.Run("g must not be nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MarriageInteractor: g must not be nil", func() {
			usecase.NewMarriageInteractor(nil, pMock)
		})
	})
	t.Run("gp must not be nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMarriageGame)
		assert.PanicsWithValue(t, "MarriageInteractor: gp must not be nil", func() {
			usecase.NewMarriageInteractor(gameMock, nil)
		})
	})
}

func TestMarriageInteractor_Reset(t *testing.T) {
	pMock, gameMock := setupMarriageMocks()
	gameMock.On("Reset").Return()
	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	assert.Equal(t, marriageMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMarriageInteractor_ResetWithConfig_Valid(t *testing.T) {
	pMock, gameMock := setupMarriageMocks()
	cfg := domain.MarriageConfig{PlayerCount: 3, CpuDifficulty: domain.MarriageCpuDifficultyHard, TargetRounds: 5}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	assert.Equal(t, marriageMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
	gameMock.AssertCalled(t, "Reset")
}

func TestMarriageInteractor_ResetWithConfig_Invalid(t *testing.T) {
	pMock := new(presenter.MockMarriagePresenter)
	gameMock := new(interfaces.MockMarriageGame)
	pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).
		Return("validation error")

	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	bad := domain.MarriageConfig{PlayerCount: 1}
	assert.Equal(t, "validation error", ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestMarriageInteractor_DrawFromStock(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockMarriagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(marriageMockOutput)
		gameMock := new(interfaces.MockMarriageGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewMarriageInteractor(gameMock, pMock)
		assert.Equal(t, marriageMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})
	t.Run("error from domain", func(t *testing.T) {
		pMock, gameMock := setupMarriageMocks()
		gameMock.On("PlayerDrawFromStock").Return(errors.New("boom"))
		ci := usecase.NewMarriageInteractor(gameMock, pMock)
		assert.Equal(t, marriageMockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
	t.Run("success", func(t *testing.T) {
		pMock, gameMock := setupMarriageMocks()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		ci := usecase.NewMarriageInteractor(gameMock, pMock)
		assert.Equal(t, marriageMockOutput, ci.DrawFromStock())
	})
}

func TestMarriageInteractor_DrawFromDiscard(t *testing.T) {
	pMock, gameMock := setupMarriageMocks()
	gameMock.On("PlayerDrawFromDiscard").Return(nil)
	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	assert.Equal(t, marriageMockOutput, ci.DrawFromDiscard())
}

func TestMarriageInteractor_Discard(t *testing.T) {
	pMock, gameMock := setupMarriageMocks()
	gameMock.On("PlayerDiscard", 0).Return(nil)
	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	assert.Equal(t, marriageMockOutput, ci.Discard(0))
	gameMock.AssertCalled(t, "PlayerDiscard", 0)
}

func TestMarriageInteractor_Discard_Error(t *testing.T) {
	pMock, gameMock := setupMarriageMocks()
	gameMock.On("PlayerDiscard", 5).Return(errors.New("bad"))
	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	assert.Equal(t, marriageMockOutput, ci.Discard(5))
}

func TestMarriageInteractor_Declare(t *testing.T) {
	pMock, gameMock := setupMarriageMocks()
	gameMock.On("PlayerDeclare", 3).Return(nil)
	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	assert.Equal(t, marriageMockOutput, ci.Declare(3))
	gameMock.AssertCalled(t, "PlayerDeclare", 3)
}

func TestMarriageInteractor_NextRound(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockMarriagePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(marriageMockOutput)
		gameMock := new(interfaces.MockMarriageGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewMarriageInteractor(gameMock, pMock)
		assert.Equal(t, marriageMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
	t.Run("normal", func(t *testing.T) {
		pMock, gameMock := setupMarriageMocks()
		gameMock.On("NextRound").Return()
		ci := usecase.NewMarriageInteractor(gameMock, pMock)
		assert.Equal(t, marriageMockOutput, ci.NextRound())
	})
}

func TestMarriageInteractor_GetConfig(t *testing.T) {
	pMock, gameMock := setupMarriageMocks()
	cfg := domain.DefaultMarriageConfig()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestMarriageInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockMarriagePresenter)
	gameMock := new(interfaces.MockMarriageGame)
	pMock.On("ActionLogOutput", gameMock).Return("log")
	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestMarriageInteractor_CpuTurns(t *testing.T) {
	// IsHumanTurn=false を 1 度返して CpuPlay を呼ばせ、次で終了フラグを立てる。
	pMock := new(presenter.MockMarriagePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(marriageMockOutput)
	gameMock := new(interfaces.MockMarriageGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetPhase").Return(domain.MarriagePhaseDraw)
	gameMock.On("GetGameEndFlag").Return(false).Once()
	gameMock.On("GetGameEndFlag").Return(true)
	gameMock.On("IsHumanTurn").Return(false)
	gameMock.On("CpuPlay").Return()

	ci := usecase.NewMarriageInteractor(gameMock, pMock)
	_ = ci.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestRestoreMarriageInteractor(t *testing.T) {
	g := domain.NewDefaultMarriage()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	pMock := new(presenter.MockMarriagePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(marriageMockOutput)

	ci, err := usecase.RestoreMarriageInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreMarriageInteractor_BadJSON(t *testing.T) {
	pMock := new(presenter.MockMarriagePresenter)
	_, err := usecase.RestoreMarriageInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

func TestMarriageInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultMarriage()
	g.Reset()
	pMock := new(presenter.MockMarriagePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(marriageMockOutput)
	ci := usecase.NewMarriageInteractor(g, pMock)
	data, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
