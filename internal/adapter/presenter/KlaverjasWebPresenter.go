//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KlaverjasWebPresenter クラヴァヤスのWebプレゼンタークラス
type KlaverjasWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KlaverjasWebPresenter) Output(g interfaces.KlaverjasGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *KlaverjasWebPresenter) buildBase(g interfaces.KlaverjasGame) *controller.KlaverjasWebOutput {
	resObj := new(controller.KlaverjasWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.TeamScores = g.GetTeamScores()
	resObj.RoundCardPoints = g.GetRoundCardPoints()
	resObj.RoundRoem = g.GetRoundRoem()
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.KlaverjasWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *KlaverjasWebPresenter) playableIndices(g interfaces.KlaverjasGame) []int {
	if g.GetPhase() != domain.KlaverjasPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *KlaverjasWebPresenter) buildPlayersOutput(g interfaces.KlaverjasGame) []*controller.KlaverjasWebOutputPlayer {
	teamScores := g.GetTeamScores()
	out := make([]*controller.KlaverjasWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		team := domain.KlaverjasTeamOf(i)
		out = append(out, &controller.KlaverjasWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			TeamScore:  teamScores[team],
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *KlaverjasWebPresenter) buildMessage(g interfaces.KlaverjasGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.KlaverjasPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "klaverjas.playPhase.lead", nil
		}
		return "", "klaverjas.playPhase.follow", nil
	case domain.KlaverjasPhaseTrickEnd:
		return "", "klaverjas.trickEnd", nil
	case domain.KlaverjasPhaseRoundEnd:
		return "", "klaverjas.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者チームメッセージを構築する
func (p *KlaverjasWebPresenter) winnerMessage(g interfaces.KlaverjasGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.KlaverjasTeamOf(i)
			break
		}
	}
	if humanTeam >= 0 && winnerTeam == humanTeam {
		return "ゲーム終了！ あなたのチームの勝ち！", "klaverjas.result.humanWin", nil
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam), "klaverjas.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *KlaverjasWebPresenter) HintOutput(g interfaces.KlaverjasGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *KlaverjasWebPresenter) ActionLogOutput(g interfaces.KlaverjasGame) string {
	return actionLogOutputJSON(g)
}
