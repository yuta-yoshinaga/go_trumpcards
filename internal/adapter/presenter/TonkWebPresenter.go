package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TonkWebPresenter Tonk Webプレゼンタークラス
type TonkWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TonkWebPresenter) Output(g interfaces.TonkGame, lastErr error) string {
	resObj := new(controller.TonkWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.UndercutRiskMax = domain.TonkUndercutRiskMax

	// **CUI は毎ターン「ノック可能/不可」を出しているのに、Web は手計算だった。**
	// 判断の基準ごと送るので、フロントは閾値の数値を写さずに済む。
	resObj.KnockThreshold = domain.TonkKnockThreshold
	resObj.BestDeadwood = -1
	if g.GetPhase() == domain.TonkPhaseDiscard && g.IsHumanTurn() {
		best, _ := g.GetBestDeadwood(g.GetCurrentPlayerIdx())
		resObj.BestDeadwood = best
	}
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.KnockerIdx = g.GetKnockerIdx()
	resObj.IsTonk = g.GetIsTonk()
	resObj.IsUndercut = g.GetIsUndercut()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.TonkWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.KnockerMelds = meldsToOutput(g.GetKnockerMelds())
	resObj.OpponentMelds = meldsToOutput(g.GetOpponentMelds())

	knockerDeadwood := g.GetKnockerDeadwood()
	resObj.KnockerDeadwood = make([]*controller.WebOutputCard, 0, len(knockerDeadwood))
	for _, card := range knockerDeadwood {
		resObj.KnockerDeadwood = append(resObj.KnockerDeadwood, cardToOutput(card))
	}

	opponentDeadwood := g.GetOpponentDeadwood()
	resObj.OpponentDeadwood = make([]*controller.WebOutputCard, 0, len(opponentDeadwood))
	for _, card := range opponentDeadwood {
		resObj.OpponentDeadwood = append(resObj.OpponentDeadwood, cardToOutput(card))
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// meldsToOutput converts domain melds into web output melds.
func meldsToOutput(melds [][]*domain.Card) []*controller.TonkWebOutputMeld {
	out := make([]*controller.TonkWebOutputMeld, 0, len(melds))
	for _, meld := range melds {
		meldOut := &controller.TonkWebOutputMeld{
			Cards: make([]*controller.WebOutputCard, 0, len(meld)),
		}
		for _, card := range meld {
			meldOut.Cards = append(meldOut.Cards, cardToOutput(card))
		}
		out = append(out, meldOut)
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TonkWebPresenter) buildPlayersOutput(g interfaces.TonkGame) []*controller.TonkWebOutputPlayer {
	out := make([]*controller.TonkWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman()
		phase := g.GetPhase()
		if phase == domain.TonkPhaseRoundEnd || phase == domain.TonkPhaseGameEnd {
			showCards = true
		}
		pObj := &controller.TonkWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TonkWebPresenter) buildMessage(g interfaces.TonkGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("tonk", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.TonkPhaseDraw:
		return "", "tonk.drawPhase", nil
	case domain.TonkPhaseDiscard:
		return "", "tonk.discardPhase", nil
	case domain.TonkPhaseRoundEnd:
		if g.GetIsTonk() {
			return "", "tonk.tonkOnDeal", nil
		}
		return "", "tonk.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *TonkWebPresenter) ActionLogOutput(g interfaces.TonkGame) string {
	return actionLogOutputJSON(g)
}
