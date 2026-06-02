package presenter

import (
	"fmt"

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
		resObj.Message = p.buildResultMessage(tg)
		resObj.MessageCode = "tienlen.result.rankings"
		resObj.MessageParams = map[string]string{"rankings": resObj.Message}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (p *TienLenWebPresenter) buildResultMessage(tg interfaces.TienLenGame) string {
	msg := "ゲーム終了！ "
	for i := 0; i < tg.GetPlayerCnt(); i++ {
		player := tg.GetPlayer(i)
		if player == nil {
			continue
		}
		rank := player.GetRank()
		if rank < 1 || rank > 4 {
			continue
		}
		var name string
		if player.GetIsHuman() {
			name = "あなた"
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		msg += fmt.Sprintf("%s:%d位 ", name, rank)
	}
	return msg
}

// ActionLogOutput 棋譜をJSON出力
func (p *TienLenWebPresenter) ActionLogOutput(tg interfaces.TienLenGame) string {
	return actionLogOutputJSON(tg)
}
