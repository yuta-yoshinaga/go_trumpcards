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

// cirullaPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type cirullaPassThrough struct{}

func (cirullaPassThrough) Output(_ interfaces.CirullaGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (cirullaPassThrough) HintOutput(_ interfaces.CirullaGame) string      { return "hint" }
func (cirullaPassThrough) ActionLogOutput(_ interfaces.CirullaGame) string { return "log" }

func newCirullaReal() (*usecase.CirullaInteractor, *domain.Cirulla) {
	g := domain.NewDefaultCirulla()
	return usecase.NewCirullaInteractor(g, cirullaPassThrough{}), g
}

func TestNewCirullaInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockCirullaPresenter)
	assert.PanicsWithValue(t, "CirullaInteractor: g must not be nil", func() {
		usecase.NewCirullaInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "CirullaInteractor: cp must not be nil", func() {
		usecase.NewCirullaInteractor(new(interfaces.MockCirullaGame), nil)
	})
}

// **開幕は人間の手番。** 親を席 1 にしてあるので席 0 から動く ── ここを 0 に
// すると、人間は最初の捕獲を選べないまま CPU が場を取り切る。
func TestCirullaInteractor_ResetStopsAtTheHuman(t *testing.T) {
	ci, g := newCirullaReal()
	require.Equal(t, "ok", ci.Reset())
	assert.Equal(t, domain.CirullaPhasePlay, g.GetPhase())
	assert.True(t, g.IsHumanTurn())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.CirullaHandSize, g.GetPlayer(0).GetCardsSize())
	assert.Len(t, g.GetTable(), domain.CirullaTableSize)
}

func TestCirullaInteractor_ResetWithConfig(t *testing.T) {
	ci, _ := newCirullaReal()
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CirullaConfig{
		CpuDifficulty: domain.CirullaCpuDifficultyEasy,
		TargetScore:   21,
	}))
	assert.Equal(t, 21, ci.GetConfig().TargetScore)

	out := ci.ResetWithConfig(domain.CirullaConfig{TargetScore: 999})
	assert.Contains(t, out, "err:")
	assert.Equal(t, 21, ci.GetConfig().TargetScore, "弾いた設定が入ってしまっている")
}

// **打ったら手番が人間に戻ってくる。** インタラクターの仕事は CPU を回し切って
// 人間の番か区切りで止めること ── 止めなければ CUI も Web も入力を待てない。
func TestCirullaInteractor_PlayRunsTheCpuAndComesBack(t *testing.T) {
	ci, g := newCirullaReal()
	require.Equal(t, "ok", ci.Reset())
	hand := g.GetPlayer(0).GetCardsSize()

	hint := g.GetHint()
	require.GreaterOrEqual(t, hint.HandIdx, 0)
	require.Equal(t, "ok", ci.Play(hint.HandIdx, hint.CaptureIdxs))

	assert.NotEqual(t, hand, g.GetPlayer(0).GetCardsSize(), "手札が減っていない")
	if g.GetPhase() == domain.CirullaPhasePlay {
		assert.True(t, g.IsHumanTurn(), "CPU の手番のまま止まっている")
		assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	}
}

// **不正な捕獲は盤面を動かさない。**
func TestCirullaInteractor_RejectsAnImpossibleCapture(t *testing.T) {
	ci, g := newCirullaReal()
	require.Equal(t, "ok", ci.Reset())
	g.SetTableForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 4, true),
		domain.NewCard(domain.CardDesignClover, 9, true),
	})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 4, true))

	assert.Contains(t, ci.Play(0, []int{1}), "err:", "9 が 4 で取れてしまっている")
	assert.Len(t, g.GetTable(), 2, "弾いた手で場が動いた")
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize(), "弾いた手で札が減った")

	assert.Contains(t, ci.Play(9, nil), "err:")
	assert.Contains(t, ci.Play(0, []int{5}), "err:", "場に無い番号が通ってしまった")
}

// **ラウンド終了では勝手に進まない。** 集計を読む時間を人間に渡す。
func TestCirullaInteractor_StopsAtRoundEnd(t *testing.T) {
	ci, g := newCirullaReal()
	require.Equal(t, "ok", ci.Reset())
	cirullaFinishRound(t, ci, g)
	require.Equal(t, domain.CirullaPhaseRoundEnd, g.GetPhase())

	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Len(t, res.Lines, 8)

	require.Equal(t, "ok", ci.NextRound())
	assert.Equal(t, domain.CirullaPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.True(t, g.IsHumanTurn(), "2 ラウンド目で人間の手番に戻らない")
}

// **札は湧かない。** 山・手札・場・取り札の総和は常に 40 枚。
func TestCirullaInteractor_CardsAreConserved(t *testing.T) {
	ci, g := newCirullaReal()
	require.Equal(t, "ok", ci.Reset())
	total := func() int {
		n := g.GetDeckRemaining() + len(g.GetTable())
		for i := 0; i < g.GetPlayerCnt(); i++ {
			n += g.GetPlayer(i).GetCardsSize() + len(g.GetPlayer(i).GetCaptured())
		}
		return n
	}
	for round := 0; round < 40 && !g.GetGameEndFlag(); round++ {
		assert.Equal(t, domain.CirullaDeckSize, total(), "ラウンド %d の頭で札が合わない", round+1)
		cirullaFinishRound(t, ci, g)
		assert.Equal(t, domain.CirullaDeckSize, total(), "ラウンド %d の終わりで札が合わない", round+1)
		require.Equal(t, "ok", ci.NextRound())
	}
}

// **目標点に届けば終わる。**
func TestCirullaInteractor_ReachesTheTargetScore(t *testing.T) {
	ci, g := newCirullaReal()
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CirullaConfig{
		CpuDifficulty: domain.CirullaCpuDifficultyEasy,
		TargetScore:   domain.CirullaMinTarget,
	}))
	for round := 0; round < 200 && !g.GetGameEndFlag(); round++ {
		cirullaFinishRound(t, ci, g)
		require.Equal(t, "ok", ci.NextRound())
	}
	require.True(t, g.GetGameEndFlag(), "11 点勝負でも終局に届かない")
	assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
	// 終局後の操作は盤面を触らない。
	assert.Equal(t, "ok", ci.NextRound())
	assert.Equal(t, "ok", ci.Play(0, nil))
	assert.True(t, g.GetGameEndFlag())
}

func TestCirullaInteractor_HintAndActionLog(t *testing.T) {
	ci, _ := newCirullaReal()
	require.Equal(t, "ok", ci.Reset())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

// **保存した盤で打ち続けられる。**
func TestCirullaInteractor_SnapshotRestoreKeepsPlaying(t *testing.T) {
	ci, g := newCirullaReal()
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CirullaConfig{
		CpuDifficulty: domain.CirullaCpuDifficultyEasy,
		TargetScore:   domain.CirullaMinTarget,
	}))
	hint := g.GetHint()
	require.GreaterOrEqual(t, hint.HandIdx, 0)
	require.Equal(t, "ok", ci.Play(hint.HandIdx, hint.CaptureIdxs))

	data, err := ci.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored, err := usecase.RestoreCirullaInteractor(data, cirullaPassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), rg.GetRoundNumber())
	assert.Equal(t, g.GetDeckRemaining(), rg.GetDeckRemaining(), "山の位置が消えている")
	assert.Equal(t, g.GetLastCapturer(), rg.GetLastCapturer(), "最後に取った席が消えている")
	assert.Equal(t, len(g.GetTable()), len(rg.GetTable()))
	assert.Equal(t, domain.CirullaMinTarget, rg.GetConfig().TargetScore, "設定が消えている")
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, g.GetPlayer(i).GetScore(), rg.GetPlayer(i).GetScore(), "席 %d の得点", i)
		assert.Equal(t, len(g.GetPlayer(i).GetCaptured()), len(rg.GetPlayer(i).GetCaptured()),
			"席 %d の取り札", i)
	}

	for round := 0; round < 200 && !rg.GetGameEndFlag(); round++ {
		cirullaFinishRound(t, restored, rg)
		require.Equal(t, "ok", restored.NextRound())
	}
	assert.True(t, rg.GetGameEndFlag(), "復元した盤で終局に届かない")
}

func TestRestoreCirullaInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreCirullaInteractor([]byte("{"), cirullaPassThrough{})
	assert.Error(t, err)
}

// cirullaFinishRound は現在のラウンドを終了まで打つ。
func cirullaFinishRound(t *testing.T, ci *usecase.CirullaInteractor, g interfaces.CirullaGame) {
	t.Helper()
	for step := 0; step < 500 && g.GetPhase() == domain.CirullaPhasePlay; step++ {
		hint := g.GetHint()
		require.GreaterOrEqual(t, hint.HandIdx, 0, "手番なのに指せる札が無い")
		require.Equal(t, "ok", ci.Play(hint.HandIdx, hint.CaptureIdxs))
	}
	require.NotEqual(t, domain.CirullaPhasePlay, g.GetPhase(), "ラウンドが終わらない")
}
