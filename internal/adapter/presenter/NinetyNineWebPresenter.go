package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// NinetyNineWebPresenter ナインティナインWebプレゼンタークラス
type NinetyNineWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *NinetyNineWebPresenter) Output(o interfaces.NinetyNineGame, lastErr error) string {
	resObj := p.buildBase(o)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(o, lastErr)
	return marshalOrError(resObj)
}

// HintOutput ヒント情報をJSON出力する
func (p *NinetyNineWebPresenter) HintOutput(o interfaces.NinetyNineGame) string {
	hint := o.GetHint()
	resObj := p.buildBase(o)

	if hint != nil {
		resObj.Hint = &controller.NinetyNineWebOutputHint{
			CardIndex:   hint.CardIndex,
			BuryIndices: hint.BuryIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *NinetyNineWebPresenter) ActionLogOutput(o interfaces.NinetyNineGame) string {
	return actionLogOutputJSON(o)
}

// buildBase 共通フィールドを構築
func (p *NinetyNineWebPresenter) buildBase(o interfaces.NinetyNineGame) *controller.NinetyNineWebOutput {
	resObj := new(controller.NinetyNineWebOutput)
	resObj.Phase = int(o.GetPhase())
	resObj.DealNumber = o.GetDealNumber()
	resObj.TargetScore = o.GetTargetScore()
	resObj.HandSize = o.GetHandSize()
	resObj.TrickNumber = o.GetTrickNumber()
	resObj.CurrentPlayerIdx = o.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = o.GetBidPlayerIdx()
	resObj.DealerIdx = o.GetDealerIdx()
	resObj.TrumpSuit = o.GetTrumpSuit()
	resObj.GameEndFlag = o.GetGameEndFlag()
	resObj.WinnerIdx = o.GetWinnerIdx()
	resObj.LeadPlayerIdx = o.GetLeadPlayerIdx()

	cfg := o.GetConfig()
	resObj.Config = controller.NinetyNineWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	resObj.CurrentTrick = p.buildTrickOutput(o.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(o)
	return resObj
}

// buildTrickOutput 現在のトリック情報を構築
func (p *NinetyNineWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.NinetyNineWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.NinetyNineWebOutputTrickCard {
		return &controller.NinetyNineWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *NinetyNineWebPresenter) buildPlayersOutput(o interfaces.NinetyNineGame) []*controller.NinetyNineWebOutputPlayer {
	out := make([]*controller.NinetyNineWebOutputPlayer, 0)
	for i := 0; i < o.GetPlayerCnt(); i++ {
		player := o.GetPlayer(i)
		pObj := &controller.NinetyNineWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			Bid:             player.GetBid(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
			BuriedCount:     len(player.GetBuried()),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *NinetyNineWebPresenter) buildMessage(o interfaces.NinetyNineGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if o.GetGameEndFlag() {
		winnerIdx := o.GetWinnerIdx()
		player := o.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("ninetynine", winnerIdx, isHuman)
	}
	switch o.GetPhase() {
	case domain.NinetyNinePhaseBid:
		return "", "ninetynine.bidPhase", nil
	case domain.NinetyNinePhasePlay:
		if len(o.GetCurrentTrick()) == 0 {
			return "", "ninetynine.playPhase.lead", nil
		}
		return "", "ninetynine.playPhase.follow", nil
	case domain.NinetyNinePhaseTrickEnd:
		return "", "ninetynine.trickEnd", nil
	case domain.NinetyNinePhaseRoundEnd:
		return "", "ninetynine.roundEnd", nil
	}
	return "", "", nil
}
