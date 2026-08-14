//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DoubleAttackBlackjackWebPresenter 追加ベット・ブラックジャックWebプレゼンタークラス
type DoubleAttackBlackjackWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。** 配る前を素の変換に通すと JSON が `null` になり、
// TS 側が非 optional な配列を約束しているのでページが落ちます。
func (cp *DoubleAttackBlackjackWebPresenter) Output(c interfaces.DoubleAttackBlackjackGame, lastErr error) string {
	resObj := new(controller.DoubleAttackBlackjackWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Hands = doubleAttackHandsToOutput(c)
	resObj.ActiveHand = c.GetActiveHandIdx()
	// **アップカードだけの間は 1 枚しか無い。** サーバがそもそも 2 枚目を持っていない
	// ので、ここで伏せる細工は要らない (持っていないものは出せない)。
	resObj.DealerCards = cardsToOutputOrEmpty(c.GetDealerCards())
	resObj.DealerHoleDealt = c.IsDealerHoleDealt()
	if c.IsDealerHoleDealt() {
		resObj.DealerScore = c.GetDealerScore()
	}
	resObj.MaxAttackBet = c.MaxAttackBet()
	resObj.CanDouble = c.CanDouble()
	resObj.CanSplit = c.CanSplit()
	resObj.AnteBet = c.GetAnteBet()
	resObj.AttackBet = c.GetAttackBet()
	resObj.BustItBet = c.GetBustItBet()
	resObj.Payout = c.GetPayout()
	resObj.BustItPayout = c.GetBustItPayout()
	resObj.Chips = c.GetChips()
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.Config = &controller.DoubleAttackWebOutCfg{
		InitialChips: c.GetConfig().InitialChips,
		DefaultAnte:  c.GetConfig().DefaultAnte,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "doubleattack.result.broke"
	}

	return marshalOrError(resObj)
}

// doubleAttackHandsToOutput は手札ごとの状態を組み立てる。
func doubleAttackHandsToOutput(c interfaces.DoubleAttackBlackjackGame) []*controller.DoubleAttackWebOutputHand {
	hands := c.GetHands()
	results := c.GetResults()
	out := make([]*controller.DoubleAttackWebOutputHand, 0, len(hands))
	for i, h := range hands {
		if h == nil {
			continue
		}
		result := domain.DoubleAttackResultNone
		if i < len(results) {
			result = results[i]
		}
		out = append(out, &controller.DoubleAttackWebOutputHand{
			Cards:     cardsToOutputOrEmpty(h.GetCards()),
			Score:     h.GetScore(),
			Bet:       h.GetBet(),
			IsSoft:    h.IsSoft(),
			Stood:     h.IsStood(),
			Doubled:   h.IsDoubled(),
			Busted:    h.IsBusted(),
			Blackjack: h.IsBlackJack(),
			Result:    int(result),
		})
	}
	return out
}

// ActionLogOutput 棋譜をJSON出力
func (cp *DoubleAttackBlackjackWebPresenter) ActionLogOutput(c interfaces.DoubleAttackBlackjackGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *DoubleAttackBlackjackWebPresenter) HintOutput(c interfaces.DoubleAttackBlackjackGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{"action": h.Action, "reason": h.Reason})
}
