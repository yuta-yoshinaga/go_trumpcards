//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BidWhistWebPresenter Bid Whist Webプレゼンタークラス
type BidWhistWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BidWhistWebPresenter) Output(g interfaces.BidWhistGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BidWhistWebPresenter) buildBase(g interfaces.BidWhistGame) *controller.BidWhistWebOutput {
	resObj := new(controller.BidWhistWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.ContractTricks = g.GetContractTricks()
	resObj.ContractDirection = g.GetContractDirection()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.HighestBid = bidWhistBidToOutput(g.GetHighestBid())
	resObj.HighestBidder = g.GetHighestBidder()
	resObj.KittyCount = len(g.GetKitty())
	resObj.KittyIndices = bidWhistKittyIndicesForHuman(g)
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	cfg := g.GetConfig()
	resObj.Config = controller.BidWhistWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// bidWhistKittyIndicesForHuman はキティ交換フェーズで人間が落札者のときのみ、
// 人間の手札のうちキティ由来カードのインデックスを返す。それ以外は空スライスを返し、
// CPU 落札者のキティ内容を人間へ漏らさない。返り値は常に非 nil。
func bidWhistKittyIndicesForHuman(g interfaces.BidWhistGame) []int {
	declarerIdx := g.GetDeclarerIdx()
	if declarerIdx < 0 || declarerIdx >= g.GetPlayerCnt() || !g.GetPlayer(declarerIdx).GetIsHuman() {
		return []int{}
	}
	return g.GetKittyIndices()
}

// bidWhistBidToOutput ビッドをWeb出力へ変換 (nil 安全)
func bidWhistBidToOutput(b *domain.BidWhistBid) *controller.BidWhistWebOutputBid {
	if b == nil {
		return nil
	}
	return &controller.BidWhistWebOutputBid{
		Tricks:    b.Tricks,
		Direction: b.Direction,
	}
}

// buildTrickOutput 現在のトリック情報を構築
func (p *BidWhistWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.BidWhistWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.BidWhistWebOutputTrickCard {
		return &controller.BidWhistWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築 (人間の手札のみ公開)
func (p *BidWhistWebPresenter) buildPlayersOutput(g interfaces.BidWhistGame) []*controller.BidWhistWebOutputPlayer {
	out := make([]*controller.BidWhistWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		out = append(out, &controller.BidWhistWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
			Bid:        bidWhistBidToOutput(player.GetBid()),
			Passed:     player.GetPassed(),
			IsDeclarer: player.GetIsDeclarer(),
		})
	}
	return out
}

// buildMessage ゲーム結果/フェーズメッセージを構築
func (p *BidWhistWebPresenter) buildMessage(g interfaces.BidWhistGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("bidwhist.result.team%dWin", winnerTeam)
		return msg, code, map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	}
	switch g.GetPhase() {
	case domain.BidWhistPhaseBid:
		return "", "bidwhist.bidPhase", nil
	case domain.BidWhistPhaseTrumpDeclaration:
		return "", "bidwhist.trumpPhase", nil
	case domain.BidWhistPhaseKittyExchange:
		return "", "bidwhist.kittyExchangePhase", nil
	case domain.BidWhistPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "bidwhist.playPhase.lead", nil
		}
		return "", "bidwhist.playPhase.follow", nil
	case domain.BidWhistPhaseTrickEnd:
		return "", "bidwhist.trickEnd", nil
	case domain.BidWhistPhaseRoundEnd:
		return "", "bidwhist.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BidWhistWebPresenter) HintOutput(g interfaces.BidWhistGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.BidWhistWebOutputHint{
			BidTricks:      hint.BidTricks,
			BidDirection:   hint.BidDirection,
			Pass:           hint.Pass,
			TrumpSuit:      hint.TrumpSuit,
			DiscardIndices: hint.DiscardIndices,
			CardIndex:      hint.CardIndex,
			Reason:         hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BidWhistWebPresenter) ActionLogOutput(g interfaces.BidWhistGame) string {
	return actionLogOutputJSON(g)
}
