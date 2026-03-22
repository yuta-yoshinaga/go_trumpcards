package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpadesWebPresenter スペードWebプレゼンタークラス
type SpadesWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SpadesWebPresenter) Output(s interfaces.SpadesGame, lastErr error) string {
	resObj := new(controller.SpadesWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = s.GetBidPlayerIdx()
	resObj.SpadesBroken = s.GetSpadesBroken()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()

	// 設定
	cfg := s.GetConfig()
	resObj.Config = controller.SpadesWebOutputConfig{
		CpuDifficulty:       int(cfg.CpuDifficulty),
		PointLimit:          cfg.PointLimit,
		NilBonus:            cfg.NilBonus,
		BagPenaltyThreshold: cfg.BagPenaltyThreshold,
	}

	trick := s.GetCurrentTrick()
	resObj.CurrentTrick = p.buildTrickOutput(trick)
	resObj.Players = p.buildPlayersOutput(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, trick, lastErr)

	return marshalOrError(resObj)
}

// buildTrickOutput 現在のトリック情報を構築
func (p *SpadesWebPresenter) buildTrickOutput(trick []*domain.SpadesTrickCard) []*controller.SpadesWebOutputTrickCard {
	out := make([]*controller.SpadesWebOutputTrickCard, 0)
	for _, tc := range trick {
		out = append(out, &controller.SpadesWebOutputTrickCard{
			PlayerIdx: tc.PlayerIdx,
			Card:      cardToOutput(tc.Card),
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SpadesWebPresenter) buildPlayersOutput(s interfaces.SpadesGame) []*controller.SpadesWebOutputPlayer {
	out := make([]*controller.SpadesWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := &controller.SpadesWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			Bid:             player.GetBid(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
			Bags:            player.GetBags(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SpadesWebPresenter) buildMessage(s interfaces.SpadesGame, trick []*domain.SpadesTrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		winnerIdx := s.GetWinnerIdx()
		player := s.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("spades", winnerIdx, isHuman)
	}
	switch s.GetPhase() {
	case domain.SpadesPhaseBid:
		return "", "spades.bidPhase", nil
	case domain.SpadesPhasePlay:
		if len(trick) == 0 {
			return "", "spades.playPhase.lead", nil
		}
		return "", "spades.playPhase.follow", nil
	case domain.SpadesPhaseTrickEnd:
		return "", "spades.trickEnd", nil
	case domain.SpadesPhaseRoundEnd:
		return "", "spades.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *SpadesWebPresenter) HintOutput(s interfaces.SpadesGame) string {
	hint := s.GetHint()
	resObj := new(controller.SpadesWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = s.GetBidPlayerIdx()
	resObj.SpadesBroken = s.GetSpadesBroken()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	cfg := s.GetConfig()
	resObj.Config = controller.SpadesWebOutputConfig{
		CpuDifficulty:       int(cfg.CpuDifficulty),
		PointLimit:          cfg.PointLimit,
		NilBonus:            cfg.NilBonus,
		BagPenaltyThreshold: cfg.BagPenaltyThreshold,
	}
	trick := s.GetCurrentTrick()
	resObj.CurrentTrick = p.buildTrickOutput(trick)
	resObj.Players = p.buildPlayersOutput(s)

	if hint != nil {
		resObj.Hint = &controller.SpadesWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SpadesWebPresenter) ActionLogOutput(s interfaces.SpadesGame) string {
	return actionLogOutputJSON(s)
}
