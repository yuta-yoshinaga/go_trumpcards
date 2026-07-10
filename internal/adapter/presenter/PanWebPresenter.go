//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PanWebPresenter パングインゲ Web プレゼンター
type PanWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *PanWebPresenter) Output(g interfaces.PanGame, lastErr error) string {
	resObj := new(controller.PanWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TargetRounds = g.GetTargetRounds()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.DeckSize = domain.PanDeckSize
	resObj.WinMeldCount = domain.PanWinMeldCount
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.PanDeclarerIdx = g.GetPanDeclarerIdx()

	if top := g.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.PanWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PanWebPresenter) buildPlayersOutput(g interfaces.PanGame) []*controller.PanWebOutputPlayer {
	out := make([]*controller.PanWebOutputPlayer, 0)
	phase := g.GetPhase()
	revealAll := phase == domain.PanPhaseRoundEnd || phase == domain.PanPhaseGameEnd
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || revealAll

		melds := player.GetLaidMelds()
		laid := make([]*controller.PanWebOutputMeld, 0, len(melds))
		for _, m := range melds {
			meldOut := &controller.PanWebOutputMeld{Cards: make([]*controller.WebOutputCard, 0, len(m))}
			for _, c := range m {
				meldOut.Cards = append(meldOut.Cards, cardToOutput(c))
			}
			laid = append(laid, meldOut)
		}

		handPoints := 0
		if showCards {
			handPoints = g.PlayerHandPoints(i)
		}

		pObj := &controller.PanWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			LaidMelds:       laid,
			MeldedCount:     g.PlayerMeldedCount(i),
			Chips:           player.GetChips(),
			HandPoints:      handPoints,
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PanWebPresenter) buildMessage(g interfaces.PanGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("pan", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.PanPhaseDraw:
		return "", "pan.drawPhase", nil
	case domain.PanPhasePlay:
		return "", "pan.playPhase", nil
	case domain.PanPhaseRoundEnd:
		return "", "pan.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *PanWebPresenter) ActionLogOutput(g interfaces.PanGame) string {
	return actionLogOutputJSON(g)
}
