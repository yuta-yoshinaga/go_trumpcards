//go:build test

package presenter

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newPasurForCui(t *testing.T) *domain.Pasur {
	t.Helper()
	p := domain.NewDefaultPasur()
	p.Reset()
	return p
}

func TestPasurCuiPresenterOutput(t *testing.T) {
	p := new(PasurCuiPresenter)
	out := p.Output(newPasurForCui(t), nil)

	assert.Contains(t, out, i18n.T("pasur.helpTitle"))
	assert.Contains(t, out, fixedPart("pasur.header"))
	// **11 の合計と絵札の扱いが規則そのもの。** 毎回書く。
	assert.Contains(t, out, i18n.T("pasur.rule"))
	// 「捕獲」はルール行に出ないので、席行の並びで数える。
	assert.Len(t, regexp.MustCompile(`捕獲\d+枚 スール\d+`).FindAllString(out, -1),
		domain.PasurDefaultPlayerCnt, "全員の席行に捕獲数とスール数が出る")
}

// **場の札には番号を振る。** `play <i> <t...>` の t はこの番号。
func TestPasurCuiPresenterNumbersTheTable(t *testing.T) {
	p := new(PasurCuiPresenter)
	g := newPasurForCui(t)
	g.SetTableForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 4, false),
	})
	out := p.Output(g, nil)
	assert.Contains(t, out, fixedPart("pasur.tableLine"))
	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "[1]")

	// **空も出す。** 「場: なし」が出ないと、直前のスールが読めない。
	//
	// `tableLine` と `tableEmpty` は前置きを共有するので、前置きでは見分けられない
	// ——札の番号が消えたことで見る。
	g.SetTableForTest(nil)
	out = p.Output(g, nil)
	assert.Contains(t, out, i18n.T("pasur.tableEmpty"))
	tableSection := out[strings.Index(out, i18n.T("pasur.tableEmpty")):]
	assert.NotContains(t, strings.SplitN(tableSection, "\n", 2)[0], "[0]", "場の札が残っていない")
}

// **場に残った札の行き先が読めること。**
func TestPasurCuiPresenterMarksTheLastCapturer(t *testing.T) {
	p := new(PasurCuiPresenter)
	g := newPasurForCui(t)
	assert.NotContains(t, p.Output(g, nil), i18n.T("pasur.roleLastCapture"), "まだ誰も取っていない")

	g.SetLastCaptureIdxForTest(2)
	assert.Contains(t, p.Output(g, nil), i18n.T("pasur.roleLastCapture"))
}

func TestPasurCuiPresenterGameEndBanners(t *testing.T) {
	p := new(PasurCuiPresenter)

	// 単独勝利。
	win := newPasurForCui(t)
	win.SetCurrentPlayerIdxForTest(0)
	win.SetTableForTest([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	win.SetHumanHandForTest(domain.NewCard(domain.CardDesignSpade, 9, false))
	require.NoError(t, win.PlayForTest(0, 0, []int{0}))
	win.EmptyHandsForTest()
	win.DrainDeckForTest()
	win.FinishGameForTest()
	assert.Contains(t, p.Output(win, nil), i18n.T("pasur.gameEndYou"))

	// 全員 0 点なら同点。
	tie := newPasurForCui(t)
	tie.EmptyHandsForTest()
	tie.DrainDeckForTest()
	tie.SetTableForTest(nil)
	tie.FinishGameForTest()
	out := p.Output(tie, nil)
	assert.Contains(t, out, fixedPart("pasur.gameEndTie"))
	assert.NotContains(t, out, i18n.T("pasur.promptPlay"), "終局後は促さない")

	// 投了すると相手側が勝つ。
	lose := newPasurForCui(t)
	lose.GiveUp()
	assert.Contains(t, p.Output(lose, nil), fixedPart("pasur.gameEndTie"),
		"CPU 3 人が同点で残る")
}

func TestPasurCuiPresenterShowsErrors(t *testing.T) {
	p := new(PasurCuiPresenter)
	assert.Contains(t, p.Output(newPasurForCui(t), assert.AnError), assert.AnError.Error())
}

func TestPasurCuiPresenterHintOutput(t *testing.T) {
	p := new(PasurCuiPresenter)
	g := newPasurForCui(t)
	g.SetCurrentPlayerIdxForTest(0)

	// 取れるときは、取る場札の番号まで出す。
	g.SetTableForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
	})
	g.SetHumanHandForTest(domain.NewCard(domain.CardDesignDiamond, 4, false))
	out := p.HintOutput(g)
	assert.Contains(t, out, fixedPart("pasur.hintCapture"))
	assert.Contains(t, out, i18n.T("pasur.hintReasonCapture"))

	// 場を空にできるならスール。
	g.SetTableForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	g.SetHumanHandForTest(domain.NewCard(domain.CardDesignDiamond, 4, false))
	assert.Contains(t, p.HintOutput(g), i18n.T("pasur.hintReasonSoor"))

	// **取れないときは取る札を書かない。**
	g.SetTableForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)})
	g.SetHumanHandForTest(domain.NewCard(domain.CardDesignDiamond, 4, false))
	out = p.HintOutput(g)
	assert.Contains(t, out, i18n.T("pasur.hintReasonTrail"))
	// `hintCapture` と `hintTrail` は "[HINT: [" を共有するので、末尾で見分ける。
	assert.Contains(t, out, "を場に置く")
	assert.NotContains(t, out, "を取る", "取れないので取る札は書かない")

	g.SetCurrentPlayerIdxForTest(1)
	assert.Contains(t, p.HintOutput(g), i18n.T("pasur.hintNone"), "相手の手番では助言しない")

	g.SetCurrentPlayerIdxForTest(0)
	g.GiveUp()
	assert.Contains(t, p.HintOutput(g), i18n.T("pasur.hintNone"))
}

func TestPasurCuiPresenterActionLogOutput(t *testing.T) {
	p := new(PasurCuiPresenter)
	g := newPasurForCui(t)
	g.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **スールは「取った結果、場が空になる」こと** (#5762)。倍化を狙えるかを
// その場で判断できるよう、CUI は場の枚数と、ヒントが空にするかどうかを出す。
func TestPasurCuiPresenterSpellsOutTheSoorOpportunity(t *testing.T) {
	p := new(PasurCuiPresenter)
	g := newPasurForCui(t)
	g.SetCurrentPlayerIdxForTest(0)
	g.SetTableForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
	})

	// 促しには「この N 枚を取り切るとスール」の N と倍率が入る。
	assert.Contains(t, p.Output(g, nil), i18n.Tf("pasur.soorNote",
		"n", "2", "mult", strconv.Itoa(domain.PasurSoorMultiplier)))

	// **場が空なら基準そのものが無い。** 負のコントロール。
	g.SetTableForTest(nil)
	assert.NotContains(t, p.Output(g, nil), fixedPart("pasur.soorNote"))

	// ヒントが場を空にする取り方を勧めているときだけ印が付く。
	g.SetTableForTest([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
	g.SetHumanHandForTest(domain.NewCard(domain.CardDesignDiamond, 4, false))
	mark := i18n.Tf("pasur.hintSoorMark", "mult", strconv.Itoa(domain.PasurSoorMultiplier))
	assert.Contains(t, p.HintOutput(g), mark)

	// 場に 1 枚残る取り方なら印は付かない。
	g.SetTableForTest([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
	})
	g.SetHumanHandForTest(domain.NewCard(domain.CardDesignDiamond, 4, false))
	assert.NotContains(t, p.HintOutput(g), mark)
}
