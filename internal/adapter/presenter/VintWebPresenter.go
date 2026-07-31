//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// VintWebPresenter ヴィント Webプレゼンタークラス
type VintWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *VintWebPresenter) Output(g interfaces.VintGame, lastErr error) string {
	resObj := new(controller.VintWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TrickLeaderIdx = g.GetTrickLeaderIdx()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.GameTarget = domain.VintGameTarget
	resObj.MinLevel = domain.VintMinLevel
	resObj.MaxLevel = domain.VintMaxLevel
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	for team := range domain.VintTeamCnt {
		resObj.TeamTricks[team] = g.VintTeamTricks(team)
		resObj.Below[team] = g.GetBelow(team)
		resObj.Above[team] = g.GetAbove(team)
		resObj.GamesWon[team] = g.GetGamesWon(team)
	}
	// **単価はスートとレベルの両方で決まる。**基準値を送り、レベル分はフロントで
	// +10 して見せられるようにする。
	for denom := range domain.VintDenomCount {
		resObj.TrickValues[denom] = domain.VintTrickValue(denom, domain.VintMinLevel)
	}

	trick := g.GetTrick()
	resObj.Trick = make([]*controller.WebOutputCard, 0, len(trick))
	for _, c := range trick {
		if out := cardToOutput(c); out != nil {
			resObj.Trick = append(resObj.Trick, out)
		}
	}

	bids := g.GetBids()
	resObj.Bids = make([]*controller.VintWebOutputBid, 0, len(bids))
	for _, b := range bids {
		if b == nil {
			continue
		}
		resObj.Bids = append(resObj.Bids, vintBidOut(b))
	}
	if hb := g.GetHighBid(); hb != nil {
		resObj.HighBid = vintBidOut(hb)
	}

	if r := g.GetLastResult(); r != nil {
		resObj.LastResult = &controller.VintWebOutputResult{
			TrickPoints:    r.TrickPoints,
			HonourPoints:   r.HonourPoints,
			AcePoints:      r.AcePoints,
			Penalty:        r.Penalty,
			Made:           r.Made,
			DeclarerTricks: r.DeclarerTricks,
			TrickValue:     r.TrickValue,
		}
	}

	// **出せる札はサーバーが決める。**追随が強制なのでフロントで再現するとずれる。
	resObj.ValidPlays = make([]int, 0)
	if g.GetPhase() == domain.VintPhasePlay && g.IsHumanTurn() {
		resObj.ValidPlays = append(resObj.ValidPlays, g.VintValidPlays(g.GetCurrentPlayerIdx())...)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.VintWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// vintBidOut は 1 件の宣言をワイヤ表現へ変換する。
func vintBidOut(b *domain.VintBid) *controller.VintWebOutputBid {
	return &controller.VintWebOutputBid{
		Player:     b.Player,
		Level:      b.Level,
		Denom:      b.Denom,
		TrickValue: domain.VintTrickValue(b.Denom, b.Level),
	}
}

// vintWebReveal は全員の手札を公開する局面かを返す。
func vintWebReveal(g interfaces.VintGame) bool {
	phase := g.GetPhase()
	return phase == domain.VintPhaseHandEnd || phase == domain.VintPhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *VintWebPresenter) buildPlayersOutput(g interfaces.VintGame) []*controller.VintWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.VintWebOutputPlayer, 0, len(players))
	reveal := vintWebReveal(g)
	for i := range players {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		// **ダミーが無いので、プレイ中は誰の手札も公開しない。**
		// ブリッジと違い、落札者の相方の手札も伏せたままである。
		if player.GetIsHuman() || reveal {
			for j := range player.GetCardsSize() {
				if c := cardToOutput(player.GetCard(j)); c != nil {
					cards = append(cards, c)
				}
			}
		}
		out = append(out, &controller.VintWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Team:          domain.VintTeamOf(i),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			TricksWon:     g.GetTricksWon(i),
			IsDealer:      i == g.GetDealerIdx(),
			IsDeclarer:    i == g.GetDeclarerIdx(),
			IsCurrentTurn: g.GetPhase() == domain.VintPhasePlay && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *VintWebPresenter) buildMessage(g interfaces.VintGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		// **チーム戦なので勝敗は席ではなくチームで見る。**人間は席 0 = チーム 0。
		if g.GetWinnerTeam() == domain.VintTeamOf(0) {
			return "your team takes the rubber", "vint.result.humanWin", nil
		}
		return "the other team takes the rubber", "vint.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.VintPhaseBid:
		return "", "vint.bidPhase", nil
	case domain.VintPhasePlay:
		return "", "vint.playPhase", nil
	case domain.VintPhaseHandEnd:
		if r := g.GetLastResult(); r != nil && !r.Made {
			return "", "vint.handSet", nil
		}
		return "", "vint.handMade", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *VintWebPresenter) ActionLogOutput(g interfaces.VintGame) string {
	return actionLogOutputJSON(g)
}
