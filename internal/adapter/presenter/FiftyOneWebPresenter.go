package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FiftyOneWebPresenter フィフティワンWebプレゼンタークラス
type FiftyOneWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FiftyOneWebPresenter) Output(fo interfaces.FiftyOneGame, lastErr error) string {
	resObj := new(controller.FiftyOneWebOutput)
	resObj.Phase = int(fo.GetPhase())
	resObj.CurrentTurn = fo.GetCurrentTurn()
	resObj.GameEndFlag = fo.GetGameEndFlag()
	resObj.WinnerIdx = fo.GetWinnerIdx()
	resObj.TurnNumber = fo.GetTurnNumber()
	resObj.StopCallerIdx = fo.GetStopCallerIdx()
	resObj.LastAction = fo.GetLastAction()
	resObj.LastHandIdx = fo.GetLastHandIdx()
	resObj.LastTableIdx = fo.GetLastTableIdx()
	resObj.Config = controller.FiftyOneWebOutputConfig{
		CpuDifficulty: int(fo.GetConfig().CpuDifficulty),
	}

	// 場札
	resObj.TableCards = cardsToOutputOrEmpty(fo.GetTableCards())

	// プレイヤー情報
	resObj.Players = make([]*controller.FiftyOneWebOutputPlayer, 0, fo.GetPlayerCnt())
	for i := 0; i < fo.GetPlayerCnt(); i++ {
		player := fo.GetPlayer(i)
		pObj := &controller.FiftyOneWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, player.GetIsHuman() || fo.GetGameEndFlag()),
			Score:     player.BestSuitScore(),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メッセージ
	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildFiftyOneMessage(fo, lastErr)

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FiftyOneWebPresenter) ActionLogOutput(fo interfaces.FiftyOneGame) string {
	return actionLogOutputJSON(fo)
}

func buildFiftyOneMessage(fo interfaces.FiftyOneGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	if fo.GetGameEndFlag() {
		winnerIdx := fo.GetWinnerIdx()
		winner := fo.GetPlayer(winnerIdx)
		if winner != nil && winner.GetIsHuman() {
			return "ゲーム終了！ あなたの勝ち！", "fiftyone.result.humanWin", nil
		}
		return fmt.Sprintf("ゲーム終了！ CPU %dの勝ち！", winnerIdx),
			"fiftyone.result.cpuWin",
			map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
	}
	return "", "", nil
}
