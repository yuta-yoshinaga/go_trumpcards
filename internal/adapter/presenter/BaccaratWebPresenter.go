//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BaccaratWebPresenter バカラWebプレゼンタークラス
type BaccaratWebPresenter struct {
}

// Output ゲーム状態を出力
func (bp *BaccaratWebPresenter) Output(b interfaces.BaccaratGame, lastErr error) string {
	resObj := new(controller.BaccaratWebOutput)

	// プレイヤーハンド
	resObj.PlayerHand = cardsToOutputOrEmpty(b.GetPlayerHand())

	// バンカーハンド
	resObj.BankerHand = cardsToOutputOrEmpty(b.GetBankerHand())

	resObj.PlayerHandValue = b.GetPlayerHandValue()
	resObj.BankerHandValue = b.GetBankerHandValue()
	resObj.Phase = b.GetPhase()
	resObj.Chips = b.GetChips()
	resObj.BetAmount = b.GetBetAmount()
	resObj.BetType = b.GetBetType()
	resObj.Result = int(b.GetResult())
	resObj.Payout = b.GetPayout()
	resObj.PlayerPairBet = b.GetPlayerPairBet()
	resObj.BankerPairBet = b.GetBankerPairBet()

	// 罫線履歴
	history := b.GetHistory()
	if history != nil {
		resObj.History = history
	} else {
		resObj.History = make([]int, 0)
	}

	// サイドベット結果
	sbResults := b.GetSideBetResults()
	if len(sbResults) > 0 {
		resObj.SideBetResults = make([]*controller.BaccaratWebOutputSideBetResult, len(sbResults))
		for i, r := range sbResults {
			resObj.SideBetResults[i] = &controller.BaccaratWebOutputSideBetResult{
				BetType:    r.BetType,
				ResultType: r.ResultType,
				ResultName: r.ResultName,
				BetAmount:  r.BetAmount,
				Payout:     r.Payout,
			}
		}
	} else {
		resObj.SideBetResults = make([]*controller.BaccaratWebOutputSideBetResult, 0)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if b.GetGameEndFlag() {
		switch b.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "baccarat.result.playerWins"
		case domain.GameResultLose:
			resObj.Message = "Banker wins!"
			resObj.MessageCode = "baccarat.result.bankerWins"
		case domain.GameResultDraw:
			resObj.Message = "Tie!"
			resObj.MessageCode = "baccarat.result.tie"
		default:
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (bp *BaccaratWebPresenter) ActionLogOutput(b interfaces.BaccaratGame) string {
	return actionLogOutputJSON(b)
}
