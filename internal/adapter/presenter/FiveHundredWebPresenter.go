//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FiveHundredWebPresenter 500 Webプレゼンタークラス
type FiveHundredWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FiveHundredWebPresenter) Output(g interfaces.FiveHundredGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *FiveHundredWebPresenter) buildBase(g interfaces.FiveHundredGame) *controller.FiveHundredWebOutput {
	resObj := new(controller.FiveHundredWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.ContractKind = g.GetContractKind()
	resObj.ContractTricks = g.GetContractTricks()
	resObj.ContractValue = g.GetContractValue()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.HighestBid = bidToOutput(g.GetHighestBid())
	resObj.HighestBidder = g.GetHighestBidder()
	resObj.JokerLeadSuit = g.GetJokerLeadSuit()
	resObj.KittyCount = len(g.GetKitty())
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	cfg := g.GetConfig()
	resObj.Config = controller.FiveHundredWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// bidToOutput ビッドをWeb出力へ変換 (nil 安全)
func bidToOutput(b *domain.FiveHundredBid) *controller.FiveHundredWebOutputBid {
	if b == nil {
		return nil
	}
	return &controller.FiveHundredWebOutputBid{
		Kind:   int(b.Kind),
		Tricks: b.Tricks,
		Suit:   b.Suit,
		Value:  b.Value(),
	}
}

// buildPlayersOutput プレイヤー情報を構築。オープンミゼールの落札者は手札を公開する。
func (p *FiveHundredWebPresenter) buildPlayersOutput(g interfaces.FiveHundredGame) []*controller.FiveHundredWebOutputPlayer {
	out := make([]*controller.FiveHundredWebOutputPlayer, 0)
	openMisereDeclarer := g.GetContractKind() == int(domain.FiveHundredContractOpenMisere)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		reveal := player.GetIsHuman() || (openMisereDeclarer && player.GetIsDeclarer())
		out = append(out, &controller.FiveHundredWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, reveal),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
			Bid:        bidToOutput(player.GetBid()),
			Passed:     player.GetPassed(),
			IsDeclarer: player.GetIsDeclarer(),
		})
	}
	return out
}

// buildMessage ゲーム結果/フェーズメッセージを構築
func (p *FiveHundredWebPresenter) buildMessage(g interfaces.FiveHundredGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("fivehundred.result.team%dWin", winnerTeam)
		return msg, code, map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	}
	switch g.GetPhase() {
	case domain.FiveHundredPhaseBid:
		return "", "fivehundred.bidPhase", nil
	case domain.FiveHundredPhaseKittyExchange:
		return "", "fivehundred.kittyExchangePhase", nil
	case domain.FiveHundredPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "fivehundred.playPhase.lead", nil
		}
		return "", "fivehundred.playPhase.follow", nil
	case domain.FiveHundredPhaseTrickEnd:
		return "", "fivehundred.trickEnd", nil
	case domain.FiveHundredPhaseRoundEnd:
		return "", "fivehundred.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *FiveHundredWebPresenter) HintOutput(g interfaces.FiveHundredGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.FiveHundredWebOutputHint{
			BidKind:        hint.BidKind,
			BidTricks:      hint.BidTricks,
			BidSuit:        hint.BidSuit,
			Pass:           hint.Pass,
			DiscardIndices: hint.DiscardIndices,
			CardIndex:      hint.CardIndex,
			JokerSuit:      hint.JokerSuit,
			Reason:         hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FiveHundredWebPresenter) ActionLogOutput(g interfaces.FiveHundredGame) string {
	return actionLogOutputJSON(g)
}
