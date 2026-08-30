package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestDoudizhuCuiPresenter_Output_BidPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()
	// Put the human on turn so the bid prompt renders.
	humanIdx := 0
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		if dg.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			break
		}
	}
	dg.SetCurrentTurn(humanIdx)

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, nil)
	assert.NotEmpty(t, out)
	// The current bidder line and the human bid prompt should both appear.
	bidderPrefix := strings.SplitN(i18n.T("doudizhu.currentBidder"), "{{", 2)[0]
	assert.Contains(t, out, bidderPrefix)
	assert.Contains(t, out, i18n.T("doudizhu.promptBid"))
}

func TestDoudizhuCuiPresenter_Output_PlayPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.SetPhase(domain.DoudizhuPhasePlay)
	dg.SetLandlordIdx(0)
	dg.SetCurrentTurn(0)
	dg.SetKittyCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	dg.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	dg.SetTableCombo(&domain.DoudizhuCombo{Type: domain.DoudizhuComboSingle, Rank: 10, Length: 1, Cards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)}})

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, nil)
	assert.NotEmpty(t, out)
}

// 最終得点の倍率はビッド額とボム回数で決まるのに、CUI はどちらも終局まで
// 出していなかった。**ボムを実際に打って**回数が追従することまで見る ——
// フィールドを立てるだけでは、表示が本当に今の値を読んでいるか分からない。
func TestDoudizhuCuiPresenter_ShowsTheBidAndBombCountDuringPlay(t *testing.T) {
	i18n.SetLang("ja")
	dg := newDoudizhuForPresenter()
	dg.SetPhase(domain.DoudizhuPhasePlay)
	dg.SetLandlordIdx(0)
	dg.SetCurrentTurn(0)
	dg.SetKittyCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})

	p := new(presenter.DoudizhuCuiPresenter)
	before := p.Output(dg, nil)
	assert.Contains(t, before, i18n.Tf("doudizhu.bidAndBombs",
		"bid", strconv.Itoa(dg.GetBaseBid()), "bombs", "0"))

	// 手番の席に 7 のフォーカード（ボム）だけを持たせて実際に打つ。
	human := dg.GetPlayer(0)
	human.Reset()
	for _, d := range []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond} {
		human.AddCard(domain.NewCard(d, 7, false))
	}
	dg.SetTableCombo(nil)
	dg.SetCurrentTurn(0)
	require.True(t, dg.IsHumanTurn(), "the seat on turn must be the human for PlayerPlay")
	require.NoError(t, dg.PlayerPlay([]int{0, 1, 2, 3}))
	require.Equal(t, 1, dg.GetBombCount(), "playing four of a kind must count as a bomb")

	after := p.Output(dg, nil)
	assert.Contains(t, after, i18n.Tf("doudizhu.bidAndBombs",
		"bid", strconv.Itoa(dg.GetBaseBid()), "bombs", "1"))
	assert.NotContains(t, after, i18n.Tf("doudizhu.bidAndBombs",
		"bid", strconv.Itoa(dg.GetBaseBid()), "bombs", "0"))
}

func TestDoudizhuCuiPresenter_Output_EndPhase(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.SetPhase(domain.DoudizhuPhasePlay)
	dg.SetLandlordIdx(0)
	dg.SetBaseBid(1)
	dg.SetCurrentTurn(0)
	dg.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	dg.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	dg.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	_ = dg.PlayerPlay([]int{0})

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, nil)
	assert.NotEmpty(t, out)
}

func TestDoudizhuCuiPresenter_Output_Error(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()

	p := new(presenter.DoudizhuCuiPresenter)
	out := p.Output(dg, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestDoudizhuCuiPresenter_ActionLogOutput(t *testing.T) {
	dg := newDoudizhuForPresenter()
	dg.Reset()

	p := new(presenter.DoudizhuCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(dg))
}

// #5617: 地主の行だけ `Player %d` を直接組み立てており、日本語ロケールでも
// "Player 2" と出ていた。同じ画面の他の行は全部 cuiPlayerName を通っている。
func TestDoudizhuCuiPresenterNamesTheLandlordLikeEveryOtherLine(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	t.Run("human landlord reads as you", func(t *testing.T) {
		dg := newDoudizhuForPresenter()
		dg.SetPhase(domain.DoudizhuPhasePlay)
		dg.SetLandlordIdx(0) // 0 は人間
		dg.SetCurrentTurn(0)

		out := new(presenter.DoudizhuCuiPresenter).Output(dg, nil)
		assert.Contains(t, out, i18n.T("doudizhu.landlord")+": "+i18n.T("cuiPlayerYou"))
		// **英語の直書きが残っていないこと。**日本語ロケールで "Player 0" は出ない。
		assert.NotContains(t, out, "Player 0")
	})

	t.Run("cpu landlord reads like the other cpu lines", func(t *testing.T) {
		dg := newDoudizhuForPresenter()
		dg.SetPhase(domain.DoudizhuPhasePlay)
		dg.SetLandlordIdx(2)
		dg.SetCurrentTurn(0)

		out := new(presenter.DoudizhuCuiPresenter).Output(dg, nil)
		assert.Contains(t, out, i18n.T("doudizhu.landlord")+": "+i18n.Tf("cuiPlayerCpu", "idx", "2"))
		assert.NotContains(t, out, "Player 2")
	})
}
