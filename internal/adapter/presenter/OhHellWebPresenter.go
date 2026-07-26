package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OhHellWebPresenter オー・ヘルWebプレゼンタークラス
type OhHellWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *OhHellWebPresenter) Output(o interfaces.OhHellGame, lastErr error) string {
	resObj := p.buildBase(o)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(o, lastErr)
	return marshalOrError(resObj)
}

// HintOutput ヒント情報をJSON出力する
func (p *OhHellWebPresenter) HintOutput(o interfaces.OhHellGame) string {
	hint := o.GetHint()
	resObj := p.buildBase(o)

	if hint != nil {
		resObj.Hint = &controller.OhHellWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *OhHellWebPresenter) ActionLogOutput(o interfaces.OhHellGame) string {
	return actionLogOutputJSON(o)
}

// buildBase 共通フィールドを構築
func (p *OhHellWebPresenter) buildBase(o interfaces.OhHellGame) *controller.OhHellWebOutput {
	resObj := new(controller.OhHellWebOutput)
	resObj.Phase = int(o.GetPhase())
	resObj.RoundNumber = o.GetRoundNumber()
	resObj.TotalRounds = o.GetTotalRounds()
	resObj.HandSize = o.GetHandSize()
	resObj.TrickNumber = o.GetTrickNumber()
	resObj.CurrentPlayerIdx = o.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = o.GetBidPlayerIdx()
	resObj.DealerIdx = o.GetDealerIdx()
	resObj.TrumpCard = cardToOutput(o.GetTrumpCard())
	resObj.TrumpSuit = o.GetTrumpSuit()
	resObj.RestrictedBid = o.GetRestrictedBid()
	resObj.GameEndFlag = o.GetGameEndFlag()
	resObj.WinnerIdx = o.GetWinnerIdx()
	resObj.LeadPlayerIdx = o.GetLeadPlayerIdx()

	cfg := o.GetConfig()
	resObj.Config = controller.OhHellWebOutputConfig{
		CpuDifficulty:  int(cfg.CpuDifficulty),
		MaxHandSize:    cfg.MaxHandSize,
		ScoringVariant: int(cfg.ScoringVariant),
		RoundDirection: int(cfg.RoundDirection),
	}

	resObj.CurrentTrick = p.buildTrickOutput(o.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(o)
	return resObj
}

// buildTrickOutput 現在のトリック情報を構築
func (p *OhHellWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.OhHellWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.OhHellWebOutputTrickCard {
		return &controller.OhHellWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *OhHellWebPresenter) buildPlayersOutput(o interfaces.OhHellGame) []*controller.OhHellWebOutputPlayer {
	out := make([]*controller.OhHellWebOutputPlayer, 0)
	for i := 0; i < o.GetPlayerCnt(); i++ {
		player := o.GetPlayer(i)
		pObj := &controller.OhHellWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			Bid:             player.GetBid(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *OhHellWebPresenter) buildMessage(o interfaces.OhHellGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if o.GetGameEndFlag() {
		winnerIdx := o.GetWinnerIdx()
		player := o.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("ohhell", winnerIdx, isHuman)
	}
	switch o.GetPhase() {
	case domain.OhHellPhaseBid:
		return "", "ohhell.bidPhase", nil
	case domain.OhHellPhasePlay:
		if len(o.GetCurrentTrick()) == 0 {
			return "", "ohhell.playPhase.lead", nil
		}
		return "", "ohhell.playPhase.follow", nil
	case domain.OhHellPhaseTrickEnd:
		return "", "ohhell.trickEnd", nil
	case domain.OhHellPhaseRoundEnd:
		return "", "ohhell.roundEnd", nil
	}
	return "", "", nil
}
