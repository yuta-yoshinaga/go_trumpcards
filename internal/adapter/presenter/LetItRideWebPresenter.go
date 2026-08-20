//go:build !js || !wasm || extra4

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// LetItRideWebPresenter レット・イット・ライドWebプレゼンタークラス
type LetItRideWebPresenter struct {
}

// Output ゲーム状態を出力
func (lp *LetItRideWebPresenter) Output(lir interfaces.LetItRideGame, lastErr error) string {
	resObj := new(controller.LetItRideWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(lir.GetPlayerHand())
	resObj.CommunityCards = letItRideMaskCommunity(lir)
	resObj.Phase = lir.GetPhase()
	resObj.Chips = lir.GetChips()
	resObj.BetAmount = lir.GetBetAmount()
	// API convention: Bet 1 is pulled first (first decision), Bet 3 is always active.
	// Domain convention is reversed (bet1Active = always active, bet3Active = pulled first),
	// so we swap bet1↔bet3 here to produce the correct wire format.
	resObj.Bet1Active = lir.GetBet3Active()
	resObj.Bet2Active = lir.GetBet2Active()
	resObj.Bet3Active = lir.GetBet1Active()
	resObj.Result = int(lir.GetResult())
	resObj.HandRank = lir.GetHandRank()
	resObj.Bet1Payout = lir.GetBet3Payout()
	resObj.Bet2Payout = lir.GetBet2Payout()
	resObj.Bet3Payout = lir.GetBet1Payout()
	resObj.TotalPayout = lir.GetTotalPayout()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if lir.GetGameEndFlag() {
		switch lir.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "letitride.result.playerWins"
		case domain.GameResultLose:
			resObj.Message = "Player loses."
			resObj.MessageCode = "letitride.result.playerLoses"
		default:
		}
	}

	return marshalOrError(resObj)
}

// PullConfirmOutput は Pull 実行前の確認内容を JSON 出力する。
// Web は自前のダイアログで確認するため、ここは状態をそのまま返す。
func (lp *LetItRideWebPresenter) PullConfirmOutput(lir interfaces.LetItRideGame) string {
	return lp.Output(lir, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (lp *LetItRideWebPresenter) ActionLogOutput(lir interfaces.LetItRideGame) string {
	return actionLogOutputJSON(lir)
}

// letItRideMaskCommunity returns community cards with appropriate masking.
// BET and FIRST_DECISION phases: both masked. SECOND_DECISION: first revealed, second masked.
// END: both revealed.
func letItRideMaskCommunity(lir interfaces.LetItRideGame) []*controller.WebOutputCard {
	cards := lir.GetCommunityCards()
	if len(cards) == 0 {
		return make([]*controller.WebOutputCard, 0)
	}

	phase := lir.GetPhase()
	result := make([]*controller.WebOutputCard, len(cards))

	switch phase {
	case domain.LetItRidePhaseBet:
		// Both masked
		for i := range cards {
			result[i] = &controller.WebOutputCard{Design: "", Value: 0}
		}
	case domain.LetItRidePhaseFirstDecision:
		// Both masked (community cards are revealed AFTER the decision)
		for i := range cards {
			result[i] = &controller.WebOutputCard{Design: "", Value: 0}
		}
	case domain.LetItRidePhaseSecondDecision:
		// First revealed, second masked
		result[0] = cardToOutput(cards[0])
		for i := 1; i < len(cards); i++ {
			result[i] = &controller.WebOutputCard{Design: "", Value: 0}
		}
	default:
		// END: all revealed
		for i, c := range cards {
			result[i] = cardToOutput(c)
		}
	}

	return result
}
