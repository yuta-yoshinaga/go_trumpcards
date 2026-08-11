//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newEstimationForCui(t *testing.T) *domain.Estimation {
	t.Helper()
	e := domain.NewDefaultEstimation()
	e.Reset()
	return e
}

func TestEstimationCuiPresenterOutput(t *testing.T) {
	p := new(EstimationCuiPresenter)
	e := newEstimationForCui(t)

	out := p.Output(e, nil)
	assert.Contains(t, out, i18n.T("estimation.helpTitle"))
	// **得点表は盤面から読めない。** Dash と Risk の振れ幅を常時出す。
	assert.Contains(t, out, i18n.T("estimation.scoreTable"))
	assert.Contains(t, out, fixedPart("estimation.header"))
}

// 切り札は未定と確定の両側を踏む。
func TestEstimationCuiPresenterTrumpLine(t *testing.T) {
	p := new(EstimationCuiPresenter)

	undecided := newEstimationForCui(t)
	assert.Contains(t, p.Output(undecided, nil), i18n.T("estimation.trumpUndecided"))

	decided := newEstimationForCui(t)
	decided.SetDealerIdxForTest(0)
	require.NoError(t, decided.SelectTrump(domain.CardDesignHeart))
	out := p.Output(decided, nil)
	assert.Contains(t, out, fixedPart("estimation.trumpLine"))
	assert.NotContains(t, out, i18n.T("estimation.trumpUndecided"))
}

// **親のときは選ばせ、そうでなければ待たせる。** 両側を踏む。
func TestEstimationCuiPresenterTrumpPrompts(t *testing.T) {
	p := new(EstimationCuiPresenter)

	dealer := newEstimationForCui(t)
	dealer.SetDealerIdxForTest(0)
	assert.Contains(t, p.Output(dealer, nil), i18n.T("estimation.promptTrump"))

	other := newEstimationForCui(t)
	other.SetDealerIdxForTest(2)
	out := p.Output(other, nil)
	assert.Contains(t, out, i18n.T("estimation.promptTrumpWait"))
	assert.NotContains(t, out, i18n.T("estimation.promptTrump"))
}

// **最後の宣言者には禁止値を先に伝える。** 押せない宣言を出させない。
func TestEstimationCuiPresenterAnnouncesTheRestrictedBid(t *testing.T) {
	p := new(EstimationCuiPresenter)

	e := newEstimationForCui(t)
	e.SetDealerIdxForTest(1)
	e.CpuSelectTrump()
	e.SetBidsForTest(map[int]int{1: 4, 2: 4, 3: 4})
	e.SetBidPlayerIdxForTest(0)
	assert.Contains(t, p.Output(e, nil), fixedPart("estimation.promptBidRestricted"))

	// 負のコントロール: まだ最後でなければ出さない。
	early := newEstimationForCui(t)
	early.SetDealerIdxForTest(1)
	early.CpuSelectTrump()
	early.SetBidPlayerIdxForTest(0)
	assert.NotContains(t, p.Output(early, nil), fixedPart("estimation.promptBidRestricted"))
}

// 宣言の種類ごとに表示が変わる。3 種すべて踏む。
func TestEstimationCuiPresenterBidLabels(t *testing.T) {
	p := new(EstimationCuiPresenter)
	e := newEstimationForCui(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(domain.CardDesignSpade))

	assert.Contains(t, p.Output(e, nil), i18n.T("estimation.bidNone"))

	e.SetBidsForTest(map[int]int{0: 0, 1: 5})
	e.GetPlayer(1).SetCallType(domain.EstimationCallRisk)
	out := p.Output(e, nil)
	assert.Contains(t, out, i18n.T("estimation.bidDash"))
	assert.Contains(t, out, fixedPart("estimation.bidRisk"))

	e.SetBidsForTest(map[int]int{2: 3})
	assert.Contains(t, p.Output(e, nil), fixedPart("estimation.bidNormal"))
}

func TestEstimationCuiPresenterRoundEnd(t *testing.T) {
	p := new(EstimationCuiPresenter)
	e := newEstimationForCui(t)
	e.SetPhaseForTest(domain.EstimationPhaseRoundEnd)

	out := p.Output(e, nil)
	assert.Contains(t, out, i18n.T("estimation.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("estimation.promptNext"))
	assert.NotContains(t, out, i18n.T("estimation.promptPlay"))
}

func TestEstimationCuiPresenterPlayPrompt(t *testing.T) {
	p := new(EstimationCuiPresenter)
	e := newEstimationForCui(t)
	e.SetPhaseForTest(domain.EstimationPhasePlay)
	e.SetCurrentPlayerIdxForTest(0)

	out := p.Output(e, nil)
	assert.Contains(t, out, i18n.T("estimation.promptPlay"))
	assert.Contains(t, out, fixedPart("estimation.promptCurrentPlayer"))
}

func TestEstimationCuiPresenterError(t *testing.T) {
	p := new(EstimationCuiPresenter)
	assert.Contains(t, p.Output(newEstimationForCui(t), assert.AnError), assert.AnError.Error())
}

func TestEstimationCuiPresenterGameEnd(t *testing.T) {
	p := new(EstimationCuiPresenter)

	for _, tc := range []struct {
		name    string
		scores  [4]int
		wantKey string
	}{
		{"you win", [4]int{50, 10, 10, 10}, "estimation.gameEndWin"},
		{"a cpu wins", [4]int{10, 50, 10, 10}, "estimation.gameEndLose"},
		{"a tie", [4]int{20, 20, 20, 20}, "estimation.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEstimationForCui(t)
			for i, s := range tc.scores {
				e.GetPlayer(i).SetTotalScore(s)
			}
			e.FinishGameForTest()

			out := p.Output(e, nil)
			assert.Contains(t, out, fixedPart(tc.wantKey))
			assert.NotContains(t, out, i18n.T("estimation.promptPlay"))
		})
	}
}

// 切り札選択中のヒントはスート名まで出す。
func TestEstimationCuiPresenterHintNamesTheTrumpSuit(t *testing.T) {
	p := new(EstimationCuiPresenter)
	e := newEstimationForCui(t)
	e.SetDealerIdxForTest(0)

	out := p.HintOutput(e)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "estimationSelectTrump", "生のキーが出ていたら未登録")
}

// 宣言中のヒントは数まで出す。
func TestEstimationCuiPresenterHintDuringBidding(t *testing.T) {
	p := new(EstimationCuiPresenter)
	e := newEstimationForCui(t)
	e.SetDealerIdxForTest(0)
	require.NoError(t, e.SelectTrump(domain.CardDesignSpade))

	out := p.HintOutput(e)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "estimationBid")
	assert.NotContains(t, out, "estimationDashCall")
}

// プレイ中の 2 つの理由キーが両方とも文言に解決される。
func TestEstimationCuiPresenterHintDuringPlay(t *testing.T) {
	p := new(EstimationCuiPresenter)

	short := newEstimationForCui(t)
	short.SetPhaseForTest(domain.EstimationPhasePlay)
	short.SetCurrentPlayerIdxForTest(0)
	short.GetPlayer(0).SetBid(5)
	assert.Contains(t, p.HintOutput(short), i18n.T("estimation.hintReasonWinTrick"))

	done := newEstimationForCui(t)
	done.SetPhaseForTest(domain.EstimationPhasePlay)
	done.SetCurrentPlayerIdxForTest(0)
	done.GetPlayer(0).SetBid(0)
	assert.Contains(t, p.HintOutput(done), i18n.T("estimation.hintReasonDuck"))
}

func TestEstimationCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(EstimationCuiPresenter)
	e := newEstimationForCui(t)
	e.GiveUp()
	assert.Contains(t, p.HintOutput(e), i18n.T("estimation.hintNone"))
}

// スート名は 4 つとも解決される。
func TestEstimationSuitName(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		assert.NotEqual(t, "?", estimationSuitName(suit))
	}
	assert.Equal(t, "?", estimationSuitName(domain.CardDesignJoker))
}

func TestEstimationCuiPresenterActionLogOutput(t *testing.T) {
	p := new(EstimationCuiPresenter)
	e := newEstimationForCui(t)
	e.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(e))
}
