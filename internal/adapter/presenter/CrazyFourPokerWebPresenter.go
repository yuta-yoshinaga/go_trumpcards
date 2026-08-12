//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CrazyFourPokerWebPresenter クレイジー 4 ポーカーWebプレゼンタークラス
type CrazyFourPokerWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。** 配る前を素の変換に通すと JSON が `null` になり、
// TS 側が非 optional な配列を約束しているのでページが落ちます。
func (cp *CrazyFourPokerWebPresenter) Output(c interfaces.CrazyFourPokerGame, lastErr error) string {
	resObj := new(controller.CrazyFourPokerWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.PlayerHand = cardsToOutputOrEmpty(c.GetPlayerHand())
	resObj.PlayerBest = cardsToOutputOrEmpty(c.GetPlayerBest())
	// **決着するまでディーラーの手は伏せる。** 5 枚から最良の 4 枚を選ぶゲームなので、
	// 見えていると判断がまるごと変わってしまう。
	if c.GetPhase() == domain.CrazyFourPokerPhaseResult {
		resObj.DealerHand = cardsToOutputOrEmpty(c.GetDealerHand())
		resObj.DealerBest = cardsToOutputOrEmpty(c.GetDealerBest())
		resObj.DealerHandRank = c.GetDealerHandRank()
		resObj.DealerQualifies = c.DealerQualifies()
	} else {
		resObj.DealerHand = make([]*controller.WebOutputCard, 0)
		resObj.DealerBest = make([]*controller.WebOutputCard, 0)
	}
	resObj.PlayerHandRank = c.GetPlayerHandRank()
	resObj.HasAcesOrBetter = c.PlayerHasAcesOrBetter()
	resObj.MaxMultiplier = c.MaxPlayMultiplier()
	resObj.AnteBet = c.GetAnteBet()
	resObj.SuperBet = c.GetSuperBet()
	resObj.QueensUpBet = c.GetQueensUpBet()
	resObj.PlayBet = c.GetPlayBet()
	resObj.PlayMultiplier = c.GetPlayMultiplier()
	resObj.Result = int(c.GetResult())
	resObj.Payout = c.GetPayout()
	resObj.Chips = c.GetChips()
	resObj.MinTotalWager = c.GetMinTotalWager()
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.Config = &controller.CrazyFourPokerWebOutCfg{
		InitialChips: c.GetConfig().InitialChips,
		DefaultAnte:  c.GetConfig().DefaultAnte,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "crazyfourpoker.result.broke"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (cp *CrazyFourPokerWebPresenter) ActionLogOutput(c interfaces.CrazyFourPokerGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *CrazyFourPokerWebPresenter) HintOutput(c interfaces.CrazyFourPokerGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{
		"multiplier": h.Multiplier,
		"reason":     h.Reason,
	})
}
