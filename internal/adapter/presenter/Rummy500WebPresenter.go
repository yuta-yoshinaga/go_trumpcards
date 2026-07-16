//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// Rummy500WebPresenter Rummy 500Webプレゼンタークラス
type Rummy500WebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *Rummy500WebPresenter) Output(g interfaces.Rummy500Game, lastErr error) string {
	resObj := new(controller.Rummy500WebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.RoundEnderIdx = g.GetRoundEnderIdx()

	// 捨て札の山
	pile := g.GetDiscardPile()
	resObj.DiscardPile = make([]*controller.WebOutputCard, 0, len(pile))
	for _, card := range pile {
		resObj.DiscardPile = append(resObj.DiscardPile, cardToOutput(card))
	}

	cfg := g.GetConfig()
	resObj.Config = controller.Rummy500WebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *Rummy500WebPresenter) buildPlayersOutput(g interfaces.Rummy500Game) []*controller.Rummy500WebOutputPlayer {
	out := make([]*controller.Rummy500WebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman()
		phase := g.GetPhase()
		if phase == domain.Rummy500PhaseRoundEnd || phase == domain.Rummy500PhaseGameEnd {
			showCards = true
		}
		// 場に出したメルド
		melds := player.GetLaidMelds()
		laid := make([]*controller.Rummy500WebOutputMeld, 0, len(melds))
		for _, m := range melds {
			meldOut := &controller.Rummy500WebOutputMeld{
				Cards: make([]*controller.WebOutputCard, 0, len(m)),
			}
			for _, c := range m {
				meldOut.Cards = append(meldOut.Cards, cardToOutput(c))
			}
			laid = append(laid, meldOut)
		}
		pObj := &controller.Rummy500WebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			LaidMelds:       laid,
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *Rummy500WebPresenter) buildMessage(g interfaces.Rummy500Game, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("rummy500", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.Rummy500PhaseDraw:
		return "", "rummy500.drawPhase", nil
	case domain.Rummy500PhasePlay:
		return "", "rummy500.playPhase", nil
	case domain.Rummy500PhaseRoundEnd:
		return "", "rummy500.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *Rummy500WebPresenter) ActionLogOutput(g interfaces.Rummy500Game) string {
	return actionLogOutputJSON(g)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (p *Rummy500WebPresenter) HintOutput(g interfaces.Rummy500Game) string {
	return p.Output(g, nil)
}
