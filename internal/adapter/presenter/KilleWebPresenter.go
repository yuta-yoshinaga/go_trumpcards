//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// killeFace は キッレの札(専用42枚デッキ)を手続き描画するための自己記述子を返す。
//
// 単一スートなので色でスートを表せない。**効果を持つ 5 種を色で見分けられる**
// ようにしてある (ADR-0033 の手続き描画パス)。
func killeFace(card *domain.Card) *CardFace {
	if card == nil || card.GetDesign() != domain.KilleDesign {
		return nil
	}
	r := domain.KilleRankOf(card)
	return &CardFace{
		Glyph: domain.KilleRankGlyph(r),
		Label: domain.KilleRankName(r),
		Color: domain.KilleRankColor(r),
		Deck:  domain.KilleDeckID,
	}
}

// KilleWebPresenter キッレ Webプレゼンタークラス
type KilleWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KilleWebPresenter) Output(g interfaces.KilleGame, lastErr error) string {
	resObj := new(controller.KilleWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.StockCount = g.GetStockCount()
	resObj.Pot = g.GetPot()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.LoserIdxs = append([]int{}, g.GetLoserIdxs()...)
	resObj.Events = p.buildEventsOutput(g)

	cfg := g.GetConfig()
	resObj.Config = controller.KilleWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		Stake:         cfg.Stake,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// killeWebReveal は全員の手札を公開する局面かを返す。
func killeWebReveal(g interfaces.KilleGame) bool {
	phase := g.GetPhase()
	return phase == domain.KillePhaseShowdown || phase == domain.KillePhaseGameEnd
}

// buildEventsOutput 交換記録を構築
func (p *KilleWebPresenter) buildEventsOutput(g interfaces.KilleGame) []*controller.KilleWebOutputEvent {
	events := g.GetEvents()
	out := make([]*controller.KilleWebOutputEvent, 0, len(events))
	for _, e := range events {
		if e == nil {
			continue
		}
		out = append(out, &controller.KilleWebOutputEvent{Kind: e.Kind, Actor: e.Actor, Target: e.Target})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *KilleWebPresenter) buildPlayersOutput(g interfaces.KilleGame) []*controller.KilleWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.KilleWebOutputPlayer, 0, len(players))
	reveal := killeWebReveal(g)
	current := g.GetCurrentPlayerIdx()
	exchanging := g.GetPhase() == domain.KillePhaseExchange
	for i := range players {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		show := player.GetIsHuman() || reveal
		var card *controller.WebOutputCard
		if show && player.GetCardsSize() > 0 {
			card = cardToOutputWithFace(player.GetCard(0), killeFace)
		}
		// **強さも伏せる。**公開していない札の強さを送ると裏が読めてしまう。
		strength := 0
		if show {
			strength = g.KilleStrength(i)
		}
		out = append(out, &controller.KilleWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Card:          card,
			Strength:      strength,
			Chips:         player.GetChips(),
			Reentries:     player.GetReentries(),
			ReentryCost:   g.KilleReentryCost(i),
			CanReenter:    player.CanReenter(),
			IsOut:         player.IsOut(),
			KnockedBy:     player.GetKnockedBy(),
			IsSatisfied:   player.IsSatisfied(),
			IsFinished:    player.GetIsFinished(),
			IsCurrentTurn: exchanging && i == current,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *KilleWebPresenter) buildMessage(g interfaces.KilleGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("kille", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.KillePhaseExchange:
		if g.GetCurrentPlayerIdx() == g.GetDealerIdx() {
			return "", "kille.dealerTurn", nil
		}
		return "", "kille.exchangePhase", nil
	case domain.KillePhaseShowdown:
		return "", "kille.showdown", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *KilleWebPresenter) ActionLogOutput(g interfaces.KilleGame) string {
	return actionLogOutputJSON(g)
}
