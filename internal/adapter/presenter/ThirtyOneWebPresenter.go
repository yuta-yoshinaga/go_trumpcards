//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ThirtyOneWebPresenter ThirtyOne Webプレゼンタークラス
type ThirtyOneWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ThirtyOneWebPresenter) Output(g interfaces.ThirtyOneGame, lastErr error) string {
	resObj := new(controller.ThirtyOneWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.KnockerIdx = g.GetKnockerIdx()
	resObj.ThirtyOneIdx = g.GetThirtyOneIdx()
	resObj.RoundWinnerIdx = g.GetRoundWinnerIdx()
	resObj.RoundLosers = append([]int{}, g.GetRoundLosers()...)

	if top := g.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.ThirtyOneWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		InitialLives:  cfg.InitialLives,
		KnockThresholds: controller.ThirtyOneWebOutputKnockThresholds{
			Easy:   domain.ThirtyOneKnockThresholdEasy,
			Normal: domain.ThirtyOneKnockThresholdNormal,
			Hard:   domain.ThirtyOneKnockThresholdHard,
		},
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// reveal ラウンド終了/ゲーム終了時に全員の手札を公開するか
func thirtyOneReveal(g interfaces.ThirtyOneGame) bool {
	phase := g.GetPhase()
	return phase == domain.ThirtyOnePhaseRoundEnd || phase == domain.ThirtyOnePhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ThirtyOneWebPresenter) buildPlayersOutput(g interfaces.ThirtyOneGame) []*controller.ThirtyOneWebOutputPlayer {
	out := make([]*controller.ThirtyOneWebOutputPlayer, 0, g.GetPlayerCnt())
	reveal := thirtyOneReveal(g)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || reveal
		score := 0
		if showCards {
			score = player.BestSuitScore()
		}
		out = append(out, &controller.ThirtyOneWebOutputPlayer{
			ID:           i,
			IsHuman:      player.GetIsHuman(),
			CardCount:    player.GetCardsSize(),
			Cards:        playerCardsToOutput(player, showCards),
			Lives:        player.GetLives(),
			Score:        score,
			IsEliminated: player.IsEliminated(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ThirtyOneWebPresenter) buildMessage(g interfaces.ThirtyOneGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("thirtyone", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.ThirtyOnePhaseDraw:
		return "", "thirtyone.drawPhase", nil
	case domain.ThirtyOnePhaseDiscard:
		return "", "thirtyone.discardPhase", nil
	case domain.ThirtyOnePhaseRoundEnd:
		if g.GetThirtyOneIdx() >= 0 {
			return "", "thirtyone.thirtyOneHit", nil
		}
		return "", "thirtyone.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *ThirtyOneWebPresenter) ActionLogOutput(g interfaces.ThirtyOneGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (p *ThirtyOneWebPresenter) HintOutput(g interfaces.ThirtyOneGame) string {
	return p.Output(g, nil)
}
