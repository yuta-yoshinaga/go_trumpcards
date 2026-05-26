package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BigTwoWebPresenter Big Two Webプレゼンタークラス
type BigTwoWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BigTwoWebPresenter) Output(bg interfaces.BigTwoGame, lastErr error) string {
	resObj := new(controller.BigTwoWebOutput)
	resObj.Players = make([]*controller.BigTwoWebOutputPlayer, 0)
	resObj.CurrentTurn = bg.GetCurrentTurn()
	resObj.LastPlayPlayerIdx = bg.GetLastPlayPlayerIdx()
	resObj.GameEndFlag = bg.GetGameEndFlag()
	resObj.TablePlayType = int(bg.GetTablePlayType())

	config := bg.GetConfig()
	resObj.Config = controller.BigTwoWebConfig{
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.TableCards = cardsToOutputOrEmpty(bg.GetTableCards())

	resObj.CpuActions = make([]*controller.BigTwoWebOutputAction, 0)
	for _, action := range bg.GetCpuActions() {
		a := &controller.BigTwoWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: cardsToOutput(action.PlayedCards),
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	humanAction := bg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.BigTwoWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: cardsToOutput(humanAction.PlayedCards),
		}
	}

	for i := 0; i < bg.GetPlayerCnt(); i++ {
		player := bg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := new(controller.BigTwoWebOutputPlayer)
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
	} else if bg.GetGameEndFlag() {
		resObj.Message = p.buildResultMessage(bg)
		resObj.MessageCode = "bigtwo.result.rankings"
		resObj.MessageParams = map[string]string{"rankings": resObj.Message}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (p *BigTwoWebPresenter) buildResultMessage(bg interfaces.BigTwoGame) string {
	msg := "ゲーム終了！ "
	for i := 0; i < bg.GetPlayerCnt(); i++ {
		player := bg.GetPlayer(i)
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
func (p *BigTwoWebPresenter) ActionLogOutput(bg interfaces.BigTwoGame) string {
	return actionLogOutputJSON(bg)
}
