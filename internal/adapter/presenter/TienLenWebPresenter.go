//go:build !js || !wasm || extra9

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TienLenWebPresenter Tien Len Webプレゼンタークラス
type TienLenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TienLenWebPresenter) Output(tg interfaces.TienLenGame, lastErr error) string {
	resObj := new(controller.TienLenWebOutput)
	resObj.Players = make([]*controller.TienLenWebOutputPlayer, 0)
	resObj.CurrentTurn = tg.GetCurrentTurn()
	resObj.LastPlayPlayerIdx = tg.GetLastPlayPlayerIdx()
	resObj.GameEndFlag = tg.GetGameEndFlag()
	resObj.TablePlayType = int(tg.GetTablePlayType())

	config := tg.GetConfig()
	resObj.Config = controller.TienLenWebConfig{
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.TableCards = cardsToOutputOrEmpty(tg.GetTableCards())

	resObj.CpuActions = make([]*controller.TienLenWebOutputAction, 0)
	for _, action := range tg.GetCpuActions() {
		a := &controller.TienLenWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: cardsToOutput(action.PlayedCards),
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	humanAction := tg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.TienLenWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: cardsToOutput(humanAction.PlayedCards),
		}
	}

	for i := 0; i < tg.GetPlayerCnt(); i++ {
		player := tg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := new(controller.TienLenWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.Cards = playerCardsToOutput(player, player.GetIsHuman())
		resObj.Players = append(resObj.Players, pObj)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if tg.GetGameEndFlag() {
		resObj.Rankings = make([]controller.WebOutputRanking, 0)
		for i := 0; i < tg.GetPlayerCnt(); i++ {
			player := tg.GetPlayer(i)
			if player != nil && player.GetRank() >= 1 && player.GetRank() <= 4 {
				resObj.Rankings = append(resObj.Rankings, controller.WebOutputRanking{Seat: i, Rank: player.GetRank()})
			}
		}
		resObj.MessageCode = "tienlen.result.rankings"
		resObj.MessageParams = map[string]string{}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TienLenWebPresenter) ActionLogOutput(tg interfaces.TienLenGame) string {
	return actionLogOutputJSON(tg)
}

// HintOutput はヒント専用のレスポンスを持たないので通常の状態出力を返す。
// Web のヒントは `useGameHint` がクライアント側で出しており、CUI (#5624) が
// 使うドメインのヒントとは別経路。
func (p *TienLenWebPresenter) HintOutput(tg interfaces.TienLenGame) string {
	return p.Output(tg, nil)
}
