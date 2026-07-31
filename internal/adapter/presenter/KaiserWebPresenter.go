//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KaiserWebPresenter カイザー Webプレゼンタークラス
type KaiserWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KaiserWebPresenter) Output(g interfaces.KaiserGame, lastErr error) string {
	resObj := new(controller.KaiserWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Contract = int(g.GetContract())
	resObj.KittySize = g.GetKittySize()
	resObj.TrickLeaderIdx = g.GetTrickLeaderIdx()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.HeartFiveBy = g.GetHeartFiveBy()
	resObj.SpadeThreeBy = g.GetSpadeThreeBy()
	resObj.BidMade = g.IsBidMade()
	resObj.TargetScore = g.GetTargetScore()
	resObj.MinBid = domain.KaiserMinBid
	resObj.MaxBid = domain.KaiserMaxBid
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	for team := range domain.KaiserTeamCnt {
		resObj.TeamHandPoints[team] = g.GetHandPoints(team)
		resObj.TeamScores[team] = g.GetScore(team)
	}

	trick := g.GetTrick()
	resObj.Trick = make([]*controller.WebOutputCard, 0, len(trick))
	for _, c := range trick {
		if out := cardToOutput(c); out != nil {
			resObj.Trick = append(resObj.Trick, out)
		}
	}

	bids := g.GetBids()
	resObj.Bids = make([]*controller.KaiserWebOutputBid, 0, len(bids))
	for _, b := range bids {
		if b == nil {
			continue
		}
		resObj.Bids = append(resObj.Bids, &controller.KaiserWebOutputBid{
			Player: b.Player, Value: b.Value, Contract: int(b.Contract),
		})
	}
	if hb := g.GetHighBid(); hb != nil {
		resObj.HighBid = &controller.KaiserWebOutputBid{
			Player: hb.Player, Value: hb.Value, Contract: int(hb.Contract),
		}
	}

	// **出せる札はサーバーが決める。**追随が強制なのでフロントで再現するとずれる。
	resObj.ValidPlays = make([]int, 0)
	if g.GetPhase() == domain.KaiserPhasePlay && g.IsHumanTurn() {
		resObj.ValidPlays = append(resObj.ValidPlays, g.KaiserValidPlays(g.GetCurrentPlayerIdx())...)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.KaiserWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		AllowNoTrump:  cfg.AllowNoTrump,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// kaiserWebReveal は手札を公開する局面かを返す。
func kaiserWebReveal(g interfaces.KaiserGame) bool {
	phase := g.GetPhase()
	return phase == domain.KaiserPhaseHandEnd || phase == domain.KaiserPhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *KaiserWebPresenter) buildPlayersOutput(g interfaces.KaiserGame) []*controller.KaiserWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.KaiserWebOutputPlayer, 0, len(players))
	reveal := kaiserWebReveal(g)
	for i := range players {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		// **相手の手札は伏せる。**枚数だけ送る。
		if player.GetIsHuman() || reveal {
			for j := range player.GetCardsSize() {
				if c := cardToOutput(player.GetCard(j)); c != nil {
					cards = append(cards, c)
				}
			}
		}
		out = append(out, &controller.KaiserWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Team:          domain.KaiserTeamOf(i),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			IsDealer:      i == g.GetDealerIdx(),
			IsDeclarer:    i == g.GetDeclarerIdx(),
			IsCurrentTurn: g.GetPhase() == domain.KaiserPhasePlay && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *KaiserWebPresenter) buildMessage(g interfaces.KaiserGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		// **チーム戦なので勝敗は席ではなくチームで見る。**人間は席 0 = チーム 0。
		if g.GetWinnerTeam() == domain.KaiserTeamOf(0) {
			return "your team wins", "kaiser.result.humanWin", nil
		}
		return "your team loses", "kaiser.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.KaiserPhaseBid:
		return "", "kaiser.bidPhase", nil
	case domain.KaiserPhaseDiscard:
		if g.GetContract() == domain.KaiserContractTrump && g.GetTrumpSuit() == 0 {
			return "", "kaiser.nameTrump", nil
		}
		return "", "kaiser.discardPhase", nil
	case domain.KaiserPhasePlay:
		return "", "kaiser.playPhase", nil
	case domain.KaiserPhaseHandEnd:
		if g.IsBidMade() {
			return "", "kaiser.handMade", nil
		}
		return "", "kaiser.handSet", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *KaiserWebPresenter) ActionLogOutput(g interfaces.KaiserGame) string {
	return actionLogOutputJSON(g)
}
