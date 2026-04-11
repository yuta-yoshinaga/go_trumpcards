package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CaribbeanStudWebPresenter カリビアンスタッドポーカーWebプレゼンタークラス
type CaribbeanStudWebPresenter struct {
}

// Output ゲーム状態を出力
func (cp *CaribbeanStudWebPresenter) Output(cs interfaces.CaribbeanStudGame, lastErr error) string {
	resObj := new(controller.CaribbeanStudWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(cs.GetPlayerHand())
	resObj.DealerHand = cardsToOutputOrEmpty(cs.GetDealerHand())
	resObj.Phase = cs.GetPhase()
	resObj.Chips = cs.GetChips()
	resObj.AnteBet = cs.GetAnteBet()
	resObj.JackpotBet = cs.GetJackpotBet()
	resObj.PlayBet = cs.GetPlayBet()
	resObj.Result = int(cs.GetResult())
	resObj.AntePayout = cs.GetAntePayout()
	resObj.PlayPayout = cs.GetPlayPayout()
	resObj.JackpotPayout = cs.GetJackpotPayout()
	resObj.TotalPayout = cs.GetTotalPayout()
	resObj.DealerQualified = cs.GetDealerQualified()
	resObj.PlayerHandRank = cs.GetPlayerHandRank()
	resObj.DealerHandRank = cs.GetDealerHandRank()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if cs.GetGameEndFlag() {
		switch cs.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "caribbeanstud.result.playerWins"
		case domain.GameResultLose:
			if cs.GetPlayBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "caribbeanstud.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "caribbeanstud.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "caribbeanstud.result.push"
		default:
		}
		if !cs.GetDealerQualified() && cs.GetPlayBet() > 0 {
			resObj.Message = "Dealer does not qualify!"
			resObj.MessageCode = "caribbeanstud.result.dealerNotQualified"
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (cp *CaribbeanStudWebPresenter) ActionLogOutput(cs interfaces.CaribbeanStudGame) string {
	return actionLogOutputJSON(cs)
}
