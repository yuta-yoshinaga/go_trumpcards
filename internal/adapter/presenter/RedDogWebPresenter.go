//go:build !js || !wasm || extra4

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RedDogWebPresenter レッドドッグWebプレゼンタークラス
type RedDogWebPresenter struct{}

func redDogHasPair(initial []*domain.Card) bool {
	return len(initial) == 2 && initial[0].GetValue() == initial[1].GetValue()
}

// Output ゲーム状態を出力
func (rp *RedDogWebPresenter) Output(rd interfaces.RedDogGame, lastErr error) string {
	resObj := new(controller.RedDogWebOutput)

	resObj.InitialCards = cardsToOutputOrEmpty(rd.GetInitialCards())
	if rd.GetThirdCard() != nil {
		resObj.ThirdCard = cardToOutput(rd.GetThirdCard())
	}
	resObj.Phase = rd.GetPhase()
	resObj.Chips = rd.GetChips()
	resObj.Ante = rd.GetAnte()
	resObj.Raise = rd.GetRaise()
	resObj.Spread = rd.GetSpread()
	resObj.Result = int(rd.GetResult())
	resObj.TotalPayout = rd.GetTotalPayout()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if rd.GetGameEndFlag() {
		pair := redDogHasPair(rd.GetInitialCards())
		switch rd.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "reddog.result.playerWins"
			if pair {
				resObj.MessageCode = "reddog.result.pairWins"
			}
		case domain.GameResultLose:
			resObj.Message = "Player loses."
			resObj.MessageCode = "reddog.result.playerLoses"
			if pair {
				resObj.MessageCode = "reddog.result.pairLoses"
			}
		default:
			resObj.Message = "Push."
			resObj.MessageCode = "reddog.result.push"
			if pair {
				resObj.MessageCode = "reddog.result.pairPush"
			}
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (rp *RedDogWebPresenter) ActionLogOutput(rd interfaces.RedDogGame) string {
	return actionLogOutputJSON(rd)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。RedDogPresenter インタフェースを満たすための実装。
func (rp *RedDogWebPresenter) HintOutput(rd interfaces.RedDogGame) string {
	return rp.Output(rd, nil)
}
