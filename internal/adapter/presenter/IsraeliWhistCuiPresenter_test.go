//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newIsraeliWhistForCui(t *testing.T) *domain.IsraeliWhist {
	t.Helper()
	w := domain.NewDefaultIsraeliWhist()
	w.Reset()
	return w
}

func TestIsraeliWhistCuiPresenterOutput(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	w := newIsraeliWhistForCui(t)

	out := p.Output(w, nil)
	assert.Contains(t, out, i18n.T("israeliwhist.helpTitle"))
	// **得点表は盤面から読めない。** 2 乗と全員一致の倍率を常時出す。
	assert.Contains(t, out, i18n.T("israeliwhist.scoreTable"))
	assert.Contains(t, out, fixedPart("israeliwhist.header"))
}

// オークション中と決着後で表示が入れ替わる。両側を踏む。
func TestIsraeliWhistCuiPresenterAuctionThenTrumpLine(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)

	during := newIsraeliWhistForCui(t)
	out := p.Output(during, nil)
	assert.Contains(t, out, fixedPart("israeliwhist.auctionLine"))
	assert.NotContains(t, out, fixedPart("israeliwhist.trumpLine"))

	after := newIsraeliWhistForCui(t)
	after.SetDeclarerForTest(1, 7, domain.CardDesignHeart)
	after.CloseAuctionForTest()
	out2 := p.Output(after, nil)
	assert.Contains(t, out2, fixedPart("israeliwhist.trumpLine"))
	assert.NotContains(t, out2, fixedPart("israeliwhist.auctionLine"))
}

// 1 段階目の立場が席ごとに出る。3 種すべて踏む。
func TestIsraeliWhistCuiPresenterRoles(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	w := newIsraeliWhistForCui(t)
	w.SetDeclarerForTest(0, 7, domain.CardDesignSpade)
	w.GetPlayer(0).SetAuction(7, domain.CardDesignSpade)
	w.GetPlayer(1).SetPassed(true)

	out := p.Output(w, nil)
	assert.Contains(t, out, fixedPart("israeliwhist.roleDeclarer"))
	assert.Contains(t, out, i18n.T("israeliwhist.rolePassed"))
	assert.Contains(t, out, i18n.T("israeliwhist.roleActive"))
}

// **落札者のノルマと禁止値は先に伝える。** 押してから断られるのでは遅い。
func TestIsraeliWhistCuiPresenterAnnouncesQuotaAndRestriction(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)

	quota := newIsraeliWhistForCui(t)
	quota.SetDeclarerForTest(0, 9, domain.CardDesignSpade)
	quota.CloseAuctionForTest()
	quota.SetBidPlayerIdxForTest(0)
	assert.Contains(t, p.Output(quota, nil), fixedPart("israeliwhist.promptBidQuota"))

	restricted := newIsraeliWhistForCui(t)
	restricted.SetDeclarerForTest(1, 5, domain.CardDesignSpade)
	restricted.CloseAuctionForTest()
	restricted.SetBidsForTest(map[int]int{1: 5, 2: 4, 3: 3})
	restricted.SetBidPlayerIdxForTest(0)
	assert.Contains(t, p.Output(restricted, nil), fixedPart("israeliwhist.promptBidRestricted"))

	// 負のコントロール: 落札者でもなく最後でもなければ、どちらも出さない。
	plain := newIsraeliWhistForCui(t)
	plain.SetDeclarerForTest(1, 5, domain.CardDesignSpade)
	plain.CloseAuctionForTest()
	plain.SetBidPlayerIdxForTest(0)
	outPlain := p.Output(plain, nil)
	assert.NotContains(t, outPlain, fixedPart("israeliwhist.promptBidQuota"))
	assert.NotContains(t, outPlain, fixedPart("israeliwhist.promptBidRestricted"))
}

// **入札できる番と待つ番で案内が変わる。** 両側を踏む。
func TestIsraeliWhistCuiPresenterAuctionPrompts(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)

	mine := newIsraeliWhistForCui(t)
	mine.SetAuctionPlayerIdxForTest(0)
	assert.Contains(t, p.Output(mine, nil), i18n.T("israeliwhist.promptAuction"))

	theirs := newIsraeliWhistForCui(t)
	theirs.SetAuctionPlayerIdxForTest(2)
	out := p.Output(theirs, nil)
	assert.Contains(t, out, i18n.T("israeliwhist.promptAuctionWait"))
	assert.NotContains(t, out, i18n.T("israeliwhist.promptAuction"))
}

func TestIsraeliWhistCuiPresenterRoundEnd(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	w := newIsraeliWhistForCui(t)
	w.SetPhaseForTest(domain.IsraeliWhistPhaseRoundEnd)

	out := p.Output(w, nil)
	assert.Contains(t, out, i18n.T("israeliwhist.promptRoundEnd"))
	assert.Contains(t, out, i18n.T("israeliwhist.promptNext"))
	assert.NotContains(t, out, i18n.T("israeliwhist.promptPlay"))
}

func TestIsraeliWhistCuiPresenterPlayPrompt(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	w := newIsraeliWhistForCui(t)
	w.SetPhaseForTest(domain.IsraeliWhistPhasePlay)
	w.SetCurrentPlayerIdxForTest(0)

	out := p.Output(w, nil)
	assert.Contains(t, out, i18n.T("israeliwhist.promptPlay"))
	assert.Contains(t, out, fixedPart("israeliwhist.promptCurrentPlayer"))
}

func TestIsraeliWhistCuiPresenterError(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	assert.Contains(t, p.Output(newIsraeliWhistForCui(t), assert.AnError), assert.AnError.Error())
}

func TestIsraeliWhistCuiPresenterGameEnd(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)

	for _, tc := range []struct {
		name    string
		scores  [4]int
		wantKey string
	}{
		{"you win", [4]int{80, 10, 10, 10}, "israeliwhist.gameEndWin"},
		{"a cpu wins", [4]int{10, 80, 10, 10}, "israeliwhist.gameEndLose"},
		{"a tie", [4]int{30, 30, 30, 30}, "israeliwhist.gameEndTie"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newIsraeliWhistForCui(t)
			for i, s := range tc.scores {
				w.GetPlayer(i).SetTotalScore(s)
			}
			w.FinishGameForTest()

			out := p.Output(w, nil)
			assert.Contains(t, out, fixedPart(tc.wantKey))
			assert.NotContains(t, out, i18n.T("israeliwhist.promptPlay"))
		})
	}
}

// オークションのヒントはスートと数まで出す。
func TestIsraeliWhistCuiPresenterHintDuringAuction(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	w := newIsraeliWhistForCui(t)
	w.SetAuctionPlayerIdxForTest(0)

	out := p.HintOutput(w)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "israeliwhistAuctionBid", "生のキーが出ていたら未登録")
	assert.NotContains(t, out, "israeliwhistAuctionPass")
}

// ノルマを満たすヒントは数まで出す。
func TestIsraeliWhistCuiPresenterHintMeetsQuota(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	w := newIsraeliWhistForCui(t)
	w.SetDeclarerForTest(0, domain.IsraeliWhistHandSize, domain.CardDesignSpade)
	w.CloseAuctionForTest()
	w.SetBidPlayerIdxForTest(0)

	out := p.HintOutput(w)
	assert.Contains(t, out, "HINT")
	assert.NotContains(t, out, "israeliwhistMeetQuota")
}

// プレイ中の 2 つの理由キーが両方とも文言に解決される。
func TestIsraeliWhistCuiPresenterHintDuringPlay(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)

	short := newIsraeliWhistForCui(t)
	short.SetPhaseForTest(domain.IsraeliWhistPhasePlay)
	short.SetCurrentPlayerIdxForTest(0)
	short.GetPlayer(0).SetBid(5)
	assert.Contains(t, p.HintOutput(short), i18n.T("israeliwhist.hintReasonWinTrick"))

	done := newIsraeliWhistForCui(t)
	done.SetPhaseForTest(domain.IsraeliWhistPhasePlay)
	done.SetCurrentPlayerIdxForTest(0)
	done.GetPlayer(0).SetBid(0)
	assert.Contains(t, p.HintOutput(done), i18n.T("israeliwhist.hintReasonDuck"))
}

func TestIsraeliWhistCuiPresenterHintNoneAfterGameEnd(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	w := newIsraeliWhistForCui(t)
	w.GiveUp()
	assert.Contains(t, p.HintOutput(w), i18n.T("israeliwhist.hintNone"))
}

// スート名は 4 つとも解決され、未決定は「なし」になる。
func TestIsraeliWhistSuitName(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		assert.NotEqual(t, i18n.T("israeliwhist.suitNone"), israeliWhistSuitName(suit))
	}
	assert.Equal(t, i18n.T("israeliwhist.suitNone"), israeliWhistSuitName(0))
}

func TestIsraeliWhistCuiPresenterActionLogOutput(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)
	w := newIsraeliWhistForCui(t)
	w.GiveUp()
	require.NotEmpty(t, p.ActionLogOutput(w))
}

// **2 倍はこのゲームの起伏そのもの** (#5752)。これまで畳まれたアクション
// ログにしか残っておらず、点が普段の倍動いた理由が読めなかった。
func TestIsraeliWhistCuiPresenterAnnouncesTheDoubledRound(t *testing.T) {
	p := new(IsraeliWhistCuiPresenter)

	setRound := func(bids, tricks [domain.IsraeliWhistPlayerCnt]int) *domain.IsraeliWhist {
		w := newIsraeliWhistForCui(t)
		for i := range domain.IsraeliWhistPlayerCnt {
			pl := w.GetPlayer(i)
			pl.SetBid(bids[i])
			pl.ResetTricks()
			for range tricks[i] {
				pl.AddTrick([]*domain.Card{})
			}
		}
		w.FinishRoundForTest()
		return w
	}

	allExact := setRound([domain.IsraeliWhistPlayerCnt]int{3, 4, 3, 3}, [domain.IsraeliWhistPlayerCnt]int{3, 4, 3, 3})
	assert.Contains(t, p.Output(allExact, nil), fixedPart("israeliwhist.doubledAllExact"))

	allMissed := setRound([domain.IsraeliWhistPlayerCnt]int{3, 4, 3, 3}, [domain.IsraeliWhistPlayerCnt]int{0, 1, 5, 7})
	missedOut := p.Output(allMissed, nil)
	assert.Contains(t, missedOut, fixedPart("israeliwhist.doubledAllMissed"))
	// **理由を取り違えない。**全員外しなのに「全員的中」と出したら嘘になる。
	assert.NotContains(t, missedOut, fixedPart("israeliwhist.doubledAllExact"))

	// 通常ラウンドでは何も出ない (負のコントロール)。
	mixed := setRound([domain.IsraeliWhistPlayerCnt]int{3, 4, 3, 3}, [domain.IsraeliWhistPlayerCnt]int{3, 0, 5, 5})
	mixedOut := p.Output(mixed, nil)
	assert.NotContains(t, mixedOut, fixedPart("israeliwhist.doubledAllExact"))
	assert.NotContains(t, mixedOut, fixedPart("israeliwhist.doubledAllMissed"))
}
