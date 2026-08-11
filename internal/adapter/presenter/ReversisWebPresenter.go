//go:build !js || !wasm || classic

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ReversisWebPresenter レヴェルシWebプレゼンタークラス
type ReversisWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ReversisWebPresenter) Output(r interfaces.ReversisGame, lastErr error) string {
	resObj := p.buildBase(r)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(r, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := r.GetHint(); hint != nil {
		resObj.Hint = &controller.ReversisWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *ReversisWebPresenter) buildBase(r interfaces.ReversisGame) *controller.ReversisWebOutput {
	resObj := new(controller.ReversisWebOutput)
	resObj.Phase = int(r.GetPhase())
	resObj.RoundNumber = r.GetRoundNumber()
	resObj.TrickNumber = r.GetTrickNumber()
	resObj.Pool = r.GetPool()
	resObj.CurrentPlayerIdx = r.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = r.GetLeadPlayerIdx()
	resObj.DealerIdx = r.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(r.GetValidPlayIndices(0))
	resObj.GameEndFlag = r.GetGameEndFlag()
	resObj.WinnerIdx = r.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(r.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(r)
	resObj.Config = controller.ReversisWebOutputConfig{Rounds: r.GetConfig().Rounds}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ReversisWebPresenter) buildPlayersOutput(r interfaces.ReversisGame) []*controller.ReversisWebOutputPlayer {
	out := make([]*controller.ReversisWebOutputPlayer, 0)
	for i := 0; i < r.GetPlayerCnt(); i++ {
		player := r.GetPlayer(i)
		out = append(out, &controller.ReversisWebOutputPlayer{
			ID:             i,
			IsHuman:        player.GetIsHuman(),
			CardCount:      player.GetCardsSize(),
			Cards:          playerCardsToOutput(player, player.GetIsHuman()),
			Chips:          player.GetChips(),
			RoundPenalty:   player.GetRoundPenalty(),
			TrickCount:     player.GetTrickCount(),
			TookQuinola:    player.GetTookQuinola(),
			TookDiamondAce: player.GetTookDiamondAce(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ReversisWebPresenter) buildMessage(r interfaces.ReversisGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if r.GetGameEndFlag() {
		if r.GetWinnerIdx() < 0 {
			return "", "reversis.result.tie", nil
		}
		return "", "reversis.result.winner", map[string]string{
			"idx":   strconv.Itoa(r.GetWinnerIdx()),
			"chips": strconv.Itoa(r.GetPlayer(r.GetWinnerIdx()).GetChips()),
		}
	}
	if r.GetPhase() == domain.ReversisPhaseRoundEnd {
		return "", "reversis.roundEnd", map[string]string{
			"round": strconv.Itoa(r.GetRoundNumber()),
			"pool":  strconv.Itoa(r.GetPool()),
		}
	}
	return "", "reversis.play", map[string]string{"pool": strconv.Itoa(r.GetPool())}
}

// HintOutput ヒント情報をJSON出力する
func (p *ReversisWebPresenter) HintOutput(r interfaces.ReversisGame) string {
	resObj := p.buildBase(r)
	if hint := r.GetHint(); hint != nil {
		resObj.Hint = &controller.ReversisWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *ReversisWebPresenter) ActionLogOutput(r interfaces.ReversisGame) string {
	return actionLogOutputJSON(r)
}
