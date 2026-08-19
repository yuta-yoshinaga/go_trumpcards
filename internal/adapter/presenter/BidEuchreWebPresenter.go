//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BidEuchreWebPresenter ビッド・ユーカー Webプレゼンタークラス
type BidEuchreWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BidEuchreWebPresenter) Output(g interfaces.BidEuchreGame, lastErr error) string {
	resObj := new(controller.BidEuchreWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.Trump = int(g.GetTrump())
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TrumpChosen = g.IsTrumpChosen()
	resObj.TrickLeaderIdx = g.GetTrickLeaderIdx()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.GameTarget = domain.BidEuchreGameTarget
	resObj.MinBid = domain.BidEuchreMinBid
	resObj.MaxBid = domain.BidEuchreMaxBid
	resObj.HandSize = domain.BidEuchreHandSize
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	for team := range domain.BidEuchreTeamCnt {
		resObj.TeamTricks[team] = g.BidEuchreTeamTricks(team)
		resObj.Scores[team] = g.GetScore(team)
	}

	trick := g.GetTrick()
	resObj.Trick = make([]*controller.WebOutputCard, 0, len(trick))
	for _, c := range trick {
		if out := cardToOutput(c); out != nil {
			resObj.Trick = append(resObj.Trick, out)
		}
	}

	bids := g.GetBids()
	resObj.Bids = make([]*controller.BidEuchreWebOutputBid, 0, len(bids))
	for _, b := range bids {
		if b == nil {
			continue
		}
		resObj.Bids = append(resObj.Bids, bidEuchreBidOut(b))
	}
	if hb := g.GetHighBid(); hb != nil {
		resObj.HighBid = bidEuchreBidOut(hb)
	}

	if r := g.GetLastResult(); r != nil {
		resObj.LastResult = &controller.BidEuchreWebOutputResult{
			Points: r.Points,
			Tricks: r.Tricks,
			Made:   r.Made,
			Bid:    r.Bid,
		}
	}

	// **出せる札はサーバーが決める。**左ボワーが切札扱いになるのでフロントで
	// 再現するとずれる。
	resObj.ValidPlays = make([]int, 0)
	if g.GetPhase() == domain.BidEuchrePhasePlay && g.IsHumanTurn() {
		resObj.ValidPlays = append(resObj.ValidPlays, g.BidEuchreValidPlays(g.GetCurrentPlayerIdx())...)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.BidEuchreWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		AllowNoTrump:  cfg.AllowNoTrump,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// bidEuchreBidOut は 1 件の宣言をワイヤ表現へ変換する。
func bidEuchreBidOut(b *domain.BidEuchreBid) *controller.BidEuchreWebOutputBid {
	return &controller.BidEuchreWebOutputBid{Player: b.Player, Value: b.Value}
}

// bidEuchreWebReveal は全員の手札を公開する局面かを返す。
func bidEuchreWebReveal(g interfaces.BidEuchreGame) bool {
	phase := g.GetPhase()
	return phase == domain.BidEuchrePhaseHandEnd || phase == domain.BidEuchrePhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BidEuchreWebPresenter) buildPlayersOutput(g interfaces.BidEuchreGame) []*controller.BidEuchreWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.BidEuchreWebOutputPlayer, 0, len(players))
	reveal := bidEuchreWebReveal(g)
	for i := range players {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		// **キティが無いので、伏せた札は他家の手札だけ。**
		if player.GetIsHuman() || reveal {
			for j := range player.GetCardsSize() {
				if c := cardToOutput(player.GetCard(j)); c != nil {
					cards = append(cards, c)
				}
			}
		}
		out = append(out, &controller.BidEuchreWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Team:          domain.BidEuchreTeamOf(i),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			TricksWon:     g.GetTricksWon(i),
			IsDealer:      i == g.GetDealerIdx(),
			IsDeclarer:    i == g.GetDeclarerIdx(),
			IsCurrentTurn: g.GetPhase() == domain.BidEuchrePhasePlay && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BidEuchreWebPresenter) buildMessage(g interfaces.BidEuchreGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		// **チーム戦なので勝敗は席ではなくチームで見る。**人間は席 0 = チーム 0。
		if g.GetWinnerTeam() == domain.BidEuchreTeamOf(0) {
			return "your team wins the game", "bideuchre.result.humanWin", nil
		}
		return "the other team wins the game", "bideuchre.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.BidEuchrePhaseBid:
		return "", "bideuchre.bidPhase", nil
	case domain.BidEuchrePhaseChooseTrump:
		return "", "bideuchre.trumpPhase", nil
	case domain.BidEuchrePhasePlay:
		return "", "bideuchre.playPhase", nil
	case domain.BidEuchrePhaseHandEnd:
		if r := g.GetLastResult(); r != nil && !r.Made {
			return "", "bideuchre.handSet", nil
		}
		return "", "bideuchre.handMade", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *BidEuchreWebPresenter) ActionLogOutput(g interfaces.BidEuchreGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput は Web では通常の状態出力と同じ (ヒントは CUI 専用)。
func (p *BidEuchreWebPresenter) HintOutput(g interfaces.BidEuchreGame) string {
	return p.Output(g, nil)
}
