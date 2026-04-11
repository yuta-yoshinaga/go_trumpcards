package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WarWebPresenter 戦争Webプレゼンタークラス
type WarWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WarWebPresenter) Output(w interfaces.WarGame, lastErr error) string {
	resObj := new(controller.WarWebOutput)
	resObj.Phase = int(w.GetPhase())
	resObj.GameEndFlag = w.GetGameEndFlag()
	resObj.WinnerIdx = w.GetWinnerIdx()
	resObj.PlayerRevealed = cardToOutput(w.GetPlayerRevealed())
	resObj.CpuRevealed = cardToOutput(w.GetCpuRevealed())
	resObj.WarPotSize = w.GetWarPotSize()
	resObj.LastWinnerIdx = w.GetLastWinnerIdx()
	resObj.LastBurialCount = w.GetLastBurialCount()
	resObj.RoundsPlayed = w.GetRoundsPlayed()
	resObj.Config = controller.WarWebOutputConfig{
		MaxRounds: w.GetConfig().MaxRounds,
	}

	resObj.Players = make([]*controller.WarWebOutputPlayer, 0, w.GetPlayerCnt())
	for i := range w.GetPlayerCnt() {
		player := w.GetPlayer(i)
		resObj.Players = append(resObj.Players, &controller.WarWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			DrawPileSize:    player.GetDrawPileSize(),
			DiscardPileSize: player.GetDiscardPileSize(),
			TotalCards:      player.TotalCards(),
		})
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildWarMessage(w, lastErr)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜出力
func (p *WarWebPresenter) ActionLogOutput(w interfaces.WarGame) string {
	return actionLogOutputJSON(w)
}

// buildWarMessage ゲーム状態に応じたメッセージを生成する
func buildWarMessage(w interfaces.WarGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	if w.GetGameEndFlag() {
		return "", "gameEnd", nil
	}
	switch w.GetPhase() {
	case domain.WarPhaseWarBury:
		return "", "war", nil
	case domain.WarPhaseResolved:
		return "", "resolved", nil
	}
	return "", "reveal", nil
}
