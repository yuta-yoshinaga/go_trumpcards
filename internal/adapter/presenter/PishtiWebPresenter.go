//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PishtiWebPresenter は Pişti Web プレゼンタークラス。
type PishtiWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (pwp *PishtiWebPresenter) Output(pg interfaces.PishtiGame, lastErr error) string {
	resObj := new(controller.PishtiWebOutput)
	resObj.Players = make([]*controller.PishtiWebOutputPlayer, 0)
	resObj.Pile = make([]*controller.WebOutputCard, 0)
	resObj.Winners = make([]int, 0)
	resObj.FinalScores = make([]int, 0)

	resObj.CurrentTurn = pg.GetCurrentTurn()
	resObj.LastCaptureIdx = pg.GetLastCaptureIdx()
	resObj.GameEndFlag = pg.GetGameEndFlag()
	resObj.Phase = string(pg.GetPhase())
	resObj.RemainingDeck = pg.GetRemainingDeck()
	resObj.Winners = append(resObj.Winners, pg.GetWinners()...)

	config := pg.GetConfig()
	resObj.Config = controller.PishtiWebConfig{
		PlayerCnt:     config.PlayerCnt,
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.Pile = cardsToOutputOrEmpty(pg.GetPile())
	resObj.PileCount = len(pg.GetPile())
	if top := pg.GetPileTop(); top != nil {
		resObj.PileTop = cardToOutput(top)
	}

	scores := pg.GetFinalScores()
	resObj.FinalScores = append(resObj.FinalScores, scores...)

	for i := 0; i < pg.GetPlayerCnt(); i++ {
		player := pg.GetPlayer(i)
		if player == nil {
			continue
		}
		score := 0
		if i < len(scores) {
			score = scores[i]
		}
		resObj.Players = append(resObj.Players, &controller.PishtiWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			CapturedCount: player.CapturedCount(),
			PistiBonus:    player.GetPistiBonus(),
			FinalScore:    score,
		})
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if pg.GetGameEndFlag() {
		// Plain Message is an English fallback; clients that understand
		// pishti.result.scores + messageParams.scores render the localised
		// version instead.
		resObj.Message = pwp.buildResultMessage(pg)
		resObj.MessageCode = "pishti.result.scores"
		resObj.MessageParams = map[string]string{
			"scores": pwp.encodeScoresParam(pg),
		}
	}

	return marshalOrError(resObj)
}

// encodeScoresParam は最終得点を "0:11,1:7,..." 形式のロケール非依存文字列へ詰める。
func (pwp *PishtiWebPresenter) encodeScoresParam(pg interfaces.PishtiGame) string {
	scores := pg.GetFinalScores()
	parts := make([]string, 0, pg.GetPlayerCnt())
	for i := 0; i < pg.GetPlayerCnt(); i++ {
		score := 0
		if i < len(scores) {
			score = scores[i]
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, score))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage はゲーム終了時のフォールバック (英語) メッセージ。
func (pwp *PishtiWebPresenter) buildResultMessage(pg interfaces.PishtiGame) string {
	scores := pg.GetFinalScores()
	msg := "Game over. "
	for i := 0; i < pg.GetPlayerCnt(); i++ {
		p := pg.GetPlayer(i)
		if p == nil {
			continue
		}
		name := fmt.Sprintf("CPU %d", i)
		if p.GetIsHuman() {
			name = "You"
		}
		score := 0
		if i < len(scores) {
			score = scores[i]
		}
		msg += fmt.Sprintf("%s:%dpt ", name, score)
	}
	return msg
}

// ActionLogOutput は棋譜を JSON 出力する。
func (pwp *PishtiWebPresenter) ActionLogOutput(pg interfaces.PishtiGame) string {
	return actionLogOutputJSON(pg)
}
