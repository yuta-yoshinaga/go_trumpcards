//go:build !js || !wasm || classic

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ScopaWebPresenter スコパ Web プレゼンタークラス。
type ScopaWebPresenter struct{}

// Output ゲーム状態を JSON 出力。
func (swp *ScopaWebPresenter) Output(sg interfaces.ScopaGame, lastErr error) string {
	resObj := new(controller.ScopaWebOutput)
	resObj.Players = make([]*controller.ScopaWebOutputPlayer, 0)
	resObj.TableCards = make([]*controller.WebOutputCard, 0)
	resObj.CpuActions = make([]*controller.ScopaWebOutputAction, 0)
	resObj.RoundWinners = make([]int, 0)

	resObj.CurrentTurn = sg.GetCurrentTurn()
	resObj.LastCaptureIdx = sg.GetLastCaptureIdx()
	resObj.GameEndFlag = sg.GetGameEndFlag()
	resObj.Phase = sg.GetPhase()
	resObj.RemainingDeck = sg.GetRemainingDeck()
	resObj.PacksDealt = sg.GetPacksDealt()
	resObj.RoundWinners = append(resObj.RoundWinners, sg.GetRoundWinners()...)

	config := sg.GetConfig()
	resObj.Config = controller.ScopaWebConfig{
		TargetScore:   config.TargetScore,
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.TableCards = cardsToOutputOrEmpty(sg.GetTableCards())

	for _, a := range sg.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, scopaActionToOutput(a))
	}
	if ha := sg.GetHumanAction(); ha != nil {
		resObj.HumanAction = scopaActionToOutput(ha)
	}

	for i := 0; i < sg.GetPlayerCnt(); i++ {
		player := sg.GetPlayer(i)
		if player == nil {
			continue
		}
		resObj.Players = append(resObj.Players, &controller.ScopaWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			CapturedCount: player.CapturedCount(),
			ScopaCount:    player.GetScopaCount(),
			TotalScore:    player.GetTotalScore(),
		})
	}

	if det := sg.GetLastRoundDetail(); det != nil {
		resObj.LastRoundDetail = &controller.ScopaWebOutputScoreDetail{
			Cards:         det.Cards,
			Diamonds:      det.Diamonds,
			Sevens:        det.Sevens,
			HasSetteBello: det.HasSetteBello,
			Scopas:        det.Scopas,
			Gained:        det.Gained,
		}
	}

	if lastErr != nil {
		// **コードを渡さないと生の識別子が画面に出る。**`NewDomainErrorCode` で
		// 作ったエラーは `Message` が空なので `Error()` はキーそのものを返す。
		// `MessageCode` を埋めないと `GameMessageBox` は翻訳を通さず、
		// `scopa.errCaptureRequired` がそのまま両言語で表示される (#6846)。
		resObj.Message = lastErr.Error()
		resObj.MessageCode, resObj.MessageParams = domain.ErrorMessageCode(lastErr)
	} else if sg.GetGameEndFlag() {
		resObj.Message = swp.buildResultMessage(sg)
		resObj.MessageCode = "scopa.result.scores"
		resObj.MessageParams = map[string]string{
			"phase":  sg.GetPhase(),
			"scores": swp.encodeScoresParam(sg),
		}
	}

	return marshalOrError(resObj)
}

// encodeScoresParam packs final scores into a stable, locale-free string the
// frontend can split (e.g. "0:11,1:7").
func (swp *ScopaWebPresenter) encodeScoresParam(sg interfaces.ScopaGame) string {
	parts := make([]string, 0, sg.GetPlayerCnt())
	for i := 0; i < sg.GetPlayerCnt(); i++ {
		p := sg.GetPlayer(i)
		if p == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, p.GetTotalScore()))
	}
	return strings.Join(parts, ",")
}

// scopaActionToOutput converts a domain action to a WebOutput representation.
func scopaActionToOutput(a *domain.ScopaAction) *controller.ScopaWebOutputAction {
	if a == nil {
		return nil
	}
	var played *controller.WebOutputCard
	if a.PlayedCard != nil {
		played = cardToOutput(a.PlayedCard)
	}
	return &controller.ScopaWebOutputAction{
		PlayerIdx:     a.PlayerIdx,
		PlayedCard:    played,
		CapturedCards: cardsToOutput(a.CapturedCards),
		IsScopa:       a.IsScopa,
	}
}

// buildResultMessage ゲーム終了時のフォールバック (英語) メッセージ。
func (swp *ScopaWebPresenter) buildResultMessage(sg interfaces.ScopaGame) string {
	msg := "Game over. "
	for i := 0; i < sg.GetPlayerCnt(); i++ {
		p := sg.GetPlayer(i)
		if p == nil {
			continue
		}
		name := fmt.Sprintf("CPU %d", i)
		if p.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%dpt ", name, p.GetTotalScore())
	}
	return msg
}

// ActionLogOutput 棋譜を JSON 出力。
// HintOutput returns the current state as JSON. The Web GUI computes its own
// hint client-side, so this mirrors Output to satisfy the ScopaPresenter
// interface shared with the CUI.
func (swp *ScopaWebPresenter) HintOutput(sg interfaces.ScopaGame) string {
	return swp.Output(sg, nil)
}

func (swp *ScopaWebPresenter) ActionLogOutput(sg interfaces.ScopaGame) string {
	return actionLogOutputJSON(sg)
}
