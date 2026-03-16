package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HeartsWebPresenter ハーツWebプレゼンタークラス
type HeartsWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *HeartsWebPresenter) Output(h interfaces.HeartsGame, lastErr error) string {
	resObj := new(controller.HeartsWebOutput)
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

	trick := h.GetCurrentTrick()
	resObj.CurrentTrick = p.buildTrickOutput(trick)
	resObj.Players = p.buildPlayersOutput(h)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(h, trick, lastErr)

	return marshalOrError(resObj)
}

// buildTrickOutput 現在のトリック情報を構築
func (p *HeartsWebPresenter) buildTrickOutput(trick []*domain.HeartsTrickCard) []*controller.HeartsWebOutputTrickCard {
	out := make([]*controller.HeartsWebOutputTrickCard, 0)
	for _, tc := range trick {
		out = append(out, &controller.HeartsWebOutputTrickCard{
			PlayerIdx: tc.PlayerIdx,
			Card:      cardToOutput(tc.Card),
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *HeartsWebPresenter) buildPlayersOutput(h interfaces.HeartsGame) []*controller.HeartsWebOutputPlayer {
	out := make([]*controller.HeartsWebOutputPlayer, 0)
	for i := 0; i < h.GetPlayerCnt(); i++ {
		player := h.GetPlayer(i)
		pObj := &controller.HeartsWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *HeartsWebPresenter) buildMessage(h interfaces.HeartsGame, trick []*domain.HeartsTrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if h.GetGameEndFlag() {
		winnerIdx := h.GetWinnerIdx()
		msg := p.buildResultMessage(h)
		player := h.GetPlayer(winnerIdx)
		if player != nil && player.GetIsHuman() {
			return msg, "hearts.result.humanWin", nil
		}
		params := map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
		return msg, "hearts.result.cpuWin", params
	}
	switch h.GetPhase() {
	case domain.HeartsPhasePass:
		return "", "hearts.passPhase", nil
	case domain.HeartsPhasePlay:
		if len(trick) == 0 {
			return "", "hearts.playPhase.lead", nil
		}
		return "", "hearts.playPhase.follow", nil
	case domain.HeartsPhaseTrickEnd:
		return "", "hearts.trickEnd", nil
	case domain.HeartsPhaseRoundEnd:
		return "", "hearts.roundEnd", nil
	}
	return "", "", nil
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
