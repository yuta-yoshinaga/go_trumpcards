//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SevenBridgeWebPresenter セブンブリッジ Web プレゼンター
type SevenBridgeWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *SevenBridgeWebPresenter) Output(g interfaces.SevenBridgeGame, lastErr error) string {
	resObj := new(controller.SevenBridgeWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
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
	resObj.Config = controller.SevenBridgeWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SevenBridgeWebPresenter) buildPlayersOutput(g interfaces.SevenBridgeGame) []*controller.SevenBridgeWebOutputPlayer {
	out := make([]*controller.SevenBridgeWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman()
		phase := g.GetPhase()
		if phase == domain.SevenBridgePhaseRoundEnd || phase == domain.SevenBridgePhaseGameEnd {
			showCards = true
		}
		melds := make([]*controller.SevenBridgeWebOutputMeld, 0, player.GetMeldCount())
		for mi := 0; mi < player.GetMeldCount(); mi++ {
			meld := player.GetMeld(mi)
			m := &controller.SevenBridgeWebOutputMeld{
				Cards: make([]*controller.WebOutputCard, 0, len(meld)),
			}
			for _, c := range meld {
				m.Cards = append(m.Cards, cardToOutput(c))
			}
			melds = append(melds, m)
		}
		pObj := &controller.SevenBridgeWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			Melds:           melds,
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SevenBridgeWebPresenter) buildMessage(g interfaces.SevenBridgeGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("sevenbridge", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.SevenBridgePhaseDraw:
		return "", "sevenbridge.drawPhase", nil
	case domain.SevenBridgePhasePlay:
		return "", "sevenbridge.playPhase", nil
	case domain.SevenBridgePhaseRoundEnd:
		return "", "sevenbridge.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *SevenBridgeWebPresenter) ActionLogOutput(g interfaces.SevenBridgeGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。SevenBridgePresenter インタフェースを満たすための実装。
func (p *SevenBridgeWebPresenter) HintOutput(g interfaces.SevenBridgeGame) string {
	return p.Output(g, nil)
}
