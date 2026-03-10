package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BaccaratWebPresenter バカラWebプレゼンタークラス
type BaccaratWebPresenter struct {
}

// NewBaccaratWebPresenter コンストラクタ
func NewBaccaratWebPresenter() *BaccaratWebPresenter {
	return &BaccaratWebPresenter{}
}

// Output ゲーム状態を出力
func (bp *BaccaratWebPresenter) Output(b interfaces.BaccaratGame, lastErr error) string {
	resObj := new(controller.BaccaratWebOutput)

	// プレイヤーハンド
	resObj.PlayerHand = make([]*controller.WebOutputCard, 0)
	for _, card := range b.GetPlayerHand() {
		resObj.PlayerHand = append(resObj.PlayerHand, cardToOutput(card))
	}

	// バンカーハンド
	resObj.BankerHand = make([]*controller.WebOutputCard, 0)
	for _, card := range b.GetBankerHand() {
		resObj.BankerHand = append(resObj.BankerHand, cardToOutput(card))
	}

	resObj.PlayerHandValue = b.GetPlayerHandValue()
	resObj.BankerHandValue = b.GetBankerHandValue()
	resObj.Phase = b.GetPhase()
	resObj.Chips = b.GetChips()
	resObj.BetAmount = b.GetBetAmount()
	resObj.BetType = b.GetBetType()
	resObj.Result = int(b.GetResult())
	resObj.Payout = b.GetPayout()

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
	if !b.GetGameEndFlag() {
		return actionLogToJSON(nil)
	}
	return actionLogToJSON(b.GetActionLog())
}
