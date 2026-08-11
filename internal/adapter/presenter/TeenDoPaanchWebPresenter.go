//go:build !js || !wasm || solo

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TeenDoPaanchWebPresenter 3-2-5 Webプレゼンタークラス
type TeenDoPaanchWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TeenDoPaanchWebPresenter) Output(g interfaces.TeenDoPaanchGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.TeenDoPaanchWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Suit: hint.Suit,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TeenDoPaanchWebPresenter) buildBase(g interfaces.TeenDoPaanchGame) *controller.TeenDoPaanchWebOutput {
	resObj := new(controller.TeenDoPaanchWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.FivePlayerIdx = g.GetFivePlayerIdx()
	resObj.LastExchange = g.GetLastExchange()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.ValidPlays = intSliceOrEmpty(g.GetValidPlayIndices(0))
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	resObj.Config = controller.TeenDoPaanchWebOutputConfig{Rounds: g.GetConfig().Rounds}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TeenDoPaanchWebPresenter) buildPlayersOutput(g interfaces.TeenDoPaanchGame) []*controller.TeenDoPaanchWebOutputPlayer {
	out := make([]*controller.TeenDoPaanchWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		out = append(out, &controller.TeenDoPaanchWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Target:     player.GetTarget(),
			TrickCount: player.GetTrickCount(),
			Met:        player.GetMet(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TeenDoPaanchWebPresenter) buildMessage(g interfaces.TeenDoPaanchGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		params := map[string]string{"idx": strconv.Itoa(g.GetWinnerIdx())}
		switch g.GetWinnerIdx() {
		case 0:
			return "", "teendopaanch.result.you", params
		case -1:
			return "", "teendopaanch.result.tie", params
		default:
			return "", "teendopaanch.result.cpu", params
		}
	}
	switch g.GetPhase() {
	case domain.TeenDoPaanchPhaseTrump:
		// **手札が揃う前に決めるのが賭けどころ。** 何枚見えているかを言う。
		params := map[string]string{"seen": strconv.Itoa(domain.TeenDoPaanchFirstDeal)}
		if g.IsHumanTrumpTurn() {
			return "", "teendopaanch.trump.choose", params
		}
		return "", "teendopaanch.trump.wait", params
	case domain.TeenDoPaanchPhaseRoundEnd:
		return "", "teendopaanch.roundEnd", map[string]string{
			"round": strconv.Itoa(g.GetRoundNumber()),
			"total": strconv.Itoa(g.GetConfig().Rounds),
		}
	default:
		// **多く取ってもうれしくない。** ノルマちょうどが目標だと毎回言う。
		return "", "teendopaanch.play", map[string]string{
			"target": strconv.Itoa(g.GetPlayer(0).GetTarget()),
			"took":   strconv.Itoa(g.GetPlayer(0).GetTrickCount()),
		}
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *TeenDoPaanchWebPresenter) HintOutput(g interfaces.TeenDoPaanchGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.TeenDoPaanchWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Suit: hint.Suit,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TeenDoPaanchWebPresenter) ActionLogOutput(g interfaces.TeenDoPaanchGame) string {
	return actionLogOutputJSON(g)
}
