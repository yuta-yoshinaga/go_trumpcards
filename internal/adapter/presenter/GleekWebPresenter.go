//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GleekWebPresenter グリーク (Gleek) のWebプレゼンタークラス
type GleekWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GleekWebPresenter) Output(g interfaces.GleekGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。** HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *GleekWebPresenter) buildBase(g interfaces.GleekGame) *controller.GleekWebOutput {
	resObj := new(controller.GleekWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.CurrentBidderIdx = g.GetCurrentBidderIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ElderIdx = g.GetElderIdx()
	resObj.BuyerIdx = g.GetBuyerIdx()
	resObj.WinningBid = g.GetWinningBid()
	resObj.HighestBid = g.HighestBid()
	resObj.NextBidAmount = g.NextBidAmount()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TurnUp = cardToOutput(g.GetTurnUp())
	resObj.DiscardCount = domain.GleekSwapSize
	resObj.RuffWinnerIdx = g.GetRuffWinnerIdx()
	resObj.DealPoints = g.DealPoints()
	resObj.Par = g.Par()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()
	resObj.IsHumanDiscardTurn = g.IsHumanDiscardTurn()
	resObj.PlayableIndices = p.playableIndices(g)
	resObj.Melds = p.buildMelds(g)

	cfg := g.GetConfig()
	resObj.Config = controller.GleekWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildMelds 申告されたメルドを出力形式に変換する。
func (p *GleekWebPresenter) buildMelds(g interfaces.GleekGame) []*controller.GleekWebOutputMeld {
	out := make([]*controller.GleekWebOutputMeld, 0)
	for _, m := range g.GetMelds() {
		if m == nil {
			continue
		}
		out = append(out, &controller.GleekWebOutputMeld{
			PlayerIdx: m.PlayerIdx, Rank: m.Rank, Count: m.Count, Value: m.Value,
		})
	}
	return out
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *GleekWebPresenter) playableIndices(g interfaces.GleekGame) []int {
	if g.GetPhase() != domain.GleekPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *GleekWebPresenter) buildPlayersOutput(g interfaces.GleekGame) []*controller.GleekWebOutputPlayer {
	scores := g.GetPlayerScores()
	bids := g.GetBids()
	passed := g.GetPassed()
	trickPoints := g.GetTrickPoints()
	ruffs := g.GetRuffs()
	buyer := g.GetBuyerIdx()
	out := make([]*controller.GleekWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		entry := &controller.GleekWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount:  player.GetTrickCount(),
			Score:       scores[i],
			IsBuyer:     i == buyer,
			Bid:         bids[i],
			Passed:      passed[i],
			TrickPoints: trickPoints[i],
			RuffSuit:    -1,
		}
		if i < len(ruffs) && ruffs[i] != nil {
			entry.Ruff = ruffs[i].Total
			entry.RuffSuit = ruffs[i].Suit
		}
		out = append(out, entry)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *GleekWebPresenter) buildMessage(g interfaces.GleekGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.GleekPhaseBid:
		return "", "gleek.bidPhase", nil
	case domain.GleekPhaseDiscard:
		return "", "gleek.discardPhase", nil
	case domain.GleekPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "gleek.playPhase.lead", nil
		}
		return "", "gleek.playPhase.follow", nil
	case domain.GleekPhaseTrickEnd:
		return "", "gleek.trickEnd", nil
	case domain.GleekPhaseRoundEnd:
		return "", "gleek.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *GleekWebPresenter) winnerMessage(g interfaces.GleekGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if winner >= 0 && winner == humanIdx {
		return "", "gleek.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "gleek.result.cpuWin", params
}

// HintOutput ヒント専用のレスポンスを返す。
func (p *GleekWebPresenter) HintOutput(g interfaces.GleekGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
		resObj.MessageCode = "gleek.hintRequested"
	} else {
		resObj.MessageCode = "gleek.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON で返す。
func (p *GleekWebPresenter) ActionLogOutput(g interfaces.GleekGame) string {
	return actionLogOutputJSON(g)
}
