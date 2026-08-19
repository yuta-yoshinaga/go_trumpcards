//go:build test

package presenter

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newStealingBundlesForCui(t *testing.T) *domain.StealingBundles {
	t.Helper()
	s := domain.NewDefaultStealingBundles()
	s.Reset()
	return s
}

func TestStealingBundlesCuiPresenterOutput(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)
	s := newStealingBundlesForCui(t)
	out := p.Output(s, nil)

	assert.Contains(t, out, i18n.T("stealingbundles.helpTitle"))
	assert.Contains(t, out, fixedPart("stealingbundles.header"))
	// **束の一番上が弱点、というのが規則そのもの。**
	assert.Contains(t, out, i18n.T("stealingbundles.rule"))
	assert.Contains(t, out, fixedPart("stealingbundles.table"))
	// 全員の席行に手札・束・一番上が出る。
	assert.Len(t, regexp.MustCompile(`手札\d+枚 束\d+枚 一番上\[`).FindAllString(out, -1),
		domain.StealingBundlesDefaultPlayerCnt)
}

// **空の場も「取れない」という情報。** 行が消えると見落としと区別が付きません。
func TestStealingBundlesCuiPresenterShowsAnEmptyTable(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)
	s := newStealingBundlesForCui(t)
	s.SetTableCardsForTest(nil)
	out := p.Output(s, nil)
	assert.Contains(t, out, fixedPart("stealingbundles.table"))
	assert.Contains(t, out, i18n.T("stealingbundles.tableEmpty"))
}

// **束の一番上は全員に見えます。**
func TestStealingBundlesCuiPresenterShowsBundleTops(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)
	s := newStealingBundlesForCui(t)
	assert.Contains(t, p.Output(s, nil), i18n.T("stealingbundles.bundleEmpty"))

	s.GetPlayer(1).SetBundle([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
	})
	out := p.Output(s, nil)
	assert.Contains(t, out, "束2枚")
	assert.Contains(t, out, cuiCardStr(domain.NewCard(domain.CardDesignDiamond, 9, false)))
}

// **取れるときは置けません。** 促しを言い分けます。
func TestStealingBundlesCuiPresenterPromptsByWhatIsLegal(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)
	s := newStealingBundlesForCui(t)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	s.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 7, false))

	out := p.Output(s, nil)
	assert.Contains(t, out, i18n.T("stealingbundles.promptMustCapture"))
	assert.Contains(t, out, i18n.T("stealingbundles.promptTake"))
	assert.NotContains(t, out, i18n.T("stealingbundles.promptTrail"))

	// **負のコントロール: 取れなければ置く促しに変わる。**
	s.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 4, false))
	for i := 1; i < s.GetPlayerCnt(); i++ {
		s.GetPlayer(i).SetBundle(nil)
	}
	out = p.Output(s, nil)
	assert.Contains(t, out, i18n.T("stealingbundles.promptNoCapture"))
	assert.Contains(t, out, i18n.T("stealingbundles.promptTrail"))
	assert.NotContains(t, out, i18n.T("stealingbundles.promptMustCapture"))
}

func TestStealingBundlesCuiPresenterGameEndBanners(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)

	won := newStealingBundlesForCui(t)
	won.DrainDeckForTest()
	won.SetCurrentPlayerIdxForTest(0)
	won.SetTableCardsForTest([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	won.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 7, false))
	for i := 1; i < won.GetPlayerCnt(); i++ {
		won.GiveHandForTest(i)
	}
	require.NoError(t, won.PlayerTake(0))
	require.True(t, won.GetGameEndFlag())
	out := p.Output(won, nil)
	assert.Contains(t, out, fixedPart("stealingbundles.gameEndYou"))
	assert.NotContains(t, out, i18n.T("stealingbundles.promptTake"), "終局後は促さない")

	lost := newStealingBundlesForCui(t)
	lost.GiveUp()
	assert.Contains(t, p.Output(lost, nil), fixedPart("stealingbundles.gameEndCpu"))
}

func TestStealingBundlesCuiPresenterShowsErrors(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)
	assert.Contains(t, p.Output(newStealingBundlesForCui(t), assert.AnError), assert.AnError.Error())
}

func TestStealingBundlesCuiPresenterHintOutput(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)
	s := newStealingBundlesForCui(t)
	s.SetCurrentPlayerIdxForTest(0)
	s.SetTableCardsForTest([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
	s.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 5, false))

	out := p.HintOutput(s)
	assert.Contains(t, out, fixedPart("stealingbundles.hintCard"))
	assert.Contains(t, out, i18n.T("stealingbundles.hintReasonTake"))

	// **略奪の助言は相手の名前も出します。**
	s.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 9, false))
	s.GetPlayer(1).SetBundle([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 9, false)})
	out = p.HintOutput(s)
	assert.Contains(t, out, fixedPart("stealingbundles.hintSteal"))
	assert.Contains(t, out, i18n.T("stealingbundles.hintReasonSteal"))

	s.GiveUp()
	assert.Contains(t, p.HintOutput(s), i18n.T("stealingbundles.hintNone"))
}

func TestStealingBundlesCuiPresenterActionLogOutput(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)
	s := newStealingBundlesForCui(t)
	s.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(s))
}

// **盗みは相手の束を丸ごと消す。** 場から取っただけの手と同じ印では、盤面に
// 痕跡が残らないこのゲームで区別が付かない (#5767)。
func TestStealingBundlesCuiPresenterDistinguishesTakesFromSteals(t *testing.T) {
	p := new(StealingBundlesCuiPresenter)

	took := newStealingBundlesForCui(t)
	took.SetCurrentPlayerIdxForTest(0)
	took.SetTableCardsForTest([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	took.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 7, false))
	require.NoError(t, took.PlayerTake(0))
	out := p.Output(took, nil)
	assert.Contains(t, out, i18n.T("stealingbundles.roleLastCaptureTake"))
	assert.NotContains(t, out, fixedPart("stealingbundles.roleLastCaptureSteal"))

	stole := newStealingBundlesForCui(t)
	stole.SetCurrentPlayerIdxForTest(0)
	stole.SetTableCardsForTest(nil)
	stole.GiveHandForTest(0, domain.NewCard(domain.CardDesignSpade, 9, false))
	stole.GetPlayer(1).SetBundle([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 9, false)})
	require.NoError(t, stole.PlayerSteal(0, 1))
	out = p.Output(stole, nil)
	// 被害者の席まで出す。
	assert.Contains(t, out, i18n.Tf("stealingbundles.roleLastCaptureSteal",
		"name", cuiPlayerName(stole.GetPlayer(1), 1)))
	assert.NotContains(t, out, i18n.T("stealingbundles.roleLastCaptureTake"))
}
