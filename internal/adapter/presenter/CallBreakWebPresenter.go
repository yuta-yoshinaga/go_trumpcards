package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CallBreakWebPresenter Call Break Web プレゼンタークラス
type CallBreakWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *CallBreakWebPresenter) Output(cb interfaces.CallBreakGame, lastErr error) string {
	resObj := p.buildBase(cb)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(cb, cb.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *CallBreakWebPresenter) buildBase(cb interfaces.CallBreakGame) *controller.CallBreakWebOutput {
	resObj := new(controller.CallBreakWebOutput)
	resObj.Phase = int(cb.GetPhase())
	resObj.RoundNumber = cb.GetRoundNumber()
	resObj.TrickNumber = cb.GetTrickNumber()
	resObj.CurrentPlayerIdx = cb.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = cb.GetBidPlayerIdx()
	resObj.SpadesBroken = cb.GetSpadesBroken()
	resObj.GameEndFlag = cb.GetGameEndFlag()
	resObj.WinnerIdx = cb.GetWinnerIdx()
	resObj.LeadPlayerIdx = cb.GetLeadPlayerIdx()

	cfg := cb.GetConfig()
	resObj.Config = controller.CallBreakWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		MaxRounds:     cfg.MaxRounds,
	}

	resObj.CurrentTrick = trickCardsToOutput(cb.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(cb)

	// Provide the human player's valid play indices so the frontend can grey out
	// cards that the Call Break rules (lead-suit / must-trump-spade) forbid.
	for i := 0; i < cb.GetPlayerCnt(); i++ {
		if pl := cb.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			resObj.ValidPlayIndices = cb.GetValidPlayIndices(i)
			break
		}
	}
	if resObj.ValidPlayIndices == nil {
		resObj.ValidPlayIndices = make([]int, 0)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CallBreakWebPresenter) buildPlayersOutput(cb interfaces.CallBreakGame) []*controller.CallBreakWebOutputPlayer {
	out := make([]*controller.CallBreakWebOutputPlayer, 0)
	for i := 0; i < cb.GetPlayerCnt(); i++ {
		player := cb.GetPlayer(i)
		pObj := &controller.CallBreakWebOutputPlayer{
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
func (p *CallBreakWebPresenter) buildMessage(cb interfaces.CallBreakGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if cb.GetGameEndFlag() {
		winnerIdx := cb.GetWinnerIdx()
		player := cb.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("callbreak", winnerIdx, isHuman)
	}
	switch cb.GetPhase() {
	case domain.CallBreakPhaseBid:
		return "", "callbreak.bidPhase", nil
	case domain.CallBreakPhasePlay:
		if len(trick) == 0 {
			return "", "callbreak.playPhase.lead", nil
		}
		return "", "callbreak.playPhase.follow", nil
	case domain.CallBreakPhaseTrickEnd:
		return "", "callbreak.trickEnd", nil
	case domain.CallBreakPhaseRoundEnd:
		return "", "callbreak.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報を JSON 出力する
func (p *CallBreakWebPresenter) HintOutput(cb interfaces.CallBreakGame) string {
	hint := cb.GetHint()
	resObj := p.buildBase(cb)
	if hint != nil {
		resObj.Hint = &controller.CallBreakWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON 出力
func (p *CallBreakWebPresenter) ActionLogOutput(cb interfaces.CallBreakGame) string {
	return actionLogOutputJSON(cb)
}
