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

const klMockOutput = `{"phase":0}`

// setupKalookiMocks 共通のモック組み合わせ。runCpuTurns ループは IsHumanTurn=true で抜ける。
func setupKalookiMocks() (*presenter.MockKalookiPresenter, *interfaces.MockKalookiGame) {
	pMock := new(presenter.MockKalookiPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(klMockOutput)
	gameMock := new(interfaces.MockKalookiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KalookiPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)
	return pMock, gameMock
}

func TestNewKalookiInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockKalookiPresenter)
	t.Run("g must not be nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KalookiInteractor: g must not be nil", func() {
			usecase.NewKalookiInteractor(nil, pMock)
		})
	})
	t.Run("gp must not be nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKalookiGame)
		assert.PanicsWithValue(t, "KalookiInteractor: gp must not be nil", func() {
			usecase.NewKalookiInteractor(gameMock, nil)
		})
	})
}

func TestKalookiInteractor_Reset(t *testing.T) {
	pMock, gameMock := setupKalookiMocks()
	gameMock.On("Reset").Return()
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, klMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestKalookiInteractor_ResetWithConfig_Valid(t *testing.T) {
	pMock, gameMock := setupKalookiMocks()
	cfg := domain.KalookiConfig{CpuDifficulty: domain.KalookiCpuDifficultyHard, PlayerCount: 3, OpeningThreshold: 60}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, klMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
	gameMock.AssertCalled(t, "Reset")
}

func TestKalookiInteractor_ResetWithConfig_Invalid(t *testing.T) {
	pMock := new(presenter.MockKalookiPresenter)
	gameMock := new(interfaces.MockKalookiGame)
	pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).
		Return("validation error")

	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	bad := domain.KalookiConfig{CpuDifficulty: domain.KalookiCpuDifficulty(-1), PlayerCount: 4, OpeningThreshold: 51}
	assert.Equal(t, "validation error", ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestKalookiInteractor_DrawFromStock(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockKalookiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(klMockOutput)
		gameMock := new(interfaces.MockKalookiGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewKalookiInteractor(gameMock, pMock)
		assert.Equal(t, klMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})
	t.Run("error from domain", func(t *testing.T) {
		pMock, gameMock := setupKalookiMocks()
		gameMock.On("PlayerDrawFromStock").Return(errors.New("boom"))
		ci := usecase.NewKalookiInteractor(gameMock, pMock)
		assert.Equal(t, klMockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
	t.Run("success", func(t *testing.T) {
		pMock, gameMock := setupKalookiMocks()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		ci := usecase.NewKalookiInteractor(gameMock, pMock)
		assert.Equal(t, klMockOutput, ci.DrawFromStock())
	})
}

func TestKalookiInteractor_DrawFromDiscard(t *testing.T) {
	pMock, gameMock := setupKalookiMocks()
	gameMock.On("PlayerDrawFromDiscard").Return(nil)
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, klMockOutput, ci.DrawFromDiscard())
}

func TestKalookiInteractor_Meld(t *testing.T) {
	pMock, gameMock := setupKalookiMocks()
	groups := [][]int{{0, 1, 2}, {3, 4, 5}}
	gameMock.On("PlayerMeld", groups).Return(nil)
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, klMockOutput, ci.Meld(groups))
	gameMock.AssertCalled(t, "PlayerMeld", groups)
}

func TestKalookiInteractor_Meld_Error(t *testing.T) {
	pMock, gameMock := setupKalookiMocks()
	groups := [][]int{{0}}
	gameMock.On("PlayerMeld", groups).Return(errors.New("bad"))
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, klMockOutput, ci.Meld(groups))
}

func TestKalookiInteractor_Layoff(t *testing.T) {
	pMock, gameMock := setupKalookiMocks()
	gameMock.On("PlayerLayoff", 1, 0, 2).Return(nil)
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, klMockOutput, ci.Layoff(1, 0, 2))
}

func TestKalookiInteractor_Discard(t *testing.T) {
	pMock, gameMock := setupKalookiMocks()
	gameMock.On("PlayerDiscard", 0).Return(nil)
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, klMockOutput, ci.Discard(0))
}

func TestKalookiInteractor_NextRound(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockKalookiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(klMockOutput)
		gameMock := new(interfaces.MockKalookiGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewKalookiInteractor(gameMock, pMock)
		assert.Equal(t, klMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
	t.Run("normal", func(t *testing.T) {
		pMock, gameMock := setupKalookiMocks()
		gameMock.On("NextRound").Return()
		ci := usecase.NewKalookiInteractor(gameMock, pMock)
		assert.Equal(t, klMockOutput, ci.NextRound())
	})
}

func TestKalookiInteractor_GetConfig(t *testing.T) {
	pMock, gameMock := setupKalookiMocks()
	cfg := domain.DefaultKalookiConfig()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestKalookiInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockKalookiPresenter)
	gameMock := new(interfaces.MockKalookiGame)
	pMock.On("ActionLogOutput", gameMock).Return("log")
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestKalookiInteractor_RunCpuTurns(t *testing.T) {
	pMock := new(presenter.MockKalookiPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(klMockOutput)
	gameMock := new(interfaces.MockKalookiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KalookiPhaseDraw)
	// First call: CPU turn, second: human turn → loop exits.
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	gameMock.On("Reset").Return()
	ci := usecase.NewKalookiInteractor(gameMock, pMock)
	ci.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestRestoreKalookiInteractor(t *testing.T) {
	g := domain.NewDefaultKalooki()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	pMock := new(presenter.MockKalookiPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(klMockOutput)

	ci, err := usecase.RestoreKalookiInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreKalookiInteractor_BadJSON(t *testing.T) {
	pMock := new(presenter.MockKalookiPresenter)
	_, err := usecase.RestoreKalookiInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

func TestKalookiInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultKalooki()
	g.Reset()
	pMock := new(presenter.MockKalookiPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(klMockOutput)
	ci := usecase.NewKalookiInteractor(g, pMock)
	data, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
