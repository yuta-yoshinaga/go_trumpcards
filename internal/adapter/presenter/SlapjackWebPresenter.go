package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SlapjackWebPresenter スラップジャック Web プレゼンター
type SlapjackWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *SlapjackWebPresenter) Output(g interfaces.SlapjackGame, lastErr error) string {
	resObj := new(controller.SlapjackWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.CurrentTurnIdx = g.GetCurrentTurnIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsTopJack = g.IsTopJack()
	resObj.CenterPileSize = g.GetCenterPileSize()
	resObj.TopCard = cardToOutput(g.GetTopCard())
	resObj.CpuDifficulty = int(g.GetConfig().CpuDifficulty)

	pending := g.GetPending()
	resObj.PendingKind = int(pending.Kind)
	resObj.PendingDeadlineMs = pending.DeadlineMs

	last := g.GetLastEvent()
	resObj.LastEventKind = int(last.Kind)
	resObj.LastEventPlayerIdx = last.PlayerIdx

	resObj.Players = make([]*controller.SlapjackWebPlayer, 0, g.GetPlayerCnt())
	for i := range g.GetPlayerCnt() {
		player := g.GetPlayer(i)
		name := "あなた"
		if !player.GetIsHuman() {
			name = "CPU"
		}
		resObj.Players = append(resObj.Players, &controller.SlapjackWebPlayer{
			Name:      name,
			IsHuman:   player.GetIsHuman(),
			StockSize: player.GetStockSize(),
		})
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildSlapjackMessage(g, lastErr)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜出力
func (p *SlapjackWebPresenter) ActionLogOutput(g interfaces.SlapjackGame) string {
	return actionLogOutputJSON(g)
}

// buildSlapjackMessage ゲーム状態に応じたメッセージを生成する
func buildSlapjackMessage(g interfaces.SlapjackGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() == 0 {
			return "", "slapjack.result.humanWin", nil
		}
		return "", "slapjack.result.cpuWin", nil
	}
	return "", "", nil
}
