package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HeartsWebPresenter ハーツWebプレゼンタークラス
type HeartsWebPresenter struct{}

// NewHeartsWebPresenter コンストラクタ
func NewHeartsWebPresenter() *HeartsWebPresenter {
	return &HeartsWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (p *HeartsWebPresenter) Output(h interfaces.HeartsGame, lastErr error) string {
	resObj := new(controller.HeartsWebOutput)
	resObj.Players = make([]*controller.HeartsWebOutputPlayer, 0)
	resObj.Phase = int(h.GetPhase())
	resObj.RoundNumber = h.GetRoundNumber()
	resObj.TrickNumber = h.GetTrickNumber()
	resObj.CurrentPlayerIdx = h.GetCurrentPlayerIdx()
	resObj.HeartsBroken = h.GetHeartsBroken()
	resObj.PassDirection = int(h.GetPassDirection())
	resObj.GameEndFlag = h.GetGameEndFlag()
	resObj.WinnerIdx = h.GetWinnerIdx()
	resObj.LeadPlayerIdx = h.GetLeadPlayerIdx()

	// 設定
	cfg := h.GetConfig()
	resObj.Config = controller.HeartsWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	// トリック
	trick := h.GetCurrentTrick()
	resObj.CurrentTrick = make([]*controller.HeartsWebOutputTrickCard, 0)
	for _, tc := range trick {
		resObj.CurrentTrick = append(resObj.CurrentTrick, &controller.HeartsWebOutputTrickCard{
			PlayerIdx: tc.PlayerIdx,
			Card:      cardToOutput(tc.Card),
		})
	}

	// プレイヤー情報
	for i := 0; i < h.GetPlayerCnt(); i++ {
		player := h.GetPlayer(i)
		pObj := &controller.HeartsWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           make([]*controller.WebOutputCard, 0),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		}
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, cardToOutput(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if h.GetGameEndFlag() {
		winnerIdx := h.GetWinnerIdx()
		resObj.Message = p.buildResultMessage(h)
		player := h.GetPlayer(winnerIdx)
		if player != nil && player.GetIsHuman() {
			resObj.MessageCode = "hearts.result.humanWin"
		} else {
			resObj.MessageCode = "hearts.result.cpuWin"
			resObj.MessageParams = map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
		}
	} else {
		phase := h.GetPhase()
		switch phase {
		case domain.HeartsPhasePass:
			resObj.MessageCode = "hearts.passPhase"
		case domain.HeartsPhasePlay:
			if len(trick) == 0 {
				resObj.MessageCode = "hearts.playPhase.lead"
			} else {
				resObj.MessageCode = "hearts.playPhase.follow"
			}
		case domain.HeartsPhaseTrickEnd:
			resObj.MessageCode = "hearts.trickEnd"
		case domain.HeartsPhaseRoundEnd:
			resObj.MessageCode = "hearts.roundEnd"
		}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (p *HeartsWebPresenter) buildResultMessage(h interfaces.HeartsGame) string {
	winnerIdx := h.GetWinnerIdx()
	player := h.GetPlayer(winnerIdx)
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
func (p *HeartsWebPresenter) ActionLogOutput(h interfaces.HeartsGame) string {
	return actionLogOutputJSON(h)
}
