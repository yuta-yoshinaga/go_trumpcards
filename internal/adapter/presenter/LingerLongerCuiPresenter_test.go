//go:build test

package presenter

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newLingerLongerForCui(t *testing.T) *domain.LingerLonger {
	t.Helper()
	l := domain.NewDefaultLingerLonger()
	l.Reset()
	return l
}

func TestLingerLongerCuiPresenterOutput(t *testing.T) {
	p := new(LingerLongerCuiPresenter)
	l := newLingerLongerForCui(t)
	out := p.Output(l, nil)

	assert.Contains(t, out, i18n.T("lingerlonger.helpTitle"))
	assert.Contains(t, out, fixedPart("lingerlonger.header"))
	// **トリックを取っても得点にならない。** 規則が直感と逆なので毎回書く。
	assert.Contains(t, out, i18n.T("lingerlonger.rule"))
	// 「手札」はルール行にも出るので、席行の並び（手札N枚 獲得N回）で数える。
	assert.Len(t, regexp.MustCompile(`手札\d+枚 獲得\d+回`).FindAllString(out, -1),
		domain.DefaultLingerLongerConfig().PlayerCnt, "全員の席行に手札と獲得回数が出る")
	// 山札の残りは常に見えている。
	assert.Contains(t, out, strconv.Itoa(l.GetStockSize()))
}

// **補充した席と脱落した席を印で出す。** どちらも盤面に痕跡が残らない。
func TestLingerLongerCuiPresenterMarksDrawsAndEliminations(t *testing.T) {
	p := new(LingerLongerCuiPresenter)
	l := newLingerLongerForCui(t)
	assert.NotContains(t, p.Output(l, nil), i18n.T("lingerlonger.roleDrew"))

	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	for i := range l.GetPlayerCnt() {
		l.GiveHandForTest(i, domain.NewCard(domain.CardDesignSpade, 13-i, false),
			domain.NewCard(domain.CardDesignHeart, 2, false))
	}
	for i := range l.GetPlayerCnt() {
		require.NoError(t, l.PlayForTest(i, 0))
	}
	assert.Contains(t, p.Output(l, nil), i18n.T("lingerlonger.roleDrew"))

	// 脱落した席には順番が付く。
	out := newLingerLongerForCui(t)
	out.GetPlayer(2).SetEliminatedAt(1)
	assert.Contains(t, p.Output(out, nil), fixedPart("lingerlonger.roleOut"))
}

// **山札が尽きたら誰も補充できない。** 黙っていると盤面から読み取れない。
func TestLingerLongerCuiPresenterAnnouncesTheEmptyStock(t *testing.T) {
	p := new(LingerLongerCuiPresenter)
	l := newLingerLongerForCui(t)
	assert.NotContains(t, p.Output(l, nil), i18n.T("lingerlonger.noStockLine"))

	l.DrainStockForTest()
	assert.Contains(t, p.Output(l, nil), i18n.T("lingerlonger.noStockLine"))
}

// **人間が脱落しても局は続く。** 促しを出したままだと操作できるように見える。
func TestLingerLongerCuiPresenterTellsTheHumanTheyAreOut(t *testing.T) {
	p := new(LingerLongerCuiPresenter)
	l := newLingerLongerForCui(t)
	assert.Contains(t, p.Output(l, nil), i18n.T("lingerlonger.promptPlay"))

	l.GetPlayer(0).SetEliminatedAt(1)
	l.GiveHandForTest(0)
	out := p.Output(l, nil)
	assert.Contains(t, out, i18n.T("lingerlonger.promptEliminated"))
	assert.NotContains(t, out, i18n.T("lingerlonger.promptPlay"))
}

func TestLingerLongerCuiPresenterGameEndBanners(t *testing.T) {
	p := new(LingerLongerCuiPresenter)

	won := newLingerLongerForCui(t)
	won.DrainStockForTest()
	won.SetLeadPlayerIdxForTest(0)
	won.SetCurrentPlayerIdxForTest(0)
	won.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignSpade, 12, false))
	for i := 1; i < won.GetPlayerCnt(); i++ {
		won.GiveHandForTest(i, domain.NewCard(domain.CardDesignSpade, 2+i, false))
	}
	for i := range won.GetPlayerCnt() {
		require.NoError(t, won.PlayForTest(i, 0))
	}
	require.True(t, won.GetGameEndFlag())
	out := p.Output(won, nil)
	assert.Contains(t, out, i18n.T("lingerlonger.gameEndYou"))
	assert.NotContains(t, out, i18n.T("lingerlonger.promptPlay"), "終局後は促さない")

	lost := newLingerLongerForCui(t)
	lost.GiveUp()
	assert.Contains(t, p.Output(lost, nil), fixedPart("lingerlonger.gameEndCpu"))
}

func TestLingerLongerCuiPresenterShowsErrors(t *testing.T) {
	p := new(LingerLongerCuiPresenter)
	assert.Contains(t, p.Output(newLingerLongerForCui(t), assert.AnError), assert.AnError.Error())
}

func TestLingerLongerCuiPresenterHintOutput(t *testing.T) {
	p := new(LingerLongerCuiPresenter)
	l := newLingerLongerForCui(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	l.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 14, false))

	out := p.HintOutput(l)
	assert.Contains(t, out, fixedPart("lingerlonger.hintCard"))
	assert.Contains(t, out, i18n.T("lingerlonger.hintReasonWinTrick"))

	// **山札が空なら取っても補充は無い。** 理由が変わる。
	l.DrainStockForTest()
	assert.Contains(t, p.HintOutput(l), i18n.T("lingerlonger.hintReasonNoStock"))

	// 取れない場面は安い札を勧める。
	duck := newLingerLongerForCui(t)
	duck.SetLeadPlayerIdxForTest(1)
	duck.SetCurrentPlayerIdxForTest(0)
	duck.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 3, false))
	duck.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 14, false)},
	})
	assert.Contains(t, p.HintOutput(duck), i18n.T("lingerlonger.hintReasonDuck"))

	l.GiveUp()
	assert.Contains(t, p.HintOutput(l), i18n.T("lingerlonger.hintNone"))
}

func TestLingerLongerCuiPresenterActionLogOutput(t *testing.T) {
	p := new(LingerLongerCuiPresenter)
	l := newLingerLongerForCui(t)
	l.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(l))
}

// CUI と Web が同じ勝因を主張すること。
//
// 山札が尽きて全員が同時に手札 0 枚になった局では「最後まで持ち続けた人」は
// 存在せず、勝ちは最後のトリックで決まる。以前はどちらの勝ちも同じ見出しだった
// (#5765)。
func TestLingerLongerEndBanner_FollowsTheWinReason(t *testing.T) {
	lasted := lingerLongerEndBanner(domain.LingerLongerWinLasted, 1, "CPU1")
	lastTrick := lingerLongerEndBanner(domain.LingerLongerWinLastTrick, 1, "CPU1")
	giveUp := lingerLongerEndBanner(domain.LingerLongerWinGiveUp, 1, "CPU1")

	// 3 つの勝因が 3 つとも違う文言になること。ここが同じだと、勝因を持ち回った
	// 意味が無い。
	assert.NotEqual(t, lasted, lastTrick)
	assert.NotEqual(t, lasted, giveUp)
	assert.NotEqual(t, lastTrick, giveUp)

	// 同時脱落の勝ちを「持ち続けた」と説明してはいけない。
	assert.Contains(t, lasted, "持ち続け")
	assert.NotContains(t, lastTrick, "持ち続け")
	assert.Contains(t, lastTrick, "最後のトリック")

	// 席 0 かどうかで呼び方が変わる。
	assert.NotEqual(t, lastTrick, lingerLongerEndBanner(domain.LingerLongerWinLastTrick, 0, "あなた"))

	// 未知の勝因は通常勝ちに寄せる (多数派に倒すほうが誤りが小さい)。
	assert.Equal(t, lasted, lingerLongerEndBanner("", 1, "CPU1"))

	// 英語ロケールでも同じ区別が保たれること。
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	assert.NotEqual(t,
		lingerLongerEndBanner(domain.LingerLongerWinLasted, 1, "CPU1"),
		lingerLongerEndBanner(domain.LingerLongerWinLastTrick, 1, "CPU1"))
}
