//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMendikotForCui(t *testing.T) *domain.Mendikot {
	t.Helper()
	m := domain.NewDefaultMendikot()
	m.Reset()
	return m
}

func TestMendikotCuiPresenterOutput(t *testing.T) {
	p := new(MendikotCuiPresenter)
	out := p.Output(newMendikotForCui(t), nil)

	assert.Contains(t, out, i18n.T("mendikot.helpTitle"))
	// **10 の枚数が勝敗そのもの。** 盤面から読めないので常時出す。
	assert.Contains(t, out, fixedPart("mendikot.tensLine"))
	assert.Contains(t, out, fixedPart("mendikot.trickLine"))
	assert.Contains(t, out, fixedPart("mendikot.scoreLine"))
	assert.Contains(t, out, i18n.T("mendikot.rule"), "勝敗条件そのものを説明する")
}

// 切り札は未決定と確定の両側を踏む。
func TestMendikotCuiPresenterTrumpLine(t *testing.T) {
	p := new(MendikotCuiPresenter)

	undecided := newMendikotForCui(t)
	assert.Contains(t, p.Output(undecided, nil), i18n.T("mendikot.trumpUndecided"))

	decided := newMendikotForCui(t)
	decided.SetTrumpForTest(domain.CardDesignHeart, 2)
	out := p.Output(decided, nil)
	assert.Contains(t, out, fixedPart("mendikot.trumpLine"))
	assert.Contains(t, out, i18n.T("mendikot.suitHeart"))
	assert.NotContains(t, out, i18n.T("mendikot.trumpUndecided"))
}

// **4 スートすべてに名前がある。** 既定の "?" に落ちない。
func TestMendikotCuiPresenterNamesEverySuit(t *testing.T) {
	p := new(MendikotCuiPresenter)
	for suit, key := range map[int]string{
		domain.CardDesignSpade:   "mendikot.suitSpade",
		domain.CardDesignClover:  "mendikot.suitClover",
		domain.CardDesignHeart:   "mendikot.suitHeart",
		domain.CardDesignDiamond: "mendikot.suitDiamond",
	} {
		m := newMendikotForCui(t)
		m.SetTrumpForTest(suit, 1)
		out := p.Output(m, nil)
		assert.Contains(t, out, i18n.T(key))
		assert.NotContains(t, out, "?", "未知スートの既定値に落ちていない")
	}
	assert.Equal(t, "?", mendikotSuitName(0))
}

// **人間の手札だけ番号付きで見える。**
func TestMendikotCuiPresenterShowsOnlyTheHumanHand(t *testing.T) {
	p := new(MendikotCuiPresenter)
	m := newMendikotForCui(t)
	out := p.Output(m, nil)

	assert.Equal(t, domain.MendikotPlayerCnt, strings.Count(out, "[T"), "全員の席行は出る")
	assert.Contains(t, out, "[0]", "人間の手札は番号付き")
}

// **決まり方で勝ち点が 1/2/3 と変わる。** 4 通りすべて別の文言になる。
func TestMendikotCuiPresenterHandEndKinds(t *testing.T) {
	p := new(MendikotCuiPresenter)
	seen := map[string]bool{}
	for _, kind := range []string{"tens", "tricks", "mendikot", "whitewash"} {
		m := newMendikotForCui(t)
		m.SetPhaseForTest(domain.MendikotPhaseHandEnd)
		line := i18n.Tf("mendikot.handEnd."+kind, "team", "0")
		require.NotContains(t, line, "mendikot.handEnd.", "翻訳キーが存在する: "+kind)
		assert.False(t, seen[line], "決まり方ごとに別の文言: "+kind)
		seen[line] = true
	}
	assert.Len(t, seen, 4)

	m := newMendikotForCui(t)
	m.GetPlayer(0).SetTens(domain.MendikotTensInDeck)
	m.SetTrickNumberForTest(domain.MendikotTricksPerRound - 1)
	m.FinishHandForTest()
	out := p.Output(m, nil)
	assert.Contains(t, out, i18n.Tf("mendikot.handEnd.mendikot", "team", "0"))
	assert.Contains(t, out, i18n.T("mendikot.promptNext"))
}

func TestMendikotCuiPresenterGameEndBanners(t *testing.T) {
	p := new(MendikotCuiPresenter)
	for _, tc := range []struct {
		name   string
		t0, t1 int
		key    string
	}{
		{"team0", 7, 2, "mendikot.gameEndTeam0"},
		{"team1", 2, 7, "mendikot.gameEndTeam1"},
		{"tie", 7, 7, "mendikot.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newMendikotForCui(t)
			m.SetScoreForTestUse(0, tc.t0)
			m.SetScoreForTestUse(1, tc.t1)
			m.FinishGameForTest()
			assert.Contains(t, p.Output(m, nil), fixedPart(tc.key))
		})
	}
}

func TestMendikotCuiPresenterShowsErrors(t *testing.T) {
	p := new(MendikotCuiPresenter)
	m := newMendikotForCui(t)
	err := m.PlayerPlay(999)
	require.Error(t, err)
	assert.Contains(t, p.Output(m, err), err.Error())
}

func TestMendikotCuiPresenterHint(t *testing.T) {
	p := new(MendikotCuiPresenter)
	m := newMendikotForCui(t)
	m.SetCurrentPlayerIdxForTest(0)

	out := p.HintOutput(m)
	assert.Contains(t, out, "HINT")
	// **理由は生の識別子ではなく訳文で出す。**
	for id, key := range mendikotHintReasonKeys {
		assert.NotContains(t, out, id, "識別子がそのまま漏れていない")
		_ = key
	}

	m.SetPhaseForTest(domain.MendikotPhaseHandEnd)
	assert.Contains(t, p.HintOutput(m), i18n.T("mendikot.hintNone"))
}

// **理由の識別子はすべて訳文を持つ。**
func TestMendikotCuiPresenterHintReasonsAllTranslate(t *testing.T) {
	assert.NotEmpty(t, mendikotHintReasonKeys)
	for id, key := range mendikotHintReasonKeys {
		assert.NotEqual(t, key, i18n.T(key), "訳が無い: "+id)
	}
}

func TestMendikotCuiPresenterActionLog(t *testing.T) {
	p := new(MendikotCuiPresenter)
	m := newMendikotForCui(t)
	assert.NotEmpty(t, p.ActionLogOutput(m))
}
