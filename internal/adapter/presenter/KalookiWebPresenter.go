//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KalookiWebPresenter カルーキ Web プレゼンター
type KalookiWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *KalookiWebPresenter) Output(g interfaces.KalookiGame, lastErr error) string {
	resObj := new(controller.KalookiWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.OpeningThreshold = g.GetOpeningThreshold()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.RoundWinnerIdx = g.GetRoundWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.KalookiWebOutputConfig{
		CpuDifficulty:    int(cfg.CpuDifficulty),
		PlayerCount:      cfg.PlayerCount,
		OpeningThreshold: cfg.OpeningThreshold,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *KalookiWebPresenter) buildPlayersOutput(g interfaces.KalookiGame) []*controller.KalookiWebOutputPlayer {
	out := make([]*controller.KalookiWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman()
		phase := g.GetPhase()
		if phase == domain.KalookiPhaseRoundEnd || phase == domain.KalookiPhaseGameEnd {
			showCards = true
		}
		melds := make([]*controller.KalookiWebOutputMeld, 0, player.GetMeldCount())
		for mi := 0; mi < player.GetMeldCount(); mi++ {
			meld := player.GetMeld(mi)
			m := &controller.KalookiWebOutputMeld{
				Cards: make([]*controller.WebOutputCard, 0, len(meld)),
			}
			for _, c := range meld {
				m.Cards = append(m.Cards, cardToOutput(c))
			}
			melds = append(melds, m)
		}
		pObj := &controller.KalookiWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			Melds:           melds,
			HasOpened:       player.HasOpened(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *KalookiWebPresenter) buildMessage(g interfaces.KalookiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("kalooki", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.KalookiPhaseDraw:
		return "", "kalooki.drawPhase", nil
	case domain.KalookiPhaseMeld:
		return "", "kalooki.meldPhase", nil
	case domain.KalookiPhaseRoundEnd:
		return "", "kalooki.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *KalookiWebPresenter) ActionLogOutput(g interfaces.KalookiGame) string {
	return actionLogOutputJSON(g)
}
