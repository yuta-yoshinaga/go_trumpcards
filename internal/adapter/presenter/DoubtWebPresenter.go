package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DoubtWebPresenter ダウトWebプレゼンタークラス
type DoubtWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (dwp *DoubtWebPresenter) Output(d interfaces.DoubtGame, lastErr error) string {
	resObj := new(controller.DoubtWebOutput)
	resObj.Players = make([]*controller.DoubtWebOutputPlayer, 0)
	resObj.CurrentTurn = d.GetCurrentTurn()
	resObj.Phase = int(d.GetPhase())
	resObj.TableCardCount = d.GetTableCardCount()
	resObj.GameEndFlag = d.GetGameEndFlag()
	resObj.WinnerIdx = d.GetWinnerIdx()
	resObj.DoubtWindowSec = d.GetConfig().DoubtWindowSec
	resObj.PenaltyDrawLimit = d.GetConfig().PenaltyDrawLimit

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.DoubtWebOutputAction, 0)
	for _, action := range d.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, dwp.actionToOutput(action))
	}

	// 人間の最後の行動
	if ha := d.GetHumanAction(); ha != nil {
		resObj.HumanAction = dwp.actionToOutput(ha)
	}

	// 最後のアクション (カードを出した情報)
	if la := d.GetLastAction(); la != nil {
		resObj.LastAction = &controller.DoubtWebOutputAction{
			PlayerIdx:    la.PlayerIdx,
			ClaimedValue: la.ClaimedValue,
			CardCount:    la.CardCount,
		}
	}

	// CPUダウター
	cpuDoubters := d.GetCpuDoubters()
	if cpuDoubters == nil {
		resObj.CpuDoubters = make([]int, 0)
	} else {
		resObj.CpuDoubters = cpuDoubters
	}

	// ダウト解決結果
	if dr := d.GetLastDoubtResult(); dr != nil {
		resObj.LastDoubtResult = &controller.DoubtWebOutputDoubtResult{
			DoubterIdx:     dr.DoubterIdx,
			CardPlayerIdx:  dr.CardPlayerIdx,
			WasLying:       dr.WasLying,
			LoserIdx:       dr.LoserIdx,
			CardCount:      dr.CardCount,
			DiscardedCount: dr.DiscardedCount,
			RevealedCards:  cardsToOutput(dr.RevealedCards),
		}
	}

	// プレイヤー情報
	for i := 0; i < d.GetPlayerCnt(); i++ {
		player := d.GetPlayer(i)
		pObj := &controller.DoubtWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			IsFinished: player.GetIsFinished(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メタAI情報
	if profile := d.GetHumanProfile(); profile != nil {
		resObj.MetaAI = &controller.DoubtWebOutputMetaAI{
			Enabled:        true,
			GamesPlayed:    profile.GamesPlayed,
			BluffRate:      profile.BluffRate(1), // medium bracket as representative
			DoubtAccuracy:  profile.DoubtAccuracy(),
			HesitationMean: profile.HesitationMean,
		}
		pd := profile.Export()
		resObj.Profile = &pd
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if d.GetGameEndFlag() {
		winnerIdx := d.GetWinnerIdx()
		player := d.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		resObj.Message, resObj.MessageCode, resObj.MessageParams = buildWinnerWebMessage(
			"doubt", winnerIdx, isHuman)
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (dwp *DoubtWebPresenter) ActionLogOutput(d interfaces.DoubtGame) string {
	return actionLogOutputJSON(d)
}

// actionToOutput DoubtCpuAction を DoubtWebOutputAction に変換
// IsBluff は意図的に除外する（ダウト解決前に隠されたゲーム状態をクライアントに漏洩しないため）
// HasTell はゲームの「リーク」メカニクスとして意図的に公開する
func (dwp *DoubtWebPresenter) actionToOutput(a *domain.DoubtCpuAction) *controller.DoubtWebOutputAction {
	return &controller.DoubtWebOutputAction{
		PlayerIdx:    a.PlayerIdx,
		ClaimedValue: a.ClaimedValue,
		CardCount:    a.CardCount,
		HasTell:      a.HasTell,
		HesitationMs: a.HesitationMs,
	}
}
