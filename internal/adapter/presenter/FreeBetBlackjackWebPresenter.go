//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FreeBetBlackjackWebPresenter フリーベット・ブラックジャックWebプレゼンタークラス
type FreeBetBlackjackWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。**
func (cp *FreeBetBlackjackWebPresenter) Output(c interfaces.FreeBetBlackjackGame, lastErr error) string {
	resObj := new(controller.FreeBetBlackjackWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Hands = freeBetHandsToOutput(c)
	resObj.ActiveHand = c.GetActiveHandIdx()
	resObj.DealerCards = cardsToOutputOrEmpty(c.GetDealerCards())
	resObj.DealerScore = c.GetDealerScore()
	resObj.DealerPushed22 = c.IsDealerPushed22()
	resObj.CanFreeDouble = c.CanFreeDouble()
	resObj.CanFreeSplit = c.CanFreeSplit()
	resObj.AnteBet = c.GetAnteBet()
	resObj.Payout = c.GetPayout()
	resObj.Chips = c.GetChips()
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.Config = &controller.FreeBetWebOutCfg{
		InitialChips: c.GetConfig().InitialChips,
		DefaultAnte:  c.GetConfig().DefaultAnte,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "freebet.result.broke"
	}

	return marshalOrError(resObj)
}

// freeBetHandsToOutput は手札ごとの状態を組み立てる。
//
// **プレイヤーの金とハウスの金は別々に載せる。** 合算すると画面が「いくら失うのか」
// を出せなくなる ── このゲームでいちばん見せたい数字はそこ。
func freeBetHandsToOutput(c interfaces.FreeBetBlackjackGame) []*controller.FreeBetWebOutputHand {
	hands := c.GetHands()
	results := c.GetResults()
	out := make([]*controller.FreeBetWebOutputHand, 0, len(hands))
	for i, h := range hands {
		if h == nil {
			continue
		}
		result := domain.FreeBetResultNone
		if i < len(results) {
			result = results[i]
		}
		out = append(out, &controller.FreeBetWebOutputHand{
			Cards:     cardsToOutputOrEmpty(h.GetCards()),
			Score:     h.GetScore(),
			Bet:       h.GetBet(),
			FreeBet:   c.GetFreeBet(i),
			IsSoft:    h.IsSoft(),
			Stood:     h.IsStood(),
			Doubled:   h.IsDoubled(),
			Busted:    h.IsBusted(),
			Blackjack: domain.FreeBetIsNatural(h),
			Result:    int(result),
		})
	}
	return out
}

// ActionLogOutput 棋譜をJSON出力
func (cp *FreeBetBlackjackWebPresenter) ActionLogOutput(c interfaces.FreeBetBlackjackGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *FreeBetBlackjackWebPresenter) HintOutput(c interfaces.FreeBetBlackjackGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{"action": h.Action, "reason": h.Reason})
}
