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

// dehlaPakadPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type dehlaPakadPassThrough struct{}

func (dehlaPakadPassThrough) Output(_ interfaces.DehlaPakadGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (dehlaPakadPassThrough) HintOutput(_ interfaces.DehlaPakadGame) string      { return "hint" }
func (dehlaPakadPassThrough) ActionLogOutput(_ interfaces.DehlaPakadGame) string { return "log" }

func newDehlaPakadReal() (*usecase.DehlaPakadInteractor, *domain.DehlaPakad) {
	g := domain.NewDefaultDehlaPakad()
	return usecase.NewDehlaPakadInteractor(g, dehlaPakadPassThrough{}), g
}

func TestNewDehlaPakadInteractor_NilGuards(t *testing.T) {
	dp := new(presenter.MockDehlaPakadPresenter)
	assert.PanicsWithValue(t, "DehlaPakadInteractor: g must not be nil", func() {
		usecase.NewDehlaPakadInteractor(nil, dp)
	})
	assert.PanicsWithValue(t, "DehlaPakadInteractor: dp must not be nil", func() {
		usecase.NewDehlaPakadInteractor(new(interfaces.MockDehlaPakadGame), nil)
	})
}

// **開幕は人間が切り札を決める。** 席 3 が親なので、その右隣の席 0 が決める ──
// ここが CPU だと、親が動かない規則と噛み合って何ハンドも決められない。
func TestDehlaPakadInteractor_ResetStopsAtTheHumanTrumpCall(t *testing.T) {
	di, g := newDehlaPakadReal()
	require.Equal(t, "ok", di.Reset())
	assert.Equal(t, domain.DehlaPakadPhaseSelectTrump, g.GetPhase())
	assert.True(t, g.IsHumanTurn())
	assert.Equal(t, 0, g.GetTrumpChooserIdx(), "人間が切り札を決められない")
	assert.Equal(t, domain.DehlaPakadFirstBatch, g.GetPlayer(0).GetCardsSize())
}

// **CPU が決める番でも盤面は進む。** 宣言させずに抜けると止まる。
func TestDehlaPakadInteractor_CpuCallsTheTrumpWithoutBeingAsked(t *testing.T) {
	di, g := newDehlaPakadReal()
	require.Equal(t, "ok", di.Reset())
	require.Equal(t, "ok", di.SelectTrump(domain.CardDesignHeart))
	dehlaPakadFinishHand(t, di, g)
	require.Equal(t, "ok", di.NextHand())

	if g.GetPlayer(g.GetTrumpChooserIdx()).GetIsHuman() {
		t.Skip("親が動かず人間がまた決める番")
	}
	assert.NotEqual(t, domain.DehlaPakadPhaseSelectTrump, g.GetPhase(), "CPU が切り札を決めていない")
	assert.Greater(t, g.GetTrumpSuit(), 0)
}

func TestDehlaPakadInteractor_ResetWithConfig(t *testing.T) {
	di, g := newDehlaPakadReal()
	require.Equal(t, "ok", di.ResetWithConfig(domain.DehlaPakadConfig{
		CpuDifficulty: domain.DehlaPakadCpuDifficultyEasy,
		TargetKots:    1,
	}))
	assert.Equal(t, 1, di.GetConfig().TargetKots)
	assert.Equal(t, domain.DehlaPakadCpuDifficultyEasy, g.GetConfig().CpuDifficulty)

	out := di.ResetWithConfig(domain.DehlaPakadConfig{TargetKots: domain.DehlaPakadMaxKots + 1})
	assert.Contains(t, out, "err:")
	assert.Equal(t, 1, di.GetConfig().TargetKots, "弾いた設定が入ってしまっている")
}

func TestDehlaPakadInteractor_RejectsAnInvalidTrump(t *testing.T) {
	di, g := newDehlaPakadReal()
	require.Equal(t, "ok", di.Reset())
	assert.Contains(t, di.SelectTrump(99), "err:")
	assert.Equal(t, domain.DehlaPakadPhaseSelectTrump, g.GetPhase(), "弾いたのに進んでいる")
	assert.Equal(t, -1, g.GetTrumpSuit())
	// 残り 8 枚も配られていない。
	assert.Equal(t, domain.DehlaPakadFirstBatch, g.GetPlayer(0).GetCardsSize())
}

// 宣言フェーズでの Play は盤面を触らない。
func TestDehlaPakadInteractor_PlayBeforeTheTrumpIsANoop(t *testing.T) {
	di, g := newDehlaPakadReal()
	require.Equal(t, "ok", di.Reset())
	before := g.GetPlayer(0).GetCardsSize()
	di.Play(0)
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize(), "宣言前に札が減った")
}

// **1 ハンドを最後まで打てる。** 10 は 4 枚とも誰かの組に入る。
func TestDehlaPakadInteractor_PlaysAHandThrough(t *testing.T) {
	di, g := newDehlaPakadReal()
	require.Equal(t, "ok", di.Reset())
	require.Equal(t, "ok", di.SelectTrump(domain.CardDesignHeart))
	dehlaPakadFinishHand(t, di, g)

	assert.Equal(t, domain.DehlaPakadPhaseHandEnd, g.GetPhase())
	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, domain.DehlaPakadTenCnt, res.TeamTens[0]+res.TeamTens[1], "10 が宙に浮いている")
	assert.Empty(t, g.GetCentrePile(), "中央に札が残っている")
}

// **コートに届いたら終局する。** 1 コート設定なら 1 度のコートで終わる。
func TestDehlaPakadInteractor_StopsWhenTheKotTargetIsReached(t *testing.T) {
	di, g := newDehlaPakadReal()
	require.Equal(t, "ok", di.ResetWithConfig(domain.DehlaPakadConfig{
		CpuDifficulty: domain.DehlaPakadCpuDifficultyEasy,
		TargetKots:    1,
	}))

	for hand := 0; hand < 300 && !g.GetGameEndFlag(); hand++ {
		dehlaPakadFinishHand(t, di, g)
		require.Equal(t, "ok", di.NextHand())
	}
	require.True(t, g.GetGameEndFlag(), "コートに届いても終局しない")
	assert.GreaterOrEqual(t, g.GetWinnerTeam(), 0)
	assert.GreaterOrEqual(t, g.GetTeamKots()[g.GetWinnerTeam()], 1)
	// 終局後の操作は盤面を触らない。
	assert.Equal(t, "ok", di.NextHand())
	assert.Equal(t, "ok", di.Play(0))
	assert.True(t, g.GetGameEndFlag())
}

func TestDehlaPakadInteractor_HintAndActionLog(t *testing.T) {
	di, _ := newDehlaPakadReal()
	require.Equal(t, "ok", di.Reset())
	assert.Equal(t, "hint", di.Hint())
	assert.Equal(t, "log", di.ActionLog())
}

// **保存した盤で指し続けられる。** 2 連勝の記憶 (prevTrickWinner) と中央の山が
// 消えると、復元した盤は別のゲームになる。
func TestDehlaPakadInteractor_SnapshotRestoreKeepsPlaying(t *testing.T) {
	di, g := newDehlaPakadReal()
	require.Equal(t, "ok", di.Reset())
	require.Equal(t, "ok", di.SelectTrump(domain.CardDesignHeart))
	if g.GetPhase() == domain.DehlaPakadPhasePlay && g.IsHumanTurn() {
		require.Equal(t, "ok", di.Play(g.GetPlayableIndices(g.GetCurrentTurn())[0]))
	}

	data, err := di.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored, err := usecase.RestoreDehlaPakadInteractor(data, dehlaPakadPassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), rg.GetTrumpSuit())
	assert.Equal(t, g.GetDealerIdx(), rg.GetDealerIdx())
	assert.Equal(t, g.GetCurrentTurn(), rg.GetCurrentTurn())
	assert.Equal(t, g.GetPrevTrickWinner(), rg.GetPrevTrickWinner(), "2 連勝の記憶が消えている")
	assert.Len(t, rg.GetCentrePile(), len(g.GetCentrePile()), "中央の山が消えている")
	for i := 0; i < domain.DehlaPakadPlayerCnt; i++ {
		require.NotNil(t, rg.GetPlayer(i), "席 %d が空", i)
		assert.Equal(t, g.GetPlayer(i).GetCardsSize(), rg.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
	}

	// 復元した盤で最後まで打てる。
	for hand := 0; hand < 300 && !rg.GetGameEndFlag(); hand++ {
		dehlaPakadFinishHand(t, restored, rg)
		require.Equal(t, "ok", restored.NextHand())
	}
	assert.True(t, rg.GetGameEndFlag(), "復元した盤で終局に届かない")
}

func TestRestoreDehlaPakadInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreDehlaPakadInteractor([]byte("{"), dehlaPakadPassThrough{})
	assert.Error(t, err)
}

// dehlaPakadFinishHand は現在のハンドを HandEnd まで打つ。
func dehlaPakadFinishHand(t *testing.T, di *usecase.DehlaPakadInteractor, g interfaces.DehlaPakadGame) {
	t.Helper()
	for step := 0; step < 600; step++ {
		switch g.GetPhase() {
		case domain.DehlaPakadPhaseSelectTrump:
			require.Equal(t, "ok", di.SelectTrump(domain.CardDesignHeart))
		case domain.DehlaPakadPhasePlay:
			valid := g.GetPlayableIndices(g.GetCurrentTurn())
			require.NotEmpty(t, valid, "出せる札が 1 枚も無い")
			require.Equal(t, "ok", di.Play(valid[0]))
		default:
			return
		}
	}
	t.Fatal("ハンドが終わらない")
}
