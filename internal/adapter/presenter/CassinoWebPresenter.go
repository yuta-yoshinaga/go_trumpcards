package presenter

import (
	"fmt"

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
		resObj.Message = cwp.buildResultMessage(cg)
		resObj.MessageCode = "cassino.result.scores"
		resObj.MessageParams = map[string]string{"phase": cassinoPhaseLabel(cg.GetPhase())}
	}

	return marshalOrError(resObj)
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

// buildResultMessage ゲーム終了時の総合結果メッセージ。
func (cwp *CassinoWebPresenter) buildResultMessage(cg interfaces.CassinoGame) string {
	msg := "ゲーム終了！ "
	for i := 0; i < cg.GetPlayerCnt(); i++ {
		p := cg.GetPlayer(i)
		if p == nil {
			continue
		}
		var name string
		if p.GetIsHuman() {
			name = "あなた"
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		msg += fmt.Sprintf("%s:%d点 ", name, p.GetTotalScore())
	}
	return msg
}

// ActionLogOutput 棋譜を JSON 出力。
func (cwp *CassinoWebPresenter) ActionLogOutput(cg interfaces.CassinoGame) string {
	return actionLogOutputJSON(cg)
}
