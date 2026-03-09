package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DoubtWebPresenter ダウトWebプレゼンタークラス
type DoubtWebPresenter struct{}

// NewDoubtWebPresenter コンストラクタ
func NewDoubtWebPresenter() *DoubtWebPresenter {
	return &DoubtWebPresenter{}
}

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
			Cards:      make([]*controller.WebOutputCard, 0),
		}
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, cardToOutput(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メタAI情報
	if profile := d.GetHumanProfile(); profile != nil {
		resObj.MetaAI = &controller.DoubtWebOutputMetaAI{
			Enabled:       true,
			GamesPlayed:   profile.GamesPlayed,
			BluffRate:     profile.BluffRate(1), // medium bracket as representative
			DoubtAccuracy: profile.DoubtAccuracy(),
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if d.GetGameEndFlag() {
		winnerIdx := d.GetWinnerIdx()
		resObj.Message = dwp.buildResultMessage(d)
		player := d.GetPlayer(winnerIdx)
		if player != nil && player.GetIsHuman() {
			resObj.MessageCode = "doubt.result.humanWin"
		} else {
			resObj.MessageCode = "doubt.result.cpuWin"
			resObj.MessageParams = map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
		}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (dwp *DoubtWebPresenter) buildResultMessage(d interfaces.DoubtGame) string {
	winnerIdx := d.GetWinnerIdx()
	player := d.GetPlayer(winnerIdx)
	if player == nil {
		return fmt.Sprintf("ゲーム終了！ CPU %dの勝ち！", winnerIdx)
	}
	var name string
	if player.GetIsHuman() {
		name = "あなた"
	} else {
		name = fmt.Sprintf("CPU %d", winnerIdx)
	}
	return fmt.Sprintf("ゲーム終了！ %sの勝ち！", name)
}

// ActionLogOutput 棋譜をJSON出力
func (dwp *DoubtWebPresenter) ActionLogOutput(d interfaces.DoubtGame) string {
	if !d.GetGameEndFlag() {
		return actionLogToJSON(nil)
	}
	return actionLogToJSON(d.GetActionLog())
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
