//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HighCardFlushWebPresenter ハイカードフラッシュWebプレゼンタークラス
type HighCardFlushWebPresenter struct {
}

// Output ゲーム状態を出力
func (hp *HighCardFlushWebPresenter) Output(hcf interfaces.HighCardFlushGame, lastErr error) string {
	resObj := new(controller.HighCardFlushWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(hcf.GetPlayerHand())
	// The dealer's 7 cards are dealt at the same time as the player's but stay
	// hidden until showdown. Only expose them at the END phase — otherwise the
	// web API would leak the dealer's hand (a cheat, and, via the UI, an
	// accessibility over-share) during betting/action. Mirrors the CUI gate.
	if hcf.GetPhase() == domain.HighCardFlushPhaseEnd {
		resObj.DealerHand = cardsToOutputOrEmpty(hcf.GetDealerHand())
	} else {
		resObj.DealerHand = cardsToOutputOrEmpty(nil)
	}
	resObj.Phase = hcf.GetPhase()
	resObj.Chips = hcf.GetChips()
	resObj.AnteBet = hcf.GetAnteBet()
	resObj.FlushBonusBet = hcf.GetFlushBonusBet()
	resObj.StraightFlushBet = hcf.GetStraightFlushBet()
	resObj.RaiseBet = hcf.GetRaiseBet()
	resObj.Result = int(hcf.GetResult())
	resObj.AntePayout = hcf.GetAntePayout()
	resObj.RaisePayout = hcf.GetRaisePayout()
	resObj.FlushBonusPayout = hcf.GetFlushBonusPayout()
	resObj.StraightFlushPayout = hcf.GetStraightFlushPayout()
	resObj.TotalPayout = hcf.GetTotalPayout()
	resObj.DealerQualified = hcf.GetDealerQualified()
	resObj.PlayerFlushLen = hcf.GetPlayerFlushLen()
	resObj.DealerFlushLen = hcf.GetDealerFlushLen()
	resObj.PlayerStraightFlushLen = hcf.GetPlayerStraightFlushLen()
	resObj.MaxRaiseMultiplier = hcf.MaxRaiseMultiplier()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if hcf.GetGameEndFlag() {
		switch hcf.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "highcardflush.result.playerWins"
		case domain.GameResultLose:
			if hcf.GetRaiseBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "highcardflush.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "highcardflush.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "highcardflush.result.push"
		default:
		}
		if !hcf.GetDealerQualified() && hcf.GetRaiseBet() > 0 {
			resObj.Message = "Dealer does not qualify!"
			resObj.MessageCode = "highcardflush.result.dealerNotQualified"
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (hp *HighCardFlushWebPresenter) ActionLogOutput(hcf interfaces.HighCardFlushGame) string {
	return actionLogOutputJSON(hcf)
}
