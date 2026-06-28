package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AllFoursWebPresenter All Fours Webプレゼンタークラス
type AllFoursWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *AllFoursWebPresenter) Output(s interfaces.AllFoursGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, s.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *AllFoursWebPresenter) buildBase(s interfaces.AllFoursGame) *controller.AllFoursWebOutput {
	resObj := new(controller.AllFoursWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.NonDealerIdx = s.GetNonDealerIdx()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.TurnUp = cardToOutput(s.GetTurnUp())
	resObj.RunCount = s.GetRunCount()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()

	cfg := s.GetConfig()
	resObj.Config = controller.AllFoursWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = p.buildTrickOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)

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
func (p *AllFoursWebPresenter) buildTrickOutput(trick []*domain.AllFoursTrickCard) []*controller.AllFoursWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.AllFoursTrickCard) *controller.AllFoursWebOutputTrickCard {
		return &controller.AllFoursWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *AllFoursWebPresenter) buildPlayersOutput(s interfaces.AllFoursGame) []*controller.AllFoursWebOutputPlayer {
	out := make([]*controller.AllFoursWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.AllFoursWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *AllFoursWebPresenter) buildMessage(s interfaces.AllFoursGame, trick []*domain.AllFoursTrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		winnerIdx := s.GetWinnerIdx()
		player := s.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("allfours", winnerIdx, isHuman)
	}
	switch s.GetPhase() {
	case domain.AllFoursPhaseBeg:
		return "", "allfours.begPhase", nil
	case domain.AllFoursPhaseGift:
		return "", "allfours.giftPhase", nil
	case domain.AllFoursPhasePlay:
		if len(trick) == 0 {
			return "", "allfours.playPhase.lead", nil
		}
		return "", "allfours.playPhase.follow", nil
	case domain.AllFoursPhaseTrickEnd:
		return "", "allfours.trickEnd", nil
	case domain.AllFoursPhaseRoundEnd:
		return "", "allfours.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *AllFoursWebPresenter) HintOutput(s interfaces.AllFoursGame) string {
	hint := s.GetHint()
	resObj := p.buildBase(s)
	if hint != nil {
		resObj.Hint = &controller.AllFoursWebOutputHint{
			CardIndex: hint.CardIndex,
			Beg:       hint.Beg,
			Run:       hint.Run,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *AllFoursWebPresenter) ActionLogOutput(s interfaces.AllFoursGame) string {
	return actionLogOutputJSON(s)
}
