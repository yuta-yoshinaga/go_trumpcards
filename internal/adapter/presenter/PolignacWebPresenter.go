//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PolignacWebPresenter ポリニャックWebプレゼンタークラス
type PolignacWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PolignacWebPresenter) Output(g interfaces.PolignacGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.PolignacWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *PolignacWebPresenter) buildBase(g interfaces.PolignacGame) *controller.PolignacWebOutput {
	resObj := new(controller.PolignacWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CapotIdx = g.GetCapotIdx()
	resObj.CapotTricks = g.GetCapotTricks()
	resObj.ValidPlays = intSliceOrEmpty(g.GetValidPlayIndices(0))
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	resObj.Config = controller.PolignacWebOutputConfig{Rounds: g.GetConfig().Rounds}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PolignacWebPresenter) buildPlayersOutput(g interfaces.PolignacGame) []*controller.PolignacWebOutputPlayer {
	out := make([]*controller.PolignacWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		out = append(out, &controller.PolignacWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			Score:         player.GetScore(),
			RoundPenalty:  player.GetRoundPenalty(),
			TrickCount:    player.GetTrickCount(),
			DeclaredCapot: player.GetDeclaredCapot(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PolignacWebPresenter) buildMessage(g interfaces.PolignacGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() < 0 {
			return "", "polignac.result.tie", nil
		}
		return "", "polignac.result.winner", map[string]string{
			"idx":   strconv.Itoa(g.GetWinnerIdx()),
			"score": strconv.Itoa(g.GetPlayer(g.GetWinnerIdx()).GetScore()),
		}
	}
	switch g.GetPhase() {
	case domain.PolignacPhaseDeclare:
		return "", "polignac.declarePhase", nil
	case domain.PolignacPhaseRoundEnd:
		return "", "polignac.roundEnd", map[string]string{"round": strconv.Itoa(g.GetRoundNumber())}
	}
	// **capot 宣言中は狙いが変わる。** 失点を避けるより、宣言を潰すほうが大きい。
	if g.GetCapotIdx() >= 0 {
		return "", "polignac.play.capotActive", map[string]string{
			"idx": strconv.Itoa(g.GetCapotIdx()),
		}
	}
	return "", "polignac.play.normal", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *PolignacWebPresenter) HintOutput(g interfaces.PolignacGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.PolignacWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *PolignacWebPresenter) ActionLogOutput(g interfaces.PolignacGame) string {
	return actionLogOutputJSON(g)
}
