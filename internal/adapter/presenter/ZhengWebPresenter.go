//go:build !js || !wasm || extra9

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ZhengWebPresenter 争上游 Webプレゼンタークラス
type ZhengWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ZhengWebPresenter) Output(zg interfaces.ZhengGame, lastErr error) string {
	resObj := new(controller.ZhengWebOutput)
	resObj.Players = make([]*controller.ZhengWebOutputPlayer, 0)
	resObj.CurrentTurn = zg.GetCurrentTurn()
	resObj.LastPlayPlayerIdx = zg.GetLastPlayPlayerIdx()
	resObj.GameEndFlag = zg.GetGameEndFlag()
	resObj.TablePlayType = int(zg.GetTablePlayType())

	config := zg.GetConfig()
	resObj.Config = controller.ZhengWebConfig{
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.TableCards = cardsToOutputOrEmpty(zg.GetTableCards())

	resObj.CpuActions = make([]*controller.ZhengWebOutputAction, 0)
	for _, action := range zg.GetCpuActions() {
		a := &controller.ZhengWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: cardsToOutput(action.PlayedCards),
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	humanAction := zg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.ZhengWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: cardsToOutput(humanAction.PlayedCards),
		}
	}

	for i := 0; i < zg.GetPlayerCnt(); i++ {
		player := zg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := new(controller.ZhengWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.Cards = playerCardsToOutput(player, player.GetIsHuman())
		pObj.FinishCards = cardsToOutputOrEmpty(nil)
		if player.GetIsFinished() {
			pObj.FinishCards = cardsToOutputOrEmpty(player.GetFinishCards())
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if zg.GetGameEndFlag() {
		resObj.Rankings = make([]controller.WebOutputRanking, 0)
		for i := 0; i < zg.GetPlayerCnt(); i++ {
			player := zg.GetPlayer(i)
			if player != nil && player.GetRank() >= 1 && player.GetRank() <= 4 {
				resObj.Rankings = append(resObj.Rankings, controller.WebOutputRanking{Seat: i, Rank: player.GetRank()})
			}
		}
		resObj.MessageCode = "zheng.result.rankings"
		resObj.MessageParams = map[string]string{}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *ZhengWebPresenter) ActionLogOutput(zg interfaces.ZhengGame) string {
	return actionLogOutputJSON(zg)
}

// HintOutput は Web ではヒント専用の応答を持たないので通常の状態を返す。
func (p *ZhengWebPresenter) HintOutput(zg interfaces.ZhengGame) string {
	return p.Output(zg, nil)
}
