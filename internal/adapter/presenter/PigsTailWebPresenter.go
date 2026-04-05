package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PigsTailWebPresenter ぶたのしっぽWebプレゼンタークラス
type PigsTailWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pwp *PigsTailWebPresenter) Output(pt interfaces.PigsTailGame, lastErr error) string {
	resObj := new(controller.PigsTailWebOutput)
	resObj.Players = make([]*controller.PigsTailWebOutputPlayer, 0)
	resObj.CircleCount = pt.GetCircleCount()
	resObj.CenterTop = cardToOutput(pt.GetCenterTopCard())
	resObj.CenterCount = len(pt.GetCenter())
	resObj.CurrentTurn = pt.GetCurrentTurn()
	resObj.GameEndFlag = pt.GetGameEndFlag()
	resObj.LoserIdx = pt.GetLoserIdx()
	resObj.LastDrawCard = cardToOutput(pt.GetLastDrawCard())
	resObj.LastPenalty = pt.GetLastPenalty()

	// プレイヤー情報
	for i := 0; i < pt.GetPlayerCnt(); i++ {
		player := pt.GetPlayer(i)
		pObj := &controller.PigsTailWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, player.GetIsHuman()),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.PigsTailWebOutputCpuAction, 0)
	for _, action := range pt.GetCpuActions() {
		a := &controller.PigsTailWebOutputCpuAction{
			DrawPlayerIdx: action.DrawPlayerIdx,
			DrawnCard:     cardToOutput(action.DrawnCard),
			PenaltyFlag:   action.PenaltyFlag,
			PenaltyCount:  action.PenaltyCount,
			HesitationMs:  action.HesitationMs,
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間プレイヤーの行動記録
	if ha := pt.GetHumanAction(); ha != nil {
		resObj.HumanAction = &controller.PigsTailWebOutputCpuAction{
			DrawPlayerIdx: ha.DrawPlayerIdx,
			DrawnCard:     cardToOutput(ha.DrawnCard),
			PenaltyFlag:   ha.PenaltyFlag,
			PenaltyCount:  ha.PenaltyCount,
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if pt.GetGameEndFlag() {
		loserIdx := pt.GetLoserIdx()
		if loserIdx >= 0 {
			loser := pt.GetPlayer(loserIdx)
			if loser != nil && loser.GetIsHuman() {
				resObj.Message = "ゲーム終了！ あなたの負け！"
				resObj.MessageCode = "pigtail.result.humanLose"
			} else {
				resObj.Message = fmt.Sprintf("ゲーム終了！ CPU %dの負け！", loserIdx)
				resObj.MessageCode = "pigtail.result.cpuLose"
				resObj.MessageParams = map[string]string{"cpuId": fmt.Sprintf("%d", loserIdx)}
			}
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pwp *PigsTailWebPresenter) ActionLogOutput(pt interfaces.PigsTailGame) string {
	return actionLogOutputJSON(pt)
}
