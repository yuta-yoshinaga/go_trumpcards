//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ChinchonWebPresenter チンチョンWebプレゼンタークラス
type ChinchonWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ChinchonWebPresenter) Output(g interfaces.ChinchonGame, lastErr error) string {
	resObj := new(controller.ChinchonWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.KnockerIdx = g.GetKnockerIdx()

	if top := g.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.ChinchonWebOutputConfig{
		CpuDifficulty:    int(cfg.CpuDifficulty),
		PlayerCount:      cfg.PlayerCount,
		KnockThreshold:   cfg.KnockThreshold,
		EliminationLimit: cfg.EliminationLimit,
	}

	// ノッカーのメルド (レイオフ用)
	knockerMelds := g.GetKnockerMelds()
	resObj.KnockerMelds = make([]*controller.ChinchonWebOutputMeld, 0, len(knockerMelds))
	for _, meld := range knockerMelds {
		meldOut := &controller.ChinchonWebOutputMeld{
			Cards: make([]*controller.WebOutputCard, 0, len(meld)),
		}
		for _, card := range meld {
			meldOut.Cards = append(meldOut.Cards, cardToOutput(card))
		}
		resObj.KnockerMelds = append(resObj.KnockerMelds, meldOut)
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ChinchonWebPresenter) buildPlayersOutput(g interfaces.ChinchonGame) []*controller.ChinchonWebOutputPlayer {
	out := make([]*controller.ChinchonWebOutputPlayer, 0)
	phase := g.GetPhase()
	revealAll := phase == domain.ChinchonPhaseLayoff || phase == domain.ChinchonPhaseRoundEnd || phase == domain.ChinchonPhaseGameEnd
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || revealAll
		pObj := &controller.ChinchonWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			Eliminated:      player.GetEliminated(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ChinchonWebPresenter) buildMessage(g interfaces.ChinchonGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		if winnerIdx < 0 {
			return "", "chinchon.draw", nil
		}
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("chinchon", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.ChinchonPhaseDraw:
		return "", "chinchon.drawPhase", nil
	case domain.ChinchonPhaseDiscard:
		return "", "chinchon.discardPhase", nil
	case domain.ChinchonPhaseLayoff:
		return "", "chinchon.layoffPhase", nil
	case domain.ChinchonPhaseRoundEnd:
		return "", "chinchon.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *ChinchonWebPresenter) ActionLogOutput(g interfaces.ChinchonGame) string {
	return actionLogOutputJSON(g)
}
