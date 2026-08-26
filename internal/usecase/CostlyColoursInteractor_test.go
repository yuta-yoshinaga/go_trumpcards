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

// costlyPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type costlyPassThrough struct{}

func (costlyPassThrough) Output(_ interfaces.CostlyColoursGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (costlyPassThrough) HintOutput(_ interfaces.CostlyColoursGame) string      { return "hint" }
func (costlyPassThrough) ActionLogOutput(_ interfaces.CostlyColoursGame) string { return "log" }

func newCostlyReal() (*usecase.CostlyColoursInteractor, *domain.CostlyColours) {
	g := domain.NewDefaultCostlyColours()
	return usecase.NewCostlyColoursInteractor(g, costlyPassThrough{}), g
}

func TestNewCostlyColoursInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockCostlyColoursPresenter)
	assert.PanicsWithValue(t, "CostlyColoursInteractor: g must not be nil", func() {
		usecase.NewCostlyColoursInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "CostlyColoursInteractor: cp must not be nil", func() {
		usecase.NewCostlyColoursInteractor(new(interfaces.MockCostlyColoursGame), nil)
	})
}

// **開幕は交換フェーズで人間の手番。** ここを飛ばすと mog の無いゲームになる。
func TestCostlyColoursInteractor_ResetStopsAtTheMog(t *testing.T) {
	ci, g := newCostlyReal()
	require.Equal(t, "ok", ci.Reset())
	assert.Equal(t, domain.CostlyColoursPhaseMog, g.GetPhase())
	assert.True(t, g.IsHumanTurn())
	assert.Equal(t, domain.CostlyColoursHandSize, g.GetPlayer(0).GetCardsSize())
	assert.NotNil(t, g.GetTurnUp(), "トランプが表に返っていない")
}

func TestCostlyColoursInteractor_ResetWithConfig(t *testing.T) {
	ci, _ := newCostlyReal()
	// **Parlett の 121 も選べる。**
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CostlyColoursConfig{
		CpuDifficulty: domain.CostlyColoursCpuDifficultyEasy,
		TargetScore:   121,
	}))
	assert.Equal(t, 121, ci.GetConfig().TargetScore)

	out := ci.ResetWithConfig(domain.CostlyColoursConfig{TargetScore: 999})
	assert.Contains(t, out, "err:")
	assert.Equal(t, 121, ci.GetConfig().TargetScore, "弾いた設定が入ってしまっている")
}

// **交換を決めたら数え上げへ進み、人間の手番で止まる。**
func TestCostlyColoursInteractor_MogAdvancesToPlay(t *testing.T) {
	for _, accept := range []bool{true, false} {
		ci, g := newCostlyReal()
		require.Equal(t, "ok", ci.Reset())
		require.Equal(t, "ok", ci.Mog(accept))
		assert.Equal(t, domain.CostlyColoursPhasePlay, g.GetPhase(), "accept=%v", accept)
		assert.Equal(t, domain.CostlyColoursHandSize, g.GetPlayer(0).GetCardsSize(),
			"交換で手札の枚数が変わっている (accept=%v)", accept)
	}
}

// **交換フェーズでは札を出せない。**
func TestCostlyColoursInteractor_RejectsAPlayBeforeTheMog(t *testing.T) {
	ci, g := newCostlyReal()
	require.Equal(t, "ok", ci.Reset())
	assert.Contains(t, ci.Play(0), "err:")
	assert.Equal(t, domain.CostlyColoursHandSize, g.GetPlayer(0).GetCardsSize())
}

// **打ったら手番が人間に戻ってくる。**
func TestCostlyColoursInteractor_PlayRunsTheCpuAndComesBack(t *testing.T) {
	ci, g := newCostlyReal()
	require.Equal(t, "ok", ci.Reset())
	require.Equal(t, "ok", ci.Mog(false))
	hand := g.GetPlayer(0).GetCardsSize()

	h := g.GetHint()
	require.GreaterOrEqual(t, h.HandIdx, 0)
	require.Equal(t, "ok", ci.Play(h.HandIdx))
	assert.Equal(t, hand-1, g.GetPlayer(0).GetCardsSize(), "手札が減っていない")
}

// **数え上げが始まったら交換はもう決められない。** インタラクターが
// ドメインの拒否を握り潰すと、遅れて届いた交換要求が黙って通ったように見える。
func TestCostlyColoursInteractor_RejectsAMogAfterThePlayHasStarted(t *testing.T) {
	ci, g := newCostlyReal()
	require.Equal(t, "ok", ci.Reset())
	require.Equal(t, "ok", ci.Mog(false))
	require.Equal(t, domain.CostlyColoursPhasePlay, g.GetPhase())

	hand := g.GetPlayer(0).GetCardsSize()
	assert.Contains(t, ci.Mog(true), "err:", "数え上げ中の交換が通っている")
	assert.Equal(t, hand, g.GetPlayer(0).GetCardsSize(), "弾いた交換で手札が動いた")
	assert.Equal(t, domain.CostlyColoursPhasePlay, g.GetPhase())
}

// **弾いた手は盤面を動かさない。**
func TestCostlyColoursInteractor_RejectsAnIllegalPlay(t *testing.T) {
	ci, g := newCostlyReal()
	require.Equal(t, "ok", ci.Reset())
	require.Equal(t, "ok", ci.Mog(false))
	hand := g.GetPlayer(0).GetCardsSize()
	assert.Contains(t, ci.Play(99), "err:")
	assert.Contains(t, ci.Play(-1), "err:")
	assert.Equal(t, hand, g.GetPlayer(0).GetCardsSize(), "弾いた手で札が減った")
}

// **ショーでは勝手に進まない。** 集計を読む時間を人間に渡す。
func TestCostlyColoursInteractor_StopsAtTheShow(t *testing.T) {
	ci, g := newCostlyReal()
	require.Equal(t, "ok", ci.Reset())
	costlyFinishDeal(t, ci, g)
	if g.GetGameEndFlag() {
		return
	}
	require.Equal(t, domain.CostlyColoursPhaseShow, g.GetPhase())
	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Len(t, res.Lines, 3)

	require.Equal(t, "ok", ci.NextDeal())
	assert.Equal(t, domain.CostlyColoursPhaseMog, g.GetPhase(), "次のディールは交換から始まる")
	assert.Equal(t, 2, g.GetDealNumber())
}

// **目標点に届けば終わる。**
func TestCostlyColoursInteractor_ReachesTheTargetScore(t *testing.T) {
	ci, g := newCostlyReal()
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CostlyColoursConfig{
		CpuDifficulty: domain.CostlyColoursCpuDifficultyEasy,
		TargetScore:   domain.CostlyColoursMinTarget,
	}))
	for deal := 0; deal < 200 && !g.GetGameEndFlag(); deal++ {
		costlyFinishDeal(t, ci, g)
		require.Equal(t, "ok", ci.NextDeal())
	}
	require.True(t, g.GetGameEndFlag(), "31 点勝負でも終局に届かない")
	assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
	// 終局後の操作は盤面を触らない。
	// **終局後は盤面を触らない。** guardGameEnd は素直に現状を返す。
	deal, phase := g.GetDealNumber(), g.GetPhase()
	assert.Equal(t, "ok", ci.NextDeal())
	assert.Equal(t, "ok", ci.Play(0))
	assert.Equal(t, "ok", ci.Mog(true))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, deal, g.GetDealNumber(), "終局後にディールが進んだ")
	assert.Equal(t, phase, g.GetPhase(), "終局後にフェーズが動いた")
}

func TestCostlyColoursInteractor_HintAndActionLog(t *testing.T) {
	ci, _ := newCostlyReal()
	require.Equal(t, "ok", ci.Reset())
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

// **保存した盤で打ち続けられる。**
func TestCostlyColoursInteractor_SnapshotRestoreKeepsPlaying(t *testing.T) {
	ci, g := newCostlyReal()
	require.Equal(t, "ok", ci.ResetWithConfig(domain.CostlyColoursConfig{
		CpuDifficulty: domain.CostlyColoursCpuDifficultyEasy,
		TargetScore:   domain.CostlyColoursMinTarget,
	}))
	require.Equal(t, "ok", ci.Mog(true))

	data, err := ci.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored, err := usecase.RestoreCostlyColoursInteractor(data, costlyPassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetTotal(), rg.GetTotal())
	require.NotNil(t, rg.GetTurnUp(), "表の 1 枚が消えている")
	assert.Equal(t, g.GetTurnUp().GetValue(), rg.GetTurnUp().GetValue())
	assert.Equal(t, domain.CostlyColoursMinTarget, rg.GetConfig().TargetScore, "設定が消えている")
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, g.GetPlayer(i).GetCardsSize(), rg.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
		assert.Equal(t, g.GetPlayer(i).GetScore(), rg.GetPlayer(i).GetScore(), "席 %d の得点", i)
	}

	for deal := 0; deal < 200 && !rg.GetGameEndFlag(); deal++ {
		costlyFinishDeal(t, restored, rg)
		require.Equal(t, "ok", restored.NextDeal())
	}
	assert.True(t, rg.GetGameEndFlag(), "復元した盤で終局に届かない")
}

func TestRestoreCostlyColoursInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreCostlyColoursInteractor([]byte("{"), costlyPassThrough{})
	assert.Error(t, err)
}

// costlyFinishDeal は現在のディールをショーまで打つ。
func costlyFinishDeal(t *testing.T, ci *usecase.CostlyColoursInteractor, g interfaces.CostlyColoursGame) {
	t.Helper()
	for step := 0; step < 200; step++ {
		if g.GetGameEndFlag() || g.GetPhase() == domain.CostlyColoursPhaseShow {
			return
		}
		if g.GetPhase() == domain.CostlyColoursPhaseMog {
			require.Equal(t, "ok", ci.Mog(g.GetHint().AcceptMog))
			continue
		}
		h := g.GetHint()
		require.GreaterOrEqual(t, h.HandIdx, 0, "手番なのに出せる札が無い")
		require.Equal(t, "ok", ci.Play(h.HandIdx))
	}
	require.Contains(t, []string{domain.CostlyColoursPhaseShow, domain.CostlyColoursPhaseGameEnd},
		g.GetPhase(), "ディールが終わらない")
}
