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

const hachihachiMockOutput = `{"phase":0}`

func newHachiHachiMocks() (*interfaces.MockHachiHachiGame, *presenter.MockHachiHachiPresenter) {
	return new(interfaces.MockHachiHachiGame), new(presenter.MockHachiHachiPresenter)
}

func TestNewHachiHachiInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockHachiHachiPresenter)
	assert.PanicsWithValue(t, "HachiHachiInteractor: kg must not be nil", func() {
		usecase.NewHachiHachiInteractor(nil, cpMock)
	})
	gameMock := new(interfaces.MockHachiHachiGame)
	assert.PanicsWithValue(t, "HachiHachiInteractor: cp must not be nil", func() {
		usecase.NewHachiHachiInteractor(gameMock, nil)
	})
}

func TestHachiHachiInteractor_Play_Error(t *testing.T) {
	gm, cp := newHachiHachiMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(hachihachiMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("GetPhase").Return(domain.HachiHachiPhasePlay)
	gm.On("PlayerPlay", 0, -1).Return(assert.AnError)

	ki := usecase.NewHachiHachiInteractor(gm, cp)
	assert.Equal(t, hachihachiMockOutput, ki.Play(0, -1))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestHachiHachiInteractor_Play_NotPlayable(t *testing.T) {
	gm, cp := newHachiHachiMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(hachihachiMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)
	gm.On("GetPhase").Return(domain.HachiHachiPhasePlay)

	ki := usecase.NewHachiHachiInteractor(gm, cp)
	assert.Equal(t, hachihachiMockOutput, ki.Play(0, -1))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestHachiHachiInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, cp := newHachiHachiMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(hachihachiMockOutput)
	ki := usecase.NewHachiHachiInteractor(gm, cp)
	out := ki.ResetWithConfig(domain.HachiHachiConfig{CpuDifficulty: 99, TargetRounds: 3})
	assert.Equal(t, hachihachiMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestHachiHachiInteractor_HintAndLog(t *testing.T) {
	gm, cp := newHachiHachiMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	ki := usecase.NewHachiHachiInteractor(gm, cp)
	assert.Equal(t, "hint", ki.Hint())
	assert.Equal(t, "log", ki.ActionLog())
}

func TestHachiHachiInteractor_GetConfig(t *testing.T) {
	gm, cp := newHachiHachiMocks()
	cfg := domain.HachiHachiConfig{CpuDifficulty: domain.HachiHachiCpuDifficultyEasy, TargetRounds: 3}
	gm.On("GetConfig").Return(cfg)
	ki := usecase.NewHachiHachiInteractor(gm, cp)
	assert.Equal(t, cfg, ki.GetConfig())
}

// TestHachiHachiInteractor_RealFlow は本物のドメインで Reset→Play→NextRound を
// advance() 経由で駆動し、終局に到達することを検証する。
func TestHachiHachiInteractor_RealFlow(t *testing.T) {
	cp := new(presenter.MockHachiHachiPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(hachihachiMockOutput)

	g := domain.NewDefaultHachiHachi()
	cfg := domain.DefaultHachiHachiConfig()
	cfg.TargetRounds = 2
	g.SetConfig(cfg)
	ki := usecase.NewHachiHachiInteractor(g, cp)
	ki.Reset()

	for step := 0; step < 20000 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.HachiHachiPhasePlay:
			require.True(t, g.IsHumanTurn(), "advance() は人間手番でのみ停止する")
			ki.Play(0, -1)
		case domain.HachiHachiPhaseRoundEnd:
			ki.NextRound()
		default:
			// GameEnd
		}
	}
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.HachiHachiPhaseGameEnd, g.GetPhase())
}

func TestHachiHachiInteractor_SnapshotAndRestore(t *testing.T) {
	cp := new(presenter.MockHachiHachiPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(hachihachiMockOutput)

	real := usecase.NewHachiHachiInteractor(domain.NewDefaultHachiHachi(), cp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreHachiHachiInteractor(data, cp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreHachiHachiInteractor([]byte("not json"), cp)
	assert.Error(t, err)

	var g domain.HachiHachi
	require.NoError(t, json.Unmarshal(data, &g))
}
