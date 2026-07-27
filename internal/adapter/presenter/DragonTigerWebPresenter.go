//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DragonTigerWebPresenter ドラゴンタイガーWebプレゼンタークラス
type DragonTigerWebPresenter struct{}

// Output ゲーム状態を出力
func (dp *DragonTigerWebPresenter) Output(dt interfaces.DragonTigerGame, lastErr error) string {
	resObj := new(controller.DragonTigerWebOutput)

	if dt.GetDragonCard() != nil {
		resObj.DragonCard = cardToOutput(dt.GetDragonCard())
	}
	if dt.GetTigerCard() != nil {
		resObj.TigerCard = cardToOutput(dt.GetTigerCard())
	}
	resObj.Phase = dt.GetPhase()
	resObj.Chips = dt.GetChips()
	resObj.BetAmount = dt.GetBetAmount()
	resObj.BetType = dt.GetBetType()
	resObj.Result = int(dt.GetResult())
	resObj.Payout = dt.GetPayout()
	if h := dt.GetHistory(); h != nil {
		resObj.History = h
	} else {
		resObj.History = make([]int, 0)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if dt.GetGameEndFlag() {
		resObj.Message, resObj.MessageCode = dragonTigerEndMessage(dt)
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (dp *DragonTigerWebPresenter) ActionLogOutput(dt interfaces.DragonTigerGame) string {
	return actionLogOutputJSON(dt)
}

// dragonTigerEndMessage は終了時の表示メッセージと i18n キーを返す。
func dragonTigerEndMessage(dt interfaces.DragonTigerGame) (string, string) {
	switch dt.GetResult() {
	case domain.GameResultWin: // dragon wins
		return "", "dragontiger.result.dragonWins"
	case domain.GameResultLose: // tiger wins
		return "", "dragontiger.result.tigerWins"
	case domain.GameResultDraw:
		if dt.GetBetType() == domain.DragonTigerBetTie {
			return "", "dragontiger.result.tieWin"
		}
		return "", "dragontiger.result.tieRefund"
	default:
		return "", ""
	}
}
