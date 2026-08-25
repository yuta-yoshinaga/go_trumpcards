//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// cometPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type cometPassThrough struct{}

func (cometPassThrough) Output(_ interfaces.CometGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (cometPassThrough) HintOutput(_ interfaces.CometGame) string      { return "hint" }
func (cometPassThrough) ActionLogOutput(_ interfaces.CometGame) string { return "log" }

func newCometReal() (*usecase.CometInteractor, *domain.Comet) {
	g := domain.NewDefaultComet()
	return usecase.NewCometInteractor(g, cometPassThrough{}), g
}

func TestNewCometInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockCometPresenter)
	assert.PanicsWithValue(t, "CometInteractor: g must not be nil", func() {
		usecase.NewCometInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "CometInteractor: cp must not be nil", func() {
		usecase.NewCometInteractor(new(interfaces.MockCometGame), nil)
	})
}

// **開幕は人間の手番。** 親の左隣が先に打つ規則なので親を最後の席にしてある ──
// 親を 0 にすると人間は最初の連なりの先頭を選べない。
func TestCometInteractor_ResetStopsAtTheHuman(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.Reset())
	assert.Equal(t, domain.CometPhasePlay, g.GetPhase())
	assert.True(t, g.IsHumanTurn())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.Equal(t, 0, g.GetNeed(), "先頭なのに要るランクが決まっている")
	assert.Positive(t, g.GetPlayer(0).GetCardsSize())
}

func TestCometInteractor_ResetWithConfig(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CometConfig{
		CpuDifficulty: domain.CometCpuDifficultyEasy,
		Players:       3,
		TargetScore:   50,
	}))
	assert.Equal(t, 50, ci.GetConfig().TargetScore)
	assert.Equal(t, 3, g.GetPlayerCnt(), "席数が反映されていない")

	out := ci.ResetWithConfig(domain.CometConfig{Players: 9, TargetScore: 50})
	assert.Contains(t, out, "err:")
	assert.Equal(t, 3, g.GetPlayerCnt(), "弾いた設定が入ってしまっている")
}

// **打ったら手番が人間に戻ってくる。** インタラクターの仕事は CPU を回し切って
// 人間の番か区切りで止めること ── 止めなければ CUI も Web も入力を待てない。
func TestCometInteractor_PlayRunsTheCpuAndComesBack(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.Reset())
	hand := g.GetPlayer(0).GetCardsSize()

	h := g.GetHint()
	require.GreaterOrEqual(t, h.HandIdx, 0)
	require.Equal(t, "ok", ci.Play(h.HandIdx))

	assert.Equal(t, hand-1, g.GetPlayer(0).GetCardsSize(), "手札が減っていない")
	if g.GetPhase() == domain.CometPhasePlay {
		assert.True(t, g.IsHumanTurn(), "CPU の手番のまま止まっている")
	}
}

// **出せる札があるならパスできない。**
func TestCometInteractor_RejectsAPassWithAPlayableCard(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.Reset())
	require.NotEmpty(t, g.PlayableIdxs(0))
	hand := g.GetPlayer(0).GetCardsSize()
	assert.Contains(t, ci.Pass(), "err:", "出せるのにパスできてしまう")
	assert.Equal(t, hand, g.GetPlayer(0).GetCardsSize(), "弾いた手で札が減った")
}

// **弾いた手は盤面を動かさない。**
func TestCometInteractor_RejectsAnIllegalPlay(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.Reset())
	hand := g.GetPlayer(0).GetCardsSize()
	assert.Contains(t, ci.Play(99), "err:")
	assert.Contains(t, ci.Play(-1), "err:")
	assert.Equal(t, hand, g.GetPlayer(0).GetCardsSize(), "弾いた手で札が減った")
}

// **局の区切りでは勝手に進まない。** 集計を読む時間を人間に渡す。
func TestCometInteractor_StopsAtRoundEnd(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.Reset())
	cometFinishRound(t, ci, g)
	require.Equal(t, domain.CometPhaseRoundEnd, g.GetPhase())

	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, 0, res.CardsLeft[res.WinnerIdx], "上がった席に札が残っている")

	require.Equal(t, "ok", ci.NextRound())
	assert.Equal(t, domain.CometPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.True(t, g.IsHumanTurn(), "2 局目で人間の手番に戻らない")
}

// **札は湧かない。** 手札と連なりと死に手の総和は常に 51 枚。
func TestCometInteractor_CardsAreConserved(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.Reset())
	total := func() int {
		n := len(g.GetPile()) + g.GetDeadCount()
		for i := 0; i < g.GetPlayerCnt(); i++ {
			n += g.GetPlayer(i).GetCardsSize()
		}
		return n
	}
	for round := 0; round < 15 && !g.GetGameEndFlag(); round++ {
		assert.Equal(t, domain.CometDeckSize, total(), "局 %d の頭で札が合わない", round+1)
		cometFinishRound(t, ci, g)
		assert.Equal(t, domain.CometDeckSize, total(), "局 %d の終わりで札が合わない", round+1)
		require.Equal(t, "ok", ci.NextRound())
	}
}

// **目標点に届けば終わる。**
func TestCometInteractor_ReachesTheTargetScore(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CometConfig{
		CpuDifficulty: domain.CometCpuDifficultyEasy,
		Players:       domain.CometDefaultPlayers,
		TargetScore:   domain.CometMinTarget,
	}))
	for round := 0; round < 200 && !g.GetGameEndFlag(); round++ {
		cometFinishRound(t, ci, g)
		require.Equal(t, "ok", ci.NextRound())
	}
	require.True(t, g.GetGameEndFlag(), "20 点勝負でも終局に届かない")
	assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
	// 終局後の操作は盤面を触らない。
	assert.Equal(t, "ok", ci.NextRound())
	assert.True(t, g.GetGameEndFlag())
}

func TestCometInteractor_HintAndActionLog(t *testing.T) {
	ci, _ := newCometReal()
	require.Equal(t, "ok", ci.Reset())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

// **保存した盤で打ち続けられる。**
func TestCometInteractor_SnapshotRestoreKeepsPlaying(t *testing.T) {
	ci, g := newCometReal()
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CometConfig{
		CpuDifficulty: domain.CometCpuDifficultyEasy,
		Players:       domain.CometDefaultPlayers,
		TargetScore:   domain.CometMinTarget,
	}))
	h := g.GetHint()
	require.GreaterOrEqual(t, h.HandIdx, 0)
	require.Equal(t, "ok", ci.Play(h.HandIdx))

	data, err := ci.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored, err := usecase.RestoreCometInteractor(data, cometPassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetNeed(), rg.GetNeed(), "次に要るランクが消えている")
	assert.Equal(t, g.GetDeadCount(), rg.GetDeadCount(), "死に手が消えている")
	assert.Equal(t, len(g.GetPile()), len(rg.GetPile()), "連なりが消えている")
	assert.Equal(t, domain.CometMinTarget, rg.GetConfig().TargetScore, "設定が消えている")
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, g.GetPlayer(i).GetCardsSize(), rg.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
		assert.Equal(t, g.GetPlayer(i).GetScore(), rg.GetPlayer(i).GetScore(), "席 %d の得点", i)
	}

	for round := 0; round < 200 && !rg.GetGameEndFlag(); round++ {
		cometFinishRound(t, restored, rg)
		require.Equal(t, "ok", restored.NextRound())
	}
	assert.True(t, rg.GetGameEndFlag(), "復元した盤で終局に届かない")
}

func TestRestoreCometInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreCometInteractor([]byte("{"), cometPassThrough{})
	assert.Error(t, err)
}

// cometFinishRound は現在の局を終了まで打つ。
func cometFinishRound(t *testing.T, ci *usecase.CometInteractor, g interfaces.CometGame) {
	t.Helper()
	for step := 0; step < 500 && g.GetPhase() == domain.CometPhasePlay; step++ {
		h := g.GetHint()
		if h.HandIdx < 0 {
			require.Equal(t, "ok", ci.Pass())
			continue
		}
		require.Equal(t, "ok", ci.Play(h.HandIdx))
	}
	require.NotEqual(t, domain.CometPhasePlay, g.GetPhase(), "局が終わらない")
}
