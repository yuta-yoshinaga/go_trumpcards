//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newHokmForCui(t *testing.T) *domain.Hokm {
	t.Helper()
	h := domain.NewDefaultHokm()
	h.Reset()
	return h
}

func TestHokmCuiPresenterOutput(t *testing.T) {
	p := new(HokmCuiPresenter)
	h := newHokmForCui(t)

	out := p.Output(h, nil)
	assert.Contains(t, out, i18n.T("hokm.helpTitle"))
	// **7 トリック先取が肝。** 13 まで打たないので進捗はここに出る。
	assert.Contains(t, out, fixedPart("hokm.raceLine"))
	assert.Contains(t, out, fixedPart("hokm.scoreLine"))
}

// 切り札は未宣言と確定の両側を踏む。
func TestHokmCuiPresenterTrumpLine(t *testing.T) {
	p := new(HokmCuiPresenter)

	undeclared := newHokmForCui(t)
	assert.Contains(t, p.Output(undeclared, nil), i18n.T("hokm.trumpUndecided"))

	declared := newHokmForCui(t)
	declared.SetHakemIdxForTest(0)
	require.NoError(t, declared.PlayerDeclareTrump(domain.CardDesignHeart))
	out := p.Output(declared, nil)
	assert.Contains(t, out, fixedPart("hokm.trumpLine"))
	assert.NotContains(t, out, i18n.T("hokm.trumpUndecided"))
}

// **親のときは宣言させ、そうでなければ待たせる。** 両側を踏む。
func TestHokmCuiPresenterTrumpPrompts(t *testing.T) {
	p := new(HokmCuiPresenter)

	hakem := newHokmForCui(t)
	hakem.SetHakemIdxForTest(0)
	assert.Contains(t, p.Output(hakem, nil), i18n.T("hokm.promptTrump"))

	other := newHokmForCui(t)
	other.SetHakemIdxForTest(2)
	out := p.Output(other, nil)
	assert.Contains(t, out, i18n.T("hokm.promptTrumpWait"))
	assert.NotContains(t, out, i18n.T("hokm.promptTrump"))
}

// 親の席に印が付く。
func TestHokmCuiPresenterMarksTheHakem(t *testing.T) {
	p := new(HokmCuiPresenter)
	h := newHokmForCui(t)
	h.SetHakemIdxForTest(1)

	assert.Contains(t, p.Output(h, nil), i18n.T("hokm.roleHakem"))
}

// **Kot は2点。** 何が起きたか言わないと得点が飛んで見える。両側を踏む。
func TestHokmCuiPresenterExplainsHowTheHandEnded(t *testing.T) {
	p := new(HokmCuiPresenter)

	kot := newHokmForCui(t)
	kot.GiveTricksForTest(0, domain.HokmTricksToWin)
	kot.FinishHandForTest(0)
	require.Equal(t, domain.HokmPhaseHandEnd, kot.GetPhase())
	outKot := p.Output(kot, nil)
	assert.Contains(t, outKot, i18n.T("hokm.promptHandEndKot"))
	assert.NotContains(t, outKot, i18n.T("hokm.promptHandEnd"))

	normal := newHokmForCui(t)
	normal.GiveTricksForTest(0, domain.HokmTricksToWin)
	normal.GiveTricksForTest(1, 1)
	normal.FinishHandForTest(0)
	outNormal := p.Output(normal, nil)
	assert.Contains(t, outNormal, i18n.T("hokm.promptHandEnd"))
	assert.NotContains(t, outNormal, i18n.T("hokm.promptHandEndKot"))
}

func TestHokmCuiPresenterPlayPrompt(t *testing.T) {
	p := new(HokmCuiPresenter)
	h := newHokmForCui(t)
	h.SetPhaseForTest(domain.HokmPhasePlay)
	h.SetCurrentPlayerIdxForTest(0)

	out := p.Output(h, nil)
	assert.Contains(t, out, i18n.T("hokm.promptPlay"))
	assert.Contains(t, out, fixedPart("hokm.promptCurrentPlayer"))
}

func TestHokmCuiPresenterError(t *testing.T) {
	p := new(HokmCuiPresenter)
	assert.Contains(t, p.Output(newHokmForCui(t), assert.AnError), assert.AnError.Error())
}

func TestHokmCuiPresenterGameEnd(t *testing.T) {
	p := new(HokmCuiPresenter)

	for _, tc := range []struct {
		name    string
		t0, t1  int
		wantKey string
	}{
		{"your team", 7, 3, "hokm.gameEndTeam0"},
		{"the other team", 3, 7, "hokm.gameEndTeam1"},
		{"a tie", 5, 5, "hokm.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHokmForCui(t)
			h.SetScoreForTestUse(0, tc.t0)
			h.SetScoreForTestUse(1, tc.t1)
			h.FinishGameForTest()

			out := p.Output(h, nil)
			assert.Contains(t, out, fixedPart(tc.wantKey))
			assert.NotContains(t, out, i18n.T("hokm.promptPlay"))
		})
	}
}

// 切り札のヒントはスート名まで出す。
func TestHokmCuiPresenterHintNamesTheSuit(t *testing.T) {
	p := new(HokmCuiPresenter)
	h := newHokmForCui(t)
	h.SetHakemIdxForTest(0)

	out := p.HintOutput(h)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "hokmDeclareTrump", "生のキーが出ていたら未登録")
}

// プレイ中の 2 つの理由キーが両方とも文言に解決される。
func TestHokmCuiPresenterHintDuringPlay(t *testing.T) {
	p := new(HokmCuiPresenter)

	solo := newHokmForCui(t)
	solo.SetHakemIdxForTest(0)
	require.NoError(t, solo.PlayerDeclareTrump(domain.CardDesignHeart))
	solo.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(solo), i18n.T("hokm.hintReasonWinTrick"))

	partner := newHokmForCui(t)
	partner.SetHakemIdxForTest(0)
	require.NoError(t, partner.PlayerDeclareTrump(domain.CardDesignHeart))
	partner.SetCurrentPlayerIdxForTest(0)
	// 味方 (2番) が切り札の A で勝っている。
	partner.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 8, false)},
	})
	assert.Contains(t, p.HintOutput(partner), i18n.T("hokm.hintReasonSaveCards"))
}

func TestHokmCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(HokmCuiPresenter)
	h := newHokmForCui(t)
	h.GiveUp()
	assert.Contains(t, p.HintOutput(h), i18n.T("hokm.hintNone"))
}

// スート名は 4 つとも解決される。
func TestHokmSuitName(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		assert.NotEqual(t, "?", hokmSuitName(suit))
	}
	assert.Equal(t, "?", hokmSuitName(domain.CardDesignJoker))
}

func TestHokmCuiPresenterActionLogOutput(t *testing.T) {
	p := new(HokmCuiPresenter)
	h := newHokmForCui(t)
	h.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(h))
}

// **親は負けたときだけ交代する** (#5753)。次に自分が切り札を選べるかを
// 左右するのに、次ハンドが始まって親バッジが動くまで分からなかった。
func TestHokmCuiPresenterAnnouncesTheHakemChange(t *testing.T) {
	p := new(HokmCuiPresenter)

	moved := newHokmForCui(t)
	moved.FinishHandForTest(1 - domain.HokmTeamOf(moved.GetHakemIdx()))
	moved.SetPhaseForTest(domain.HokmPhaseHandEnd)
	movedOut := p.Output(moved, nil)
	assert.Contains(t, movedOut, i18n.T("hokm.hakemMoves"))
	assert.NotContains(t, movedOut, i18n.T("hokm.hakemStays"))

	stayed := newHokmForCui(t)
	stayed.FinishHandForTest(domain.HokmTeamOf(stayed.GetHakemIdx()))
	stayed.SetPhaseForTest(domain.HokmPhaseHandEnd)
	stayedOut := p.Output(stayed, nil)
	assert.Contains(t, stayedOut, i18n.T("hokm.hakemStays"))
	assert.NotContains(t, stayedOut, i18n.T("hokm.hakemMoves"))
}
