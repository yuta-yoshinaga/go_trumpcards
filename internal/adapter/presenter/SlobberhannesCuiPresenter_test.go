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

func newSlobberhannesForCui(t *testing.T) *domain.Slobberhannes {
	t.Helper()
	s := domain.NewDefaultSlobberhannes()
	s.Reset()
	return s
}

// 置換前の固定部分で照合する（色コードと {{}} 展開が混ざるため）。
func fixedPart(key string) string {
	return strings.SplitN(i18n.T(key), "{{", 2)[0]
}

func TestSlobberhannesCuiPresenterOutput(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)
	s := newSlobberhannesForCui(t)
	s.SetTrickNumberForTest(3)
	s.SetCurrentPlayerIdxForTest(0)

	out := p.Output(s, nil)

	assert.Contains(t, out, i18n.T("slobberhannes.helpTitle"))
	// **切り札が無いことを明示する。** 無いという情報は画面に出ないと伝わらない。
	assert.Contains(t, out, i18n.T("slobberhannes.noTrumpLine"))
	assert.Contains(t, out, i18n.T("slobberhannes.promptPlay"))
	assert.Contains(t, out, fixedPart("slobberhannes.header"))
}

// 最初と最後のトリックでだけ警告が出る。中間では出ない（負のコントロール）。
func TestSlobberhannesCuiPresenterPositionWarnings(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)

	t.Run("first trick warns", func(t *testing.T) {
		s := newSlobberhannesForCui(t)
		s.SetTrickNumberForTest(0)
		out := p.Output(s, nil)
		assert.Contains(t, out, i18n.T("slobberhannes.warnFirst"))
		assert.NotContains(t, out, i18n.T("slobberhannes.warnLast"))
	})

	t.Run("last trick warns", func(t *testing.T) {
		s := newSlobberhannesForCui(t)
		s.SetTrickNumberForTest(domain.SlobberhannesTricksPerRound - 1)
		out := p.Output(s, nil)
		assert.Contains(t, out, i18n.T("slobberhannes.warnLast"))
		assert.NotContains(t, out, i18n.T("slobberhannes.warnFirst"))
	})

	t.Run("middle trick is silent", func(t *testing.T) {
		s := newSlobberhannesForCui(t)
		s.SetTrickNumberForTest(3)
		out := p.Output(s, nil)
		assert.NotContains(t, out, i18n.T("slobberhannes.warnFirst"))
		assert.NotContains(t, out, i18n.T("slobberhannes.warnLast"))
	})
}

// 罰の内訳が記号で出る。無傷とそうでない場合の両方を踏む。
func TestSlobberhannesCuiPresenterPenaltyMarks(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)

	clean := newSlobberhannesForCui(t)
	clean.SetTrickNumberForTest(3)
	assert.Contains(t, p.Output(clean, nil), i18n.T("slobberhannes.markClean"))

	hit := newSlobberhannesForCui(t)
	hit.SetTrickNumberForTest(3)
	hit.SetPenaltiesForTest(0, true, false, true)
	out := p.Output(hit, nil)
	assert.Contains(t, out, i18n.T("slobberhannes.markFirst"))
	assert.Contains(t, out, i18n.T("slobberhannes.markQueen"))
}

func TestSlobberhannesCuiPresenterError(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)
	s := newSlobberhannesForCui(t)
	assert.Contains(t, p.Output(s, assert.AnError), assert.AnError.Error())
}

func TestSlobberhannesCuiPresenterRoundEnd(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)
	s := newSlobberhannesForCui(t)
	s.SetTrickNumberForTest(3)
	s.SetPhase(domain.SlobberhannesPhaseRoundEnd)

	out := p.Output(s, nil)
	assert.Contains(t, out, i18n.T("slobberhannes.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("slobberhannes.promptNext"))
	assert.NotContains(t, out, i18n.T("slobberhannes.promptPlay"), "ラウンド終了中は出し手を促さない")
}

func TestSlobberhannesCuiPresenterGameEnd(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)

	t.Run("a winner", func(t *testing.T) {
		s := newSlobberhannesForCui(t)
		s.GetPlayer(1).SetScore(2)
		s.Finish()
		out := p.Output(s, nil)
		assert.Contains(t, out, fixedPart("slobberhannes.gameEndWinner"))
		assert.NotContains(t, out, i18n.T("slobberhannes.promptPlay"))
	})

	t.Run("a tie", func(t *testing.T) {
		s := newSlobberhannesForCui(t)
		s.Finish()
		assert.Contains(t, p.Output(s, nil), i18n.T("slobberhannes.gameEndTie"))
	})
}

func TestSlobberhannesCuiPresenterHintOutput(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)
	s := newSlobberhannesForCui(t)
	s.SetTrickNumberForTest(3)
	s.SetCurrentPlayerIdxForTest(0)

	out := p.HintOutput(s)
	assert.Contains(t, out, "HINT")
	// 理由キーが i18n に解決されている。生のキーが出ていたら未登録。
	assert.NotContains(t, out, "slobberhannesLeadLow")
	assert.NotContains(t, out, "slobberhannesAvoid")
}

// 3 つの理由キーがすべて日本語文言になる。
func TestSlobberhannesCuiPresenterHintReasons(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)

	lead := newSlobberhannesForCui(t)
	lead.SetTrickNumberForTest(3)
	lead.SetCurrentPlayerIdxForTest(0)
	assert.Contains(t, p.HintOutput(lead), i18n.T("slobberhannes.hintReasonLeadLow"))

	danger := newSlobberhannesForCui(t)
	danger.SetTrickNumberForTest(0)
	danger.SetCurrentPlayerIdxForTest(0)
	danger.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
	})
	assert.Contains(t, p.HintOutput(danger), i18n.T("slobberhannes.hintReasonAvoid"))

	safe := newSlobberhannesForCui(t)
	safe.SetTrickNumberForTest(3)
	safe.SetCurrentPlayerIdxForTest(0)
	safe.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
	})
	assert.Contains(t, p.HintOutput(safe), i18n.T("slobberhannes.hintReasonDump"))
}

func TestSlobberhannesCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)
	s := newSlobberhannesForCui(t)
	s.GiveUp()
	assert.Contains(t, p.HintOutput(s), i18n.T("slobberhannes.hintNone"))
}

func TestSlobberhannesCuiPresenterActionLogOutput(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)
	s := newSlobberhannesForCui(t)
	s.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(s))
}

// **♣Q は位置ではなく中身の罰点** (#5745)。場に出た瞬間に言わないと、
// 取ってから penaltyMarks で気づくことになる。
func TestSlobberhannesCuiPresenterQueenWarning(t *testing.T) {
	p := new(SlobberhannesCuiPresenter)

	s := newSlobberhannesForCui(t)
	s.SetTrickNumberForTest(3) // 位置の警告が出ない中間のトリック
	s.SetCurrentPlayerIdxForTest(0)
	s.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 12, true)},
	})

	out := p.Output(s, nil)
	assert.Contains(t, out, fixedPart("slobberhannes.warnQueen"))
	// 位置の警告は出ていない (中間のトリック)。
	assert.NotContains(t, out, fixedPart("slobberhannes.warnFirst"))

	// 負のコントロール: 別スートの Q では出ない。
	other := newSlobberhannesForCui(t)
	other.SetTrickNumberForTest(3)
	other.SetCurrentPlayerIdxForTest(0)
	other.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 12, true)},
	})
	assert.NotContains(t, p.Output(other, nil), fixedPart("slobberhannes.warnQueen"))

	// 最初のトリックと同時でも両方出る (受け入れ条件4)。
	both := newSlobberhannesForCui(t)
	both.SetTrickNumberForTest(0)
	both.SetCurrentPlayerIdxForTest(0)
	both.SetCurrentTrickForTest([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 12, true)},
	})
	bothOut := p.Output(both, nil)
	assert.Contains(t, bothOut, fixedPart("slobberhannes.warnFirst"))
	assert.Contains(t, bothOut, fixedPart("slobberhannes.warnQueen"))
}
