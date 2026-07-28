//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpoonsWebPresenter はスプーンの Web プレゼンター。
type SpoonsWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *SpoonsWebPresenter) Output(g interfaces.SpoonsGame, lastErr error) string {
	resObj := new(controller.SpoonsWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.FeederIdx = g.GetFeederIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.SpoonsRemaining = g.GetSpoonsRemaining()
	resObj.GrabWindowOpen = g.IsGrabWindowOpen()
	resObj.FirstGrabberIdx = g.GetFirstGrabberIdx()
	resObj.RoundLoserIdx = g.GetRoundLoserIdx()
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.DrawPileSize = g.GetDrawPileSize()
	resObj.CpuDifficulty = int(g.GetConfig().CpuDifficulty)

	resObj.Players = make([]*controller.SpoonsWebPlayer, 0, g.GetPlayerCnt())
	for i := range g.GetPlayerCnt() {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := "あなた"
		if !player.GetIsHuman() {
			name = "CPU"
		}
		// 手札は人間 (idx 0) のみ公開する。
		resObj.Players = append(resObj.Players, &controller.SpoonsWebPlayer{
			Name:       name,
			IsHuman:    player.GetIsHuman(),
			HandSize:   player.GetCardsSize(),
			Hand:       playerCardsToOutput(player, player.GetIsHuman()),
			Letters:    player.GetLetters(),
			Eliminated: player.GetEliminated(),
			HasSpoon:   player.GetHasSpoon(),
		})
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildSpoonsMessage(g, lastErr)
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を出力する。
func (p *SpoonsWebPresenter) ActionLogOutput(g interfaces.SpoonsGame) string {
	return actionLogOutputJSON(g)
}

// buildSpoonsMessage はゲーム状態に応じたメッセージを生成する。
func buildSpoonsMessage(g interfaces.SpoonsGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() == 0 {
			return "", "spoons.result.humanWin", nil
		}
		return "", "spoons.result.cpuWin", nil
	}
	return "", "", nil
}
