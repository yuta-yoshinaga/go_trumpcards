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

func newMinibridgeForCui(t *testing.T) *domain.Minibridge {
	t.Helper()
	m := domain.NewDefaultMinibridge()
	m.Reset()
	return m
}

func TestMinibridgeCuiPresenterOutput(t *testing.T) {
	p := new(MinibridgeCuiPresenter)
	out := p.Output(newMinibridgeForCui(t), nil)

	assert.Contains(t, out, i18n.T("minibridge.helpTitle"))
	assert.Contains(t, out, fixedPart("minibridge.header"))
	// **競りが無いこと自体が規則。** 毎回書く。
	assert.Contains(t, out, i18n.T("minibridge.rule"))
	// 「HCP」はルール行にも出るので、席行の並び（HCPn 獲得n）で数える。
	assert.Len(t, regexp.MustCompile(`HCP\d+ 獲得\d+`).FindAllString(out, -1),
		domain.MinibridgePlayerCnt, "4 席すべてに HCP と獲得数が出る")
	assert.Contains(t, out, fixedPart("minibridge.scoreLine"))
}

// **HCP の合計は必ず 40。** 盤面に出ている数字を足して確かめる。
func TestMinibridgeCuiPresenterShowsFortyPointsInTotal(t *testing.T) {
	p := new(MinibridgeCuiPresenter)
	out := p.Output(newMinibridgeForCui(t), nil)
	total := 0
	for _, m := range regexp.MustCompile(`HCP(\d+) 獲得`).FindAllStringSubmatch(out, -1) {
		n, err := strconv.Atoi(m[1])
		require.NoError(t, err)
		total += n
	}
	assert.Equal(t, domain.MinibridgeTotalHcp, total)
}

// 契約は未決定と確定の両側を踏む。
func TestMinibridgeCuiPresenterContractLine(t *testing.T) {
	p := new(MinibridgeCuiPresenter)

	undecided := newMinibridgeForCui(t)
	assert.Contains(t, p.Output(undecided, nil), i18n.T("minibridge.contractUndecided"))

	decided := newMinibridgeForCui(t)
	decided.SetPhaseForTest(domain.MinibridgePhasePlay)
	decided.SetContractForTest(0, 4, domain.CardDesignHeart)
	out := p.Output(decided, nil)
	assert.Contains(t, out, fixedPart("minibridge.contractLine"))
	assert.Contains(t, out, i18n.T("minibridge.suitHeart"))
	// **「4♥」だけでは何トリック要るか分からない。**
	assert.Contains(t, out, strconv.Itoa(domain.MinibridgeBookTricks+4))
	assert.Contains(t, out, i18n.T("minibridge.roleDeclarer"))
	assert.Contains(t, out, i18n.T("minibridge.roleDummy"))
	assert.NotContains(t, out, i18n.T("minibridge.contractUndecided"))
}

// **4 スートとノートランプの 5 つすべてに名前がある。**
func TestMinibridgeCuiPresenterNamesEverySuit(t *testing.T) {
	p := new(MinibridgeCuiPresenter)
	for suit, key := range map[int]string{
		domain.CardDesignSpade:   "minibridge.suitSpade",
		domain.CardDesignClover:  "minibridge.suitClover",
		domain.CardDesignHeart:   "minibridge.suitHeart",
		domain.CardDesignDiamond: "minibridge.suitDiamond",
		0:                        "minibridge.suitNoTrump",
	} {
		m := newMinibridgeForCui(t)
		m.SetPhaseForTest(domain.MinibridgePhasePlay)
		m.SetContractForTest(0, 1, suit)
		assert.Contains(t, p.Output(m, nil), i18n.T(key), "suit=%d", suit)
	}
}

// **ダミーは契約が決まってから公開される。** 出さないと人間が操作できない。
func TestMinibridgeCuiPresenterDummyHand(t *testing.T) {
	p := new(MinibridgeCuiPresenter)

	before := newMinibridgeForCui(t)
	assert.NotContains(t, p.Output(before, nil), fixedPart("minibridge.dummyHand"))

	after := newMinibridgeForCui(t)
	require.NoError(t, after.SelectContractForTest(after.GetDeclarerIdx(), 2, 0))
	assert.Contains(t, p.Output(after, nil), fixedPart("minibridge.dummyHand"))
}

func TestMinibridgeCuiPresenterPrompts(t *testing.T) {
	p := new(MinibridgeCuiPresenter)

	mine := newMinibridgeForCui(t)
	mine.SetContractForTest(0, 0, 0)
	mine.SetPhaseForTest(domain.MinibridgePhaseContract)
	assert.Contains(t, p.Output(mine, nil), i18n.T("minibridge.promptContract"))

	theirs := newMinibridgeForCui(t)
	theirs.SetContractForTest(1, 0, 0)
	theirs.SetPhaseForTest(domain.MinibridgePhaseContract)
	assert.Contains(t, p.Output(theirs, nil), i18n.T("minibridge.promptContractWait"))

	playing := newMinibridgeForCui(t)
	playing.SetContractForTest(0, 2, 0)
	playing.SetPhaseForTest(domain.MinibridgePhasePlay)
	playing.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.Output(playing, nil), i18n.T("minibridge.promptPlay"))

	// **ダミーの手番は「ダミーの手札から出す」と言い分ける。**
	playing.SetCurrentPlayerIdxForTest(2)
	out := p.Output(playing, nil)
	assert.Contains(t, out, i18n.T("minibridge.promptPlayDummy"))

	roundEnd := newMinibridgeForCui(t)
	roundEnd.SetPhaseForTest(domain.MinibridgePhaseRoundEnd)
	assert.Contains(t, p.Output(roundEnd, nil), i18n.T("minibridge.promptNext"))
}

func TestMinibridgeCuiPresenterGameEndBanners(t *testing.T) {
	p := new(MinibridgeCuiPresenter)
	for _, tc := range []struct {
		name   string
		winner int
		key    string
	}{
		{"あなたのペア", 0, "minibridge.gameEndYou"},
		{"相手ペア", 1, "minibridge.gameEndCpu"},
		{"同点", -1, "minibridge.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newMinibridgeForCui(t)
			if tc.winner >= 0 {
				m.SetTeamScore(tc.winner, 100)
			}
			m.FinishGameForTest()
			out := p.Output(m, nil)
			assert.Contains(t, out, i18n.T(tc.key))
			assert.NotContains(t, out, i18n.T("minibridge.promptPlay"), "終局後は促さない")
		})
	}
}

func TestMinibridgeCuiPresenterShowsErrors(t *testing.T) {
	p := new(MinibridgeCuiPresenter)
	assert.Contains(t, p.Output(newMinibridgeForCui(t), assert.AnError), assert.AnError.Error())
}

func TestMinibridgeCuiPresenterHintOutput(t *testing.T) {
	p := new(MinibridgeCuiPresenter)

	// 契約の助言は札ではなくレベルと種別を指す。
	m := newMinibridgeForCui(t)
	m.SetContractForTest(0, 0, 0)
	m.SetPhaseForTest(domain.MinibridgePhaseContract)
	out := p.HintOutput(m)
	assert.Contains(t, out, fixedPart("minibridge.hintContract"))
	assert.Contains(t, out, i18n.T("minibridge.hintReasonContract"))
	assert.NotContains(t, out, fixedPart("minibridge.hintCard"))

	// 本番の助言は札を指す。
	m.SetContractForTest(0, 2, domain.CardDesignHeart)
	m.SetPhaseForTest(domain.MinibridgePhasePlay)
	m.SetCurrentPlayerIdxForTest(0)
	out = p.HintOutput(m)
	assert.Contains(t, out, fixedPart("minibridge.hintCard"))
	assert.Contains(t, out, i18n.T("minibridge.hintReasonWinTrick"))

	// **ダミーの手番なら、指しているのはダミーの手札。**
	m.SetCurrentPlayerIdxForTest(2)
	out = p.HintOutput(m)
	assert.Contains(t, out, i18n.T("minibridge.hintReasonDummy"))

	m.FinishGameForTest()
	assert.Contains(t, p.HintOutput(m), i18n.T("minibridge.hintNone"))
}

func TestMinibridgeCuiPresenterActionLogOutput(t *testing.T) {
	p := new(MinibridgeCuiPresenter)
	m := newMinibridgeForCui(t)
	m.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(m))
}
