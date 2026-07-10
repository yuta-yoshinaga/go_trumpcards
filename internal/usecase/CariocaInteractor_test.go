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

const cariocaMockOutput = `{"phase":0}`

// setupCariocaMocks 共通のモック組み合わせ。runCpuTurns ループは IsHumanTurn=true で抜ける。
func setupCariocaMocks() (*presenter.MockCariocaPresenter, *interfaces.MockCariocaGame) {
	pMock := new(presenter.MockCariocaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(cariocaMockOutput)
	gameMock := new(interfaces.MockCariocaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.CariocaPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)
	return pMock, gameMock
}

func TestNewCariocaInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockCariocaPresenter)
	t.Run("g must not be nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CariocaInteractor: g must not be nil", func() {
			usecase.NewCariocaInteractor(nil, pMock)
		})
	})
	t.Run("gp must not be nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCariocaGame)
		assert.PanicsWithValue(t, "CariocaInteractor: gp must not be nil", func() {
			usecase.NewCariocaInteractor(gameMock, nil)
		})
	})
}

func TestCariocaInteractor_Reset(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	gameMock.On("Reset").Return()
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cariocaMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestCariocaInteractor_ResetWithConfig_Valid(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	cfg := domain.CariocaConfig{PlayerCount: 5, CpuDifficulty: domain.CariocaCpuDifficultyHard, FailContractPenalty: 50}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cariocaMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
	gameMock.AssertCalled(t, "Reset")
}

func TestCariocaInteractor_ResetWithConfig_Invalid(t *testing.T) {
	pMock := new(presenter.MockCariocaPresenter)
	gameMock := new(interfaces.MockCariocaGame)
	pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).
		Return("validation error")

	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	bad := domain.CariocaConfig{PlayerCount: 2}
	assert.Equal(t, "validation error", ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestCariocaInteractor_DrawFromStock(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockCariocaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(cariocaMockOutput)
		gameMock := new(interfaces.MockCariocaGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewCariocaInteractor(gameMock, pMock)
		assert.Equal(t, cariocaMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})
	t.Run("error from domain", func(t *testing.T) {
		pMock, gameMock := setupCariocaMocks()
		gameMock.On("PlayerDrawFromStock").Return(errors.New("boom"))
		ci := usecase.NewCariocaInteractor(gameMock, pMock)
		assert.Equal(t, cariocaMockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
	t.Run("success", func(t *testing.T) {
		pMock, gameMock := setupCariocaMocks()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		ci := usecase.NewCariocaInteractor(gameMock, pMock)
		assert.Equal(t, cariocaMockOutput, ci.DrawFromStock())
	})
}

func TestCariocaInteractor_DrawFromDiscard(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	gameMock.On("PlayerDrawFromDiscard").Return(nil)
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cariocaMockOutput, ci.DrawFromDiscard())
}

func TestCariocaInteractor_MeldContract(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	indices := [][]int{{0, 1, 2}, {3, 4, 5}}
	gameMock.On("PlayerMeldContract", indices).Return(nil)
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cariocaMockOutput, ci.MeldContract(indices))
	gameMock.AssertCalled(t, "PlayerMeldContract", indices)
}

func TestCariocaInteractor_MeldContract_Error(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	indices := [][]int{{0}}
	gameMock.On("PlayerMeldContract", indices).Return(errors.New("bad"))
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cariocaMockOutput, ci.MeldContract(indices))
}

func TestCariocaInteractor_MeldExtra(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	idx := []int{0, 1, 2}
	gameMock.On("PlayerMeldExtra", idx).Return(nil)
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cariocaMockOutput, ci.MeldExtra(idx))
}

func TestCariocaInteractor_Layoff(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	gameMock.On("PlayerLayoff", 1, 0, 2).Return(nil)
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cariocaMockOutput, ci.Layoff(1, 0, 2))
}

func TestCariocaInteractor_Discard(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	gameMock.On("PlayerDiscard", 0).Return(nil)
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cariocaMockOutput, ci.Discard(0))
}

func TestCariocaInteractor_NextRound(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockCariocaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(cariocaMockOutput)
		gameMock := new(interfaces.MockCariocaGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewCariocaInteractor(gameMock, pMock)
		assert.Equal(t, cariocaMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
	t.Run("normal", func(t *testing.T) {
		pMock, gameMock := setupCariocaMocks()
		gameMock.On("NextRound").Return()
		ci := usecase.NewCariocaInteractor(gameMock, pMock)
		assert.Equal(t, cariocaMockOutput, ci.NextRound())
	})
}

func TestCariocaInteractor_GetConfig(t *testing.T) {
	pMock, gameMock := setupCariocaMocks()
	cfg := domain.DefaultCariocaConfig()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestCariocaInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockCariocaPresenter)
	gameMock := new(interfaces.MockCariocaGame)
	pMock.On("ActionLogOutput", gameMock).Return("log")
	ci := usecase.NewCariocaInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreCariocaInteractor(t *testing.T) {
	g := domain.NewDefaultCarioca()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	pMock := new(presenter.MockCariocaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(cariocaMockOutput)

	ci, err := usecase.RestoreCariocaInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreCariocaInteractor_BadJSON(t *testing.T) {
	pMock := new(presenter.MockCariocaPresenter)
	_, err := usecase.RestoreCariocaInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

func TestCariocaInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultCarioca()
	g.Reset()
	pMock := new(presenter.MockCariocaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(cariocaMockOutput)
	ci := usecase.NewCariocaInteractor(g, pMock)
	data, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
