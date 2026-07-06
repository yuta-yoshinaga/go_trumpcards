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

const koikoiMockOutput = `{"phase":0}`

func newKoiKoiMocks() (*interfaces.MockKoiKoiGame, *presenter.MockKoiKoiPresenter) {
	return new(interfaces.MockKoiKoiGame), new(presenter.MockKoiKoiPresenter)
}

func TestNewKoiKoiInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockKoiKoiPresenter)
	assert.PanicsWithValue(t, "KoiKoiInteractor: kg must not be nil", func() {
		usecase.NewKoiKoiInteractor(nil, cpMock)
	})
	gameMock := new(interfaces.MockKoiKoiGame)
	assert.PanicsWithValue(t, "KoiKoiInteractor: cp must not be nil", func() {
		usecase.NewKoiKoiInteractor(gameMock, nil)
	})
}

func TestKoiKoiInteractor_Play_Error(t *testing.T) {
	gm, cp := newKoiKoiMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(koikoiMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerPlay", 0, -1).Return(assert.AnError)

	ki := usecase.NewKoiKoiInteractor(gm, cp)
	assert.Equal(t, koikoiMockOutput, ki.Play(0, -1))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestKoiKoiInteractor_Play_NotPlayable(t *testing.T) {
	gm, cp := newKoiKoiMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(koikoiMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)

	ki := usecase.NewKoiKoiInteractor(gm, cp)
	assert.Equal(t, koikoiMockOutput, ki.Play(0, -1))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestKoiKoiInteractor_Decide_Error(t *testing.T) {
	gm, cp := newKoiKoiMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(koikoiMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerDecide", false).Return(assert.AnError)

	ki := usecase.NewKoiKoiInteractor(gm, cp)
	assert.Equal(t, koikoiMockOutput, ki.Decide(false))
	gm.AssertNotCalled(t, "CpuDecide")
}

func TestKoiKoiInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, cp := newKoiKoiMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(koikoiMockOutput)
	ki := usecase.NewKoiKoiInteractor(gm, cp)
	out := ki.ResetWithConfig(domain.KoiKoiConfig{CpuDifficulty: 99, TargetScore: 15})
	assert.Equal(t, koikoiMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestKoiKoiInteractor_HintAndLog(t *testing.T) {
	gm, cp := newKoiKoiMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	ki := usecase.NewKoiKoiInteractor(gm, cp)
	assert.Equal(t, "hint", ki.Hint())
	assert.Equal(t, "log", ki.ActionLog())
}

func TestKoiKoiInteractor_GetConfig(t *testing.T) {
	gm, cp := newKoiKoiMocks()
	cfg := domain.KoiKoiConfig{CpuDifficulty: domain.KoiKoiCpuDifficultyEasy, TargetScore: 15}
	gm.On("GetConfig").Return(cfg)
	ki := usecase.NewKoiKoiInteractor(gm, cp)
	assert.Equal(t, cfg, ki.GetConfig())
}

// TestKoiKoiInteractor_RealFlow は本物のドメインで Reset→Play/Decide→NextRound を
// advance() 経由で駆動し、終局に到達することを検証する。
func TestKoiKoiInteractor_RealFlow(t *testing.T) {
	cp := new(presenter.MockKoiKoiPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(koikoiMockOutput)

	g := domain.NewDefaultKoiKoi()
	ki := usecase.NewKoiKoiInteractor(g, cp)
	ki.Reset()

	for step := 0; step < 20000 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.KoiKoiPhasePlay:
			require.True(t, g.IsHumanTurn(), "advance() は人間手番でのみ停止する")
			ki.Play(0, -1)
		case domain.KoiKoiPhaseKoiKoiDecision:
			require.True(t, g.IsHumanTurn())
			ki.Decide(false)
		case domain.KoiKoiPhaseRoundEnd:
			ki.NextRound()
		default:
			// GameEnd
		}
	}
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.KoiKoiPhaseGameEnd, g.GetPhase())
}

func TestKoiKoiInteractor_SnapshotAndRestore(t *testing.T) {
	cp := new(presenter.MockKoiKoiPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(koikoiMockOutput)

	real := usecase.NewKoiKoiInteractor(domain.NewDefaultKoiKoi(), cp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreKoiKoiInteractor(data, cp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreKoiKoiInteractor([]byte("not json"), cp)
	assert.Error(t, err)

	var g domain.KoiKoi
	require.NoError(t, json.Unmarshal(data, &g))
}
