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

const cinchMockOutput = `{"phase":2}`

func newCinchMocks() (*interfaces.MockCinchGame, *presenter.MockCinchPresenter) {
	return new(interfaces.MockCinchGame), new(presenter.MockCinchPresenter)
}

func TestNewCinchInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockCinchPresenter)
	assert.PanicsWithValue(t, "CinchInteractor: cg must not be nil", func() {
		usecase.NewCinchInteractor(nil, cpMock)
	})
	gameMock := new(interfaces.MockCinchGame)
	assert.PanicsWithValue(t, "CinchInteractor: cp must not be nil", func() {
		usecase.NewCinchInteractor(gameMock, nil)
	})
}

func TestCinchInteractor_Bid_Error(t *testing.T) {
	gm, cp := newCinchMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(cinchMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerBid", 2).Return(assert.AnError)

	ci := usecase.NewCinchInteractor(gm, cp)
	assert.Equal(t, cinchMockOutput, ci.Bid(2))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestCinchInteractor_Bid_GameEnded(t *testing.T) {
	gm, cp := newCinchMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(cinchMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ci := usecase.NewCinchInteractor(gm, cp)
	assert.Equal(t, cinchMockOutput, ci.Bid(2))
	gm.AssertNotCalled(t, "PlayerBid", mock.Anything)
}

func TestCinchInteractor_NameTrump_Error(t *testing.T) {
	gm, cp := newCinchMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(cinchMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("NameTrump", 0).Return(assert.AnError)

	ci := usecase.NewCinchInteractor(gm, cp)
	assert.Equal(t, cinchMockOutput, ci.NameTrump(0))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestCinchInteractor_Play_NotPlayable(t *testing.T) {
	gm, cp := newCinchMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(cinchMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)

	ci := usecase.NewCinchInteractor(gm, cp)
	assert.Equal(t, cinchMockOutput, ci.Play(0))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestCinchInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, cp := newCinchMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(cinchMockOutput)
	ci := usecase.NewCinchInteractor(gm, cp)
	out := ci.ResetWithConfig(domain.CinchConfig{CpuDifficulty: 99, PointLimit: 21})
	assert.Equal(t, cinchMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestCinchInteractor_HintAndLog(t *testing.T) {
	gm, cp := newCinchMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	ci := usecase.NewCinchInteractor(gm, cp)
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestCinchInteractor_GetConfig(t *testing.T) {
	gm, cp := newCinchMocks()
	cfg := domain.CinchConfig{CpuDifficulty: domain.CinchDifficultyEasy, PointLimit: 21}
	gm.On("GetConfig").Return(cfg)
	ci := usecase.NewCinchInteractor(gm, cp)
	assert.Equal(t, cfg, ci.GetConfig())
}

// TestCinchInteractor_RealFlow は本物のドメインを使って
// Reset→Bid→(NameTrump)→Play→NextRound の一連の流れで advance() を駆動する。
func TestCinchInteractor_RealFlow(t *testing.T) {
	cp := new(presenter.MockCinchPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(cinchMockOutput)

	// CPU を Easy にして人間手番が確実に来る保証はないため、ドメインの状態を直接検査する。
	g := domain.NewDefaultCinch()
	cfg := domain.DefaultCinchConfig()
	cfg.CpuDifficulty = domain.CinchDifficultyNormal
	g.SetConfig(cfg)
	ci := usecase.NewCinchInteractor(g, cp)

	ci.Reset()
	// Reset 後、人間 (bid 手番) かゲームが進行しているはず。
	assert.False(t, g.GetGameEndFlag())

	// bid フェーズを進める: human 手番ならパス、そうでなければ advance が既に消化済み。
	for step := 0; step < 100 && g.GetPhase() == domain.CinchPhaseBid; step++ {
		if g.IsHumanTurn() {
			ci.Bid(domain.CinchPassBid)
		} else {
			break
		}
	}
	// nameTrump が human なら宣言。
	if g.GetPhase() == domain.CinchPhaseNameTrump && g.IsHumanTurn() {
		ci.NameTrump(domain.CardDesignSpade)
	}
	// play フェーズを最後まで進める。
	for step := 0; step < 500 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.CinchPhasePlay:
			if g.IsHumanTurn() {
				idx := g.GetPlayableIndices(g.GetCurrentTurn())
				require.NotEmpty(t, idx)
				ci.Play(idx[0])
			} else {
				// advance がここまで進めているはずだが、念のため。
				break
			}
		case domain.CinchPhaseRoundEnd:
			ci.NextRound()
		default:
			// bid/nameTrump に戻った (次ディール)。human 操作で進める。
			if g.GetPhase() == domain.CinchPhaseBid && g.IsHumanTurn() {
				ci.Bid(domain.CinchPassBid)
			} else if g.GetPhase() == domain.CinchPhaseNameTrump && g.IsHumanTurn() {
				ci.NameTrump(domain.CardDesignSpade)
			} else {
				break
			}
		}
	}
	// 少なくとも 1 ディールは進行し、直前ディールの内訳が得られる。
	assert.NotNil(t, g.GetLastDealDetail())
}

func TestCinchInteractor_SnapshotAndRestore(t *testing.T) {
	cp := new(presenter.MockCinchPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(cinchMockOutput)

	real := usecase.NewCinchInteractor(domain.NewDefaultCinch(), cp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreCinchInteractor(data, cp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreCinchInteractor([]byte("not json"), cp)
	assert.Error(t, err)

	// round-trip via json.Marshal sanity.
	var g domain.Cinch
	require.NoError(t, json.Unmarshal(data, &g))
}
