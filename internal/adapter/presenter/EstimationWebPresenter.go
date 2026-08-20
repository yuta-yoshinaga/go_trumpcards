//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// EstimationWebPresenter エスティメーションWebプレゼンタークラス
type EstimationWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *EstimationWebPresenter) Output(e interfaces.EstimationGame, lastErr error) string {
	resObj := p.buildBase(e)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(e, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := e.GetHint(); hint != nil {
		resObj.Hint = &controller.EstimationWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Value: hint.Value,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *EstimationWebPresenter) buildBase(e interfaces.EstimationGame) *controller.EstimationWebOutput {
	resObj := new(controller.EstimationWebOutput)
	resObj.Phase = int(e.GetPhase())
	resObj.RoundNumber = e.GetRoundNumber()
	resObj.TrickNumber = e.GetTrickNumber()
	resObj.TrumpSuit = e.GetTrumpSuit()
	// **禁止値をワイヤに載せる。** 載せないとクライアントは押せない宣言を
	// 出してしまい、サーバに拒否されるまで分からない。
	resObj.RestrictedBid = e.GetRestrictedBid()
	resObj.CurrentPlayerIdx = e.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = e.GetBidPlayerIdx()
	resObj.LeadPlayerIdx = e.GetLeadPlayerIdx()
	resObj.DealerIdx = e.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(e.GetValidPlayIndices(0))
	resObj.GameEndFlag = e.GetGameEndFlag()
	resObj.WinnerIdx = e.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(e.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(e)
	resObj.Config = controller.EstimationWebOutputConfig{Rounds: e.GetConfig().Rounds}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *EstimationWebPresenter) buildPlayersOutput(e interfaces.EstimationGame) []*controller.EstimationWebOutputPlayer {
	out := make([]*controller.EstimationWebOutputPlayer, 0)
	for i := 0; i < e.GetPlayerCnt(); i++ {
		player := e.GetPlayer(i)
		out = append(out, &controller.EstimationWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Bid:        player.GetBid(),
			CallType:   int(player.GetCallType()),
			TrickCount: player.GetTrickCount(),
			RoundScore: player.GetRoundScore(),
			TotalScore: player.GetTotalScore(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *EstimationWebPresenter) buildMessage(e interfaces.EstimationGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if e.GetGameEndFlag() {
		if e.GetWinnerIdx() == 0 {
			return "", "estimation.result.win", map[string]string{"score": strconv.Itoa(e.GetPlayer(0).GetTotalScore())}
		}
		if e.GetWinnerIdx() < 0 {
			return "", "estimation.result.tie", nil
		}
		return "", "estimation.result.lose", map[string]string{"idx": strconv.Itoa(e.GetWinnerIdx())}
	}
	switch e.GetPhase() {
	case domain.EstimationPhaseTrump:
		if e.IsHumanTrumpTurn() {
			return "", "estimation.trump.choose", nil
		}
		return "", "estimation.trump.wait", nil
	case domain.EstimationPhaseBid:
		// **最後の宣言者だけ選べない数がある。** 案内を変えないと、押せない
		// 宣言を出してから拒否されることになる。
		if r := e.GetRestrictedBid(); r >= 0 && e.IsHumanBidTurn() {
			return "", "estimation.bid.restricted", map[string]string{"n": strconv.Itoa(r)}
		}
		return "", "estimation.bid.choose", nil
	case domain.EstimationPhaseRoundEnd:
		return "", "estimation.roundEnd", map[string]string{
			"round": strconv.Itoa(e.GetRoundNumber()),
			"score": strconv.Itoa(e.GetPlayer(0).GetRoundScore()),
		}
	}
	return "", "estimation.play", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *EstimationWebPresenter) HintOutput(e interfaces.EstimationGame) string {
	resObj := p.buildBase(e)
	if hint := e.GetHint(); hint != nil {
		resObj.Hint = &controller.EstimationWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Value: hint.Value,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *EstimationWebPresenter) ActionLogOutput(e interfaces.EstimationGame) string {
	return actionLogOutputJSON(e)
}
