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

const basraMockOutput = `{"phase":1}`

func newBasraMocks() (*interfaces.MockBasraGame, *presenter.MockBasraPresenter) {
	return new(interfaces.MockBasraGame), new(presenter.MockBasraPresenter)
}

func TestNewBasraInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockBasraPresenter)
	assert.PanicsWithValue(t, "BasraInteractor: bg must not be nil", func() {
		usecase.NewBasraInteractor(nil, cpMock)
	})
	gameMock := new(interfaces.MockBasraGame)
	assert.PanicsWithValue(t, "BasraInteractor: cp must not be nil", func() {
		usecase.NewBasraInteractor(gameMock, nil)
	})
}

func TestBasraInteractor_Play_Error(t *testing.T) {
	gm, cp := newBasraMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(basraMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerPlay", 0, mock.Anything).Return(assert.AnError)

	bi := usecase.NewBasraInteractor(gm, cp)
	assert.Equal(t, basraMockOutput, bi.Play(0, nil))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestBasraInteractor_Play_NotPlayable(t *testing.T) {
	gm, cp := newBasraMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(basraMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)

	bi := usecase.NewBasraInteractor(gm, cp)
	assert.Equal(t, basraMockOutput, bi.Play(0, nil))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestBasraInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, cp := newBasraMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(basraMockOutput)
	bi := usecase.NewBasraInteractor(gm, cp)
	out := bi.ResetWithConfig(domain.BasraConfig{CpuDifficulty: 99})
	assert.Equal(t, basraMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestBasraInteractor_HintAndLog(t *testing.T) {
	gm, cp := newBasraMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	bi := usecase.NewBasraInteractor(gm, cp)
	assert.Equal(t, "hint", bi.Hint())
	assert.Equal(t, "log", bi.ActionLog())
}

func TestBasraInteractor_GetConfig(t *testing.T) {
	gm, cp := newBasraMocks()
	cfg := domain.BasraConfig{CpuDifficulty: domain.BasraCpuDifficultyEasy}
	gm.On("GetConfig").Return(cfg)
	bi := usecase.NewBasraInteractor(gm, cp)
	assert.Equal(t, cfg, bi.GetConfig())
}

// TestBasraInteractor_RealFlow は本物のドメインで Reset→Play→…→NextRound の流れを
// advance() 経由で駆動し、ゲーム終了に到達することを検証する。
func TestBasraInteractor_RealFlow(t *testing.T) {
	cp := new(presenter.MockBasraPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(basraMockOutput)

	g := domain.NewDefaultBasra()
	cfg := domain.DefaultBasraConfig()
	cfg.CpuDifficulty = domain.BasraCpuDifficultyNormal
	g.SetConfig(cfg)
	bi := usecase.NewBasraInteractor(g, cp)

	bi.Reset()

	for step := 0; step < 5000 && !g.GetGameEndFlag(); step++ {
		if g.GetPhase() != domain.BasraPhasePlay {
			break
		}
		require.True(t, g.IsHumanTurn(), "advance() は人間手番でのみ停止する")
		opts := g.GetCaptureOptions(g.GetCurrentTurn())
		bi.Play(0, opts[0])
	}
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.BasraPhaseGameEnd, g.GetPhase())

	// NextRound で新規ゲームが始まる。
	bi.NextRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, domain.BasraPhasePlay, g.GetPhase())
}

// TestBasraInteractor_GameEndScoresImmediately は、ゲーム終了に到達した時点で
// (NextRound を呼ぶ前に) すでに得点内訳 (lastDealDetail) が埋まっていることを検証する。
// これがないとフロントエンドは結果画面を描画できない (MUST 修正)。
func TestBasraInteractor_GameEndScoresImmediately(t *testing.T) {
	cp := new(presenter.MockBasraPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(basraMockOutput)

	g := domain.NewDefaultBasra()
	bi := usecase.NewBasraInteractor(g, cp)
	bi.Reset()

	for step := 0; step < 5000 && !g.GetGameEndFlag(); step++ {
		if g.GetPhase() != domain.BasraPhasePlay {
			break
		}
		opts := g.GetCaptureOptions(g.GetCurrentTurn())
		bi.Play(0, opts[0])
	}
	require.True(t, g.GetGameEndFlag())
	det := g.GetLastDealDetail()
	require.NotNil(t, det, "ゲーム終了時点で得点内訳が埋まっているべき")
	require.NotNil(t, det.Gained)
	assert.NotEmpty(t, g.GetWinners())
}

func TestBasraInteractor_SnapshotAndRestore(t *testing.T) {
	cp := new(presenter.MockBasraPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(basraMockOutput)

	real := usecase.NewBasraInteractor(domain.NewDefaultBasra(), cp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreBasraInteractor(data, cp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreBasraInteractor([]byte("not json"), cp)
	assert.Error(t, err)

	var g domain.Basra
	require.NoError(t, json.Unmarshal(data, &g))
}
