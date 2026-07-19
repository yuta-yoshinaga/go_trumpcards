//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// EscobaWebPresenter エスコバ Web プレゼンタークラス。
type EscobaWebPresenter struct{}

// Output ゲーム状態を JSON 出力。
func (ewp *EscobaWebPresenter) Output(eg interfaces.EscobaGame, lastErr error) string {
	resObj := new(controller.EscobaWebOutput)
	resObj.Players = make([]*controller.EscobaWebOutputPlayer, 0)
	resObj.TableCards = make([]*controller.WebOutputCard, 0)
	resObj.HandCaptures = make([][][]int, 0)

	resObj.Phase = eg.GetPhase()
	resObj.RoundNumber = eg.GetRoundNumber()
	resObj.CurrentTurn = eg.GetCurrentTurn()
	resObj.DealerIdx = eg.GetDealerIdx()
	resObj.StockRemaining = eg.GetStockRemaining()
	resObj.LastCaptureIdx = eg.GetLastCaptureIdx()
	resObj.WinnerIdx = eg.GetWinnerIdx()
	resObj.GameEndFlag = eg.GetGameEndFlag()
	resObj.IsHumanTurn = eg.IsHumanTurn()

	config := eg.GetConfig()
	resObj.Config = controller.EscobaWebConfig{
		TargetScore:   config.TargetScore,
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.TableCards = cardsToOutputOrEmpty(eg.GetTableCards())

	humanIdx := -1
	for i := 0; i < eg.GetPlayerCnt(); i++ {
		player := eg.GetPlayer(i)
		if player == nil {
			continue
		}
		if player.GetIsHuman() {
			humanIdx = i
		}
		// Reveal the captured pile only for the human; CPUs stay count-only to
		// preserve the memory-game aspect (どの札を取ったかは相手に見せない)。
		captured := make([]*controller.WebOutputCard, 0)
		if player.GetIsHuman() {
			captured = cardsToOutputOrEmpty(player.GetCapturedCards())
		}
		resObj.Players = append(resObj.Players, &controller.EscobaWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			HandCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			CapturedCount: player.CapturedCount(),
			CapturedCards: captured,
			EscobaCount:   player.GetScopaCount(),
			Score:         player.GetTotalScore(),
		})
	}

	// On the human's turn, surface the capture options for each of their hand cards.
	if humanIdx >= 0 && humanIdx == eg.GetCurrentTurn() && eg.IsHumanTurn() {
		if hp := eg.GetPlayer(humanIdx); hp != nil {
			resObj.HandCaptures = make([][][]int, hp.GetCardsSize())
			for h := 0; h < hp.GetCardsSize(); h++ {
				resObj.HandCaptures[h] = eg.GetValidCaptures(h)
			}
		}
	}

	if det := eg.GetLastRoundDetail(); det != nil {
		resObj.LastRoundDetail = &controller.EscobaWebOutputScoreDetail{
			Cards:      det.Cards,
			Espadas:    det.Espadas,
			Sevens:     det.Sevens,
			Oros:       det.Oros,
			Escobas:    det.Escobas,
			Gained:     det.Gained,
			AceEspada:  det.AceEsp,
			SeteEspada: det.SeteEsp,
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if eg.GetGameEndFlag() {
		resObj.Message = ewp.buildResultMessage(eg)
		resObj.MessageCode = "escoba.result.winner"
		resObj.MessageParams = map[string]string{
			"phase":  eg.GetPhase(),
			"winner": fmt.Sprintf("%d", eg.GetWinnerIdx()),
		}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了時のフォールバック (英語) メッセージ。
func (ewp *EscobaWebPresenter) buildResultMessage(eg interfaces.EscobaGame) string {
	msg := fmt.Sprintf("Game over. Player %d wins. ", eg.GetWinnerIdx())
	for i := 0; i < eg.GetPlayerCnt(); i++ {
		p := eg.GetPlayer(i)
		if p == nil {
			continue
		}
		msg += fmt.Sprintf("P%d:%dpt ", i, p.GetTotalScore())
	}
	return msg
}

// ActionLogOutput 棋譜を JSON 出力。
// HintOutput returns the current state as JSON. The Web GUI computes its own
// hint client-side, so this mirrors Output to satisfy the EscobaPresenter
// interface shared with the CUI.
func (ewp *EscobaWebPresenter) HintOutput(eg interfaces.EscobaGame) string {
	return ewp.Output(eg, nil)
}

func (ewp *EscobaWebPresenter) ActionLogOutput(eg interfaces.EscobaGame) string {
	return actionLogOutputJSON(eg)
}
