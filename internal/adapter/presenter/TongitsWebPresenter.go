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

	// challenge の勝敗は残り点の少なさで決まるので、フロントが再計算せずに
	// 済むよう手番のプレイヤーの残り点を送る。**点数の規則を 2 箇所に持たせない。**
	resObj.RemainingPoints = -1
	if g.GetPhase() == domain.TongitsPhaseDiscard && g.IsHumanTurn() {
		cur := g.GetPlayer(g.GetCurrentPlayerIdx())
		total := 0
		for i := 0; i < cur.GetCardsSize(); i++ {
			total += domain.TongitsCardValue(cur.GetCard(i))
		}
		resObj.RemainingPoints = total
	}
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsTongits = g.GetIsTongits()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.TongitsWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
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
			Melds:           tongitsMeldsToOutput(player.GetMelds()),
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
