//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// YanivWebPresenter Yaniv Webプレゼンタークラス
type YanivWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *YanivWebPresenter) Output(g interfaces.YanivGame, lastErr error) string {
	resObj := new(controller.YanivWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.CallerIdx = g.GetCallerIdx()
	resObj.AsafWinnerIdx = g.GetAsafWinnerIdx()
	resObj.IsAsaf = g.GetIsAsaf()
	resObj.RoundScores = append([]int{}, g.GetRoundScores()...)

	resObj.PickupCards = make([]*controller.WebOutputCard, 0, len(g.GetPickupCards()))
	for _, c := range g.GetPickupCards() {
		resObj.PickupCards = append(resObj.PickupCards, cardToOutput(c))
	}

	cfg := g.GetConfig()
	resObj.Config = controller.YanivWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		ScoreLimit:    cfg.ScoreLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *YanivWebPresenter) buildPlayersOutput(g interfaces.YanivGame) []*controller.YanivWebOutputPlayer {
	out := make([]*controller.YanivWebOutputPlayer, 0, g.GetPlayerCnt())
	reveal := yanivReveal(g)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || reveal
		total := 0
		if showCards {
			total = player.HandTotal()
		}
		out = append(out, &controller.YanivWebOutputPlayer{
			ID:           i,
			IsHuman:      player.GetIsHuman(),
			CardCount:    player.GetCardsSize(),
			Cards:        playerCardsToOutput(player, showCards),
			Score:        player.GetScore(),
			HandTotal:    total,
			IsEliminated: player.IsEliminated(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *YanivWebPresenter) buildMessage(g interfaces.YanivGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("yaniv", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.YanivPhaseDiscard:
		return "", "yaniv.discardPhase", nil
	case domain.YanivPhaseDraw:
		return "", "yaniv.drawPhase", nil
	case domain.YanivPhaseRoundEnd:
		if g.GetIsAsaf() {
			return "", "yaniv.asafResult", nil
		}
		if g.GetCallerIdx() >= 0 {
			return "", "yaniv.yanivResult", nil
		}
		return "", "yaniv.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *YanivWebPresenter) ActionLogOutput(g interfaces.YanivGame) string {
	return actionLogOutputJSON(g)
}
