//go:build !js || !wasm || extra5

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TongitsWebPresenter Tongits Webプレゼンタークラス
type TongitsWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TongitsWebPresenter) Output(g interfaces.TongitsGame, lastErr error) string {
	resObj := new(controller.TongitsWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.UndercutRiskMax = domain.TongitsUndercutRiskMax

	// **CUI は毎ターン「ノック可能/不可」を出しているのに、Web は手計算だった。**
	// 判断の基準ごと送るので、フロントは閾値の数値を写さずに済む。
	resObj.KnockThreshold = domain.TongitsKnockThreshold
	resObj.BestDeadwood = -1
	if g.GetPhase() == domain.TongitsPhaseDiscard && g.IsHumanTurn() {
		best, _ := g.GetBestDeadwood(g.GetCurrentPlayerIdx())
		resObj.BestDeadwood = best
	}
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.KnockerIdx = g.GetKnockerIdx()
	resObj.IsTongits = g.GetIsTongits()
	resObj.IsUndercut = g.GetIsUndercut()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.TongitsWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.KnockerMelds = tongitsMeldsToOutput(g.GetKnockerMelds())
	resObj.OpponentMelds = tongitsMeldsToOutput(g.GetOpponentMelds())

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
func tongitsMeldsToOutput(melds [][]*domain.Card) []*controller.TongitsWebOutputMeld {
	out := make([]*controller.TongitsWebOutputMeld, 0, len(melds))
	for _, meld := range melds {
		meldOut := &controller.TongitsWebOutputMeld{
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
func (p *TongitsWebPresenter) buildPlayersOutput(g interfaces.TongitsGame) []*controller.TongitsWebOutputPlayer {
	out := make([]*controller.TongitsWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman()
		phase := g.GetPhase()
		if phase == domain.TongitsPhaseRoundEnd || phase == domain.TongitsPhaseGameEnd {
			showCards = true
		}
		pObj := &controller.TongitsWebOutputPlayer{
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
func (p *TongitsWebPresenter) buildMessage(g interfaces.TongitsGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("tongits", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.TongitsPhaseDraw:
		return "", "tongits.drawPhase", nil
	case domain.TongitsPhaseDiscard:
		return "", "tongits.discardPhase", nil
	case domain.TongitsPhaseRoundEnd:
		if g.GetIsTongits() {
			return "", "tongits.tongitsOnDeal", nil
		}
		return "", "tongits.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *TongitsWebPresenter) ActionLogOutput(g interfaces.TongitsGame) string {
	return actionLogOutputJSON(g)
}
