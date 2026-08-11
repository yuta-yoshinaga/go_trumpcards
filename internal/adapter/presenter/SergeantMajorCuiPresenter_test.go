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

func newSergeantMajorForCui(t *testing.T) *domain.SergeantMajor {
	t.Helper()
	s := domain.NewDefaultSergeantMajor()
	s.Reset()
	return s
}

func TestSergeantMajorCuiPresenterOutput(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	out := p.Output(newSergeantMajorForCui(t), nil)

	assert.Contains(t, out, i18n.T("sergeantmajor.helpTitle"))
	assert.Contains(t, out, fixedPart("sergeantmajor.header"))
	// **ノルマは席順で決まる。** 規則そのものを毎回書く。
	assert.Contains(t, out, i18n.T("sergeantmajor.rule"))
	// 「ノルマ」はルール行にも出るので、席行の並び（ノルマN 獲得N）で数える。
	assert.Len(t, regexp.MustCompile(`ノルマ\d+ 獲得\d+`).FindAllString(out, -1),
		domain.SergeantMajorPlayerCnt, "全員の席行にノルマと獲得数が出る")
}

// 切り札は未宣言と確定の両側を踏む。
func TestSergeantMajorCuiPresenterTrumpLine(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)

	undeclared := newSergeantMajorForCui(t)
	out := p.Output(undeclared, nil)
	assert.Contains(t, out, fixedPart("sergeantmajor.trumpUndecided"))
	assert.Contains(t, out, strconv.Itoa(domain.SergeantMajorKittySize), "キティの枚数を出す")

	declared := newSergeantMajorForCui(t)
	declared.SetDealerIdxForTest(0)
	require.NoError(t, declared.DeclareTrump(domain.CardDesignHeart))
	out = p.Output(declared, nil)
	assert.Contains(t, out, fixedPart("sergeantmajor.trumpLine"))
	assert.Contains(t, out, i18n.T("sergeantmajor.suitHeart"))
	assert.NotContains(t, out, fixedPart("sergeantmajor.trumpUndecided"))
}

// **4 スートすべてに名前がある。**
func TestSergeantMajorCuiPresenterNamesEverySuit(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	for suit, key := range map[int]string{
		domain.CardDesignSpade:   "sergeantmajor.suitSpade",
		domain.CardDesignClover:  "sergeantmajor.suitClover",
		domain.CardDesignHeart:   "sergeantmajor.suitHeart",
		domain.CardDesignDiamond: "sergeantmajor.suitDiamond",
	} {
		s := newSergeantMajorForCui(t)
		s.SetTrumpSuitForTest(suit)
		assert.Contains(t, p.Output(s, nil), i18n.T(key))
	}
	assert.Equal(t, "?", sergeantMajorSuitName(0))
}

// **親の席には印が付く。** 切り札を決め、ノルマ 8 を負うのがその席。
func TestSergeantMajorCuiPresenterMarksTheDealer(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	s := newSergeantMajorForCui(t)
	s.SetDealerIdxForTest(1)

	out := p.Output(s, nil)
	assert.Contains(t, out, i18n.T("sergeantmajor.roleDealer"))
	assert.Len(t, regexp.MustCompile(regexp.QuoteMeta(i18n.T("sergeantmajor.roleDealer"))).FindAllString(out, -1),
		1, "印が付くのは 1 席だけ")
}

// **宣言・捨て札は親かどうかで案内が違う。** 両側を踏む。
func TestSergeantMajorCuiPresenterPrompts(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)

	mine := newSergeantMajorForCui(t)
	mine.SetDealerIdxForTest(0)
	assert.Contains(t, p.Output(mine, nil), i18n.T("sergeantmajor.promptTrump"))

	theirs := newSergeantMajorForCui(t)
	theirs.SetDealerIdxForTest(1)
	assert.Contains(t, p.Output(theirs, nil), i18n.T("sergeantmajor.promptTrumpWait"))

	myDiscard := newSergeantMajorForCui(t)
	myDiscard.SetDealerIdxForTest(0)
	require.NoError(t, myDiscard.DeclareTrump(domain.CardDesignSpade))
	assert.Contains(t, p.Output(myDiscard, nil), fixedPart("sergeantmajor.promptDiscard"))

	theirDiscard := newSergeantMajorForCui(t)
	theirDiscard.SetDealerIdxForTest(1)
	require.NoError(t, theirDiscard.DeclareTrump(domain.CardDesignSpade))
	assert.Contains(t, p.Output(theirDiscard, nil), i18n.T("sergeantmajor.promptDiscardWait"))
}

// **前ラウンドの札のやり取りは盤面に痕跡が残らない。**
func TestSergeantMajorCuiPresenterReportsTheExchange(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	fresh := newSergeantMajorForCui(t)
	assert.NotContains(t, p.Output(fresh, nil), fixedPart("sergeantmajor.exchangeLine"))

	s := newSergeantMajorForCui(t)
	s.SetPhaseForTest(domain.SergeantMajorPhasePlay)
	s.SetSurplusForTest([]int{1, -1, 0})
	s.ExchangeForTest()
	assert.Contains(t, p.Output(s, nil), fixedPart("sergeantmajor.exchangeLine"))
}

func TestSergeantMajorCuiPresenterRoundEndPrompt(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	s := newSergeantMajorForCui(t)
	s.SetPhaseForTest(domain.SergeantMajorPhaseRoundEnd)
	assert.Contains(t, p.Output(s, nil), i18n.T("sergeantmajor.promptNext"))
}

func TestSergeantMajorCuiPresenterGameEndBanners(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	for _, tc := range []struct {
		name   string
		scores [domain.SergeantMajorPlayerCnt]int
		key    string
	}{
		{"you", [3]int{4, -2, -2}, "sergeantmajor.gameEndYou"},
		{"cpu", [3]int{-2, 4, -2}, "sergeantmajor.gameEndCpu"},
		{"tie", [3]int{0, 0, 0}, "sergeantmajor.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSergeantMajorForCui(t)
			for i, n := range tc.scores {
				s.GetPlayer(i).SetScore(n)
			}
			s.FinishGameForTest()
			assert.Contains(t, p.Output(s, nil), fixedPart(tc.key))
		})
	}
}

func TestSergeantMajorCuiPresenterShowsErrors(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	s := newSergeantMajorForCui(t)
	err := s.DeclareTrump(99)
	require.Error(t, err)
	assert.Contains(t, p.Output(s, err), err.Error())
}

// **助言は宣言／捨て札／プレイで形が違う。** 3 通りすべてを踏む。
func TestSergeantMajorCuiPresenterHint(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	s := newSergeantMajorForCui(t)
	s.SetDealerIdxForTest(0)

	assert.Contains(t, p.HintOutput(s), fixedPart("sergeantmajor.hintTrump"))

	require.NoError(t, s.DeclareTrump(domain.CardDesignHeart))
	discardHint := p.HintOutput(s)
	assert.Contains(t, discardHint, fixedPart("sergeantmajor.hintDiscard"))

	require.NoError(t, s.DiscardForTest(0, []int{0, 1, 2, 3}))
	s.SetCurrentPlayerIdxForTest(0)
	cardHint := p.HintOutput(s)
	assert.Contains(t, cardHint, "HINT")
	for id := range sergeantMajorHintReasonKeys {
		assert.NotContains(t, cardHint, id, "識別子がそのまま漏れていない")
	}

	s.FinishGameForTest()
	assert.Contains(t, p.HintOutput(s), i18n.T("sergeantmajor.hintNone"))
}

// **理由の識別子はすべて訳文を持つ。**
func TestSergeantMajorCuiPresenterHintReasonsAllTranslate(t *testing.T) {
	assert.NotEmpty(t, sergeantMajorHintReasonKeys)
	for id, key := range sergeantMajorHintReasonKeys {
		assert.NotEqual(t, key, i18n.T(key), "訳が無い: "+id)
	}
}

func TestSergeantMajorCuiPresenterActionLog(t *testing.T) {
	p := new(SergeantMajorCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(newSergeantMajorForCui(t)))
}
