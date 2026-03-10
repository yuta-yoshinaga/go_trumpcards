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

	resObj.CurrentTrick = p.buildTrickOutput(h)
	resObj.Players = p.buildPlayersOutput(h)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(h, lastErr)

	return marshalOrError(resObj)
}

// buildTrickOutput 現在のトリック情報を構築
func (p *HeartsWebPresenter) buildTrickOutput(h interfaces.HeartsGame) []*controller.HeartsWebOutputTrickCard {
	out := make([]*controller.HeartsWebOutputTrickCard, 0)
	for _, tc := range h.GetCurrentTrick() {
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
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *HeartsWebPresenter) buildMessage(h interfaces.HeartsGame, lastErr error) (string, string, map[string]string) {
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
	trick := h.GetCurrentTrick()
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
