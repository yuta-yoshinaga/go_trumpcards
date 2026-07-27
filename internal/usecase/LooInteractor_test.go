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

const looMockOutput = `{"phase":1}`

func newLooMocks() (*interfaces.MockLooGame, *presenter.MockLooPresenter) {
	return new(interfaces.MockLooGame), new(presenter.MockLooPresenter)
}

func TestNewLooInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockLooPresenter)
	assert.PanicsWithValue(t, "LooInteractor: lg must not be nil", func() {
		usecase.NewLooInteractor(nil, cpMock)
	})
	gameMock := new(interfaces.MockLooGame)
	assert.PanicsWithValue(t, "LooInteractor: cp must not be nil", func() {
		usecase.NewLooInteractor(gameMock, nil)
	})
}

func TestLooInteractor_Decide_Error(t *testing.T) {
	gm, cp := newLooMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(looMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerDecide", true).Return(assert.AnError)

	li := usecase.NewLooInteractor(gm, cp)
	assert.Equal(t, looMockOutput, li.Decide(true))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestLooInteractor_Decide_GameEnded(t *testing.T) {
	gm, cp := newLooMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(looMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	li := usecase.NewLooInteractor(gm, cp)
	assert.Equal(t, looMockOutput, li.Decide(true))
	gm.AssertNotCalled(t, "PlayerDecide", mock.Anything)
}

func TestLooInteractor_Play_NotPlayable(t *testing.T) {
	gm, cp := newLooMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(looMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)

	li := usecase.NewLooInteractor(gm, cp)
	assert.Equal(t, looMockOutput, li.Play(0))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestLooInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, cp := newLooMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(looMockOutput)
	li := usecase.NewLooInteractor(gm, cp)
	out := li.ResetWithConfig(domain.LooConfig{CpuDifficulty: 99, Ante: 3})
	assert.Equal(t, looMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestLooInteractor_HintAndLog(t *testing.T) {
	gm, cp := newLooMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	li := usecase.NewLooInteractor(gm, cp)
	assert.Equal(t, "hint", li.Hint())
	assert.Equal(t, "log", li.ActionLog())
}

func TestLooInteractor_GetConfig(t *testing.T) {
	gm, cp := newLooMocks()
	cfg := domain.LooConfig{CpuDifficulty: domain.LooCpuDifficultyEasy, Ante: 3}
	gm.On("GetConfig").Return(cfg)
	li := usecase.NewLooInteractor(gm, cp)
	assert.Equal(t, cfg, li.GetConfig())
}

// TestLooInteractor_RealFlow は本物のドメインで Reset→Decide→Play→NextRound の
// 一連の流れを advance() 経由で駆動する。
func TestLooInteractor_RealFlow(t *testing.T) {
	cp := new(presenter.MockLooPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(looMockOutput)

	g := domain.NewDefaultLoo()
	cfg := domain.DefaultLooConfig()
	cfg.CpuDifficulty = domain.LooCpuDifficultyNormal
	g.SetConfig(cfg)
	li := usecase.NewLooInteractor(g, cp)

	li.Reset()
	assert.False(t, g.GetGameEndFlag())

	// decide フェーズを進める。
	for step := 0; step < 100 && g.GetPhase() == domain.LooPhaseDecide; step++ {
		if g.IsHumanTurn() {
			li.Decide(true)
		} else {
			break
		}
	}

driveLoop:
	for step := 0; step < 500; step++ {
		switch g.GetPhase() {
		case domain.LooPhasePlay:
			if g.IsHumanTurn() {
				idx := g.GetPlayableIndices(g.GetCurrentTurn())
				require.NotEmpty(t, idx)
				li.Play(idx[0])
			} else {
				break driveLoop
			}
		case domain.LooPhaseRoundEnd:
			li.NextRound()
			break driveLoop
		case domain.LooPhaseDecide:
			if g.IsHumanTurn() {
				li.Decide(true)
			} else {
				break driveLoop
			}
		default:
			break driveLoop
		}
	}
	// 少なくとも 1 ディールは進行している。
	assert.GreaterOrEqual(t, g.GetRoundNumber(), 1)
}

// TestLooInteractor_RoundEndScoresImmediately は、ディールが RoundEnd に到達した
// 時点で (NextRound を呼ぶ前に) すでに精算済み (lastDealDetail が埋まり、チップが
// 反映されている) ことを検証する。これがないとフロントエンドはディール結果画面を
// 表示できない (MUST 修正)。
func TestLooInteractor_RoundEndScoresImmediately(t *testing.T) {
	cp := new(presenter.MockLooPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(looMockOutput)

	g := domain.NewDefaultLoo()
	cfg := domain.DefaultLooConfig()
	cfg.CpuDifficulty = domain.LooCpuDifficultyNormal
	g.SetConfig(cfg)
	li := usecase.NewLooInteractor(g, cp)

	li.Reset()

	reachedRoundEnd := false
	for step := 0; step < 1000; step++ {
		switch g.GetPhase() {
		case domain.LooPhaseDecide:
			if g.IsHumanTurn() {
				li.Decide(true)
			} else {
				t.Fatalf("advance() が decide フェーズで CPU 手番のまま停止した")
			}
		case domain.LooPhasePlay:
			require.True(t, g.IsHumanTurn(), "play フェーズで停止するのは人間手番のときのみ")
			idx := g.GetPlayableIndices(g.GetCurrentTurn())
			require.NotEmpty(t, idx)
			li.Play(idx[0])
		case domain.LooPhaseRoundEnd:
			reachedRoundEnd = true
		default:
			t.Fatalf("想定外のフェーズ %v", g.GetPhase())
		}
		if reachedRoundEnd {
			break
		}
	}
	require.True(t, reachedRoundEnd, "1000 ステップ内に RoundEnd に到達しなかった")

	// NextRound を呼ぶ *前* に精算済みであること。
	det := g.GetLastDealDetail()
	require.NotNil(t, det, "RoundEnd 到達時点で精算済み (lastDealDetail 非 nil) であるべき")
	require.NotNil(t, det.Tricks)

	// 二重精算防止: RoundEnd で ScoreRound を再度呼んでもチップは変わらない。
	before := make([]int, domain.LooPlayerCnt)
	for i := 0; i < domain.LooPlayerCnt; i++ {
		before[i] = g.GetPlayer(i).GetChips()
	}
	g.ScoreRound()
	for i := 0; i < domain.LooPlayerCnt; i++ {
		assert.Equal(t, before[i], g.GetPlayer(i).GetChips(), "scored フラグにより二重精算されない")
	}
}

// TestLooInteractor_HumanCompletesTrick は人間がトリックを完了しても
// ゲームが停止せず進行することを検証する (RECURRING LESSON 5)。
func TestLooInteractor_HumanCompletesTrick(t *testing.T) {
	cp := new(presenter.MockLooPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(looMockOutput)

	g := domain.NewDefaultLoo()
	g.Reset()
	// 2 人参加 (0=人間, 1=CPU)。人間が最後に出してトリックを完了させる状況を作る。
	g.GetPlayer(0).SetPlaying(true)
	g.GetPlayer(1).SetPlaying(true)
	g.GetPlayer(2).SetPlaying(false)
	g.GetPlayer(3).SetPlaying(false)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.LooPhasePlay)
	g.SetTrickNumber(1)
	g.SetLeadPlayerIdx(1)
	g.SetCurrentTurn(0)
	// CPU (1) が既に 1 枚出した。人間 (0) が最後の 1 枚を出すとトリック完了。
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
	})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	g.GetPlayer(1).Reset()
	for i := 0; i < 4; i++ {
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 2+i, false))
	}

	li := usecase.NewLooInteractor(g, cp)
	// 人間が最高のハートを出してトリック完了 → フリーズしないこと。
	li.Play(0)
	// トリックが解決され、次トリック (play) または play 継続へ進んでいるはず。
	assert.NotEqual(t, domain.LooPhaseTrickEnd, g.GetPhase())
}

func TestLooInteractor_SnapshotAndRestore(t *testing.T) {
	cp := new(presenter.MockLooPresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(looMockOutput)

	real := usecase.NewLooInteractor(domain.NewDefaultLoo(), cp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreLooInteractor(data, cp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreLooInteractor([]byte("not json"), cp)
	assert.Error(t, err)

	var g domain.Loo
	require.NoError(t, json.Unmarshal(data, &g))
}
