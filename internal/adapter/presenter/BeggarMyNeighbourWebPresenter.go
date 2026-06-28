package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BeggarMyNeighbourWebPresenter Beggar-My-Neighbour Webプレゼンタークラス
type BeggarMyNeighbourWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BeggarMyNeighbourWebPresenter) Output(g interfaces.BeggarMyNeighbourGame, lastErr error) string {
	resObj := new(controller.BeggarMyNeighbourWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.PenaltyOwnerIdx = g.GetPenaltyOwnerIdx()
	resObj.PenaltyRemaining = g.GetPenaltyRemaining()
	resObj.CentralPileSize = g.GetCentralPileSize()
	resObj.LastCardPlayed = cardToOutput(g.GetLastCardPlayed())
	resObj.RoundsPlayed = g.GetRoundsPlayed()
	resObj.Config = controller.BeggarMyNeighbourWebOutputConfig{
		MaxRounds: g.GetConfig().MaxRounds,
	}

	resObj.Players = make([]*controller.BeggarMyNeighbourWebOutputPlayer, 0, g.GetPlayerCnt())
	for i := range g.GetPlayerCnt() {
		player := g.GetPlayer(i)
		resObj.Players = append(resObj.Players, &controller.BeggarMyNeighbourWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			DrawPileSize:    player.GetDrawPileSize(),
			DiscardPileSize: player.GetDiscardPileSize(),
			TotalCards:      player.TotalCards(),
		})
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildBeggarMyNeighbourMessage(g, lastErr)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜出力
func (p *BeggarMyNeighbourWebPresenter) ActionLogOutput(g interfaces.BeggarMyNeighbourGame) string {
	return actionLogOutputJSON(g)
}

// buildBeggarMyNeighbourMessage ゲーム状態に応じたメッセージを生成する
func buildBeggarMyNeighbourMessage(g interfaces.BeggarMyNeighbourGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	if g.GetGameEndFlag() {
		switch g.GetWinnerIdx() {
		case 0:
			return "", "beggarmyneighbour.result.humanWin", nil
		case 1:
			return "", "beggarmyneighbour.result.cpuWin", nil
		default:
			return "", "beggarmyneighbour.result.draw", nil
		}
	}
	return "", "", nil
}
