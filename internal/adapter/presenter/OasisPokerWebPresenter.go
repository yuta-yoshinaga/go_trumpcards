//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OasisPokerWebPresenter オアシスポーカーWebプレゼンタークラス
type OasisPokerWebPresenter struct {
}

// Output ゲーム状態を出力
func (op *OasisPokerWebPresenter) Output(g interfaces.OasisPokerGame, lastErr error) string {
	resObj := new(controller.OasisPokerWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(g.GetPlayerHand())
	// 終了フェーズ以外はディーラーの2-5枚目を隠す。
	if g.GetPhase() == domain.OasisPokerPhaseEnd {
		resObj.DealerHand = cardsToOutputOrEmpty(g.GetDealerHand())
	} else {
		resObj.DealerHand = oasisPokerMaskDealerHand(g.GetDealerHand())
	}
	resObj.Phase = g.GetPhase()
	resObj.Chips = g.GetChips()
	resObj.AnteBet = g.GetAnteBet()
	resObj.JackpotBet = g.GetJackpotBet()
	resObj.ExchangeCount = g.GetExchangeCount()
	resObj.ExchangeFee = g.GetExchangeFee()
	resObj.PlayBet = g.GetPlayBet()
	resObj.Result = int(g.GetResult())
	resObj.AntePayout = g.GetAntePayout()
	resObj.PlayPayout = g.GetPlayPayout()
	resObj.JackpotPayout = g.GetJackpotPayout()
	resObj.TotalPayout = g.GetTotalPayout()
	resObj.DealerQualified = g.GetDealerQualified()
	resObj.PlayerHandRank = g.GetPlayerHandRank()
	resObj.DealerHandRank = g.GetDealerHandRank()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		switch g.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "oasispoker.result.playerWins"
		case domain.GameResultLose:
			if g.GetPlayBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "oasispoker.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "oasispoker.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "oasispoker.result.push"
		default:
		}
		// ディーラー未クオリファイは結果より優先して表示する（払い戻しが固定のため最重要情報）。
		if !g.GetDealerQualified() && g.GetPlayBet() > 0 {
			resObj.Message = "Dealer does not qualify!"
			resObj.MessageCode = "oasispoker.result.dealerNotQualified"
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (op *OasisPokerWebPresenter) ActionLogOutput(g interfaces.OasisPokerGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (op *OasisPokerWebPresenter) HintOutput(g interfaces.OasisPokerGame) string {
	return op.Output(g, nil)
}

// oasisPokerMaskDealerHand returns the dealer hand with all cards except the first masked.
// 1枚目だけは表示し、それ以降は Design 空文字・Value 0 でマスクする。
func oasisPokerMaskDealerHand(cards []*domain.Card) []*controller.WebOutputCard {
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
