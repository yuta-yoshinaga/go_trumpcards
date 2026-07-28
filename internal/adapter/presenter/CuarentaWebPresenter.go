//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CuarentaWebPresenter クアレンタ Web プレゼンタークラス。
type CuarentaWebPresenter struct{}

// Output ゲーム状態を JSON 出力。
func (cwp *CuarentaWebPresenter) Output(cg interfaces.CuarentaGame, lastErr error) string {
	resObj := new(controller.CuarentaWebOutput)
	resObj.Players = make([]*controller.CuarentaWebOutputPlayer, 0)
	resObj.TableCards = make([]*controller.WebOutputCard, 0)
	resObj.TeamScores = make([]int, 0)
	resObj.CpuActions = make([]*controller.CuarentaWebOutputAction, 0)
	resObj.RoundWinners = make([]int, 0)

	resObj.CurrentTurn = cg.GetCurrentTurn()
	resObj.LastCaptureIdx = cg.GetLastCaptureIdx()
	resObj.GameEndFlag = cg.GetGameEndFlag()
	resObj.Phase = cg.GetPhase()
	resObj.RemainingDeck = cg.GetRemainingDeck()
	resObj.RoundWinners = append(resObj.RoundWinners, cg.GetRoundWinners()...)

	config := cg.GetConfig()
	resObj.Config = controller.CuarentaWebConfig{
		TargetScore:   config.TargetScore,
		CpuDifficulty: int(config.CpuDifficulty),
	}

	for t := 0; t < domain.CuarentaTeamCnt; t++ {
		resObj.TeamScores = append(resObj.TeamScores, cg.GetTeamScore(t))
	}

	resObj.TableCards = cardsToOutputOrEmpty(cg.GetTableCards())

	for _, a := range cg.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, cuarentaActionToOutput(a))
	}
	if ha := cg.GetHumanAction(); ha != nil {
		resObj.HumanAction = cuarentaActionToOutput(ha)
	}

	for i := 0; i < cg.GetPlayerCnt(); i++ {
		player := cg.GetPlayer(i)
		if player == nil {
			continue
		}
		resObj.Players = append(resObj.Players, &controller.CuarentaWebOutputPlayer{
			ID:            i,
			Team:          domain.CuarentaTeamOf(i),
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			CapturedCount: player.CapturedCount(),
		})
	}

	if det := cg.GetLastRoundDetail(); det != nil {
		resObj.LastRoundDetail = &controller.CuarentaWebOutputScoreDetail{
			CapturedCount: det.CapturedCount,
			Caida:         det.Caida,
			Ronda:         det.Ronda,
			Limpia:        det.Limpia,
			MostCards:     det.MostCards,
			Gained:        det.Gained,
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if cg.GetGameEndFlag() {
		resObj.Message = cwp.buildResultMessage(cg)
		resObj.MessageCode = "cuarenta.result.scores"
		resObj.MessageParams = map[string]string{
			"scores": cwp.encodeScoresParam(cg),
		}
	}

	return marshalOrError(resObj)
}

// encodeScoresParam packs final team scores into a stable, locale-free string
// the frontend can split (e.g. "0:40,1:31").
func (cwp *CuarentaWebPresenter) encodeScoresParam(cg interfaces.CuarentaGame) string {
	parts := make([]string, 0, domain.CuarentaTeamCnt)
	for t := 0; t < domain.CuarentaTeamCnt; t++ {
		parts = append(parts, fmt.Sprintf("%d:%d", t, cg.GetTeamScore(t)))
	}
	return strings.Join(parts, ",")
}

// cuarentaActionToOutput converts a domain action to a WebOutput representation.
func cuarentaActionToOutput(a *domain.CuarentaAction) *controller.CuarentaWebOutputAction {
	if a == nil {
		return nil
	}
	var played *controller.WebOutputCard
	if a.PlayedCard != nil {
		played = cardToOutput(a.PlayedCard)
	}
	return &controller.CuarentaWebOutputAction{
		PlayerIdx:     a.PlayerIdx,
		PlayedCard:    played,
		CapturedCards: cardsToOutput(a.CapturedCards),
		IsCaida:       a.IsCaida,
		IsLimpia:      a.IsLimpia,
		RondaBonus:    a.RondaBonus,
	}
}

// buildResultMessage ゲーム終了時のフォールバック (英語) メッセージ。
func (cwp *CuarentaWebPresenter) buildResultMessage(cg interfaces.CuarentaGame) string {
	msg := "Game over. "
	for t := 0; t < domain.CuarentaTeamCnt; t++ {
		msg += fmt.Sprintf("Team %d:%dpt ", t, cg.GetTeamScore(t))
	}
	return msg
}

// ActionLogOutput 棋譜を JSON 出力。
func (cwp *CuarentaWebPresenter) ActionLogOutput(cg interfaces.CuarentaGame) string {
	return actionLogOutputJSON(cg)
}
