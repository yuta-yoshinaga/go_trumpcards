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

const tablanetMockOutput = `{"phase":1}`

func newTablanetMocks() (*interfaces.MockTablanetGame, *presenter.MockTablanetPresenter) {
	return new(interfaces.MockTablanetGame), new(presenter.MockTablanetPresenter)
}

func TestNewTablanetInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockTablanetPresenter)
	assert.PanicsWithValue(t, "TablanetInteractor: bg must not be nil", func() {
		usecase.NewTablanetInteractor(nil, cpMock)
	})
	gameMock := new(interfaces.MockTablanetGame)
	assert.PanicsWithValue(t, "TablanetInteractor: cp must not be nil", func() {
		usecase.NewTablanetInteractor(gameMock, nil)
	})
}

func TestTablanetInteractor_Play_Error(t *testing.T) {
	gm, cp := newTablanetMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(tablanetMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerPlay", 0, mock.Anything).Return(assert.AnError)

	bi := usecase.NewTablanetInteractor(gm, cp)
	assert.Equal(t, tablanetMockOutput, bi.Play(0, nil))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestTablanetInteractor_Play_NotPlayable(t *testing.T) {
	gm, cp := newTablanetMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(tablanetMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)

	bi := usecase.NewTablanetInteractor(gm, cp)
	assert.Equal(t, tablanetMockOutput, bi.Play(0, nil))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestTablanetInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, cp := newTablanetMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(tablanetMockOutput)
	bi := usecase.NewTablanetInteractor(gm, cp)
	out := bi.ResetWithConfig(domain.TablanetConfig{CpuDifficulty: 99})
	assert.Equal(t, tablanetMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestTablanetInteractor_HintAndLog(t *testing.T) {
	gm, cp := newTablanetMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	bi := usecase.NewTablanetInteractor(gm, cp)
	assert.Equal(t, "hint", bi.Hint())
	assert.Equal(t, "log", bi.ActionLog())
}

func TestTablanetInteractor_GetConfig(t *testing.T) {
	gm, cp := newTablanetMocks()
	cfg := domain.TablanetConfig{CpuDifficulty: domain.TablanetCpuDifficultyEasy}
	gm.On("GetConfig").Return(cfg)
	bi := usecase.NewTablanetInteractor(gm, cp)
	assert.Equal(t, cfg, bi.GetConfig())
}

// TestTablanetInteractor_RealFlow は本物のドメインで Reset→Play→…→NextRound の流れを
// advance() 経由で駆動し、ゲーム終了に到達することを検証する。
func TestTablanetInteractor_RealFlow(t *testing.T) {
	cp := new(presenter.MockTablanetPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(tablanetMockOutput)

	g := domain.NewDefaultTablanet()
	cfg := domain.DefaultTablanetConfig()
	cfg.CpuDifficulty = domain.TablanetCpuDifficultyNormal
	g.SetConfig(cfg)
	bi := usecase.NewTablanetInteractor(g, cp)

	bi.Reset()

	for step := 0; step < 5000 && !g.GetGameEndFlag(); step++ {
		if g.GetPhase() != domain.TablanetPhasePlay {
			break
		}
		require.True(t, g.IsHumanTurn(), "advance() は人間手番でのみ停止する")
		opts := g.GetCaptureOptions(g.GetCurrentTurn())
		bi.Play(0, opts[0])
	}
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.TablanetPhaseGameEnd, g.GetPhase())

	// NextRound で新規ゲームが始まる。
	bi.NextRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, domain.TablanetPhasePlay, g.GetPhase())
}

// TestTablanetInteractor_GameEndScoresImmediately は、ゲーム終了に到達した時点で
// (NextRound を呼ぶ前に) すでに得点内訳 (lastDealDetail) が埋まっていることを検証する。
// これがないとフロントエンドは結果画面を描画できない (MUST 修正)。
func TestTablanetInteractor_GameEndScoresImmediately(t *testing.T) {
	cp := new(presenter.MockTablanetPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(tablanetMockOutput)

	g := domain.NewDefaultTablanet()
	bi := usecase.NewTablanetInteractor(g, cp)
	bi.Reset()

	for step := 0; step < 5000 && !g.GetGameEndFlag(); step++ {
		if g.GetPhase() != domain.TablanetPhasePlay {
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

func TestTablanetInteractor_SnapshotAndRestore(t *testing.T) {
	cp := new(presenter.MockTablanetPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(tablanetMockOutput)

	real := usecase.NewTablanetInteractor(domain.NewDefaultTablanet(), cp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreTablanetInteractor(data, cp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreTablanetInteractor([]byte("not json"), cp)
	assert.Error(t, err)

	var g domain.Tablanet
	require.NoError(t, json.Unmarshal(data, &g))
}
