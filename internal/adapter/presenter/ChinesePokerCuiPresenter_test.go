//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestChinesePokerCuiPresenter_Output_BetPhase(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	cp := domain.NewDefaultChinesePoker()
	result := pp.Output(cp, nil)
	assert.Contains(t, result, "----------")
}

func TestChinesePokerCuiPresenter_Output_WithError(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	cp := domain.NewDefaultChinesePoker()
	testErr := domain.NewDomainError(domain.ErrWrongPhase, "test error")
	result := pp.Output(cp, testErr)
	assert.Contains(t, result, "test error")
}

func TestChinesePokerCuiPresenter_Output_SetHandsPhase(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	cp := domain.NewDefaultChinesePoker()
	_ = cp.Bet(100)
	result := pp.Output(cp, nil)
	assert.Contains(t, result, "[0]")
}

func TestChinesePokerCuiPresenter_Output_EndPhaseWin(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetPhase(domain.ChinesePokerPhaseEnd)
	cp.SetResult(domain.GameResultWin)
	cp.SetBet(100)
	cp.SetPayout(200)
	cp.SetPlayerFront([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
	})
	cp.SetPlayerMiddle([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
	})
	cp.SetPlayerBack([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
	})
	cp.SetDealerFront([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignDiamond, 6, false),
	})
	cp.SetDealerMiddle([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 11, false),
	})
	cp.SetDealerBack([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 12, false),
	})

	result := pp.Output(cp, nil)
	assert.Contains(t, result, "----------")
	assert.Contains(t, result, "200")
}

func TestChinesePokerCuiPresenter_Output_EndPhaseLose(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetPhase(domain.ChinesePokerPhaseEnd)
	cp.SetResult(domain.GameResultLose)
	cp.SetBet(100)
	result := pp.Output(cp, nil)
	assert.Contains(t, result, "----------")
}

func TestChinesePokerCuiPresenter_Output_ScoopWin(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	cp := domain.NewDefaultChinesePoker()
	cp.SetGameEndFlag(true)
	cp.SetPhase(domain.ChinesePokerPhaseEnd)
	cp.SetResult(domain.GameResultWin)
	cp.SetScoop(true)
	cp.SetBet(100)
	cp.SetPayout(400)
	result := pp.Output(cp, nil)
	assert.Contains(t, result, "----------")
}

func TestChinesePokerCuiPresenter_ActionLogOutput(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	cp := domain.NewDefaultChinesePoker()
	result := pp.ActionLogOutput(cp)
	assert.NotEmpty(t, result)
}

func TestChinesePokerCuiPresenter_PhaseStr(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	assert.NotEmpty(t, pp.phaseStr(domain.ChinesePokerPhaseBet))
	assert.NotEmpty(t, pp.phaseStr(domain.ChinesePokerPhaseSetHands))
	assert.NotEmpty(t, pp.phaseStr(domain.ChinesePokerPhaseEnd))
	assert.NotEmpty(t, pp.phaseStr(99))
}

func TestChinesePokerCuiPresenter_RankStr(t *testing.T) {
	pp := &ChinesePokerCuiPresenter{}
	assert.NotEmpty(t, pp.frontRankStr(domain.ThreeCardHandPair))
	assert.NotEmpty(t, pp.frontRankStr(99))
	assert.NotEmpty(t, pp.fiveCardRankStr(domain.PokerHandFlush))
	assert.NotEmpty(t, pp.fiveCardRankStr(99))
}

// **フロントの役名も日本語ロケールで日本語。**英語の ThreeCardHandNames を
// そのまま返していた (#4985 レビュー指摘の follow-up)。
func TestChinesePokerCuiPresenter_FrontRankIsTranslated(t *testing.T) {
	pp := new(ChinesePokerCuiPresenter)
	assert.Equal(t, "ハイカード", pp.frontRankStr(domain.ThreeCardHandHighCard))
	assert.Equal(t, "ストレートフラッシュ", pp.frontRankStr(domain.ThreeCardHandStraightFlush))
	assert.NotContains(t, pp.frontRankStr(domain.ThreeCardHandFlush), "Flush")
	// 範囲外は未知ランクの文言に落ちる。キー文字列を出さない。
	assert.NotContains(t, pp.frontRankStr(99), "pokerhand.")
	assert.NotEmpty(t, pp.frontRankStr(99))
}
