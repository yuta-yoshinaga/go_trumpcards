package presenter_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestHachiHachiCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.HachiHachiCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand + indexed field
}

func TestHachiHachiCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.HachiHachiCuiPresenter)

	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.HachiHachiPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.HachiHachiPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	// 空フィールド表示。
	g.SetPhase(domain.HachiHachiPhasePlay)
	g.SetFieldCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestHachiHachiCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.HachiHachiCuiPresenter)

	g := domain.NewDefaultHachiHachi()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.HintOutput(g))

	// ヒント無し (CPU 手番)。
	g.SetCurrentTurn(1)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestHachiHachiCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	g.Reset()
	p := new(presenter.HachiHachiCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(g))
}

func TestHachiHachiCuiPresenter_RoundResult(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	cfg := domain.DefaultHachiHachiConfig()
	cfg.TargetRounds = 3
	g.SetConfig(cfg)
	g.Reset()
	for step := 0; step < 20000 && g.GetPhase() == domain.HachiHachiPhasePlay; step++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerPlay(0, -1))
		} else {
			g.CpuPlay()
		}
	}
	require.Equal(t, domain.HachiHachiPhaseRoundEnd, g.GetPhase())

	p := new(presenter.HachiHachiCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

// #5711: Web はラウンドを制したプレイヤーの行に 👑 を付けるのに、CUI は
// raw/bonus/delta を並べるだけで、誰が一番手だったかは数字を見比べるしかなかった。
func TestHachiHachiCuiPresenter_MarksTheRoundWinner(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true) // 精算行を素の文字列で組み立てて照合するため
	defer color.SetNoColor(origNoColor)
	p := new(presenter.HachiHachiCuiPresenter)

	atRoundEnd := func(best int) *domain.HachiHachi {
		g := domain.NewDefaultHachiHachi()
		g.Reset()
		g.SetPhase(domain.HachiHachiPhaseRoundEnd)
		g.SetLastRoundResult(&domain.HachiHachiRoundResult{
			Best: best,
			Scores: []domain.HachiHachiPlayerScore{
				{PlayerIdx: 0, RawScore: 70, Bonus: 0, Delta: -18},
				{PlayerIdx: 1, RawScore: 120, Bonus: 10, Delta: 42},
				{PlayerIdx: 2, RawScore: 74, Bonus: 0, Delta: -14},
			},
		})
		return g
	}

	// 精算行そのものを組み立てて照合する。**プレイヤー状態の行にも名前が出る**ので、
	// 名前で行を拾うと精算行ではない行を見てしまう。
	scoreLine := func(name string, raw, bonus int, delta string) string {
		return i18n.Tf("hachihachi.scoreLine",
			"name", name, "raw", strconv.Itoa(raw), "bonus", strconv.Itoa(bonus), "delta", delta)
	}

	t.Run("marks the best player and nobody else", func(t *testing.T) {
		out := p.Output(atRoundEnd(1), nil)

		marker := i18n.T("hachihachi.roundBestMark")
		assert.Contains(t, out, scoreLine("CPU 1", 120, 10, "+42")+"  "+marker)
		assert.Contains(t, out, scoreLine("CPU 2", 74, 0, "-14")+"\n")
		assert.NotContains(t, out, scoreLine("CPU 2", 74, 0, "-14")+"  "+marker)
	})

	// 総取りが決まらなかったラウンド (Best = -1) では誰にも付けない。
	t.Run("marks nobody when there is no best player", func(t *testing.T) {
		out := p.Output(atRoundEnd(-1), nil)

		assert.NotContains(t, out, i18n.T("hachihachi.roundBestMark"))
	})
}
