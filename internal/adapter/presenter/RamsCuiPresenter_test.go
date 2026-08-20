//go:build test

package presenter

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newRamsForCui(t *testing.T) *domain.Rams {
	t.Helper()
	r := domain.NewDefaultRams()
	r.Reset()
	return r
}

func TestRamsCuiPresenterOutput(t *testing.T) {
	p := new(RamsCuiPresenter)
	r := newRamsForCui(t)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("rams.helpTitle"))
	// **ポット・切り札・リスクの3点が参加判断の材料。** 常時出す。
	assert.Contains(t, out, fixedPart("rams.potLine"))
	assert.Contains(t, out, fixedPart("rams.trumpLine"))
	assert.Contains(t, out, fixedPart("rams.riskLine"))
	// 配り直後は参加可否を促す。
	assert.Contains(t, out, i18n.T("rams.promptDecide"))
	assert.NotContains(t, out, i18n.T("rams.promptPlay"))
}

// 参加状況の表示が3通り出る（未定 / 参加 / 降り）。
func TestRamsCuiPresenterStatusStrings(t *testing.T) {
	p := new(RamsCuiPresenter)

	undecided := newRamsForCui(t)
	assert.Contains(t, p.Output(undecided, nil), i18n.T("rams.statusUndecided"))

	decided := newRamsForCui(t)
	require.NoError(t, decided.Decide(true))
	out := p.Output(decided, nil)
	assert.Contains(t, out, i18n.T("rams.statusIn"), "参加した人が出る")

	dropped := newRamsForCui(t)
	require.NoError(t, dropped.Decide(false))
	assert.Contains(t, p.Output(dropped, nil), i18n.T("rams.statusOut"), "降りた人が出る")
}

// **降りたラウンドは操作を促さない。** 待ちに見えてはいけない。
func TestRamsCuiPresenterWatchingWhenDroppedOut(t *testing.T) {
	p := new(RamsCuiPresenter)
	r := newRamsForCui(t)
	r.GetPlayer(0).SetDecided(true)
	r.GetPlayer(0).SetInRound(false)
	r.SetPhaseForTest(domain.RamsPhasePlay)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("rams.promptWatching"))
	assert.NotContains(t, out, i18n.T("rams.promptPlay"))
}

func TestRamsCuiPresenterPromptsPlayWhenIn(t *testing.T) {
	p := new(RamsCuiPresenter)
	r := newRamsForCui(t)
	r.GetPlayer(0).SetDecided(true)
	r.GetPlayer(0).SetInRound(true)
	r.SetPhaseForTest(domain.RamsPhasePlay)
	r.SetCurrentPlayerIdxForTest(0)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("rams.promptPlay"))
	assert.NotContains(t, out, i18n.T("rams.promptWatching"))
}

func TestRamsCuiPresenterRoundEnd(t *testing.T) {
	p := new(RamsCuiPresenter)
	r := newRamsForCui(t)
	r.SetPhaseForTest(domain.RamsPhaseRoundEnd)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("rams.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("rams.promptNext"))
}

func TestRamsCuiPresenterError(t *testing.T) {
	p := new(RamsCuiPresenter)
	assert.Contains(t, p.Output(newRamsForCui(t), assert.AnError), assert.AnError.Error())
}

func TestRamsCuiPresenterGameEnd(t *testing.T) {
	p := new(RamsCuiPresenter)

	t.Run("a winner", func(t *testing.T) {
		r := newRamsForCui(t)
		r.GetPlayer(1).SetChips(200)
		r.FinishGameForTest()
		out := p.Output(r, nil)
		assert.Contains(t, out, fixedPart("rams.gameEndWinner"))
		assert.NotContains(t, out, i18n.T("rams.promptDecide"))
	})

	t.Run("a tie", func(t *testing.T) {
		r := newRamsForCui(t)
		r.FinishGameForTest()
		assert.Contains(t, p.Output(r, nil), i18n.T("rams.gameEndTie"))
	})
}

// **選択フェーズのヒントは札ではなく参加可否。** 書式そのものが違う。
func TestRamsCuiPresenterHintInDecidePhase(t *testing.T) {
	p := new(RamsCuiPresenter)
	r := newRamsForCui(t)

	out := p.HintOutput(r)
	assert.Contains(t, out, "HINT")
	// 参加/降りのどちらかの文言が出て、生のキーは出ない。
	assert.True(t,
		strings.Contains(out, i18n.T("rams.hintReasonPlayIn")) ||
			strings.Contains(out, i18n.T("rams.hintReasonPassOut")),
		"参加可否の理由が出る: %s", out)
	assert.NotContains(t, out, "ramsPlayIn")
	assert.NotContains(t, out, "ramsPassOut")
}

// プレイ中のヒントは札を指す。2 つの理由キーの両方を踏む。
func TestRamsCuiPresenterHintInPlayPhase(t *testing.T) {
	p := new(RamsCuiPresenter)

	fresh := newRamsForCui(t)
	require.NoError(t, fresh.Decide(true))
	fresh.SetPhaseForTest(domain.RamsPhasePlay)
	fresh.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(fresh), i18n.T("rams.hintReasonTakeTrick"))

	safe := newRamsForCui(t)
	require.NoError(t, safe.Decide(true))
	safe.SetPhaseForTest(domain.RamsPhasePlay)
	safe.SetCurrentPlayerIdxForTest(0)
	safe.GetPlayer(0).SetRoundTricks(1)
	assert.Contains(t, p.HintOutput(safe), i18n.T("rams.hintReasonAlreadySafe"))
}

func TestRamsCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(RamsCuiPresenter)
	r := newRamsForCui(t)
	r.GiveUp()
	assert.Contains(t, p.HintOutput(r), i18n.T("rams.hintNone"))
}

func TestRamsCuiPresenterActionLogOutput(t *testing.T) {
	p := new(RamsCuiPresenter)
	r := newRamsForCui(t)
	r.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(r))
}

// **参加判断もリードも親の左隣から始まる** (#5748)。誰が親かが出ていないと、
// 自分が何番目に決断するのかが読めない。
func TestRamsCuiPresenterMarksTheDealer(t *testing.T) {
	p := new(RamsCuiPresenter)
	r := newRamsForCui(t)
	r.SetDealerIdxForTest(2)

	out := ramsPlain(p.Output(r, nil))

	// 印が付くのは 1 席だけ。行ごとに突き合わせる。
	assert.Contains(t, out, ramsPlain(cuiPlayerName(r.GetPlayer(2), 2))+i18n.T("rams.dealerMark"))
	assert.Equal(t, 1, strings.Count(out, i18n.T("rams.dealerMark")))

	// 親が移れば印も移る。
	moved := newRamsForCui(t)
	moved.SetDealerIdxForTest(0)
	movedOut := ramsPlain(p.Output(moved, nil))
	assert.Contains(t, movedOut, ramsPlain(cuiPlayerName(moved.GetPlayer(0), 0))+i18n.T("rams.dealerMark"))
	assert.NotContains(t, movedOut, ramsPlain(cuiPlayerName(moved.GetPlayer(2), 2))+i18n.T("rams.dealerMark"))
}

// ramsPlain は色付けのエスケープを落とす。
var ramsAnsi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func ramsPlain(s string) string { return ramsAnsi.ReplaceAllString(s, "") }
