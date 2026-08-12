//go:build test

package presenter

import (
	"math/rand"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newGoofspielForCui(t *testing.T) *domain.Goofspiel {
	t.Helper()
	g := domain.NewDefaultGoofspiel()
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	return g
}

func TestGoofspielCuiPresenterOutput(t *testing.T) {
	p := new(GoofspielCuiPresenter)
	g := newGoofspielForCui(t)
	out := p.Output(g, nil)

	assert.Contains(t, out, i18n.T("goofspiel.helpTitle"))
	assert.Contains(t, out, fixedPart("goofspiel.header"))
	// **同時入札であることが規則そのもの。**
	assert.Contains(t, out, i18n.T("goofspiel.rule"))
	assert.Contains(t, out, fixedPart("goofspiel.prize"))
	assert.Len(t, regexp.MustCompile(`残り\d+枚 得点\d+点`).FindAllString(out, -1),
		domain.GoofspielDefaultPlayerCnt, "全員の席行に残り枚数と得点が出る")
	assert.Contains(t, out, i18n.T("goofspiel.promptBid"))
	assert.NotContains(t, out, fixedPart("goofspiel.carried"), "持ち越しは無い")
}

// **伏せたことは見せますが、中身は公開まで見せません。**
func TestGoofspielCuiPresenterHidesTheBidUntilRevealed(t *testing.T) {
	p := new(GoofspielCuiPresenter)
	g := newGoofspielForCui(t)
	require.NoError(t, g.BidForTest(0, 5))

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("goofspiel.roleBid"))
	// **fixedPart は使えません。** この鍵は placeholder で始まるので固定部が空になり、
	// 「空文字を含まない」という常に落ちる検査になります。
	assert.NotContains(t, out, "を出した]")
	assert.Contains(t, out, i18n.T("goofspiel.promptWaiting"))
	assert.NotContains(t, out, i18n.T("goofspiel.promptBid"))

	require.NoError(t, g.BidForTest(1, 8))
	g.ResolveForTest()
	out = p.Output(g, nil)
	assert.Contains(t, out, "を出した]")
	assert.Contains(t, out, i18n.T("goofspiel.promptNext"))
}

// **同点は誰も取りません。** 勝者が居ない結果を言い分けます。
func TestGoofspielCuiPresenterDistinguishesATie(t *testing.T) {
	p := new(GoofspielCuiPresenter)
	g := newGoofspielForCui(t)
	g.SetCurrentPrizeForTest(domain.NewCard(domain.GoofspielPrizeSuit(), 9, false))
	require.NoError(t, g.BidForTest(0, 6))
	require.NoError(t, g.BidForTest(1, 6))
	g.ResolveForTest()

	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("goofspiel.promptTie"))
	assert.NotContains(t, out, "点を獲得しました")

	// **負のコントロール: 決着したラウンドは勝者を出す。**
	won := newGoofspielForCui(t)
	won.SetCurrentPrizeForTest(domain.NewCard(domain.GoofspielPrizeSuit(), 9, false))
	require.NoError(t, won.BidForTest(0, 4))
	require.NoError(t, won.BidForTest(1, 11))
	won.ResolveForTest()
	out = p.Output(won, nil)
	assert.Contains(t, out, "点を獲得しました")
	assert.NotContains(t, out, i18n.T("goofspiel.promptTie"))
}

// **持ち越しは「今回の賞が増える」こと。** 見えないと計算が合いません。
func TestGoofspielCuiPresenterShowsTheCarryOver(t *testing.T) {
	p := new(GoofspielCuiPresenter)
	cfg := domain.DefaultGoofspielConfig()
	cfg.TieRule = domain.GoofspielTieCarryOver
	g := domain.NewGoofspiel(nil, cfg)
	g.SetRand(rand.New(rand.NewSource(1)))
	g.Reset()
	g.SetCurrentPrizeForTest(domain.NewCard(domain.GoofspielPrizeSuit(), 9, false))
	require.NoError(t, g.BidForTest(0, 6))
	require.NoError(t, g.BidForTest(1, 6))
	g.ResolveForTest()
	require.NoError(t, g.NextRound())

	assert.Contains(t, p.Output(g, nil), fixedPart("goofspiel.carried"))
}

func TestGoofspielCuiPresenterGameEndBanners(t *testing.T) {
	p := new(GoofspielCuiPresenter)

	g := newGoofspielForCui(t)
	for !g.GetGameEndFlag() {
		require.NoError(t, g.PlayerBid(g.GetPlayer(0).GetCardsSize()-1))
		if g.GetPhase() == domain.GoofspielPhaseReveal && !g.GetGameEndFlag() {
			require.NoError(t, g.NextRound())
		}
	}
	out := p.Output(g, nil)
	// **fixedPart は使えません。** you と cpu は「ゲーム終了！ 」を共有するので、
	// 固定部で見ると片方の検査がもう片方でも通ってしまいます。
	assert.Contains(t, out, "点でいちばん多く取りました")
	assert.NotContains(t, out, i18n.T("goofspiel.promptBid"), "終局後は促さない")

	// 投了は必ず CPU の勝ち。**「あなたが」ではないことまで見ます。**
	lost := newGoofspielForCui(t)
	lost.GiveUp()
	lostOut := p.Output(lost, nil)
	assert.Contains(t, lostOut, "点でいちばん多く取りました")
	assert.NotContains(t, lostOut, "あなたが")
}

func TestGoofspielCuiPresenterShowsErrors(t *testing.T) {
	p := new(GoofspielCuiPresenter)
	assert.Contains(t, p.Output(newGoofspielForCui(t), assert.AnError), assert.AnError.Error())
}

func TestGoofspielCuiPresenterHintOutput(t *testing.T) {
	p := new(GoofspielCuiPresenter)
	g := newGoofspielForCui(t)

	assert.Contains(t, p.HintOutput(g), fixedPart("goofspiel.hintCard"))

	g.SetCurrentPrizeForTest(domain.NewCard(domain.GoofspielPrizeSuit(), 12, false))
	assert.Contains(t, p.HintOutput(g), i18n.T("goofspiel.hintReasonHighPrize"))

	g.SetCurrentPrizeForTest(domain.NewCard(domain.GoofspielPrizeSuit(), 2, false))
	assert.Contains(t, p.HintOutput(g), i18n.T("goofspiel.hintReasonLowPrize"))

	g.GiveUp()
	assert.Contains(t, p.HintOutput(g), i18n.T("goofspiel.hintNone"))
}

func TestGoofspielCuiPresenterActionLogOutput(t *testing.T) {
	p := new(GoofspielCuiPresenter)
	g := newGoofspielForCui(t)
	g.GiveUp()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
