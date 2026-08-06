//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GinRummyWebPresenter ジンラミーWebプレゼンタークラス
type GinRummyWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GinRummyWebPresenter) Output(g interfaces.GinRummyGame, lastErr error) string {
	resObj := new(controller.GinRummyWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.KnockerIdx = g.GetKnockerIdx()
	resObj.IsGin = g.GetIsGin()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.GinRummyWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	// ノッカーのメルド
	resObj.LayoffTargets = p.buildLayoffTargets(g)

	knockerMelds := g.GetKnockerMelds()
	resObj.KnockerMelds = make([]*controller.GinRummyWebOutputMeld, 0, len(knockerMelds))
	for _, meld := range knockerMelds {
		meldOut := &controller.GinRummyWebOutputMeld{
			Cards: make([]*controller.WebOutputCard, 0, len(meld)),
		}
		for _, card := range meld {
			meldOut.Cards = append(meldOut.Cards, cardToOutput(card))
		}
		resObj.KnockerMelds = append(resObj.KnockerMelds, meldOut)
	}

	// ノッカーのデッドウッド
	knockerDeadwood := g.GetKnockerDeadwood()
	resObj.KnockerDeadwood = make([]*controller.WebOutputCard, 0, len(knockerDeadwood))
	for _, card := range knockerDeadwood {
		resObj.KnockerDeadwood = append(resObj.KnockerDeadwood, cardToOutput(card))
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *GinRummyWebPresenter) buildPlayersOutput(g interfaces.GinRummyGame) []*controller.GinRummyWebOutputPlayer {
	out := make([]*controller.GinRummyWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman()
		// ラウンド終了/ゲーム終了時はCPUの手札も表示
		phase := g.GetPhase()
		if phase == domain.GinRummyPhaseRoundEnd || phase == domain.GinRummyPhaseGameEnd || phase == domain.GinRummyPhaseLayoff {
			showCards = true
		}
		pObj := &controller.GinRummyWebOutputPlayer{
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
func (p *GinRummyWebPresenter) buildMessage(g interfaces.GinRummyGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("ginrummy", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.GinRummyPhaseDraw:
		return "", "ginrummy.drawPhase", nil
	case domain.GinRummyPhaseDiscard:
		return "", "ginrummy.discardPhase", nil
	case domain.GinRummyPhaseLayoff:
		return "", "ginrummy.layoffPhase", nil
	case domain.GinRummyPhaseRoundEnd:
		return "", "ginrummy.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *GinRummyWebPresenter) ActionLogOutput(g interfaces.GinRummyGame) string {
	return actionLogOutputJSON(g)
}

// buildLayoffTargets は人間の手札 1 枚ごとに、足せるノッカーのメルド番号を返す。
//
// **レイオフフェーズの主題そのもの。**ディスカードフェーズは meldedIndices で
// 区分を見せているのに、レイオフには補助が無く、押してサーバーの応答で初めて
// 成否が分かる状態だった (#4823)。判定はドメインをそのまま使う。
func (p *GinRummyWebPresenter) buildLayoffTargets(g interfaces.GinRummyGame) [][]int {
	human := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			human = i
			break
		}
	}
	if human < 0 {
		return nil
	}
	hand := g.GetPlayer(human)
	out := make([][]int, 0, hand.GetCardsSize())
	for i := 0; i < hand.GetCardsSize(); i++ {
		targets := g.LayoffTargets(hand.GetCard(i))
		if targets == nil {
			targets = []int{}
		}
		out = append(out, targets)
	}
	return out
}
