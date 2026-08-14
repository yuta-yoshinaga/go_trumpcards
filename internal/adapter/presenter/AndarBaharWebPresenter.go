//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AndarBaharWebPresenter アンダーバハールWebプレゼンタークラス
type AndarBaharWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。** 空の列を `cardsToOutput` に通すと JSON が `null` になり、
// TS 側が非 optional な `Card[]` を約束しているのでページが `.length` で落ちます。
func (ap *AndarBaharWebPresenter) Output(ab interfaces.AndarBaharGame, lastErr error) string {
	resObj := new(controller.AndarBaharWebOutput)

	resObj.Joker = cardToOutput(ab.GetJoker())
	resObj.AndarCards = cardsToOutputOrEmpty(ab.GetAndarCards())
	resObj.BaharCards = cardsToOutputOrEmpty(ab.GetBaharCards())
	resObj.FirstColumn = ab.GetFirstColumn()
	resObj.DealtCount = ab.DealtCount()
	resObj.Phase = ab.GetPhase()
	resObj.Chips = ab.GetChips()
	resObj.BetAmount = ab.GetBetAmount()
	resObj.BetTarget = ab.GetBetTarget()
	resObj.SideAmount = ab.GetSideAmount()
	resObj.SideBand = ab.GetSideBand()
	resObj.Winner = ab.GetWinner()
	resObj.Result = int(ab.GetResult())
	resObj.Payout = ab.GetPayout()
	resObj.History = intSliceOrEmpty(ab.GetHistory())

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if ab.GetGameEndFlag() {
		resObj.Message, resObj.MessageCode = andarBaharEndMessage(ab)
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (ap *AndarBaharWebPresenter) ActionLogOutput(ab interfaces.AndarBaharGame) string {
	return actionLogOutputJSON(ab)
}

// HintOutput ヒントをJSON出力
func (ap *AndarBaharWebPresenter) HintOutput(ab interfaces.AndarBaharGame) string {
	return marshalOrError(map[string]string{"hint": ab.GetHint()})
}

// andarBaharEndMessage は決着時の表示メッセージと i18n キーを返す。
func andarBaharEndMessage(ab interfaces.AndarBaharGame) (string, string) {
	if ab.GetResult() == domain.GameResultWin {
		return "", "andarbahar.result.win"
	}
	return "", "andarbahar.result.lose"
}
