//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ThreeCardWebPresenter スリーカードポーカーWebプレゼンタークラス
type ThreeCardWebPresenter struct {
}

// Output ゲーム状態を出力
func (tp *ThreeCardWebPresenter) Output(tc interfaces.ThreeCardGame, lastErr error) string {
	resObj := new(controller.ThreeCardWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(tc.GetPlayerHand())
	resObj.DealerHand = cardsToOutputOrEmpty(tc.GetDealerHand())
	resObj.Phase = tc.GetPhase()
	resObj.Chips = tc.GetChips()
	resObj.AnteBet = tc.GetAnteBet()
	resObj.PairPlusBet = tc.GetPairPlusBet()
	resObj.PlayBet = tc.GetPlayBet()
	resObj.Result = int(tc.GetResult())
	resObj.AntePayout = tc.GetAntePayout()
	resObj.PlayPayout = tc.GetPlayPayout()
	resObj.AnteBonusPayout = tc.GetAnteBonusPayout()
	resObj.PairPlusPayout = tc.GetPairPlusPayout()
	resObj.TotalPayout = tc.GetTotalPayout()
	resObj.DealerQualified = tc.GetDealerQualified()
	resObj.PlayerHandRank = tc.GetPlayerHandRank()
	resObj.DealerHandRank = tc.GetDealerHandRank()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if tc.GetGameEndFlag() {
		switch tc.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "threecard.result.playerWins"
		case domain.GameResultLose:
			if tc.GetPlayBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "threecard.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "threecard.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "threecard.result.push"
		default:
		}
		if !tc.GetDealerQualified() && tc.GetPlayBet() > 0 {
			resObj.Message = "Dealer does not qualify!"
			resObj.MessageCode = "threecard.result.dealerNotQualified"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput returns the current game state as JSON. The Web GUI computes its
// own play/fold hint client-side (useGameHint), so this simply mirrors Output;
// it exists to satisfy the ThreeCardPresenter interface shared with the CUI.
func (tp *ThreeCardWebPresenter) HintOutput(tc interfaces.ThreeCardGame) string {
	return tp.Output(tc, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (tp *ThreeCardWebPresenter) ActionLogOutput(tc interfaces.ThreeCardGame) string {
	return actionLogOutputJSON(tc)
}
