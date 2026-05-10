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

const crMockOutput = `{"phase":0}`

// setupContractRummyMocks 共通のモック組み合わせ。runCpuTurns ループは IsHumanTurn=true で抜ける。
func setupContractRummyMocks() (*presenter.MockContractRummyPresenter, *interfaces.MockContractRummyGame) {
	pMock := new(presenter.MockContractRummyPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(crMockOutput)
	gameMock := new(interfaces.MockContractRummyGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ContractRummyPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)
	return pMock, gameMock
}

func TestNewContractRummyInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockContractRummyPresenter)
	t.Run("g must not be nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ContractRummyInteractor: g must not be nil", func() {
			usecase.NewContractRummyInteractor(nil, pMock)
		})
	})
	t.Run("gp must not be nil", func(t *testing.T) {
		gameMock := new(interfaces.MockContractRummyGame)
		assert.PanicsWithValue(t, "ContractRummyInteractor: gp must not be nil", func() {
			usecase.NewContractRummyInteractor(gameMock, nil)
		})
	})
}

func TestContractRummyInteractor_Reset(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	gameMock.On("Reset").Return()
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, crMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestContractRummyInteractor_ResetWithConfig_Valid(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	cfg := domain.ContractRummyConfig{CpuDifficulty: domain.ContractRummyCpuDifficultyHard, FailContractPenalty: 50}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, crMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
	gameMock.AssertCalled(t, "Reset")
}

func TestContractRummyInteractor_ResetWithConfig_Invalid(t *testing.T) {
	pMock := new(presenter.MockContractRummyPresenter)
	gameMock := new(interfaces.MockContractRummyGame)
	pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).
		Return("validation error")

	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	bad := domain.ContractRummyConfig{CpuDifficulty: domain.ContractRummyCpuDifficulty(-1)}
	assert.Equal(t, "validation error", ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestContractRummyInteractor_DrawFromStock(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockContractRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(crMockOutput)
		gameMock := new(interfaces.MockContractRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewContractRummyInteractor(gameMock, pMock)
		assert.Equal(t, crMockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})
	t.Run("error from domain", func(t *testing.T) {
		pMock, gameMock := setupContractRummyMocks()
		gameMock.On("PlayerDrawFromStock").Return(errors.New("boom"))
		ci := usecase.NewContractRummyInteractor(gameMock, pMock)
		assert.Equal(t, crMockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
	t.Run("success", func(t *testing.T) {
		pMock, gameMock := setupContractRummyMocks()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		ci := usecase.NewContractRummyInteractor(gameMock, pMock)
		assert.Equal(t, crMockOutput, ci.DrawFromStock())
	})
}

func TestContractRummyInteractor_DrawFromDiscard(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	gameMock.On("PlayerDrawFromDiscard").Return(nil)
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, crMockOutput, ci.DrawFromDiscard())
}

func TestContractRummyInteractor_MeldContract(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	indices := [][]int{{0, 1, 2}, {3, 4, 5}}
	gameMock.On("PlayerMeldContract", indices).Return(nil)
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, crMockOutput, ci.MeldContract(indices))
	gameMock.AssertCalled(t, "PlayerMeldContract", indices)
}

func TestContractRummyInteractor_MeldContract_Error(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	indices := [][]int{{0}}
	gameMock.On("PlayerMeldContract", indices).Return(errors.New("bad"))
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, crMockOutput, ci.MeldContract(indices))
}

func TestContractRummyInteractor_MeldExtra(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	idx := []int{0, 1, 2}
	gameMock.On("PlayerMeldExtra", idx).Return(nil)
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, crMockOutput, ci.MeldExtra(idx))
}

func TestContractRummyInteractor_Layoff(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	gameMock.On("PlayerLayoff", 1, 0, 2).Return(nil)
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, crMockOutput, ci.Layoff(1, 0, 2))
}

func TestContractRummyInteractor_Discard(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	gameMock.On("PlayerDiscard", 0).Return(nil)
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, crMockOutput, ci.Discard(0))
}

func TestContractRummyInteractor_NextRound(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockContractRummyPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(crMockOutput)
		gameMock := new(interfaces.MockContractRummyGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewContractRummyInteractor(gameMock, pMock)
		assert.Equal(t, crMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
	t.Run("normal", func(t *testing.T) {
		pMock, gameMock := setupContractRummyMocks()
		gameMock.On("NextRound").Return()
		ci := usecase.NewContractRummyInteractor(gameMock, pMock)
		assert.Equal(t, crMockOutput, ci.NextRound())
	})
}

func TestContractRummyInteractor_GetConfig(t *testing.T) {
	pMock, gameMock := setupContractRummyMocks()
	cfg := domain.DefaultContractRummyConfig()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestContractRummyInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockContractRummyPresenter)
	gameMock := new(interfaces.MockContractRummyGame)
	pMock.On("ActionLogOutput", gameMock).Return("log")
	ci := usecase.NewContractRummyInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreContractRummyInteractor(t *testing.T) {
	g := domain.NewDefaultContractRummy()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	pMock := new(presenter.MockContractRummyPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(crMockOutput)

	ci, err := usecase.RestoreContractRummyInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreContractRummyInteractor_BadJSON(t *testing.T) {
	pMock := new(presenter.MockContractRummyPresenter)
	_, err := usecase.RestoreContractRummyInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

func TestContractRummyInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultContractRummy()
	g.Reset()
	pMock := new(presenter.MockContractRummyPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(crMockOutput)
	ci := usecase.NewContractRummyInteractor(g, pMock)
	data, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
