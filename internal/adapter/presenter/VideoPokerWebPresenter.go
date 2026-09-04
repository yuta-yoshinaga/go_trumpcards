//go:build !js || !wasm || casino

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// VideoPokerWebPresenter ビデオポーカーWebプレゼンタークラス
type VideoPokerWebPresenter struct {
}

// Output ゲーム状態を出力
func (vpp *VideoPokerWebPresenter) Output(vp interfaces.VideoPokerGame, lastErr error) string {
	resObj := new(controller.VideoPokerWebOutput)

	resObj.Hand = cardsToOutputOrEmpty(vp.GetHand())
	resObj.Phase = vp.GetPhase()
	resObj.Chips = vp.GetChips()
	resObj.BetAmount = vp.GetBetAmount()
	resObj.Result = int(vp.GetResult())
	resObj.Payout = vp.GetPayout()
	resObj.HandRank = vp.GetHandRank()
	resObj.HandName = vp.GetHandName()
	resObj.HandKey = vp.GetHandKey()
	resObj.HeldIndices = vp.GetHeldIndices()
	resObj.VariantName = vp.GetVariantName()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if vp.GetGameEndFlag() {
		switch vp.GetResult() {
		case domain.GameResultWin:
			resObj.Message = vp.GetHandName() + "! You win!"
			resObj.MessageCode = "videopoker.result.win"
			resObj.MessageParams = map[string]string{
				"handName": vp.GetHandName(),
				"payout":   strconv.Itoa(vp.GetPayout()),
			}
		default:
			resObj.Message = "No winning hand."
			resObj.MessageCode = "videopoker.result.lose"
		}
	} else if vp.GetChipsRefilled() {
		// Reset tops the balance back up when it falls under the minimum bet.
		// Without a word for it, chips simply appear and the player is left to
		// guess whether they misread the previous round.
		resObj.Message = "Your balance ran out, so it was topped up."
		resObj.MessageCode = "videopoker.chipsRefilled"
		resObj.MessageParams = map[string]string{"chips": strconv.Itoa(domain.VideoPokerDefaultChips)}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (vpp *VideoPokerWebPresenter) ActionLogOutput(vp interfaces.VideoPokerGame) string {
	return actionLogOutputJSON(vp)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (vpp *VideoPokerWebPresenter) HintOutput(vp interfaces.VideoPokerGame) string {
	return vpp.Output(vp, nil)
}
