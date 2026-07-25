package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WhistWebPresenter ホイストWebプレゼンタークラス
type WhistWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WhistWebPresenter) Output(w interfaces.WhistGame, lastErr error) string {
	resObj := p.buildBase(w)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(w, w.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *WhistWebPresenter) buildBase(w interfaces.WhistGame) *controller.WhistWebOutput {
	resObj := new(controller.WhistWebOutput)
	resObj.Phase = int(w.GetPhase())
	resObj.RoundNumber = w.GetRoundNumber()
	resObj.TrickNumber = w.GetTrickNumber()
	resObj.CurrentPlayerIdx = w.GetCurrentPlayerIdx()
	resObj.TrumpSuit = w.GetTrumpSuit()
	resObj.DealerIdx = w.GetDealerIdx()
	resObj.TeamScores = [2]int{w.GetTeamScore(0), w.GetTeamScore(1)}
	resObj.GameEndFlag = w.GetGameEndFlag()
	resObj.WinnerTeam = w.GetWinnerTeam()
	resObj.LeadPlayerIdx = w.GetLeadPlayerIdx()

	// 設定
	cfg := w.GetConfig()
	resObj.Config = controller.WhistWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = p.buildTrickOutput(w.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(w)
	return resObj
}

// buildTrickOutput 現在のトリック情報を構築
func (p *WhistWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.WhistWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.WhistWebOutputTrickCard {
		return &controller.WhistWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *WhistWebPresenter) buildPlayersOutput(w interfaces.WhistGame) []*controller.WhistWebOutputPlayer {
	out := make([]*controller.WhistWebOutputPlayer, 0)
	for i := 0; i < w.GetPlayerCnt(); i++ {
		player := w.GetPlayer(i)
		pObj := &controller.WhistWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
			Team:            player.GetTeam(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *WhistWebPresenter) buildMessage(w interfaces.WhistGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if w.GetGameEndFlag() {
		winnerTeam := w.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("whist.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch w.GetPhase() {
	case domain.WhistPhasePlay:
		if len(trick) == 0 {
			return "", "whist.playPhase.lead", nil
		}
		return "", "whist.playPhase.follow", nil
	case domain.WhistPhaseTrickEnd:
		return "", "whist.trickEnd", nil
	case domain.WhistPhaseRoundEnd:
		return "", "whist.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *WhistWebPresenter) HintOutput(w interfaces.WhistGame) string {
	hint := w.GetHint()
	resObj := p.buildBase(w)
	if hint != nil {
		resObj.Hint = &controller.WhistWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *WhistWebPresenter) ActionLogOutput(w interfaces.WhistGame) string {
	return actionLogOutputJSON(w)
}
