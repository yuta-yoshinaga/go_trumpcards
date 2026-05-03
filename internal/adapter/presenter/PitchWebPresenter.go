package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PitchWebPresenter ピッチWebプレゼンタークラス
type PitchWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PitchWebPresenter) Output(s interfaces.PitchGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, s.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *PitchWebPresenter) buildBase(s interfaces.PitchGame) *controller.PitchWebOutput {
	resObj := new(controller.PitchWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = s.GetBidPlayerIdx()
	resObj.CurrentBid = s.GetCurrentBid()
	resObj.BidWinnerIdx = s.GetBidWinnerIdx()
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()

	cfg := s.GetConfig()
	resObj.Config = controller.PitchWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = p.buildTrickOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)

	// human の有効プレイインデックスを供給
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if pl := s.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			resObj.ValidPlayIndices = s.GetValidPlayIndices(i)
			break
		}
	}
	if resObj.ValidPlayIndices == nil {
		resObj.ValidPlayIndices = make([]int, 0)
	}
	return resObj
}

// buildTrickOutput 現在のトリック情報を構築
func (p *PitchWebPresenter) buildTrickOutput(trick []*domain.PitchTrickCard) []*controller.PitchWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.PitchTrickCard) *controller.PitchWebOutputTrickCard {
		return &controller.PitchWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PitchWebPresenter) buildPlayersOutput(s interfaces.PitchGame) []*controller.PitchWebOutputPlayer {
	out := make([]*controller.PitchWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := &controller.PitchWebOutputPlayer{
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
func (p *PitchWebPresenter) buildMessage(s interfaces.PitchGame, trick []*domain.PitchTrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		winnerIdx := s.GetWinnerIdx()
		player := s.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("pitch", winnerIdx, isHuman)
	}
	switch s.GetPhase() {
	case domain.PitchPhaseBid:
		return "", "pitch.bidPhase", nil
	case domain.PitchPhasePlay:
		if len(trick) == 0 {
			return "", "pitch.playPhase.lead", nil
		}
		return "", "pitch.playPhase.follow", nil
	case domain.PitchPhaseTrickEnd:
		return "", "pitch.trickEnd", nil
	case domain.PitchPhaseRoundEnd:
		return "", "pitch.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *PitchWebPresenter) HintOutput(s interfaces.PitchGame) string {
	hint := s.GetHint()
	resObj := p.buildBase(s)
	if hint != nil {
		resObj.Hint = &controller.PitchWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *PitchWebPresenter) ActionLogOutput(s interfaces.PitchGame) string {
	return actionLogOutputJSON(s)
}
