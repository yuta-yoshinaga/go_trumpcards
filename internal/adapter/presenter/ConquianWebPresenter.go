//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ConquianWebPresenter コンキャンWebプレゼンタークラス
type ConquianWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ConquianWebPresenter) Output(g interfaces.ConquianGame, lastErr error) string {
	resObj := new(controller.ConquianWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.RoundWinnerIdx = g.GetRoundWinnerIdx()
	resObj.TookDiscard = g.GetTookDiscard()

	if top := g.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.ConquianWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetWins:    cfg.TargetWins,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ConquianWebPresenter) buildPlayersOutput(g interfaces.ConquianGame) []*controller.ConquianWebOutputPlayer {
	out := make([]*controller.ConquianWebOutputPlayer, 0)
	phase := g.GetPhase()
	revealAll := phase == domain.ConquianPhaseRoundEnd || phase == domain.ConquianPhaseGameEnd
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || revealAll

		melds := player.GetMelds()
		meldsOut := make([]*controller.ConquianWebOutputMeld, 0, len(melds))
		for _, meld := range melds {
			meldOut := &controller.ConquianWebOutputMeld{
				Cards: make([]*controller.WebOutputCard, 0, len(meld)),
			}
			for _, card := range meld {
				meldOut.Cards = append(meldOut.Cards, cardToOutput(card))
			}
			meldsOut = append(meldsOut, meldOut)
		}

		pObj := &controller.ConquianWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, showCards),
			Melds:     meldsOut,
			Wins:      player.GetWins(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ConquianWebPresenter) buildMessage(g interfaces.ConquianGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		if winnerIdx < 0 {
			return "", "conquian.draw", nil
		}
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("conquian", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.ConquianPhaseDraw:
		return "", "conquian.drawPhase", nil
	case domain.ConquianPhaseMeld:
		return "", "conquian.meldPhase", nil
	case domain.ConquianPhaseRoundEnd:
		return "", "conquian.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *ConquianWebPresenter) ActionLogOutput(g interfaces.ConquianGame) string {
	return actionLogOutputJSON(g)
}
