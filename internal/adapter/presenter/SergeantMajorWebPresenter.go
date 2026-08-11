//go:build !js || !wasm || extra

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SergeantMajorWebPresenter サージェントメジャーWebプレゼンタークラス
type SergeantMajorWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SergeantMajorWebPresenter) Output(s interfaces.SergeantMajorGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.SergeantMajorWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Suit: hint.Suit,
			Indices: intSliceOrEmpty(hint.Indices),
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SergeantMajorWebPresenter) buildBase(s interfaces.SergeantMajorGame) *controller.SergeantMajorWebOutput {
	resObj := new(controller.SergeantMajorWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.KittySize = s.GetKittySize()
	resObj.DiscardCount = s.GetDiscardCount()
	resObj.LastExchange = s.GetLastExchange()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(0))
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	resObj.Config = controller.SergeantMajorWebOutputConfig{Rounds: s.GetConfig().Rounds}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SergeantMajorWebPresenter) buildPlayersOutput(s interfaces.SergeantMajorGame) []*controller.SergeantMajorWebOutputPlayer {
	out := make([]*controller.SergeantMajorWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.SergeantMajorWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Target:     player.GetTarget(),
			TrickCount: player.GetTrickCount(),
			Score:      player.GetScore(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SergeantMajorWebPresenter) buildMessage(s interfaces.SergeantMajorGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		params := map[string]string{"idx": strconv.Itoa(s.GetWinnerIdx())}
		switch s.GetWinnerIdx() {
		case 0:
			return "", "sergeantmajor.result.you", params
		case -1:
			return "", "sergeantmajor.result.tie", params
		default:
			return "", "sergeantmajor.result.cpu", params
		}
	}
	switch s.GetPhase() {
	case domain.SergeantMajorPhaseTrump:
		// **親だけが決める。** ノルマ 8 と切り札宣言は同じ席。
		if s.IsHumanTrumpTurn() {
			return "", "sergeantmajor.trump.choose", nil
		}
		return "", "sergeantmajor.trump.wait", nil
	case domain.SergeantMajorPhaseDiscard:
		params := map[string]string{
			"kitty": strconv.Itoa(domain.SergeantMajorKittySize),
			"n":     strconv.Itoa(s.GetDiscardCount()),
		}
		if s.IsHumanDiscardTurn() {
			return "", "sergeantmajor.discard.choose", params
		}
		return "", "sergeantmajor.discard.wait", params
	case domain.SergeantMajorPhaseRoundEnd:
		return "", "sergeantmajor.roundEnd", map[string]string{
			"round": strconv.Itoa(s.GetRoundNumber()),
			"total": strconv.Itoa(s.GetConfig().Rounds),
		}
	default:
		// **多く取れば取っただけ得。** ノルマとの差がそのまま得点。
		return "", "sergeantmajor.play", map[string]string{
			"target": strconv.Itoa(s.GetPlayer(0).GetTarget()),
			"took":   strconv.Itoa(s.GetPlayer(0).GetTrickCount()),
		}
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *SergeantMajorWebPresenter) HintOutput(s interfaces.SergeantMajorGame) string {
	resObj := p.buildBase(s)
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.SergeantMajorWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Suit: hint.Suit,
			Indices: intSliceOrEmpty(hint.Indices),
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SergeantMajorWebPresenter) ActionLogOutput(s interfaces.SergeantMajorGame) string {
	return actionLogOutputJSON(s)
}
