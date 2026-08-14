//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MonteBankWebPresenter モンテバンクWebプレゼンタークラス
type MonteBankWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。**
func (cp *MonteBankWebPresenter) Output(c interfaces.MonteBankGame, lastErr error) string {
	resObj := new(controller.MonteBankWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Layout = monteBankLayoutToOutput(c)
	if gate := c.GetGate(); gate != nil {
		resObj.Gate = cardToOutput(gate)
	}
	resObj.Pick = c.GetPick()
	resObj.Bet = c.GetBet()
	resObj.Result = int(c.GetResult())
	resObj.Payout = c.GetPayout()
	resObj.Chips = c.GetChips()
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.PayoutMultiplier = domain.MonteBankPayout
	cfg := c.GetConfig()
	resObj.Config = &controller.MonteBankWebOutCfg{
		InitialChips: cfg.InitialChips, DefaultBet: cfg.DefaultBet,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "montebank.result.gameEnd"
	}

	return marshalOrError(resObj)
}

// monteBankLayoutToOutput は場札とその賭けやすさを組み立てる。
//
// **同じスートが何枚出ているかをサーバが数える。** ページに数え直させると、
// 控除率を決めている唯一の規則が 2 か所に分かれる。1 枚なら互角、2 枚以上なら
// 賭けるだけ損 ── その判定 (`isEven`) もここで付ける。
func monteBankLayoutToOutput(c interfaces.MonteBankGame) []*controller.MonteBankWebOutputCard {
	layout := c.GetLayout()
	out := make([]*controller.MonteBankWebOutputCard, 0, len(layout))
	for i, card := range layout {
		if card == nil {
			continue
		}
		count := c.SuitCountInLayout(card.GetDesign())
		out = append(out, &controller.MonteBankWebOutputCard{
			Card:            cardToOutput(card),
			SuitCount:       count,
			RemainingOfSuit: c.RemainingOfSuit(card.GetDesign()),
			IsEven:          count == 1,
			IsPicked:        i == c.GetPick(),
		})
	}
	return out
}

// ActionLogOutput 棋譜をJSON出力
func (cp *MonteBankWebPresenter) ActionLogOutput(c interfaces.MonteBankGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *MonteBankWebPresenter) HintOutput(c interfaces.MonteBankGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{"pickIdx": h.PickIdx, "reason": h.Reason})
}
