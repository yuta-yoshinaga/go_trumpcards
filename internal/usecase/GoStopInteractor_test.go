//go:build test

package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const gostopMockOutput = `{"phase":0}`

func newGoStopMocks() (*interfaces.MockGoStopGame, *presenter.MockGoStopPresenter) {
	return new(interfaces.MockGoStopGame), new(presenter.MockGoStopPresenter)
}

func TestNewGoStopInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockGoStopPresenter)
	assert.PanicsWithValue(t, "GoStopInteractor: kg must not be nil", func() {
		usecase.NewGoStopInteractor(nil, cpMock)
	})
	gameMock := new(interfaces.MockGoStopGame)
	assert.PanicsWithValue(t, "GoStopInteractor: cp must not be nil", func() {
		usecase.NewGoStopInteractor(gameMock, nil)
	})
}

func TestGoStopInteractor_Play_Error(t *testing.T) {
	gm, cp := newGoStopMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(gostopMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerPlay", 0, -1).Return(assert.AnError)

	ki := usecase.NewGoStopInteractor(gm, cp)
	assert.Equal(t, gostopMockOutput, ki.Play(0, -1))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestGoStopInteractor_Play_NotPlayable(t *testing.T) {
	gm, cp := newGoStopMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(gostopMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)

	ki := usecase.NewGoStopInteractor(gm, cp)
	assert.Equal(t, gostopMockOutput, ki.Play(0, -1))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestGoStopInteractor_Decide_Error(t *testing.T) {
	gm, cp := newGoStopMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(gostopMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerDecide", false).Return(assert.AnError)

	ki := usecase.NewGoStopInteractor(gm, cp)
	assert.Equal(t, gostopMockOutput, ki.Decide(false))
	gm.AssertNotCalled(t, "CpuDecide")
}

func TestGoStopInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, cp := newGoStopMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(gostopMockOutput)
	ki := usecase.NewGoStopInteractor(gm, cp)
	out := ki.ResetWithConfig(domain.GoStopConfig{CpuDifficulty: 99, TargetScore: 7})
	assert.Equal(t, gostopMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestGoStopInteractor_HintAndLog(t *testing.T) {
	gm, cp := newGoStopMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	ki := usecase.NewGoStopInteractor(gm, cp)
	assert.Equal(t, "hint", ki.Hint())
	assert.Equal(t, "log", ki.ActionLog())
}

func TestGoStopInteractor_GetConfig(t *testing.T) {
	gm, cp := newGoStopMocks()
	cfg := domain.GoStopConfig{CpuDifficulty: domain.GoStopCpuDifficultyEasy, TargetScore: 7}
	gm.On("GetConfig").Return(cfg)
	ki := usecase.NewGoStopInteractor(gm, cp)
	assert.Equal(t, cfg, ki.GetConfig())
}

func TestGoStopInteractor_RealFlow(t *testing.T) {
	cp := new(presenter.MockGoStopPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(gostopMockOutput)

	g := domain.NewDefaultGoStop()
	ki := usecase.NewGoStopInteractor(g, cp)
	ki.Reset()

	for step := 0; step < 20000 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.GoStopPhasePlay:
			require.True(t, g.IsHumanTurn(), "advance() は人間手番でのみ停止する")
			ki.Play(0, -1)
		case domain.GoStopPhaseGoDecision:
			require.True(t, g.IsHumanTurn())
			ki.Decide(false)
		case domain.GoStopPhaseRoundEnd:
			ki.NextRound()
		default:
		}
	}
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.GoStopPhaseGameEnd, g.GetPhase())
}

func TestGoStopInteractor_SnapshotAndRestore(t *testing.T) {
	cp := new(presenter.MockGoStopPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(gostopMockOutput)

	real := usecase.NewGoStopInteractor(domain.NewDefaultGoStop(), cp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreGoStopInteractor(data, cp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreGoStopInteractor([]byte("not json"), cp)
	assert.Error(t, err)

	var g domain.GoStop
	require.NoError(t, json.Unmarshal(data, &g))
}
