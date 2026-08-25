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

// sutdaPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type sutdaPassThrough struct{}

func (sutdaPassThrough) Output(_ interfaces.SutdaGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (sutdaPassThrough) HintOutput(_ interfaces.SutdaGame) string      { return "hint" }
func (sutdaPassThrough) ActionLogOutput(_ interfaces.SutdaGame) string { return "log" }

func newSutdaReal() (*usecase.SutdaInteractor, *domain.Sutda) {
	g := domain.NewDefaultSutda()
	return usecase.NewSutdaInteractor(g, sutdaPassThrough{}), g
}

func TestNewSutdaInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.MockSutdaPresenter)
	assert.PanicsWithValue(t, "SutdaInteractor: g must not be nil", func() {
		usecase.NewSutdaInteractor(nil, sp)
	})
	assert.PanicsWithValue(t, "SutdaInteractor: sp must not be nil", func() {
		usecase.NewSutdaInteractor(new(interfaces.MockSutdaGame), nil)
	})
}

// **開幕は人間の手番。** 親を最後の席にしてあるので席 0 から動く ── ここを 0 に
// すると、人間は一度もベットを選べないまま卓が回る。
func TestSutdaInteractor_ResetStopsAtTheHuman(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.Reset())
	assert.Equal(t, domain.SutdaPhaseBet, g.GetPhase())
	assert.True(t, g.IsHumanTurn())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.SutdaHandSize, g.GetPlayer(0).GetCardsSize())
}

func TestSutdaInteractor_ResetWithConfig(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.ResetWithConfig(domain.SutdaConfig{
		CpuDifficulty: domain.SutdaCpuDifficultyEasy,
		Seats:         5,
		StartChips:    500,
	}))
	assert.Equal(t, 5, si.GetConfig().Seats)
	assert.Equal(t, 5, g.GetPlayerCnt(), "席数が反映されていない")
	assert.Equal(t, 500-domain.SutdaAnte, g.GetPlayer(0).GetChips())

	out := si.ResetWithConfig(domain.SutdaConfig{Seats: 99, StartChips: 500})
	assert.Contains(t, out, "err:")
	assert.Equal(t, 5, si.GetConfig().Seats, "弾いた設定が入ってしまっている")
}

// **降りたら手番は戻らない。** 降りた直後にショーダウンへ届く。
func TestSutdaInteractor_FoldRunsTheHandOut(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.Reset())
	require.Equal(t, "ok", si.Action(domain.SutdaActionFold))
	assert.True(t, g.GetPlayer(0).IsFolded())
	assert.NotEqual(t, domain.SutdaPhaseBet, g.GetPhase(), "降りたのにベッティングが続いている")
}

// **レイズしても盤面は自分だけで終わらない。** 上げられた側が応じる番になる。
func TestSutdaInteractor_RaiseGivesTheOthersTheTurnBack(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.Reset())
	before := g.GetCurrentBet()
	require.Equal(t, "ok", si.Action(domain.SutdaActionRaise))
	assert.Greater(t, g.GetCurrentBet(), before)
	// CPU が応じ終えるまで回るので、戻ってくるのは人間の番かショーダウン。
	assert.True(t, g.IsHumanTurn() || g.GetPhase() != domain.SutdaPhaseBet)
}

func TestSutdaInteractor_RejectsAnUnknownAction(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.Reset())
	before := g.GetPot()
	assert.Contains(t, si.Action("zzz"), "err:")
	assert.Equal(t, before, g.GetPot(), "弾いた行動でポットが動いた")
}

// **1 ハンドを最後まで打てる。** ショーダウンで役が確定する。
func TestSutdaInteractor_PlaysAHandThrough(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.Reset())
	sutdaFinishHand(t, si, g)
	require.NotEqual(t, domain.SutdaPhaseBet, g.GetPhase())
	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Positive(t, res.Pot)
	assert.NotEmpty(t, res.Winners)
}

// **チップは湧かない。** ハンドを跨いでも卓の総額は変わらない。
func TestSutdaInteractor_ChipsAreConserved(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.Reset())
	total := func() int {
		n := g.GetPot()
		for i := 0; i < g.GetPlayerCnt(); i++ {
			n += g.GetPlayer(i).GetChips()
		}
		return n
	}
	want := total()
	for hand := 0; hand < 30 && !g.GetGameEndFlag(); hand++ {
		sutdaFinishHand(t, si, g)
		assert.Equal(t, want, total(), "ハンド %d でチップが合わない", hand+1)
		require.Equal(t, "ok", si.NextHand())
	}
}

// **チップが尽きたら終局する。**
func TestSutdaInteractor_StopsWhenSomeoneIsBroke(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.ResetWithConfig(domain.SutdaConfig{
		CpuDifficulty: domain.SutdaCpuDifficultyEasy,
		Seats:         2,
		StartChips:    domain.SutdaMinChips,
	}))
	for hand := 0; hand < 3000 && !g.GetGameEndFlag(); hand++ {
		sutdaFinishHand(t, si, g)
		require.Equal(t, "ok", si.NextHand())
	}
	require.True(t, g.GetGameEndFlag(), "チップが尽きても終わらない")
	assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
	// 終局後の操作は盤面を触らない。
	assert.Equal(t, "ok", si.NextHand())
	assert.Equal(t, "ok", si.Action(domain.SutdaActionCall))
	assert.True(t, g.GetGameEndFlag())
}

func TestSutdaInteractor_HintAndActionLog(t *testing.T) {
	si, _ := newSutdaReal()
	require.Equal(t, "ok", si.Reset())
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

// **保存した盤で打ち続けられる。**
func TestSutdaInteractor_SnapshotRestoreKeepsPlaying(t *testing.T) {
	si, g := newSutdaReal()
	require.Equal(t, "ok", si.Reset())
	require.Equal(t, "ok", si.Action(domain.SutdaActionRaise))

	data, err := si.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored, err := usecase.RestoreSutdaInteractor(data, sutdaPassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetPot(), rg.GetPot())
	assert.Equal(t, g.GetCurrentBet(), rg.GetCurrentBet())
	assert.Equal(t, g.GetRaiseCount(), rg.GetRaiseCount(), "レイズ回数が消えている")
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, g.GetPlayer(i).GetChips(), rg.GetPlayer(i).GetChips(), "席 %d のチップ", i)
	}

	for hand := 0; hand < 3000 && !rg.GetGameEndFlag(); hand++ {
		sutdaFinishHand(t, restored, rg)
		require.Equal(t, "ok", restored.NextHand())
	}
	assert.True(t, rg.GetGameEndFlag(), "復元した盤で終局に届かない")
}

func TestRestoreSutdaInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreSutdaInteractor([]byte("{"), sutdaPassThrough{})
	assert.Error(t, err)
}

// sutdaFinishHand は現在のハンドをショーダウンまで打つ。
func sutdaFinishHand(t *testing.T, si *usecase.SutdaInteractor, g interfaces.SutdaGame) {
	t.Helper()
	for step := 0; step < 500 && g.GetPhase() == domain.SutdaPhaseBet; step++ {
		require.Equal(t, "ok", si.Action(domain.SutdaActionCall))
	}
	require.NotEqual(t, domain.SutdaPhaseBet, g.GetPhase(), "ベッティングが終わらない")
}
