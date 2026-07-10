//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CanastaWebPresenter カナスタWebプレゼンタークラス
type CanastaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CanastaWebPresenter) Output(g interfaces.CanastaGame, lastErr error) string {
	resObj := new(controller.CanastaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.DiscardPileCount = g.GetDiscardPileCount()
	resObj.IsFrozen = g.GetIsFrozen()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.CanastaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CanastaWebPresenter) buildPlayersOutput(g interfaces.CanastaGame) []*controller.CanastaWebOutputPlayer {
	out := make([]*controller.CanastaWebOutputPlayer, 0)
	phase := g.GetPhase()
	showAllCards := phase == domain.CanastaPhaseRoundEnd || phase == domain.CanastaPhaseGameEnd

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || showAllCards

		// メルド出力
		melds := make([]*controller.CanastaWebOutputMeld, 0, len(player.GetMelds()))
		for _, m := range player.GetMelds() {
			meldOut := &controller.CanastaWebOutputMeld{
				Cards:     make([]*controller.WebOutputCard, 0, len(m.Cards)),
				IsNatural: m.IsNatural,
				IsCanasta: m.IsCanasta(),
				Rank:      m.GetRank(),
			}
			for _, card := range m.Cards {
				meldOut.Cards = append(meldOut.Cards, cardToOutput(card))
			}
			melds = append(melds, meldOut)
		}

		// 赤3出力
		red3s := make([]*controller.WebOutputCard, 0, len(player.GetRed3s()))
		for _, card := range player.GetRed3s() {
			red3s = append(red3s, cardToOutput(card))
		}

		pObj := &controller.CanastaWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			Melds:           melds,
			Red3Count:       len(player.GetRed3s()),
			Red3s:           red3s,
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			HasCanasta:      player.HasCanasta(),
			HasInitMeld:     player.GetHasInitMeld(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *CanastaWebPresenter) buildMessage(g interfaces.CanastaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("canasta", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.CanastaPhaseDraw:
		return "", "canasta.drawPhase", nil
	case domain.CanastaPhaseMeld:
		return "", "canasta.meldPhase", nil
	case domain.CanastaPhaseDiscard:
		return "", "canasta.discardPhase", nil
	case domain.CanastaPhaseRoundEnd:
		return "", "canasta.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *CanastaWebPresenter) ActionLogOutput(g interfaces.CanastaGame) string {
	return actionLogOutputJSON(g)
}
