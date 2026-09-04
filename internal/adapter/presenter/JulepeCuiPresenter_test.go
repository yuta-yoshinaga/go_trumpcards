//go:build test

package presenter

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newJulepeForCui(t *testing.T) *domain.Julepe {
	t.Helper()
	r := domain.NewDefaultJulepe()
	r.Reset()
	return r
}

func TestJulepeCuiPresenterOutput(t *testing.T) {
	p := new(JulepeCuiPresenter)
	r := newJulepeForCui(t)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("julepe.helpTitle"))
	// **ポット・切り札・リスクの3点が参加判断の材料。** 常時出す。
	assert.Contains(t, out, fixedPart("julepe.potLine"))
	assert.Contains(t, out, fixedPart("julepe.trumpLine"))
	assert.Contains(t, out, fixedPart("julepe.riskLine"))
	// 配り直後は参加可否を促す。
	assert.Contains(t, out, i18n.T("julepe.promptDecide"))
	assert.NotContains(t, out, i18n.T("julepe.promptPlay"))
}

// 参加状況の表示が3通り出る（未定 / 参加 / 降り）。
func TestJulepeCuiPresenterStatusStrings(t *testing.T) {
	p := new(JulepeCuiPresenter)

	undecided := newJulepeForCui(t)
	assert.Contains(t, p.Output(undecided, nil), i18n.T("julepe.statusUndecided"))

	decided := newJulepeForCui(t)
	require.NoError(t, decided.Decide(true))
	out := p.Output(decided, nil)
	assert.Contains(t, out, i18n.T("julepe.statusIn"), "参加した人が出る")

	dropped := newJulepeForCui(t)
	require.NoError(t, dropped.Decide(false))
	assert.Contains(t, p.Output(dropped, nil), i18n.T("julepe.statusOut"), "降りた人が出る")
}

// **降りたラウンドは操作を促さない。** 待ちに見えてはいけない。
func TestJulepeCuiPresenterWatchingWhenDroppedOut(t *testing.T) {
	p := new(JulepeCuiPresenter)
	r := newJulepeForCui(t)
	r.GetPlayer(0).SetDecided(true)
	r.GetPlayer(0).SetInRound(false)
	r.SetPhaseForTest(domain.JulepePhasePlay)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("julepe.promptWatching"))
	assert.NotContains(t, out, i18n.T("julepe.promptPlay"))
}

func TestJulepeCuiPresenterPromptsPlayWhenIn(t *testing.T) {
	p := new(JulepeCuiPresenter)
	r := newJulepeForCui(t)
	r.GetPlayer(0).SetDecided(true)
	r.GetPlayer(0).SetInRound(true)
	r.SetPhaseForTest(domain.JulepePhasePlay)
	r.SetCurrentPlayerIdxForTest(0)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("julepe.promptPlay"))
	assert.NotContains(t, out, i18n.T("julepe.promptWatching"))
}

func TestJulepeCuiPresenterRoundEnd(t *testing.T) {
	p := new(JulepeCuiPresenter)
	r := newJulepeForCui(t)
	r.SetPhaseForTest(domain.JulepePhaseRoundEnd)

	out := p.Output(r, nil)
	assert.Contains(t, out, i18n.T("julepe.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("julepe.promptNext"))
}

func TestJulepeCuiPresenterError(t *testing.T) {
	p := new(JulepeCuiPresenter)
	assert.Contains(t, p.Output(newJulepeForCui(t), assert.AnError), assert.AnError.Error())
}

func TestJulepeCuiPresenterGameEnd(t *testing.T) {
	p := new(JulepeCuiPresenter)

	t.Run("a winner", func(t *testing.T) {
		r := newJulepeForCui(t)
		r.GetPlayer(1).SetChips(200)
		r.FinishGameForTest()
		out := p.Output(r, nil)
		assert.Contains(t, out, fixedPart("julepe.gameEndWinner"))
		assert.NotContains(t, out, i18n.T("julepe.promptDecide"))
	})

	t.Run("a tie", func(t *testing.T) {
		r := newJulepeForCui(t)
		r.FinishGameForTest()
		assert.Contains(t, p.Output(r, nil), i18n.T("julepe.gameEndTie"))
	})
}

// **選択フェーズのヒントは札ではなく参加可否。** 書式そのものが違う。
func TestJulepeCuiPresenterHintInDecidePhase(t *testing.T) {
	p := new(JulepeCuiPresenter)
	r := newJulepeForCui(t)

	out := p.HintOutput(r)
	assert.Contains(t, out, "HINT")
	// 参加/降りのどちらかの文言が出て、生のキーは出ない。
	assert.True(t,
		strings.Contains(out, i18n.T("julepe.hintReasonPlayIn")) ||
			strings.Contains(out, i18n.T("julepe.hintReasonPassOut")),
		"参加可否の理由が出る: %s", out)
	assert.NotContains(t, out, "julepePlayIn")
	assert.NotContains(t, out, "julepePassOut")
}

// プレイ中のヒントは札を指す。2 つの理由キーの両方を踏む。
func TestJulepeCuiPresenterHintInPlayPhase(t *testing.T) {
	p := new(JulepeCuiPresenter)

	fresh := newJulepeForCui(t)
	require.NoError(t, fresh.Decide(true))
	fresh.SetPhaseForTest(domain.JulepePhasePlay)
	fresh.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(fresh), i18n.T("julepe.hintReasonTakeTrick"))

	// **「安全」の線は規定トリック数であって 1 トリックではない。**
	// 既定の卓では規定が 2 なので、1 トリックではまだ安全にならない。
	safe := newJulepeForCui(t)
	require.NoError(t, safe.Decide(true))
	safe.SetPhaseForTest(domain.JulepePhasePlay)
	safe.SetCurrentPlayerIdxForTest(0)
	safe.GetPlayer(0).SetRoundTricks(safe.GetRequiredTricks())
	assert.Contains(t, p.HintOutput(safe), i18n.T("julepe.hintReasonAlreadySafe"))
}

func TestJulepeCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(JulepeCuiPresenter)
	r := newJulepeForCui(t)
	r.GiveUp()
	assert.Contains(t, p.HintOutput(r), i18n.T("julepe.hintNone"))
}

func TestJulepeCuiPresenterActionLogOutput(t *testing.T) {
	p := new(JulepeCuiPresenter)
	r := newJulepeForCui(t)
	r.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(r))
}

// **参加判断もリードも親の左隣から始まる** (#5748)。誰が親かが出ていないと、
// 自分が何番目に決断するのかが読めない。
func TestJulepeCuiPresenterMarksTheDealer(t *testing.T) {
	p := new(JulepeCuiPresenter)
	r := newJulepeForCui(t)
	r.SetDealerIdxForTest(2)

	out := julepePlain(p.Output(r, nil))

	// 印が付くのは 1 席だけ。行ごとに突き合わせる。
	assert.Contains(t, out, julepePlain(cuiPlayerName(r.GetPlayer(2), 2))+i18n.T("julepe.dealerMark"))
	assert.Equal(t, 1, strings.Count(out, i18n.T("julepe.dealerMark")))

	// 親が移れば印も移る。
	moved := newJulepeForCui(t)
	moved.SetDealerIdxForTest(0)
	movedOut := julepePlain(p.Output(moved, nil))
	assert.Contains(t, movedOut, julepePlain(cuiPlayerName(moved.GetPlayer(0), 0))+i18n.T("julepe.dealerMark"))
	assert.NotContains(t, movedOut, julepePlain(cuiPlayerName(moved.GetPlayer(2), 2))+i18n.T("julepe.dealerMark"))
}

// #6616: Web は validPlays で出せる札をリング表示しているのに、CUI は素の一覧
// だけで、番号を打ってエラーを踏むまで分からなかった。
func TestJulepeCuiPresenterMarksThePlayableCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(JulepeCuiPresenter)

	setup := func(phase domain.JulepePhase, turn int, currentTrick []*domain.TrickCard) *domain.Julepe {
		players := []*domain.JulepePlayer{
			domain.NewJulepePlayer(true),
			domain.NewJulepePlayer(false),
			domain.NewJulepePlayer(false),
			domain.NewJulepePlayer(false),
		}
		cfg := domain.DefaultJulepeConfig()
		r := domain.NewJulepe(domain.NewTrumpCards(0), players, cfg)
		human := r.GetPlayer(0)
		human.SetDecided(true)
		human.SetInRound(true)
		human.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		human.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		human.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		human.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))

		r.SetTrumpSuitForTest(domain.CardDesignClover)
		r.SetPhaseForTest(phase)
		r.SetCurrentPlayerIdxForTest(turn)
		r.SetCurrentTrickForTest(currentTrick)
		return r
	}

	t.Run("marks only cards allowed by the follow rule", func(t *testing.T) {
		trick := []*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
		}
		r := setup(domain.JulepePhasePlay, 0, trick)
		out := p.Output(r, nil)

		// 出せる札（SPADE 7, SPADE 1）には印が付く
		assert.Contains(t, out, "[0]SPADE 7"+CuiLegalMark)
		assert.Contains(t, out, "[1]SPADE 1"+CuiLegalMark)
		// 出せない札（HEART 8, DIAMOND 10）には印が付かない
		assert.NotContains(t, out, "HEART 8"+CuiLegalMark)
		assert.NotContains(t, out, "DIAMOND 10"+CuiLegalMark)
		// 印の総数が合法手の数（2枚）と一致する
		assert.Equal(t, 2, strings.Count(out, CuiLegalMark))
	})

	t.Run("marks nothing on another player's turn", func(t *testing.T) {
		trick := []*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
		}
		r := setup(domain.JulepePhasePlay, 1, trick)
		out := p.Output(r, nil)
		assert.NotContains(t, out, CuiLegalMark, "他家の手番では目印を出さない")
	})

	t.Run("marks nothing outside the play phase", func(t *testing.T) {
		r := setup(domain.JulepePhaseDecide, 0, nil)
		out := p.Output(r, nil)
		assert.NotContains(t, out, CuiLegalMark, "選択フェーズでは目印を出さない")
	})
}

// julepePlain は色付けのエスケープを落とす。
var julepeAnsi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func julepePlain(s string) string { return julepeAnsi.ReplaceAllString(s, "") }
