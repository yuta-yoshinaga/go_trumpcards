//go:build !js || !wasm || extra4

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SixBidSoloWebPresenter シックスビッド・ソロ Webプレゼンタークラス
type SixBidSoloWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SixBidSoloWebPresenter) Output(g interfaces.SixBidSoloGame, lastErr error) string {
	resObj := new(controller.SixBidSoloWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Declared = g.IsDeclared()
	resObj.SpreadOpen = g.IsSpreadOpen()
	resObj.TrickLeaderIdx = g.GetTrickLeaderIdx()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TotalPoints = domain.SixBidSoloTotalPoints
	resObj.BaseTarget = domain.SixBidSoloBaseTarget
	resObj.HandSize = domain.SixBidSoloHandSize
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	// **目標点はビッドとスートの両方で決まる。**ギャランティーは ♥ で 74、他は 80。
	for kind := range int(domain.SixBidSoloBidCount) {
		resObj.BidTargets[kind] = domain.SixBidSoloTargetPoints(domain.SixBidSoloBidKind(kind), g.GetTrumpSuit())
	}

	trick := g.GetTrick()
	resObj.Trick = make([]*controller.WebOutputCard, 0, len(trick))
	for _, c := range trick {
		if out := cardToOutput(c); out != nil {
			resObj.Trick = append(resObj.Trick, out)
		}
	}

	// **ウィドウは精算まで伏せたまま。**中身が見えると読みが崩れる。
	widow := g.GetWidow()
	resObj.WidowSize = len(widow)
	resObj.Widow = make([]*controller.WebOutputCard, 0, len(widow))
	if sixBidSoloWebReveal(g) {
		for _, c := range widow {
			if out := cardToOutput(c); out != nil {
				resObj.Widow = append(resObj.Widow, out)
			}
		}
	}

	if c := cardToOutput(g.GetCalledCard()); c != nil {
		resObj.CalledCard = c
	}

	bids := g.GetBids()
	resObj.Bids = make([]*controller.SixBidSoloWebOutputBid, 0, len(bids))
	for _, b := range bids {
		if b == nil {
			continue
		}
		resObj.Bids = append(resObj.Bids, sixBidSoloBidOut(b))
	}
	if hb := g.GetHighBid(); hb != nil {
		resObj.HighBid = sixBidSoloBidOut(hb)
	}

	if r := g.GetLastResult(); r != nil {
		resObj.LastResult = &controller.SixBidSoloWebOutputResult{
			Kind:           int(r.Kind),
			Declarer:       r.Declarer,
			DeclarerPoints: r.DeclarerPoints,
			WidowPoints:    r.WidowPoints,
			Target:         r.Target,
			Made:           r.Made,
			Value:          r.Value,
			Deltas:         r.Deltas,
		}
	}

	// **出せる札はサーバーが決める。**追随が強制なのでフロントで再現するとずれる。
	resObj.ValidPlays = make([]int, 0)
	if g.GetPhase() == domain.SixBidSoloPhasePlay && g.IsHumanTurn() {
		resObj.ValidPlays = append(resObj.ValidPlays, g.SixBidSoloValidPlays(g.GetCurrentPlayerIdx())...)
	}

	cfg := g.GetConfig()
	resObj.TargetHands = cfg.TargetHands
	resObj.Config = controller.SixBidSoloWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetHands:   cfg.TargetHands,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// sixBidSoloBidOut は 1 件の宣言をワイヤ表現へ変換する。
func sixBidSoloBidOut(b *domain.SixBidSoloBid) *controller.SixBidSoloWebOutputBid {
	return &controller.SixBidSoloWebOutputBid{Player: b.Player, Kind: int(b.Kind)}
}

// sixBidSoloWebReveal は伏せ札を公開する局面かを返す。
func sixBidSoloWebReveal(g interfaces.SixBidSoloGame) bool {
	phase := g.GetPhase()
	return phase == domain.SixBidSoloPhaseHandEnd || phase == domain.SixBidSoloPhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SixBidSoloWebPresenter) buildPlayersOutput(g interfaces.SixBidSoloGame) []*controller.SixBidSoloWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.SixBidSoloWebOutputPlayer, 0, len(players))
	reveal := sixBidSoloWebReveal(g)
	for i := range players {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		// **スプレッド・ミゼールでは宣言者の手札も公開される。**それが賭けの中身。
		spread := g.IsSpreadOpen() && i == g.GetDeclarerIdx()
		if player.GetIsHuman() || reveal || spread {
			for j := range player.GetCardsSize() {
				if c := cardToOutput(player.GetCard(j)); c != nil {
					cards = append(cards, c)
				}
			}
		}
		out = append(out, &controller.SixBidSoloWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			Points:        g.GetPoints(i),
			TricksWon:     g.GetTricksWon(i),
			Score:         g.GetScore(i),
			IsDealer:      i == g.GetDealerIdx(),
			IsDeclarer:    i == g.GetDeclarerIdx(),
			IsCurrentTurn: g.GetPhase() == domain.SixBidSoloPhasePlay && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SixBidSoloWebPresenter) buildMessage(g interfaces.SixBidSoloGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() == 0 {
			return "you win", "sixbidsolo.result.humanWin", nil
		}
		return "a cpu wins", "sixbidsolo.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.SixBidSoloPhaseBid:
		return "", "sixbidsolo.bidPhase", nil
	case domain.SixBidSoloPhaseDeclare:
		return "", "sixbidsolo.declarePhase", nil
	case domain.SixBidSoloPhasePlay:
		return "", "sixbidsolo.playPhase", nil
	case domain.SixBidSoloPhaseHandEnd:
		if r := g.GetLastResult(); r != nil && !r.Made {
			return "", "sixbidsolo.handSet", nil
		}
		return "", "sixbidsolo.handMade", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *SixBidSoloWebPresenter) ActionLogOutput(g interfaces.SixBidSoloGame) string {
	return actionLogOutputJSON(g)
}
