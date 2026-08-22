//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CaribbeanDrawWebPresenter カリビアン・ドロー・ポーカーWebプレゼンタークラス
type CaribbeanDrawWebPresenter struct {
}

// Output ゲーム状態を出力
func (cp *CaribbeanDrawWebPresenter) Output(cs interfaces.CaribbeanDrawGame, lastErr error) string {
	resObj := new(controller.CaribbeanDrawWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(cs.GetPlayerHand())
	// During the action phase only the first dealer card is visible; mask the rest.
	if cs.GetPhase() == domain.CaribbeanDrawPhaseEnd {
		resObj.DealerHand = cardsToOutputOrEmpty(cs.GetDealerHand())
	} else {
		resObj.DealerHand = caribbeanDrawMaskDealerHand(cs.GetDealerHand())
	}
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
	resObj.DrawCost = cs.GetDrawCost()
	resObj.PlayerHandRank = cs.GetPlayerHandRank()
	resObj.DealerHandRank = cs.GetDealerHandRank()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if cs.GetGameEndFlag() {
		switch cs.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "caribbeandraw.result.playerWins"
		case domain.GameResultLose:
			if cs.GetPlayBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "caribbeandraw.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "caribbeandraw.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "caribbeandraw.result.push"
		default:
		}
		// Dealer-not-qualified intentionally overrides any earlier win/lose/push message:
		// the qualification status is the most useful piece of information for the player
		// in that scenario (ante pays 1:1, play bet pushes regardless of hand strength).
		if !cs.GetDealerQualified() && cs.GetPlayBet() > 0 {
			resObj.Message = "Dealer does not qualify!"
			resObj.MessageCode = "caribbeandraw.result.dealerNotQualified"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。CaribbeanDrawPresenter インタフェースを
// 満たすための実装。
func (cp *CaribbeanDrawWebPresenter) HintOutput(cs interfaces.CaribbeanDrawGame) string {
	return cp.Output(cs, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (cp *CaribbeanDrawWebPresenter) ActionLogOutput(cs interfaces.CaribbeanDrawGame) string {
	return actionLogOutputJSON(cs)
}

// caribbeanDrawMaskDealerHand returns the dealer hand with all cards except the first masked.
// Only the first card is revealed during the action phase; the rest have empty Design and zero Value.
func caribbeanDrawMaskDealerHand(cards []*domain.Card) []*controller.WebOutputCard {
	if len(cards) == 0 {
		return make([]*controller.WebOutputCard, 0)
	}
	result := make([]*controller.WebOutputCard, len(cards))
	result[0] = cardToOutput(cards[0])
	for i := 1; i < len(cards); i++ {
		result[i] = &controller.WebOutputCard{Design: "", Value: 0}
	}
	return result
}
