//go:build test

package presenter

import (
	"math/rand"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newPigForCui(t *testing.T) *domain.Pig {
	t.Helper()
	g := domain.NewDefaultPig()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

func TestPigCuiPresenterOutput(t *testing.T) {
	p := new(PigCuiPresenter)
	g := newPigForCui(t)
	out := p.Output(g, nil)

	assert.Contains(t, out, i18n.T("pig.helpTitle"))
	assert.Contains(t, out, fixedPart("pig.header"))
	// **罰の理由が直感と違う。** 取り合いではなく、遅れることが負け。
	assert.Contains(t, out, i18n.T("pig.rule"))
	// 「手札」はルール行にも出るので、席行の並びで数える。
	assert.Len(t, regexp.MustCompile(`手札\d+枚 文字\[`).FindAllString(out, -1),
		domain.PigDefaultPlayerCnt, "全員の席行に手札と文字が出る")
	assert.Contains(t, out, strconv.Itoa(domain.PigDeckSize(domain.PigDefaultPlayerCnt)))
	assert.Contains(t, out, i18n.T("pig.promptPass"))
}

// **選び終えた席に印を出す。** 同時に渡すので、待ちが発生します。
func TestPigCuiPresenterMarksWhoHasChosen(t *testing.T) {
	p := new(PigCuiPresenter)
	g := newPigForCui(t)
	assert.NotContains(t, p.Output(g, nil), i18n.T("pig.roleChosen"))

	require.NoError(t, g.ChoosePassForTest(0, 0))
	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("pig.roleChosen"))
	assert.Contains(t, out, i18n.T("pig.promptWaiting"))
	assert.NotContains(t, out, i18n.T("pig.promptPass"), "もう選ぶ場面ではない")
}

// **合図は声に出さない。** こちらから名乗る必要があることを出す。
func TestPigCuiPresenterPromptsForTheSignal(t *testing.T) {
	p := new(PigCuiPresenter)
	g := newPigForCui(t)
	g.OpenSignalForTest(2)

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("pig.promptSignal"))
	assert.Contains(t, out, i18n.T("pig.promptSignalCmd"))
	assert.Contains(t, out, fixedPart("pig.roleNoticed"))
	assert.NotContains(t, out, i18n.T("pig.promptPass"))

	// 名乗ったあとは待ちの表示に変わる。
	require.NoError(t, g.PlayerSignal())
	out = p.Output(g, nil)
	assert.Contains(t, out, fixedPart("pig.promptSignalDone"))
	assert.NotContains(t, out, i18n.T("pig.promptSignal"))
}

// **ラウンドの罰は 1 回きりの出来事。** 配り直す前に読ませる。
func TestPigCuiPresenterShowsTheRoundResult(t *testing.T) {
	p := new(PigCuiPresenter)
	g := newPigForCui(t)
	g.OpenSignalForTest(1)
	g.NoticeForTest(2)
	g.NoticeForTest(3)
	require.Equal(t, domain.PigPhaseRoundEnd, g.GetPhase())

	out := p.Output(g, nil)
	assert.Contains(t, out, fixedPart("pig.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("pig.promptNext"))
	assert.Contains(t, out, "P", "受け取った文字を出す")
	// **3 文字で脱落**という目標が席行から読める (#5766)。まだ 0 文字の席でも
	// 出るので、行そのものを組み立てて突き合わせる ("PIG" は溜まった文字の
	// 側にも現れるため、部分一致では素通りする)。
	line := i18n.Tf("pig.playerLine",
		"name", cuiPlayerName(g.GetPlayer(0), 0), "role", "",
		"cards", strconv.Itoa(g.GetPlayer(0).GetCardsSize()),
		"letters", g.GetPlayer(0).GetLetterWord(), "target", domain.PigLetterTargetWord)
	assert.Contains(t, out, line)
}

// **人間が脱落しても局は続く。**
func TestPigCuiPresenterTellsTheHumanTheyAreOut(t *testing.T) {
	p := new(PigCuiPresenter)
	g := newPigForCui(t)
	assert.Contains(t, p.Output(g, nil), i18n.T("pig.promptPass"))

	g.GetPlayer(0).SetEliminated(true)
	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("pig.promptEliminated"))
	assert.Contains(t, out, i18n.T("pig.roleOut"))
	assert.NotContains(t, out, i18n.T("pig.promptPass"))
}

func TestPigCuiPresenterGameEndBanners(t *testing.T) {
	p := new(PigCuiPresenter)

	won := newPigForCui(t)
	for i := 1; i < won.GetPlayerCnt()-1; i++ {
		won.GetPlayer(i).SetLetters(domain.PigMaxLetters)
		won.GetPlayer(i).SetEliminated(true)
	}
	won.GetPlayer(won.GetPlayerCnt() - 1).SetLetters(domain.PigMaxLetters - 1)
	won.OpenSignalForTest(0)
	require.True(t, won.GetGameEndFlag())
	out := p.Output(won, nil)
	assert.Contains(t, out, i18n.T("pig.gameEndYou"))
	assert.NotContains(t, out, i18n.T("pig.promptPass"), "終局後は促さない")

	lost := newPigForCui(t)
	lost.GiveUp()
	assert.Contains(t, p.Output(lost, nil), fixedPart("pig.gameEndCpu"))
}

func TestPigCuiPresenterShowsErrors(t *testing.T) {
	p := new(PigCuiPresenter)
	assert.Contains(t, p.Output(newPigForCui(t), assert.AnError), assert.AnError.Error())
}

func TestPigCuiPresenterHintOutput(t *testing.T) {
	p := new(PigCuiPresenter)
	g := newPigForCui(t)

	out := p.HintOutput(g)
	assert.Contains(t, out, fixedPart("pig.hintCard"))

	// **合図の場面では札を指さない。**
	g.OpenSignalForTest(1)
	out = p.HintOutput(g)
	assert.Contains(t, out, fixedPart("pig.hintSignal"))
	assert.Contains(t, out, i18n.T("pig.hintReasonSignal"))
	assert.NotContains(t, out, fixedPart("pig.hintCard"))

	g.GiveUp()
	assert.Contains(t, p.HintOutput(g), i18n.T("pig.hintNone"))
}

func TestPigCuiPresenterActionLogOutput(t *testing.T) {
	p := new(PigCuiPresenter)
	g := newPigForCui(t)
	g.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
