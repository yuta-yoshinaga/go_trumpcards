package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestBourreCuiPresenter_Output(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)

	// Decide phase render
	out := p.Output(bg, nil)
	assert.Contains(t, out, "==========")

	// Drive to game end, rendering along the way
	for i := 0; i < 300000 && !bg.GetGameEndFlag(); i++ {
		if bg.GetPhase() == domain.BourrePhaseRoundEnd {
			bg.NextHand()
			continue
		}
		_ = p.Output(bg, nil)
		bg.CpuPlay()
	}
	endOut := p.Output(bg, nil)
	assert.Contains(t, endOut, "==========")
	assert.NotEmpty(t, strings.TrimSpace(endOut))
}

func TestBourreCuiPresenter_Error(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)
	out := p.Output(bg, assertErr{})
	assert.Contains(t, out, "boom")
}

type assertErr struct{}

func (assertErr) Error() string { return "boom" }

func TestBourreCuiPresenter_ActionLog(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(bg))
}

// **CPU 名も i18n と太字を通す。**bourreName は人間だけ i18n を通し、CPU は
// `fmt.Sprintf("CPU %d")` の英語リテラルを返していた (#4719)。
func TestBourreCuiPresenter_CpuNamesGoThroughTheSharedHelper(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)
	// **リセット直後は bourreName を通らない。**トリックが場に出るまで進める。
	var out string
	for i := 0; i < 300000; i++ {
		out = p.Output(bg, nil)
		if strings.Contains(out, i18n.T("cuiPlayerCpu")) || len(bg.GetCurrentTrick()) > 0 {
			break
		}
		if bg.GetPhase() == domain.BourrePhaseRoundEnd {
			bg.NextHand()
			continue
		}
		if bg.GetGameEndFlag() {
			break
		}
		bg.CpuPlay()
	}
	require.NotEmpty(t, bg.GetCurrentTrick(), "no trick was ever rendered")

	// **プレイヤー行は元から cuiPlayerName を使っている。**出力全体で探すと
	// 修正前でも当たるので、素の "CPU 1" が 1 つも残っていないことで見る。
	bold1 := color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))
	assert.Contains(t, out, bold1)
	assert.NotContains(t, strings.ReplaceAll(out, bold1, ""), "CPU 1")
	assert.NotContains(t, strings.ReplaceAll(out, color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "2")), ""), "CPU 2")
}
