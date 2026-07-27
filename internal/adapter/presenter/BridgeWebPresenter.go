//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BridgeWebPresenter ブリッジWebプレゼンタークラス
type BridgeWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BridgeWebPresenter) Output(b interfaces.BridgeGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, b.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BridgeWebPresenter) buildBase(b interfaces.BridgeGame) *controller.BridgeWebOutput {
	resObj := new(controller.BridgeWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.RoundNumber = b.GetRoundNumber()
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = b.GetBidPlayerIdx()
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.TrumpSuit = trumpSuitForAPI(b.GetTrumpSuit(), b.GetContractSuit())
	resObj.ContractLevel = b.GetContractLevel()
	resObj.ContractSuit = b.GetContractSuit()
	resObj.Doubled = b.GetDoubled()
	resObj.DeclarerIdx = b.GetDeclarerIdx()
	resObj.DummyIdx = b.GetDummyIdx()
	resObj.Vulnerability = [2]bool{b.GetVulnerability(0), b.GetVulnerability(1)}
	resObj.TeamScores = [2]int{b.GetTeamScore(0), b.GetTeamScore(1)}
	resObj.GamesWon = [2]int{b.GetGamesWon(0), b.GetGamesWon(1)}
	resObj.BelowLine = [2]int{b.GetBelowLine(0), b.GetBelowLine(1)}
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerTeam = b.GetWinnerTeam()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()
	resObj.OpeningLeadDone = b.IsOpeningLeadDone()

	// ダミーの手札
	resObj.DummyHand = cardsToOutputOrEmpty(b.GetDummyHand())

	// 設定
	cfg := b.GetConfig()
	resObj.Config = controller.BridgeWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	resObj.BidHistory = p.buildBidHistoryOutput(b.GetBidHistory())
	resObj.CurrentTrick = p.buildTrickOutput(b.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(b)
	return resObj
}

// buildBidHistoryOutput ビッド履歴を構築
func (p *BridgeWebPresenter) buildBidHistoryOutput(history []*domain.BridgeBidEntry) []*controller.BridgeWebOutputBidEntry {
	out := make([]*controller.BridgeWebOutputBidEntry, 0)
	for _, entry := range history {
		out = append(out, &controller.BridgeWebOutputBidEntry{
			PlayerIdx: entry.PlayerIdx,
			BidType:   int(entry.BidType),
			Level:     entry.Level,
			Suit:      entry.Suit,
		})
	}
	return out
}

// buildTrickOutput 現在のトリック情報を構築
func (p *BridgeWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.BridgeWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.BridgeWebOutputTrickCard {
		return &controller.BridgeWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BridgeWebPresenter) buildPlayersOutput(b interfaces.BridgeGame) []*controller.BridgeWebOutputPlayer {
	out := make([]*controller.BridgeWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		pObj := &controller.BridgeWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BridgeWebPresenter) buildMessage(b interfaces.BridgeGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if b.GetGameEndFlag() {
		winnerTeam := b.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("bridge.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch b.GetPhase() {
	case domain.BridgePhaseBid:
		return "", "bridge.bidPhase", nil
	case domain.BridgePhasePlay:
		if len(trick) == 0 {
			return "", "bridge.playPhase.lead", nil
		}
		return "", "bridge.playPhase.follow", nil
	case domain.BridgePhaseTrickEnd:
		return "", "bridge.trickEnd", nil
	case domain.BridgePhaseRoundEnd:
		return "", "bridge.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BridgeWebPresenter) HintOutput(b interfaces.BridgeGame) string {
	hint := b.GetHint()
	resObj := p.buildBase(b)
	if hint != nil {
		resObj.Hint = &controller.BridgeWebOutputHint{
			CardIndex: hint.CardIndex,
			BidType:   hint.BidType,
			BidLevel:  hint.BidLevel,
			BidSuit:   hint.BidSuit,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BridgeWebPresenter) ActionLogOutput(b interfaces.BridgeGame) string {
	return actionLogOutputJSON(b)
}

// trumpSuitForAPI converts internal CardDesign-based trumpSuit to BridgeBidSuit
// enum for consistent API output. Uses contractSuit to distinguish NoTrump (-1
// internally) from "not yet determined" (-1 before auction).
// Output: 0=not determined, 1=Club, 2=Diamond, 3=Heart, 4=Spade, 5=NoTrump.
func trumpSuitForAPI(trumpSuit int, contractSuit int) int {
	if trumpSuit == -1 {
		if contractSuit == domain.BridgeBidSuitNT {
			return domain.BridgeBidSuitNT // 5 = NoTrump
		}
		return 0 // not yet determined
	}
	switch trumpSuit {
	case domain.CardDesignClover:
		return domain.BridgeBidSuitClub
	case domain.CardDesignDiamond:
		return domain.BridgeBidSuitDiamond
	case domain.CardDesignHeart:
		return domain.BridgeBidSuitHeart
	case domain.CardDesignSpade:
		return domain.BridgeBidSuitSpade
	default:
		return 0
	}
}
