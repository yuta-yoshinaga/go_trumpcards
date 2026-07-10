//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ThreeThirteenWebPresenter スリー・サーティーン Web プレゼンター
type ThreeThirteenWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *ThreeThirteenWebPresenter) Output(g interfaces.ThreeThirteenGame, lastErr error) string {
	resObj := new(controller.ThreeThirteenWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.Round = g.GetRound()
	resObj.WildRank = g.WildRank()
	resObj.DealCount = g.GetDealCount()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.KnockerIdx = g.GetKnockerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.ThreeThirteenWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PlayerCount:   cfg.PlayerCount,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ThreeThirteenWebPresenter) buildPlayersOutput(g interfaces.ThreeThirteenGame) []*controller.ThreeThirteenWebOutputPlayer {
	out := make([]*controller.ThreeThirteenWebOutputPlayer, 0)
	phase := g.GetPhase()
	revealAll := phase == domain.ThreeThirteenPhaseRoundEnd || phase == domain.ThreeThirteenPhaseGameEnd
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || revealAll
		pObj := &controller.ThreeThirteenWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			Deadwood:        g.GetPlayerDeadwoodValue(i),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ThreeThirteenWebPresenter) buildMessage(g interfaces.ThreeThirteenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("threethirteen", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.ThreeThirteenPhaseDraw:
		return "", "threethirteen.drawPhase", nil
	case domain.ThreeThirteenPhaseDiscard:
		return "", "threethirteen.discardPhase", nil
	case domain.ThreeThirteenPhaseRoundEnd:
		return "", "threethirteen.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *ThreeThirteenWebPresenter) ActionLogOutput(g interfaces.ThreeThirteenGame) string {
	return actionLogOutputJSON(g)
}
