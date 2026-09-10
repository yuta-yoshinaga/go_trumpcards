//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BiribaWebPresenter ビリバWebプレゼンタークラス
type BiribaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BiribaWebPresenter) Output(g interfaces.BiribaGame, lastErr error) string {
	resObj := new(controller.BiribaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.DiscardPileCount = g.GetDiscardPileCount()
	resObj.PozzettoCount = g.GetPozzettoCount()
	resObj.IsFrozen = g.GetIsFrozen()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	// 捨て札パイル全体を古い順（下から上）に公開する。ビリバでは山ごと引き取るため
	// パイルの中身は全プレイヤーに見えている情報であり、取得判断の核心となる。
	pile := g.GetDiscardPile()
	resObj.DiscardPile = make([]*controller.WebOutputCard, 0, len(pile))
	for _, card := range pile {
		resObj.DiscardPile = append(resObj.DiscardPile, cardToOutput(card))
	}

	cfg := g.GetConfig()
	resObj.Config = controller.BiribaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	// ヒントはドメインが人間の手番でだけ返す (#5628)。
	//
	// **理由はキーに直してから運ぶ。**ドメインが返すのは `draw_discard_pair` の
	// ような内部識別子で、そのまま渡すとフロントは存在しないキーを引き、翻訳の
	// 代わりに生の識別子が画面に出る。CUI 側も同じ対応表を通している。
	if h := g.GetHint(); h != nil {
		resObj.Hint = &controller.BiribaWebOutputHint{
			Action:  h.Action,
			Indices: h.Indices,
			Reason:  biribaWebHintReasonKeys[h.Reason],
		}
	}

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BiribaWebPresenter) buildPlayersOutput(g interfaces.BiribaGame) []*controller.BiribaWebOutputPlayer {
	out := make([]*controller.BiribaWebOutputPlayer, 0)
	phase := g.GetPhase()
	showAllCards := phase == domain.BiribaPhaseRoundEnd || phase == domain.BiribaPhaseGameEnd

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || showAllCards

		// メルド出力
		melds := make([]*controller.BiribaWebOutputMeld, 0, len(player.GetMelds()))
		for _, m := range player.GetMelds() {
			meldOut := &controller.BiribaWebOutputMeld{
				Cards:     make([]*controller.WebOutputCard, 0, len(m.Cards)),
				IsNatural: m.IsNatural,
				IsBiriba:  m.IsBiriba(),
				Rank:      m.GetRank(),
			}
			for _, card := range m.Cards {
				meldOut.Cards = append(meldOut.Cards, cardToOutput(card))
			}
			melds = append(melds, meldOut)
		}

		// 赤3出力
		red3s := make([]*controller.WebOutputCard, 0, len(player.GetRed3s()))
		for _, card := range player.GetRed3s() {
			red3s = append(red3s, cardToOutput(card))
		}

		pObj := &controller.BiribaWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			Melds:           melds,
			Red3Count:       len(player.GetRed3s()),
			Red3s:           red3s,
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			HasBiriba:       player.HasBiriba(),
			HasInitMeld:     player.GetHasInitMeld(),
			TookPozzetto:    player.GetTookPozzetto(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BiribaWebPresenter) buildMessage(g interfaces.BiribaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("biriba", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.BiribaPhaseDraw:
		return "", "biriba.drawPhase", nil
	case domain.BiribaPhaseMeld:
		return "", "biriba.meldPhase", nil
	case domain.BiribaPhaseDiscard:
		return "", "biriba.discardPhase", nil
	case domain.BiribaPhaseRoundEnd:
		return "", "biriba.roundEnd", nil
	}
	return "", "", nil
}

// biribaWebHintReasonKeys maps the domain's internal hint reasons to the
// locale suffixes the web catalogue uses. Same pairs as the CUI's
// biribaHintReasonKeys, without the `biriba.` prefix (the page adds `hint.`).
var biribaWebHintReasonKeys = map[string]string{
	"draw_discard_pair": "hintReasonDrawDiscard",
	"draw_stock_safe":   "hintReasonDrawStock",
	"meld_available":    "hintReasonMeld",
	"no_meld":           "hintReasonNoMeld",
	"discard_safe":      "hintReasonDiscard",
}

// HintOutput ヒント情報をJSON出力する。専用のレスポンスは持たず通常の状態出力を
// 返すが、**その状態にドメインのヒントが載っている** (#5628)。CUI と同じ値なので、
// 2 つの画面が同じ盤面で違う手を勧めることがない。
func (p *BiribaWebPresenter) HintOutput(g interfaces.BiribaGame) string {
	return p.Output(g, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BiribaWebPresenter) ActionLogOutput(g interfaces.BiribaGame) string {
	return actionLogOutputJSON(g)
}

// BiribaWebHintReasonKeyForTest exposes the reason mapping to tests in the
// external test package.
func BiribaWebHintReasonKeyForTest(reason string) string { return biribaWebHintReasonKeys[reason] }
