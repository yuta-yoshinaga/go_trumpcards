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

const machiavelliMockOutput = `{"phase":0}`

// setupMachiavelliMocks 共通のモック組み合わせ。runCpuTurns ループは IsHumanTurn=true で抜ける。
func setupMachiavelliMocks() (*presenter.MockMachiavelliPresenter, *interfaces.MockMachiavelliGame) {
	pMock := new(presenter.MockMachiavelliPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(machiavelliMockOutput)
	gameMock := new(interfaces.MockMachiavelliGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MachiavelliPhaseTurn)
	gameMock.On("IsHumanTurn").Return(true)
	return pMock, gameMock
}

func TestNewMachiavelliInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockMachiavelliPresenter)
	t.Run("g must not be nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MachiavelliInteractor: g must not be nil", func() {
			usecase.NewMachiavelliInteractor(nil, pMock)
		})
	})
	t.Run("gp must not be nil", func(t *testing.T) {
		gameMock := new(interfaces.MockMachiavelliGame)
		assert.PanicsWithValue(t, "MachiavelliInteractor: gp must not be nil", func() {
			usecase.NewMachiavelliInteractor(gameMock, nil)
		})
	})
}

func TestMachiavelliInteractor_Reset(t *testing.T) {
	pMock, gameMock := setupMachiavelliMocks()
	gameMock.On("Reset").Return()
	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	assert.Equal(t, machiavelliMockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestMachiavelliInteractor_ResetWithConfig_Valid(t *testing.T) {
	pMock, gameMock := setupMachiavelliMocks()
	cfg := domain.MachiavelliConfig{PlayerCount: 5, CpuDifficulty: domain.MachiavelliCpuDifficultyHard, TargetRounds: 4}
	gameMock.On("SetConfig", cfg).Return()
	gameMock.On("Reset").Return()

	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	assert.Equal(t, machiavelliMockOutput, ci.ResetWithConfig(cfg))
	gameMock.AssertCalled(t, "SetConfig", cfg)
	gameMock.AssertCalled(t, "Reset")
}

func TestMachiavelliInteractor_ResetWithConfig_Invalid(t *testing.T) {
	pMock := new(presenter.MockMachiavelliPresenter)
	gameMock := new(interfaces.MockMachiavelliGame)
	pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).
		Return("validation error")

	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	bad := domain.MachiavelliConfig{PlayerCount: 1}
	assert.Equal(t, "validation error", ci.ResetWithConfig(bad))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	gameMock.AssertNotCalled(t, "Reset")
}

func TestMachiavelliInteractor_Draw(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockMachiavelliPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(machiavelliMockOutput)
		gameMock := new(interfaces.MockMachiavelliGame)
		gameMock.On("GetGameEndFlag").Return(true)
		gameMock.On("IsHumanTurn").Return(false)
		ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
		assert.Equal(t, machiavelliMockOutput, ci.Draw())
		gameMock.AssertNotCalled(t, "PlayerDraw")
	})
	t.Run("error from domain", func(t *testing.T) {
		pMock, gameMock := setupMachiavelliMocks()
		gameMock.On("PlayerDraw").Return(errors.New("boom"))
		ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
		assert.Equal(t, machiavelliMockOutput, ci.Draw())
		gameMock.AssertCalled(t, "PlayerDraw")
	})
	t.Run("success", func(t *testing.T) {
		pMock, gameMock := setupMachiavelliMocks()
		gameMock.On("PlayerDraw").Return(nil)
		ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
		assert.Equal(t, machiavelliMockOutput, ci.Draw())
	})
}

func TestMachiavelliInteractor_Play(t *testing.T) {
	pMock, gameMock := setupMachiavelliMocks()
	refs := [][]domain.MachiavelliCardRef{{{Design: 1, Value: 3}, {Design: 1, Value: 4}, {Design: 1, Value: 5}}}
	idx := []int{0, 1, 2}
	gameMock.On("PlayerPlay", refs, idx).Return(nil)
	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	assert.Equal(t, machiavelliMockOutput, ci.Play(refs, idx))
	gameMock.AssertCalled(t, "PlayerPlay", refs, idx)
}

func TestMachiavelliInteractor_Play_Error(t *testing.T) {
	pMock, gameMock := setupMachiavelliMocks()
	refs := [][]domain.MachiavelliCardRef{{{Design: 1, Value: 3}}}
	gameMock.On("PlayerPlay", refs, []int{0}).Return(errors.New("bad"))
	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	assert.Equal(t, machiavelliMockOutput, ci.Play(refs, []int{0}))
}

func TestMachiavelliInteractor_NewMeld(t *testing.T) {
	pMock, gameMock := setupMachiavelliMocks()
	idx := []int{0, 1, 2}
	gameMock.On("PlayerNewMeld", idx).Return(nil)
	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	assert.Equal(t, machiavelliMockOutput, ci.NewMeld(idx))
}

func TestMachiavelliInteractor_Layoff(t *testing.T) {
	pMock, gameMock := setupMachiavelliMocks()
	gameMock.On("PlayerLayoff", 0, 2).Return(nil)
	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	assert.Equal(t, machiavelliMockOutput, ci.Layoff(0, 2))
}

func TestMachiavelliInteractor_NextRound(t *testing.T) {
	t.Run("game ended → no-op", func(t *testing.T) {
		pMock := new(presenter.MockMachiavelliPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(machiavelliMockOutput)
		gameMock := new(interfaces.MockMachiavelliGame)
		gameMock.On("GetGameEndFlag").Return(true)
		ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
		assert.Equal(t, machiavelliMockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})
	t.Run("normal", func(t *testing.T) {
		pMock, gameMock := setupMachiavelliMocks()
		gameMock.On("NextRound").Return()
		ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
		assert.Equal(t, machiavelliMockOutput, ci.NextRound())
	})
}

func TestMachiavelliInteractor_GetConfig(t *testing.T) {
	pMock, gameMock := setupMachiavelliMocks()
	cfg := domain.DefaultMachiavelliConfig()
	gameMock.On("GetConfig").Return(cfg)
	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ci.GetConfig())
}

func TestMachiavelliInteractor_ActionLog(t *testing.T) {
	pMock := new(presenter.MockMachiavelliPresenter)
	gameMock := new(interfaces.MockMachiavelliGame)
	pMock.On("ActionLogOutput", gameMock).Return("log")
	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	assert.Equal(t, "log", ci.ActionLog())
}

func TestMachiavelliInteractor_CpuLoop(t *testing.T) {
	// CPU が数手番プレイしてから人間の手番で止まることを確認する。
	pMock := new(presenter.MockMachiavelliPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(machiavelliMockOutput)
	gameMock := new(interfaces.MockMachiavelliGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.MachiavelliPhaseTurn)
	// 1 回目は CPU、2 回目以降は人間。
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	gameMock.On("Reset").Return()
	ci := usecase.NewMachiavelliInteractor(gameMock, pMock)
	ci.Reset()
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestRestoreMachiavelliInteractor(t *testing.T) {
	g := domain.NewDefaultMachiavelli()
	g.Reset()
	data, err := g.MarshalJSON()
	assert.NoError(t, err)

	pMock := new(presenter.MockMachiavelliPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(machiavelliMockOutput)

	ci, err := usecase.RestoreMachiavelliInteractor(data, pMock)
	assert.NoError(t, err)
	assert.NotNil(t, ci)
}

func TestRestoreMachiavelliInteractor_BadJSON(t *testing.T) {
	pMock := new(presenter.MockMachiavelliPresenter)
	_, err := usecase.RestoreMachiavelliInteractor([]byte("not json"), pMock)
	assert.Error(t, err)
}

func TestMachiavelliInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultMachiavelli()
	g.Reset()
	pMock := new(presenter.MockMachiavelliPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(machiavelliMockOutput)
	ci := usecase.NewMachiavelliInteractor(g, pMock)
	data, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
