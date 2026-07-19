//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BurracoWebPresenter ブラーコWebプレゼンタークラス
type BurracoWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BurracoWebPresenter) Output(g interfaces.BurracoGame, lastErr error) string {
	resObj := new(controller.BurracoWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.DiscardPileCount = g.GetDiscardPileCount()
	resObj.PozzettoCount = g.GetPozzettoCount()
	resObj.IsFrozen = g.GetIsFrozen()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	// 捨て札パイル全体を古い順（下から上）に公開する。ブラーコでは山ごと引き取るため
	// パイルの中身は全プレイヤーに見えている情報であり、取得判断の核心となる。
	pile := g.GetDiscardPile()
	resObj.DiscardPile = make([]*controller.WebOutputCard, 0, len(pile))
	for _, card := range pile {
		resObj.DiscardPile = append(resObj.DiscardPile, cardToOutput(card))
	}

	cfg := g.GetConfig()
	resObj.Config = controller.BurracoWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BurracoWebPresenter) buildPlayersOutput(g interfaces.BurracoGame) []*controller.BurracoWebOutputPlayer {
	out := make([]*controller.BurracoWebOutputPlayer, 0)
	phase := g.GetPhase()
	showAllCards := phase == domain.BurracoPhaseRoundEnd || phase == domain.BurracoPhaseGameEnd

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || showAllCards

		// メルド出力
		melds := make([]*controller.BurracoWebOutputMeld, 0, len(player.GetMelds()))
		for _, m := range player.GetMelds() {
			meldOut := &controller.BurracoWebOutputMeld{
				Cards:     make([]*controller.WebOutputCard, 0, len(m.Cards)),
				IsNatural: m.IsNatural,
				IsBurraco: m.IsBurraco(),
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

		pObj := &controller.BurracoWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			Melds:           melds,
			Red3Count:       len(player.GetRed3s()),
			Red3s:           red3s,
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			HasBurraco:      player.HasBurraco(),
			HasInitMeld:     player.GetHasInitMeld(),
			TookPozzetto:    player.GetTookPozzetto(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BurracoWebPresenter) buildMessage(g interfaces.BurracoGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("burraco", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.BurracoPhaseDraw:
		return "", "burraco.drawPhase", nil
	case domain.BurracoPhaseMeld:
		return "", "burraco.meldPhase", nil
	case domain.BurracoPhaseDiscard:
		return "", "burraco.discardPhase", nil
	case domain.BurracoPhaseRoundEnd:
		return "", "burraco.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *BurracoWebPresenter) ActionLogOutput(g interfaces.BurracoGame) string {
	return actionLogOutputJSON(g)
}
