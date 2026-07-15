package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CassinoWebPresenter カシノ Web プレゼンタークラス。
type CassinoWebPresenter struct{}

// Output ゲーム状態を JSON 出力。
func (cwp *CassinoWebPresenter) Output(cg interfaces.CassinoGame, lastErr error) string {
	resObj := new(controller.CassinoWebOutput)
	resObj.Players = make([]*controller.CassinoWebOutputPlayer, 0)
	resObj.TableCards = make([]*controller.WebOutputCard, 0)
	resObj.Builds = make([]*controller.CassinoWebOutputBuild, 0)
	resObj.CpuActions = make([]*controller.CassinoWebOutputAction, 0)
	resObj.RoundWinners = make([]int, 0)

	resObj.CurrentTurn = cg.GetCurrentTurn()
	resObj.LastCaptureIdx = cg.GetLastCaptureIdx()
	resObj.GameEndFlag = cg.GetGameEndFlag()
	resObj.Phase = cg.GetPhase()
	resObj.RemainingDeck = cg.GetRemainingDeck()
	resObj.PacksDealt = cg.GetPacksDealt()
	resObj.RoundWinners = append(resObj.RoundWinners, cg.GetRoundWinners()...)

	config := cg.GetConfig()
	resObj.Config = controller.CassinoWebConfig{
		TargetScore:       config.TargetScore,
		MultiBuildEnabled: config.MultiBuildEnabled,
		SweepBonusEnabled: config.SweepBonusEnabled,
		CpuDifficulty:     int(config.CpuDifficulty),
	}

	// 場の単独カード
	resObj.TableCards = cardsToOutputOrEmpty(cg.GetTableCards())

	// ビルド
	for _, b := range cg.GetBuilds() {
		groups := make([][]*controller.WebOutputCard, 0, len(b.Groups))
		for _, g := range b.Groups {
			groups = append(groups, cardsToOutput(g))
		}
		resObj.Builds = append(resObj.Builds, &controller.CassinoWebOutputBuild{
			OwnerIdx: b.OwnerIdx,
			Value:    b.Value,
			Groups:   groups,
			IsMulti:  b.IsMulti,
		})
	}

	// CPU 行動
	for _, a := range cg.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, cassinoActionToOutput(a))
	}
	// 人間 行動
	if ha := cg.GetHumanAction(); ha != nil {
		resObj.HumanAction = cassinoActionToOutput(ha)
	}

	// プレイヤー
	for i := 0; i < cg.GetPlayerCnt(); i++ {
		player := cg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := &controller.CassinoWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			CapturedCount: player.CapturedCount(),
			SweepCount:    player.GetSweepCount(),
			TotalScore:    player.GetTotalScore(),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// 得点詳細
	if det := cg.GetLastRoundDetail(); det != nil {
		resObj.LastRoundDetail = &controller.CassinoWebOutputScoreDetail{
			Cards:           det.Cards,
			Spades:          det.Spades,
			Aces:            det.Aces,
			HasBigCasino:    det.HasBigCasino,
			HasLittleCasino: det.HasLittleCasino,
			Sweeps:          det.Sweeps,
			Gained:          det.Gained,
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if cg.GetGameEndFlag() {
		// Plain `Message` is provided as an English fallback; clients that
		// understand `cassino.result.scores` + `messageParams.scores` should
		// render the localised version instead.
		resObj.Message = cwp.buildResultMessage(cg)
		resObj.MessageCode = "cassino.result.scores"
		resObj.MessageParams = map[string]string{
			"phase":  cg.GetPhase(),
			"scores": cwp.encodeScoresParam(cg),
		}
	}

	return marshalOrError(resObj)
}

// encodeScoresParam packs final scores into a stable, locale-free string the
// frontend can split (e.g. "0:11,1:7,2:5,3:3"). The frontend re-formats it
// using the active i18n locale.
func (cwp *CassinoWebPresenter) encodeScoresParam(cg interfaces.CassinoGame) string {
	parts := make([]string, 0, cg.GetPlayerCnt())
	for i := 0; i < cg.GetPlayerCnt(); i++ {
		p := cg.GetPlayer(i)
		if p == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, p.GetTotalScore()))
	}
	return strings.Join(parts, ",")
}

// cassinoActionToOutput converts a domain action to a WebOutput representation.
func cassinoActionToOutput(a *domain.CassinoAction) *controller.CassinoWebOutputAction {
	if a == nil {
		return nil
	}
	var played *controller.WebOutputCard
	if a.PlayedCard != nil {
		played = cardToOutput(a.PlayedCard)
	}
	return &controller.CassinoWebOutputAction{
		PlayerIdx:     a.PlayerIdx,
		Type:          string(a.Type),
		PlayedCard:    played,
		CapturedCards: cardsToOutput(a.CapturedCards),
		BuildValue:    a.BuildValue,
		IsSweep:       a.IsSweep,
	}
}

// buildResultMessage ゲーム終了時のフォールバック (英語) メッセージ。
// クライアントが messageCode を理解する場合はこのメッセージを使わず i18n
// 経由で表示する想定。
func (cwp *CassinoWebPresenter) buildResultMessage(cg interfaces.CassinoGame) string {
	msg := "Game over. "
	for i := 0; i < cg.GetPlayerCnt(); i++ {
		p := cg.GetPlayer(i)
		if p == nil {
			continue
		}
		var name string
		if p.GetIsHuman() {
			name = "You"
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		msg += fmt.Sprintf("%s:%dpt ", name, p.GetTotalScore())
	}
	return msg
}

// ActionLogOutput 棋譜を JSON 出力。
func (cwp *CassinoWebPresenter) ActionLogOutput(cg interfaces.CassinoGame) string {
	return actionLogOutputJSON(cg)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint /
// suggestCassinoAction) で算出するため、通常の状態出力を返す。CassinoPresenter
// インタフェースを満たすための実装。
func (cwp *CassinoWebPresenter) HintOutput(cg interfaces.CassinoGame) string {
	return cwp.Output(cg, nil)
}
