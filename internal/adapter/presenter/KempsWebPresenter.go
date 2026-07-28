//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KempsWebPresenter はケムプスの Web プレゼンター。
type KempsWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *KempsWebPresenter) Output(g interfaces.KempsGame, lastErr error) string {
	resObj := new(controller.KempsWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.SignalType = int(g.GetSignalType())
	resObj.PartnerSignaling = g.IsPartnerSignaling()
	resObj.OpponentSignaling = g.IsOpponentSignaling()
	resObj.FourHolderIdx = g.GetFourHolderIdx()
	resObj.RoundResult = g.GetRoundResult()
	resObj.RoundWinnerTeam = g.GetRoundWinnerTeam()
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CpuDifficulty = int(g.GetConfig().CpuDifficulty)
	resObj.TargetScore = g.GetConfig().TargetScore

	resObj.TeamScores = make([]int, domain.KempsTeamCnt)
	for team := range domain.KempsTeamCnt {
		resObj.TeamScores[team] = g.GetTeamScore(team)
	}

	resObj.Field = make([]*controller.WebOutputCard, 0, g.GetFieldSize())
	for i := range g.GetFieldSize() {
		resObj.Field = append(resObj.Field, cardToOutput(g.GetFieldCard(i)))
	}

	resObj.Players = make([]*controller.KempsWebPlayer, 0, g.GetPlayerCnt())
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
		resObj.Players = append(resObj.Players, &controller.KempsWebPlayer{
			Name:           name,
			IsHuman:        player.GetIsHuman(),
			Team:           domain.KempsTeamOf(i),
			HandSize:       player.GetCardsSize(),
			Hand:           playerCardsToOutput(player, player.GetIsHuman()),
			HasFourOfAKind: player.GetIsHuman() && player.HasFourOfAKind(),
		})
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = buildKempsMessage(g, lastErr)
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を出力する。
func (p *KempsWebPresenter) ActionLogOutput(g interfaces.KempsGame) string {
	return actionLogOutputJSON(g)
}

// buildKempsMessage はゲーム状態に応じたメッセージを生成する。
func buildKempsMessage(g interfaces.KempsGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "error", nil
	}
	if g.GetGameEndFlag() {
		if g.GetWinnerTeam() == domain.KempsTeamOf(0) {
			return "", "kemps.result.humanWin", nil
		}
		return "", "kemps.result.cpuWin", nil
	}
	return "", "", nil
}
