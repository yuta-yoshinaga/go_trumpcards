//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BostonWebPresenter ボストン Webプレゼンタークラス
type BostonWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BostonWebPresenter) Output(g interfaces.BostonGame, lastErr error) string {
	resObj := new(controller.BostonWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.PartnerIdx = g.GetPartnerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Exposed = g.IsExposed()
	resObj.TrickLeaderIdx = g.GetTrickLeaderIdx()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.DeclarerTricks = g.BostonDeclarerTricks()
	resObj.BidMade = g.IsBidMade()
	resObj.HandSize = domain.BostonHandSize
	resObj.TargetHands = g.GetTargetHands()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	trick := g.GetTrick()
	resObj.Trick = make([]*controller.WebOutputCard, 0, len(trick))
	for _, c := range trick {
		if out := cardToOutput(c); out != nil {
			resObj.Trick = append(resObj.Trick, out)
		}
	}

	bids := g.GetBids()
	resObj.Bids = make([]*controller.BostonWebOutputBid, 0, len(bids))
	for _, b := range bids {
		if b == nil {
			continue
		}
		resObj.Bids = append(resObj.Bids, bostonBidOut(b))
	}
	if hb := g.GetHighBid(); hb != nil {
		resObj.HighBid = bostonBidOut(hb)
	}
	resObj.BidOptions = bostonBidOptions()

	// **出せる札はサーバーが決める。**追随が強制なのでフロントで再現するとずれる。
	resObj.ValidPlays = make([]int, 0)
	if g.GetPhase() == domain.BostonPhasePlay && g.IsHumanTurn() {
		resObj.ValidPlays = append(resObj.ValidPlays, g.BostonValidPlays(g.GetCurrentPlayerIdx())...)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.BostonWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetHands:   cfg.TargetHands,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// bostonBidOut は 1 件の宣言をワイヤ表現へ変換する。
func bostonBidOut(b *domain.BostonBidRecord) *controller.BostonWebOutputBid {
	return &controller.BostonWebOutputBid{
		Player: b.Player,
		Level:  int(b.Level),
		Name:   domain.BostonBidName(b.Level),
		Suit:   b.Suit,
	}
}

// bostonBidOptions は序列表そのものを返す。
//
// **フロントで並べ直さない。**ミゼールがトリック宣言の間に挟まる序列なので、
// クライアント側で組み直すと必ずずれる。
func bostonBidOptions() []*controller.BostonWebOutputBidOption {
	out := make([]*controller.BostonWebOutputBidOption, 0, int(domain.BostonBidLevelCount)-1)
	for l := domain.BostonBidFive; l < domain.BostonBidLevelCount; l++ {
		out = append(out, &controller.BostonWebOutputBidOption{
			Level:          int(l),
			Name:           domain.BostonBidName(l),
			Kind:           int(domain.BostonBidKindOf(l)),
			Tricks:         domain.BostonBidTricks(l),
			NeedsTrump:     domain.BostonBidNeedsTrump(l),
			Exposed:        domain.BostonBidIsExposed(l),
			CanCallPartner: domain.BostonBidCanCallPartner(l),
			Payout:         domain.BostonBidPayout(l),
		})
	}
	return out
}

// bostonWebReveal は全員の手札を公開する局面かを返す。
func bostonWebReveal(g interfaces.BostonGame) bool {
	phase := g.GetPhase()
	return phase == domain.BostonPhaseHandEnd || phase == domain.BostonPhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BostonWebPresenter) buildPlayersOutput(g interfaces.BostonGame) []*controller.BostonWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.BostonWebOutputPlayer, 0, len(players))
	reveal := bostonWebReveal(g)
	for i := range players {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		// **on the Table の宣言では落札者の手札が全員に見える。**
		// 第 1 トリックが済んでからで、宣言と同時ではない。
		exposedNow := g.IsExposed() && i == g.GetDeclarerIdx() && g.GetTrickNumber() >= 1
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		if player.GetIsHuman() || reveal || exposedNow {
			for j := range player.GetCardsSize() {
				if c := cardToOutput(player.GetCard(j)); c != nil {
					cards = append(cards, c)
				}
			}
		}
		out = append(out, &controller.BostonWebOutputPlayer{
			ID:             i,
			IsHuman:        player.GetIsHuman(),
			CardCount:      player.GetCardsSize(),
			Cards:          cards,
			TricksWon:      g.GetTricksWon(i),
			Chips:          g.GetChips(i),
			IsDealer:       i == g.GetDealerIdx(),
			IsDeclarer:     i == g.GetDeclarerIdx(),
			IsPartner:      g.GetPartnerIdx() >= 0 && i == g.GetPartnerIdx(),
			IsDeclarerSide: g.BostonIsDeclarerSide(i),
			IsCurrentTurn:  g.GetPhase() == domain.BostonPhasePlay && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BostonWebPresenter) buildMessage(g interfaces.BostonGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("boston", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.BostonPhaseBid:
		return "", "boston.bidPhase", nil
	case domain.BostonPhaseCallPartner:
		return "", "boston.callPartner", nil
	case domain.BostonPhasePlay:
		return "", "boston.playPhase", nil
	case domain.BostonPhaseHandEnd:
		if g.IsBidMade() {
			return "", "boston.handMade", nil
		}
		return "", "boston.handFailed", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *BostonWebPresenter) ActionLogOutput(g interfaces.BostonGame) string {
	return actionLogOutputJSON(g)
}
