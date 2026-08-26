//go:build test

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

func ramschNoColor(t *testing.T) {
	t.Helper()
	orig := color.NoColor()
	color.SetNoColor(true)
	t.Cleanup(func() { color.SetNoColor(orig) })
}

func newRamschForPresenter() *domain.Ramsch {
	g := domain.NewDefaultRamsch()
	g.Reset()
	return g
}

// **切り札と「点は罰点」を毎回出す。** どちらもスカート系のつもりで来た人が
// 真っ先に取り違えるところで、書いていなければ盤が読めない。
func TestRamschCuiPresenter_Output_AlwaysStatesTheFixedTrumpAndInvertedScoring(t *testing.T) {
	ramschNoColor(t)
	p := new(presenter.RamschCuiPresenter)

	out := p.Output(newRamschForPresenter(), nil)
	assert.Contains(t, out, i18n.T("ramsch.trumpFixed"))
	assert.Contains(t, out, i18n.T("ramsch.scoringNote"))
	// 反対言語が漏れていないこと（T() は未翻訳ならキーを返すので、キー名が
	// そのまま出ていないことも見る）。
	assert.NotContains(t, out, "ramsch.trumpFixed")
	assert.NotContains(t, out, "ramsch.scoringNote")
}

// **ラウンド終了で、誰がいくつ取ったかを全員分出す。** 罰点なので「最多が誰か」
// が結果そのものだが、1 点差のこともある。合計だけでは読めない。
func TestRamschCuiPresenter_Output_RoundEndNamesTheLoserAndEveryonesPoints(t *testing.T) {
	ramschNoColor(t)
	p := new(presenter.RamschCuiPresenter)

	g := newRamschForPresenter()
	g.SetPhase(domain.RamschPhaseRoundEnd)
	g.SetCardPointsForTest([domain.RamschPlayerCnt]int{30, 78, 12})
	g.ScoreRound()

	out := p.Output(g, nil)
	// 3 席すべての点が、名前とともに 1 行ずつ出ること。
	for i, want := range []struct {
		name   string
		points string
	}{{"あなた", "30"}, {"CPU 1", "78"}, {"CPU 2", "12"}} {
		assert.Contains(t, out,
			i18n.Tf("ramsch.roundPointsLine", "name", want.name, "points", want.points),
			"player %d の点の行が無い", i)
	}
	// 最多を取った席が名指しされること。
	assert.Contains(t, out, i18n.Tf("ramsch.roundLoser", "name", "CPU 1", "points", "78"))
	assert.NotContains(t, out, i18n.T("ramsch.roundTied"))
}

// **Durchmarsch は別の文言で出す。** 「最多を取った人が負け」の行をそのまま
// 出すと、総取りした人が負けたように読める。
func TestRamschCuiPresenter_Output_AnnouncesDurchmarsch(t *testing.T) {
	ramschNoColor(t)
	p := new(presenter.RamschCuiPresenter)

	g := newRamschForPresenter()
	g.SetPhase(domain.RamschPhaseRoundEnd)
	g.SetCardPointsForTest([domain.RamschPlayerCnt]int{domain.RamschTotalCardPoints, 0, 0})
	g.SetDurchmarschForTest(0)
	g.ScoreRound()

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.Tf("ramsch.durchmarsch", "name", "あなた"))
	// **負のコントロール**: 敗者行は出ていないこと。
	assert.NotContains(t, out, i18n.Tf("ramsch.roundLoser",
		"name", "あなた", "points", "120"))
}

// 同点なら「全員が負う」と書く。誰か 1 人を名指ししない。
func TestRamschCuiPresenter_Output_SaysEveryoneTiedLoses(t *testing.T) {
	ramschNoColor(t)
	p := new(presenter.RamschCuiPresenter)

	g := newRamschForPresenter()
	g.SetPhase(domain.RamschPhaseRoundEnd)
	g.SetCardPointsForTest([domain.RamschPlayerCnt]int{50, 50, 20})
	g.ScoreRound()

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("ramsch.roundTied"))
}

// ヒントは着手を指し、理由を添える。
func TestRamschCuiPresenter_HintOutput(t *testing.T) {
	ramschNoColor(t)
	p := new(presenter.RamschCuiPresenter)

	g := newRamschForPresenter()
	require.True(t, g.IsHumanTurn())
	out := p.HintOutput(g)
	assert.NotEqual(t, i18n.T("ramsch.hintNone")+"\n", out, "手番なのに助言していない")
	assert.True(t, strings.Contains(out, i18n.T("ramsch.hintReasonAvoidPoints")) ||
		strings.Contains(out, i18n.T("ramsch.hintReasonLeadLow")) ||
		strings.Contains(out, i18n.T("ramsch.hintReasonForcedDiscard")),
		"理由が出ていない: %s", out)

	// 手番でなければ助言しない。
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, i18n.T("ramsch.hintNone")+"\n", p.HintOutput(g))
}

// エラーは画面に出る（黙って握り潰さない）。
func TestRamschCuiPresenter_Output_ShowsTheError(t *testing.T) {
	ramschNoColor(t)
	p := new(presenter.RamschCuiPresenter)
	out := p.Output(newRamschForPresenter(), domain.ErrWrongPhase)
	assert.Contains(t, out, domain.ErrWrongPhase.Error())
}

// アクションログは配りの記録から始まる。
func TestRamschCuiPresenter_ActionLogOutput(t *testing.T) {
	ramschNoColor(t)
	p := new(presenter.RamschCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(newRamschForPresenter()))
}
