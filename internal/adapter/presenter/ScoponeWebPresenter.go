//go:build !js || !wasm || classic

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ScoponeWebPresenter スコポーネ Web プレゼンタークラス。
type ScoponeWebPresenter struct{}

// Output ゲーム状態を JSON 出力。
func (swp *ScoponeWebPresenter) Output(sg interfaces.ScoponeGame, lastErr error) string {
	resObj := new(controller.ScoponeWebOutput)
	resObj.Players = make([]*controller.ScoponeWebOutputPlayer, 0)
	resObj.TableCards = make([]*controller.WebOutputCard, 0)
	resObj.HandCaptures = make([][][]int, 0)

	resObj.Phase = sg.GetPhase()
	resObj.RoundNumber = sg.GetRoundNumber()
	resObj.CurrentTurn = sg.GetCurrentTurn()
	resObj.DealerIdx = sg.GetDealerIdx()
	resObj.LastCaptureIdx = sg.GetLastCaptureIdx()
	resObj.WinnerTeam = sg.GetWinnerTeam()
	resObj.GameEndFlag = sg.GetGameEndFlag()
	resObj.IsHumanTurn = sg.IsHumanTurn()

	resObj.TeamScores = make([]int, domain.ScoponeTeamCnt)
	for t := 0; t < domain.ScoponeTeamCnt; t++ {
		resObj.TeamScores[t] = sg.GetTeamScore(t)
	}

	config := sg.GetConfig()
	resObj.Config = controller.ScoponeWebConfig{
		TargetScore:   config.TargetScore,
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.TableCards = cardsToOutputOrEmpty(sg.GetTableCards())

	humanIdx := -1
	for i := 0; i < sg.GetPlayerCnt(); i++ {
		player := sg.GetPlayer(i)
		if player == nil {
			continue
		}
		if player.GetIsHuman() {
			humanIdx = i
		}
		resObj.Players = append(resObj.Players, &controller.ScoponeWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Team:          domain.ScoponeTeamOf(i),
			HandCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			CapturedCount: player.CapturedCount(),
			ScopaCount:    player.GetScopaCount(),
		})
	}

	// On the human's turn, surface the capture options for each of their hand cards.
	if humanIdx >= 0 && humanIdx == sg.GetCurrentTurn() && sg.IsHumanTurn() {
		if hp := sg.GetPlayer(humanIdx); hp != nil {
			resObj.HandCaptures = make([][][]int, hp.GetCardsSize())
			for h := 0; h < hp.GetCardsSize(); h++ {
				resObj.HandCaptures[h] = sg.GetValidCaptures(h)
			}
		}
	}

	if det := sg.GetLastRoundDetail(); det != nil {
		resObj.LastRoundDetail = &controller.ScoponeWebOutputScoreDetail{
			Cards:      det.Cards,
			Diamonds:   det.Diamonds,
			Sevens:     det.Sevens,
			Scopas:     det.Scopas,
			Gained:     det.Gained,
			Settebello: det.SettebelloTm,
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if sg.GetGameEndFlag() {
		resObj.Message = swp.buildResultMessage(sg)
		resObj.MessageCode = "scopone.result.scores"
		resObj.MessageParams = map[string]string{
			"phase":      sg.GetPhase(),
			"winnerTeam": fmt.Sprintf("%d", sg.GetWinnerTeam()),
			"scores":     swp.encodeScoresParam(sg),
		}
	}

	return marshalOrError(resObj)
}

// encodeScoresParam packs final team scores into a stable, locale-free string
// the frontend can split (e.g. "0:11,1:7").
func (swp *ScoponeWebPresenter) encodeScoresParam(sg interfaces.ScoponeGame) string {
	parts := make([]string, 0, domain.ScoponeTeamCnt)
	for t := 0; t < domain.ScoponeTeamCnt; t++ {
		parts = append(parts, fmt.Sprintf("%d:%d", t, sg.GetTeamScore(t)))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage ゲーム終了時のフォールバック (英語) メッセージ。
func (swp *ScoponeWebPresenter) buildResultMessage(sg interfaces.ScoponeGame) string {
	msg := fmt.Sprintf("Game over. Team %d wins. ", sg.GetWinnerTeam())
	for t := 0; t < domain.ScoponeTeamCnt; t++ {
		msg += fmt.Sprintf("Team %d:%dpt ", t, sg.GetTeamScore(t))
	}
	return msg
}

// ActionLogOutput 棋譜を JSON 出力。
// HintOutput returns the current state as JSON. The Web GUI computes its own
// hint client-side, so this mirrors Output to satisfy the ScoponePresenter
// interface shared with the CUI.
func (swp *ScoponeWebPresenter) HintOutput(sg interfaces.ScoponeGame) string {
	return swp.Output(sg, nil)
}

func (swp *ScoponeWebPresenter) ActionLogOutput(sg interfaces.ScoponeGame) string {
	return actionLogOutputJSON(sg)
}
