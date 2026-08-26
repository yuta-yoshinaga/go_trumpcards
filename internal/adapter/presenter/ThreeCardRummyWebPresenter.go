//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ThreeCardRummyWebPresenter スリーカード・ラミーWebプレゼンタークラス
type ThreeCardRummyWebPresenter struct {
}

// Output ゲーム状態を出力
func (tp *ThreeCardRummyWebPresenter) Output(tc interfaces.ThreeCardRummyGame, lastErr error) string {
	resObj := new(controller.ThreeCardRummyWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(tc.GetPlayerHand())
	// **勝負するまでディーラーの手は伏せる。** 3 枚とも見えていたら相手の合計が
	// 数えられてしまい、play/fold に判断の余地が無くなる (CUI 側も終了フェーズ
	// まで出さない)。
	if tc.GetPhase() == domain.ThreeCardRummyPhaseEnd {
		resObj.DealerHand = cardsToOutputOrEmpty(tc.GetDealerHand())
	} else {
		resObj.DealerHand = threeCardRummyMaskHand(tc.GetDealerHand())
	}
	resObj.Phase = tc.GetPhase()
	resObj.Chips = tc.GetChips()
	resObj.AnteBet = tc.GetAnteBet()
	resObj.LowBonusBet = tc.GetLowBonusBet()
	resObj.PlayBet = tc.GetPlayBet()
	resObj.Result = int(tc.GetResult())
	resObj.AntePayout = tc.GetAntePayout()
	resObj.PlayPayout = tc.GetPlayPayout()
	resObj.AnteBonusPayout = tc.GetAnteBonusPayout()
	resObj.LowBonusPayout = tc.GetLowBonusPayout()
	resObj.TotalPayout = tc.GetTotalPayout()
	resObj.DealerQualified = tc.GetDealerQualified()
	resObj.PlayerScore = tc.GetPlayerScore()
	resObj.DealerScore = tc.GetDealerScore()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if tc.GetGameEndFlag() {
		switch tc.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "threecardrummy.result.playerWins"
		case domain.GameResultLose:
			if tc.GetPlayBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "threecardrummy.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "threecardrummy.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "threecardrummy.result.push"
		default:
		}
		if !tc.GetDealerQualified() && tc.GetPlayBet() > 0 {
			resObj.Message = "Dealer does not qualify!"
			resObj.MessageCode = "threecardrummy.result.dealerNotQualified"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput returns the current game state as JSON. The Web GUI computes its
// own play/fold hint client-side (useGameHint), so this simply mirrors Output;
// it exists to satisfy the ThreeCardRummyPresenter interface shared with the CUI.
func (tp *ThreeCardRummyWebPresenter) HintOutput(tc interfaces.ThreeCardRummyGame) string {
	return tp.Output(tc, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (tp *ThreeCardRummyWebPresenter) ActionLogOutput(tc interfaces.ThreeCardRummyGame) string {
	return actionLogOutputJSON(tc)
}

// threeCardRummyMaskHand replaces every card with the blank the Web GUI renders
// as a card back. Unlike Caribbean Stud there is no up-card: the player decides
// on their own total alone.
func threeCardRummyMaskHand(cards []*domain.Card) []*controller.WebOutputCard {
	masked := make([]*controller.WebOutputCard, len(cards))
	for i := range cards {
		masked[i] = &controller.WebOutputCard{Design: "", Value: 0}
	}
	return masked
}
